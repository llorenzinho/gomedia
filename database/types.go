package database

import (
	"time"
)

type Media struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	Filename  string `gorm:"index"`
	Size      int64
	BasePath  *string
	Check     bool
}
