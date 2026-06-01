package handlers

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/kevidn/be-sipa/config"
	"github.com/kevidn/be-sipa/models"
)

func GetAdminDashboardStats(c *fiber.Ctx) error {
	var totalJenisSurat int64
	var suratAktif int64
	var totalHariLibur int64
	var totalPengguna int64
	var totalRole int64

	// Count Total Jenis Surat
	config.DB.Model(&models.JenisSurat{}).Count(&totalJenisSurat)

	// Count Surat Aktif
	config.DB.Model(&models.JenisSurat{}).Where("status = ?", "Aktif").Count(&suratAktif)

	// Count Hari Libur
	config.DB.Model(&models.HariLibur{}).Count(&totalHariLibur)

	// Count Users and Distinct Roles
	config.DB.Model(&models.User{}).Count(&totalPengguna)
	config.DB.Model(&models.User{}).Distinct("role").Count(&totalRole)

	// Calculate Average SLA
	var jenisSurats []models.JenisSurat
	config.DB.Find(&jenisSurats)
	var totalSla float64 = 0
	var countSla int = 0
	for _, js := range jenisSurats {
		// Asumsi format SLA adalah angka, atau string seperti "1", "2 Hari Kerja"
		// Kita akan coba extract angka pertama
		parts := strings.Fields(js.SLA)
		if len(parts) > 0 {
			if val, err := strconv.ParseFloat(parts[0], 64); err == nil {
				totalSla += val
				countSla++
			}
		}
	}
	avgSla := 0.0
	if countSla > 0 {
		avgSla = totalSla / float64(countSla)
		// Round to 1 decimal place
		avgSla = math.Round(avgSla*10) / 10
	}

	// Get Recent Activities
	var recentActivities []models.ActivityLog
	config.DB.Order("created_at desc").Limit(4).Find(&recentActivities)

	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"total_jenis_surat": totalJenisSurat,
			"surat_aktif":       suratAktif,
			"total_hari_libur":  totalHariLibur,
			"avg_sla":           fmt.Sprintf("%.1f", avgSla),
			"total_pengguna":    totalPengguna,
			"total_role":        totalRole,
			"recent_activities": recentActivities,
		},
	})
}

func CreateHariLibur(c *fiber.Ctx) error {
	var input models.HariLibur
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Input tidak valid"})
	}

	if err := config.DB.Create(&input).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menambah hari libur"})
	}

	// Record activity
	RecordLog(c.Locals("id_user").(string), "Manajemen Hari Libur", "Menambah hari libur: "+input.NamaLibur, "Berhasil", c.IP(), "")

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "success", "data": input})
}

func GetAllHariLibur(c *fiber.Ctx) error {
	var hariLibur []models.HariLibur
	config.DB.Order("tanggal asc").Find(&hariLibur)
	return c.JSON(fiber.Map{"status": "success", "data": hariLibur})
}

func DeleteHariLibur(c *fiber.Ctx) error {
	id := c.Params("id")
	var existing models.HariLibur
	if err := config.DB.First(&existing, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Hari libur tidak ditemukan"})
	}

	config.DB.Delete(&existing)
	RecordLog(c.Locals("id_user").(string), "Manajemen Hari Libur", "Menghapus hari libur: "+existing.NamaLibur, "Berhasil", c.IP(), "")

	return c.JSON(fiber.Map{"status": "success", "message": "Berhasil dihapus"})
}

// ========================
// MANAJEMEN PENGGUNA
// ========================

func GetAllUsers(c *fiber.Ctx) error {
	search := c.Query("search")
	roleFilter := c.Query("role")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	var users []models.User
	query := config.DB.Model(&models.User{})

	if search != "" {
		query = query.Where("nama_lengkap ILIKE ? OR nim ILIKE ? OR email ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	if roleFilter != "" && roleFilter != "Semua" {
		query = query.Where("role = ?", roleFilter)
	}

	var total int64
	query.Count(&total)

	offset := (page - 1) * limit
	query.Order("created_at desc").Limit(limit).Offset(offset).Find(&users)

	lastPage := math.Ceil(float64(total) / float64(limit))
	if lastPage == 0 {
		lastPage = 1
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   users,
		"meta": fiber.Map{
			"total":        total,
			"current_page": page,
			"last_page":    lastPage,
			"limit":        limit,
		},
	})
}

func CreateUser(c *fiber.Ctx) error {
	var input models.User
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Input tidak valid"})
	}

	// Default password setup is handled properly in a real app. For this prototype, we'll assign a default hash or let it be.
	// Normally we'd hash the password here if provided. For simplicity, assume frontend doesn't send password and we set a default.
	if input.PasswordHash == "" {
		// "$2a$10$wB.Q/X3hWj3D/4jYpQvHquWb9Y1I2O9o0WzQ.0wYpQvHquWb9Y1I2" is bcrypt for "password"
		input.PasswordHash = "$2a$10$wB.Q/X3hWj3D/4jYpQvHquWb9Y1I2O9o0WzQ.0wYpQvHquWb9Y1I2"
	}
	if input.StatusAkun == "" {
		input.StatusAkun = "Aktif"
	}

	if err := config.DB.Create(&input).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menambah pengguna"})
	}

	RecordLog(c.Locals("id_user").(string), "Manajemen Pengguna", "Menambah pengguna: "+input.NamaLengkap, "Berhasil", c.IP(), "")
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "success", "data": input})
}

func UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	var existing models.User
	if err := config.DB.Where("id_user = ?", id).First(&existing).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Pengguna tidak ditemukan"})
	}

	var input models.User
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Input tidak valid"})
	}

	// Hanya update field yang diizinkan (role, status)
	config.DB.Model(&existing).Select("Role", "StatusAkun", "NamaLengkap", "Email", "NIM").Updates(input)

	RecordLog(c.Locals("id_user").(string), "Manajemen Pengguna", "Memperbarui pengguna: "+existing.NamaLengkap, "Berhasil", c.IP(), "")
	return c.JSON(fiber.Map{"status": "success", "data": existing})
}

func DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	var existing models.User
	if err := config.DB.Where("id_user = ?", id).First(&existing).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Pengguna tidak ditemukan"})
	}

	config.DB.Delete(&existing)
	RecordLog(c.Locals("id_user").(string), "Manajemen Pengguna", "Menghapus pengguna: "+existing.NamaLengkap, "Berhasil", c.IP(), "")

	return c.JSON(fiber.Map{"status": "success", "message": "Berhasil dihapus"})
}
