package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pau-sc/go-propagator/internal/config"
	"github.com/pau-sc/go-propagator/internal/handler"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed loading config", "error", err)
		os.Exit(1)
	}

	wh := handler.NewWebhookHandler(logger)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /info", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("pinged")
	})
	mux.Handle("POST /webhook", wh)

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: mux,
	}

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("server starting")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("failed starting server", "error", err)
			os.Exit(1)
		}
	}()

	sig := <-quit
	logger.Info("shutting down...", "signal", sig)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Timeout)*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("shutdown error", "error", err)
		os.Exit(1)
	}

	logger.Info("stopped!")
}
