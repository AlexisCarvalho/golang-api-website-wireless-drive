package config

import (
	"path/filepath"

	"gorm.io/driver/sqlite"

	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	basePath := GetEnv("BASE_PATH", ".")
	dbName := GetEnv("DB_NAME", "wirelessDrive.db")
	dbPath := filepath.Join(basePath, dbName)

	database, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		panic("Falha ao conectar ao banco de dados!")
	}

	DB = database
}
