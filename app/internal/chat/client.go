package chat

import (
	"time"

	"github.com/sreenathsvrm/chat-room/app/internal/models"
)

type Client struct {
	ID       string
	Channel  chan models.Message
	Active   bool
	LastSeen time.Time
}

func NewClient(id string) *Client {
	return &Client{
		ID:       id,
		Channel:  make(chan models.Message, 100),
		Active:   true,
		LastSeen: time.Now(),
	}
}
