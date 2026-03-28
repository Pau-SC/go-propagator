package handler

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/pausic/go-propagator/internal/config"
)

func newTarget(t *testing.T, bodies *[]string, mu *sync.Mutex) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		*bodies = append(*bodies, string(body))
		mu.Unlock()
	}))
}

func TestServeHTTP(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		targetCount   int
		targetDown    int
		wantStatus    int
		wantForwarded int
	}{
		{"single webhook", `{"event":"test"}`, 1, 0, http.StatusOK, 1},
		{"multiple webhooks", `{"event":"test"}`, 3, 0, http.StatusOK, 3},
		{"one target down", `{"event":"test"}`, 2, 1, http.StatusMultiStatus, 2},
		{"all targets down", `{"event":"test"}`, 0, 3, http.StatusInternalServerError, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mu sync.Mutex
			var received []string
			var urls []string

			for range tt.targetCount {
				srv := newTarget(t, &received, &mu)
				defer srv.Close()
				urls = append(urls, srv.URL)
			}

			for range tt.targetDown {
				urls = append(urls, "http://localhost:1")
			}

			cfg := &config.Config{Webhooks: urls, Concurrent: tt.targetCount + tt.targetDown}
			h := NewWebhookHandler(slog.New(slog.NewTextHandler(io.Discard, nil)), cfg)

			req := httptest.NewRequest("POST", "/webhook", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			got := len(received)
			if got != tt.wantForwarded {
				t.Errorf("forwarded to %d targets, want %d", got, tt.wantForwarded)
			}

			for i, body := range received {
				if body != tt.body {
					t.Errorf("target[%d] got %q, want %q", i, body, tt.body)
				}
			}
		})
	}
}

func TestServeHTTPBadBody(t *testing.T) {
	cfg := &config.Config{Webhooks: []string{}, Concurrent: 1}
	h := NewWebhookHandler(slog.New(slog.NewTextHandler(io.Discard, nil)), cfg)

	req := httptest.NewRequest("POST", "/webhook", &badReader{})
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

type badReader struct{}

func (b *badReader) Read([]byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}
