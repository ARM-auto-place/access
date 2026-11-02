package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"

	"github.com/ARM-auto-place/access/internal/config"
	"github.com/ARM-auto-place/access/internal/service"
	"github.com/ARM-auto-place/access/internal/storage/cache"
	"github.com/ARM-auto-place/access/internal/storage/postgres"
	customhandels "github.com/ARM-auto-place/access/internal/transport/custom-handels"
	"github.com/ARM-auto-place/access/internal/transport/jsonRPC/internalapi"
)

func main() {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(".env"); err != nil {
			panic("Error loading .env file: " + err.Error())
		}
	}

	cfg := config.Values()
	logger := cfg.Logger()
	log.Logger = logger

	// postgres
	connectManager, err := postgres.New(logger, cfg.Postgres)
	if err != nil {
		logger.Err(err).Msg("Failed to connect to postgres")
		return
	}
	defer connectManager.Close()
	accessStorage := postgres.NewAccessStorage(logger, connectManager)

	// ristretto
	roleCache, err := cache.NewRoleCache(logger, cfg.Cache)
	if err != nil {
		logger.Err(err).Msg("Failed to create role cache")
		return
	}
	defer roleCache.Close()

	accessService := service.NewAccessService(logger, accessStorage, roleCache, cfg.Cache)

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app := internalapi.New(
		logger,
		internalapi.SetReadBufferSize(cfg.MaxRequestHeaderSize),
		internalapi.MaxBodySize(cfg.MaxRequestBodySize),
		internalapi.ReadTimeout(cfg.ReadTimeout),
		internalapi.WriteTimeout(cfg.WriteTimeout),
		internalapi.MaxBatchSize(100),
		internalapi.MaxBatchWorkers(1),
		internalapi.WithRequestID("X-Request-Id"),
		internalapi.Use(func(ctx *fiber.Ctx) error {
			requestCtx, cancel := context.WithTimeout(shutdownCtx, cfg.WriteTimeout)
			defer cancel()
			ctx.SetUserContext(log.Ctx(ctx.UserContext()).WithContext(requestCtx))
			return ctx.Next()
		}),
		internalapi.Use(cors.New(cors.Config{
			AllowOriginsFunc: isAllowedOrigin(cfg.AllowedOrigins),
			AllowMethods:     "POST,OPTIONS",
			AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Request-Id",
		})),
		internalapi.InternalAPI(internalapi.NewInternalAPI(accessService).WithErrorHandler(customhandels.ErrorHandler)),
	).WithLog().WithMetrics()

	app.ServeHealth(cfg.HealthBind, map[string]string{"status": "ok"})
	app.ServeMetrics(logger, cfg.MetricsPath, cfg.MetricsBind)

	refreshDone := make(chan struct{})
	go func() {
		defer close(refreshDone)
		accessService.RunCacheRefresh(shutdownCtx)
	}()
	defer func() {
		stop()
		<-refreshDone
	}()

	go func() {
		logger.Info().Msgf("HTTP server started address:%s", cfg.ServiceBind)
		if err := app.Fiber().Listen(cfg.ServiceBind); err != nil {
			logger.Err(err).Msg("Error serving HTTP server")
			stop()
		}
	}()

	<-shutdownCtx.Done()
	logger.Info().Msg("Server stopping")

	done := make(chan struct{})
	go func() {
		app.Shutdown()
		<-refreshDone
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(cfg.ShutdownTimeout):
		logger.Warn().Msg("Server shutdown timed out")
	}
	logger.Info().Msg("Server stopped")
}

func isAllowedOrigin(allowedOrigins []string) func(origin string) bool {
	return func(origin string) bool {
		for _, allowed := range allowedOrigins {
			if allowed == origin {
				return true
			}
		}
		return false
	}
}
