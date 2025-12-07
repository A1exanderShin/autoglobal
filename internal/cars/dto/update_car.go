package dto

type UpdateCarRequest struct {
	Brand string `json:"brand"`
	Model string `json:"model"`
	Year  int    `json:"year"`
	Price int    `json:"price"`
	URL   string `json:"url"`
}
