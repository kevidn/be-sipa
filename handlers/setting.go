package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kevidn/be-sipa/config"
	"github.com/kevidn/be-sipa/models"
)

func GetSystemSettings(c *fiber.Ctx) error {
	var setting models.SystemSetting
	// Always use ID 1 for global settings
	if err := config.DB.First(&setting, 1).Error; err != nil {
		// If not exists, create default
		setting = models.SystemSetting{ID: 1}
		config.DB.Create(&setting)
	}
	return c.JSON(fiber.Map{"status": "success", "data": setting})
}

func UpdateSystemSettings(c *fiber.Ctx) error {
	var input models.SystemSetting
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Input tidak valid"})
	}

	var setting models.SystemSetting
	config.DB.First(&setting, 1)
	
	config.DB.Model(&setting).Updates(input)

	RecordLog(c.Locals("id_user").(string), "Pengaturan Sistem", "Memperbarui konfigurasi sistem", "Berhasil", c.IP(), "")

	return c.JSON(fiber.Map{"status": "success", "data": setting})
}
