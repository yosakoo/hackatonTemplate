package ratelimiter

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type limitsEntry struct {
	Identifier string `json:"identifier"`
	Limit      int    `json:"limit"`
	Window     int    `json:"window"`
}

type rateLimitFile struct {
	LimitsIP   limitsEntry   `json:"limitsIp"`
	LimitsAPI  []limitsEntry `json:"limitsAPI"`
	LimitsUser limitsEntry   `json:"limitsUser"`
}

func LoadIPConfigFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var f rateLimitFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}

	return &Config{
		Limit:        f.LimitsIP.Limit,
		Window:       time.Duration(f.LimitsIP.Window) * time.Second,
		CacheTimeout: time.Second,
	}, nil
}

func LoadAPIConfigsFromFile(path string) (map[string]*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var f rateLimitFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}

	result := make(map[string]*Config, len(f.LimitsAPI))
	for _, e := range f.LimitsAPI {
		result[e.Identifier] = &Config{
			Limit:        e.Limit,
			Window:       time.Duration(e.Window) * time.Second,
			CacheTimeout: time.Second,
		}
	}

	return result, nil
}
