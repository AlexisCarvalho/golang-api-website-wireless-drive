package config

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	dbPath := GetEnv("DB_PATH", "calculaPagamento.db")

	database, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		panic("Falha ao conectar ao banco de dados!")
	}

	DB = database
}
