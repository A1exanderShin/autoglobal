package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/A1exanderShin/autoglobal/internal/cars/dto"
	"github.com/A1exanderShin/autoglobal/internal/cars/service"
	response "github.com/A1exanderShin/autoglobal/internal/lib"
	"github.com/go-chi/chi/v5"
)

type CarHandlers struct {
	svc *service.Service
}

func NewCarHandlers(svc *service.Service) *CarHandlers {
	return &CarHandlers{svc: svc}
}

func (h *CarHandlers) CreateCar(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateCarRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	id, err := h.svc.CreateCar(r.Context(), req)
	if err != nil {
		if err == service.ErrValidation {
			response.Error(w, http.StatusBadRequest, "validation failed")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}
	response.WriteJSON(w, http.StatusCreated, map[string]any{"id": id})
	return
}

func (h *CarHandlers) GetCar(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid car ID")
		return
	}

	car, err := h.svc.GetCar(r.Context(), id)

	if err != nil {
		if err == service.ErrNotFound {
			response.Error(w, http.StatusNotFound, "not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}
	response.WriteJSON(w, http.StatusOK, car)
	return
}

func (h *CarHandlers) ListCars(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.ListCars(r.Context())

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}
	response.WriteJSON(w, http.StatusOK, list)
	return
}
