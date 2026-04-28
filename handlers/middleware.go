package handlers

import (
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token tidak ditemukan"})
	}

	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
	
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		secretKey = "rahasia_default"
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token tidak valid"})
	}

	claims := token.Claims.(jwt.MapClaims)
	c.Locals("id_user", fmt.Sprintf("%v", claims["id_user"]))
	c.Locals("role", fmt.Sprintf("%v", claims["role"]))

	return c.Next()
}
