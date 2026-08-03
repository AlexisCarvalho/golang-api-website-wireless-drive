package handler

import (
	"errors"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"wireless_drive/internal/config"
	"wireless_drive/internal/middleware"
	"wireless_drive/internal/service"
	"wireless_drive/internal/utils"

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

// streamSecret signs and validates short-lived stream/download tokens.
var streamSecret = mustLoadStreamSecret()

func mustLoadStreamSecret() []byte {
	secret := config.GetEnv("STREAM_SECRET", "")
	if secret == "" {
		log.Fatal("STREAM_SECRET is not configured")
	}
	return []byte(secret)
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
		mediaRoutes.POST("/:id/generate-thumbnail", middleware.AuthMiddleware(), h.GenerateThumbnail)
		mediaRoutes.POST("/:id/delete-thumbnail", middleware.AuthMiddleware(), h.DeleteThumbnail)
		mediaRoutes.GET("/:id/file", middleware.AuthMiddleware(), h.GetMediaFile)
		mediaRoutes.GET("/:id/stream-url", middleware.AuthMiddleware(), h.GetStreamURL)
		mediaRoutes.GET("/:id/stream", h.StreamMedia)
		mediaRoutes.GET("/:id/download", h.DownloadMedia)
		mediaRoutes.GET("/:id", middleware.AuthMiddleware(), h.GetMedia)
		mediaRoutes.PUT("/:id", middleware.AuthMiddleware(), h.UpdateMedia)
		mediaRoutes.DELETE("/:id", middleware.AuthMiddleware(), h.DeleteMedia)
	}
}

// +-----------+
// | ENDPOINTS |
// +-----------+

// UploadMedia uploads a media file
func (h *MediaHandler) UploadMedia(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	// Get file from the form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File not found"})
		return
	}

	title := c.PostForm("title")
	if title == "" {
		title = file.Filename
	}

	// Upload the media
	media, err := h.service.UploadMedia(file, title, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Media uploaded successfully",
		"media":   media,
	})
}

// GetMediaByOwner returns all media for a given owner
func (h *MediaHandler) GetMediaByOwner(c *gin.Context) {
	currentUserID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	medias, ok := fetchOrServerError(c, h.service.GetMediaByOwner, currentUserID)
	if !ok {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"medias": medias,
	})
}

// GetMediaWithMissingThumbnails returns media items without thumbnails
func (h *MediaHandler) GetMediaWithMissingThumbnails(c *gin.Context) {
	currentUserID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	medias, ok := fetchOrServerError(c, h.service.GetMediaByOwner, currentUserID)
	if !ok {
		return
	}

	// Filter only media without thumbnails
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

// GenerateThumbnail creates a thumbnail for a specific media item using local ffmpeg or external API as a fallback
func (h *MediaHandler) GenerateThumbnail(c *gin.Context) {
	mediaID, ok := parseIDParam(c)
	if !ok {
		return
	}

	media, ok := fetchOrNotFound(c, h.service.GetMediaByID, mediaID, "Media not found")
	if !ok {
		return
	}

	currentUserID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	if !requireOwner(c, media.OwnerID, currentUserID) {
		return
	}

	if media.Type != string(utils.IMAGE) && media.Type != string(utils.VIDEO) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Media type does not support thumbnails"})
		return
	}

	fileType := utils.StringToFileType(media.Type)
	fullPath := utils.GetFullMediaPath(fileType, media.Filename)

	thumbnailName, err := h.service.GenerateThumbnail(fileType, media.Filename, fullPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error generating thumbnail: %v", err)})
		return
	}

	if thumbnailName != "" {
		media.Thumbnail = thumbnailName
		if err := h.service.UpdateMedia(media); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating media"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Thumbnail generated successfully",
		"thumbnail": thumbnailName,
	})
}

func (h *MediaHandler) DeleteThumbnail(c *gin.Context) {
	mediaID, ok := parseIDParam(c)
	if !ok {
		return
	}

	media, ok := fetchOrNotFound(c, h.service.GetMediaByID, mediaID, "Media not found")
	if !ok {
		return
	}

	currentUserID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	if !requireOwner(c, media.OwnerID, currentUserID) {
		return
	}

	if err := h.service.DeleteThumbnail(mediaID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error deleting thumbnail: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Thumbnail deleted successfully",
	})
}

// GetMediaFile returns the media file
func (h *MediaHandler) GetMediaFile(c *gin.Context) {
	mediaID, ok := parseIDParam(c)
	if !ok {
		return
	}

	media, ok := fetchOrNotFound(c, h.service.GetMediaByID, mediaID, "Media not found")
	if !ok {
		return
	}

	currentUserID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	if !requireOwner(c, media.OwnerID, currentUserID) {
		return
	}

	// Get the file path
	fileType := utils.StringToFileType(media.Type)
	fullPath := utils.GetFullMediaPath(fileType, media.Filename)

	// Serve the file
	file, err := os.Open(fullPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error opening file"})
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting file information"})
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

func (h *MediaHandler) GetStreamURL(c *gin.Context) {
	mediaID, ok := parseIDParam(c)
	if !ok {
		return
	}

	currentUserID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	media, ok := fetchOrNotFound(c, h.service.GetMediaByID, mediaID, "Media not found")
	if !ok {
		return
	}

	if !requireOwner(c, media.OwnerID, currentUserID) {
		return
	}

	claims := StreamClaims{
		MediaID: media.ID,
		UserID:  currentUserID,

		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(
				time.Now().Add(30 * time.Minute),
			),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	signedToken, err := token.SignedString(streamSecret)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error generating token",
		})
		return
	}

	action := c.DefaultQuery("type", "stream")

	var url string

	switch action {
	case "download":
		url = fmt.Sprintf(
			"/api/media/%d/download?token=%s",
			media.ID,
			signedToken,
		)

	default:
		url = fmt.Sprintf(
			"/api/media/%d/stream?token=%s",
			media.ID,
			signedToken,
		)
	}

	c.JSON(http.StatusOK, gin.H{
		"url": url,
	})
}

func (h *MediaHandler) StreamMedia(c *gin.Context) {
	h.serveMedia(c, false)
}

func (h *MediaHandler) DownloadMedia(c *gin.Context) {
	h.serveMedia(c, true)
}

// GetMedia returns a specific media item by ID
func (h *MediaHandler) GetMedia(c *gin.Context) {
	mediaID, ok := parseIDParam(c)
	if !ok {
		return
	}

	media, ok := fetchOrNotFound(c, h.service.GetMediaByID, mediaID, "Media not found")
	if !ok {
		return
	}

	currentUserID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	if !requireOwner(c, media.OwnerID, currentUserID) {
		return
	}

	c.JSON(http.StatusOK, media)
}

// UpdateMedia updates a media item's metadata
func (h *MediaHandler) UpdateMedia(c *gin.Context) {
	mediaID, ok := parseIDParam(c)
	if !ok {
		return
	}

	var updateReq struct {
		Title string `json:"title"`
	}

	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	media, ok := fetchOrNotFound(c, h.service.GetMediaByID, mediaID, "Media not found")
	if !ok {
		return
	}

	currentUserID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	if !requireOwner(c, media.OwnerID, currentUserID) {
		return
	}

	// Update fields
	if updateReq.Title != "" {
		media.Title = updateReq.Title
	}

	if err := h.service.UpdateMedia(media); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating media"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Media updated successfully",
		"media":   media,
	})
}

// DeleteMedia deletes a media item and its files
func (h *MediaHandler) DeleteMedia(c *gin.Context) {
	mediaID, ok := parseIDParam(c)
	if !ok {
		return
	}

	// Get the media item to verify permissions
	media, ok := fetchOrNotFound(c, h.service.GetMediaByID, mediaID, "Media not found")
	if !ok {
		return
	}

	currentUserID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	if !requireOwner(c, media.OwnerID, currentUserID) {
		return
	}

	// Delete the media item
	if err := h.service.DeleteMedia(mediaID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting media"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Media deleted successfully",
	})
}

// +-------------------+
// | GENERAL FUNCTIONS |
// +-------------------+

func sanitizeFileName(name string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)

	return strings.TrimSpace(replacer.Replace(name))
}

func (h *MediaHandler) serveMedia(
	c *gin.Context,
	download bool,
) {

	tokenString := c.Query("token")

	token, err := jwt.ParseWithClaims(
		tokenString,
		&StreamClaims{},
		func(token *jwt.Token) (interface{}, error) {

			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("método de assinatura inválido")
			}

			return streamSecret, nil
		},
	)

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid or expired token",
		})
		return
	}

	claims, ok := token.Claims.(*StreamClaims)

	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid claims",
		})
		return
	}

	media, ok := fetchOrNotFound(c, h.service.GetMediaByID, claims.MediaID, "Media not found")
	if !ok {
		return
	}

	if !requireOwner(c, media.OwnerID, claims.UserID) {
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
			"error": "File not found",
		})
		return
	}

	defer file.Close()

	stat, err := file.Stat()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error getting file information",
		})
		return
	}

	contentType := mime.TypeByExtension(
		filepath.Ext(media.Filename),
	)

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	c.Header("Content-Type", contentType)
	c.Header("Accept-Ranges", "bytes")

	fileName := media.Filename

	if download {

		extension := filepath.Ext(media.Filename)

		title := media.Title

		// Remove extension only if it's already on title
		if strings.HasSuffix(
			strings.ToLower(title),
			strings.ToLower(extension),
		) {
			title = title[:len(title)-len(extension)]
		}

		title = sanitizeFileName(title)

		fileName = fmt.Sprintf(
			"%d_%s%s",
			media.ID,
			title,
			extension,
		)

		c.Header(
			"Content-Disposition",
			fmt.Sprintf(`attachment; filename="%s"`, fileName),
		)
	}

	http.ServeContent(
		c.Writer,
		c.Request,
		fileName,
		stat.ModTime(),
		file,
	)
}

// +-----------------+
// | REQUEST HELPERS |
// +-----------------+

// getCurrentUserID extracts and normalizes the authenticated user ID
// stored in the gin context by the auth middleware.
func getCurrentUserID(c *gin.Context) (uint, bool) {
	switch v := c.MustGet("userID").(type) {
	case uint:
		return v, true
	case uint64:
		return uint(v), true
	case float64:
		return uint(v), true
	default:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return 0, false
	}
}

// parseIDParam parses the ":id" URL param as a uint.
func parseIDParam(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return 0, false
	}
	return uint(id), true
}

// fetchOrNotFound runs fetch(arg) and responds 404 with notFoundMsg on error.
func fetchOrNotFound[T any, A any](c *gin.Context, fetch func(A) (T, error), arg A, notFoundMsg string) (T, bool) {
	value, err := fetch(arg)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": notFoundMsg})
		return value, false
	}
	return value, true
}

// fetchOrServerError runs fetch(arg) and responds 500 with err's message on error.
func fetchOrServerError[T any, A any](c *gin.Context, fetch func(A) (T, error), arg A) (T, bool) {
	value, err := fetch(arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return value, false
	}
	return value, true
}

// requireOwner responds 403 and returns false when ownerID does not match
// currentUserID.
func requireOwner(c *gin.Context, ownerID, currentUserID uint) bool {
	if ownerID != currentUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return false
	}
	return true
}
