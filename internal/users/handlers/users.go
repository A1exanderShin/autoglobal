package handlers

import "github.com/A1exanderShin/autoglobal/internal/cars/service"

type UsersHandlers struct {
	svc *service.UsersService
}

func NewUsersHandlers(svc *service.Service) *UsersHandlers {
	return &UsersHandlers{svc}
}
