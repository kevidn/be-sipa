package main

import (
	"os"

	"github.com/kevidn/be-sipa/config"
	"github.com/kevidn/be-sipa/handlers"
	"github.com/kevidn/be-sipa/models"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	config.InitDB()

	// Auto-migrate schema: buat/update tabel users jika belum ada
	config.DB.AutoMigrate(&models.User{})

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: os.Getenv("FRONTEND_URL"),
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
	}))

	// Health check
	app.Get("/api/ping", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "active", "message": "Sistem Akademik API"})
	})

	// Auth routes
	app.Post("/api/login", handlers.Login)
	app.Post("/api/register", handlers.Register)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	app.Listen(":" + port)
}
