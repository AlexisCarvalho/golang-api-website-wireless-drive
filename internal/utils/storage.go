package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"wireless_drive/internal/config"
)

type FileType string

const (
	IMAGE  FileType = "image"
	VIDEO  FileType = "video"
	AUDIO  FileType = "audio"
	OTHERS FileType = "others"
)

// StringToFileType retorna o tipo de arquivo baseado na string
func StringToFileType(t string) FileType {
	switch t {
	case "image":
		return IMAGE
	case "video":
		return VIDEO
	case "audio":
		return AUDIO
	default:
		return OTHERS
	}
}

// GetMediaStoragePath retorna o caminho completo para armazenar a mídia baseado no tipo
func GetMediaStoragePath(fileType FileType) string {
	uploadsDir := config.GetEnv("UPLOADS_DIR", "uploads")
	basePath := config.GetEnv("BASE_PATH", ".")

	mediaPath := filepath.Join(basePath, uploadsDir, string(fileType))
	return mediaPath
}

// EnsureUploadDir cria os diretórios necessários se não existirem
func EnsureUploadDir(fileType FileType) error {
	path := GetMediaStoragePath(fileType)
	return os.MkdirAll(path, 0755)
}

// GetFullMediaPath retorna o caminho completo do arquivo
func GetFullMediaPath(fileType FileType, filename string) string {
	return filepath.Join(GetMediaStoragePath(fileType), filename)
}

// DetermineFileType determina o tipo de arquivo baseado na MIME type
func DetermineFileType(mimeType string) FileType {
	switch {
	case len(mimeType) >= 6 && mimeType[:6] == "image/":
		return IMAGE
	case len(mimeType) >= 6 && mimeType[:6] == "video/":
		return VIDEO
	case len(mimeType) >= 6 && mimeType[:6] == "audio/":
		return AUDIO
	default:
		return OTHERS
	}
}

// DeleteMediaFile deleta o arquivo físico do armazenamento
func DeleteMediaFile(fileType FileType, filename string) error {
	fullPath := GetFullMediaPath(fileType, filename)
	return os.Remove(fullPath)
}

// DeleteThumbnailFile deleta o arquivo de thumbnail
func DeleteThumbnailFile(filename string) error {
	thumbsDir := config.GetEnv("THUMBS_DIR", "thumbs")
	basePath := config.GetEnv("BASE_PATH", ".")
	fullPath := filepath.Join(basePath, thumbsDir, filename)
	return os.Remove(fullPath)
}

// EnsureThumbsDir cria o diretório de thumbnails se não existir
func EnsureThumbsDir() error {
	thumbsDir := config.GetEnv("THUMBS_DIR", "thumbs")
	basePath := config.GetEnv("BASE_PATH", ".")
	path := filepath.Join(basePath, thumbsDir)
	return os.MkdirAll(path, 0755)
}

// GetThumbsPath retorna o caminho completo do diretório de thumbnails
func GetThumbsPath() string {
	thumbsDir := config.GetEnv("THUMBS_DIR", "thumbs")
	basePath := config.GetEnv("BASE_PATH", ".")
	return filepath.Join(basePath, thumbsDir)
}

// GetFullThumbPath retorna o caminho completo do arquivo de thumbnail
func GetFullThumbPath(filename string) string {
	return filepath.Join(GetThumbsPath(), filename)
}

// GenerateThumbnailName gera um nome único para o thumbnail baseado no nome do arquivo
func GenerateThumbnailName(filename string) string {
	ext := filepath.Ext(filename)
	nameWithoutExt := filename[:len(filename)-len(ext)]
	return fmt.Sprintf("thumb_%s.jpg", nameWithoutExt)
}
