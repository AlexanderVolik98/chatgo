package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"chatgo/backend/internal/models"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (s *UserRepository) Create(ctx context.Context, username, email, passwordHash string) (*models.User, error) {
	u := &models.User{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, username, email, password_hash, created_at
	`, username, email, passwordHash).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrAlreadyExists
		}
		return nil, err
	}
	return u, nil
}

func (s *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	u := &models.User{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, username, email, password_hash, created_at FROM users WHERE email = $1
	`, email).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	u := &models.User{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, username, email, password_hash, created_at FROM users WHERE id = $1
	`, id).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserRepository) UpdateUsername(ctx context.Context, id int64, username string) (*models.User, error) {
	u := &models.User{}
	err := s.pool.QueryRow(ctx, `
		UPDATE users
		SET username = $1
		WHERE id = $2
		RETURNING id, username, email, password_hash, created_at
	`, username, id).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return nil, ErrAlreadyExists
	}

	if err != nil {
		return nil, err
	}

	return u, nil
}
