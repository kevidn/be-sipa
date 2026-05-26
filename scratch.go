package main

import (
	"fmt"
	"time"

	"github.com/kevidn/be-sipa/config"
	"github.com/kevidn/be-sipa/models"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	config.InitDB()

	var count int64
	today := time.Now().Format("2006-01-02")

	// Print total Selesai
	config.DB.Model(&models.Surat{}).Where("status = ?", "Selesai").Count(&count)
	fmt.Printf("Total Selesai: %d\n", count)

	// Print Selesai Hari Ini
	config.DB.Model(&models.Surat{}).Where("status = ? AND DATE(updated_at) = ?", "Selesai", today).Count(&count)
	fmt.Printf("Selesai Hari Ini (DATE): %d\n", count)
	
	// Alternative for Postgres
	config.DB.Model(&models.Surat{}).Where("status = ? AND updated_at::date = CURRENT_DATE", "Selesai").Count(&count)
	fmt.Printf("Selesai Hari Ini (CURRENT_DATE): %d\n", count)
}
