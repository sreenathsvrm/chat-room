package chat

import (
	"errors"
	"sync"
	"time"

	"github.com/sreenathsvrm/chat-room/app/internal/models"
)

type ChatRoom struct {
	clients      map[string]*Client
	broadcast    chan models.Message
	mu           sync.RWMutex
	repository   *MessageRepository
	messageCache []models.Message
	cacheSize    int
}

func NewChatRoom(repository *MessageRepository, cacheSize int) *ChatRoom {
	return &ChatRoom{
		clients:    make(map[string]*Client),
		broadcast:  make(chan models.Message, 1000),
		repository: repository,
		cacheSize:  cacheSize,
	}
}

func (cr *ChatRoom) Run() {
	go cr.broadcastMessages()
	go cr.cleanInactiveClients()
}

func (cr *ChatRoom) broadcastMessages() {
	for msg := range cr.broadcast {
		cr.mu.RLock()
		for _, client := range cr.clients {
			if client.Active {
				select {
				case client.Channel <- msg:
					client.LastSeen = time.Now()
				default:
					// Channel full, client might be disconnected
					client.Active = false
				}
			}
		}
		cr.mu.RUnlock()

		// Persist message
		if err := cr.repository.Save(&msg); err != nil {
			// Handle error (log it, retry, etc.)
		}

		// Cache message
		cr.messageCache = append(cr.messageCache, msg)
		if len(cr.messageCache) > cr.cacheSize {
			cr.messageCache = cr.messageCache[1:]
		}
	}
}

func (cr *ChatRoom) cleanInactiveClients() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		cr.mu.Lock()
		for id, client := range cr.clients {
			if !client.Active || time.Since(client.LastSeen) > 5*time.Minute {
				close(client.Channel)
				delete(cr.clients, id)
			}
		}
		cr.mu.Unlock()
	}
}

func (cr *ChatRoom) Join(clientID string) (*Client, error) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if _, exists := cr.clients[clientID]; exists {
		return nil, ErrClientExists
	}

	client := NewClient(clientID)
	cr.clients[clientID] = client
	return client, nil
}

func (cr *ChatRoom) Leave(clientID string) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if client, exists := cr.clients[clientID]; exists {
		client.Active = false
		close(client.Channel)
		delete(cr.clients, clientID)
	}
}

func (cr *ChatRoom) Send(senderID, text string) error {
	cr.mu.RLock()
	_, exists := cr.clients[senderID]
	cr.mu.RUnlock()
	if !exists {
		return ErrClientNotFound
	}
	msg := models.Message{
		SenderID:  senderID,
		Message:   text,
		CreatedAt: time.Now(),
	}
	cr.broadcast <- msg
	return nil
}

func (cr *ChatRoom) GetMessages(clientID string, since time.Time) ([]models.Message, error) {
	cr.mu.RLock()
	_, exists := cr.clients[clientID]
	cr.mu.RUnlock()

	if !exists {
		return nil, ErrClientNotFound
	}

	// First check cache
	var messages []models.Message
	for _, msg := range cr.messageCache {
		if msg.CreatedAt.After(since) {
			messages = append(messages, msg)
		}
	}

	// If cache doesn't have enough, query DB
	if len(messages) == 0 {
		dbMessages, err := cr.repository.GetMessages(since)
		if err != nil {
			return nil, err
		}
		messages = dbMessages
	}

	return messages, nil
}

var (
	ErrClientExists   = errors.New("client already exists")
	ErrClientNotFound = errors.New("client not found")
)
