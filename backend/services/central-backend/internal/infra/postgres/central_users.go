package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"mercadia.dev/pos/services/central-backend/internal/app"
	"mercadia.dev/pos/services/central-backend/internal/domain"
)

func (s *Store) SaveUser(ctx context.Context, user domain.CentralUser) error {
	roles, err := json.Marshal(user.Roles)
	if err != nil {
		return fmt.Errorf("marshal user roles: %w", err)
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO central_users (id, email, display_name, password_hash, roles, active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			email = EXCLUDED.email,
			display_name = EXCLUDED.display_name,
			password_hash = EXCLUDED.password_hash,
			roles = EXCLUDED.roles,
			active = EXCLUDED.active
	`, user.ID, strings.ToLower(user.Email), user.DisplayName, user.PasswordHash, roles, user.Active, user.CreatedAt)
	if err != nil {
		return fmt.Errorf("save central user: %w", err)
	}
	return nil
}

func (s *Store) FindUser(ctx context.Context, userID string) (domain.CentralUser, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, email, display_name, password_hash, roles, active, created_at
		FROM central_users WHERE id = $1
	`, userID)
	return scanCentralUser(row)
}

func (s *Store) FindUserByEmail(ctx context.Context, email string) (domain.CentralUser, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, email, display_name, password_hash, roles, active, created_at
		FROM central_users WHERE lower(email) = lower($1)
	`, email)
	return scanCentralUser(row)
}

func (s *Store) ListUsers(ctx context.Context) ([]domain.CentralUser, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, email, display_name, password_hash, roles, active, created_at
		FROM central_users ORDER BY email
	`)
	if err != nil {
		return nil, fmt.Errorf("list central users: %w", err)
	}
	defer rows.Close()

	users := make([]domain.CentralUser, 0)
	for rows.Next() {
		user, err := scanCentralUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list central users: %w", err)
	}
	return users, nil
}

func (s *Store) SaveSession(ctx context.Context, session domain.CentralSession) error {
	roles, err := json.Marshal(session.Roles)
	if err != nil {
		return fmt.Errorf("marshal session roles: %w", err)
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO central_sessions (token, user_id, roles, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (token) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			roles = EXCLUDED.roles,
			created_at = EXCLUDED.created_at,
			expires_at = EXCLUDED.expires_at
	`, session.Token, session.UserID, roles, session.CreatedAt, session.ExpiresAt)
	if err != nil {
		return fmt.Errorf("save central session: %w", err)
	}
	return nil
}

func (s *Store) FindSessionByToken(ctx context.Context, token string) (domain.CentralSession, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT token, user_id, roles, created_at, expires_at
		FROM central_sessions WHERE token = $1
	`, token)

	var session domain.CentralSession
	var rolesJSON []byte
	if err := row.Scan(&session.Token, &session.UserID, &rolesJSON, &session.CreatedAt, &session.ExpiresAt); err != nil {
		if err == pgx.ErrNoRows {
			return domain.CentralSession{}, app.ErrSessionNotFound
		}
		return domain.CentralSession{}, fmt.Errorf("find central session: %w", err)
	}
	if err := json.Unmarshal(rolesJSON, &session.Roles); err != nil {
		return domain.CentralSession{}, fmt.Errorf("decode session roles: %w", err)
	}
	return session, nil
}

type centralUserRow interface {
	Scan(dest ...any) error
}

func scanCentralUser(row centralUserRow) (domain.CentralUser, error) {
	var user domain.CentralUser
	var rolesJSON []byte
	if err := row.Scan(&user.ID, &user.Email, &user.DisplayName, &user.PasswordHash, &rolesJSON, &user.Active, &user.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return domain.CentralUser{}, app.ErrCentralUserNotFound
		}
		return domain.CentralUser{}, fmt.Errorf("scan central user: %w", err)
	}
	if err := json.Unmarshal(rolesJSON, &user.Roles); err != nil {
		return domain.CentralUser{}, fmt.Errorf("decode user roles: %w", err)
	}
	return user, nil
}
