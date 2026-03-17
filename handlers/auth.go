package handlers

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/kevidn/be-sipa/config"
	"github.com/kevidn/be-sipa/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
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
	NimNip      string `json:"nim_nip"`      // dipakai sebagai username
	NamaLengkap string `json:"nama_lengkap"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Role        string `json:"role"`         // wajib: Mahasiswa | Dosen | Kaprodi | Tendik
	PhoneNumber string `json:"phone_number"` // opsional
}

// generateUserID membuat ID user acak berformat 'USR' + 7 angka, contoh: USR0012345
func generateUserID() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("USR%07d", r.Intn(9999999))
}

func Register(c *fiber.Ctx) error {
	var input RegisterInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Input tidak valid"})
	}

	// --- Validasi field wajib ---
	input.NimNip = strings.TrimSpace(input.NimNip)
	input.NamaLengkap = strings.TrimSpace(input.NamaLengkap)
	input.Email = strings.TrimSpace(input.Email)

	input.Role = strings.TrimSpace(input.Role)

	if input.NimNip == "" || input.NamaLengkap == "" || input.Email == "" || input.Password == "" || input.Role == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "NIM/NIP, Nama Lengkap, Email, Password, dan Role wajib diisi"})
	}

	if !RoleValid[input.Role] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Role tidak valid. Pilih salah satu: Mahasiswa, Dosen, Kaprodi, Tendik"})
	}

	if len(input.Password) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Password minimal 8 karakter"})
	}

	// --- Cek duplikasi username (NIM/NIP) ---
	var existing models.User
	if result := config.DB.Where("username = ?", input.NimNip).First(&existing); result.Error == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "NIM/NIP sudah terdaftar"})
	}

	// --- Cek duplikasi email ---
	if result := config.DB.Where("email = ?", input.Email).First(&existing); result.Error == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Email sudah terdaftar"})
	}

	// --- Hash password ---
	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal memproses password"})
	}

	// --- Buat user baru ---
	newUser := models.User{
		IDUser:       generateUserID(),
		Username:     input.NimNip,
		PasswordHash: string(hashed),
		NamaLengkap:  input.NamaLengkap,
		Email:        input.Email,
		PhoneNumber:  strings.TrimSpace(input.PhoneNumber),
		Role:         input.Role,
		StatusAkun:   "Aktif",
	}

	if result := config.DB.Create(&newUser); result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyimpan akun: " + result.Error.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "Akun berhasil didaftarkan",
		"data": fiber.Map{
			"id_user":      newUser.IDUser,
			"username":     newUser.Username,
			"nama_lengkap": newUser.NamaLengkap,
			"email":        newUser.Email,
			"role":         newUser.Role,
			"status_akun":  newUser.StatusAkun,
		},
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