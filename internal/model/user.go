package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name     string `json:"name"`
	Code     string `json:"code" gorm:"unique"`
	Password string `json:"password"`
}
