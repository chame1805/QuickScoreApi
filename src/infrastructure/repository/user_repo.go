package repository

import (
"database/sql"

"apiGolan/src/domain"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) domain.UserRepository {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(user *domain.User) error {
	query := `INSERT INTO users (name, email, password, role) VALUES (?, ?, ?, ?)`
	result, err := r.db.Exec(query, user.Name, user.Email, user.Password, user.Role)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = int(id)
	return nil
}

func (r *UserRepo) FindByEmail(email string) (*domain.User, error) {
	user := &domain.User{}
	query := `SELECT id, name, email, password, role, created_at FROM users WHERE email = ?`
	err := r.db.QueryRow(query, email).Scan(
&user.ID, &user.Name, &user.Email, &user.Password, &user.Role, &user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) FindByID(id int) (*domain.User, error) {
	user := &domain.User{}
	query := `SELECT id, name, email, password, role, created_at FROM users WHERE id = ?`
	err := r.db.QueryRow(query, id).Scan(
&user.ID, &user.Name, &user.Email, &user.Password, &user.Role, &user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}
