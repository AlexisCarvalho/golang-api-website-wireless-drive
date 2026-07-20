package handler

import (
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"wireless_gallery/internal/middleware"
	"wireless_gallery/internal/service"
	"wireless_gallery/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type MediaHandler struct {
	service service.MediaService
}

type StreamClaims struct {
	MediaID uint `json:"media_id"`
	UserID  uint `json:"user_id"`

	jwt.RegisteredClaims
}

func NewMediaHandler(s service.MediaService) *MediaHandler {
	return &MediaHandler{service: s}
}

func (h *MediaHandler) RegisterRoutes(r *gin.Engine) {
	mediaRoutes := r.Group("/api/media")
	{
		mediaRoutes.POST("/upload", middleware.AuthMiddleware(), h.UploadMedia)
		mediaRoutes.GET("/owner", middleware.AuthMiddleware(), h.GetMediaByOwner)
		mediaRoutes.GET("/owner/missing-thumbnails", middleware.AuthMiddleware(), h.GetMediaWithMissingThumbnails)
		mediaRoutes.GET("/:id/file", middleware.AuthMiddleware(), h.GetMediaFile)
		mediaRoutes.GET("/:id/stream-url", middleware.AuthMiddleware(), h.GetStreamURL)
		mediaRoutes.GET("/:id/stream", h.StreamMedia)
		mediaRoutes.POST("/:id/generate-thumbnail", middleware.AuthMiddleware(), h.GenerateThumbnail)
		mediaRoutes.POST("/:id/delete-thumbnail", middleware.AuthMiddleware(), h.DeleteThumbnail)
		mediaRoutes.GET("/:id", middleware.AuthMiddleware(), h.GetMedia)
		mediaRoutes.PUT("/:id", middleware.AuthMiddleware(), h.UpdateMedia)
		mediaRoutes.DELETE("/:id", middleware.AuthMiddleware(), h.DeleteMedia)
	}
}

func (h *MediaHandler) GetStreamURL(c *gin.Context) {
	id := c.Param("id")

	mediaID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID inválido",
		})
		return
	}

	currentUserIDInterface := c.MustGet("userID")

	var currentUserID uint

	switch v := currentUserIDInterface.(type) {
	case uint:
		currentUserID = v
	case uint64:
		currentUserID = uint(v)
	case float64:
		currentUserID = uint(v)
	default:
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "ID de usuário inválido",
		})
		return
	}

	media, err := h.service.GetMediaByID(uint(mediaID))

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Mídia não encontrada",
		})
		return
	}

	if media.OwnerID != currentUserID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Acesso negado",
		})
		return
	}

	claims := StreamClaims{
		MediaID: media.ID,
		UserID:  currentUserID,

		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(
				time.Now().Add(5 * time.Minute),
			),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	signedToken, err := token.SignedString(
		[]byte(os.Getenv("STREAM_SECRET")),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao gerar token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url": fmt.Sprintf(
			"/api/media/%d/stream?token=%s",
			media.ID,
			signedToken,
		),
	})
}

func (h *MediaHandler) StreamMedia(c *gin.Context) {

	tokenString := c.Query("token")

	token, err := jwt.ParseWithClaims(
		tokenString,
		&StreamClaims{},
		func(token *jwt.Token) (interface{}, error) {

			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("método de assinatura inválido")
			}

			return []byte(os.Getenv("STREAM_SECRET")), nil
		},
	)

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Token inválido ou expirado",
		})
		return
	}

	claims, ok := token.Claims.(*StreamClaims)

	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Claims inválidas",
		})
		return
	}

	media, err := h.service.GetMediaByID(
		claims.MediaID,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Mídia não encontrada",
		})
		return
	}

	// Verifica se o token pertence ao dono da mídia
	if media.OwnerID != claims.UserID {

		c.JSON(http.StatusForbidden, gin.H{
			"error": "Acesso negado",
		})

		return
	}

	fileType := utils.StringToFileType(media.Type)

	fullPath := utils.GetFullMediaPath(
		fileType,
		media.Filename,
	)

	file, err := os.Open(fullPath)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Arquivo não encontrado",
		})
		return
	}

	defer file.Close()

	stat, err := file.Stat()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao obter informações do arquivo",
		})
		return
	}

	contentType := mime.TypeByExtension(
		filepath.Ext(media.Filename),
	)

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	c.Header(
		"Content-Type",
		contentType,
	)

	c.Header(
		"Accept-Ranges",
		"bytes",
	)

	http.ServeContent(
		c.Writer,
		c.Request,
		media.Filename,
		stat.ModTime(),
		file,
	)
}

// UploadMedia faz upload de um arquivo de mídia
func (h *MediaHandler) UploadMedia(c *gin.Context) {
	userIDInterface := c.MustGet("userID")
	var userID uint

	switch v := userIDInterface.(type) {
	case uint:
		userID = v
	case uint64:
		userID = uint(v)
	case float64:
		userID = uint(v)
	default:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID de usuário inválido"})
		return
	}

	// Obtém arquivo do form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Arquivo não encontrado"})
		return
	}

	// Obtém título e descrição opcionais
	title := c.PostForm("title")
	if title == "" {
		title = file.Filename
	}
	description := c.PostForm("description")

	// Faz upload
	media, err := h.service.UploadMedia(file, title, description, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Mídia enviada com sucesso",
		"media":   media,
	})
}

// GetMediaFile retorna o arquivo da mídia
func (h *MediaHandler) GetMediaFile(c *gin.Context) {
	id := c.Param("id")
	mediaID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	media, err := h.service.GetMediaByID(uint(mediaID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mídia não encontrada"})
		return
	}

	// Verifica permissão
	currentUserIDInterface := c.MustGet("userID")
	var currentUserID uint

	switch v := currentUserIDInterface.(type) {
	case uint:
		currentUserID = v
	case uint64:
		currentUserID = uint(v)
	case float64:
		currentUserID = uint(v)
	default:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID de usuário inválido"})
		return
	}

	if media.OwnerID != currentUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado"})
		return
	}

	// Obtém o caminho do arquivo
	fileType := utils.StringToFileType(media.Type)
	fullPath := utils.GetFullMediaPath(fileType, media.Filename)

	// Serve o arquivo
	file, err := os.Open(fullPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao abrir arquivo"})
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao obter informações do arquivo"})
		return
	}

	c.Header("Content-Disposition", "inline; filename=\""+media.Filename+"\"")
	c.Header("Accept-Ranges", "bytes")
	contentType := mime.TypeByExtension(filepath.Ext(media.Filename))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	c.Header("Content-Type", contentType)

	http.ServeContent(
		c.Writer,
		c.Request,
		media.Filename,
		stat.ModTime(),
		file,
	)
}

// GetMedia retorna uma mídia específica pelo ID
func (h *MediaHandler) GetMedia(c *gin.Context) {
	id := c.Param("id")
	mediaID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	media, err := h.service.GetMediaByID(uint(mediaID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mídia não encontrada"})
		return
	}

	// Verifica permissão
	currentUserIDInterface := c.MustGet("userID")
	var currentUserID uint

	switch v := currentUserIDInterface.(type) {
	case uint:
		currentUserID = v
	case uint64:
		currentUserID = uint(v)
	case float64:
		currentUserID = uint(v)
	default:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID de usuário inválido"})
		return
	}

	if media.OwnerID != currentUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado"})
		return
	}

	c.JSON(http.StatusOK, media)
}

// GetMediaByOwner retorna todas as mídias de um proprietário
func (h *MediaHandler) GetMediaByOwner(c *gin.Context) {
	// Verifica se o usuário está tentando acessar mídias de outro usuário
	currentUserIDInterface := c.MustGet("userID")
	var currentUserID uint

	switch v := currentUserIDInterface.(type) {
	case uint:
		currentUserID = v
	case uint64:
		currentUserID = uint(v)
	case float64:
		currentUserID = uint(v)
	default:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID de usuário inválido"})
		return
	}

	medias, err := h.service.GetMediaByOwner(currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"medias": medias,
	})
}

// UpdateMedia atualiza os dados de uma mídia
func (h *MediaHandler) UpdateMedia(c *gin.Context) {
	id := c.Param("id")
	mediaID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var updateReq struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}

	media, err := h.service.GetMediaByID(uint(mediaID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mídia não encontrada"})
		return
	}

	// Verifica se o usuário é o proprietário
	currentUserIDInterface := c.MustGet("userID")
	var currentUserID uint

	switch v := currentUserIDInterface.(type) {
	case uint:
		currentUserID = v
	case uint64:
		currentUserID = uint(v)
	case float64:
		currentUserID = uint(v)
	default:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID de usuário inválido"})
		return
	}

	if media.OwnerID != currentUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado"})
		return
	}

	// Atualiza dados
	if updateReq.Title != "" {
		media.Title = updateReq.Title
	}

	media.Description = updateReq.Description

	if err := h.service.UpdateMedia(media); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar mídia"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Mídia atualizada com sucesso",
		"media":   media,
	})
}

// DeleteMedia deleta uma mídia e seus arquivos
func (h *MediaHandler) DeleteMedia(c *gin.Context) {
	id := c.Param("id")
	mediaID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	// Obtém a mídia para verificar permissões
	media, err := h.service.GetMediaByID(uint(mediaID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mídia não encontrada"})
		return
	}

	// Verifica se o usuário é o proprietário
	currentUserIDInterface := c.MustGet("userID")
	var currentUserID uint

	switch v := currentUserIDInterface.(type) {
	case uint:
		currentUserID = v
	case uint64:
		currentUserID = uint(v)
	case float64:
		currentUserID = uint(v)
	default:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID de usuário inválido"})
		return
	}

	if media.OwnerID != currentUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado"})
		return
	}

	// Deleta a mídia
	if err := h.service.DeleteMedia(uint(mediaID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao deletar mídia"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Mídia deletada com sucesso",
	})
}

// GetMediaWithMissingThumbnails retorna mídias sem thumbnail
func (h *MediaHandler) GetMediaWithMissingThumbnails(c *gin.Context) {
	// Verifica se o usuário está tentando acessar mídias de outro usuário
	currentUserIDInterface := c.MustGet("userID")
	var currentUserID uint

	switch v := currentUserIDInterface.(type) {
	case uint:
		currentUserID = v
	case uint64:
		currentUserID = uint(v)
	case float64:
		currentUserID = uint(v)
	default:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID de usuário inválido"})
		return
	}

	medias, err := h.service.GetMediaByOwner(currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Filtra apenas as mídias sem thumbnail
	var missingThumbnails []interface{}
	for _, media := range medias {
		if media.Thumbnail == "" && (media.Type == "image" || media.Type == "video") {
			missingThumbnails = append(missingThumbnails, gin.H{
				"ID":       media.ID,
				"Title":    media.Title,
				"Type":     media.Type,
				"Filename": media.Filename,
				"MimeType": media.MimeType,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"count":  len(missingThumbnails),
		"medias": missingThumbnails,
	})
}

// GenerateThumbnail gera thumbnail para uma mídia específica usando a API externa
func (h *MediaHandler) GenerateThumbnail(c *gin.Context) {
	id := c.Param("id")
	mediaID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	media, err := h.service.GetMediaByID(uint(mediaID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mídia não encontrada"})
		return
	}

	// Verifica se o usuário é o proprietário
	currentUserIDInterface := c.MustGet("userID")
	var currentUserID uint

	switch v := currentUserIDInterface.(type) {
	case uint:
		currentUserID = v
	case uint64:
		currentUserID = uint(v)
	case float64:
		currentUserID = uint(v)
	default:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID de usuário inválido"})
		return
	}

	if media.OwnerID != currentUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado"})
		return
	}

	// Tenta gerar o thumbnail através da API de thumbnail
	thumbnailName, err := h.service.GenerateThumbnailViaAPI(media)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao gerar thumbnail: %v", err)})
		return
	}

	// Atualiza o registro da mídia com o novo thumbnail
	media.Thumbnail = thumbnailName
	if err := h.service.UpdateMedia(media); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar mídia"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Thumbnail gerado com sucesso",
		"thumbnail": thumbnailName,
	})
}

func (h *MediaHandler) DeleteThumbnail(c *gin.Context) {
	id := c.Param("id")
	mediaID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	media, err := h.service.GetMediaByID(uint(mediaID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mídia não encontrada"})
		return
	}

	// Verifica se o usuário é o proprietário
	currentUserIDInterface := c.MustGet("userID")
	var currentUserID uint

	switch v := currentUserIDInterface.(type) {
	case uint:
		currentUserID = v
	case uint64:
		currentUserID = uint(v)
	case float64:
		currentUserID = uint(v)
	default:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID de usuário inválido"})
		return
	}

	if media.OwnerID != currentUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado"})
		return
	}

	if err := h.service.DeleteThumbnail(uint(mediaID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao deletar thumbnail: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Thumbnail deletado com sucesso",
	})
}
