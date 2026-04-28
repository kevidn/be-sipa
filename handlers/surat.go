package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kevidn/be-sipa/config"
	"github.com/kevidn/be-sipa/models"
)

type SuratInput struct {
	JenisSurat string `json:"jenis_surat"`
	Keperluan  string `json:"keperluan"`
	Semester   string `json:"semester"`
	FileUrl    string `json:"file_url"`
}

func generateNomorSurat(jenis string) string {
	var prefix string
	switch jenis {
	case "Surat Keterangan Masih Kuliah":
		prefix = "SKM"
	case "Surat Ijin Survei Penelitian (Skripsi)":
		prefix = "SKRIPSI"
	case "Surat Rekomendasi Beasiswa":
		prefix = "BEA"
	default:
		prefix = "SRT"
	}

	year := time.Now().Format("2006")
	
	var count int64
	config.DB.Model(&models.Surat{}).Where("nomor_surat LIKE ?", prefix+"-"+year+"-%").Count(&count)
	
	return fmt.Sprintf("%s-%s-%03d", prefix, year, count+1)
}

func SubmitSurat(c *fiber.Ctx) error {
	var input SuratInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Input tidak valid"})
	}

	userID := c.Locals("id_user").(string)

	newSurat := models.Surat{
		UserID:     userID,
		NomorSurat: generateNomorSurat(input.JenisSurat),
		JenisSurat: input.JenisSurat,
		Keperluan:  input.Keperluan,
		Semester:   input.Semester,
		FileUrl:    input.FileUrl,
		Status:     "Diproses",
	}

	if err := config.DB.Create(&newSurat).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyimpan pengajuan"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "Pengajuan berhasil dikirim",
		"data":    newSurat,
	})
}

func GetHistorySurat(c *fiber.Ctx) error {
	userID := c.Locals("id_user").(string)
	role := c.Locals("role").(string)

	var surat []models.Surat
	query := config.DB.Order("created_at DESC")

	// Jika mahasiswa, hanya lihat miliknya sendiri
	if strings.ToLower(role) == "mahasiswa" {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Find(&surat).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data riwayat"})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   surat,
	})
}
