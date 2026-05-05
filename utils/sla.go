package utils

import (
	"time"
)

// CalculateDeadline calculates the deadline based on working days (Mon-Fri)
func CalculateDeadline(start time.Time, days int) time.Time {
	result := start
	addedDays := 0
	for addedDays < days {
		result = result.AddDate(0, 0, 1)
		// Check if it's a weekend (Saturday or Sunday)
		if result.Weekday() != time.Saturday && result.Weekday() != time.Sunday {
			// In a real system, you'd also check against a list of public holidays here
			addedDays++
		}
	}
	// Set to end of working day (e.g., 16:00)
	return time.Date(result.Year(), result.Month(), result.Day(), 16, 0, 0, 0, result.Location())
}

// GetSLADays returns the number of working days for each letter type based on BR-002 and BR-003
func GetSLADays(jenisSurat string) int {
	switch jenisSurat {
	case "Surat Keterangan Masih Kuliah", "Surat Keterangan Kelakuan Baik":
		return 2
	case "Surat Ijin Survei Penelitian (Skripsi)", "Surat Rekomendasi Beasiswa":
		return 5
	case "Surat Tunjangan/Pensiun/Akses":
		return 3
	case "Surat Keterangan Tidak Menerima Beasiswa":
		return 2
	default:
		return 3
	}
}
