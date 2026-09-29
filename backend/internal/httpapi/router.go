package httpapi

import (
	"net/http"

	"chatgo/backend/internal/auth"
	"chatgo/backend/internal/middleware"
	"chatgo/backend/internal/repository"
	"chatgo/backend/internal/ws"
)

type Deps struct {
	Tokens *auth.TokenManager
	Users  *repository.UserRepository
	Rooms  *repository.RoomRepository
	Msgs   *repository.MessageRepository
	WS     *ws.Service
}

func NewRouter(d Deps) http.Handler {
	authH := NewAuthHandler(d.Users, d.Tokens)
	roomH := NewRoomHandler(d.Rooms)
	msgH := NewMessageHandler(d.Rooms, d.Msgs)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("POST /api/auth/register", authH.Register)
	mux.HandleFunc("POST /api/auth/login", authH.Login)

	mux.HandleFunc("GET /ws", d.WS.ServeHTTP)

	protected := http.NewServeMux()
	protected.HandleFunc("GET /api/me", withUser(authH.Me))

	protected.HandleFunc("GET /api/rooms", withUser(roomH.List))
	protected.HandleFunc("POST /api/rooms", withUser(roomH.Create))
	protected.HandleFunc("GET /api/rooms/{id}", withUserAndID(roomH.Get))
	protected.HandleFunc("PATCH /api/rooms/{id}", withUserAndID(roomH.Update))
	protected.HandleFunc("POST /api/rooms/{id}/join", withUserAndID(roomH.Join))
	protected.HandleFunc("POST /api/rooms/{id}/leave", withUserAndID(roomH.Leave))
	protected.HandleFunc("GET /api/rooms/{id}/messages", withUserAndID(msgH.List))
	protected.HandleFunc("POST /api/rooms/{id}/members/{userId}/role", withUserAndTwoIDs(roomH.SetMemberRole))
	protected.HandleFunc("DELETE /api/rooms/{id}/members/{userId}", withUserAndTwoIDs(roomH.RemoveMember))
	protected.HandleFunc("DELETE /api/messages/{id}", withUserAndID(msgH.Delete))

	mux.Handle("/api/", middleware.RequireAuth(d.Tokens)(protected))

	return middleware.CORS(mux)
}

func withUser(fn func(http.ResponseWriter, *http.Request, int64)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserID(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		fn(w, r, userID)
	}
}

func withUserAndID(fn func(http.ResponseWriter, *http.Request, int64, int64)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserID(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		id, err := parseID(r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		fn(w, r, userID, id)
	}
}

func withUserAndTwoIDs(fn func(http.ResponseWriter, *http.Request, int64, int64, int64)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserID(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		id, err := parseID(r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		otherID, err := parseID(r.PathValue("userId"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid userId")
			return
		}
		fn(w, r, userID, id, otherID)
	}
}
