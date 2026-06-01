package models

import (
	"time"
)

type JenisSurat struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Kode            string    `gorm:"type:varchar(20);uniqueIndex;not null" json:"kode"`
	KodeSifat       string    `gorm:"type:varchar(20)" json:"kode_sifat"`
	KodeKlasifikasi string    `gorm:"type:varchar(20)" json:"kode_klasifikasi"`
	Nama            string    `gorm:"type:varchar(100);not null" json:"nama"`
	SLA             string    `gorm:"type:varchar(50)" json:"sla"` // e.g. 1-2 hari kerja
	TemplateFile    string    `gorm:"type:varchar(255)" json:"template_file"`
	Status          string    `gorm:"type:varchar(20);default:'Aktif'" json:"status"` // Aktif, Non-Aktif
	Persyaratan     string    `gorm:"type:text" json:"persyaratan"` // JSON or comma-separated: KTM, Transkrip, etc.
	TotalPengajuan  int       `gorm:"default:0" json:"total_pengajuan"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (JenisSurat) TableName() string {
	return "jenis_surat"
}
