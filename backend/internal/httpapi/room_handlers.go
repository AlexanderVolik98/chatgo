package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"chatgo/backend/internal/models"
	"chatgo/backend/internal/repository"
)

type RoomHandler struct {
	rooms *repository.RoomRepository
}

func NewRoomHandler(rooms *repository.RoomRepository) *RoomHandler {
	return &RoomHandler{rooms: rooms}
}

type createRoomRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPrivate   bool   `json:"is_private"`
}

func (h *RoomHandler) List(w http.ResponseWriter, r *http.Request, userID int64) {
	rooms, err := h.rooms.ListVisibleTo(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list rooms")
		return
	}
	if rooms == nil {
		rooms = []models.Room{}
	}
	writeJSON(w, http.StatusOK, rooms)
}

func (h *RoomHandler) Create(w http.ResponseWriter, r *http.Request, userID int64) {
	var req createRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if len(req.Name) < 2 {
		writeError(w, http.StatusBadRequest, "room name must be at least 2 characters")
		return
	}

	room, err := h.rooms.Create(r.Context(), req.Name, req.Description, req.IsPrivate, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create room")
		return
	}
	writeJSON(w, http.StatusCreated, room)
}

type roomDetail struct {
	models.Room
	Members       []models.RoomMember          `json:"members"`
	MembersByRole map[models.RoomRole][]string `json:"members_by_role"`
}

func groupMembersByRole(members []models.RoomMember) map[models.RoomRole][]string {
	membersByRole := make(map[models.RoomRole][]string)

	for _, member := range members {
		membersByRole[member.Role] = append(membersByRole[member.Role], member.Username)
	}

	return membersByRole
}

func (h *RoomHandler) Get(w http.ResponseWriter, r *http.Request, userID, roomID int64) {
	room, err := h.rooms.GetByID(r.Context(), roomID)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "room not found")
		return
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load room")
		return
	}

	if room.IsPrivate {
		if _, err := h.rooms.GetMember(r.Context(), roomID, userID); errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusForbidden, "not a member of this private room")
			return
		}
	}

	members, err := h.rooms.ListMembers(r.Context(), roomID)

	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load members")
		return
	}

	grouped := groupMembersByRole(members)

	writeJSON(w, http.StatusOK, roomDetail{Room: *room, Members: members, MembersByRole: grouped})
}

type updateRoomRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *RoomHandler) Update(w http.ResponseWriter, r *http.Request, userID, roomID int64) {
	member, err := h.requireAdmin(r, userID, roomID)
	if err != nil {
		writeRepositoryError(w, err)
		return
	}
	_ = member

	var req updateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if len(req.Name) < 2 {
		writeError(w, http.StatusBadRequest, "room name must be at least 2 characters")
		return
	}

	if err := h.rooms.Update(r.Context(), roomID, req.Name, req.Description); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update room")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *RoomHandler) Join(w http.ResponseWriter, r *http.Request, userID, roomID int64) {
	room, err := h.rooms.GetByID(r.Context(), roomID)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "room not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load room")
		return
	}
	if room.IsPrivate {
		writeError(w, http.StatusForbidden, "this room is private, ask an admin to add you")
		return
	}

	if err := h.rooms.AddMember(r.Context(), roomID, userID, models.RoleMember); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to join room")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "joined"})
}

func (h *RoomHandler) Leave(w http.ResponseWriter, r *http.Request, userID, roomID int64) {
	member, err := h.rooms.GetMember(r.Context(), roomID, userID)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not a member of this room")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load membership")
		return
	}
	if member.Role == models.RoleOwner {
		writeError(w, http.StatusForbidden, "the owner cannot leave the room; delete it or transfer ownership first")
		return
	}

	if err := h.rooms.RemoveMember(r.Context(), roomID, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to leave room")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "left"})
}

type setRoleRequest struct {
	Role string `json:"role"`
}

func (h *RoomHandler) SetMemberRole(w http.ResponseWriter, r *http.Request, actorID, roomID, targetID int64) {
	actor, err := h.requireAdmin(r, actorID, roomID)
	if err != nil {
		writeRepositoryError(w, err)
		return
	}

	var req setRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	newRole := models.RoomRole(req.Role)
	if newRole != models.RoleAdmin && newRole != models.RoleMember {
		writeError(w, http.StatusBadRequest, "role must be 'admin' or 'member'")
		return
	}

	target, err := h.rooms.GetMember(r.Context(), roomID, targetID)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "target user is not a member of this room")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load target member")
		return
	}

	if target.Role == models.RoleOwner {
		writeError(w, http.StatusForbidden, "changing the owner role is prohibited")
		return
	}

	if target.Role == models.RoleAdmin && actor.Role == models.RoleAdmin {
		writeError(w, http.StatusForbidden, "an admin cannot change the admin role")
		return
	}

	if err := h.rooms.SetMemberRole(r.Context(), roomID, targetID, newRole); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update role")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *RoomHandler) RemoveMember(w http.ResponseWriter, r *http.Request, actorID, roomID, targetID int64) {
	actor, err := h.requireAdmin(r, actorID, roomID)
	if err != nil {
		writeRepositoryError(w, err)
		return
	}

	target, err := h.rooms.GetMember(r.Context(), roomID, targetID)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "target user is not a member of this room")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load target member")
		return
	}

	if target.Role == models.RoleOwner {
		writeError(w, http.StatusForbidden, "remove the owner role is prohibited")
		return
	}

	if target.Role == models.RoleAdmin && actor.Role == models.RoleAdmin {
		writeError(w, http.StatusForbidden, "an admin cannot remove the admin role")
		return
	}

	if err := h.rooms.RemoveMember(r.Context(), roomID, targetID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to remove member")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

// requireAdmin loads the acting user's membership and ensures it is at least admin.
func (h *RoomHandler) requireAdmin(r *http.Request, userID, roomID int64) (*models.RoomMember, error) {
	member, err := h.rooms.GetMember(r.Context(), roomID, userID)
	if err != nil {
		return nil, err
	}
	if member.Role != models.RoleOwner && member.Role != models.RoleAdmin {
		return nil, repository.ErrForbidden
	}
	return member, nil
}

func writeRepositoryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, "not a member of this room")
	case errors.Is(err, repository.ErrForbidden):
		writeError(w, http.StatusForbidden, "admin privileges required")
	default:
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

func parseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
