package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yohagos/collab-api/internal/domain"
)

type messageRepository struct {
	db *pgxpool.Pool
}

func NewMessageRepository(db *pgxpool.Pool) domain.MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Save(ctx context.Context, msg *domain.Message) error {
	query := `
		INSERT INTO messages (id, room_id, user_id, username, content, type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(ctx, query, msg.ID, msg.RoomID, msg.UserID, msg.Username, msg.Content, msg.Type, msg.CreatedAt)
	return err
}

func (r *messageRepository) GetByRoom(ctx context.Context, roomID domain.RoomID, limit int) ([]*domain.Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := `
		SELECT id, room_id, user_id, username, content, type, created_at
		FROM messages
		WHERE room_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, roomID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []*domain.Message
	for rows.Next() {
		var msg domain.Message
		err := rows.Scan(
			&msg.ID,
			&msg.RoomID,
			&msg.UserID,
			&msg.Username,
			&msg.Content,
			&msg.Type,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, &msg)
	}

	return msgs, nil
}
