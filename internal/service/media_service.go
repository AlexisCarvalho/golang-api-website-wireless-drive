package service

import (
	"fmt"
	"mime/multipart"
	"os"
	"wireless_gallery/internal/model"
	"wireless_gallery/internal/repository"
	"wireless_gallery/internal/utils"
)

type MediaService interface {
	UploadMedia(file *multipart.FileHeader, title, description string, ownerID uint) (*model.Media, error)
	GetMediaByID(id uint) (*model.Media, error)
	GetMediaByOwner(ownerID uint) ([]model.Media, error)
	DeleteMedia(id uint) error
	UpdateMedia(media *model.Media) error
	GenerateThumbnailViaAPI(media *model.Media) (string, error)
}

type mediaService struct {
	repo repository.MediaRepository
}

func NewMediaService(r repository.MediaRepository) MediaService {
	return &mediaService{repo: r}
}

func (s *mediaService) UploadMedia(file *multipart.FileHeader, title, description string, ownerID uint) (*model.Media, error) {
	// Abre o arquivo enviado
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer src.Close()

	// Lê o arquivo para determinar tipo
	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo: %w", err)
	}

	// Reseta o pointer do arquivo
	src.Seek(0, 0)

	// Detecta MIME type
	mimeType := file.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Determina o tipo de arquivo
	fileType := utils.DetermineFileType(mimeType)

	// Cria diretório se não existir
	if err := utils.EnsureUploadDir(fileType); err != nil {
		return nil, fmt.Errorf("erro ao criar diretório de upload: %w", err)
	}

	// Gera nome único para o arquivo
	filename := fmt.Sprintf("%d_%s", ownerID, file.Filename)
	fullPath := utils.GetFullMediaPath(fileType, filename)

	// Copia arquivo para o destino
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar arquivo de destino: %w", err)
	}
	defer dst.Close()

	if _, err := src.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("erro ao resetar arquivo: %w", err)
	}

	if _, err := dst.ReadFrom(src); err != nil {
		os.Remove(fullPath)
		return nil, fmt.Errorf("erro ao copiar arquivo: %w", err)
	}

	// Cria thumbnail
	thumbnailName := utils.GenerateThumbnailName(filename)
	if err := utils.EnsureThumbsDir(); err != nil {
		return nil, fmt.Errorf("erro ao criar diretório de thumbnails: %w", err)
	}

	thumbPath := utils.GetFullThumbPath(thumbnailName)

	switch fileType {
	case utils.IMAGE:
		if err := utils.GenerateImageThumbnail(fullPath, thumbPath, 320, 320); err != nil {
			// Log do erro mas não falha o upload se thumbnail falhar
			fmt.Printf("Aviso: falha ao gerar thumbnail de imagem: %v\n", err)
			thumbnailName = ""
		}
	case utils.VIDEO:
		if err := utils.GenerateVideoThumbnail(fullPath, thumbPath, 320, 320); err != nil {
			// Log do erro mas não falha o upload se thumbnail falhar
			fmt.Printf("Aviso: falha ao gerar thumbnail de vídeo: %v\n", err)
			thumbnailName = ""
		}
	}

	// Cria registro no banco de dados
	media := &model.Media{
		Title:       title,
		Description: description,
		Type:        string(fileType),
		Filename:    filename,
		Thumbnail:   thumbnailName,
		MimeType:    mimeType,
		OwnerID:     ownerID,
	}

	if err := s.repo.Create(media); err != nil {
		os.Remove(fullPath)
		if thumbnailName != "" {
			os.Remove(thumbPath)
		}
		return nil, fmt.Errorf("erro ao salvar mídia no banco de dados: %w", err)
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
	// Obtém a mídia para deletar os arquivos
	media, err := s.repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("mídia não encontrada: %w", err)
	}

	// Deleta arquivo principal
	fileType := utils.StringToFileType(media.Type)
	if err := utils.DeleteMediaFile(fileType, media.Filename); err != nil {
		// Log do erro mas continua a deleção
		fmt.Printf("Aviso: erro ao deletar arquivo: %v\n", err)
	}

	// Deleta thumbnail se existir
	if media.Thumbnail != "" {
		if err := utils.DeleteThumbnailFile(media.Thumbnail); err != nil {
			fmt.Printf("Aviso: erro ao deletar thumbnail: %v\n", err)
		}
	}

	// Deleta registro do banco de dados
	return s.repo.Delete(id)
}

func (s *mediaService) UpdateMedia(media *model.Media) error {
	return s.repo.Update(media)
}

// GenerateThumbnailViaAPI gera thumbnail através da API externa
func (s *mediaService) GenerateThumbnailViaAPI(media *model.Media) (string, error) {
	return utils.GenerateThumbnailViaExternalAPI(media.Type, utils.StringToFileType(media.Type), media.Filename)
}
