package models

import (
	"time"
)

type Surat struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      string    `gorm:"type:varchar(10);not null" json:"id_user"`
	User        User      `gorm:"foreignKey:UserID;references:IDUser" json:"user"`
	NomorSurat  string    `gorm:"type:varchar(50);uniqueIndex" json:"nomor_surat"`
	JenisSurat  string    `gorm:"type:varchar(100);not null" json:"jenis_surat"`
	Keperluan   string    `gorm:"type:text" json:"keperluan"`
	Semester    string    `gorm:"type:varchar(20)" json:"semester"`
	Status      string    `gorm:"type:varchar(20);not null;default:'Diproses'" json:"status"` // Diproses, Diterima Tendik, Selesai, Ditolak
	FileUrl     string    `gorm:"type:varchar(255)" json:"file_url"`
	Komentar    string    `gorm:"type:text" json:"komentar"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Surat) TableName() string {
	return "surat_pengajuan"
}
