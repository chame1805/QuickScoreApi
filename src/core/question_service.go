package core

import (
	"errors"
	"strings"

	"apiGolan/src/domain"
)

// QuestionService contiene la lógica de negocio para preguntas y respuestas
type QuestionService struct {
	questionRepo domain.QuestionRepository
	answerRepo   domain.AnswerRepository
	scoreRepo    domain.ScoreRepository
	roomRepo     domain.RoomRepository
}

func NewQuestionService(
	questionRepo domain.QuestionRepository,
	answerRepo domain.AnswerRepository,
	scoreRepo domain.ScoreRepository,
	roomRepo domain.RoomRepository,
) *QuestionService {
	return &QuestionService{
		questionRepo: questionRepo,
		answerRepo:   answerRepo,
		scoreRepo:    scoreRepo,
		roomRepo:     roomRepo,
	}
}

// LaunchQuestion crea una nueva pregunta y la lanza a la sala (solo host)
func (s *QuestionService) LaunchQuestion(roomCode string, hostID int, text, correctAnswer string, points int) (*domain.Question, error) {
	room, err := s.roomRepo.FindByCode(roomCode)
	if err != nil || room == nil {
		return nil, errors.New("sala no encontrada")
	}
	if room.HostID != hostID {
		return nil, errors.New("solo el host puede lanzar preguntas")
	}
	if room.Status != domain.RoomStatusActive {
		return nil, errors.New("la sesión debe estar activa para lanzar preguntas")
	}
	if text == "" || correctAnswer == "" {
		return nil, errors.New("la pregunta y la respuesta correcta son requeridas")
	}
	if points <= 0 {
		points = 10 // valor por defecto
	}

	// Cerrar pregunta abierta anterior si existe
	existing, _ := s.questionRepo.FindOpenByRoom(room.ID)
	if existing != nil {
		_ = s.questionRepo.CloseQuestion(existing.ID)
	}

	q := &domain.Question{
		RoomID:        room.ID,
		Text:          text,
		CorrectAnswer: correctAnswer,
		Points:        points,
		Status:        domain.QuestionStatusOpen,
	}
	if err := s.questionRepo.Create(q); err != nil {
		return nil, err
	}
	return q, nil
}

// CloseQuestion cierra la pregunta activa (solo host)
func (s *QuestionService) CloseQuestion(roomCode string, hostID, questionID int) error {
	room, err := s.roomRepo.FindByCode(roomCode)
	if err != nil || room == nil {
		return errors.New("sala no encontrada")
	}
	if room.HostID != hostID {
		return errors.New("solo el host puede cerrar preguntas")
	}
	return s.questionRepo.CloseQuestion(questionID)
}

// SubmitAnswer procesa la respuesta de un participante
// Devuelve si fue correcta y los puntos ganados
func (s *QuestionService) SubmitAnswer(roomCode string, userID int, questionID int, answerText string) (bool, int, error) {
	room, err := s.roomRepo.FindByCode(roomCode)
	if err != nil || room == nil {
		return false, 0, errors.New("sala no encontrada")
	}
	if room.Status != domain.RoomStatusActive {
		return false, 0, errors.New("la sesión no está activa")
	}

	question, err := s.questionRepo.FindByID(questionID)
	if err != nil || question == nil {
		return false, 0, errors.New("pregunta no encontrada")
	}
	if question.RoomID != room.ID {
		return false, 0, errors.New("la pregunta no pertenece a esta sala")
	}
	if question.Status != domain.QuestionStatusOpen {
		return false, 0, errors.New("la pregunta ya está cerrada")
	}

	// Verificar que no haya respondido ya
	already, _ := s.answerRepo.HasAnswered(questionID, userID)
	if already {
		return false, 0, errors.New("ya respondiste esta pregunta")
	}

	// Evaluar respuesta (case-insensitive, trimmed)
	isCorrect := strings.EqualFold(
		strings.TrimSpace(answerText),
		strings.TrimSpace(question.CorrectAnswer),
	)

	answer := &domain.Answer{
		QuestionID: questionID,
		UserID:     userID,
		Text:       answerText,
		IsCorrect:  isCorrect,
	}
	if err := s.answerRepo.Create(answer); err != nil {
		return false, 0, err
	}

	pointsEarned := 0
	if isCorrect {
		pointsEarned = question.Points
		if err := s.scoreRepo.AddPoints(room.ID, userID, pointsEarned); err != nil {
			return true, 0, err
		}
	}

	return isCorrect, pointsEarned, nil
}

// GetCurrentQuestion devuelve la pregunta actualmente abierta en una sala
func (s *QuestionService) GetCurrentQuestion(roomCode string) (*domain.Question, error) {
	room, err := s.roomRepo.FindByCode(roomCode)
	if err != nil || room == nil {
		return nil, errors.New("sala no encontrada")
	}
	q, err := s.questionRepo.FindOpenByRoom(room.ID)
	if err != nil {
		return nil, err
	}
	return q, nil // puede ser nil si no hay pregunta activa
}

// GetRoomService agrega ResetPoints al scoreService (host)
func (s *QuestionService) GetAnswers(roomCode string, hostID, questionID int) ([]domain.Answer, error) {
	room, err := s.roomRepo.FindByCode(roomCode)
	if err != nil || room == nil {
		return nil, errors.New("sala no encontrada")
	}
	if room.HostID != hostID {
		return nil, errors.New("solo el host puede ver las respuestas")
	}
	return s.answerRepo.FindByQuestion(questionID)
}
