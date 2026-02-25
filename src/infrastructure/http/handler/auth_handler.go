package handler

import (
	"encoding/json"
	"net/http"

	"apiGolan/src/applications/usecase"
	jwtutil "apiGolan/src/infrastructure/jwt"
)

type AuthHandler struct {
	uc *usecase.AuthUseCase
}

func NewAuthHandler(uc *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{uc: uc}
}

// POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input usecase.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "cuerpo de la petición inválido", http.StatusBadRequest)
		return
	}

	user, err := h.uc.Register(input)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	jsonResponse(w, http.StatusCreated, map[string]interface{}{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
		"role":  user.Role,
	})
}

// POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input usecase.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "cuerpo de la petición inválido", http.StatusBadRequest)
		return
	}

	user, err := h.uc.Login(input)
	if err != nil {
		jsonError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	token, err := jwtutil.Generate(user.ID, string(user.Role))
	if err != nil {
		jsonError(w, "error al generar token", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

// ── helpers compartidos por todos los handlers ────────────

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, msg string, status int) {
	jsonResponse(w, status, map[string]string{"error": msg})
}
