package main

import (
	"fmt"
	"math/rand"
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

func (manager *ClientManager) start() { //在webstocket中
	for {
		select {
		case conn := <-manager.register: //註冊新連線，每當有新連線註冊時透過manager.register通道接收新連線
			manager.mu.Lock()
			manager.clients[conn] = true
			manager.mu.Unlock()
		case conn := <-manager.unregister: //取消註冊連線，每當有連線要取消註冊時透過manager.unregister通道接收該連線。
			if _, ok := manager.clients[conn]; ok {
				manager.mu.Lock()
				close(conn.send)
				delete(manager.clients, conn)
				manager.mu.Unlock()
			}
		case message := <-manager.broadcast: //當有訊息要廣播時，透過manager.broadcast通道接收該訊息
			manager.mu.Lock()
			for conn := range manager.clients {
				select {
				case conn.send <- message: //對每個連線，嘗試將訊息發送到conn.send 通道
				default:
					close(conn.send)
					delete(manager.clients, conn) //若無法接通連線則表示連線已斷開，關閉連線
				}
			}
			manager.mu.Unlock() //確保在對Client操作時的正確性
		}
	}
}

func generateUserId() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, 4)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	color := getRandomColor()
	userID := "User_" + string(result)
	userID = "<span style=\"color: " + color + "; font-weight: bold;\">" + userID + "</span>"
	return userID
}

func getRandomColor() string {
	return "#" + fmt.Sprintf("%06X", rand.Intn(0xFFFFFF))
}
