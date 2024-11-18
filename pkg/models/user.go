package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email    string `gorm:"uniqueIndex;size:255"`
	Password string `gorm:"size:255"`
}
