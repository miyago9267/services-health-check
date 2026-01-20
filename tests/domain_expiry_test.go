package tests

import (
	"context"
	"testing"

	"services-health-check/internal/checkers/domain"
)

func TestDomainExpiryMissingDomain(t *testing.T) {
	checker := &domain.ExpiryChecker{NameValue: "domain"}
	res, err := checker.Check(context.Background())
	if err == nil {
		t.Fatalf("expected error")
	}
	if res.Status != "UNKNOWN" {
		t.Fatalf("unexpected status: %s", res.Status)
	}
}
