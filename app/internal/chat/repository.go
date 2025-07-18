package chat

import (
	"time"

	"gorm.io/gorm"

	"github.com/sreenathsvrm/chat-room/app/internal/models"
)

type MessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Save(message *models.Message) error {
	return r.db.Create(message).Error
}

func (r *MessageRepository) GetMessages(since time.Time) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.Where("created_at > ?", since).Order("created_at asc").Find(&messages).Error
	return messages, err
}
