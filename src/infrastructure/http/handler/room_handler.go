package handler

import (
	"net/http"
	"strings"

	"apiGolan/src/applications/usecase"
	"apiGolan/src/infrastructure/http/middleware"
	jwtutil "apiGolan/src/infrastructure/jwt"
	ws "apiGolan/src/infrastructure/websocket"
)

type RoomHandler struct {
	uc  *usecase.RoomUseCase
	hub *ws.Hub
}

func NewRoomHandler(uc *usecase.RoomUseCase, hub *ws.Hub) *RoomHandler {
	return &RoomHandler{uc: uc, hub: hub}
}

// POST /rooms  → solo host
func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	claims := getClaims(r)
	room, err := h.uc.CreateRoom(claims.UserID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonResponse(w, http.StatusCreated, room)
}

// POST /rooms/{code}/join  → solo participant
func (h *RoomHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	code := extractCode(r.URL.Path, "/rooms/", "/join")
	claims := getClaims(r)

	if err := h.uc.JoinRoom(code, claims.UserID); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"message": "te uniste a la sala"})
}

// PATCH /rooms/{code}/start  → solo host
func (h *RoomHandler) StartSession(w http.ResponseWriter, r *http.Request) {
	code := extractCode(r.URL.Path, "/rooms/", "/start")
	claims := getClaims(r)

	if err := h.uc.StartSession(code, claims.UserID); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.hub.Broadcast(code, ws.Message{
		Event:    "session_started",
		RoomCode: code,
		Payload:  map[string]string{"status": "active"},
	})

	jsonResponse(w, http.StatusOK, map[string]string{"message": "sesión iniciada"})
}

// PATCH /rooms/{code}/end  → solo host
func (h *RoomHandler) EndSession(w http.ResponseWriter, r *http.Request) {
	code := extractCode(r.URL.Path, "/rooms/", "/end")
	claims := getClaims(r)

	if err := h.uc.EndSession(code, claims.UserID); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.hub.Broadcast(code, ws.Message{
		Event:    "session_ended",
		RoomCode: code,
		Payload:  map[string]string{"status": "finished"},
	})

	jsonResponse(w, http.StatusOK, map[string]string{"message": "sesión finalizada"})
}

// GET /rooms/{code}
func (h *RoomHandler) GetRoom(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/rooms/")
	room, err := h.uc.GetRoom(code)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, http.StatusOK, room)
}

// helper: extrae el código de la URL  /rooms/{code}/start → code
func extractCode(path, prefix, suffix string) string {
	s := strings.TrimPrefix(path, prefix)
	return strings.TrimSuffix(s, suffix)
}

// helper: obtiene los claims del contexto (puesto por el middleware)
func getClaims(r *http.Request) *jwtutil.Claims {
	return r.Context().Value(middleware.UserClaimsKey).(*jwtutil.Claims)
}
