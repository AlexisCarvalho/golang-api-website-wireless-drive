package handler

import (
	"net/http"
	"wireless_gallery/internal/auth"
	"wireless_gallery/internal/model"

	"wireless_gallery/internal/middleware"

	"github.com/gin-gonic/gin"

	service "wireless_gallery/internal/service"
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
		"message": "Usuário autenticado",
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}

	user := h.service.Authenticate(req.Code, req.Password)
	if user == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Credenciais inválidas"})
		return
	}

	token, err := auth.GenerateJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
