package main

import (
	"embed"
	"io/fs"
	"net/http"
	"path/filepath"
	"time"

	"wireless_drive/internal/config"
	"wireless_drive/internal/handler"
	"wireless_drive/internal/model"
	"wireless_drive/internal/repository"
	service "wireless_drive/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

//
// EMBED OF FRONTEND
//

//go:embed static/*
var staticFiles embed.FS

func main() {
	// =========================
	// INITIAL CONFIGURATION
	// =========================
	config.LoadEnv()
	config.ConnectDB()

	config.DB.AutoMigrate(&model.User{}, &model.Media{})

	// =========================
	// GIN
	// =========================
	r := gin.Default()

	// CORS PRIMEIRO e MÁXIMO PERMISSIVO (antes de qualquer rota ou middleware)
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Requested-With"}
	corsConfig.ExposeHeaders = []string{"Content-Length"}
	corsConfig.AllowCredentials = false // false com AllowAllOrigins
	corsConfig.MaxAge = 12 * time.Hour
	r.Use(cors.New(corsConfig))

	// =========================
	// FRONTEND (SPA)
	// =========================

	// Remove the prefix "static" of embed
	subFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic("erro ao carregar arquivos estáticos")
	}

	// Static Files (JS, CSS, assets...)
	r.StaticFS("/static", http.FS(subFS))

	// Serve thumbnails from disk - considera BASE_PATH
	basePath := config.GetEnv("BASE_PATH", ".")
	thumbsDir := config.GetEnv("THUMBS_DIR", "thumbs")
	thumbsPath := filepath.Join(basePath, thumbsDir)

	r.Static("/thumbs", thumbsPath)

	// =========================
	// REPOSITORIES / SERVICES
	// =========================
	userRepo := repository.NewUserRepository(config.DB)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	mediaRepo := repository.NewMediaRepository(config.DB)
	mediaService := service.NewMediaService(mediaRepo)
	mediaHandler := handler.NewMediaHandler(mediaService)

	// =========================
	// API ENDPOINTS (registered before page routes)
	// =========================

	userHandler.RegisterRoutes(r)
	mediaHandler.RegisterRoutes(r)

	// =========================
	// PAGE ROUTES (after API to avoid intercepting API calls)
	// =========================

	// Serve pages - using /website/ prefix to avoid conflicts with API routes
	r.GET("/website/account", func(c *gin.Context) {
		c.FileFromFS("account.html", http.FS(subFS))
	})

	r.GET("/website/dashboard", func(c *gin.Context) {
		c.FileFromFS("dashboard.html", http.FS(subFS))
	})

	r.GET("/website/upload", func(c *gin.Context) {
		c.FileFromFS("upload.html", http.FS(subFS))
	})

	r.GET("/website/media", func(c *gin.Context) {
		c.FileFromFS("media.html", http.FS(subFS))
	})

	// Redirect / to /website/account
	r.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/website/account")
	})

	// =========================
	// START SERVER
	// =========================
	r.Run(":8085")
}
