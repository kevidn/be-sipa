package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/kevidn/be-sipa/config"
	"github.com/kevidn/be-sipa/models"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	config.InitDB()

	var surat []models.Surat
	query := config.DB.Order("updated_at DESC").Limit(5).Preload("User")
	if err := query.Find(&surat).Error; err != nil {
		log.Fatal(err)
	}

	for _, s := range surat {
		data, _ := json.MarshalIndent(s, "", "  ")
		fmt.Printf("Item: %s\n", string(data))
	}
}
