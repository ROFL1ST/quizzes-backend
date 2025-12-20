package controllers

import (
	"bufio"
	"fmt"
	"github.com/ROFL1ST/quizzes-backend/utils" // Import utils
	"github.com/gofiber/fiber/v2"
	"time"
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
)

func StreamNotifications(c *fiber.Ctx) error {
	userVal := c.Locals("user_id")
	if userVal == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	userID := uint(userVal.(float64))

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	msgChan := make(chan string)

	// Gunakan utils.NotifManager
	utils.NotifManager.Lock.Lock()
	utils.NotifManager.Clients[userID] = msgChan
	utils.NotifManager.Lock.Unlock()

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		defer func() {
			utils.NotifManager.Lock.Lock()
			delete(utils.NotifManager.Clients, userID)
			utils.NotifManager.Lock.Unlock()
			close(msgChan)
		}()

		for {
			select {
			case msg := <-msgChan:
				fmt.Fprintf(w, "data: %s\n\n", msg)
				if err := w.Flush(); err != nil {
					return
				}
			case <-ticker.C:
				fmt.Fprintf(w, ":keepalive\n\n")
				if err := w.Flush(); err != nil {
					return
				}
			}
		}
	})

	return nil
}

func GetMyNotifications(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)

	var notifs []models.Notification
	// Ambil notifikasi terbaru dulu, limit 50
	if err := config.DB.Where("user_id = ?", userID).Order("created_at desc").Limit(50).Find(&notifs).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil notifikasi", nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Success", notifs)
}

// PUT /api/notifications/:id/read
func MarkNotificationRead(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)
	notifID := c.Params("id")

	// Pastikan notifikasi milik user tersebut
	result := config.DB.Model(&models.Notification{}).Where("id = ? AND user_id = ?", notifID, userID).Update("is_read", true)
	
	if result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Error updating notification", nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Notification read", nil)
}

// DELETE /api/notifications (Clear All)
func ClearAllNotifications(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)
	
	// Soft delete semua notifikasi user
	config.DB.Where("user_id = ?", userID).Delete(&models.Notification{})

	return utils.SuccessResponse(c, fiber.StatusOK, "All notifications cleared", nil)
}

func MarkAllNotificationsRead(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)

	// Update semua notifikasi milik user menjadi is_read = true
	result := config.DB.Model(&models.Notification{}).
		Where("user_id = ?", userID).
		Update("is_read", true)

	if result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengupdate notifikasi", nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Semua notifikasi ditandai sudah dibaca", nil)
}