package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func init() {
	exePath, err := os.Executable()
	if err != nil {
		log.Println("Failed to determine executable path:", err)
		return
	}

	envPath := filepath.Join(filepath.Dir(exePath), ".env")

	if err := godotenv.Load(envPath); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
