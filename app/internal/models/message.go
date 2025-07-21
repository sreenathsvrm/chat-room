package models

import "time"

type Message struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	SenderID  string    `json:"sender_id" gorm:"not null"`
	Message   string    `json:"message" gorm:"column:message;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`
}
