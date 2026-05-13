package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yohagos/collab-api/internal/domain"
	"go.uber.org/zap"
)

type RoomService struct {
	repo   domain.RoomRepository
	logger *zap.Logger
}

func NewRoomService(repo domain.RoomRepository, logger *zap.Logger) *RoomService {
	return &RoomService{
		repo:   repo,
		logger: logger,
	}
}

func (s *RoomService) CreateRoom(ctx context.Context, name, description string, userID domain.UserID) (*domain.Room, error) {
	if name == "" || len(name) < 3 {
		return nil, domain.ErrInvalidRoomName
	}

	room := &domain.Room{
		ID:          domain.RoomID(uuid.New()),
		Name:        name,
		Description: description,
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, room); err != nil {
		return nil, err
	}

	s.logger.Info(
		"Room created",
		zap.String("room_id", room.ID.String()),
		zap.String("name", room.Name),
		zap.String("created_by", userID.String()),
	)

	return room, nil
}

func (s *RoomService) GetRoom(ctx context.Context, roomID domain.RoomID) (*domain.Room, error) {
	return s.repo.GetByID(ctx, roomID)
}

func (s *RoomService) JoinRoom(
	ctx context.Context,
	roomID domain.RoomID,
	userID domain.UserID,
	username string,
) error {
	// TODO
	s.logger.Info(
		"User joined room",
		zap.String("room_id", roomID.String()),
		zap.String("user_id", userID.String()),
		zap.String("username", username),
	)

	return nil
}

func (s *RoomService) ListRooms(ctx context.Context, limit int) ([]*domain.Room, error) {
	rooms, err := s.repo.List(ctx, limit)
	if err != nil {
		return nil, err
	}
	return rooms, nil
}
