package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/kevidn/be-sipa/config"
	"github.com/kevidn/be-sipa/models"
)

func main() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env file found, relying on system env variables")
	}

	config.InitDB()
	
	// Set program_studi to 'S1 Sistem Informasi' for any user with empty program_studi
	result := config.DB.Model(&models.User{}).
		Where("program_studi = ? OR program_studi IS NULL", "").
		Update("program_studi", "S1 Sistem Informasi")
		
	if result.Error != nil {
		log.Fatalf("Failed to update users: %v", result.Error)
	}
	
	fmt.Printf("Successfully updated %d users to 'S1 Sistem Informasi'\n", result.RowsAffected)
}
