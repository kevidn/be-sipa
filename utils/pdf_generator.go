package utils

import (
	"bytes"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
)

type KitirData struct {
	NomorSurat  string
	NamaLengkap string
	NIM         string
	JenisSurat  string
	Tanggal     time.Time
}

func GenerateKitirPDF(data KitirData) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 24)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(190, 15, "KITIR PENGANTAR", "", 1, "C", false, 0, "")

	// Subtitle
	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(190, 10, "SISTEM PELAYANAN AKADEMIK UNESA", "", 1, "C", false, 0, "")

	// Printed Date
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(120, 120, 120)
	pdf.CellFormat(190, 8, fmt.Sprintf("Tanggal Cetak: %s", time.Now().Format("02-01-2006")), "B", 1, "C", false, 0, "")
	pdf.Ln(10)

	// Nomor Pengajuan
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(190, 6, "NOMOR PENGAJUAN", "", 1, "L", false, 0, "")
	
	pdf.SetFont("Arial", "B", 24)
	pdf.SetTextColor(0, 168, 107) // sipa-green
	pdf.CellFormat(190, 12, data.NomorSurat, "", 1, "L", false, 0, "")
	pdf.Ln(5)

	// User details
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(95, 6, "NAMA MAHASISWA", "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 6, "NIM", "", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(95, 8, data.NamaLengkap, "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 8, data.NIM, "", 1, "L", false, 0, "")
	pdf.Ln(5)

	// Jenis Surat
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(190, 6, "JENIS LAYANAN", "", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(190, 8, data.JenisSurat, "", 1, "L", false, 0, "")
	pdf.Ln(5)

	// Tanggal Masuk
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(190, 6, "TANGGAL MASUK", "", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(190, 8, data.Tanggal.Format("02 January 2006, 15:04"), "", 1, "L", false, 0, "")
	pdf.Ln(15)

	// Footer / Peringatan
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(190, 6, "PERINGATAN", "T", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(100, 100, 100)
	peringatanText := "Simpan kitir ini dengan baik. Nomor pengajuan digunakan untuk melacak status surat atau saat mengambil surat fisik di tata usaha fakultas."
	pdf.MultiCell(190, 6, peringatanText, "", "C", false)

	// Output to buffer
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
