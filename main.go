package main

import (
	"log"
	"net/http"
	"os"

	"apiGolan/src/applications/usecase"
	"apiGolan/src/core"
	infradb "apiGolan/src/infrastructure/db"
	"apiGolan/src/infrastructure/http/handler"
	"apiGolan/src/infrastructure/http/middleware"
	"apiGolan/src/infrastructure/http/router"
	"apiGolan/src/infrastructure/repository"
	"apiGolan/src/infrastructure/websocket"
)

// @title QuickScore API
// @version 1.0
// @description API de autenticación, salas, ranking y websocket.
// @host quickscoreapi.duckdns.org:8090
// @BasePath /
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// ── 1. Conectar a MySQL ────────────────────────────────
	db, err := infradb.Connect()
	if err != nil {
		log.Fatal("No se pudo conectar a la base de datos:", err)
	}
	defer db.Close()
	log.Println("Conexión a MySQL exitosa")

	// ── 2. Repositorios (infraestructura) ─────────────────
	userRepo := repository.NewUserRepo(db)
	roomRepo := repository.NewRoomRepo(db)
	participantRepo := repository.NewParticipantRepo(db)
	scoreRepo := repository.NewScoreRepo(db)

	// ── 3. Servicios del core (lógica de negocio) ─────────
	userService := core.NewUserService(userRepo)
	roomService := core.NewRoomService(roomRepo, participantRepo, scoreRepo)
	scoreService := core.NewScoreService(scoreRepo, roomRepo)

	// ── 4. Casos de uso (application layer) ───────────────
	authUC := usecase.NewAuthUseCase(userService)
	roomUC := usecase.NewRoomUseCase(roomService)
	scoreUC := usecase.NewScoreUseCase(scoreService)

	// ── 5. WebSocket Hub ──────────────────────────────────
	hub := websocket.NewHub()

	// ── 6. Handlers HTTP ──────────────────────────────────
	authHandler := handler.NewAuthHandler(authUC)
	roomHandler := handler.NewRoomHandler(roomUC, hub)
	scoreHandler := handler.NewScoreHandler(scoreUC, hub)

	// ── 7. Router ─────────────────────────────────────────
	mux := router.Setup(authHandler, roomHandler, scoreHandler, hub)

	// Envuelve todo el mux en CORS para que funcione desde cualquier frontend
	handlerWithCORS := middleware.CORS(mux)

	port := getEnv("PORT", "8080")
	log.Printf("API corriendo en http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, handlerWithCORS))
}
func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
