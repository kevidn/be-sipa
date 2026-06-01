package models

import (
	"time"
)

type SystemSetting struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	EmailNotification bool      `gorm:"default:true" json:"email_notification"`
	PushNotification  bool      `gorm:"default:true" json:"push_notification"`
	SLAPerformanceHrs int       `gorm:"default:6" json:"sla_performance_hrs"`
	BackupOtomatis    bool      `gorm:"default:true" json:"backup_otomatis"`
	BackupInterval    string    `gorm:"type:varchar(50);default:'Harian'" json:"backup_interval"`
	MaxFileUploadMB   int       `gorm:"default:5" json:"max_file_upload_mb"`
	SessionTimeoutMin int       `gorm:"default:30" json:"session_timeout_min"`
	SMTPServer        string    `gorm:"type:varchar(255);default:'smtp.unesa.ac.id'" json:"smtp_server"`
	SMTPPort          int       `gorm:"default:587" json:"smtp_port"`
	SMTPUsername      string    `gorm:"type:varchar(255)" json:"smtp_username"`
	SMTPPassword      string    `gorm:"type:varchar(255)" json:"smtp_password"`
	MaxLoginAttempts  int       `gorm:"default:5" json:"max_login_attempts"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (SystemSetting) TableName() string {
	return "system_settings"
}
