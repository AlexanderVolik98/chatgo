package models

import "time"

type RoomRole string

const (
	RoleOwner  RoomRole = "owner"
	RoleAdmin  RoomRole = "admin"
	RoleMember RoomRole = "member"
)

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type Room struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     int64     `json:"owner_id"`
	IsPrivate   bool      `json:"is_private"`
	CreatedAt   time.Time `json:"created_at"`
}

type RoomMember struct {
	RoomID   int64     `json:"room_id"`
	UserID   int64     `json:"user_id"`
	Username string    `json:"username"`
	Role     RoomRole  `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type Message struct {
	ID        int64     `json:"id"`
	RoomID    int64     `json:"room_id"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
