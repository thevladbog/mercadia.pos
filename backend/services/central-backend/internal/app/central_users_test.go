package app_test

import (
	"context"
	"testing"

	"mercadia.dev/pos/services/central-backend/internal/app"
	"mercadia.dev/pos/services/central-backend/internal/domain"
	"mercadia.dev/pos/services/central-backend/internal/infra/memory"
)

func TestCentralUsersServiceCRUD(t *testing.T) {
	store := memory.NewStore()
	seedCentralAdmin(t, store)

	auth := app.NewAuthService(store, store)
	usersService := app.NewCentralUsersService(store)

	adminSession, err := auth.CreateSession(context.Background(), app.CreateSessionCommand{
		Email:    "admin@example.com",
		Password: "admin-pass",
	})
	if err != nil {
		t.Fatalf("create admin session: %v", err)
	}

	created, err := usersService.CreateUser(context.Background(), app.CreateCentralUserCommand{
		UserID:      "viewer-1",
		Email:       "viewer@example.com",
		DisplayName: "Viewer",
		Password:    "viewer-pass",
		Roles:       []domain.CentralRole{domain.CentralRoleViewer},
		Session:     adminSession,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if created.Email != "viewer@example.com" || !created.Active {
		t.Fatalf("created = %+v", created)
	}

	listed, err := usersService.ListUsers(context.Background(), adminSession)
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	if len(listed) != 2 {
		t.Fatalf("listed = %+v", listed)
	}

	viewerSession, err := auth.CreateSession(context.Background(), app.CreateSessionCommand{
		Email:    "viewer@example.com",
		Password: "viewer-pass",
	})
	if err != nil {
		t.Fatalf("create viewer session: %v", err)
	}
	if _, err := usersService.CreateUser(context.Background(), app.CreateCentralUserCommand{
		UserID:      "blocked-1",
		Email:       "blocked@example.com",
		Password:    "blocked-pass",
		Roles:       []domain.CentralRole{domain.CentralRoleViewer},
		Session:     viewerSession,
	}); err == nil {
		t.Fatal("expected permission denied for viewer creating users")
	}

	inactive := false
	updated, err := usersService.UpdateUser(context.Background(), app.UpdateCentralUserCommand{
		UserID:  "viewer-1",
		Active:  &inactive,
		Session: adminSession,
	})
	if err != nil {
		t.Fatalf("update user: %v", err)
	}
	if updated.Active {
		t.Fatalf("updated = %+v", updated)
	}
}

func TestBootstrapSeedCentralAdmin(t *testing.T) {
	store := memory.NewStore()
	if err := app.BootstrapSeedCentralAdmin(context.Background(), store, app.SeedCentralAdminConfig{
		Email:       "seed@example.com",
		Password:    "seed-pass",
		DisplayName: "Seed Admin",
		UserID:      "seed-admin",
	}); err != nil {
		t.Fatalf("bootstrap seed: %v", err)
	}

	auth := app.NewAuthService(store, store)
	session, err := auth.CreateSession(context.Background(), app.CreateSessionCommand{
		Email:    "seed@example.com",
		Password: "seed-pass",
	})
	if err != nil {
		t.Fatalf("login seeded admin: %v", err)
	}
	if session.UserID != "seed-admin" {
		t.Fatalf("session = %+v", session)
	}

	if err := app.BootstrapSeedCentralAdmin(context.Background(), store, app.SeedCentralAdminConfig{
		Email:    "other@example.com",
		Password: "other-pass",
		UserID:   "other-admin",
	}); err != nil {
		t.Fatalf("second bootstrap: %v", err)
	}
	users, err := store.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("users = %+v", users)
	}
}
