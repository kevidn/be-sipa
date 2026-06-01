package cron

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/kevidn/be-sipa/config"
	"github.com/kevidn/be-sipa/models"
	"github.com/kevidn/be-sipa/utils"
)

func InitScheduler() {
	go func() {
		for {
			CheckSLAAndSendWarnings()
			time.Sleep(1 * time.Hour)
		}
	}()

	go func() {
		for {
			RunDatabaseBackup()
			time.Sleep(24 * time.Hour)
		}
	}()
}

func CheckSLAAndSendWarnings() {
	var setting models.SystemSetting
	if err := config.DB.First(&setting, 1).Error; err != nil {
		log.Println("Gagal memuat pengaturan sistem untuk cron SLA")
		return
	}

	if !setting.EmailNotification {
		return // Notifikasi dimatikan
	}

	var suratList []models.Surat
	// Cari surat yang sedang diproses
	config.DB.Preload("Processor").Where("status = ?", "Diproses").Find(&suratList)

	now := time.Now()

	for _, surat := range suratList {
		if surat.DeadlineSLA == nil {
			continue
		}
		
		hoursUntilDeadline := surat.DeadlineSLA.Sub(now).Hours()
		
		if !surat.IsSLAWarningSent && hoursUntilDeadline > 0 && hoursUntilDeadline <= float64(setting.SLAPerformanceHrs) {
			// Sudah masuk zona kritis SLA, kirim email ke processor (kaprodi/tendik)
			if surat.Processor != nil && surat.Processor.Email != "" {
				utils.SendSLAWarningEmail(surat.Processor.Email, surat.Processor.NamaLengkap, surat.NomorSurat, int(hoursUntilDeadline))
				log.Printf("Mengirim peringatan SLA ke %s untuk Surat %s", surat.Processor.Email, surat.NomorSurat)
				
				config.DB.Model(&surat).Updates(map[string]interface{}{
					"is_sla_warning_sent": true,
					"sla_status":          "Mendekati",
				})
			}
		} else if hoursUntilDeadline <= 0 && surat.SlaStatus != "Terlampaui" {
			// Sudah terlewat, eskalasi ke Kaprodi
			var kaprodi models.User
			config.DB.Where("role = ?", "Kaprodi").First(&kaprodi)
			
			tendikName := "Sistem / Belum ditugaskan"
			if surat.Processor != nil {
				tendikName = surat.Processor.NamaLengkap
			}
			
			if kaprodi.Email != "" {
				utils.SendSLAEscalationEmail(kaprodi.Email, kaprodi.NamaLengkap, surat.NomorSurat, tendikName)
				log.Printf("Mengirim eskalasi SLA ke Kaprodi (%s) untuk Surat %s", kaprodi.Email, surat.NomorSurat)
			}
			
			config.DB.Model(&surat).Update("sla_status", "Terlampaui")
		}
	}
}

func RunDatabaseBackup() {
	var setting models.SystemSetting
	if err := config.DB.First(&setting, 1).Error; err != nil {
		return
	}

	if !setting.BackupOtomatis {
		return
	}

	// Buat folder backups jika belum ada
	backupDir := "backups"
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		os.Mkdir(backupDir, 0755)
	}

	timestamp := time.Now().Format("20060102_150405")
	fileName := filepath.Join(backupDir, fmt.Sprintf("backup_%s.sql", timestamp))

	// Konfigurasi koneksi dari .env
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	dbName := os.Getenv("DB_NAME")
	password := os.Getenv("DB_PASSWORD")
	
	if host == "" || user == "" || dbName == "" {
		log.Println("DB Config missing, cannot run pg_dump")
		return
	}

	// Peringatan: Fitur ini mengharuskan pg_dump ada di dalam PATH OS (Environment Variable)
	cmd := exec.Command("pg_dump", "-h", host, "-p", port, "-U", user, "-d", dbName, "-F", "c", "-f", fileName)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", password))

	if err := cmd.Run(); err != nil {
		log.Printf("Gagal melakukan backup database: %v", err)
	} else {
		log.Printf("Backup database berhasil disimpan di: %s", fileName)
	}

	// Hapus file backup yang berumur lebih dari 7 hari (Sesuai NFR-006)
	cleanOldBackups(backupDir, 7)
}

func cleanOldBackups(dir string, retentionDays int) {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("Gagal membaca direktori backup untuk pembersihan: %v", err)
		return
	}

	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		info, err := f.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			oldFile := filepath.Join(dir, f.Name())
			os.Remove(oldFile)
			log.Printf("Menghapus backup lama: %s", oldFile)
		}
	}
}
