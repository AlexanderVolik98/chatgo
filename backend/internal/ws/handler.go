package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"chatgo/backend/internal/auth"
	"chatgo/backend/internal/models"
	"chatgo/backend/internal/repository"
)

type inboundMessage struct {
	Type    string `json:"type"`
	RoomID  int64  `json:"room_id"`
	Content string `json:"content"`
}

type outboundMessage struct {
	Type    string          `json:"type"`
	RoomID  int64           `json:"room_id,omitempty"`
	Message *models.Message `json:"message,omitempty"`
	Error   string          `json:"error,omitempty"`
}

type Service struct {
	hub     *Hub
	rooms   *repository.RoomRepository
	msgs    *repository.MessageRepository
	tokens  *auth.TokenManager
	upgrade websocket.Upgrader
}

func NewService(hub *Hub, rooms *repository.RoomRepository, msgs *repository.MessageRepository, tokens *auth.TokenManager) *Service {
	return &Service{
		hub:    hub,
		rooms:  rooms,
		msgs:   msgs,
		tokens: tokens,
		upgrade: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

// ServeHTTP upgrades the connection to a websocket. Authentication is done via
// a `token` query parameter since browsers cannot set custom headers on the
// native WebSocket handshake.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	claims, err := s.tokens.Parse(tokenStr)
	if err != nil {
		http.Error(w, "invalid or missing token", http.StatusUnauthorized)
		return
	}

	conn, err := s.upgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
		return
	}

	client := NewClient(s.hub, conn, claims.UserID, "")
	go client.WritePump()
	go s.readPump(client)
}

func (s *Service) readPump(c *Client) {
	defer func() {
		s.hub.LeaveAll(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	ctx := context.Background()

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		var in inboundMessage
		if err := json.Unmarshal(raw, &in); err != nil {
			s.sendError(c, "invalid message")
			continue
		}

		switch in.Type {
		case "join":
			s.handleJoin(ctx, c, in.RoomID)
		case "leave":
			s.hub.Leave(in.RoomID, c)
			delete(c.rooms, in.RoomID)
		case "message":
			s.handleMessage(ctx, c, in.RoomID, in.Content)
		default:
			s.sendError(c, "unknown message type")
		}
	}
}

func (s *Service) handleJoin(ctx context.Context, c *Client, roomID int64) {
	member, err := s.rooms.GetMember(ctx, roomID, c.UserID)
	if err != nil {
		s.sendError(c, "not a member of this room")
		return
	}
	if c.Username == "" {
		c.Username = member.Username
	}
	c.rooms[roomID] = true
	s.hub.Join(roomID, c)
}

func (s *Service) handleMessage(ctx context.Context, c *Client, roomID int64, content string) {
	if content == "" {
		return
	}
	if !c.rooms[roomID] {
		s.sendError(c, "join the room before sending messages")
		return
	}

	msg, err := s.msgs.Create(ctx, roomID, c.UserID, content)
	if err != nil {
		s.sendError(c, "failed to save message")
		return
	}
	msg.Username = c.Username

	payload, err := json.Marshal(outboundMessage{Type: "message", RoomID: roomID, Message: msg})
	if err != nil {
		return
	}
	s.hub.Broadcast(roomID, payload)
}

func (s *Service) sendError(c *Client, text string) {
	payload, _ := json.Marshal(outboundMessage{Type: "error", Error: text})
	select {
	case c.send <- payload:
	default:
	}
}
