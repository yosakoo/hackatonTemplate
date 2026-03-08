package ratelimiter

import (
	"context"
	"log/slog"
)

type SlogLogger struct{ L *slog.Logger }

func (s *SlogLogger) Warnw(ctx context.Context, msg string, args ...any) {
	s.L.WarnContext(ctx, msg, args...)
}
