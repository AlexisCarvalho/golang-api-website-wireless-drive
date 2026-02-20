package utils

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
)

// GenerateImageThumbnail gera thumbnail para imagens usando FFmpeg
func GenerateImageThumbnail(inputPath, outputPath string, width, height int) error {
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-vf", fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease", width, height),
		"-y", // Overwrite output file
		outputPath,
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("erro ao gerar thumbnail de imagem: %w", err)
	}

	return nil
}

// GenerateVideoThumbnail gera thumbnail para vídeos no primeiro frame
func GenerateVideoThumbnail(inputPath, outputPath string, width, height int) error {
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-ss", "00:00:01", // Captura no 1º segundo
		"-vf", fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease", width, height),
		"-vframes", "1",
		"-y", // Overwrite output file
		outputPath,
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("erro ao gerar thumbnail de vídeo: %w", err)
	}

	return nil
}

// GetVideoMetadata retorna informações do vídeo (duração, resolução, etc)
func GetVideoMetadata(filePath string) (map[string]string, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration:stream=width,height,codec_type",
		"-of", "default=noprint_wrappers=1:nokey=1:nokey=1",
		filePath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter metadados do vídeo: %w", err)
	}

	metadata := make(map[string]string)
	metadata["output"] = string(output)
	return metadata, nil
}

// ConvertVideoFormat converte vídeo para um formato específico
func ConvertVideoFormat(inputPath, outputPath, format string) error {
	// Mapeia formatos conhecidos para parâmetros FFmpeg apropriados
	formatParams := map[string][]string{
		"mp4":  {"-c:v", "libx264", "-crf", "23", "-c:a", "aac", "-b:a", "128k"},
		"webm": {"-c:v", "libvpx", "-crf", "30", "-c:a", "libopus"},
		"mkv":  {"-c:v", "libx264", "-crf", "23", "-c:a", "aac"},
	}

	params, exists := formatParams[format]
	if !exists {
		return fmt.Errorf("formato de vídeo não suportado: %s", format)
	}

	args := []string{"-i", inputPath}
	args = append(args, params...)
	args = append(args, "-y", outputPath)

	cmd := exec.Command("ffmpeg", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("erro ao converter vídeo: %w", err)
	}

	return nil
}

// CompressVideo comprime o vídeo para otimização
func CompressVideo(inputPath, outputPath string, crf int) error {
	if crf < 0 || crf > 51 {
		crf = 23 // valor padrão
	}

	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-c:v", "libx264",
		"-crf", fmt.Sprintf("%d", crf),
		"-c:a", "aac",
		"-b:a", "128k",
		"-y",
		outputPath,
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("erro ao comprimir vídeo: %w", err)
	}

	return nil
}

// IsFFmpegInstalled verifica se FFmpeg está instalado
func IsFFmpegInstalled() bool {
	cmd := exec.Command("ffmpeg", "-version")
	return cmd.Run() == nil
}

// IsFFprobeInstalled verifica se FFprobe está instalado
func IsFFprobeInstalled() bool {
	cmd := exec.Command("ffprobe", "-version")
	return cmd.Run() == nil
}

// GenerateThumbnailViaExternalAPI gera thumbnail através da API externa
func GenerateThumbnailViaExternalAPI(mediaType string, fileType FileType, filename string) (string, error) {
	apiURL := os.Getenv("THUMBNAIL_API_URL")
	if apiURL == "" {
		apiURL = "http://192.168.0.139:8086"
	}

	// Abre o arquivo de mídia
	filePath := GetFullMediaPath(fileType, filename)
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer file.Close()

	// Determina o endpoint baseado no tipo
	endpoint := "/thumbnail/image"
	if mediaType == "video" {
		endpoint = "/thumbnail/video"
	}

	// Cria um buffer para o arquivo
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Adiciona o arquivo ao formulário
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("erro ao criar form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return "", fmt.Errorf("erro ao copiar arquivo: %w", err)
	}

	// Adiciona parâmetros opcionais
	writer.WriteField("width", "320")
	writer.WriteField("height", "320")
	if mediaType == "video" {
		writer.WriteField("second", "1")
	}

	writer.Close()

	// Faz requisição POST
	resp, err := http.Post(apiURL+endpoint, writer.FormDataContentType(), body)
	if err != nil {
		return "", fmt.Errorf("erro ao conectar com API de thumbnail: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API retornou status %d", resp.StatusCode)
	}

	// Lê a resposta (thumbnail em bytes)
	thumbData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("erro ao ler resposta: %w", err)
	}

	// Salva o thumbnail
	if err := EnsureThumbsDir(); err != nil {
		return "", fmt.Errorf("erro ao criar diretório de thumbnails: %w", err)
	}

	thumbName := GenerateThumbnailName(filename)
	thumbPath := GetFullThumbPath(thumbName)

	if err := ioutil.WriteFile(thumbPath, thumbData, 0644); err != nil {
		return "", fmt.Errorf("erro ao salvar thumbnail: %w", err)
	}

	return thumbName, nil
}
