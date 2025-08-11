package cache

import (
	"context"
	"sync"

	"L0/internal/model"
)

type Cache struct {
	mu     sync.RWMutex
	orders map[string]model.Order
}

func New() *Cache {
	return &Cache{
		orders: make(map[string]model.Order),
	}
}

func (c *Cache) Add(order model.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.orders[order.OrderUID] = order
}

func (c *Cache) Get(uid string) (model.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	order, exists := c.orders[uid]
	return order, exists
}

func (c *Cache) LoadFromDB(ctx context.Context, db DBProvider) error {
	orders, err := db.GetAllOrders(ctx)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	for _, order := range orders {
		c.orders[order.OrderUID] = order
	}
	return nil
}

type DBProvider interface {
	GetAllOrders(ctx context.Context) ([]model.Order, error)
}
