package domain

import "time"

// QuestionStatus indica si una pregunta está abierta o cerrada para respuestas
type QuestionStatus string

const (
	QuestionStatusOpen   QuestionStatus = "open"
	QuestionStatusClosed QuestionStatus = "closed"
)

// Question representa una pregunta lanzada por el host dentro de una sala activa
type Question struct {
	ID            int            `json:"id"`
	RoomID        int            `json:"room_id"`
	Text          string         `json:"text"`
	CorrectAnswer string         `json:"-"`          // nunca se expone al cliente
	Points        int            `json:"points"`     // puntos que vale la pregunta
	Status        QuestionStatus `json:"status"`
	CreatedAt     time.Time      `json:"created_at"`
}

// Answer representa la respuesta enviada por un participante a una pregunta
type Answer struct {
	ID         int       `json:"id"`
	QuestionID int       `json:"question_id"`
	UserID     int       `json:"user_id"`
	Text       string    `json:"answer"`
	IsCorrect  bool      `json:"is_correct"`
	AnsweredAt time.Time `json:"answered_at"`
}

// QuestionRepository define las operaciones de persistencia para preguntas
type QuestionRepository interface {
	Create(q *Question) error
	FindByID(id int) (*Question, error)
	FindOpenByRoom(roomID int) (*Question, error) // pregunta actualmente abierta
	CloseQuestion(id int) error
	FindByRoom(roomID int) ([]Question, error)
}

// AnswerRepository define las operaciones de persistencia para respuestas
type AnswerRepository interface {
	Create(a *Answer) error
	HasAnswered(questionID, userID int) (bool, error)
	FindByQuestion(questionID int) ([]Answer, error)
}
