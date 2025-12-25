package model

import "context"

type Task struct {
	ID      int64
	URL     string
	Context context.Context
}
