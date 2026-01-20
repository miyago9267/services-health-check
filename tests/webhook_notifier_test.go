package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"services-health-check/internal/core/notify"
	"services-health-check/internal/notifiers/webhook"
)

type webhookPayload struct {
	Details string `json:"details"`
}

func TestWebhookPayload(t *testing.T) {
	var got webhookPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := &webhook.Notifier{NameValue: "webhook", URL: server.URL, Timeout: 2 * time.Second}
	event := notify.Event{Service: "svc", Status: "WARN", Summary: "sum", Details: "a; b", OccurredAt: time.Now()}
	if err := n.Send(context.Background(), event); err != nil {
		t.Fatalf("send error: %v", err)
	}
	if !strings.Contains(got.Details, "- a") {
		t.Fatalf("expected list details, got: %q", got.Details)
	}
}
