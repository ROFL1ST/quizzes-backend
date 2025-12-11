package utils

import (
	"encoding/json"
	"sync"
)


type NotificationPayload struct {
	Message string `json:"message"`
	Url     string `json:"url"`  
	Type    string `json:"type"`
}


type NotificationManager struct {
	Clients map[uint]chan string
	Lock    sync.Mutex
}

var NotifManager = NotificationManager{
	Clients: make(map[uint]chan string),
}


func SendNotification(userID uint, message string, url string, notifType string) {
	NotifManager.Lock.Lock()
	defer NotifManager.Lock.Unlock()

	payload := NotificationPayload{
		Message: message,
		Url:     url,
		Type:    notifType,
	}
	jsonMsg, _ := json.Marshal(payload)

	if ch, ok := NotifManager.Clients[userID]; ok {
		select {
		case ch <- string(jsonMsg):
		default:
		}
	}
}