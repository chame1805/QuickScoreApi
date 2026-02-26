package router

import (
	"net/http"

	"apiGolan/src/infrastructure/http/handler"
	"apiGolan/src/infrastructure/http/middleware"
	jwtutil "apiGolan/src/infrastructure/jwt"
	ws "apiGolan/src/infrastructure/websocket"

	_ "apiGolan/docs"
	"github.com/gorilla/websocket"
	httpSwagger "github.com/swaggo/http-swagger"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func Setup(
	authH *handler.AuthHandler,
	roomH *handler.RoomHandler,
	scoreH *handler.ScoreHandler,
	questionH *handler.QuestionHandler,
	hub *ws.Hub,
) http.Handler {
	mux := http.NewServeMux()

	// ── Rutas públicas ─────────────────────────────────────
	mux.HandleFunc("POST /auth/register", authH.Register)
	mux.HandleFunc("POST /auth/login", authH.Login)

	auth := middleware.Auth
	onlyHost := func(h http.Handler) http.Handler {
		return auth(middleware.OnlyHost(h))
	}

	// ── Cualquier usuario autenticado ──────────────────────
	mux.Handle("GET /rooms/{code}", auth(http.HandlerFunc(roomH.GetRoom)))
	mux.Handle("POST /rooms/{code}/join", auth(http.HandlerFunc(roomH.JoinRoom)))
	mux.Handle("GET /rooms/{code}/ranking", auth(http.HandlerFunc(scoreH.GetRanking)))
	mux.Handle("GET /rooms/{code}/participants", auth(http.HandlerFunc(roomH.GetParticipants)))
	mux.Handle("GET /rooms/{code}/online", auth(http.HandlerFunc(roomH.GetOnlineUsers(hub))))
	mux.Handle("GET /rooms/{code}/questions/current", auth(http.HandlerFunc(questionH.GetCurrentQuestion)))
	mux.Handle("POST /rooms/{code}/answer", auth(http.HandlerFunc(questionH.SubmitAnswer)))

	// ── Solo host ──────────────────────────────────────────
	mux.Handle("POST /rooms", onlyHost(http.HandlerFunc(roomH.CreateRoom)))
	mux.Handle("PATCH /rooms/{code}/start", onlyHost(http.HandlerFunc(roomH.StartSession)))
	mux.Handle("PATCH /rooms/{code}/end", onlyHost(http.HandlerFunc(roomH.EndSession)))
	mux.Handle("POST /rooms/{code}/score", onlyHost(http.HandlerFunc(scoreH.AddPoints)))
	mux.Handle("POST /rooms/{code}/score/reset", onlyHost(http.HandlerFunc(scoreH.ResetUserPoints)))
	mux.Handle("POST /rooms/{code}/score/reset-all", onlyHost(http.HandlerFunc(scoreH.ResetAllPoints)))
	mux.Handle("POST /rooms/{code}/kick", onlyHost(http.HandlerFunc(roomH.KickParticipant)))
	mux.Handle("POST /rooms/{code}/questions", onlyHost(http.HandlerFunc(questionH.LaunchQuestion)))
	mux.Handle("PATCH /rooms/{code}/questions/{question_id}/close", onlyHost(http.HandlerFunc(questionH.CloseQuestion)))
	mux.Handle("GET /rooms/{code}/questions/{question_id}/answers", onlyHost(http.HandlerFunc(questionH.GetAnswers)))

	// ── WebSocket ──────────────────────────────────────────
	// ws://host:8080/ws?room=ABC123&token=<jwt>&name=Juan
	//
	// El token es OBLIGATORIO. Se valida antes de hacer el upgrade.
	// Así el Hub sabe quién es cada cliente desde el primer momento.
	mux.HandleFunc("GET /ws", func(w http.ResponseWriter, r *http.Request) {
		roomCode := r.URL.Query().Get("room")
		tokenStr := r.URL.Query().Get("token")
		name := r.URL.Query().Get("name") // nombre display del usuario

		if roomCode == "" || tokenStr == "" {
			http.Error(w, `{"error":"room y token son requeridos"}`, http.StatusBadRequest)
			return
		}

		// Validar JWT ANTES de hacer el upgrade a WebSocket
		claims, err := jwtutil.Validate(tokenStr)
		if err != nil {
			http.Error(w, `{"error":"token inválido o expirado"}`, http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		// Registrar con identidad completa
		info := ws.ClientInfo{
			UserID: claims.UserID,
			Name:   name,
			Role:   claims.Role,
		}
		hub.Register(conn, roomCode, info)
	})

	// ── Swagger ────────────────────────────────────────────
	swaggerHandler := httpSwagger.Handler(httpSwagger.URL("/docs/doc.json"))
	mux.Handle("/docs/", swaggerHandler)
	mux.HandleFunc("GET /docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/index.html", http.StatusFound)
	})

	return mux
}
