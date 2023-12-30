// client.go
package main

import "github.com/gorilla/websocket"

type Client struct {
	socket *websocket.Conn
	send   chan []byte
	UserId string
}

func (c *Client) readPump(manager *ClientManager) {
	defer func() {
		manager.unregister <- c
		c.socket.Close()

		// Notify other clients when a user leaves
		leftMessage := map[string]interface{}{
			"userId":  "server",
			"message": c.UserId + " left the chat",
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

		// Broadcast the message to all clients
		manager.broadcast <- encodeMessage(data)

		// Close the connection if it's a close message
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
				// Connection closed by the server
				c.socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			data := map[string]interface{}{
				"userId":  c.UserId,
				"message": string(message),
			}

			// Extract the message value and send it to the client
			value := data["message"].(string)
			c.socket.WriteMessage(websocket.TextMessage, []byte(value))
		}
	}
}
