package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"services-health-check/internal/core/notify"
	"services-health-check/internal/notifiers/slack"
)

type slackPayload struct {
	Attachments []struct {
		Color  string `json:"color"`
		Blocks []struct {
			Type string `json:"type"`
		} `json:"blocks"`
	} `json:"attachments"`
}

func TestSlackPayload(t *testing.T) {
	var got slackPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := &slack.Notifier{NameValue: "slack", URL: server.URL, Timeout: 2 * time.Second}
	event := notify.Event{Service: "svc", Status: "CRIT", Summary: "sum", Details: "a; b", OccurredAt: time.Now()}
	if err := n.Send(context.Background(), event); err != nil {
		t.Fatalf("send error: %v", err)
	}
	if len(got.Attachments) == 0 {
		t.Fatalf("missing attachments")
	}
	if got.Attachments[0].Color == "" {
		t.Fatalf("missing color")
	}
	if len(got.Attachments[0].Blocks) == 0 {
		t.Fatalf("missing blocks")
	}
}
