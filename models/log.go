package models

import (
	"time"
)

type ActivityLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      string    `gorm:"type:varchar(10);not null" json:"id_user"`
	User        User      `gorm:"foreignKey:UserID;references:IDUser" json:"user"`
	Aksi        string    `gorm:"type:varchar(100);not null" json:"aksi"` // Menyetujui Pengajuan, Login, dll
	Keterangan  string    `gorm:"type:text" json:"keterangan"`
	Status      string    `gorm:"type:varchar(20);default:'Berhasil'" json:"status"` // Berhasil, Gagal
	IPAddress   string    `gorm:"type:varchar(50)" json:"ip_address"`
	ReferansiID string    `gorm:"type:varchar(50)" json:"referensi_id"` // Nomor Surat atau ID lainnya
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (ActivityLog) TableName() string {
	return "activity_logs"
}
