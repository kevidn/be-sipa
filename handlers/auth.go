package handlers

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kevidn/be-sipa/config"
	"github.com/kevidn/be-sipa/models"
	"golang.org/x/crypto/bcrypt"
)

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RoleValid adalah daftar role yang diizinkan saat registrasi
var RoleValid = map[string]bool{
	"Mahasiswa": true,
	"Dosen":     true,
	"Kaprodi":   true,
	"Tendik":    true,
}

type RegisterInput struct {
	NamaLengkap string `json:"nama_lengkap"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	PhoneNumber string `json:"phone_number"`
	Role        string `json:"role"`
	Password    string `json:"password"`
}

func generateUserID() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("USR%07d", r.Intn(9999999))
}

func Register(c *fiber.Ctx) error {
	var input RegisterInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Input data tidak valid"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal memproses kata sandi"})
	}

	idUser := fmt.Sprintf("U%d", time.Now().UnixNano()%1000000)

	newUser := models.User{
		IDUser:       idUser,
		Username:     input.Username,
		PasswordHash: string(hashedPassword),
		NamaLengkap:  input.NamaLengkap,
		Email:        input.Email,
		Role:         input.Role,
		PhoneNumber:  input.PhoneNumber,
		StatusAkun:   "Aktif",
	}

	result := config.DB.Create(&newUser)
	if result.Error != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Username atau Email sudah terdaftar!"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Pendaftaran berhasil",
		"data":    newUser.Username,
	})
}

func Login(c *fiber.Ctx) error {
	var input LoginInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Input tidak valid"})
	}

	var user models.User
	result := config.DB.Where("username = ?", input.Username).First(&user)
	if result.Error != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Username tidak ditemukan"})
	}

	if user.StatusAkun != "Aktif" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Akun Anda: " + user.StatusAkun})
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Password salah!"})
	}

	now := time.Now()
	config.DB.Model(&user).Update("last_login", &now)

	claims := jwt.MapClaims{
		"id_user":        user.IDUser,
		"role":           user.Role,
		"is_sla_monitor": user.IsSlaMonitor,
		"exp":            time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		secretKey = "rahasia_default"
	}

	t, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal membuat token"})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Login berhasil",
		"token":   t,
		"data":    user,
	})
}
