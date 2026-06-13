package main

import "testing"

// TestRunUnknownMode is a smoke test: the scaffold compiles and run() rejects an
// unknown mode rather than panicking. Replace/extend as surfaces land.
func TestRunUnknownMode(t *testing.T) {
	if err := run([]string{"-mode", "nope"}); err == nil {
		t.Fatal("expected error for unknown mode, got nil")
	}
}
