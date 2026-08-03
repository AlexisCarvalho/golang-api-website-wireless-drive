package handler

import (
	"net/http"
	"wireless_drive/internal/auth"
	"wireless_drive/internal/model"

	"wireless_drive/internal/middleware"

	"github.com/gin-gonic/gin"

	service "wireless_drive/internal/service"
)

type UserHandler struct {
	service service.UserService
}

func NewUserHandler(s service.UserService) *UserHandler {
	return &UserHandler{service: s}
}

func (h *UserHandler) RegisterRoutes(r *gin.Engine) {
	userRoutes := r.Group("/api/users")
	{
		userRoutes.POST("/register", h.RegisterUser)
		userRoutes.POST("/login", h.Login)
		userRoutes.GET("/verify-authenticated", middleware.AuthMiddleware(), h.VerifyAuthenticated)
	}

}

func (h *UserHandler) VerifyAuthenticated(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	c.JSON(200, gin.H{
		"message": "User authenticated",
		"userID":  userID,
	})
}

func (h *UserHandler) RegisterUser(c *gin.Context) {
	var entry model.User
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.RegisterUser(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, entry)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req struct {
		Code     string `json:"code"`
		Password string `json:"password"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	user := h.service.Authenticate(req.Code, req.Password)
	if user == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := auth.GenerateJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
