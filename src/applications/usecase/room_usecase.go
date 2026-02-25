package usecase

import (
	"apiGolan/src/core"
	"apiGolan/src/domain"
)

// RoomUseCase orquesta las operaciones de salas
type RoomUseCase struct {
	roomService *core.RoomService
}

func NewRoomUseCase(roomService *core.RoomService) *RoomUseCase {
	return &RoomUseCase{roomService: roomService}
}

// CreateRoomOutput es lo que se devuelve al crear una sala
type CreateRoomOutput struct {
	Code   string            `json:"code"`
	Status domain.RoomStatus `json:"status"`
}

func (uc *RoomUseCase) CreateRoom(hostID int) (*CreateRoomOutput, error) {
	room, err := uc.roomService.CreateRoom(hostID)
	if err != nil {
		return nil, err
	}
	return &CreateRoomOutput{Code: room.Code, Status: room.Status}, nil
}

func (uc *RoomUseCase) JoinRoom(code string, userID int) error {
	return uc.roomService.JoinRoom(code, userID)
}

func (uc *RoomUseCase) StartSession(code string, hostID int) error {
	return uc.roomService.StartSession(code, hostID)
}

func (uc *RoomUseCase) EndSession(code string, hostID int) error {
	return uc.roomService.EndSession(code, hostID)
}

func (uc *RoomUseCase) GetRoom(code string) (*domain.Room, error) {
	return uc.roomService.GetRoom(code)
}
