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

	// Auto-migrate schema
	config.DB.AutoMigrate(&models.User{}, &models.Surat{}, &models.ActivityLog{}, &models.JenisSurat{}, &models.SystemSetting{})

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
	
	// Surat routes
	app.Post("/api/surat", handlers.AuthMiddleware, handlers.SubmitSurat)
	app.Get("/api/surat", handlers.AuthMiddleware, handlers.GetHistorySurat)
	app.Get("/api/surat/stats", handlers.AuthMiddleware, handlers.GetDashboardStats)
	app.Get("/api/surat/:id", handlers.AuthMiddleware, handlers.GetDetailSurat)
	app.Put("/api/surat/:id/status", handlers.AuthMiddleware, handlers.UpdateStatusSurat)

	// File Upload
	app.Post("/api/upload", handlers.AuthMiddleware, handlers.UploadFile)

	// Activity Logs
	app.Get("/api/logs", handlers.AuthMiddleware, handlers.GetActivityLogs)
	app.Get("/api/logs/stats", handlers.AuthMiddleware, handlers.GetLogStats)

	// Manajemen Jenis Surat
	app.Get("/api/jenis-surat", handlers.AuthMiddleware, handlers.GetAllJenisSurat)
	app.Post("/api/jenis-surat", handlers.AuthMiddleware, handlers.CreateJenisSurat)
	app.Put("/api/jenis-surat/:id", handlers.AuthMiddleware, handlers.UpdateJenisSurat)
	app.Delete("/api/jenis-surat/:id", handlers.AuthMiddleware, handlers.DeleteJenisSurat)

	// System Settings
	app.Get("/api/settings", handlers.AuthMiddleware, handlers.GetSystemSettings)
	app.Put("/api/settings", handlers.AuthMiddleware, handlers.UpdateSystemSettings)

	// Kaprodi
	app.Get("/api/kaprodi/stats", handlers.AuthMiddleware, handlers.GetKaprodiStats)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	app.Listen(":" + port)
}
