package handler

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"

	"github.com/pau-sc/go-propagator/internal/config"
)

type WebhookHandler struct {
	logger *slog.Logger
	cfg    *config.Config
}

func NewWebhookHandler(logger *slog.Logger, cfg *config.Config) *WebhookHandler {
	return &WebhookHandler{logger: logger, cfg: cfg}
}

func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Routing...")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	lenWebhooks := len(h.cfg.Webhooks)
	sem := make(chan struct{}, h.cfg.Concurrent)
	results := make(chan bool, lenWebhooks)

	for _, url := range h.cfg.Webhooks {
		sem <- struct{}{}
		go func(url string) {
			defer func() { <-sem }()

			res, err := http.Post(url, r.Header.Get("Content-Type"), bytes.NewReader(body))
			if err != nil {
				h.logger.Error("failed to forward", "url", url, "error", err)
				results <- false
				return
			}

			res.Body.Close()
			h.logger.Info("forwarded", "url", url, "status", res.StatusCode)
			results <- res.StatusCode == http.StatusOK
		}(url)
	}

	failedRequests := 0
	for range lenWebhooks {
		if !<-results {
			failedRequests++
		}
	}

	switch {
	case failedRequests == 0:
		w.WriteHeader(http.StatusOK)

	case failedRequests < lenWebhooks:
		w.WriteHeader(http.StatusMultiStatus)

	case failedRequests == lenWebhooks:
		w.WriteHeader(http.StatusInternalServerError)

	}
}
