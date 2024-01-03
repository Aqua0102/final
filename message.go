// message.go
package main

import (
	"time"
)

func encodeMessage(data map[string]interface{}) []byte {
	userID, _ := data["userId"].(string)
	message, _ := data["message"].(string)

	// Replace swear words in the message
	message = maskSwearWord(message)

	// Format the message with timestamp, user ID, and content
	timestamp := "<span style=\"color: " + "#FFFFFF" + "; font-weight: bold;\">" + time.Now().Format("15:04:05") + "</span>"
	result := timestamp + " " + userID + " " + ": " + message

	return []byte(result)
}
