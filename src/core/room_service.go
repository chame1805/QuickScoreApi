package core

import (
	"errors"
	"math/rand"
	"strings"
	"time"

	"apiGolan/src/domain"
)

// RoomService contiene la lógica de negocio para salas
type RoomService struct {
	roomRepo        domain.RoomRepository
	participantRepo domain.ParticipantRepository
	scoreRepo       domain.ScoreRepository
}

func NewRoomService(
	roomRepo domain.RoomRepository,
	participantRepo domain.ParticipantRepository,
	scoreRepo domain.ScoreRepository,
) *RoomService {
	return &RoomService{
		roomRepo:        roomRepo,
		participantRepo: participantRepo,
		scoreRepo:       scoreRepo,
	}
}

// CreateRoom genera un código único y crea la sala
func (s *RoomService) CreateRoom(hostID int) (*domain.Room, error) {
	code := generateCode()

	room := &domain.Room{
		Code:   code,
		HostID: hostID,
		Status: domain.RoomStatusWaiting,
	}

	if err := s.roomRepo.Create(room); err != nil {
		return nil, err
	}

	return room, nil
}

func (s *RoomService) JoinRoom(code string, userID int) error {
	room, err := s.roomRepo.FindByCode(code)
	if err != nil || room == nil {
		return errors.New("sala no encontrada")
	}

	if room.Status == domain.RoomStatusFinished {
		return errors.New("la sala ya terminó")
	}

	exists, _ := s.participantRepo.ExistsInRoom(room.ID, userID)
	if exists {
		return errors.New("ya estás en esta sala")
	}

	participant := &domain.Participant{
		RoomID: room.ID,
		UserID: userID,
	}

	if err := s.participantRepo.Add(participant); err != nil {
		return err
	}

	// Inicializar el score en 0 al unirse
	return s.scoreRepo.Upsert(room.ID, userID, 0)
}

// StartSession marca la sala como activa (solo el host puede hacer esto)
func (s *RoomService) StartSession(code string, requesterID int) error {
	return s.changeStatus(code, requesterID, domain.RoomStatusWaiting, domain.RoomStatusActive)
}

// EndSession marca la sala como finalizada
func (s *RoomService) EndSession(code string, requesterID int) error {
	return s.changeStatus(code, requesterID, domain.RoomStatusActive, domain.RoomStatusFinished)
}

func (s *RoomService) changeStatus(code string, requesterID int, from, to domain.RoomStatus) error {
	room, err := s.roomRepo.FindByCode(code)
	if err != nil || room == nil {
		return errors.New("sala no encontrada")
	}

	// Solo el host puede cambiar el estado
	if room.HostID != requesterID {
		return errors.New("solo el host puede realizar esta acción")
	}

	if room.Status != from {
		return errors.New("la sala no está en el estado correcto para esta acción")
	}

	return s.roomRepo.UpdateStatus(code, to)
}

// GetRoom devuelve una sala por código
func (s *RoomService) GetRoom(code string) (*domain.Room, error) {
	room, err := s.roomRepo.FindByCode(code)
	if err != nil || room == nil {
		return nil, errors.New("sala no encontrada")
	}
	return room, nil
}

// generateCode genera un código aleatorio de 6 caracteres tipo ABC123
func generateCode() string {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var sb strings.Builder
	for i := 0; i < 6; i++ {
		sb.WriteByte(chars[rng.Intn(len(chars))])
	}
	return sb.String()
}
