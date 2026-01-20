package tests

import (
	"testing"

	"services-health-check/internal/notifiers/format"
)

func TestDetailsList(t *testing.T) {
	got := format.DetailsList("a; b; c")
	want := "- a\n- b\n- c"
	if got != want {
		t.Fatalf("unexpected list: got %q want %q", got, want)
	}
}

func TestDetailsListEmpty(t *testing.T) {
	got := format.DetailsList("  ")
	if got != "n/a" {
		t.Fatalf("unexpected empty output: %q", got)
	}
}
