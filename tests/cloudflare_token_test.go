package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"services-health-check/internal/checkers/cloudflare"
)

func TestCloudflareTokenCheckerSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"result": map[string]any{
				"id":     "abc",
				"status": "active",
			},
		})
	}))
	defer server.Close()

	checker := &cloudflare.TokenChecker{
		NameValue: "cf",
		Token:     "token",
		Timeout:   2 * time.Second,
		BaseURL:   server.URL,
	}

	res, err := checker.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != "OK" {
		t.Fatalf("unexpected status: %s", res.Status)
	}
}

func TestCloudflareTokenCheckerFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"errors":  []map[string]any{{"message": "invalid token"}},
		})
	}))
	defer server.Close()

	checker := &cloudflare.TokenChecker{
		NameValue: "cf",
		Token:     "token",
		Timeout:   2 * time.Second,
		BaseURL:   server.URL,
	}

	res, err := checker.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != "CRIT" {
		t.Fatalf("unexpected status: %s", res.Status)
	}
}
