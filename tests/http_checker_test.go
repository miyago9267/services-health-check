package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httpcheck "services-health-check/internal/checkers/http"
)

func TestHTTPCheckerOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	checker := &httpcheck.Checker{NameValue: "http", URL: server.URL, Timeout: 2 * time.Second}
	res, err := checker.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != "OK" {
		t.Fatalf("unexpected status: %s", res.Status)
	}
}

func TestHTTPCheckerCrit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	checker := &httpcheck.Checker{NameValue: "http", URL: server.URL, Timeout: 2 * time.Second}
	res, err := checker.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != "CRIT" {
		t.Fatalf("unexpected status: %s", res.Status)
	}
}
