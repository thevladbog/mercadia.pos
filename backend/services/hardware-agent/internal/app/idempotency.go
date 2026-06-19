package app

import (
	"context"
	"errors"
	"time"
)

var (
	ErrIdempotencyKeyRequired   = errors.New("idempotency key is required")
	ErrIdempotencyKeyReused     = errors.New("idempotency key reused for different command")
	ErrIdempotencyResultMissing = errors.New("idempotency result missing")
)

type IdempotencyStore interface {
	Find(ctx context.Context, operation string, key string) (IdempotencyRecord, bool, error)
	Save(ctx context.Context, record IdempotencyRecord) error
}

type IdempotencyRecord struct {
	Operation   string
	Key         string
	TargetID    string
	Fingerprint string
	Result      any
	CreatedAt   time.Time
}
