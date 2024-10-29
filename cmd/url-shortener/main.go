package main

import (
	"fmt"
	"net/http"
	"os"

	"log/slog"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"

	"github.com/Talonmortem/Rest-API-service/internal/config"
	"github.com/Talonmortem/Rest-API-service/internal/http-server/handlers/url/delete"
	"github.com/Talonmortem/Rest-API-service/internal/http-server/handlers/url/redirect"
	"github.com/Talonmortem/Rest-API-service/internal/http-server/handlers/url/save"
	mwLogger "github.com/Talonmortem/Rest-API-service/internal/http-server/middleware/logger"
	"github.com/Talonmortem/Rest-API-service/internal/lib/logger/handlers/slogpretty"
	"github.com/Talonmortem/Rest-API-service/internal/lib/logger/sl"
	"github.com/Talonmortem/Rest-API-service/internal/storage/sqlite"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	//init config: cleanenv

	cfg := config.MustLoad()

	//init logger: slog

	log := setupLogger(cfg.Env)
	log.Info("Starting application", "env", cfg.Env)
	log.Debug("Debug messages enabled")
	log.Error("Error messages enabled")

	//init storage: sqlite
	fmt.Println(cfg.StoragePath)
	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("Failed to initialize storage", sl.Err(err))
		os.Exit(1)
	}
	log.Info("Storage initialized", "path", cfg.StoragePath)

	_ = storage
	//TODO: init router: chi, render

	router := chi.NewRouter()

	//middleware
	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(middleware.Heartbeat("/ping"))

	fmt.Println(cfg.HTTPServer.User)
	fmt.Println(cfg.HTTPServer.Password)
	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))
		r.Delete("/url/{alias}", delete.New(log, storage))
		r.Post("/", save.New(log, storage))

	})

	router.Get("/{alias}", redirect.New(log, storage))

	log.Info("starting server", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}
	//TODO: run server

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Error("server stoped")

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default:
		panic(fmt.Sprintf("unknown env: %s", env))
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)
	return slog.New(handler)

}
