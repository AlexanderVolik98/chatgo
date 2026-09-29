package httpapi

import (
	"context"
	"time"

	"chatgo/backend/internal/models"
	"chatgo/backend/internal/repository"
)

// mockUserRepository is an in-memory UserRepository for tests.
type mockUserRepository struct {
	byEmail map[string]*models.User
	nextID  int64
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{byEmail: make(map[string]*models.User)}
}

func (s *mockUserRepository) Create(ctx context.Context, username, email, passwordHash string) (*models.User, error) {
	if _, exists := s.byEmail[email]; exists {
		return nil, repository.ErrAlreadyExists
	}

	s.nextID++
	u := &models.User{
		ID:           s.nextID,
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}
	s.byEmail[email] = u
	return u, nil
}

func (s *mockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	u, ok := s.byEmail[email]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return u, nil
}

func (s *mockUserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	for _, u := range s.byEmail {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, repository.ErrNotFound
}
