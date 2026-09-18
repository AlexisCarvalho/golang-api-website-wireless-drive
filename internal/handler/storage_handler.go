package handler

import (
	"net/http"
	"wireless_drive/internal/middleware"
	"wireless_drive/internal/utils"

	"github.com/gin-gonic/gin"
)

type StorageHandler struct{}

func NewStorageHandler() *StorageHandler {
	return &StorageHandler{}
}

func (h *StorageHandler) RegisterRoutes(r *gin.Engine) {
	r.GET("/api/storage/disk-usage", middleware.AuthMiddleware(), h.GetDataDiskUsage)
}

func (h *StorageHandler) GetDataDiskUsage(c *gin.Context) {
	usage, err := utils.GetDataDiskUsage()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, usage)
}
