package config

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dbname := os.Getenv("DB_NAME")
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSLMODE"),
		os.Getenv("DB_TIMEZONE"),
	)

	database, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})

	if err != nil {
		panic("Gagal terkoneksi ke database!")
	}

	fmt.Printf("✅ Terhubung ke database: %s\n", dbname)
	DB = database

	// Verifikasi kolom (Debug)
	var columns []string
	DB.Raw("SELECT column_name FROM information_schema.columns WHERE table_name = 'users'").Scan(&columns)
	fmt.Printf("DEBUG: Kolom di tabel users: %v\n", columns)
}
