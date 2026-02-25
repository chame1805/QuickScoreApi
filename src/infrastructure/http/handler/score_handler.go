package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"apiGolan/src/applications/usecase"
	ws "apiGolan/src/infrastructure/websocket"
)

type ScoreHandler struct {
	uc  *usecase.ScoreUseCase
	hub *ws.Hub
}

func NewScoreHandler(uc *usecase.ScoreUseCase, hub *ws.Hub) *ScoreHandler {
	return &ScoreHandler{uc: uc, hub: hub}
}

// POST /rooms/{code}/score  → solo host
// Body: { "target_user_id": 3, "delta": 10 }
func (h *ScoreHandler) AddPoints(w http.ResponseWriter, r *http.Request) {
	code := extractCode(r.URL.Path, "/rooms/", "/score")
	claims := getClaims(r)

	var input usecase.AddPointsInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "cuerpo de la petición inválido", http.StatusBadRequest)
		return
	}

	input.RoomCode = code
	input.RequesterID = claims.UserID

	if err := h.uc.AddPoints(input); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Obtener el ranking actualizado y hacer broadcast a todos en la sala
	ranking, err := h.uc.GetRanking(code)
	if err == nil {
		h.hub.Broadcast(code, ws.Message{
			Event:    "score_update",
			RoomCode: code,
			Payload:  ranking,
		})
	}

	jsonResponse(w, http.StatusOK, map[string]string{"message": "puntos actualizados"})
}

// GET /rooms/{code}/ranking  → todos
func (h *ScoreHandler) GetRanking(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/rooms/")
	code = strings.TrimSuffix(code, "/ranking")

	ranking, err := h.uc.GetRanking(code)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	jsonResponse(w, http.StatusOK, ranking)
}
