package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kevidn/be-sipa/config"
	"github.com/kevidn/be-sipa/models"
)

func RecordLog(userID, aksi, keterangan, status, ip, refID string) {
	log := models.ActivityLog{
		UserID:      userID,
		Aksi:        aksi,
		Keterangan:  keterangan,
		Status:      status,
		IPAddress:   ip,
		ReferansiID: refID,
	}
	config.DB.Create(&log)
}

func GetActivityLogs(c *fiber.Ctx) error {
	var logs []models.ActivityLog
	config.DB.Order("created_at DESC").Preload("User").Find(&logs)

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   logs,
	})
}

func GetLogStats(c *fiber.Ctx) error {
	var total, hariIni, berhasil, gagal int64

	config.DB.Model(&models.ActivityLog{}).Count(&total)
	
	today := time.Now().Format("2006-01-02")
	config.DB.Model(&models.ActivityLog{}).Where("DATE(created_at) = ?", today).Count(&hariIni)
	
	config.DB.Model(&models.ActivityLog{}).Where("status = ?", "Berhasil").Count(&berhasil)
	config.DB.Model(&models.ActivityLog{}).Where("status = ?", "Gagal").Count(&gagal)

	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"total_log": total,
			"hari_ini":  hariIni,
			"berhasil":  berhasil,
			"gagal":     gagal,
		},
	})
}
