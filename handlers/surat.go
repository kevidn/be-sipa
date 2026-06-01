package handlers

import (
	"fmt"
	"strconv"
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

func generateNomorSurat(jenisSurat models.JenisSurat) string {
	sifat := jenisSurat.KodeSifat
	if sifat == "" {
		sifat = "B" // Default Biasa
	}

	klasifikasi := jenisSurat.KodeKlasifikasi
	if klasifikasi == "" {
		klasifikasi = "KM" // Default Kemahasiswaan
	}

	unit := "UN38.9" // Statis untuk Fakultas Teknik UNESA
	year := time.Now().Format("2006")

	var count int64
	// Format surat: [Sifat]/[Urut]/[Unit]/[Klasifikasi]/[Tahun]
	// Cari urutan berdasarkan tahun berjalan, unit, dan klasifikasi
	config.DB.Model(&models.Surat{}).Where("nomor_surat LIKE ?", "%/"+unit+"/"+klasifikasi+"/"+year).Count(&count)

	return fmt.Sprintf("%s/%03d/%s/%s/%s", sifat, count+1, unit, klasifikasi, year)
}

func generateKitirNumber(jenisSurat string) string {
	var prefix string
	switch jenisSurat {
	case "Surat Keterangan Tidak Menerima Beasiswa":
		prefix = "SKTB"
	case "Surat Keterangan Kelakuan Baik":
		prefix = "SKKB"
	case "Surat Ijin Survei Penelitian (Skripsi)":
		prefix = "SKRIPSI"
	case "Surat Rekomendasi Beasiswa":
		prefix = "BEA"
	case "Surat Keterangan Masih Kuliah":
		prefix = "SKM"
	default:
		prefix = "SIPA"
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
	
	nomorSurat := generateKitirNumber(input.JenisSurat)

	newSurat := models.Surat{
		UserID:      userID,
		JenisSurat:  input.JenisSurat,
		Keperluan:   input.Keperluan,
		Semester:    input.Semester,
		FileUrl:     input.FileUrl,
		Status:      "Diajukan",
		DeadlineSLA: &deadline,
		SlaStatus:   "Aman",
		NomorSurat:  nomorSurat,
	}

	if err := config.DB.Create(&newSurat).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyimpan pengajuan"})
	}

	// Kirim email konfirmasi (REQ-FR019) dengan lampiran PDF Kitir
	var user models.User
	config.DB.Where("id_user = ?", userID).First(&user)

	kitirData := utils.KitirData{
		NomorSurat:  newSurat.NomorSurat,
		NamaLengkap: user.NamaLengkap,
		NIM:         user.NIM,
		JenisSurat:  newSurat.JenisSurat,
		Tanggal:     time.Now(),
	}

	pdfBytes, errPDF := utils.GenerateKitirPDF(kitirData)
	if errPDF == nil {
		utils.SendEmailWithKitir(user.Email, utils.MailData{
			NamaLengkap: user.NamaLengkap,
			NomorSurat:  newSurat.NomorSurat,
			JenisSurat:  newSurat.JenisSurat,
			Status:      "Diajukan",
		}, pdfBytes)
	} else {
		// Fallback ke email tanpa lampiran
		utils.SendStatusUpdateEmail(user.Email, utils.MailData{
			NamaLengkap: user.NamaLengkap,
			NomorSurat:  newSurat.NomorSurat,
			JenisSurat:  newSurat.JenisSurat,
			Status:      "Diajukan",
		})
	}

	CreateNotification(userID, "Surat Diajukan", fmt.Sprintf("%s %s telah berhasil diajukan dan menunggu verifikasi Tendik", newSurat.JenisSurat, newSurat.NomorSurat), "Process", "/pengajuan")

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

	query := config.DB.Model(&models.Surat{})
	// Jika mahasiswa, hanya lihat miliknya sendiri (BR-006)
	if strings.ToLower(role) == "mahasiswa" {
		query = query.Where("user_id = ?", userID)
	}

	search := c.Query("search", "")
	jenis := c.Query("jenis", "")
	status := c.Query("status", "")

	if search != "" {
		query = query.Where("nomor_surat ILIKE ?", "%"+search+"%")
	}
	if jenis != "" && jenis != "Semua Jenis Surat" {
		query = query.Where("jenis_surat = ?", jenis)
	}
	if status != "" && status != "Semua Status" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	pageStr := c.Query("page", "1")
	limitStr := c.Query("limit", "10")
	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	if page < 1 { page = 1 }
	if limit < 1 { limit = 10 }
	offset := (page - 1) * limit

	query = query.Preload("User").Preload("Processor")
	if sortBy == "mahasiswa" {
		query = query.Joins("User").Order("User.nama_lengkap " + order)
	} else {
		query = query.Order(dbSortCol + " " + order)
	}

	if err := query.Offset(offset).Limit(limit).Find(&surat).Error; err != nil {
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

	lastPage := 1
	if limit > 0 {
		lastPage = int(total) / limit
		if int(total)%limit > 0 {
			lastPage++
		}
	}

	return c.JSON(fiber.Map{
		"status":    "success",
		"data":      surat,
		"total":     total,
		"page":      page,
		"last_page": lastPage,
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
		Status  string `json:"status"`
		Catatan string `json:"catatan"`
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

	// BR-008: Pengajuan berstatus final (Selesai/Ditolak) tidak dapat diubah lagi
	if surat.Status == "Selesai" || surat.Status == "Ditolak" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Pengajuan dengan status final tidak dapat diubah lagi"})
	}

	// State Machine Verification
	isValidTransition := false
	switch surat.Status {
	case "Diajukan":
		if input.Status == "Diterima Tendik" || input.Status == "Ditolak" {
			isValidTransition = true
		}
	case "Diterima Tendik":
		if input.Status == "Diproses" {
			isValidTransition = true
		}
	case "Diproses", "SLA Terlampaui":
		if input.Status == "Selesai" || input.Status == "Ditolak" {
			isValidTransition = true
		}
	}

	if !isValidTransition {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("Transisi status dari '%s' ke '%s' tidak valid", surat.Status, input.Status)})
	}

	// Update Status, Komentar, & Processor
	processorID := c.Locals("id_user").(string)
	updates := map[string]interface{}{
		"status":       input.Status,
		"komentar":     input.Catatan,
		"processor_id": processorID,
	}

	if input.Status == "Selesai" {
		now := time.Now()
		updates["tanggal_selesai"] = &now
	}

	// Generate Nomor Surat (Penomoran Surat Flowchart)
	if surat.NomorSurat == "" && input.Status != "Ditolak" && input.Status != "Diajukan" {
		var js models.JenisSurat
		if err := config.DB.Where("nama = ?", surat.JenisSurat).First(&js).Error; err == nil {
			newNomor := generateNomorSurat(js)
			updates["nomor_surat"] = newNomor
			surat.NomorSurat = newNomor // Update lokal untuk email notifikasi di bawah
		}
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

	var notifType string
	notifTitle := "Surat " + input.Status
	switch input.Status {
	case "Selesai":
		notifType = "Success"
	case "Ditolak":
		notifType = "Error"
	case "Diproses", "Diterima Tendik":
		notifType = "Process"
	default:
		notifType = "Info"
	}

	notifMsg := fmt.Sprintf("%s %s saat ini berstatus: %s", surat.JenisSurat, surat.NomorSurat, input.Status)
	if input.Status == "Ditolak" {
		notifMsg = fmt.Sprintf("%s %s ditolak. Alasan: %s", surat.JenisSurat, surat.NomorSurat, input.Catatan)
	}

	CreateNotification(surat.UserID, notifTitle, notifMsg, notifType, fmt.Sprintf("/pengajuan/%d", surat.ID))

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Status pengajuan berhasil diperbarui",
		"data":    surat,
	})
}

func GetDashboardStats(c *fiber.Ctx) error {
	var totalAntrian, sedangDiproses, totalSelesai, prioritasTinggi, dokumenLengkap int64

	config.DB.Model(&models.Surat{}).Where("status = ?", "Diajukan").Count(&totalAntrian)
	config.DB.Model(&models.Surat{}).Where("status IN ?", []string{"Diterima Tendik", "Diproses"}).Count(&sedangDiproses)

	config.DB.Model(&models.Surat{}).Where("status = ?", "Selesai").Count(&totalSelesai)

	var slaTerlampauiList []models.Surat
	config.DB.Preload("User").Where("sla_status = ? AND status NOT IN ('Selesai', 'Ditolak')", "Terlampaui").Find(&slaTerlampauiList)

	config.DB.Model(&models.Surat{}).Where("prioritas = ? AND status NOT IN ('Selesai', 'Ditolak')", "Tinggi").Count(&prioritasTinggi)
	config.DB.Model(&models.Surat{}).Where("is_document_complete = ? AND status NOT IN ('Selesai', 'Ditolak')", true).Count(&dokumenLengkap)

	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"total_antrian":       totalAntrian,
			"sedang_diproses":     sedangDiproses,
			"total_selesai":       totalSelesai,
			"sla_terlampaui":      int64(len(slaTerlampauiList)),
			"sla_terlampaui_list": slaTerlampauiList,
			"prioritas_tinggi":    prioritasTinggi,
			"dokumen_lengkap":     dokumenLengkap,
		},
	})
}
