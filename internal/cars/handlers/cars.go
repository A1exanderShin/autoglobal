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

func (h *CarHandlers) UpdateCar(w http.ResponseWriter, r *http.Request) {
	// получить id-строку
	idStr := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid car ID")
		return
	}

	// чтение JSON тела запроса
	var req dto.UpdateCarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// сервисный слой
	err = h.svc.UpdateCar(r.Context(), id, req)
	if err != nil {
		if err == service.ErrValidation {
			response.Error(w, http.StatusBadRequest, "validation failed")
			return
		}
		if err == service.ErrNotFound {
			response.Error(w, http.StatusNotFound, "not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.OK(w, "car updated")
}

func (h *CarHandlers) DeleteCar(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid car ID")
		return
	}

	err = h.svc.DeleteCar(r.Context(), id)
	if err != nil {
		if err == service.ErrNotFound {
			response.Error(w, http.StatusNotFound, "not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.OK(w, "deleted")
}

func parseCarFilters(r *http.Request) dto.CarFilters {
	q := r.URL.Query()

	f := dto.CarFilters{}

	// brand
	if v := q.Get("brand"); v != "" {
		f.Brand = v
	}

	if v := q.Get("model"); v != "" {
		f.Model = v
	}

	// minYear
	if v := q.Get("min_year"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.MinYear = n
		}
	}

	// maxYear
	if v := q.Get("max_year"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.MaxYear = n
		}
	}

	// minPrice
	if v := q.Get("min_price"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.MinPrice = n
		}
	}

	// maxPrice
	if v := q.Get("max_price"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.MaxPrice = n
		}
	}

	// sort
	f.Sort = q.Get("sort")

	// page
	if v := q.Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.Page = n
		}
	} else {
		f.Page = 1
	}

	// limit
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.Limit = n
		}
	} else {
		f.Limit = 20
	}

	return f
}

func (h *CarHandlers) ListFiltered(w http.ResponseWriter, r *http.Request) {
	// сбор query-параметров
	q := r.URL.Query()

	brand := q.Get("brand")
	model := q.Get("model")

	minYear, _ := strconv.Atoi(q.Get("min_year"))
	maxYear, _ := strconv.Atoi(q.Get("max_year"))
	minPrice, _ := strconv.Atoi(q.Get("min_price"))
	maxPrice, _ := strconv.Atoi(q.Get("max_price"))

	sort := q.Get("sort")
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	// собираем dto.Carfilters
	filters := dto.CarFilters{
		Brand:    brand,
		Model:    model,
		MinYear:  minYear,
		MaxYear:  maxYear,
		MinPrice: minPrice,
		MaxPrice: maxPrice,
		Sort:     sort,
		Page:     page,
		Limit:    limit,
	}

	list, err := h.svc.ListFiltered(r.Context(), filters)

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.WriteJSON(w, http.StatusOK, list)
}
