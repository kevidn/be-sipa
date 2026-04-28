package models

import (
	"time"
)

type User struct {
	IDUser       string     `gorm:"primaryKey;type:varchar(10)" json:"id_user"`
	Username     string     `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	PasswordHash string     `gorm:"type:char(60);not null" json:"-"` // json:"-" menyembunyikan password dari response API
	NamaLengkap  string     `gorm:"type:varchar(150);not null" json:"nama_lengkap"`
	Email                string     `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	ResetPasswordToken   string     `gorm:"type:varchar(255)" json:"-"`
	ResetPasswordExpires *time.Time `gorm:"type:timestamp null" json:"-"`
	Role         string     `gorm:"type:varchar(20);not null" json:"role"` // Enum: Mahasiswa, Dosen, Kaprodi, dll
	IDUnitKerja  string     `gorm:"type:varchar(10)" json:"id_unit_kerja"`
	NIM          string     `gorm:"type:varchar(20)" json:"nim"`
	ProgramStudi string     `gorm:"type:varchar(100)" json:"program_studi"`
	StatusAkun   string     `gorm:"type:varchar(20);not null;default:'Aktif'" json:"status_akun"` // Enum: Aktif, Nonaktif, dll
	LastLogin    *time.Time `gorm:"type:timestamp null" json:"last_login"`                        // Pakai pointer agar bisa menerima nilai NULL
	CreatedAt    time.Time  `gorm:"autoCreateTime;type:timestamp;not null" json:"created_at"`
	IsSlaMonitor bool       `gorm:"not null;default:false" json:"is_sla_monitor"`
	PhoneNumber  string     `gorm:"type:varchar(20)" json:"phone_number"`
}

func (User) TableName() string {
	return "users"
}
