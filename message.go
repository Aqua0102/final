// message.go
package main

import (
	"fmt"
	"github.com/kyokomi/emoji/v2"
	"math/rand"
	"time"
)

func encodeMessage(data map[string]interface{}) []byte {
	userID, _ := data["userId"].(string)
	message, _ := data["message"].(string)

	// Replace swear words in the message
	message = maskSwearWord(message)
	message = emoji.Sprint(message)

	// Format the message with timestamp, user ID, and content
	timestamp := "<span style=\"color: " + "#FFFFFF" + "; font-weight: bold;\">" + time.Now().Format("15:04:05") + "</span>"
	result := timestamp + " " + userID + " " + ": " + message

	return []byte(result)
}

func getRandomColor() string {
	return "#" + fmt.Sprintf("%06X", rand.Intn(0xFFFFFF))
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
