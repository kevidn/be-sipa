package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kevidn/be-sipa/config"
	"github.com/kevidn/be-sipa/models"
)

// Helper function to be called internally by other handlers
func CreateNotification(userID, title, message, notifType, link string) error {
	notif := models.Notification{
		UserID:  userID,
		Title:   title,
		Message: message,
		Type:    notifType,
		Link:    link,
	}
	if err := config.DB.Create(&notif).Error; err != nil {
		return err
	}
	return nil
}

// GET /api/notifications
func GetNotifications(c *fiber.Ctx) error {
	userID := c.Locals("id_user").(string)
	var notifications []models.Notification

	// Get latest notifications, sort by created_at DESC
	if err := config.DB.Where("user_id = ?", userID).Order("created_at desc").Limit(50).Find(&notifications).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil notifikasi"})
	}

	return c.JSON(fiber.Map{
		"data": notifications,
	})
}

// PUT /api/notifications/:id/read
func MarkAsRead(c *fiber.Ctx) error {
	userID := c.Locals("id_user").(string)
	id := c.Params("id")

	var notif models.Notification
	if err := config.DB.Where("id = ? AND user_id = ?", id, userID).First(&notif).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Notifikasi tidak ditemukan"})
	}

	notif.IsRead = true
	if err := config.DB.Save(&notif).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengupdate notifikasi"})
	}

	return c.JSON(fiber.Map{"message": "Notifikasi ditandai sudah dibaca"})
}

// PUT /api/notifications/read-all
func MarkAllAsRead(c *fiber.Ctx) error {
	userID := c.Locals("id_user").(string)

	if err := config.DB.Model(&models.Notification{}).Where("user_id = ?", userID).Update("is_read", true).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengupdate notifikasi"})
	}

	return c.JSON(fiber.Map{"message": "Semua notifikasi ditandai sudah dibaca"})
}

// DELETE /api/notifications/:id
func DeleteNotification(c *fiber.Ctx) error {
	userID := c.Locals("id_user").(string)
	id := c.Params("id")

	var notif models.Notification
	if err := config.DB.Where("id = ? AND user_id = ?", id, userID).First(&notif).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Notifikasi tidak ditemukan"})
	}

	if err := config.DB.Delete(&notif).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menghapus notifikasi"})
	}

	return c.JSON(fiber.Map{"message": "Notifikasi dihapus"})
}
