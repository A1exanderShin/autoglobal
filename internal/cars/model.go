package cars

type Car struct {
	ID    int64  `db:"id"`
	Brand string `db:"brand"`
	Model string `db:"model"`
	Year  int    `db:"year"`
	Price int    `db:"price"`
	URL   string `db:"url"`
}
