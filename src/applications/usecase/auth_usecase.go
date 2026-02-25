package usecase

import (
"apiGolan/src/core"
"apiGolan/src/domain"
)

type AuthUseCase struct {
	userService *core.UserService
}

func NewAuthUseCase(userService *core.UserService) *AuthUseCase {
	return &AuthUseCase{userService: userService}
}

type RegisterInput struct {
	Name     string      `json:"name"`
	Email    string      `json:"email"`
	Password string      `json:"password"`
	Role     domain.Role `json:"role"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (uc *AuthUseCase) Register(input RegisterInput) (*domain.User, error) {
	if input.Role == "" {
		input.Role = domain.RoleParticipant
	}
	return uc.userService.Register(input.Name, input.Email, input.Password, input.Role)
}

func (uc *AuthUseCase) Login(input LoginInput) (*domain.User, error) {
	return uc.userService.Login(input.Email, input.Password)
}
