package ratelimiter

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	_base    = 10
	_bitSize = 64
)

// Config определяет параметры rate limiter.
type Config struct {
	Limit        int           // максимальное количество реквестов
	Window       time.Duration // время окна для лимита (например, 1 * time.Second)
	CacheTimeout time.Duration // как часто синхронизироваться с Redis
}

// cacheEntry хранит временные метки запросов и время последней синхронизации
type cacheEntry struct {
	entries  map[int64]struct{} // timestamps in UnixNano
	lastSync time.Time
	mu       sync.RWMutex // добавляем отдельный мьютекс для каждого entry
}

// libClient — реализация Client
type libClient struct {
	logger Logger
	cache  map[string]*cacheEntry
	redis  *redis.Client
	config *Config
	mu     sync.RWMutex
}

// New создаёт новый экземпляр rate limiterа
func New(logger Logger, redisClient *redis.Client, config *Config) (Client, error) {
	if err := validate(config); err != nil {
		return nil, fmt.Errorf("during config validation: %w", err)
	}

	return &libClient{
		logger: logger,
		cache:  make(map[string]*cacheEntry),
		redis:  redisClient,
		config: config,
	}, nil
}

// IsAllowed проверяет, разрешён ли запрос по ключу в текущем окне.
func (c *libClient) IsAllowed(ctx context.Context, key string) bool {
	now := time.Now().UnixNano()
	windowStart := now - int64(c.config.Window)

	c.mu.Lock()

	entry, exists := c.cache[key]
	if !exists {
		entry = &cacheEntry{
			entries:  make(map[int64]struct{}),
			lastSync: time.Now(),
		}
		c.cache[key] = entry
	}
	c.mu.Unlock()

	entry.mu.Lock()
	defer entry.mu.Unlock()

	c.cleanupOldEntries(entry, windowStart)

	if time.Since(entry.lastSync) > c.config.CacheTimeout {
		c.syncFromRedis(ctx, key, entry, windowStart)
	}

	if len(entry.entries) >= c.config.Limit {
		return false
	}

	entry.entries[now] = struct{}{}

	c.sendToRedis(ctx, key, now)

	return true
}

// cleanupOldEntries удаляет из локального кеша записи вне временного окна.
func (c *libClient) cleanupOldEntries(entry *cacheEntry, windowStart int64) {
	for ts := range entry.entries {
		if ts < windowStart {
			delete(entry.entries, ts)
		}
	}
}

// syncFromRedis синхронизирует данные из Redis в локальный кеш
func (c *libClient) syncFromRedis(ctx context.Context, key string, entry *cacheEntry, windowStart int64) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	minStr := strconv.FormatInt(windowStart, _base)

	remoteData, err := c.redis.ZRangeArgs(ctx, redis.ZRangeArgs{
		Key:     key,
		Start:   minStr,
		Stop:    "+inf",
		ByScore: true,
	}).Result()
	if err != nil {
		c.logger.Warnw(ctx, "failed to fetch entries from Redis", "key", key, "error", err)
		return
	}

	for _, s := range remoteData {
		if ts, err := strconv.ParseInt(s, _base, _bitSize); err == nil && ts >= windowStart {
			entry.entries[ts] = struct{}{}
		}
	}

	for ts := range entry.entries {
		if ts < windowStart {
			delete(entry.entries, ts)
		}
	}

	maxStr := strconv.FormatInt(windowStart-1, _base)
	if _, err := c.redis.ZRemRangeByScore(ctx, key, "-inf", maxStr).Result(); err != nil {
		c.logger.Warnw(ctx, "failed to cleanup old entries in Redis", "key", key, "error", err)
	}

	entry.lastSync = time.Now()
}

func (c *libClient) sendToRedis(ctx context.Context, key string, timestamp int64) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	input := redis.Z{
		Score:  float64(timestamp),
		Member: strconv.FormatInt(timestamp, _base),
	}

	if _, err := c.redis.ZAdd(ctx, key, input).Result(); err != nil {
		c.logger.Warnw(ctx, "failed to add entry to Redis", "key", key, "error", err)
	}
}

func validate(config *Config) error {
	if config == nil {
		return ErrEmptyConfig
	}

	if config.Limit <= 0 {
		return ErrEmptyLimit
	}

	if config.Window <= 0 {
		return ErrEmptyWindow
	}

	if config.CacheTimeout <= 0 {
		config.CacheTimeout = time.Second
	}

	return nil
}
