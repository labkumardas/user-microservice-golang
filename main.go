package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"user-microservice-golang/config"
	"user-microservice-golang/controller"
	"user-microservice-golang/repository"
	"user-microservice-golang/router"
	"user-microservice-golang/service"
)

func main() {
	// ── Load .env (ignored if missing in production) ─────────────────────────
	_ = godotenv.Load()

	// ── Configuration ────────────────────────────────────────────────────────
	cfg := config.Load()

	// ── Logger ───────────────────────────────────────────────────────────────
	logger, err := config.NewLogger(cfg.AppEnv)
	if err != nil {
		panic(fmt.Sprintf("failed to initialise logger: %v", err))
	}
	defer logger.Sync() //nolint:errcheck

	// ── MongoDB ──────────────────────────────────────────────────────────────
	mongoClient, err := config.NewMongoClient(cfg, logger)
	if err != nil {
		logger.Fatal("failed to connect to MongoDB", zap.Error(err))
	}

	// ── Wire dependencies (manual DI) ────────────────────────────────────────
	userRepo := repository.NewUserRepository(mongoClient, logger)
	userSvc := service.NewUserService(userRepo, cfg, logger)
	userCtrl := controller.NewUserController(userSvc, logger)

	// ── Router ───────────────────────────────────────────────────────────────
	engine := router.Setup(cfg, userCtrl, logger)

	// ── HTTP server ──────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start in background
	go func() {
		logger.Info("user-microservice-golang starting", zap.String("port", cfg.ServerPort), zap.String("env", cfg.AppEnv))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	// ── Graceful shutdown ────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server…")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", zap.Error(err))
	}

	if err := mongoClient.Disconnect(ctx); err != nil {
		logger.Error("MongoDB disconnect error", zap.Error(err))
	}

	logger.Info("server exited cleanly")
}
