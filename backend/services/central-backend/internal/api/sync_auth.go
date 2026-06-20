package api

import (
	"context"
	"net/http"

	"mercadia.dev/pos/services/central-backend/internal/app"
)

const syncAPIKeyHeader = "X-Sync-Api-Key" //nolint:gosec // HTTP header name, not a credential

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

func RequireSyncAPIKeyOrSession(
	auth *app.AuthService,
	syncKeys *app.SyncAPIKeyService,
	permission app.CentralPermission,
	next http.HandlerFunc,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if syncKeys != nil && syncKeys.Enabled() {
			if err := syncKeys.Validate(r.Header.Get(syncAPIKeyHeader)); err != nil {
				writeAppError(w, err)
				return
			}
			next(w, r)
			return
		}

		token := r.Header.Get(sessionTokenHeader)
		session, err := auth.ResolveSession(r.Context(), token)
		if err != nil {
			writeAppError(w, err)
			return
		}
		if err := app.CheckCentralPermission(session.Roles, permission); err != nil {
			writeAppError(w, err)
			return
		}
		ctx := context.WithValue(r.Context(), sessionContextKey{}, session)
		next(w, r.WithContext(ctx))
	}
}

// RequireSyncAPIKeyOrSessionAuth allows either a valid sync API key or a central session
// with the given permission. Used for catalog read endpoints consumed by admin-web.
func RequireSyncAPIKeyOrSessionAuth(
	auth *app.AuthService,
	syncKeys *app.SyncAPIKeyService,
	permission app.CentralPermission,
	next http.HandlerFunc,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		providedKey := r.Header.Get(syncAPIKeyHeader)
		if syncKeys != nil && syncKeys.Enabled() && providedKey != "" {
			if err := syncKeys.Validate(providedKey); err != nil {
				writeAppError(w, err)
				return
			}
			next(w, r)
			return
		}

		token := r.Header.Get(sessionTokenHeader)
		session, err := auth.ResolveSession(r.Context(), token)
		if err != nil {
			writeAppError(w, err)
			return
		}
		if err := app.CheckCentralPermission(session.Roles, permission); err != nil {
			writeAppError(w, err)
			return
		}
		ctx := context.WithValue(r.Context(), sessionContextKey{}, session)
		next(w, r.WithContext(ctx))
	}
}

func syncAPIKeyProtectedDescription(description string) string {
	return description + " When `MERCADIA_CENTRAL_BACKEND_SYNC_API_KEY` is set, requires `X-Sync-Api-Key` header."
}

func storeRegistrationProtectedDescription(description string) string {
	return description + " Requires `X-Sync-Api-Key` when configured, otherwise `X-Session-Token` with `users.manage`."
}

func catalogListProtectedDescription(description string) string {
	return description + " Requires `X-Session-Token` with `reporting.read`, or `X-Sync-Api-Key` when configured."
}
