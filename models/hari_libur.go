package models

import (
	"time"
)

type HariLibur struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Tanggal   time.Time `gorm:"type:date;not null" json:"tanggal"`
	NamaLibur string    `gorm:"type:varchar(150);not null" json:"nama_libur"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (HariLibur) TableName() string {
	return "hari_libur"
}
