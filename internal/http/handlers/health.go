package handlers

import "net/http"

// w http.ResponseWriter — объект, через который мы отправляем ответ клиенту
// r *http.Request — запрос, который пришёл от клиента
func Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}
