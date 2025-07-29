package main

// @title Telephone Book API
// @version 1.0
// @description API для телефонного справочника
// @host localhost:8080
// @BasePath /

import (
	"context"
	stdlog "log"
	"log/slog"
	"net/http"
	"os"
	_ "telephone-book/docs" // swagger docs
	sso "telephone-book/internal/clients/sso/grpc"
	"telephone-book/internal/config"
	"telephone-book/internal/http_server/handlers/auth/check_role"
	"telephone-book/internal/http_server/handlers/auth/login"
	"telephone-book/internal/http_server/handlers/auth/register"
	"telephone-book/internal/http_server/handlers/departments"
	"telephone-book/internal/http_server/handlers/utility/birthday"
	"telephone-book/internal/http_server/handlers/utility/emergency"
	imports "telephone-book/internal/http_server/handlers/utility/import"
	"telephone-book/internal/http_server/handlers/utility/search"
	"telephone-book/internal/http_server/handlers/workers"
	"telephone-book/internal/http_server/middleware"
	"telephone-book/internal/lib/logger/sl"
	"telephone-book/internal/lib/logger/slogpretty"
	"telephone-book/internal/storage/postgresql"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
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

	// Swagger UI
	router.Get("/swagger/*", httpSwagger.WrapHandler)

	ctx := context.Background()

	// SSO
	router.Route("/auth", func(r chi.Router) {
		r.Post("/login", login.New(ctx, ssoClient, log))
		r.Post("/register", register.New(ctx, ssoClient, log))
		r.Get("/check-role", check_role.CheckRole(ctx, log))
	})

	// Срочные службы
	router.Route("/emergency", func(r chi.Router) {
		r.Get("/", emergency.New(ctx, log, storage))

	})

	// Поиск
	router.Route("/search", func(r chi.Router) {
		r.Get("/", search.New(ctx, log, storage))
	})

	//Др сегодня и завтра
	router.Route("/birthday", func(r chi.Router) {
		r.Get("/today", birthday.Today(ctx, log, storage))
		r.Get("/tomorrow", birthday.Tomorrow(ctx, log, storage))
	})

	//Работники
	router.Route("/workers", func(r chi.Router) {
		r.Post("/", workers.Create(ctx, log, storage))
		r.Post("/with-photo", workers.CreateWithPhoto(ctx, log, storage))
		r.Get("/{email}", workers.GetByEmail(ctx, log, storage))
		r.Get("/{email}/photo", workers.GetPhoto(ctx, log, storage))
		r.Post("/{email}/photo", workers.UploadPhoto(ctx, log, storage))
		r.Put("/{email}/photo", workers.UpdatePhoto(ctx, log, storage))
		r.Delete("/{email}/photo", workers.DeletePhoto(ctx, log, storage))
		r.Put("/", workers.Update(ctx, log, storage))
		r.Delete("/", workers.Delete(ctx, log, storage))
		r.Post("/all", workers.GetAll(ctx, log, storage))
		r.Post("/import", imports.New(ctx, log, storage))
	})

	// Отделы
	router.Route("/departments", func(r chi.Router) {
		r.Get("/", departments.GetAll(ctx, log, storage))
		r.Post("/", departments.Create(ctx, log, storage))
		r.Put("/", departments.Update(ctx, log, storage))
		r.Delete("/", departments.Delete(ctx, log, storage))
		r.Get("/{department}", departments.GetSections(ctx, log, storage))
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
