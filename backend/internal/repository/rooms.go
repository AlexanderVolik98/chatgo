package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"chatgo/backend/internal/models"
)

type RoomRepository struct {
	pool *pgxpool.Pool
}

func NewRoomRepository(pool *pgxpool.Pool) *RoomRepository {
	return &RoomRepository{pool: pool}
}

// Create creates a room and makes the creator its owner in one transaction.
func (s *RoomRepository) Create(ctx context.Context, name, description string, isPrivate bool, ownerID int64) (*models.Room, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	r := &models.Room{}
	err = tx.QueryRow(ctx, `
		INSERT INTO rooms (name, description, owner_id, is_private)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, description, owner_id, is_private, created_at
	`, name, description, ownerID, isPrivate).Scan(&r.ID, &r.Name, &r.Description, &r.OwnerID, &r.IsPrivate, &r.CreatedAt)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO room_members (room_id, user_id, role) VALUES ($1, $2, 'owner')
	`, r.ID, ownerID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r, nil
}

func (s *RoomRepository) GetByID(ctx context.Context, id int64) (*models.Room, error) {
	r := &models.Room{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, description, owner_id, is_private, created_at FROM rooms WHERE id = $1
	`, id).Scan(&r.ID, &r.Name, &r.Description, &r.OwnerID, &r.IsPrivate, &r.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return r, nil
}

// ListVisibleTo returns public rooms plus private rooms the user is a member of.
func (s *RoomRepository) ListVisibleTo(ctx context.Context, userID int64) ([]models.Room, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT r.id, r.name, r.description, r.owner_id, r.is_private, r.created_at
		FROM rooms r
		LEFT JOIN room_members rm ON rm.room_id = r.id AND rm.user_id = $1
		WHERE r.is_private = false OR rm.user_id IS NOT NULL
		ORDER BY r.id DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []models.Room
	for rows.Next() {
		var r models.Room
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.OwnerID, &r.IsPrivate, &r.CreatedAt); err != nil {
			return nil, err
		}
		rooms = append(rooms, r)
	}
	return rooms, rows.Err()
}

func (s *RoomRepository) Update(ctx context.Context, id int64, name, description string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE rooms SET name = $2, description = $3 WHERE id = $1
	`, id, name, description)
	return err
}

func (s *RoomRepository) AddMember(ctx context.Context, roomID, userID int64, role models.RoomRole) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO room_members (room_id, user_id, role) VALUES ($1, $2, $3)
		ON CONFLICT (room_id, user_id) DO NOTHING
	`, roomID, userID, role)
	return err
}

func (s *RoomRepository) RemoveMember(ctx context.Context, roomID, userID int64) error {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM room_members WHERE room_id = $1 AND user_id = $2
	`, roomID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *RoomRepository) SetMemberRole(ctx context.Context, roomID, userID int64, role models.RoomRole) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE room_members SET role = $3 WHERE room_id = $1 AND user_id = $2
	`, roomID, userID, role)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *RoomRepository) GetMember(ctx context.Context, roomID, userID int64) (*models.RoomMember, error) {
	m := &models.RoomMember{}
	err := s.pool.QueryRow(ctx, `
		SELECT rm.room_id, rm.user_id, u.username, rm.role, rm.joined_at
		FROM room_members rm JOIN users u ON u.id = rm.user_id
		WHERE rm.room_id = $1 AND rm.user_id = $2
	`, roomID, userID).Scan(&m.RoomID, &m.UserID, &m.Username, &m.Role, &m.JoinedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *RoomRepository) ListMembers(ctx context.Context, roomID int64) ([]models.RoomMember, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT rm.room_id, rm.user_id, u.username, rm.role, rm.joined_at
		FROM room_members rm JOIN users u ON u.id = rm.user_id
		WHERE rm.room_id = $1
		ORDER BY rm.role, u.username
	`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.RoomMember
	for rows.Next() {
		var m models.RoomMember
		if err := rows.Scan(&m.RoomID, &m.UserID, &m.Username, &m.Role, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}
