package repository

import (
	"database/sql"

	"apiGolan/src/domain"
)

// QuestionRepo implementa domain.QuestionRepository usando MySQL
type QuestionRepo struct {
	db *sql.DB
}

func NewQuestionRepo(db *sql.DB) domain.QuestionRepository {
	return &QuestionRepo{db: db}
}

func (r *QuestionRepo) Create(q *domain.Question) error {
	query := `INSERT INTO questions (room_id, text, correct_answer, points, status) VALUES (?, ?, ?, ?, ?)`
	result, err := r.db.Exec(query, q.RoomID, q.Text, q.CorrectAnswer, q.Points, q.Status)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	q.ID = int(id)
	return nil
}

func (r *QuestionRepo) FindByID(id int) (*domain.Question, error) {
	q := &domain.Question{}
	query := `SELECT id, room_id, text, correct_answer, points, status, created_at FROM questions WHERE id = ?`
	err := r.db.QueryRow(query, id).Scan(&q.ID, &q.RoomID, &q.Text, &q.CorrectAnswer, &q.Points, &q.Status, &q.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return q, err
}

func (r *QuestionRepo) FindOpenByRoom(roomID int) (*domain.Question, error) {
	q := &domain.Question{}
	query := `SELECT id, room_id, text, correct_answer, points, status, created_at FROM questions WHERE room_id = ? AND status = 'open' LIMIT 1`
	err := r.db.QueryRow(query, roomID).Scan(&q.ID, &q.RoomID, &q.Text, &q.CorrectAnswer, &q.Points, &q.Status, &q.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return q, err
}

func (r *QuestionRepo) CloseQuestion(id int) error {
	_, err := r.db.Exec(`UPDATE questions SET status = 'closed' WHERE id = ?`, id)
	return err
}

func (r *QuestionRepo) FindByRoom(roomID int) ([]domain.Question, error) {
	query := `SELECT id, room_id, text, correct_answer, points, status, created_at FROM questions WHERE room_id = ? ORDER BY created_at DESC`
	rows, err := r.db.Query(query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []domain.Question
	for rows.Next() {
		var q domain.Question
		if err := rows.Scan(&q.ID, &q.RoomID, &q.Text, &q.CorrectAnswer, &q.Points, &q.Status, &q.CreatedAt); err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}
	return questions, nil
}

// AnswerRepo implementa domain.AnswerRepository usando MySQL
type AnswerRepo struct {
	db *sql.DB
}

func NewAnswerRepo(db *sql.DB) domain.AnswerRepository {
	return &AnswerRepo{db: db}
}

func (r *AnswerRepo) Create(a *domain.Answer) error {
	query := `INSERT INTO answers (question_id, user_id, text, is_correct) VALUES (?, ?, ?, ?)`
	result, err := r.db.Exec(query, a.QuestionID, a.UserID, a.Text, a.IsCorrect)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	a.ID = int(id)
	return nil
}

func (r *AnswerRepo) HasAnswered(questionID, userID int) (bool, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM answers WHERE question_id = ? AND user_id = ?`, questionID, userID).Scan(&count)
	return count > 0, err
}

func (r *AnswerRepo) FindByQuestion(questionID int) ([]domain.Answer, error) {
	query := `SELECT id, question_id, user_id, text, is_correct, answered_at FROM answers WHERE question_id = ?`
	rows, err := r.db.Query(query, questionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var answers []domain.Answer
	for rows.Next() {
		var a domain.Answer
		if err := rows.Scan(&a.ID, &a.QuestionID, &a.UserID, &a.Text, &a.IsCorrect, &a.AnsweredAt); err != nil {
			return nil, err
		}
		answers = append(answers, a)
	}
	return answers, nil
}
