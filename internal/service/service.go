package service

import (
	"context"
	"errors"
	"fmt"

	"order-svc/internal/cache"
	"order-svc/internal/model"
	"order-svc/internal/repo"
)

type Service struct {
	repo  *repo.PG
	cache *cache.LRU[string, *model.Order]
}

func New(r *repo.PG, c *cache.LRU[string, *model.Order]) *Service {
	return &Service{repo: r, cache: c}
}

func (s *Service) Put(ctx context.Context, o *model.Order) error {
	if err := s.repo.UpsertOrder(ctx, o); err != nil {
		return err
	}
	s.cache.Set(o.OrderUID, o) // write-through
	return nil
}

func (s *Service) Get(ctx context.Context, uid string) (*model.Order, error) {
	if o, ok := s.cache.Get(uid); ok {
		return o, nil
	}
	o, err := s.repo.GetOrder(ctx, uid)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("repo: %w", err)
	}
	s.cache.Set(uid, o)
	return o, nil
}

func (s *Service) Warmup(ctx context.Context, n int) error {
	orders, err := s.repo.LoadRecent(ctx, n)
	if err != nil {
		return err
	}
	for _, o := range orders {
		s.cache.Set(o.OrderUID, o)
	}
	return nil
}
