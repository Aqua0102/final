// profanity.go
package main

import (
	"bufio"
	"log"
	"os"
	"strings"
)

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

func maskSwearWord(input string) string {
	message := input
	for _, word := range profanityList {
		runes := []rune(word)
		message = strings.ReplaceAll(message, word, strings.Repeat("*", len(runes)))
	}
	return message
}
