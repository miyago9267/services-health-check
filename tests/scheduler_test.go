package tests

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"services-health-check/internal/app"
)

func TestAppRunWithInterval(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := fmt.Sprintf(`checks:
  - type: http
    name: test-http
    url: %s
    interval: 50ms
channels: []
routes: []
log:
  level: info
  format: text
`, server.URL)

	file, err := os.CreateTemp("", "healthd-*.yaml")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	defer os.Remove(file.Name())
	if _, err := file.WriteString(config); err != nil {
		t.Fatalf("write config: %v", err)
	}
	_ = file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx, file.Name())
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("app run error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("timeout waiting for app run")
	}
}
