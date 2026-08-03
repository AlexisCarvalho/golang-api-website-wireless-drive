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

// allFileTypes lists every media type whose upload directory must exist
var allFileTypes = []FileType{IMAGE, VIDEO, AUDIO, OTHERS}

const (
	defaultUploadsDir = "uploads"
	defaultThumbsDir  = "thumbs"
)

var (
	// Holds BASE_PATH exactly as configured, with no "." fallback.
	// BASE_PATH must be explicitly set, since silently creating directories
	// relative to the working directory at startup could mask a missing configuration.
	// if you want to use the same folder the executable is as BASE_PATH to store the files
	// just put (BASE_PATH=.) on the .env
	basePath = config.GetEnv("BASE_PATH", "")

	uploadsDir = config.GetEnv("UPLOADS_DIR", defaultUploadsDir)
	thumbsDir  = config.GetEnv("THUMBS_DIR", defaultThumbsDir)
)

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// StringToFileType returns the FileType based on a string
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

// GetMediaStoragePath returns the full storage path for media based on its type
func GetMediaStoragePath(fileType FileType) string {
	return filepath.Join(basePath, uploadsDir, string(fileType))
}

// EnsureBaseDir creates the base directory. Unlike the other Ensure*
// functions, BASE_PATH must be explicitly configured — there is no
// fallback.
func EnsureBaseDir() error {
	if basePath == "" {
		return fmt.Errorf("BASE_PATH is not configured")
	}

	return os.MkdirAll(basePath, 0755)
}

// EnsureThumbsDir creates the thumbnail directory if it does not exist
func EnsureThumbsDir() error {
	return os.MkdirAll(GetThumbsPath(), 0755)
}

// EnsureUploadDir creates the required upload directories if they do not exist
func EnsureUploadDir(fileType FileType) error {
	return os.MkdirAll(GetMediaStoragePath(fileType), 0755)
}

// EnsureAllDirs guarantees that the base directory, the thumbs directory
// and every media type's upload directory exist. Should be called once
// during application startup.
func EnsureAllDirs() error {
	if err := EnsureBaseDir(); err != nil {
		return fmt.Errorf("failed to ensure base dir: %w", err)
	}

	if err := EnsureThumbsDir(); err != nil {
		return fmt.Errorf("failed to ensure thumbs dir: %w", err)
	}

	for _, ft := range allFileTypes {
		if err := EnsureUploadDir(ft); err != nil {
			return fmt.Errorf("failed to ensure upload dir for %q: %w", ft, err)
		}
	}

	return nil
}

// GetFullMediaPath returns the full file path for media
func GetFullMediaPath(fileType FileType, filename string) string {
	return filepath.Join(GetMediaStoragePath(fileType), filename)
}

// DetermineFileType determines the file type from the MIME type
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

// DeleteMediaFile deletes the physical media file from storage
func DeleteMediaFile(fileType FileType, filename string) error {
	return os.Remove(GetFullMediaPath(fileType, filename))
}

// DeleteThumbnailFile deletes the thumbnail file
func DeleteThumbnailFile(filename string) error {
	return os.Remove(GetFullThumbPath(filename))
}

// GetThumbsPath returns the full path of the thumbnails directory
func GetThumbsPath() string {
	return filepath.Join(basePath, thumbsDir)
}

// GetFullThumbPath returns the full path of a thumbnail file
func GetFullThumbPath(filename string) string {
	return filepath.Join(GetThumbsPath(), filename)
}

// GenerateThumbnailName generates a unique thumbnail name based on the file name
func GenerateThumbnailName(filename string) string {
	ext := filepath.Ext(filename)
	nameWithoutExt := filename[:len(filename)-len(ext)]
	return fmt.Sprintf("thumb_%s.jpg", nameWithoutExt)
}
