package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/kevidn/be-sipa/config"
	"github.com/kevidn/be-sipa/models"
)

func main() {
	godotenv.Load()
	config.InitDB()

	query := config.DB.Model(&models.Surat{})

	var total int64
	query.Count(&total) // Mutates?

	log.Printf("Total after count: %d", total)

	var surat []models.Surat
	err := query.Offset(0).Limit(5).Find(&surat).Error
	log.Printf("Surat len after find: %d, err: %v", len(surat), err)
}
