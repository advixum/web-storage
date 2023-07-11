package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
	Files    []File
}

type File struct {
	gorm.Model
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null"`
	ListName  string    `gorm:"not null"`
	Name      string    `gorm:"unique;not null"`
	Extension string    `gorm:"not null"`
	Path      string    `gorm:"unique;not null"`
	Date      time.Time `gorm:"not null"`
	Size      int64     `gorm:"not null"`
}
