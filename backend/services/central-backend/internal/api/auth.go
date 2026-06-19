package api

import (
	"context"
	"net/http"

	"mercadia.dev/pos/platform/httpapi"
	"mercadia.dev/pos/services/central-backend/internal/app"
)

const sessionTokenHeader = "X-Session-Token"

type sessionContextKey struct{}

func RequireSession(auth *app.AuthService, permission app.CentralPermission, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

func SessionFromContext(ctx context.Context) (app.SessionResult, bool) {
	session, ok := ctx.Value(sessionContextKey{}).(app.SessionResult)
	return session, ok
}

func protectedResponseSpecs(successCode string, successDescription string, successSchema httpapi.Schema) map[string]httpapi.ResponseSpec {
	return map[string]httpapi.ResponseSpec{
		successCode: {Description: successDescription, Schema: successSchema},
		"401":       {Description: "Session is missing or invalid", Schema: httpapi.ProblemSchema()},
		"403":       {Description: "Permission denied", Schema: httpapi.ProblemSchema()},
	}
}

func sessionProtectedDescription(description string) string {
	return description + " Requires `X-Session-Token` header."
}
