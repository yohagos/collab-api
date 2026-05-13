package domain

import "errors"

var (
	ErrRoomNotFound    = errors.New("room not found")
	ErrUserNotInRoom   = errors.New("user not in room")
	ErrInvalidRoomName = errors.New("invalid room name")
	ErrInvalidMessage  = errors.New("invalid message")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrInvalidToken    = errors.New("invalid tpken")
)
