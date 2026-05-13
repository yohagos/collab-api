package domain

import "time"

type CursorPosition struct {
	RoomID    RoomID    `json:"room_id"`
	UserID    UserID    `json:"user_id"`
	Username  string    `json:"username"`
	X         float64   `json:"x"`
	Y         float64   `json:"y"`
	Color     string    `json:"color,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}
