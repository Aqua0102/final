package main

import (
	"log"

	"net/http"

	"github.com/gin-gonic/gin"
)

func init() {
	if err := loadProfanityList(); err != nil {
		log.Fatal(err)
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
		http.ServeFile(c.Writer, c.Request, "static/index.html")
	})

	r.GET("/ws", func(c *gin.Context) {
		serveWs(&manager, c.Writer, c.Request)
	})
	r.Run(":80")
}
