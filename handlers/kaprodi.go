package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kevidn/be-sipa/config"
	"github.com/kevidn/be-sipa/models"
)

func GetKaprodiStats(c *fiber.Ctx) error {
	var total, selesai, dalamProses, ditolak, slaTerlampaui int64

	config.DB.Model(&models.Surat{}).Count(&total)
	config.DB.Model(&models.Surat{}).Where("status = ?", "Selesai").Count(&selesai)
	config.DB.Model(&models.Surat{}).Where("status IN ?", []string{"Diajukan", "Diterima Tendik", "Diproses"}).Count(&dalamProses)
	config.DB.Model(&models.Surat{}).Where("status = ?", "Ditolak").Count(&ditolak)
	config.DB.Model(&models.Surat{}).Where("sla_status = ?", "Terlampaui").Count(&slaTerlampaui)

	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"total_pengajuan": total,
			"selesai":         selesai,
			"dalam_proses":    dalamProses,
			"ditolak":         ditolak,
			"sla_terlampaui":  slaTerlampaui,
		},
	})
}
