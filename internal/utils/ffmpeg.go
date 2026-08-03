package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"wireless_drive/internal/config"
)

const (
	// Default binary paths used when the corresponding env vars are not set
	// Try to access the ffmpeg and ffprobe native installed on the system
	defaultFFmpegPath  = "ffmpeg"
	defaultFFprobePath = "ffprobe"

	// Video thumbnail is captured at this timestamp
	videoThumbnailTimestamp = "00:00:05"

	// Thumbnail defaults
	thumbnailWidth  = 320
	thumbnailHeight = 320
)

var (
	ffmpegPath  = config.GetEnv("FFMPEG_PATH", defaultFFmpegPath)
	ffprobePath = config.GetEnv("FFPROBE_PATH", defaultFFprobePath)

	// thumbnailAPIURL has no default: GenerateThumbnailViaExternalAPI
	// checks it explicitly and fails fast if it was never configured.
	thumbnailAPIURL = config.GetEnv("THUMBNAIL_API_URL", "")
)

// GenerateImageThumbnail generates a thumbnail for images using FFmpeg
func GenerateImageThumbnail(inputPath, outputPath string) error {
	cmd := exec.Command(ffmpegPath,
		"-i", inputPath,
		"-vf", fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease", thumbnailWidth, thumbnailHeight),
		"-y", // Overwrite output file
		outputPath,
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error generating image thumbnail: %w", err)
	}

	return nil
}

// GenerateVideoThumbnail generates a thumbnail for videos from the first second
func GenerateVideoThumbnail(inputPath, outputPath string) error {
	cmd := exec.Command(ffmpegPath,
		"-i", inputPath,
		"-ss", videoThumbnailTimestamp,
		"-vf", fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease", thumbnailWidth, thumbnailHeight),
		"-vframes", "1",
		"-y", // Overwrite output file
		outputPath,
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error generating video thumbnail: %w", err)
	}

	return nil
}

// IsFFmpegInstalled checks whether FFmpeg is installed
func IsFFmpegInstalled() bool {
	cmd := exec.Command(ffmpegPath, "-version")
	return cmd.Run() == nil
}

// IsFFprobeInstalled checks whether FFprobe is installed
func IsFFprobeInstalled() bool {
	cmd := exec.Command(ffprobePath, "-version")
	return cmd.Run() == nil
}

// GenerateThumbnailViaExternalAPI generates a thumbnail through the external API using streaming upload.
func GenerateThumbnailViaExternalAPI(fileType FileType, filename string) (string, error) {
	if thumbnailAPIURL == "" {
		return "", fmt.Errorf("THUMBNAIL_API_URL is not configured")
	}

	filePath := GetFullMediaPath(fileType, filename)

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	endpoint := "/thumbnail/image"
	if string(fileType) == "video" {
		endpoint = "/thumbnail/video"
	}

	// Pipe used for streaming
	pr, pw := io.Pipe()

	writer := multipart.NewWriter(pw)

	// Writes the multipart body while the HTTP client streams the data
	go func() {
		defer pw.Close()
		defer writer.Close()

		part, err := writer.CreateFormFile("file", filename)
		if err != nil {
			pw.CloseWithError(err)
			return
		}

		if _, err := io.Copy(part, file); err != nil {
			pw.CloseWithError(err)
			return
		}

		if err := writer.WriteField("width", strconv.Itoa(thumbnailWidth)); err != nil {
			pw.CloseWithError(err)
			return
		}

		if err := writer.WriteField("height", strconv.Itoa(thumbnailHeight)); err != nil {
			pw.CloseWithError(err)
			return
		}

		if string(fileType) == "video" {
			parts := strings.Split(videoThumbnailTimestamp, ":")
			seconds := parts[2]
			if err := writer.WriteField("second", seconds); err != nil {
				pw.CloseWithError(err)
				return
			}
		}
	}()

	req, err := http.NewRequest(http.MethodPost, thumbnailAPIURL+endpoint, pr)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error connecting to thumbnail API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API retornou status %d: %s", resp.StatusCode, string(body))
	}

	thumbData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	thumbName := GenerateThumbnailName(filename)
	thumbPath := GetFullThumbPath(thumbName)

	if err := os.WriteFile(thumbPath, thumbData, 0644); err != nil {
		return "", fmt.Errorf("error saving thumbnail: %w", err)
	}

	return thumbName, nil
}
