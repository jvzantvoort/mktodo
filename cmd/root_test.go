package cmd

import (
	"testing"
)

func TestRootCommand(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("rootCmd should not be nil")
	}

	if rootCmd.Use != "mktodo" {
		t.Errorf("expected Use to be 'mktodo', got '%s'", rootCmd.Use)
	}
}

func TestSetVersion(t *testing.T) {
	SetVersion("1.0.0", "abc123", "2026-05-02")

	if version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", version)
	}

	if commit != "abc123" {
		t.Errorf("expected commit 'abc123', got '%s'", commit)
	}

	if date != "2026-05-02" {
		t.Errorf("expected date '2026-05-02', got '%s'", date)
	}
}
