package app

import (
	"context"
	"os"
	"strings"
	"time"

	"mercadia.dev/pos/services/central-backend/internal/domain"
)

type SeedCentralAdminConfig struct {
	Email       string
	Password    string
	DisplayName string
	UserID      string
}

func SeedCentralAdminConfigFromEnv() (SeedCentralAdminConfig, bool) {
	email := strings.TrimSpace(os.Getenv("MERCADIA_CENTRAL_BACKEND_SEED_ADMIN_EMAIL"))
	password := os.Getenv("MERCADIA_CENTRAL_BACKEND_SEED_ADMIN_PASSWORD")
	if email == "" || password == "" {
		return SeedCentralAdminConfig{}, false
	}
	displayName := strings.TrimSpace(os.Getenv("MERCADIA_CENTRAL_BACKEND_SEED_ADMIN_DISPLAY_NAME"))
	if displayName == "" {
		displayName = "Central Admin"
	}
	userID := strings.TrimSpace(os.Getenv("MERCADIA_CENTRAL_BACKEND_SEED_ADMIN_USER_ID"))
	if userID == "" {
		userID = "seed-admin"
	}
	return SeedCentralAdminConfig{
		Email:       email,
		Password:    password,
		DisplayName: displayName,
		UserID:      userID,
	}, true
}

func BootstrapSeedCentralAdmin(ctx context.Context, users CentralUserRepository, config SeedCentralAdminConfig) error {
	existing, err := users.ListUsers(ctx)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return nil
	}

	passwordHash, err := HashPassword(config.Password)
	if err != nil {
		return err
	}

	user, err := domain.NewCentralUser(domain.CentralUser{
		ID:           config.UserID,
		Email:        config.Email,
		DisplayName:  config.DisplayName,
		PasswordHash: passwordHash,
		Roles:        []domain.CentralRole{domain.CentralRoleAdmin},
		CreatedAt:    time.Now().UTC(),
	})
	if err != nil {
		return err
	}
	return users.SaveUser(ctx, user)
}
