package app_test

import (
	"errors"
	"testing"

	"mercadia.dev/pos/services/central-backend/internal/app"
)

func TestSyncAPIKeyDisabledWhenUnset(t *testing.T) {
	service := app.NewSyncAPIKeyService("")
	if service.Enabled() {
		t.Fatal("expected sync API key auth to be disabled")
	}
	if err := service.Validate(""); err != nil {
		t.Fatalf("validate with disabled auth: %v", err)
	}
	if err := service.Validate("anything"); err != nil {
		t.Fatalf("validate with disabled auth: %v", err)
	}
}

func TestSyncAPIKeyRequiredWhenEnabled(t *testing.T) {
	service := app.NewSyncAPIKeyService("test-key")
	if err := service.Validate(""); !errors.Is(err, app.ErrSyncAPIKeyRequired) {
		t.Fatalf("expected required error, got %v", err)
	}
}

func TestSyncAPIKeyInvalidWhenWrong(t *testing.T) {
	service := app.NewSyncAPIKeyService("test-key")
	if err := service.Validate("wrong-key"); !errors.Is(err, app.ErrSyncAPIKeyInvalid) {
		t.Fatalf("expected invalid error, got %v", err)
	}
	if err := service.Validate("test-ke"); !errors.Is(err, app.ErrSyncAPIKeyInvalid) {
		t.Fatalf("expected invalid error for length mismatch, got %v", err)
	}
}

func TestSyncAPIKeyValidWhenCorrect(t *testing.T) {
	service := app.NewSyncAPIKeyService("test-key")
	if err := service.Validate("test-key"); err != nil {
		t.Fatalf("validate correct key: %v", err)
	}
}
