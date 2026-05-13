package application

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yohagos/collab-api/internal/domain"
	"go.uber.org/zap"
)

type MockRoomRepository struct {
	mock.Mock
}

func (m *MockRoomRepository) Create(ctx context.Context, room *domain.Room) error {
	args := m.Called(ctx, room)
	return args.Error(0)
}

func (m *MockRoomRepository) GetByID(ctx context.Context, id domain.RoomID) (*domain.Room, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Room), args.Error(1)
}

func (m *MockRoomRepository) List(ctx context.Context, limit int) ([]*domain.Room, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]*domain.Room), args.Error(1)
}

func TestRoomService_CreateRoom(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(MockRoomRepository)
	service := NewRoomService(mockRepo, logger)

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Room")).Return(nil)

	room, err := service.CreateRoom(context.Background(), "Team Meeting", "Daily Sync", domain.UserID{})

	assert.NoError(t, err)
	assert.NotNil(t, room)
	assert.Equal(t, "Team Meeting", room.Name)
	mockRepo.AssertExpectations(t)
}
