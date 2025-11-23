package middleware

import (
	"log"
	"net/http"
	"time"
)

// responseWriter — обёртка над стандартным http.ResponseWriter
// Она нужна, чтобы поймать HTTP-статус ответа, который хендлер вернёт
// Обычный ResponseWriter НЕ даёт получить статус-код
type responseWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader — переопределённый метод
// Каждый раз, когда хендлер вызывает WriteHeader(404)
// мы сохраняем этот код внутрь rw.status
// Так мы сможем залогировать фактический статус
func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code                    // сохраняем статус
	rw.ResponseWriter.WriteHeader(code) // вызываем оригинальный метод
}

// Logger — middleware для логирования всех HTTP-запросов
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Достаём request-id, который создал предыдущий middleware
		// Это нужно, чтобы все логи были связаны между собой
		id, _ := r.Context().Value(RequestIDKey).(string)
		if id == "" {
			id = "unknown"
		}

		// Создаём обёртку над ResponseWriter
		// В отличие от обычного w, эта обёртка умеет запоминать статус ответа
		// status: 200 — значение по умолчанию, если хендлер сам статус не укажет
		wrappedWriter := &responseWriter{
			ResponseWriter: w,
			status:         200,
		}

		// Запоминаем время начала обработки запроса
		// Позже сможем вычислить длительность
		start := time.Now()

		// Передаём управление следующему middleware или хендлеру
		// ВАЖНО: мы передаём wrappedWriter, чтобы перехватить WriteHeader()
		next.ServeHTTP(wrappedWriter, r)

		// Вычисляем время выполнения хендлера
		duration := time.Since(start)

		// Достаём метод, путь и статус-код для логов
		method := r.Method             // GET, POST и т.д.
		path := r.URL.Path             // /cars, /health, /cars/12
		status := wrappedWriter.status // финальный статус ответа

		// Формат логов, который обычно используется в сервисах
		log.Printf("[reqID=%s] %s %s %d %s",
			id, method, path, status, duration)
	})
}
