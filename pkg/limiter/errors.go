package ratelimiter

import "errors"

var (
	ErrEmptyConfig = errors.New("empty config")
	ErrEmptyLimit  = errors.New("empty limit")
	ErrEmptyWindow = errors.New("empty window")
)
