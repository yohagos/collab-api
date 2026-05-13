package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yohagos/collab-api/internal/domain"
	"go.uber.org/zap"
)

type MessageService struct {
	repo   domain.MessageRepository
	logger *zap.Logger
}

func NewMessageService(repo domain.MessageRepository, logger *zap.Logger) *MessageService {
	return &MessageService{
		repo:   repo,
		logger: logger,
	}
}

func (s *MessageService) SaveMessage(
	ctx context.Context,
	roomID domain.RoomID,
	userID domain.UserID,
	username, content string,
) (*domain.Message, error) {
	msg := &domain.Message{
		ID:        domain.MessageID(uuid.New()),
		RoomID:    roomID,
		UserID:    userID,
		Username:  username,
		Content:   content,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.repo.Save(ctx, msg); err != nil {
		return nil, err
	}

	s.logger.Info(
		"Message saved",
		zap.String("message_id", msg.ID.String()),
		zap.String("room_id", roomID.String()),
	)
	return msg, nil
}

func (s *MessageService) GetMessagesByRoom(ctx context.Context, roomID domain.RoomID, limit int) ([]*domain.Message, error) {
	return s.repo.GetByRoom(ctx, roomID, limit)
}
