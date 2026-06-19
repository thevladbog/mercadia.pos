package app_test

import (
	"context"
	"testing"
	"time"

	"mercadia.dev/pos/services/central-backend/internal/app"
	"mercadia.dev/pos/services/central-backend/internal/domain"
	"mercadia.dev/pos/services/central-backend/internal/infra/memory"
)

func seedCentralAdmin(t *testing.T, store *memory.Store) {
	t.Helper()
	passwordHash, err := app.HashPassword("admin-pass")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user, err := domain.NewCentralUser(domain.CentralUser{
		ID:           "admin-1",
		Email:        "admin@example.com",
		DisplayName:  "Admin",
		PasswordHash: passwordHash,
		Roles:        []domain.CentralRole{domain.CentralRoleAdmin},
		CreatedAt:    time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("new central user: %v", err)
	}
	if err := store.SaveUser(context.Background(), user); err != nil {
		t.Fatalf("save user: %v", err)
	}
}

func TestAuthCreateSessionSuccess(t *testing.T) {
	store := memory.NewStore()
	seedCentralAdmin(t, store)

	auth := app.NewAuthService(store, store)
	result, err := auth.CreateSession(context.Background(), app.CreateSessionCommand{
		Email:    "admin@example.com",
		Password: "admin-pass",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if result.Token == "" || result.UserID != "admin-1" {
		t.Fatalf("session = %+v", result)
	}

	resolved, err := auth.ResolveSession(context.Background(), result.Token)
	if err != nil {
		t.Fatalf("resolve session: %v", err)
	}
	if resolved.UserID != "admin-1" {
		t.Fatalf("resolved = %+v", resolved)
	}
}

func TestAuthCreateSessionInvalidCredentials(t *testing.T) {
	store := memory.NewStore()
	seedCentralAdmin(t, store)

	auth := app.NewAuthService(store, store)
	if _, err := auth.CreateSession(context.Background(), app.CreateSessionCommand{
		Email:    "admin@example.com",
		Password: "wrong",
	}); err == nil {
		t.Fatal("expected invalid credentials")
	}
}

func TestAuthResolveSessionExpired(t *testing.T) {
	store := memory.NewStore()
	seedCentralAdmin(t, store)

	// Save expired session directly.
	now := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)
	session, err := domain.NewCentralSession(domain.CentralUser{
		ID:    "admin-1",
		Email: "admin@example.com",
		Roles: []domain.CentralRole{domain.CentralRoleAdmin},
	}, "expired-token", now, time.Hour)
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	session.ExpiresAt = now.Add(-time.Minute)
	if err := store.SaveSession(context.Background(), session); err != nil {
		t.Fatalf("save session: %v", err)
	}

	expiredAuth := app.NewAuthService(store, store)
	if _, err := expiredAuth.ResolveSession(context.Background(), "expired-token"); err == nil {
		t.Fatal("expected expired session")
	}
}

func TestCentralRBACPermissions(t *testing.T) {
	if err := app.CheckCentralPermission([]domain.CentralRole{domain.CentralRoleViewer}, app.PermissionReportingRead); err != nil {
		t.Fatalf("viewer reporting read: %v", err)
	}
	if err := app.CheckCentralPermission([]domain.CentralRole{domain.CentralRoleViewer}, app.PermissionUsersManage); err == nil {
		t.Fatal("expected permission denied for viewer user manage")
	}
	if err := app.CheckCentralPermission([]domain.CentralRole{domain.CentralRoleAdmin}, app.PermissionUsersManage); err != nil {
		t.Fatalf("admin user manage: %v", err)
	}
}
