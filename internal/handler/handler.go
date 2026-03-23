package handler

import (
	"log/slog"
	"net/http"
)

type WebhookHandler struct {
	logger *slog.Logger
}

func NewWebhookHandler(logger *slog.Logger) *WebhookHandler {
	return &WebhookHandler{logger: logger}
}

func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Routing...")

}
