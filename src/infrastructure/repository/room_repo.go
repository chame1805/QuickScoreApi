package repository

import (
	"database/sql"

	"apiGolan/src/domain"
)

// RoomRepo implementa domain.RoomRepository usando MySQL
type RoomRepo struct {
	db *sql.DB
}

func NewRoomRepo(db *sql.DB) domain.RoomRepository {
	return &RoomRepo{db: db}
}

func (r *RoomRepo) Create(room *domain.Room) error {
	query := `INSERT INTO rooms (code, host_id, status) VALUES (?, ?, ?)`
	result, err := r.db.Exec(query, room.Code, room.HostID, room.Status)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	room.ID = int(id)
	return nil
}

func (r *RoomRepo) FindByCode(code string) (*domain.Room, error) {
	room := &domain.Room{}
	query := `SELECT id, code, host_id, status, created_at FROM rooms WHERE code = ?`
	err := r.db.QueryRow(query, code).Scan(
		&room.ID, &room.Code, &room.HostID, &room.Status, &room.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return room, nil
}

func (r *RoomRepo) UpdateStatus(code string, status domain.RoomStatus) error {
	query := `UPDATE rooms SET status = ? WHERE code = ?`
	_, err := r.db.Exec(query, status, code)
	return err
}

// ParticipantRepo implementa domain.ParticipantRepository usando MySQL
type ParticipantRepo struct {
	db *sql.DB
}

func NewParticipantRepo(db *sql.DB) domain.ParticipantRepository {
	return &ParticipantRepo{db: db}
}

func (r *ParticipantRepo) Add(p *domain.Participant) error {
	query := `INSERT INTO participants (room_id, user_id) VALUES (?, ?)`
	result, err := r.db.Exec(query, p.RoomID, p.UserID)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	p.ID = int(id)
	return nil
}

func (r *ParticipantRepo) ExistsInRoom(roomID, userID int) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM participants WHERE room_id = ? AND user_id = ?`
	err := r.db.QueryRow(query, roomID, userID).Scan(&count)
	return count > 0, err
}

func (r *ParticipantRepo) FindByRoom(roomID int) ([]domain.Participant, error) {
	query := `SELECT id, room_id, user_id, joined_at FROM participants WHERE room_id = ?`
	rows, err := r.db.Query(query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []domain.Participant
	for rows.Next() {
		var p domain.Participant
		if err := rows.Scan(&p.ID, &p.RoomID, &p.UserID, &p.JoinedAt); err != nil {
			return nil, err
		}
		participants = append(participants, p)
	}
	return participants, nil
}
