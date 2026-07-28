package model

import (
	"time"

	"gorm.io/gorm"
)

type Media struct {
	gorm.Model
	Title     string
	Type      string
	Filename  string
	Thumbnail string
	MimeType  string
	OwnerID   uint
	CreatedAt time.Time
}
