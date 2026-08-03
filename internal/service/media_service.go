package service

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"wireless_drive/internal/model"
	"wireless_drive/internal/repository"
	"wireless_drive/internal/utils"

	"github.com/google/uuid"
)

type MediaService interface {
	UploadMedia(file *multipart.FileHeader, title string, ownerID uint) (*model.Media, error)
	GetMediaByID(id uint) (*model.Media, error)
	GetMediaByOwner(ownerID uint) ([]model.Media, error)
	DeleteMedia(id uint) error
	DeleteThumbnail(mediaID uint) error
	UpdateMedia(media *model.Media) error
	GenerateThumbnail(fileType utils.FileType, filename, fullPath string) (string, error)
}

type mediaService struct {
	repo repository.MediaRepository
}

func NewMediaService(r repository.MediaRepository) MediaService {
	return &mediaService{repo: r}
}

func (s *mediaService) UploadMedia(file *multipart.FileHeader, title string, ownerID uint) (*model.Media, error) {
	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer src.Close()

	// Read the file to determine its type
	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Reset the file pointer
	src.Seek(0, 0)

	// Detect MIME type
	mimeType := file.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Determine the file type
	fileType := utils.DetermineFileType(mimeType)

	// Generate a unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s%s", uuid.NewString(), ext)
	fullPath := utils.GetFullMediaPath(fileType, filename)

	// Copy the file to the destination
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("error creating destination file: %w", err)
	}
	defer dst.Close()

	if _, err := src.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("error resetting file: %w", err)
	}

	if _, err := dst.ReadFrom(src); err != nil {
		os.Remove(fullPath)
		return nil, fmt.Errorf("error copying file: %w", err)
	}

	thumbnailName, err := s.GenerateThumbnail(fileType, filename, fullPath)
	if err != nil {
		return nil, err
	}

	// Create the database record
	media := &model.Media{
		Title:     title,
		Type:      string(fileType),
		Filename:  filename,
		Thumbnail: thumbnailName,
		MimeType:  mimeType,
		OwnerID:   ownerID,
	}

	if err := s.repo.Create(media); err != nil {
		os.Remove(fullPath)
		if thumbnailName != "" {
			os.Remove(utils.GetFullThumbPath(thumbnailName))
		}
		return nil, fmt.Errorf("error saving media to database: %w", err)
	}

	return media, nil
}

func (s *mediaService) GetMediaByID(id uint) (*model.Media, error) {
	return s.repo.FindByID(id)
}

func (s *mediaService) GetMediaByOwner(ownerID uint) ([]model.Media, error) {
	return s.repo.FindByOwnerID(ownerID)
}

func (s *mediaService) DeleteMedia(id uint) error {
	// Get the media item to delete its files
	media, err := s.repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("media not found: %w", err)
	}

	// Delete the main media file
	fileType := utils.StringToFileType(media.Type)
	if err := utils.DeleteMediaFile(fileType, media.Filename); err != nil {
		// Log the error but continue deletion
		fmt.Printf("Warning: failed to delete file: %v\n", err)
	}

	// Delete the thumbnail if it exists
	if media.Thumbnail != "" {
		if err := utils.DeleteThumbnailFile(media.Thumbnail); err != nil {
			fmt.Printf("Warning: failed to delete thumbnail: %v\n", err)
		}
	}

	// Delete the database record
	return s.repo.Delete(id)
}

func (s *mediaService) DeleteThumbnail(mediaID uint) error {
	media, err := s.repo.FindByID(mediaID)
	if err != nil {
		return fmt.Errorf("media not found: %w", err)
	}

	if media.Thumbnail == "" {
		return nil
	}

	if err := utils.DeleteThumbnailFile(media.Thumbnail); err != nil {
		fmt.Printf("Warning: failed to delete thumbnail: %v\n", err)
	}

	media.Thumbnail = ""
	return s.repo.Update(media)
}

func (s *mediaService) UpdateMedia(media *model.Media) error {
	return s.repo.Update(media)
}

// GenerateThumbnail creates a thumbnail using local ffmpeg when available, otherwise external API as a fallback.
// If generation fails, it returns an empty string without interrupting the main flow.
func (s *mediaService) GenerateThumbnail(fileType utils.FileType, filename, fullPath string) (string, error) {
	if fileType != utils.IMAGE && fileType != utils.VIDEO {
		return "", nil
	}

	thumbnailName := utils.GenerateThumbnailName(filename)
	if utils.IsFFmpegInstalled() && utils.IsFFprobeInstalled() {

		thumbPath := utils.GetFullThumbPath(thumbnailName)
		var err error
		switch fileType {
		case utils.IMAGE:
			err = utils.GenerateImageThumbnail(fullPath, thumbPath)
		case utils.VIDEO:
			err = utils.GenerateVideoThumbnail(fullPath, thumbPath)
		}

		if err != nil {
			fmt.Printf("Warning: failed to generate thumbnail locally: %v\n", err)
		}

		return thumbnailName, nil
	}

	thumbnailName, err := utils.GenerateThumbnailViaExternalAPI(fileType, filename)
	if err != nil {
		fmt.Printf("Warning: failed to generate thumbnail via external API: %v\n", err)
		return "", nil
	}

	return thumbnailName, nil
}
