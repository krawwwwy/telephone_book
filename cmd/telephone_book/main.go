package main

import (
	"context"
	stdlog "log"
	"log/slog"
	"net/http"
	"os"
	sso "telephone-book/internal/clients/sso/grpc"
	"telephone-book/internal/config"
	"telephone-book/internal/http_server/handlers/workers/create"
	"telephone-book/internal/http_server/handlers/workers/delete"
	"telephone-book/internal/http_server/handlers/workers/read"
	"telephone-book/internal/http_server/handlers/workers/update"
	"telephone-book/internal/http_server/middleware"
	"telephone-book/internal/lib/logger/sl"
	"telephone-book/internal/lib/logger/slogpretty"
	"telephone-book/internal/storage/postgresql"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env, cfg.Logger)
	log.With(
		slog.String("env", cfg.Env),
	)
	log.Info("logger initialized successfully", slog.Any("cfg", cfg))

	ssoClient, err := sso.New(
		context.Background(),
		log,
		cfg.Clients.SSO.Address,
		cfg.Clients.SSO.Timeout,
		cfg.Clients.SSO.RetriesCount,
	)
	if err != nil {
		log.With(slog.Any("cfg", cfg.Clients.SSO)).Error("failed to init sso client", sl.Err(err))
	}

	log.Info("sso client initialized successfully", slog.Any("sso", cfg.Clients.SSO))
	_ = ssoClient

	newUser, err := ssoClient.Register(context.Background(), "testing@mail.ru", "test")
	if err != nil {
		log.Error("failed test 1 grpc", sl.Err(err))
	} else {
		log.Debug("successfully test 1 grpc", slog.Int("user_id", int(newUser)), slog.String("method", "Register"))
	}

	storage, err := postgresql.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	router := chi.NewRouter()

	router.Use(chimiddleware.RequestID) // tracing
	router.Use(chimiddleware.Logger)
	router.Use(chimiddleware.Recoverer)
	router.Use(chimiddleware.URLFormat)
	router.Use(middleware.CORS)                           // Добавляем CORS middleware
	router.Use(middleware.AuthMiddleware(ssoClient, log)) // Добавляем Auth middleware

	ctx := context.Background()

	// Create a new worker
	router.Post("/worker", create.New(ctx, log, storage))

	// Get all workers
	router.Get("/workers/all", read.GetAll(ctx, log, storage))

	// Get, update, delete workers by ID
	router.Route("/workers", func(r chi.Router) {
		r.Get("/", read.GetByEmail(ctx, log, storage))
		r.Put("/", update.New(ctx, log, storage))
		r.Delete("/", delete.New(ctx, log, storage))
	})

	log.Info("starting server", slog.String("address", cfg.HTTPServer.Address))

	srv := http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("error starting server", sl.Err(err))
	}

	log.Error("server stopped")

}

func setupLogger(env string, logger string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		if logger == "pretty" {
			log = setupPrettySlog(env)
		} else {
			log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
		}
	case envDev:
		if logger == "pretty" {
			log = setupPrettySlog(env)
		} else {
			log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
		}
	case envProd:
		if logger == "pretty" {
			log = setupPrettySlog(env)
		} else {
			log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
		}
	}

	if log != nil {
		return log
	}

	stdlog.Fatal("can not setup logger")
	return nil
}

func setupPrettySlog(env string) *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}
	switch env {
	case envLocal:
		opts.SlogOpts.Level = slog.LevelDebug
	case envDev:
		opts.SlogOpts.Level = slog.LevelDebug
	case envProd:
		opts.SlogOpts.Level = slog.LevelInfo
	}
	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
