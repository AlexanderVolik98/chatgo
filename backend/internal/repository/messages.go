package repository

import (
	"context"
	"errors"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"chatgo/backend/internal/models"
)

type MessageRepository struct {
	pool *pgxpool.Pool
}

func NewMessageRepository(pool *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{pool: pool}
}

func (s *MessageRepository) Create(ctx context.Context, roomID, userID int64, content string) (*models.Message, error) {
	m := &models.Message{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO messages (room_id, user_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, room_id, user_id, content, created_at
	`, roomID, userID, content).Scan(&m.ID, &m.RoomID, &m.UserID, &m.Content, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	m.Username = "" // filled in by caller if needed
	return m, nil
}

// ListBefore returns up to `limit` messages in a room, ordered oldest to newest,
// optionally paginating backwards from beforeID (exclusive).
func (s *MessageRepository) ListBefore(ctx context.Context, roomID int64, beforeID int64, limit int) ([]models.Message, error) {
	var (
		rows pgx.Rows
		err  error
	)

	if beforeID == 0 {
		rows, err = s.pool.Query(ctx, `
			SELECT messages.id, room_id, user_id, content, messages.created_at, username
			FROM messages JOIN public.users u on messages.user_id = u.id
			WHERE room_id = $1
			ORDER BY id DESC
			LIMIT $2
		`, roomID, limit)
	} else {
		rows, err = s.pool.Query(ctx, `
		SELECT messages.id, room_id, user_id, content, messages.created_at, username
			FROM messages
			JOIN users ON messages.user_id = users.id
			WHERE room_id = $1 AND id < $2
			ORDER BY id DESC
			LIMIT $3
		`, roomID, beforeID, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.RoomID, &m.UserID, &m.Content, &m.CreatedAt, &m.Username); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Query returns newest first; callers expect oldest first.
	slices.Reverse(msgs)

	return msgs, nil
}

func (s *MessageRepository) GetByID(ctx context.Context, id int64) (*models.Message, error) {
	m := &models.Message{}
	err := s.pool.QueryRow(ctx, `
		SELECT m.id, m.room_id, m.user_id, u.username, m.content, m.created_at
		FROM messages m JOIN users u ON u.id = m.user_id
		WHERE m.id = $1
	`, id).Scan(&m.ID, &m.RoomID, &m.UserID, &m.Username, &m.Content, &m.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *MessageRepository) Delete(ctx context.Context, id int64) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM messages WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
