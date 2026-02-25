package domain

import "time"

// RoomStatus representa el estado actual de una sala
type RoomStatus string

const (
	RoomStatusWaiting  RoomStatus = "waiting"
	RoomStatusActive   RoomStatus = "active"
	RoomStatusFinished RoomStatus = "finished"
)

// Room representa una sala de competencia creada por un host
type Room struct {
	ID        int        `json:"id"`
	Code      string     `json:"code"`       // código corto tipo ABC123
	HostID    int        `json:"host_id"`
	Status    RoomStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
}
