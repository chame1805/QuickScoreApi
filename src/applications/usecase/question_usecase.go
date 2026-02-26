package usecase

import (
	"apiGolan/src/core"
	"apiGolan/src/domain"
)

// QuestionUseCase orquesta las operaciones de preguntas y respuestas
type QuestionUseCase struct {
	questionService *core.QuestionService
}

func NewQuestionUseCase(questionService *core.QuestionService) *QuestionUseCase {
	return &QuestionUseCase{questionService: questionService}
}

// LaunchQuestionInput son los datos que llegan cuando el host lanza una pregunta
type LaunchQuestionInput struct {
	Text          string `json:"text"`
	CorrectAnswer string `json:"correct_answer"`
	Points        int    `json:"points"`
	RoomCode      string `json:"-"` // se toma de la URL
	HostID        int    `json:"-"` // se toma del token
}

// LaunchQuestionOutput es lo que se devuelve al lanzar una pregunta
// No incluye la respuesta correcta para no exponerla al cliente
type LaunchQuestionOutput struct {
	ID       int                    `json:"id"`
	RoomID   int                    `json:"room_id"`
	Text     string                 `json:"text"`
	Points   int                    `json:"points"`
	Status   domain.QuestionStatus  `json:"status"`
}

// SubmitAnswerInput son los datos que envía un participante al responder
type SubmitAnswerInput struct {
	QuestionID int    `json:"question_id"`
	Answer     string `json:"answer"`
	RoomCode   string `json:"-"`
	UserID     int    `json:"-"`
}

// SubmitAnswerOutput es el resultado de evaluar la respuesta
type SubmitAnswerOutput struct {
	IsCorrect    bool   `json:"is_correct"`
	PointsEarned int    `json:"points_earned"`
	Message      string `json:"message"`
}

func (uc *QuestionUseCase) LaunchQuestion(input LaunchQuestionInput) (*LaunchQuestionOutput, error) {
	q, err := uc.questionService.LaunchQuestion(input.RoomCode, input.HostID, input.Text, input.CorrectAnswer, input.Points)
	if err != nil {
		return nil, err
	}
	return &LaunchQuestionOutput{
		ID:     q.ID,
		RoomID: q.RoomID,
		Text:   q.Text,
		Points: q.Points,
		Status: q.Status,
	}, nil
}

func (uc *QuestionUseCase) CloseQuestion(roomCode string, hostID, questionID int) error {
	return uc.questionService.CloseQuestion(roomCode, hostID, questionID)
}

func (uc *QuestionUseCase) SubmitAnswer(input SubmitAnswerInput) (*SubmitAnswerOutput, error) {
	isCorrect, points, err := uc.questionService.SubmitAnswer(input.RoomCode, input.UserID, input.QuestionID, input.Answer)
	if err != nil {
		return nil, err
	}

	msg := "Respuesta incorrecta"
	if isCorrect {
		msg = "¡Correcto! Ganaste puntos"
	}

	return &SubmitAnswerOutput{
		IsCorrect:    isCorrect,
		PointsEarned: points,
		Message:      msg,
	}, nil
}

func (uc *QuestionUseCase) GetCurrentQuestion(roomCode string) (*LaunchQuestionOutput, error) {
	q, err := uc.questionService.GetCurrentQuestion(roomCode)
	if err != nil {
		return nil, err
	}
	if q == nil {
		return nil, nil
	}
	return &LaunchQuestionOutput{
		ID:     q.ID,
		RoomID: q.RoomID,
		Text:   q.Text,
		Points: q.Points,
		Status: q.Status,
	}, nil
}

func (uc *QuestionUseCase) GetAnswers(roomCode string, hostID, questionID int) ([]domain.Answer, error) {
	return uc.questionService.GetAnswers(roomCode, hostID, questionID)
}
