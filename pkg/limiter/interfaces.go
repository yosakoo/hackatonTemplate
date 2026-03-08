package ratelimiter

import (
	"context"
)

type Logger interface {
	Warnw(ctx context.Context, msg string, args ...any)
}

type Client interface {
	IsAllowed(ctx context.Context, key string) bool
}
