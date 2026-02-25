package router

import (
	"net/http"

	"apiGolan/src/infrastructure/http/handler"
	"apiGolan/src/infrastructure/http/middleware"
	ws "apiGolan/src/infrastructure/websocket"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // permitir cualquier origen (ajustar en producción)
}

func Setup(
	authH *handler.AuthHandler,
	roomH *handler.RoomHandler,
	scoreH *handler.ScoreHandler,
	hub *ws.Hub,
) http.Handler {
	mux := http.NewServeMux()

	// ── Rutas públicas ─────────────────────────────────────
	mux.HandleFunc("POST /auth/register", authH.Register)
	mux.HandleFunc("POST /auth/login", authH.Login)

	// ── Rutas protegidas (requieren JWT) ───────────────────
	auth := middleware.Auth

	// Sala — cualquier usuario autenticado puede ver o unirse
	mux.Handle("GET /rooms/{code}", auth(http.HandlerFunc(roomH.GetRoom)))
	mux.Handle("POST /rooms/{code}/join", auth(http.HandlerFunc(roomH.JoinRoom)))
	mux.Handle("GET /rooms/{code}/ranking", auth(http.HandlerFunc(scoreH.GetRanking)))

	// Sala — solo host
	onlyHost := func(h http.Handler) http.Handler {
		return auth(middleware.OnlyHost(h))
	}
	mux.Handle("POST /rooms", onlyHost(http.HandlerFunc(roomH.CreateRoom)))
	mux.Handle("PATCH /rooms/{code}/start", onlyHost(http.HandlerFunc(roomH.StartSession)))
	mux.Handle("PATCH /rooms/{code}/end", onlyHost(http.HandlerFunc(roomH.EndSession)))
	mux.Handle("POST /rooms/{code}/score", onlyHost(http.HandlerFunc(scoreH.AddPoints)))

	// ── WebSocket ──────────────────────────────────────────
	// ws://host:8080/ws?room=ABC123&token=<jwt>
	mux.HandleFunc("GET /ws", func(w http.ResponseWriter, r *http.Request) {
		roomCode := r.URL.Query().Get("room")
		if roomCode == "" {
			http.Error(w, "parámetro room requerido", http.StatusBadRequest)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		hub.Register(conn, roomCode)
	})

	return mux
}
