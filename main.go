package main

import (
	"log"
	"os"

	"github.com/kevidn/be-sipa/config"
	"github.com/kevidn/be-sipa/handlers"
	"github.com/kevidn/be-sipa/models"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Info: File .env tidak ditemukan, menggunakan variabel environment dari sistem cloud.")
	}

	config.InitDB()

	// Auto-migrate schema: buat/update tabel users jika belum ada
	config.DB.AutoMigrate(&models.User{})

	app := fiber.New()

	allowOrigins := os.Getenv("FRONTEND_URL")
	if allowOrigins == "" {
		allowOrigins = "*"
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins: allowOrigins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
	}))

	// Health check
	app.Get("/api/ping", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "active", "message": "Sistem Akademik API"})
	})

	// Auth routes
	app.Post("/api/login", handlers.Login)
	app.Post("/api/register", handlers.Register)
	app.Post("/api/forgot-password", handlers.ForgotPassword)
	app.Post("/api/reset-password", handlers.ResetPassword)
	app.Post("/api/logout", handlers.Logout)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	app.Listen(":" + port)
}
