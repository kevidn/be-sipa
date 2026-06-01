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
		log.Println("No .env file found")
	}

	config.InitDB()

	var users []models.User
	if err := config.DB.Where("nim = ? OR nim = ?", "", nil).Find(&users).Error; err != nil {
		log.Fatalf("Error finding users: %v", err)
	}

	count := 0
	for _, user := range users {
		// Asumsi: Username pada role Mahasiswa adalah NIM
		user.NIM = user.Username
		if err := config.DB.Save(&user).Error; err != nil {
			log.Printf("Error updating user %s: %v", user.Username, err)
		} else {
			count++
		}
	}

	fmt.Printf("Berhasil memperbarui %d data NIM yang kosong.\n", count)
}
