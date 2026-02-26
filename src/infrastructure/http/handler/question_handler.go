package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"apiGolan/src/applications/usecase"
	ws "apiGolan/src/infrastructure/websocket"
)

type QuestionHandler struct {
	uc  *usecase.QuestionUseCase
	hub *ws.Hub
}

func NewQuestionHandler(uc *usecase.QuestionUseCase, hub *ws.Hub) *QuestionHandler {
	return &QuestionHandler{uc: uc, hub: hub}
}

// LaunchQuestion godoc
// @Summary Host lanza una pregunta a la sala
// @Tags questions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param code path string true "Código de sala"
// @Param body body usecase.LaunchQuestionInput true "Datos de la pregunta"
// @Success 201 {object} usecase.LaunchQuestionOutput
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /rooms/{code}/questions [post]
func (h *QuestionHandler) LaunchQuestion(w http.ResponseWriter, r *http.Request) {
	code := extractRoomCode(r.URL.Path, "/questions")
	claims := getClaims(r)

	var input usecase.LaunchQuestionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "cuerpo de la petición inválido", http.StatusBadRequest)
		return
	}
	input.RoomCode = code
	input.HostID = claims.UserID

	output, err := h.uc.LaunchQuestion(input)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Notificar a todos en la sala que hay una nueva pregunta
	h.hub.Broadcast(code, ws.Message{
		Event:    "new_question",
		RoomCode: code,
		Payload:  output,
	})

	jsonResponse(w, http.StatusCreated, output)
}

// CloseQuestion godoc
// @Summary Host cierra la pregunta activa
// @Tags questions
// @Produce json
// @Security BearerAuth
// @Param code path string true "Código de sala"
// @Param question_id path int true "ID de pregunta"
// @Success 200 {object} map[string]string
// @Router /rooms/{code}/questions/{question_id}/close [patch]
func (h *QuestionHandler) CloseQuestion(w http.ResponseWriter, r *http.Request) {
	// Path: /rooms/{code}/questions/{question_id}/close
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	// parts = ["rooms", code, "questions", id, "close"]
	if len(parts) < 5 {
		jsonError(w, "ruta inválida", http.StatusBadRequest)
		return
	}
	code := parts[1]
	questionID, err := strconv.Atoi(parts[3])
	if err != nil {
		jsonError(w, "id de pregunta inválido", http.StatusBadRequest)
		return
	}
	claims := getClaims(r)

	if err := h.uc.CloseQuestion(code, claims.UserID, questionID); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.hub.Broadcast(code, ws.Message{
		Event:    "question_closed",
		RoomCode: code,
		Payload:  map[string]int{"question_id": questionID},
	})

	jsonResponse(w, http.StatusOK, map[string]string{"message": "pregunta cerrada"})
}

// GetCurrentQuestion godoc
// @Summary Obtener la pregunta activa de una sala
// @Tags questions
// @Produce json
// @Security BearerAuth
// @Param code path string true "Código de sala"
// @Success 200 {object} usecase.LaunchQuestionOutput
// @Success 204 "Sin pregunta activa"
// @Router /rooms/{code}/questions/current [get]
func (h *QuestionHandler) GetCurrentQuestion(w http.ResponseWriter, r *http.Request) {
	code := extractRoomCode(r.URL.Path, "/questions/current")

	q, err := h.uc.GetCurrentQuestion(code)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if q == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	jsonResponse(w, http.StatusOK, q)
}

// SubmitAnswer godoc
// @Summary Participante envía su respuesta
// @Tags questions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param code path string true "Código de sala"
// @Param body body usecase.SubmitAnswerInput true "Respuesta"
// @Success 200 {object} usecase.SubmitAnswerOutput
// @Failure 400 {object} map[string]string
// @Router /rooms/{code}/answer [post]
func (h *QuestionHandler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	code := extractRoomCode(r.URL.Path, "/answer")
	claims := getClaims(r)

	var input usecase.SubmitAnswerInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "cuerpo de la petición inválido", http.StatusBadRequest)
		return
	}
	input.RoomCode = code
	input.UserID = claims.UserID

	output, err := h.uc.SubmitAnswer(input)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Si fue correcta, hacer broadcast del ranking actualizado
	if output.IsCorrect {
		// Importar score use case sería un ciclo de dependencia,
		// así que el broadcast del ranking lo hacemos desde el score handler
		// via ws event genérico de "answer_correct"
		h.hub.Broadcast(code, ws.Message{
			Event:    "answer_correct",
			RoomCode: code,
			Payload: map[string]interface{}{
				"user_id":       claims.UserID,
				"question_id":   input.QuestionID,
				"points_earned": output.PointsEarned,
			},
		})
	}

	jsonResponse(w, http.StatusOK, output)
}

// GetAnswers godoc
// @Summary Host obtiene todas las respuestas de una pregunta
// @Tags questions
// @Produce json
// @Security BearerAuth
// @Param code path string true "Código de sala"
// @Param question_id path int true "ID de pregunta"
// @Success 200 {array} domain.Answer
// @Router /rooms/{code}/questions/{question_id}/answers [get]
func (h *QuestionHandler) GetAnswers(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	// parts = ["rooms", code, "questions", id, "answers"]
	if len(parts) < 5 {
		jsonError(w, "ruta inválida", http.StatusBadRequest)
		return
	}
	code := parts[1]
	questionID, err := strconv.Atoi(parts[3])
	if err != nil {
		jsonError(w, "id de pregunta inválido", http.StatusBadRequest)
		return
	}
	claims := getClaims(r)

	answers, err := h.uc.GetAnswers(code, claims.UserID, questionID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusForbidden)
		return
	}
	jsonResponse(w, http.StatusOK, answers)
}

// extractRoomCode extrae el código de sala de paths tipo /rooms/{code}/questions
func extractRoomCode(path, suffix string) string {
	s := strings.TrimPrefix(path, "/rooms/")
	return strings.TrimSuffix(s, suffix)
}
