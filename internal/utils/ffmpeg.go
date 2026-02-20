package utils

import (
	"fmt"
	"os/exec"
	"path/filepath"
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

// GenerateThumbnailName gera um nome único para o thumbnail
func GenerateThumbnailName(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	nameWithoutExt := originalFilename[:len(originalFilename)-len(ext)]
	return nameWithoutExt + "_thumb.jpg"
}
