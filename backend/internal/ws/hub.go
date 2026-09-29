package ws

import "sync"

// Hub keeps track of which clients are subscribed to which rooms and
// fans out messages to them. One instance is shared by all connections,
// each of which runs in its own goroutine — so every method here needs
// to be safe for concurrent use.
type Hub struct {
	mu sync.RWMutex

	// Keyed by *Client: pointer identity means "this exact connection".
	rooms map[int64]map[*Client]bool
}

func NewHub() *Hub {
	return &Hub{rooms: make(map[int64]map[*Client]bool)}
}

// Join subscribes client c to room roomID.
func (h *Hub) Join(roomID int64, c *Client) {
	h.mu.Lock()
	if len(h.rooms[roomID]) == 0 {
		h.rooms[roomID] = make(map[*Client]bool)
	}

	h.rooms[roomID][c] = true

	h.mu.Unlock()
}

// Leave unsubscribes client c from room roomID.
func (h *Hub) Leave(roomID int64, c *Client) {
	h.mu.Lock()
	if h.rooms[roomID] != nil {
		delete(h.rooms[roomID], c)
	}

	if len(h.rooms[roomID]) == 0 {
		delete(h.rooms, roomID)
	}

	h.mu.Unlock()
}

// LeaveAll removes a client from every room it was subscribed to, used on disconnect.
func (h *Hub) LeaveAll(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for roomID, cRoom := range h.rooms {
		delete(cRoom, c)
		if len(cRoom) == 0 {
			delete(h.rooms, roomID)
		}
	}
}

// Broadcast sends payload to every client subscribed to room roomID.
func (h *Hub) Broadcast(roomID int64, payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for c := range h.rooms[roomID] {
		// Drop the message for a slow client instead of blocking the broadcast.
		select {
		case c.send <- payload:
		default:
		}
	}
}
