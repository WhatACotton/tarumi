package models

import "time"

type ReplyMessage struct {
	ReplyToken string         `json:"replyToken"`
	Messages   []MessageEntry `json:"messages"`
}

type MessageEntry struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type LineTalkRoom struct {
	OwnerID      string
	FriendRoomID string
}

type WaitingFriendRoomsQueue struct {
	OwnerID      string
	IdentifyCode string
	CreatedAt    time.Time
}

type LineWebhookEvent struct {
	Destination string      `json:"destination"`
	Events      []LineEvent `json:"events"`
}

type LineEvent struct {
	Type            string          `json:"type"`
	Message         *LineMessage    `json:"message,omitempty"`
	WebhookEventID  string          `json:"webhookEventId"`
	DeliveryContext DeliveryContext `json:"deliveryContext"`
	Timestamp       int64           `json:"timestamp"`
	Source          LineSource      `json:"source"`
	ReplyToken      string          `json:"replyToken"`
	Mode            string          `json:"mode"`
}

type LineMessage struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	QuoteToken string `json:"quoteToken"`
	Text       string `json:"text"`
}

type DeliveryContext struct {
	IsRedelivery bool `json:"isRedelivery"`
}

type LineSource struct {
	Type   string `json:"type"`
	UserID string `json:"userId"`
}
