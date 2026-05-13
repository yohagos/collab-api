package domain

import "time"

type Message struct {
	ID        MessageID `json:"id"`
	RoomID    RoomID    `json:"room_id"`
	UserID    UserID    `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}