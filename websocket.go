package main

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func serveWs(manager *ClientManager, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	userId := generateUserId()

	// Send a connected message to the client
	connectedMessage := "Connected! You are " + userId
	conn.WriteMessage(websocket.TextMessage, []byte(connectedMessage))

	// Send a welcome message to all clients
	welcomeMessage := map[string]interface{}{
		"userId":  "server",
		"message": userId + " joined the chat",
	}
	manager.broadcast <- encodeMessage(welcomeMessage)

	client := &Client{socket: conn, send: make(chan []byte), UserId: userId}
	manager.register <- client

	go client.writePump()
	client.readPump(manager)

	// Send a leave message when the user disconnects
	leftMessage := map[string]interface{}{
		"userId":  "server",
		"message": userId + " left the chat",
	}
	manager.broadcast <- encodeMessage(leftMessage)
}

type ClientManager struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (manager *ClientManager) start() {
	for {
		select {
		case conn := <-manager.register:
			manager.mu.Lock()
			manager.clients[conn] = true
			manager.mu.Unlock()
		case conn := <-manager.unregister:
			if _, ok := manager.clients[conn]; ok {
				manager.mu.Lock()
				close(conn.send)
				delete(manager.clients, conn)
				manager.mu.Unlock()
			}
		case message := <-manager.broadcast:
			manager.mu.Lock()
			for conn := range manager.clients {
				select {
				case conn.send <- message:
				default:
					close(conn.send)
					delete(manager.clients, conn)
				}
			}
			manager.mu.Unlock()
		}
	}
}
