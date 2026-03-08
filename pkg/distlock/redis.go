package distlock

import (
	"context"
	errors "errors"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisLock - распределённая блокировка на базе Redis SET NX EX
// Гарантирует, что только одна реплика выполняет критическую секцию
type RedisLock struct {
	client *redis.Client
}

// New создаёт RedisLock.
func New(client *redis.Client) *RedisLock {
	return &RedisLock{client: client}
}

// Acquire пытается захватить блокировку с указанным ключом и TTL
// Возвращает true, если блокировка успешно получена
func (l *RedisLock) Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	res, err := l.client.SetArgs(ctx, key, "1", redis.SetArgs{
		Mode: "NX",
		TTL:  ttl,
	}).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return res == "OK", nil
}

// Release освобождает блокировку
func (l *RedisLock) Release(ctx context.Context, key string) {
	l.client.Del(ctx, key)
}

// CronTask описывает периодическую задачу с распределённой блокировкой
type CronTask struct {
	Name     string        // имя задачи (для логов и ключа блокировки)
	Interval time.Duration // интервал между запусками
	LockTTL  time.Duration // TTL блокировки (должен быть < Interval)
	Fn       func(ctx context.Context) error
}

// RunCron запускает периодическую задачу в горутине.
// Перед каждым запуском захватывает распределённую блокировку,
// чтобы при нескольких репликах задача выполнялась только на одной.
func RunCron(ctx context.Context, lock *RedisLock, logger *slog.Logger, task CronTask) {
	go func() {
		ticker := time.NewTicker(task.Interval)
		defer ticker.Stop()

		key := "lock:" + task.Name

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				acquired, err := lock.Acquire(ctx, key, task.LockTTL)
				if err != nil {
					logger.Error("cron lock acquire error", "task", task.Name, "error", err)
					continue
				}

				if !acquired {
					continue
				}

				if err := task.Fn(ctx); err != nil {
					logger.Error("cron task failed", "task", task.Name, "error", err)
				}
			}
		}
	}()
}
