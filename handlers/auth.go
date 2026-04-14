package handlers

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kevidn/be-sipa/config"
	"github.com/kevidn/be-sipa/models"
	"github.com/kevidn/be-sipa/utils"
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

func ForgotPassword(c *fiber.Ctx) error {
	type ForgotInput struct {
		Email string `json:"email"`
	}
	var input ForgotInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Input tidak valid"})
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		// Untuk keamanan, jangan beri tahu jika email tidak ada
		return c.JSON(fiber.Map{"message": "Jika email terdaftar, tautan reset akan dikirim."})
	}

	token := uuid.New().String()
	expiry := time.Now().UTC().Add(time.Hour * 1) // 1 jam - Menggunakan UTC

	user.ResetPasswordToken = token
	user.ResetPasswordExpires = &expiry

	fmt.Printf("DEBUG: Saving token %s for user %s (Expires UTC: %v)\n", token, user.Email, expiry)
	if err := config.DB.Save(&user).Error; err != nil {
		fmt.Printf("DEBUG: Save error: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal memproses permintaan"})
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, token)

	mailData := utils.MailData{
		NamaLengkap: user.NamaLengkap,
		ResetLink:   resetLink,
	}

	if err := utils.SendResetPasswordEmail(user.Email, mailData); err != nil {
		fmt.Printf("SMTP Error: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengirim email reset"})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Tautan reset kata sandi telah dikirim ke email Anda.",
	})
}

func ResetPassword(c *fiber.Ctx) error {
	type ResetInput struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	var input ResetInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Input tidak valid"})
	}

	var user models.User
	fmt.Printf("DEBUG: Looking for token: %s\n", input.Token)
	if err := config.DB.Where("reset_password_token = ?", input.Token).First(&user).Error; err != nil {
		fmt.Printf("DEBUG: Token not found: %v\n", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Tautan tidak valid atau telah kadaluarsa"})
	}

	now := time.Now().UTC()
	fmt.Printf("DEBUG: Checking expiry. Now (UTC): %v, Expires (DB): %v\n", now, user.ResetPasswordExpires)

	if user.ResetPasswordExpires == nil || now.After(*user.ResetPasswordExpires) {
		fmt.Printf("DEBUG: Token expired. Now is after Expires.\n")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Tautan reset telah kadaluarsa"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal memproses kata sandi baru"})
	}

	// Gunakan Updates dengan map untuk memastikan field "" dan nil ikut terhapus
	if err := config.DB.Model(&user).Updates(map[string]interface{}{
		"password_hash":          string(hashedPassword),
		"reset_password_token":   "",
		"reset_password_expires": nil,
	}).Error; err != nil {
		fmt.Printf("DEBUG: Updates error: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal memperbarui kata sandi"})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Kata sandi berhasil diperbarui. Silakan masuk kembali.",
	})
}

func Logout(c *fiber.Ctx) error {
	// Untuk JWT stateless, logout biasanya ditangani di sisi klien dengan menghapus token.
	// Endpoint ini bisa digunakan untuk logging atau invalidasi token di masa depan.
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Logout berhasil",
	})
}
