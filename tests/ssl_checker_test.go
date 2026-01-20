package tests

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"services-health-check/internal/checkers/ssl"
)

func TestSSLCheckerSkipVerify(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	addr := server.Listener.Addr().String()
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("split host: %v", err)
	}

	checker := &ssl.Checker{
		NameValue:  "ssl",
		Address:    addr,
		ServerName: host,
		Timeout:    2 * time.Second,
		SkipVerify: true,
	}

	res, err := checker.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != "OK" {
		t.Fatalf("unexpected status: %s", res.Status)
	}
}

func TestSSLCheckerMissingAddress(t *testing.T) {
	checker := &ssl.Checker{NameValue: "ssl"}
	res, err := checker.Check(context.Background())
	if err == nil {
		t.Fatalf("expected error")
	}
	if res.Status != "UNKNOWN" {
		t.Fatalf("unexpected status: %s", res.Status)
	}
}
