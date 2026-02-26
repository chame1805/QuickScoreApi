package domain

// UserRepository define las operaciones de persistencia para usuarios.
// Esta interfaz vive en el dominio; la implementación concreta está en infrastructure.
type UserRepository interface {
	Create(user *User) error
	FindByEmail(email string) (*User, error)
	FindByID(id int) (*User, error)
}

// RoomRepository define las operaciones de persistencia para salas.
type RoomRepository interface {
	Create(room *Room) error
	FindByCode(code string) (*Room, error)
	UpdateStatus(code string, status RoomStatus) error
}

// ParticipantRepository define las operaciones de persistencia para participantes.
type ParticipantRepository interface {
	Add(participant *Participant) error
	ExistsInRoom(roomID, userID int) (bool, error)
	FindByRoom(roomID int) ([]Participant, error)
	FindByRoomWithUsers(roomID int) ([]ParticipantWithUser, error) // con datos de usuario
	Remove(roomID, userID int) error                               // expulsar participante
}

// ScoreRepository define las operaciones de persistencia para puntos.
type ScoreRepository interface {
	Upsert(roomID, userID, points int) error          // crea o actualiza el score
	AddPoints(roomID, userID, delta int) error         // suma o resta puntos
	GetRanking(roomID int) ([]RankingEntry, error)     // ranking ordenado por puntos
	ResetPoints(roomID, userID int) error              // resetear puntos de un participante
	ResetAllPoints(roomID int) error                   // resetear todos los puntos de la sala
}

// ParticipantWithUser combina participante y datos del usuario para listados
type ParticipantWithUser struct {
	UserID   int    `json:"user_id"`
	UserName string `json:"user_name"`
	Email    string `json:"email"`
	JoinedAt string `json:"joined_at"`
}
