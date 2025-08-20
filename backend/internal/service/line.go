package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/whatacotton/tarumi/internal/models"
)

type LineMessagingAPIClient struct {
}

func NewLineMessagingAPIClient() *LineMessagingAPIClient {
	return &LineMessagingAPIClient{}
}

func (c *LineMessagingAPIClient) ReplyMessage(message string, replyToken string) error {
	// Implement the logic to send a message using the LINE Messaging API
	url := "https://api.line.me/v2/bot/message/reply"
	accessToken := os.Getenv("LINE_CHANNEL_ACCESS_TOKEN")

	body := models.ReplyMessage{
		ReplyToken: replyToken,
		Messages: []models.MessageEntry{
			{Type: "text", Text: message},
		},
	}

	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	defer resp.Body.Close()
	fmt.Println("Status:", resp.Status)
	return nil
}
