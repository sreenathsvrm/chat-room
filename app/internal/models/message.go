package models

import "time"

type Message struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	SenderID  string    `json:"sender_id" gorm:"not null"`
	Text      string    `json:"text" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`
}
