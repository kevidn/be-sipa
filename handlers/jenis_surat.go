package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kevidn/be-sipa/config"
	"github.com/kevidn/be-sipa/models"
)

func GetAllJenisSurat(c *fiber.Ctx) error {
	var list []models.JenisSurat
	// Subquery to get count from surat_pengajuan table
	config.DB.Select("jenis_surat.*, (SELECT COUNT(*) FROM surat_pengajuan WHERE surat_pengajuan.jenis_surat = jenis_surat.nama) as total_pengajuan").
		Order("id ASC").Find(&list)
	return c.JSON(fiber.Map{"status": "success", "data": list})
}

func CreateJenisSurat(c *fiber.Ctx) error {
	var input models.JenisSurat
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Input tidak valid"})
	}

	if err := config.DB.Create(&input).Error; err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Kode surat sudah ada"})
	}

	RecordLog(c.Locals("id_user").(string), "Manajemen Jenis Surat", "Menambah jenis surat: "+input.Nama, "Berhasil", c.IP(), input.Kode)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "success", "data": input})
}

func UpdateJenisSurat(c *fiber.Ctx) error {
	id := c.Params("id")
	var input models.JenisSurat
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Input tidak valid"})
	}

	var existing models.JenisSurat
	if err := config.DB.First(&existing, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Data tidak ditemukan"})
	}

	config.DB.Model(&existing).Updates(input)

	RecordLog(c.Locals("id_user").(string), "Manajemen Jenis Surat", "Memperbarui jenis surat: "+existing.Nama, "Berhasil", c.IP(), existing.Kode)

	return c.JSON(fiber.Map{"status": "success", "data": existing})
}

func DeleteJenisSurat(c *fiber.Ctx) error {
	id := c.Params("id")
	var existing models.JenisSurat
	if err := config.DB.First(&existing, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Data tidak ditemukan"})
	}

	config.DB.Delete(&existing)

	RecordLog(c.Locals("id_user").(string), "Manajemen Jenis Surat", "Menghapus jenis surat: "+existing.Nama, "Berhasil", c.IP(), existing.Kode)

	return c.JSON(fiber.Map{"status": "success", "message": "Berhasil dihapus"})
}
