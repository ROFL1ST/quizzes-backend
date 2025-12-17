package utils

import (
	"encoding/json"
	"github.com/ROFL1ST/quizzes-backend/config" // Pastikan import config DB
	"github.com/ROFL1ST/quizzes-backend/models"
	"sync"
)

// Struktur Manager untuk SSE (Realtime)
type NotificationManager struct {
	Clients map[uint]chan string
	Lock    sync.Mutex
}

var NotifManager = NotificationManager{
	Clients: make(map[uint]chan string),
}

func SendNotification(userID uint, notifType, title, message, link string) {

	notif := models.Notification{
		UserID:  userID,
		Type:    notifType,
		Title:   title,
		Message: message,
		Link:    link,
		IsRead:  false,
	}
	config.DB.Create(&notif)

	NotifManager.Lock.Lock()
	clientChan, ok := NotifManager.Clients[userID]
	NotifManager.Lock.Unlock()

	if ok {

		payload, _ := json.Marshal(map[string]interface{}{
			"id":      notif.ID,
			"type":    notifType,
			"title":   title,
			"message": message,
			"link":    link,
		})

		// Non-blocking send
		select {
		case clientChan <- string(payload):
		default:
		}
	}
}
