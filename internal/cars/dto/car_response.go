package dto

// Для ответа GET
type CarResponse struct {
	ID    int64  `json:"id"`
	Brand string `json:"brand"`
	Model string `json:"model"`
	Year  int    `json:"year"`
	Price int    `json:"price"`
}
