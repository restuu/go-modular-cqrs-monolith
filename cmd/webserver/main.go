package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	orderapi "go-modular-cqrs-monolith/modules/order/api"
	productapi "go-modular-cqrs-monolith/modules/product/api"
	"go-modular-cqrs-monolith/platform/config"
	"go-modular-cqrs-monolith/platform/httpx/middleware"
	"go-modular-cqrs-monolith/platform/httpx/router"
	"go-modular-cqrs-monolith/platform/metrics"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config", "error", err)
	}

	ctx := context.Background()

	// --- infrastructure: database ---
	db, err := pgxpool.New(ctx, cfg.Database.DSN)
	if err != nil {
		log.Fatal("failed to open database", "error", err)
	}
	if err := db.Ping(ctx); err != nil {
		log.Fatal("failed to ping database", "error", err)
	}
	logger.Info("database connected")
	defer db.Close()

	sqlDB := stdlib.OpenDBFromPool(db)

	productSvc := productapi.NewService(productapi.ServiceDeps{
		DB: sqlDB,
	})

	orderSvc := orderapi.NewService(orderapi.ServiceDeps{
		DB:             sqlDB,
		ProductService: productSvc,
	})

	modules := []router.RouteRegistrar{
		func(app *fiber.App, publicLimiter fiber.Handler, authLimiter fiber.Handler, authMW fiber.Handler) {
			orderSvc.RegisterRouter(orderapi.RouterDeps{
				App:            app,
				AuthMiddleware: authMW,
				AuthLimiter:    authMW,
				PublicLimiter:  publicLimiter,
			})
		},
	}

	// --- HTTP layer ---
	reg, recorder := metrics.New()
	reg.MustRegister(metrics.NewPgxCollector(db))

	r := router.New(router.Deps{
		Logger:         logger,
		AuthMiddleware: middleware.Auth(cfg.JWT.Secret),
		InternalSecret: cfg.Internal.APISecret,
		Recorder:       recorder,
		MetricsHandler: metrics.Handler(reg),
		HealthProbes: map[string]func(ctx context.Context) error{
			"postgres": db.Ping,
		},
		Modules: modules,
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		addr := fmt.Sprintf(":%s", cfg.App.Port)
		logger.Info("API server starting", "addr", addr, "env", cfg.App.Env)
		if err := r.Listen(addr); err != nil {
			log.Fatal("server error", err)
		}
	}()

	<-quit
	logger.Info("shutting down...")
	if err := r.Shutdown(); err != nil {
		log.Fatal("server shutdown failed", err)
	}
	logger.Info("server stopped cleanly")
}
