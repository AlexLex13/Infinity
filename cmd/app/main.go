package main

import (
	"context"
	"github.com/AlexLex13/Infinity/internal/config"
	"github.com/AlexLex13/Infinity/internal/http-server/handlers/save"
	"github.com/AlexLex13/Infinity/internal/http-server/middleware/logger"
	"github.com/AlexLex13/Infinity/internal/lib/logger/sl"
	"github.com/AlexLex13/Infinity/internal/storage/sqlite"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("initializing server", slog.String("env", cfg.Env))
	log.Debug("logger debug mode enabled")

	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to initialize storage", sl.Err(err))
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID) // Добавляет request_id в каждый запрос, для трейсинга
	router.Use(middleware.Recoverer) // Если где-то внутри сервера (обработчика запроса) произойдет паника, приложение не должно упасть
	router.Use(middleware.URLFormat) // Парсер URLов поступающих запросов
	router.Use(middleware.Logger)
	router.Use(logger.New(log))

	router.Post("/", save.New(log, storage))

	log.Info("Server starting", slog.String("address", cfg.Address))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error("failed to start server")
		}
	}()

	log.Info("server started")

	<-done
	log.Info("stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", sl.Err(err))

		return
	}

	log.Info("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
