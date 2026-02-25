package usecase

import (
	"apiGolan/src/core"
	"apiGolan/src/domain"
)

// ScoreUseCase orquesta las operaciones de puntos y ranking
type ScoreUseCase struct {
	scoreService *core.ScoreService
}

func NewScoreUseCase(scoreService *core.ScoreService) *ScoreUseCase {
	return &ScoreUseCase{scoreService: scoreService}
}

// AddPointsInput son los datos que llegan cuando el host da puntos
type AddPointsInput struct {
	RoomCode    string `json:"room_code"`
	TargetID    int    `json:"target_user_id"`
	Delta       int    `json:"delta"` // positivo o negativo
	RequesterID int    `json:"-"`     // se toma del token, no del body
}

func (uc *ScoreUseCase) AddPoints(input AddPointsInput) error {
	return uc.scoreService.AddPoints(input.RoomCode, input.RequesterID, input.TargetID, input.Delta)
}

func (uc *ScoreUseCase) GetRanking(code string) ([]domain.RankingEntry, error) {
	return uc.scoreService.GetRanking(code)
}
