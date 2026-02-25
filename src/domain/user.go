package domain

import "time"

// Role define el tipo de usuario en el sistema
type Role string

const (
RoleHost        Role = "host"
RoleParticipant Role = "participant"
)

// User representa a un usuario del sistema (host o participante)
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // nunca se expone en JSON
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}
