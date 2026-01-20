package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"services-health-check/internal/core/notify"
	"services-health-check/internal/notifiers/discord"
)

type discordPayload struct {
	Embeds []struct {
		Title  string `json:"title"`
		Fields []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"fields"`
	} `json:"embeds"`
}

func TestDiscordPayload(t *testing.T) {
	var got discordPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := &discord.Notifier{NameValue: "discord", URL: server.URL, Timeout: 2 * time.Second}
	event := notify.Event{Service: "svc", Status: "WARN", Summary: "sum", Details: "a; b", OccurredAt: time.Now()}
	if err := n.Send(context.Background(), event); err != nil {
		t.Fatalf("send error: %v", err)
	}
	if len(got.Embeds) == 0 {
		t.Fatalf("missing embeds")
	}
	if got.Embeds[0].Title == "" {
		t.Fatalf("missing title")
	}
	if len(got.Embeds[0].Fields) == 0 {
		t.Fatalf("missing fields")
	}
}
