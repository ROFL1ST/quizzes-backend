package controllers

import (
	"bufio"
	"fmt"
	"github.com/ROFL1ST/quizzes-backend/utils" // Import utils
	"github.com/gofiber/fiber/v2"
	"time"
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