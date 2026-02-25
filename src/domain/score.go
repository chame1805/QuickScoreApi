package domain

import "time"

// Participant representa la relación entre un usuario y una sala
type Participant struct {
	ID       int       `json:"id"`
	RoomID   int       `json:"room_id"`
	UserID   int       `json:"user_id"`
	JoinedAt time.Time `json:"joined_at"`
}

// Score representa los puntos de un participante dentro de una sala
type Score struct {
	ID        int       `json:"id"`
	RoomID    int       `json:"room_id"`
	UserID    int       `json:"user_id"`
	Points    int       `json:"points"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RankingEntry representa una entrada del ranking con los datos del usuario
type RankingEntry struct {
	UserID   int    `json:"user_id"`
	UserName string `json:"user_name"`
	Points   int    `json:"points"`
	Position int    `json:"position"`
}
