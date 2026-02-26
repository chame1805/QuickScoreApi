package core

import (
	"errors"

	"apiGolan/src/domain"
)

// ScoreService contiene la lógica de negocio para puntos y ranking
type ScoreService struct {
	scoreRepo domain.ScoreRepository
	roomRepo  domain.RoomRepository
}

func NewScoreService(scoreRepo domain.ScoreRepository, roomRepo domain.RoomRepository) *ScoreService {
	return &ScoreService{scoreRepo: scoreRepo, roomRepo: roomRepo}
}

// AddPoints suma o resta puntos a un participante (solo el host puede hacerlo)
// delta puede ser positivo (+10) o negativo (-5)
func (s *ScoreService) AddPoints(code string, requesterID, targetUserID, delta int) error {
	room, err := s.roomRepo.FindByCode(code)
	if err != nil || room == nil {
		return errors.New("sala no encontrada")
	}

	// Solo el host de esa sala puede dar puntos
	if room.HostID != requesterID {
		return errors.New("solo el host puede modificar los puntos")
	}

	// Solo se pueden dar puntos con la sesión activa
	if room.Status != domain.RoomStatusActive {
		return errors.New("la sesión no está activa")
	}

	return s.scoreRepo.AddPoints(room.ID, targetUserID, delta)
}

// GetRanking devuelve el ranking ordenado de una sala
func (s *ScoreService) GetRanking(code string) ([]domain.RankingEntry, error) {
	room, err := s.roomRepo.FindByCode(code)
	if err != nil || room == nil {
		return nil, errors.New("sala no encontrada")
	}

	return s.scoreRepo.GetRanking(room.ID)
}

// ResetUserPoints resetea los puntos de un participante específico (solo host)
func (s *ScoreService) ResetUserPoints(code string, hostID, targetUserID int) error {
	room, err := s.roomRepo.FindByCode(code)
	if err != nil || room == nil {
		return errors.New("sala no encontrada")
	}
	if room.HostID != hostID {
		return errors.New("solo el host puede resetear puntos")
	}
	return s.scoreRepo.ResetPoints(room.ID, targetUserID)
}

// ResetAllPoints resetea los puntos de todos en la sala (solo host)
func (s *ScoreService) ResetAllPoints(code string, hostID int) error {
	room, err := s.roomRepo.FindByCode(code)
	if err != nil || room == nil {
		return errors.New("sala no encontrada")
	}
	if room.HostID != hostID {
		return errors.New("solo el host puede resetear puntos")
	}
	return s.scoreRepo.ResetAllPoints(room.ID)
}
