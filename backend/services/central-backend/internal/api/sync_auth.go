package api

import (
	"net/http"

	"mercadia.dev/pos/services/central-backend/internal/app"
)

const syncAPIKeyHeader = "X-Sync-Api-Key"

func RequireSyncAPIKey(syncKeys *app.SyncAPIKeyService, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if syncKeys != nil && syncKeys.Enabled() {
			if err := syncKeys.Validate(r.Header.Get(syncAPIKeyHeader)); err != nil {
				writeAppError(w, err)
				return
			}
		}
		next(w, r)
	}
}

func syncAPIKeyProtectedDescription(description string) string {
	return description + " When `MERCADIA_CENTRAL_BACKEND_SYNC_API_KEY` is set, requires `X-Sync-Api-Key` header."
}
