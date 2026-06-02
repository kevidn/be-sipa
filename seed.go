package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/kevidn/be-sipa/config"
	"github.com/kevidn/be-sipa/models"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Info: File .env tidak ditemukan")
	}

	config.InitDB()

	jenisSurats := []models.JenisSurat{
		{
			Kode:        "SKMK",
			Nama:        "Surat Keterangan Masih Kuliah",
			TemplateFile: "",
			SLA:         "3 Hari Kerja",
			Status:      "Aktif",
			Persyaratan: "Kartu Tanda Mahasiswa (KTM), Kartu Rencana Studi (KRS)",
		},
		{
			Kode:        "SISP",
			Nama:        "Surat Ijin Survei/Penelitian Skripsi",
			TemplateFile: "",
			SLA:         "3 Hari Kerja",
			Status:      "Aktif",
			Persyaratan: "Kartu Tanda Mahasiswa (KTM), Transkrip Nilai",
		},
		{
			Kode:        "SKTP",
			Nama:        "Surat Keterangan Tunjangan/Pensiun",
			TemplateFile: "",
			SLA:         "3 Hari Kerja",
			Status:      "Aktif",
			Persyaratan: "Kartu Tanda Mahasiswa (KTM), Kartu Rencana Studi (KRS)",
		},
		{
			Kode:        "SKTB",
			Nama:        "Surat Keterangan Tidak Terima Beasiswa",
			TemplateFile: "",
			SLA:         "3 Hari Kerja",
			Status:      "Aktif",
			Persyaratan: "Kartu Tanda Mahasiswa (KTM), Transkrip Nilai",
		},
		{
			Kode:        "SRB",
			Nama:        "Surat Rekomendasi Beasiswa",
			TemplateFile: "",
			SLA:         "3 Hari Kerja",
			Status:      "Aktif",
			Persyaratan: "Kartu Tanda Mahasiswa (KTM), Kartu Rencana Studi (KRS), Transkrip Nilai",
		},
		{
			Kode:        "SKKB",
			Nama:        "Surat Keterangan Kelakuan Baik",
			TemplateFile: "",
			SLA:         "3 Hari Kerja",
			Status:      "Aktif",
			Persyaratan: "Kartu Tanda Mahasiswa (KTM), Transkrip Nilai",
		},
	}

	for _, js := range jenisSurats {
		var existing models.JenisSurat
		if config.DB.Where("kode = ?", js.Kode).First(&existing).Error == nil {
			// Update
			existing.Nama = js.Nama
			existing.Persyaratan = js.Persyaratan
			existing.SLA = js.SLA
			config.DB.Save(&existing)
			fmt.Println("Updated", js.Nama)
		} else {
			// Create
			config.DB.Create(&js)
			fmt.Println("Created", js.Nama)
		}
	}
	
	// Ensure the mismatched names from previous seed are deleted or updated.
	config.DB.Where("kode = '' OR kode IS NULL").Delete(&models.JenisSurat{})
	
	fmt.Println("Seeding complete")
}
