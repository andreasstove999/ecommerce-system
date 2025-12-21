package inventory

import (
	"context"
	"errors"
	"sync"
)

// Service orchestrates higher-level reservation workflows on top of the Repository.
// It adds basic idempotency keyed by orderID to avoid double-decrementing stock
// if the same order is processed multiple times.
type Service struct {
	repo Repository

	mu        sync.Mutex
	completed map[string]ReserveResult
}

func NewService(repo Repository) *Service {
	return &Service{
		repo:      repo,
		completed: make(map[string]ReserveResult),
	}
}

func (s *Service) ReserveForOrder(ctx context.Context, orderID string, lines []Line) (ReserveResult, error) {
	if orderID == "" {
		return ReserveResult{}, errors.New("orderID is required")
	}

	s.mu.Lock()
	if res, ok := s.completed[orderID]; ok {
		s.mu.Unlock()
		return res, nil
	}
	s.mu.Unlock()

	res, err := s.repo.Reserve(ctx, orderID, lines)
	if err != nil {
		return ReserveResult{}, err
	}

	s.mu.Lock()
	s.completed[orderID] = res
	s.mu.Unlock()

	return res, nil
}
