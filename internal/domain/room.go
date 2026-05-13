package domain

import (
	"time"

	"github.com/google/uuid"
)

type (
	RoomID uuid.UUID
	UserID uuid.UUID
	MessageID uuid.UUID
)

func (id RoomID) String() string {return uuid.UUID(id).String()}
func (id UserID) String() string {return uuid.UUID(id).String()}
func (id MessageID) String() string {return uuid.UUID(id).String()}

type Room struct {
	ID          RoomID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedBy   UserID `json:"created_by"`
	CreatedAt   time.Time `json:"creeated_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RoomMember struct {
	UserID   UserID `json:"id"`
	Username string    `json:"username"`
	JoinedAt time.Time `json:"joined_at"`
	LastSeen time.Time `json:"last_seen"`
}

type User struct {
	ID       UserID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
}
