package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v5"
	echomw "github.com/labstack/echo/v5/middleware"
	"github.com/redis/go-redis/v9"

	"hackathonTemplate/internal/config"
	"hackathonTemplate/internal/observability"
	"hackathonTemplate/pkg/apierr"
	database "hackathonTemplate/pkg/datebase"
	"hackathonTemplate/pkg/jwtauth"
	"hackathonTemplate/pkg/middleware"
)

func main() {
	cfg := config.Load()

	logger := newLogger(cfg.Log.Level)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ── Database ──────────────────────────────────────────────────────────────
	dbPoolCfg, err := database.NewPoolConfigFromDatabaseConfig(cfg.Database)
	if err != nil {
		logger.Error("failed to build db pool config", "error", err)
		os.Exit(1)
	}

	pool, err := database.NewPostgreSQL(ctx, dbPoolCfg)
	if err != nil {
		logger.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer database.Close(pool)

	// Auto-migrations
	if err := database.RunMigrations(pool, "migrations"); err != nil {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// ── Redis ─────────────────────────────────────────────────────────────────
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer rdb.Close()

	// ── JWT ───────────────────────────────────────────────────────────────────
	jwt := jwtauth.New(jwtauth.Config{
		Secret:         cfg.JWT.Secret,
		AccessTokenTTL: cfg.JWT.AccessTokenTTL,
	})

	// ── Echo ──────────────────────────────────────────────────────────────────
	e := echo.New()

	// Global middleware
	e.Use(echomw.Recover())
	e.Use(echomw.CORS(cfg.CORS.AllowOrigins...))
	e.Use(middleware.Logging(logger))
	e.Use(middleware.HTTPMetrics())

	e.HTTPErrorHandler = apierr.HTTPErrorHandler(logger)

	// Observability routes (/health, /ready, /metrics)
	obsHandler := observability.NewHandler(&observability.CompositeChecker{
		Checkers: []observability.ReadinessChecker{
			&observability.PoolChecker{Pool: pool},
		},
	})
	observability.RegisterRoutes(e, obsHandler)

	// ── Application routes ────────────────────────────────────────────────────
	// Example of a protected group — remove or replace with your own handlers:
	//
	//   api := e.Group("/api/v1", middleware.JWTAuth(jwt))
	//   api.GET("/me", myHandler.Me)
	//
	_ = jwt // used above; keep the reference until you wire real handlers

	// ── HTTP server ───────────────────────────────────────────────────────────
	sc := echo.StartConfig{
		Address:         cfg.HTTP.Addr,
		HideBanner:      true,
		GracefulTimeout: cfg.HTTP.ShutdownTimeout,
		OnShutdownError: func(err error) {
			logger.Error("graceful shutdown failed", "error", err)
		},
		BeforeServeFunc: func(s *http.Server) error {
			s.ReadTimeout = cfg.HTTP.ReadTimeout
			s.WriteTimeout = cfg.HTTP.WriteTimeout
			return nil
		},
	}

	logger.Info("starting HTTP server", "addr", cfg.HTTP.Addr)

	if err := sc.Start(ctx, e); err != nil {
		logger.Error("server stopped with error", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped")
}

func newLogger(level string) *slog.Logger {
	var l slog.Level
	if err := l.UnmarshalText([]byte(level)); err != nil {
		l = slog.LevelInfo
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: l}))
}
