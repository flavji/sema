package websockethelper

import (
	"fmt"
	"log"
	"net/http"
	"sema/models/delta"
	"sync"

	"github.com/gorilla/websocket"
)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	EnableCompression: false,
}

type WebSocketManager struct {
	connections map[string]map[*websocket.Conn]bool // Section -> map of connections
	mu          sync.Mutex
}

func SpawnWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		connections: make(map[string]map[*websocket.Conn]bool),
	}
}

func (manager *WebSocketManager) OpenConnection(id string, conn *websocket.Conn) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if manager.connections[id] == nil {
		manager.connections[id] = make(map[*websocket.Conn]bool)
	}

	manager.connections[id][conn] = true
	log.Printf("WebSocket opened for section %s", id)
}

func (manager *WebSocketManager) CloseConnection(id string, conn *websocket.Conn) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if manager.connections[id] != nil {
		delete(manager.connections[id], conn)

		if len(manager.connections[id]) == 0 {
			delete(manager.connections, id)
		}
	}
}

func (manager *WebSocketManager) SendToIDExpectConn(id string, message delta.Delta, expectConn *websocket.Conn) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	for conn := range manager.connections[id] {
		if expectConn == conn {
			fmt.Println("Broadcastor: ", conn.RemoteAddr())
			continue
		}
		fmt.Println("Sending to: ", conn.RemoteAddr())
		if err := conn.WriteJSON(message); err != nil {
			log.Println("Error sending message:", err)
			conn.Close()
			manager.CloseConnection(id, conn) // Close and remove the connection if it fails
			
		}
	}
}


func (manager *WebSocketManager) GetNumofConns(id string) int {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	return len(manager.connections[id])
}



// Add error checking 
func (manager *WebSocketManager) RequestSectionContents(id string, expectConn *websocket.Conn) {
  manager.mu.Lock()
  defer manager.mu.Unlock()

  for conn := range manager.connections[id] {
    if conn == expectConn {
      continue
    }
    request := map[string]string{"action": "request_contents"}
    if err := conn.WriteJSON(request); err != nil {
      log.Println("Error requesting section contents:", err)
      conn.Close()
      manager.CloseConnection(id, conn)
    }
    return
  }
}

