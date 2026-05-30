package models

import (
	"time"
)

type Notification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"type:varchar(10);not null;index" json:"id_user"`
	User      User      `gorm:"foreignKey:UserID;references:IDUser" json:"user"`
	Title     string    `gorm:"type:varchar(150);not null" json:"title"`
	Message   string    `gorm:"type:text;not null" json:"message"`
	Type      string    `gorm:"type:varchar(20);default:'Info'" json:"type"` // Info, Success, Process, Error, Warning
	Link      string    `gorm:"type:varchar(255)" json:"link"` // Halaman yang akan dituju saat "Lihat Detail"
	IsRead    bool      `gorm:"default:false" json:"is_read"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (Notification) TableName() string {
	return "notifications"
}
