package repository

type CarRow struct {
	ID     int64
	Brand  string
	Model  string
	Year   int
	Price  int
	URL    string
	UserID *int64
}
