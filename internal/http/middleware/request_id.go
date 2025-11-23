package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// ctxKey — типизированный ключ для хранения данных в контексте
// Используем отдельный тип, чтобы избежать конфликтов со сторонними ключами
type ctxKey string

// RequestIDKey — ключ, под которым в контексте будет храниться request-id
const RequestIDKey ctxKey = "requestID"

// RequestID — middleware, которое генерирует уникальный request-id
// для каждого HTTP-запроса и сохраняет его в контекст и заголовки ответа
// Используется для трейсинга, логирования и корреляции запросов
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Генерируем уникальный UUID (пример: "2f736fa0-4ed1-4cc2-aa91-864b94a0a8b2")
		id := uuid.New().String()

		// Создаём новый контекст, куда записываем request-id
		// Теперь любой код "под" этим запросом сможет достать ID
		ctx := context.WithValue(r.Context(), RequestIDKey, id)

		// Добавляем request-id в HTTP-заголовки ответа
		// Клиент сможет получить этот ID для трейсинга
		w.Header().Set("X-Request-ID", id)

		// Обновляем запрос с новым контекстом
		r = r.WithContext(ctx)

		// Передаём выполнение следующему middleware или нужному хендлеру
		next.ServeHTTP(w, r)
	})
}
