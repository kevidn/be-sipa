package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kevidn/be-sipa/config"
	"github.com/kevidn/be-sipa/models"
)

func UploadFile(c *fiber.Ctx) error {
	// Parse the multipart form
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Gagal membaca file upload"})
	}

	// Fetch MaxFileUploadMB from DB
	var setting models.SystemSetting
	config.DB.First(&setting, 1)
	maxMB := setting.MaxFileUploadMB
	if maxMB <= 0 {
		maxMB = 5 // Fallback
	}

	// Validate file size
	if fileHeader.Size > int64(maxMB)*1024*1024 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("Ukuran file terlalu besar. Maksimal %dMB", maxMB)})
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext != ".pdf" && ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".doc" && ext != ".docx" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format file tidak didukung. Harap upload PDF, JPG, PNG, DOC, atau DOCX"})
	}

	// Open the file
	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal membuka file"})
	}
	defer file.Close()

	// Read file content into a buffer
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal membaca isi file"})
	}

	// Read environment variables
	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	supabaseBucket := os.Getenv("SUPABASE_BUCKET")

	if supabaseUrl == "" || supabaseKey == "" || supabaseBucket == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Konfigurasi Supabase Storage tidak lengkap"})
	}

	// Generate a unique filename
	uniqueId := uuid.New().String()
	timestamp := time.Now().Format("20060102150405")
	newFilename := fmt.Sprintf("%s_%s%s", timestamp, uniqueId[:8], ext)

	// Determine content type
	contentType := "application/octet-stream"
	if ext == ".pdf" {
		contentType = "application/pdf"
	} else if ext == ".png" {
		contentType = "image/png"
	} else if ext == ".jpg" || ext == ".jpeg" {
		contentType = "image/jpeg"
	} else if ext == ".doc" {
		contentType = "application/msword"
	} else if ext == ".docx" {
		contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	}

	// Upload to Supabase Storage
	// Endpoint: POST [SUPABASE_URL]/storage/v1/object/[BUCKET]/[FILENAME]
	uploadUrl := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseUrl, supabaseBucket, newFilename)

	req, err := http.NewRequest("POST", uploadUrl, bytes.NewReader(fileBytes))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal membuat request upload"})
	}

	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("Content-Type", contentType)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengirim file ke server penyimpanan"})
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Gagal mengunggah file ke Supabase",
			"details": string(bodyBytes),
		})
	}

	// Generate Public URL
	// Endpoint: [SUPABASE_URL]/storage/v1/object/public/[BUCKET]/[FILENAME]
	publicUrl := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseUrl, supabaseBucket, newFilename)

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "File berhasil diunggah",
		"data": fiber.Map{
			"file_url": publicUrl,
		},
	})
}
