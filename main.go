package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ClientManager struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

type Client struct {
	socket *websocket.Conn
	send   chan []byte
	UserId string
}

var profanityList []string

func loadProfanityList() error {
	file, err := os.Open("swear_word.txt")
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		profanityList = append(profanityList, word)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func init() {
	if err := loadProfanityList(); err != nil {
		log.Fatal(err)
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

func main() {
	r := gin.Default()

	manager := ClientManager{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}

	go manager.start()

	r.GET("/", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "index.html")
	})

	r.GET("/ws", func(c *gin.Context) {
		serveWs(&manager, c.Writer, c.Request)
	})

	r.Run(":80")
}

func serveWs(manager *ClientManager, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	userId := generateUserId()

	// 向前端發送一條完整的訊息
	connectedMessage := "Connected! You are " + userId
	conn.WriteMessage(websocket.TextMessage, []byte(connectedMessage))

	// Send a welcome message to the connected client
	welcomeMessage := map[string]interface{}{
		"userId":  "server",
		"message": userId + " joined the chat",
	}
	manager.broadcast <- encodeMessage(welcomeMessage)

	// 建立 Client 時傳遞 UserId
	client := &Client{socket: conn, send: make(chan []byte), UserId: userId}
	manager.register <- client

	go client.writePump()
	client.readPump(manager)

	// 在用戶斷開連線時發送離開訊息
	leftMessage := map[string]interface{}{
		"userId":  "server",
		"message": userId + " leave the chat",
	}
	manager.broadcast <- encodeMessage(leftMessage)
}

func (c *Client) readPump(manager *ClientManager) {
	defer func() {
		manager.unregister <- c
		c.socket.Close()

		// 離開聊天室時顯示成員已離開
		leftMessage := map[string]interface{}{
			"userId":  "server",
			"message": c.UserId + " leave the chat",
		}
		manager.broadcast <- encodeMessage(leftMessage)
	}()

	for {
		messageType, p, err := c.socket.ReadMessage()
		if err != nil {
			return
		}
		message := append([]byte{}, p...)
		data := map[string]interface{}{
			"userId":  c.UserId,
			"message": string(message),
		}

		manager.broadcast <- encodeMessage(data)

		if messageType == websocket.CloseMessage {
			return
		}
	}
}

func (c *Client) writePump() {
	defer func() {
		c.socket.Close()
	}()

	for {

		select {
		case message, ok := <-c.send:
			if !ok {
				c.socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			data := map[string]interface{}{
				"userId":  c.UserId,
				"message": string(message),
			}

			value := data["message"].(string)

			c.socket.WriteMessage(websocket.TextMessage, []byte(value))

		}
	}
}

func encodeMessage(data map[string]interface{}) []byte { // 將 UserID 和 Message 組合在一起並編碼
	userID, _ := data["userId"].(string)
	message, _ := data["message"].(string)

	message = maskSwearWord(message)
	//消除髒話
	timestamp := "<span style=\"color: " + "#FFFFFF" + "; font-weight: bold;\">" + time.Now().Format("15:04:05") + "</span>"

	result := timestamp + " " + userID + " " + ": " + message

	return []byte(result)
}

func getRandomColor() string {
	return "#" + fmt.Sprintf("%06X", rand.Intn(0xFFFFFF))
}

func generateUserId() string { // 生成 UserID
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

func maskSwearWord(input string) string {
	message := input
	for _, word := range profanityList {
		runes := []rune(word)
		message = strings.ReplaceAll(message, word, strings.Repeat("*", len(runes)))
	}
	return message
}
