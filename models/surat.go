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
	Status      string    `gorm:"type:varchar(20);not null;default:'Diajukan'" json:"status"` // Diajukan, Diterima Tendik, Diproses, Selesai, Ditolak
	FileUrl     string    `gorm:"type:text" json:"file_url"`
	Komentar    string    `gorm:"type:text" json:"komentar"`
	DeadlineSLA        *time.Time `gorm:"type:timestamp null" json:"deadline_sla"`
	SlaStatus          string     `gorm:"type:varchar(20);default:'Aman'" json:"sla_status"` // Aman, Mendekati, Terlampaui
	Prioritas          string     `gorm:"type:varchar(20);default:'Normal'" json:"prioritas"` // Normal, Sedang, Tinggi
	IsDocumentComplete bool       `gorm:"default:true" json:"is_document_complete"`
	ProcessorID        *string    `gorm:"type:varchar(10)" json:"id_processor"`
	Processor          *User      `gorm:"foreignKey:ProcessorID;references:IDUser" json:"processor"`
	CreatedAt          time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Surat) TableName() string {
	return "surat_pengajuan"
}
