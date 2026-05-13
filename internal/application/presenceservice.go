package application

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yohagos/collab-api/internal/domain"
	"go.uber.org/zap"
)

type PresenceService struct {
	rds    *redis.Client
	logger *zap.Logger
}

func NewPresenceService(redisClient *redis.Client, logger *zap.Logger) *PresenceService {
	return &PresenceService{
		rds:    redisClient,
		logger: logger,
	}
}

func (s *PresenceService) UserJoined(
	ctx context.Context,
	roomID domain.RoomID,
	userID domain.UserID,
	username string,
) error {
	key := "room:" + roomID.String() + ":users"

	member := map[string]interface{}{
		"username":  username,
		"last_seen": time.Now().Unix(),
	}

	return s.rds.HSet(ctx, key, userID.String(), member).Err()
}

func (s *PresenceService) GetOnlineUsers(
	ctx context.Context,
	roomID domain.RoomID,
) ([]domain.RoomMember, error) {
	key := "room:" + roomID.String() + ":users"

	result, err := s.rds.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var members []domain.RoomMember
	for range result {
		members = append(members, domain.RoomMember{
			UserID:   domain.UserID{},
			Username: "Online User",
			JoinedAt: time.Now(),
			LastSeen: time.Now(),
		})
	}
	return nil, nil
}

func (s *PresenceService) UserLeft(ctx context.Context, roomID domain.RoomID, userID domain.UserID) error {
	key := "room:" + roomID.String() + ":users"
	return s.rds.HDel(ctx, key, userID.String()).Err()
}

func (s *PresenceService) SetUserTyping(ctx context.Context, roomID domain.RoomID, userID domain.UserID, username string) error {
	key := "room:" + roomID.String() + ":typing"

	_, err := s.rds.HSet(ctx, key, userID.String(), username).Result()
	if err != nil {
		return err
	}

	return s.rds.Expire(ctx, key, 3 * time.Second).Err()
}

func (s *PresenceService) GetTypingUsers(ctx context.Context, roomID domain.RoomID) ([]string, error) {
	key := "room:" + roomID.String() + ":typing"

	usernames, err := s.rds.HVals(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	return usernames, nil
}
