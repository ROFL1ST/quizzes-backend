package utils

import (
	"encoding/json"
	"fmt"
	"sync"
)

type LobbyManagerStruct struct {
	Clients map[uint]map[uint]chan string
	Lock    sync.Mutex
}

var LobbyManager = LobbyManagerStruct{
	Clients: make(map[uint]map[uint]chan string),
}

// BroadcastLobby mengirim pesan dengan format Standar SSE
// Format:
// event: nama_event
// data: {json_payload}
// <baris kosong>
func BroadcastLobby(challengeID uint, msgType string, payload interface{}) {
	LobbyManager.Lock.Lock()
	defer LobbyManager.Lock.Unlock()

	clients, ok := LobbyManager.Clients[challengeID]
	if !ok {
		return
	}

	// 1. Marshal Payload jadi JSON string
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error marshal payload:", err)
		return
	}

	// 2. Format Pesan SSE yang Benar
	// "event: ... \n data: ... \n\n"
	msgString := fmt.Sprintf("event: %s\ndata: %s\n\n", msgType, string(jsonPayload))

	// 3. Kirim ke semua client
	for _, ch := range clients {
		select {
		case ch <- msgString:
		default:
		}
	}
}

func AddClientToLobby(challengeID uint, userID uint) chan string {
	LobbyManager.Lock.Lock()
	defer LobbyManager.Lock.Unlock()

	if _, ok := LobbyManager.Clients[challengeID]; !ok {
		LobbyManager.Clients[challengeID] = make(map[uint]chan string)
	}

	// Buffer channel
	msgChan := make(chan string, 10)
	LobbyManager.Clients[challengeID][userID] = msgChan
	return msgChan
}

func RemoveClientFromLobby(challengeID uint, userID uint) {
	LobbyManager.Lock.Lock()
	defer LobbyManager.Lock.Unlock()

	if clients, ok := LobbyManager.Clients[challengeID]; ok {
		if ch, exists := clients[userID]; exists {
			close(ch)
			delete(clients, userID)
		}
		if len(clients) == 0 {
			delete(LobbyManager.Clients, challengeID)
		}
	}
}