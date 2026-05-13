package domain

import (
	"context"
)

type RoomRepository interface {
	Create(ctx context.Context, room *Room) error
	GetByID(ctx context.Context, id RoomID) (*Room, error)
	List(ctx context.Context, limit int) ([]*Room, error)
}

type MessageRepository interface {
	Save(ctx context.Context, msg *Message) error
	GetByRoom(ctx context.Context, roomID RoomID, limit int) ([]*Message, error)
}
