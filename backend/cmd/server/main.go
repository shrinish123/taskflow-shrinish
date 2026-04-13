package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/shrinishv/taskflow-shrinish/backend/internal/config"
	"github.com/shrinishv/taskflow-shrinish/backend/internal/database"
	"github.com/shrinishv/taskflow-shrinish/backend/internal/handler"
	"github.com/shrinishv/taskflow-shrinish/backend/internal/repository"
	"github.com/shrinishv/taskflow-shrinish/backend/internal/router"
	"github.com/shrinishv/taskflow-shrinish/backend/internal/service"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	slog.Info("running database migrations")
	m, err := migrate.New("file://migrations", cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to create migrator", "error", err)
		os.Exit(1)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	srcErr, dbErr := m.Close()
	if srcErr != nil {
		slog.Error("failed to close migration source", "error", srcErr)
	}
	if dbErr != nil {
		slog.Error("failed to close migration db", "error", dbErr)
	}
	slog.Info("migrations completed successfully")

	ctx := context.Background()
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if os.Getenv("SEED_DB") == "true" {
		if err := database.Seed(ctx, pool); err != nil {
			slog.Error("failed to seed database", "error", err)
			os.Exit(1)
		}
	}

	userRepo := repository.NewUserRepository(pool)
	projectRepo := repository.NewProjectRepository(pool)
	taskRepo := repository.NewTaskRepository(pool)

	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	projectService := service.NewProjectService(projectRepo, taskRepo)
	taskService := service.NewTaskService(taskRepo, projectRepo, userRepo)

	authHandler := handler.NewAuthHandler(authService)
	projectHandler := handler.NewProjectHandler(projectService)
	taskHandler := handler.NewTaskHandler(taskService)

	r := router.New(authService, authHandler, projectHandler, taskHandler, cfg.AllowedOrigins)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("=== TaskFlow API server ready ===")
		slog.Info("API running", "url", "http://localhost:"+cfg.Port)
		slog.Info("Health check", "url", "http://localhost:"+cfg.Port+"/api/health")
		slog.Info("Postman collection", "file", "postman_collection.json")
		slog.Info("Test credentials", "email", "test@example.com", "password", "password123")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	select {
	case sig := <-quit:
		slog.Info("shutting down server", "signal", sig.String())
	case err := <-serverErr:
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped gracefully")
}
