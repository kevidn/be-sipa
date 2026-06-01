package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/kevidn/be-sipa/config"
	"github.com/kevidn/be-sipa/models"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	godotenv.Load()
	config.InitDB()

	// Hash password
	password := "password"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Gagal hash password: %v", err)
	}

	admin := models.User{
		IDUser:       "ADM-001",
		Username:     "superadmin",
		PasswordHash: string(hashedPassword),
		NamaLengkap:  "Super Admin",
		Email:        "admin@sipa.unesa.ac.id",
		Role:         "Admin Sistem",
		StatusAkun:   "Aktif",
	}

	// Insert or Update
	if err := config.DB.Save(&admin).Error; err != nil {
		log.Fatalf("Gagal insert admin: %v", err)
	}

	fmt.Println("Berhasil membuat user admin!")
}
