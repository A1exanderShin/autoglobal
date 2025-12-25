package service

import (
	"context"
	"errors"
	"sync"

	"github.com/A1exanderShin/autoglobal/internal/ingestion/model"
)

type Service struct {
	tasks  chan model.Task
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func New(parent context.Context, queueSize int) *Service {
	if parent == nil {
		panic("parent context is nil")
	}

	if queueSize <= 0 {
		panic("queueSize must be greater than zero")
	}

	ctx, cancel := context.WithCancel(parent)

	tasks := make(chan model.Task, queueSize)

	return &Service{
		tasks:  tasks,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (s *Service) Start(workerCount int) {
	if workerCount <= 0 {
		panic("workerCount must be greater than zero")
	}

	for i := 0; i < workerCount; i++ {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			for {
				select {
				case <-s.ctx.Done():
					return

				case task, ok := <-s.tasks:
					if !ok {
						return
					}
				}
			}
		}()
	}
}

func (s *Service) Submit(ctx context.Context, url string) error {
	if s.ctx.Err() != nil {
		return s.ctx.Err()
	}

	task := model.Task{
		URL: url,
	}

	select {
	case <-s.ctx.Done():
		return s.ctx.Err()

	case <-ctx.Done():
		return ctx.Err()

	case s.tasks <- task:
		return nil

	default:
		return errors.New("ingestion queue is full")
	}
}
