package dto

type CarFilters struct {
	Brand    string `json:"brand"`
	Model    string `json:"model"`
	MinYear  int    `json:"min_year"`
	MaxYear  int    `json:"max_year"`
	MinPrice int    `json:"min_price"`
	MaxPrice int    `json:"max_price"`

	Sort  string `json:"sort"`
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
}
