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
// @description API de autenticación, salas, ranking, preguntas y websocket.
// @host quickscoreapi.duckdns.org
// @BasePath /
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	db, err := infradb.Connect()
	if err != nil {
		log.Fatal("No se pudo conectar a la base de datos:", err)
	}
	defer db.Close()
	log.Println("Conexión a MySQL exitosa")

	// Repositorios
	userRepo := repository.NewUserRepo(db)
	roomRepo := repository.NewRoomRepo(db)
	participantRepo := repository.NewParticipantRepo(db)
	scoreRepo := repository.NewScoreRepo(db)
	questionRepo := repository.NewQuestionRepo(db)
	answerRepo := repository.NewAnswerRepo(db)

	// Servicios (core)
	userService := core.NewUserService(userRepo)
	roomService := core.NewRoomService(roomRepo, participantRepo, scoreRepo)
	scoreService := core.NewScoreService(scoreRepo, roomRepo)
	questionService := core.NewQuestionService(questionRepo, answerRepo, scoreRepo, roomRepo)

	// Casos de uso (application)
	authUC := usecase.NewAuthUseCase(userService)
	roomUC := usecase.NewRoomUseCase(roomService)
	scoreUC := usecase.NewScoreUseCase(scoreService)
	questionUC := usecase.NewQuestionUseCase(questionService)

	// WebSocket Hub
	hub := websocket.NewHub()

	// Handlers HTTP
	authHandler := handler.NewAuthHandler(authUC)
	roomHandler := handler.NewRoomHandler(roomUC, hub)
	scoreHandler := handler.NewScoreHandler(scoreUC, hub)
	questionHandler := handler.NewQuestionHandler(questionUC, hub)

	// Router
	mux := router.Setup(authHandler, roomHandler, scoreHandler, questionHandler, hub)
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
