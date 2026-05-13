package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yohagos/collab-api/internal/domain"
)

type roomRepository struct {
	db *pgxpool.Pool
}

func NewRoomRepository(db *pgxpool.Pool) domain.RoomRepository {
	return &roomRepository{db: db}
}

func (r *roomRepository) Create(ctx context.Context, room *domain.Room) error {
	query := `
		INSERT INTO rooms (id, name, description, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(
		ctx, 
		query,
		room.ID,
		room.Name,
		room.Description,
		room.CreatedBy,
		room.CreatedAt,
		room.UpdatedAt,
	)
	
	return err
}

func (r *roomRepository) GetByID(ctx context.Context, id domain.RoomID) (*domain.Room, error) {
	query := `
		SELECT id, name, description, created_by, created_at, updated_at
		FROM rooms
		WHERE id = $1
	`

	var room domain.Room
	err := r.db.QueryRow(ctx, query, id).Scan(
		&room.ID,
		&room.Name,
		&room.Description,
		&room.CreatedBy,
		&room.CreatedAt,
		&room.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrRoomNotFound
		}
		return nil, err
	}
	
	return &room, nil
}

func (r *roomRepository) List(ctx context.Context, limit int) ([]*domain.Room, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := `
		SELECT id, name, description, created_by, created_at, updated_at
		FROM rooms
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []*domain.Room
	for rows.Next() {
		var room domain.Room
		err := rows.Scan(
			&room.ID,
			&room.Name,
			&room.Description,
			&room.CreatedBy,
			&room.CreatedAt,
			&room.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, &room)
	}
	return rooms, nil
}
