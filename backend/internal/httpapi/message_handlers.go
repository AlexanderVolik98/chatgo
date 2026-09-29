package httpapi

import (
	"errors"
	"net/http"
	"strconv"

	"chatgo/backend/internal/models"
	"chatgo/backend/internal/repository"
)

type MessageHandler struct {
	rooms    *repository.RoomRepository
	messages *repository.MessageRepository
}

func NewMessageHandler(rooms *repository.RoomRepository, messages *repository.MessageRepository) *MessageHandler {
	return &MessageHandler{rooms: rooms, messages: messages}
}

func (h *MessageHandler) List(w http.ResponseWriter, r *http.Request, userID, roomID int64) {
	if _, err := h.rooms.GetMember(r.Context(), roomID, userID); errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusForbidden, "join the room to view its messages")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check membership")
		return
	}

	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	var beforeID int64
	if v := r.URL.Query().Get("before"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			beforeID = n
		}
	}

	messages, err := h.messages.ListBefore(r.Context(), roomID, beforeID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load messages")
		return
	}
	if messages == nil {
		messages = []models.Message{}
	}
	writeJSON(w, http.StatusOK, messages)
}

func (h *MessageHandler) Delete(w http.ResponseWriter, r *http.Request, userID, messageID int64) {
	msg, err := h.messages.GetByID(r.Context(), messageID)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "message not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load message")
		return
	}

	if msg.UserID != userID {
		member, err := h.rooms.GetMember(r.Context(), msg.RoomID, userID)
		if err != nil || (member.Role != models.RoleOwner && member.Role != models.RoleAdmin) {
			writeError(w, http.StatusForbidden, "only the author or a room admin can delete this message")
			return
		}
	}

	if err := h.messages.Delete(r.Context(), messageID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete message")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
