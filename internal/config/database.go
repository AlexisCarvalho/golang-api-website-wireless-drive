package config

import (
	"path/filepath"

	"gorm.io/driver/sqlite"

	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	// Use the BASE_PATH exactly as configured, with no "." fallback.
	// BASE_PATH must be explicitly set because it's used for more sensible things.
	// If you want to use the same folder the executable is as BASE_PATH to store the database and files
	// just put (BASE_PATH=.) on the .env
	basePath := GetEnv("BASE_PATH", "")

	if basePath == "" {
		panic("BASE_PATH is not configured")
	}

	dbName := GetEnv("DB_NAME", "wireless_drive.db")
	dbPath := filepath.Join(basePath, dbName)

	database, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		panic("Database connection failed!")
	}

	DB = database
}
