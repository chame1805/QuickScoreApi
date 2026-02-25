package core

import (
"errors"

"apiGolan/src/domain"
"golang.org/x/crypto/bcrypt"
)

// UserService contiene la lógica de negocio relacionada con usuarios.
type UserService struct {
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// Register valida los datos, hashea la contraseña y crea el usuario
func (s *UserService) Register(name, email, password string, role domain.Role) (*domain.User, error) {
	if name == "" || email == "" || password == "" {
		return nil, errors.New("nombre, email y contraseña son requeridos")
	}

	existing, _ := s.repo.FindByEmail(email)
	if existing != nil {
		return nil, errors.New("el email ya está registrado")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("error al procesar la contraseña")
	}

	user := &domain.User{
		Name:     name,
		Email:    email,
		Password: string(hashed),
		Role:     role,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login valida credenciales y devuelve el usuario si son correctas
func (s *UserService) Login(email, password string) (*domain.User, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil || user == nil {
		return nil, errors.New("credenciales inválidas")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("credenciales inválidas")
	}

	return user, nil
}
