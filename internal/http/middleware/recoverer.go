package middleware

import (
	"log"
	"net/http"

	response "github.com/A1exanderShin/autoglobal/internal/lib"
)

// Recoverer — middleware, которое ловит паники в хендлерах и предотвращает падение всего HTTP-сервера
// Если внутри запроса произошла паника, клиент получит JSON 500, а разработчик увидит лог с request-id
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// defer-функция выполнится в самом конце обработки запроса,
		// даже если внутри случится паника. Это "страховка" сервера.
		defer func() {

			// recover() возвращает данные паники, если она произошла.
			// Если паники нет, err будет nil.
			if err := recover(); err != nil {

				// Достаём request-id из контекста, чтобы связать логи.
				id, _ := r.Context().Value(RequestIDKey).(string)
				if id == "" {
					id = "unknown"
				}

				// Записываем информацию о панике в лог.
				log.Printf("[reqID=%s] PANIC: %v", id, err)

				// Отправляем клиенту единый JSON-ответ с кодом 500.
				// Используем нашу общую обёртку для ошибок.
				response.Error(w, http.StatusInternalServerError, "internal server error")
			}
		}()

		// Если паники нет — запрос просто проходит дальше по цепочке.
		next.ServeHTTP(w, r)
	})
}
