package repository

import (
	"database/sql"

	"apiGolan/src/domain"
)

// ScoreRepo implementa domain.ScoreRepository usando MySQL
type ScoreRepo struct {
	db *sql.DB
}

func NewScoreRepo(db *sql.DB) domain.ScoreRepository {
	return &ScoreRepo{db: db}
}

// Upsert crea el registro de puntos si no existe, o lo actualiza si ya existe
func (r *ScoreRepo) Upsert(roomID, userID, points int) error {
	query := `
		INSERT INTO scores (room_id, user_id, points)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE points = VALUES(points)
	`
	_, err := r.db.Exec(query, roomID, userID, points)
	return err
}

// AddPoints suma o resta puntos al score actual del participante
func (r *ScoreRepo) AddPoints(roomID, userID, delta int) error {
	query := `
		INSERT INTO scores (room_id, user_id, points)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE points = points + VALUES(points)
	`
	_, err := r.db.Exec(query, roomID, userID, delta)
	return err
}

// GetRanking devuelve los participantes ordenados por puntos de mayor a menor
// Incluye el nombre del usuario haciendo JOIN con la tabla users
func (r *ScoreRepo) GetRanking(roomID int) ([]domain.RankingEntry, error) {
	query := `
		SELECT s.user_id, u.name, s.points
		FROM scores s
		JOIN users u ON u.id = s.user_id
		WHERE s.room_id = ?
		ORDER BY s.points DESC
	`
	rows, err := r.db.Query(query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ranking []domain.RankingEntry
	position := 1
	for rows.Next() {
		var entry domain.RankingEntry
		if err := rows.Scan(&entry.UserID, &entry.UserName, &entry.Points); err != nil {
			return nil, err
		}
		entry.Position = position
		position++
		ranking = append(ranking, entry)
	}
	return ranking, nil
}

// GetByRoomAndUser devuelve el score de un usuario específico en una sala
func (r *ScoreRepo) GetByRoomAndUser(roomID, userID int) (*domain.Score, error) {
	score := &domain.Score{}
	query := `SELECT id, room_id, user_id, points, updated_at FROM scores WHERE room_id = ? AND user_id = ?`
	err := r.db.QueryRow(query, roomID, userID).Scan(
		&score.ID, &score.RoomID, &score.UserID, &score.Points, &score.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return score, err
}
