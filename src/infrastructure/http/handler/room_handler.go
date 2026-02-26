package handler

import (
    "encoding/json"
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

// CreateRoom godoc
// @Summary Crear sala
// @Tags rooms
// @Produce json
// @Security BearerAuth
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /rooms [post]
func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
    claims := getClaims(r)
    room, err := h.uc.CreateRoom(claims.UserID)
    if err != nil {
        jsonError(w, err.Error(), http.StatusBadRequest)
        return
    }
    jsonResponse(w, http.StatusCreated, room)
}

// JoinRoom godoc
// @Summary Unirse a sala
// @Tags rooms
// @Produce json
// @Security BearerAuth
// @Param code path string true "Código de sala"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /rooms/{code}/join [post]
func (h *RoomHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
    code := extractCode(r.URL.Path, "/rooms/", "/join")
    claims := getClaims(r)

    if err := h.uc.JoinRoom(code, claims.UserID); err != nil {
        jsonError(w, err.Error(), http.StatusBadRequest)
        return
    }

    jsonResponse(w, http.StatusOK, map[string]string{"message": "te uniste a la sala"})
}

// StartSession godoc
// @Summary Iniciar sesión de sala
// @Tags rooms
// @Produce json
// @Security BearerAuth
// @Param code path string true "Código de sala"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /rooms/{code}/start [patch]
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

// EndSession godoc
// @Summary Finalizar sesión de sala
// @Tags rooms
// @Produce json
// @Security BearerAuth
// @Param code path string true "Código de sala"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /rooms/{code}/end [patch]
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

// GetRoom godoc
// @Summary Obtener sala por código
// @Tags rooms
// @Produce json
// @Security BearerAuth
// @Param code path string true "Código de sala"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /rooms/{code} [get]
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

// GetParticipants godoc
// @Summary Listar participantes de una sala
// @Tags rooms
// @Produce json
// @Security BearerAuth
// @Param code path string true "Código de sala"
// @Success 200 {array} domain.ParticipantWithUser
// @Router /rooms/{code}/participants [get]
func (h *RoomHandler) GetParticipants(w http.ResponseWriter, r *http.Request) {
    code := extractCode(r.URL.Path, "/rooms/", "/participants")
    participants, err := h.uc.GetParticipants(code)
    if err != nil {
        jsonError(w, err.Error(), http.StatusBadRequest)
        return
    }
    if participants == nil {
        jsonResponse(w, http.StatusOK, []interface{}{})
        return
    }
    jsonResponse(w, http.StatusOK, participants)
}

// KickParticipant godoc
// @Summary Host expulsa a un participante
// @Tags rooms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param code path string true "Código de sala"
// @Param body body map[string]int true "user_id del participante a expulsar"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /rooms/{code}/kick [post]
func (h *RoomHandler) KickParticipant(w http.ResponseWriter, r *http.Request) {
    code := extractCode(r.URL.Path, "/rooms/", "/kick")
    claims := getClaims(r)

    var body struct {
        UserID int `json:"user_id"`
    }
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.UserID == 0 {
        jsonError(w, "user_id requerido", http.StatusBadRequest)
        return
    }

    if err := h.uc.KickParticipant(code, claims.UserID, body.UserID); err != nil {
        jsonError(w, err.Error(), http.StatusBadRequest)
        return
    }

    h.hub.Broadcast(code, ws.Message{
        Event:    "participant_kicked",
        RoomCode: code,
        Payload:  map[string]int{"user_id": body.UserID},
    })

    jsonResponse(w, http.StatusOK, map[string]string{"message": "participante expulsado"})
}
// GetOnlineUsers godoc
// @Summary Ver quién está conectado ahora mismo en la sala (vía WS)
// @Tags rooms
// @Produce json
// @Security BearerAuth
// @Param code path string true "Código de sala"
// @Success 200 {array} ws.ClientInfo
// @Router /rooms/{code}/online [get]
func (h *RoomHandler) GetOnlineUsers(hub *ws.Hub) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        code := extractCode(r.URL.Path, "/rooms/", "/online")
        online := hub.GetOnlineUsers(code)
        if online == nil {
            online = []ws.ClientInfo{}
        }
        jsonResponse(w, http.StatusOK, online)
    }
}