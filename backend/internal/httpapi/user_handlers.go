package httpapi

import (
	"chatgo/backend/internal/models"
	"chatgo/backend/internal/repository"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type UserRepository interface {
	Create(ctx context.Context, username, email, passwordHash string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id int64) (*models.User, error)
	UpdateUsername(ctx context.Context, id int64, username string) (*models.User, error)
}

type updateUserRequest struct {
	Name string `json:"name"`
}

type UserHandler struct {
	users UserRepository
}

type UserUpdateResponse struct {
	User models.User `json:"user"`
}

func NewUserHandler(users UserRepository) *UserHandler {
	return &UserHandler{users: users}
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request, userID int64) {
	user, err := h.users.GetByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (u *UserHandler) Update(w http.ResponseWriter, r *http.Request, userID int64) {
	var req updateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if len(req.Name) < 2 {
		writeError(w, http.StatusBadRequest, "username must be at least 2 characters")
		return
	}

	user, err := u.users.UpdateUsername(r.Context(), userID, req.Name)

	if errors.Is(err, repository.ErrAlreadyExists) {
		writeError(w, http.StatusConflict, "username already exists")
		return
	}

	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to rename user")
		return
	}

	writeJSON(w, http.StatusOK, UserUpdateResponse{User: *user})
}
