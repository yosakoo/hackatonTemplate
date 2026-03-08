package observability

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PoolChecker struct {
	Pool *pgxpool.Pool
}

func (c *PoolChecker) Ready(ctx context.Context) error {
	if c.Pool == nil {
		return nil
	}

	return c.Pool.Ping(ctx)
}

type CompositeChecker struct {
	Checkers []ReadinessChecker
}

func (c *CompositeChecker) Ready(ctx context.Context) error {
	for _, ch := range c.Checkers {
		if err := ch.Ready(ctx); err != nil {
			return err
		}
	}

	return nil
}
