package response

import (
	"encoding/json"
	"net/http"
)

// SuccessResponse — единый формат успешного JSON-ответа
// Status всегда "ok", Data — полезная нагрузка
type SuccessResponse struct {
	Status string `json:"status"`
	Data   any    `json:"data"`
}

// ErrorResponse — единый формат ошибочного JSON-ответа
// Status всегда "error", Message — текст ошибки
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"error"`
}

// WriteJSON — низкоуровневая функция, которая отправляет JSON-клиенту
// 1) Устанавливает заголовок Content-Type
// 2) Устанавливает HTTP-статус
// 3) Кодирует payload в JSON и пишет в ответ
func WriteJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

// OK — helper-функция для успешного JSON-ответа
// Формирует SuccessResponse и отправляет его с HTTP 200
func OK(w http.ResponseWriter, data interface{}) {
	resp := SuccessResponse{
		Status: "ok",
		Data:   data,
	}

	WriteJSON(w, http.StatusOK, resp)
}

// Error — helper-функция для ошибочного ответа
// Формирует ErrorResponse и отправляет его с переданным HTTP-кодом
func Error(w http.ResponseWriter, statuscode int, msg string) {
	resp := ErrorResponse{
		Status:  "error",
		Message: msg,
	}

	WriteJSON(w, statuscode, resp)
}
