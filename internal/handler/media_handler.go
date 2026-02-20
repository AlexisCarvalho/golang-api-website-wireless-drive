package handler

import (
	"net/http"
	"strconv"
	"wireless_gallery/internal/middleware"
	"wireless_gallery/internal/service"
	"wireless_gallery/internal/utils"

	"github.com/gin-gonic/gin"
)

type MediaHandler struct {
	service service.MediaService
}

func NewMediaHandler(s service.MediaService) *MediaHandler {
	return &MediaHandler{service: s}
}

func (h *MediaHandler) RegisterRoutes(r *gin.Engine) {
	mediaRoutes := r.Group("/api/media")
	{
		mediaRoutes.POST("/upload", middleware.AuthMiddleware(), h.UploadMedia)
		mediaRoutes.GET("/:id", middleware.AuthMiddleware(), h.GetMedia)
		mediaRoutes.GET("/:id/file", middleware.AuthMiddleware(), h.GetMediaFile)
		mediaRoutes.GET("/owner/:ownerID", middleware.AuthMiddleware(), h.GetMediaByOwner)
		mediaRoutes.PUT("/:id", middleware.AuthMiddleware(), h.UpdateMedia)
		mediaRoutes.DELETE("/:id", middleware.AuthMiddleware(), h.DeleteMedia)
	}
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
	media, err := h.service.UploadMedia(file, title, description, uint(userID))
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
	c.File(fullPath)
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

	c.JSON(http.StatusOK, media)
}

// GetMediaByOwner retorna todas as mídias de um proprietário
func (h *MediaHandler) GetMediaByOwner(c *gin.Context) {
	ownerID := c.Param("ownerID")
	userID, err := strconv.ParseUint(ownerID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do proprietário inválido"})
		return
	}

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

	if currentUserID != uint(userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado"})
		return
	}

	medias, err := h.service.GetMediaByOwner(uint(userID))
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
	if updateReq.Description != "" {
		media.Description = updateReq.Description
	}

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
