package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kevidn/be-sipa/config"
	"github.com/kevidn/be-sipa/models"
	"github.com/kevidn/be-sipa/utils"
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
	case "Surat Tunjangan/Pensiun/Akses":
		prefix = "TPA"
	case "Surat Keterangan Tidak Menerima Beasiswa":
		prefix = "SKTB"
	case "Surat Rekomendasi Beasiswa":
		prefix = "BEA"
	case "Surat Keterangan Kelakuan Baik":
		prefix = "SKKB"
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

	// BR-010: Cegah mahasiswa mengajukan surat yang sama jika masih ada pengajuan aktif
	var activeSurat models.Surat
	err := config.DB.Where("user_id = ? AND jenis_surat = ? AND status NOT IN ('Selesai', 'Ditolak')", userID, input.JenisSurat).First(&activeSurat).Error
	if err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Anda masih memiliki pengajuan aktif untuk jenis surat ini."})
	}

	// Hitung Deadline SLA (BR-002, BR-003, BR-004)
	slaDays := utils.GetSLADays(input.JenisSurat)
	deadline := utils.CalculateDeadline(time.Now(), slaDays)

	newSurat := models.Surat{
		UserID:      userID,
		NomorSurat:  generateNomorSurat(input.JenisSurat),
		JenisSurat:  input.JenisSurat,
		Keperluan:   input.Keperluan,
		Semester:    input.Semester,
		FileUrl:     input.FileUrl,
		Status:      "Diajukan",
		DeadlineSLA: &deadline,
		SlaStatus:   "Aman",
	}

	if err := config.DB.Create(&newSurat).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyimpan pengajuan"})
	}

	// Kirim email konfirmasi (REQ-FR019)
	var user models.User
	config.DB.Where("id_user = ?", userID).First(&user)
	
	utils.SendStatusUpdateEmail(user.Email, utils.MailData{
		NamaLengkap: user.NamaLengkap,
		NomorSurat:  newSurat.NomorSurat,
		JenisSurat:  newSurat.JenisSurat,
		Status:      "Diajukan",
	})

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
	
	sortBy := c.Query("sort", "created_at")
	order := c.Query("order", "desc")

	if strings.ToLower(order) != "asc" {
		order = "desc"
	}

	dbSortCol := "surat_pengajuan.created_at"
	switch sortBy {
	case "nomor_surat":
		dbSortCol = "surat_pengajuan.nomor_surat"
	case "jenis_surat":
		dbSortCol = "surat_pengajuan.jenis_surat"
	case "status":
		dbSortCol = "surat_pengajuan.status"
	case "tanggal_masuk":
		dbSortCol = "surat_pengajuan.created_at"
	case "sla":
		dbSortCol = "surat_pengajuan.deadline_sla"
	case "prioritas":
		dbSortCol = "surat_pengajuan.prioritas"
	}

	query := config.DB.Preload("User").Preload("Processor")
	if sortBy == "mahasiswa" {
		query = query.Joins("User").Order("User.nama_lengkap " + order)
	} else {
		query = query.Order(dbSortCol + " " + order)
	}
	// Jika mahasiswa, hanya lihat miliknya sendiri (BR-006)
	if strings.ToLower(role) == "mahasiswa" {
		query = query.Where("user_id = ?", userID)
	}


	if err := query.Find(&surat).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data riwayat"})
	}

	// Update SLA status on the fly for non-completed letters
	now := time.Now()
	for i := range surat {
		if surat[i].Status != "Selesai" && surat[i].Status != "Ditolak" && surat[i].DeadlineSLA != nil {
			var newSlaStatus string
			if now.After(*surat[i].DeadlineSLA) {
				newSlaStatus = "Terlampaui"
			} else if now.After(surat[i].DeadlineSLA.Add(-24 * time.Hour)) {
				newSlaStatus = "Mendekati"
			} else {
				newSlaStatus = "Aman"
			}

			if surat[i].SlaStatus != newSlaStatus {
				surat[i].SlaStatus = newSlaStatus
				config.DB.Model(&surat[i]).Update("sla_status", newSlaStatus)
			}
		}
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   surat,
	})
}

func GetDetailSurat(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("id_user").(string)
	role := c.Locals("role").(string)

	var surat models.Surat
	query := config.DB.Preload("User").Preload("Processor")

	if err := query.First(&surat, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Pengajuan tidak ditemukan"})
	}

	// Akses kontrol: Mahasiswa hanya bisa lihat surat miliknya sendiri
	if strings.ToLower(role) == "mahasiswa" && surat.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Anda tidak memiliki akses ke data ini"})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   surat,
	})
}

func UpdateStatusSurat(c *fiber.Ctx) error {
	id := c.Params("id")
	role := c.Locals("role").(string)

	// BR-005 & BR-007: Hanya Tendik yang dapat mengubah status. Kaprodi tidak bisa.
	if strings.ToLower(role) != "tendik" && strings.ToLower(role) != "admin sistem" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Hanya Tendik atau Admin yang dapat mengubah status pengajuan"})
	}

	type UpdateInput struct {
		Status   string `json:"status"`
		Catatan  string `json:"catatan"`
	}

	var input UpdateInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Input tidak valid"})
	}

	// BR-008: Status 'Ditolak' WAJIB disertai catatan
	if input.Status == "Ditolak" && strings.TrimSpace(input.Catatan) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Status 'Ditolak' wajib menyertakan alasan penolakan"})
	}

	var surat models.Surat
	if err := config.DB.Preload("User").First(&surat, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Pengajuan tidak ditemukan"})
	}

	// BR-009: Pengajuan berstatus 'Selesai' tidak dapat diubah statusnya kembali
	if surat.Status == "Selesai" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Pengajuan yang sudah Selesai tidak dapat diubah lagi"})
	}

	// Update Status, Komentar, & Processor
	processorID := c.Locals("id_user").(string)
	updates := map[string]interface{}{
		"status":       input.Status,
		"komentar":     input.Catatan,
		"processor_id": processorID,
	}

	if err := config.DB.Model(&surat).Updates(updates).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal memperbarui status"})
	}

	// Record Activity Log
	keterangan := fmt.Sprintf("Memperbarui status pengajuan %s menjadi %s", surat.JenisSurat, input.Status)
	if input.Status == "Ditolak" {
		keterangan += fmt.Sprintf(" dengan alasan: %s", input.Catatan)
	}
	RecordLog(processorID, "Update Status", keterangan, "Berhasil", c.IP(), surat.NomorSurat)

	// Kirim Email Notifikasi (REQ-FR020)
	utils.SendStatusUpdateEmail(surat.User.Email, utils.MailData{
		NamaLengkap: surat.User.NamaLengkap,
		NomorSurat:  surat.NomorSurat,
		JenisSurat:  surat.JenisSurat,
		Status:      input.Status,
		Catatan:     input.Catatan,
	})

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Status pengajuan berhasil diperbarui",
		"data":    surat,
	})
}

func GetDashboardStats(c *fiber.Ctx) error {
	var totalAntrian, sedangDiproses, totalSelesai, slaTerlampaui, prioritasTinggi, dokumenLengkap int64

	config.DB.Model(&models.Surat{}).Where("status = ?", "Diajukan").Count(&totalAntrian)
	config.DB.Model(&models.Surat{}).Where("status IN ?", []string{"Diterima Tendik", "Diproses"}).Count(&sedangDiproses)
	
	config.DB.Model(&models.Surat{}).Where("status = ?", "Selesai").Count(&totalSelesai)
	
	config.DB.Model(&models.Surat{}).Where("sla_status = ? AND status NOT IN ('Selesai', 'Ditolak')", "Terlampaui").Count(&slaTerlampaui)

	config.DB.Model(&models.Surat{}).Where("prioritas = ? AND status NOT IN ('Selesai', 'Ditolak')", "Tinggi").Count(&prioritasTinggi)
	config.DB.Model(&models.Surat{}).Where("is_document_complete = ? AND status NOT IN ('Selesai', 'Ditolak')", true).Count(&dokumenLengkap)

	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"total_antrian":    totalAntrian,
			"sedang_diproses":  sedangDiproses,
			"total_selesai":    totalSelesai,
			"sla_terlampaui":   slaTerlampaui,
			"prioritas_tinggi": prioritasTinggi,
			"dokumen_lengkap":  dokumenLengkap,
		},
	})
}



