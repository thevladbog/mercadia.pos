package app

import (
	"crypto/subtle"
	"errors"
	"os"
	"strings"
)

var (
	ErrSyncAPIKeyRequired = errors.New("sync api key required")
	ErrSyncAPIKeyInvalid  = errors.New("sync api key invalid")
)

type SyncAPIKeyService struct {
	configuredKey string
}

func NewSyncAPIKeyService(configuredKey string) *SyncAPIKeyService {
	return &SyncAPIKeyService{
		configuredKey: strings.TrimSpace(configuredKey),
	}
}

func NewSyncAPIKeyServiceFromEnv() *SyncAPIKeyService {
	return NewSyncAPIKeyService(os.Getenv("MERCADIA_CENTRAL_BACKEND_SYNC_API_KEY"))
}

func (s *SyncAPIKeyService) Enabled() bool {
	return s.configuredKey != ""
}

func (s *SyncAPIKeyService) Validate(provided string) error {
	if !s.Enabled() {
		return nil
	}
	provided = strings.TrimSpace(provided)
	if provided == "" {
		return ErrSyncAPIKeyRequired
	}
	if len(provided) != len(s.configuredKey) {
		return ErrSyncAPIKeyInvalid
	}
	if subtle.ConstantTimeCompare([]byte(provided), []byte(s.configuredKey)) != 1 {
		return ErrSyncAPIKeyInvalid
	}
	return nil
}
