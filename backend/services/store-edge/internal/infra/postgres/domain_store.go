package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"mercadia.dev/pos/services/store-edge/internal/app"
	"mercadia.dev/pos/services/store-edge/internal/domain"
)

func (s *Store) SeedDemoActors(ctx context.Context) error {
	return s.Run(ctx, func(ctx context.Context) error {
		for _, actor := range demoActors() {
			roles, err := json.Marshal(actor.Roles)
			if err != nil {
				return fmt.Errorf("marshal actor roles: %w", err)
			}
			var credentialPolicy []byte
			if actor.CredentialPolicy != nil {
				credentialPolicy, err = json.Marshal(actor.CredentialPolicy)
				if err != nil {
					return fmt.Errorf("marshal actor credential policy: %w", err)
				}
			}
			bindings := actor.CredentialBindings
			if bindings == nil {
				bindings = []domain.CredentialBinding{}
			}
			credentialBindings, err := json.Marshal(bindings)
			if err != nil {
				return fmt.Errorf("marshal actor credential bindings: %w", err)
			}
			_, err = s.conn(ctx).Exec(ctx, `
			INSERT INTO store_actors (id, pin, roles, credential_policy, credential_bindings)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO UPDATE SET
				pin = EXCLUDED.pin,
				roles = EXCLUDED.roles,
				credential_policy = EXCLUDED.credential_policy,
				credential_bindings = EXCLUDED.credential_bindings
		`, actor.ID, actor.PIN, roles, credentialPolicy, credentialBindings)
			if err != nil {
				return fmt.Errorf("seed actor %s: %w", actor.ID, err)
			}
		}
		storePolicy := domain.CredentialPolicy{
			Required: true,
			AllowedKinds: []domain.CredentialKind{
				domain.CredentialKindIButton,
				domain.CredentialKindMSRCard,
				domain.CredentialKindBarcodeCard,
			},
		}
		policyJSON, err := json.Marshal(storePolicy)
		if err != nil {
			return fmt.Errorf("marshal store credential policy: %w", err)
		}
		if _, err := s.conn(ctx).Exec(ctx, `
		INSERT INTO store_credential_policies (store_id, policy)
		VALUES ($1, $2)
		ON CONFLICT (store_id) DO UPDATE SET policy = EXCLUDED.policy
	`, "store-1", policyJSON); err != nil {
			return fmt.Errorf("seed store credential policy: %w", err)
		}
		return nil
	})
}

func demoActors() []domain.Actor {
	notRequired := domain.CredentialPolicy{Required: false}
	return []domain.Actor{
		{ID: "cashier-1", PIN: "1234", Roles: []domain.Role{domain.RoleCashier}, CredentialPolicy: &notRequired},
		{
			ID:    "senior-1",
			PIN:   "5678",
			Roles: []domain.Role{domain.RoleSeniorCashier},
			CredentialBindings: []domain.CredentialBinding{
				{Kind: domain.CredentialKindIButton, TokenHash: app.HashCredentialToken("demo-ibutton-senior-1"), MaskedToken: "iButton demo ****0001", Active: true},           // #nosec G101 -- deterministic demo binding fixture, not a secret.
				{Kind: domain.CredentialKindMSRCard, TokenHash: app.HashCredentialToken("demo-msr-senior-1"), MaskedToken: "MSR staff demo ****0001", Active: true},             // #nosec G101 -- deterministic demo binding fixture, not a secret.
				{Kind: domain.CredentialKindBarcodeCard, TokenHash: app.HashCredentialToken("demo-barcode-senior-1"), MaskedToken: "Barcode staff demo ****0001", Active: true}, // #nosec G101 -- deterministic demo binding fixture, not a secret.
			},
		},
		{ID: "admin-1", PIN: "9999", Roles: []domain.Role{domain.RoleAdmin}, CredentialPolicy: &notRequired},
	}
}

func (s *Store) FindActor(ctx context.Context, actorID string) (domain.Actor, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, pin, roles, credential_policy, credential_bindings FROM store_actors WHERE id = $1
	`, actorID)

	actor, err := scanActor(row)
	if err != nil {
		return domain.Actor{}, err
	}
	return actor, nil
}

func (s *Store) ListActors(ctx context.Context) ([]domain.Actor, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, pin, roles, credential_policy, credential_bindings FROM store_actors ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("list actors: %w", err)
	}
	defer rows.Close()

	actors := make([]domain.Actor, 0)
	for rows.Next() {
		actor, err := scanActor(rows)
		if err != nil {
			return nil, err
		}
		actors = append(actors, actor)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list actors: %w", err)
	}
	return actors, nil
}

func (s *Store) SaveActor(ctx context.Context, actor domain.Actor) error {
	roles, err := json.Marshal(actor.Roles)
	if err != nil {
		return fmt.Errorf("marshal actor roles: %w", err)
	}
	var credentialPolicy []byte
	if actor.CredentialPolicy != nil {
		credentialPolicy, err = json.Marshal(actor.CredentialPolicy)
		if err != nil {
			return fmt.Errorf("marshal actor credential policy: %w", err)
		}
	}
	bindings := actor.CredentialBindings
	if bindings == nil {
		bindings = []domain.CredentialBinding{}
	}
	credentialBindings, err := json.Marshal(bindings)
	if err != nil {
		return fmt.Errorf("marshal actor credential bindings: %w", err)
	}
	commandTag, err := s.conn(ctx).Exec(ctx, `
		UPDATE store_actors
		SET pin = $2, roles = $3, credential_policy = $4, credential_bindings = $5
		WHERE id = $1
	`, actor.ID, actor.PIN, roles, credentialPolicy, credentialBindings)
	if err != nil {
		return fmt.Errorf("save actor: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return app.ErrActorNotFound
	}
	return nil
}

func (s *Store) UpdateActorCredentialPolicy(ctx context.Context, actorID string, policy *domain.CredentialPolicy) (domain.Actor, error) {
	var actor domain.Actor
	err := s.Run(ctx, func(ctx context.Context) error {
		lockedActor, err := s.findActorForCredentialUpdate(ctx, actorID)
		if err != nil {
			return err
		}
		policyJSON, err := marshalCredentialPolicy(policy)
		if err != nil {
			return err
		}
		commandTag, err := s.conn(ctx).Exec(ctx, `
			UPDATE store_actors SET credential_policy = $2 WHERE id = $1
		`, actorID, policyJSON)
		if err != nil {
			return fmt.Errorf("update actor credential policy: %w", err)
		}
		if commandTag.RowsAffected() == 0 {
			return app.ErrActorNotFound
		}
		lockedActor.CredentialPolicy = cloneCredentialPolicyPointer(policy)
		actor = lockedActor
		return nil
	})
	if err != nil {
		return domain.Actor{}, err
	}
	return actor, nil
}

func (s *Store) UpdateActorCredentialBindings(ctx context.Context, actorID string, update func([]domain.CredentialBinding) ([]domain.CredentialBinding, error)) (domain.Actor, error) {
	var actor domain.Actor
	err := s.Run(ctx, func(ctx context.Context) error {
		lockedActor, err := s.findActorForCredentialUpdate(ctx, actorID)
		if err != nil {
			return err
		}
		bindings := append([]domain.CredentialBinding(nil), lockedActor.CredentialBindings...)
		updatedBindings, err := update(bindings)
		if err != nil {
			return err
		}
		bindingsJSON, err := marshalCredentialBindings(updatedBindings)
		if err != nil {
			return err
		}
		commandTag, err := s.conn(ctx).Exec(ctx, `
			UPDATE store_actors SET credential_bindings = $2 WHERE id = $1
		`, actorID, bindingsJSON)
		if err != nil {
			return fmt.Errorf("update actor credential bindings: %w", err)
		}
		if commandTag.RowsAffected() == 0 {
			return app.ErrActorNotFound
		}
		lockedActor.CredentialBindings = append([]domain.CredentialBinding(nil), updatedBindings...)
		actor = lockedActor
		return nil
	})
	if err != nil {
		return domain.Actor{}, err
	}
	return actor, nil
}

func (s *Store) findActorForCredentialUpdate(ctx context.Context, actorID string) (domain.Actor, error) {
	row := s.conn(ctx).QueryRow(ctx, `
		SELECT id, pin, roles, credential_policy, credential_bindings
		FROM store_actors
		WHERE id = $1
		FOR UPDATE
	`, actorID)
	return scanActor(row)
}

func marshalCredentialPolicy(policy *domain.CredentialPolicy) ([]byte, error) {
	if policy == nil {
		return nil, nil
	}
	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return nil, fmt.Errorf("marshal actor credential policy: %w", err)
	}
	return policyJSON, nil
}

func marshalCredentialBindings(bindings []domain.CredentialBinding) ([]byte, error) {
	if bindings == nil {
		bindings = []domain.CredentialBinding{}
	}
	bindingsJSON, err := json.Marshal(bindings)
	if err != nil {
		return nil, fmt.Errorf("marshal actor credential bindings: %w", err)
	}
	return bindingsJSON, nil
}

func cloneCredentialPolicyPointer(policy *domain.CredentialPolicy) *domain.CredentialPolicy {
	if policy == nil {
		return nil
	}
	cloned := *policy
	cloned.AllowedKinds = append([]domain.CredentialKind(nil), policy.AllowedKinds...)
	return &cloned
}

func scanActor(row pgx.Row) (domain.Actor, error) {
	var actor domain.Actor
	var rolesJSON []byte
	var credentialPolicyJSON []byte
	var credentialBindingsJSON []byte
	if err := row.Scan(&actor.ID, &actor.PIN, &rolesJSON, &credentialPolicyJSON, &credentialBindingsJSON); err != nil {
		if err == pgx.ErrNoRows {
			return domain.Actor{}, app.ErrActorNotFound
		}
		return domain.Actor{}, fmt.Errorf("scan actor: %w", err)
	}
	if err := json.Unmarshal(rolesJSON, &actor.Roles); err != nil {
		return domain.Actor{}, fmt.Errorf("decode actor roles: %w", err)
	}
	if len(credentialPolicyJSON) > 0 && string(credentialPolicyJSON) != "null" {
		var policy domain.CredentialPolicy
		if err := json.Unmarshal(credentialPolicyJSON, &policy); err != nil {
			return domain.Actor{}, fmt.Errorf("decode actor credential policy: %w", err)
		}
		actor.CredentialPolicy = &policy
	}
	if len(credentialBindingsJSON) > 0 {
		if err := json.Unmarshal(credentialBindingsJSON, &actor.CredentialBindings); err != nil {
			return domain.Actor{}, fmt.Errorf("decode actor credential bindings: %w", err)
		}
	}
	return actor, nil
}

func (s *Store) FindStoreCredentialPolicy(ctx context.Context, storeID string) (domain.CredentialPolicy, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT policy FROM store_credential_policies WHERE store_id = $1
	`, storeID)

	var policyJSON []byte
	if err := row.Scan(&policyJSON); err != nil {
		if err == pgx.ErrNoRows {
			return domain.CredentialPolicy{}, nil
		}
		return domain.CredentialPolicy{}, fmt.Errorf("find store credential policy: %w", err)
	}
	var policy domain.CredentialPolicy
	if err := json.Unmarshal(policyJSON, &policy); err != nil {
		return domain.CredentialPolicy{}, fmt.Errorf("decode store credential policy: %w", err)
	}
	return policy, nil
}

func (s *Store) SaveStoreCredentialPolicy(ctx context.Context, storeID string, policy domain.CredentialPolicy) error {
	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("marshal store credential policy: %w", err)
	}
	_, err = s.conn(ctx).Exec(ctx, `
		INSERT INTO store_credential_policies (store_id, policy)
		VALUES ($1, $2)
		ON CONFLICT (store_id) DO UPDATE SET policy = EXCLUDED.policy
	`, storeID, policyJSON)
	if err != nil {
		return fmt.Errorf("save store credential policy: %w", err)
	}
	return nil
}

func (s *Store) SaveSession(ctx context.Context, session domain.Session) error {
	roles, err := json.Marshal(session.Roles)
	if err != nil {
		return fmt.Errorf("marshal session roles: %w", err)
	}
	credentialFactor, err := json.Marshal(session.CredentialFactor)
	if err != nil {
		return fmt.Errorf("marshal session credential factor: %w", err)
	}
	_, err = s.conn(ctx).Exec(ctx, `
		INSERT INTO sessions (token, actor_id, roles, credential_factor, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (token) DO UPDATE SET
			actor_id = EXCLUDED.actor_id,
			roles = EXCLUDED.roles,
			credential_factor = EXCLUDED.credential_factor,
			created_at = EXCLUDED.created_at,
			expires_at = EXCLUDED.expires_at
	`, session.Token, session.ActorID, roles, credentialFactor, session.CreatedAt, session.ExpiresAt)
	if err != nil {
		return fmt.Errorf("save session: %w", err)
	}
	return nil
}

func (s *Store) FindSessionByToken(ctx context.Context, token string) (domain.Session, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT token, actor_id, roles, credential_factor, created_at, expires_at
		FROM sessions WHERE token = $1
	`, token)

	var session domain.Session
	var rolesJSON []byte
	var credentialFactorJSON []byte
	if err := row.Scan(&session.Token, &session.ActorID, &rolesJSON, &credentialFactorJSON, &session.CreatedAt, &session.ExpiresAt); err != nil {
		if err == pgx.ErrNoRows {
			return domain.Session{}, app.ErrSessionNotFound
		}
		return domain.Session{}, fmt.Errorf("find session: %w", err)
	}
	if err := json.Unmarshal(rolesJSON, &session.Roles); err != nil {
		return domain.Session{}, fmt.Errorf("decode session roles: %w", err)
	}
	if len(credentialFactorJSON) > 0 && string(credentialFactorJSON) != "null" && string(credentialFactorJSON) != "{}" {
		var factor domain.SessionCredentialFactor
		if err := json.Unmarshal(credentialFactorJSON, &factor); err != nil {
			return domain.Session{}, fmt.Errorf("decode session credential factor: %w", err)
		}
		if factor.Kind != "" {
			session.CredentialFactor = &factor
		}
	}
	return session, nil
}

func (s *Store) FindStoreAuthSettings(ctx context.Context, storeID string) (domain.StoreAuthSettings, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT store_id, failed_attempt_limit, lockout_duration_seconds, pos_auto_lock_seconds,
			updated_by_id, updated_at
		FROM store_auth_settings WHERE store_id = $1
	`, storeID)
	var settings domain.StoreAuthSettings
	if err := row.Scan(
		&settings.StoreID,
		&settings.FailedAttemptLimit,
		&settings.LockoutDurationSeconds,
		&settings.POSAutoLockSeconds,
		&settings.UpdatedByID,
		&settings.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return domain.DefaultStoreAuthSettings(storeID), nil
		}
		return domain.StoreAuthSettings{}, fmt.Errorf("find store auth settings: %w", err)
	}
	return settings, nil
}

func (s *Store) SaveStoreAuthSettings(ctx context.Context, settings domain.StoreAuthSettings) error {
	_, err := s.conn(ctx).Exec(ctx, `
		INSERT INTO store_auth_settings (
			store_id, failed_attempt_limit, lockout_duration_seconds, pos_auto_lock_seconds,
			updated_by_id, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (store_id) DO UPDATE SET
			failed_attempt_limit = EXCLUDED.failed_attempt_limit,
			lockout_duration_seconds = EXCLUDED.lockout_duration_seconds,
			pos_auto_lock_seconds = EXCLUDED.pos_auto_lock_seconds,
			updated_by_id = EXCLUDED.updated_by_id,
			updated_at = EXCLUDED.updated_at
	`, settings.StoreID, settings.FailedAttemptLimit, settings.LockoutDurationSeconds,
		settings.POSAutoLockSeconds, settings.UpdatedByID, settings.UpdatedAt)
	if err != nil {
		return fmt.Errorf("save store auth settings: %w", err)
	}
	return nil
}

func (s *Store) SaveAuthAttempt(ctx context.Context, attempt domain.AuthAttempt) error {
	_, err := s.conn(ctx).Exec(ctx, `
		INSERT INTO auth_attempts (
			id, store_id, actor_id, terminal_id, credential_kind, credential_fingerprint,
			successful, failure_reason, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, attempt.ID, attempt.StoreID, attempt.ActorID, attempt.TerminalID, attempt.CredentialKind,
		attempt.CredentialFingerprint, attempt.Successful, attempt.FailureReason, attempt.CreatedAt)
	if err != nil {
		return fmt.Errorf("save auth attempt: %w", err)
	}
	return nil
}

func (s *Store) ListAuthAttempts(ctx context.Context, filter app.AuthAttemptFilter, params app.PageParams) (app.PageResult[domain.AuthAttempt], error) {
	conditions := []string{"store_id = $1"}
	args := []any{filter.StoreID}
	if filter.ActorID != "" {
		args = append(args, filter.ActorID)
		conditions = append(conditions, fmt.Sprintf("actor_id = $%d", len(args)))
	}
	if filter.TerminalID != "" {
		args = append(args, filter.TerminalID)
		conditions = append(conditions, fmt.Sprintf("terminal_id = $%d", len(args)))
	}
	if filter.Successful != nil {
		args = append(args, *filter.Successful)
		conditions = append(conditions, fmt.Sprintf("successful = $%d", len(args)))
	}
	if !filter.Since.IsZero() {
		args = append(args, filter.Since)
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", len(args)))
	}
	if !filter.Until.IsZero() {
		args = append(args, filter.Until)
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", len(args)))
	}

	where := strings.Join(conditions, " AND ")
	var totalCount int
	if err := s.pool.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*)
		FROM auth_attempts
		WHERE %s
	`, where), args...).Scan(&totalCount); err != nil {
		return app.PageResult[domain.AuthAttempt]{}, fmt.Errorf("count auth attempts: %w", err)
	}

	queryArgs := append([]any(nil), args...)
	queryArgs = append(queryArgs, params.Limit, params.Offset)
	limitParam := len(queryArgs) - 1
	offsetParam := len(queryArgs)
	rows, err := s.pool.Query(ctx, fmt.Sprintf(`
		SELECT id, store_id, actor_id, terminal_id, credential_kind, credential_fingerprint,
			successful, failure_reason, created_at
		FROM auth_attempts
		WHERE %s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d OFFSET $%d
	`, where, limitParam, offsetParam), queryArgs...)
	if err != nil {
		return app.PageResult[domain.AuthAttempt]{}, fmt.Errorf("list auth attempts: %w", err)
	}
	defer rows.Close()

	attempts := []domain.AuthAttempt{}
	for rows.Next() {
		var attempt domain.AuthAttempt
		var kind string
		if err := rows.Scan(
			&attempt.ID,
			&attempt.StoreID,
			&attempt.ActorID,
			&attempt.TerminalID,
			&kind,
			&attempt.CredentialFingerprint,
			&attempt.Successful,
			&attempt.FailureReason,
			&attempt.CreatedAt,
		); err != nil {
			return app.PageResult[domain.AuthAttempt]{}, fmt.Errorf("scan auth attempt: %w", err)
		}
		attempt.CredentialKind = domain.CredentialKind(kind)
		attempts = append(attempts, attempt)
	}
	if err := rows.Err(); err != nil {
		return app.PageResult[domain.AuthAttempt]{}, fmt.Errorf("list auth attempts rows: %w", err)
	}
	return app.PageResult[domain.AuthAttempt]{Items: attempts, TotalCount: totalCount}, nil
}

func (s *Store) CountFailedAuthAttemptsSinceLastSuccess(ctx context.Context, storeID string, actorID string, since time.Time) (int, error) {
	threshold := since
	foundSuccess := false
	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(MAX(created_at), $3), COUNT(*) > 0 FROM auth_attempts
		WHERE store_id = $1 AND actor_id = $2 AND successful = TRUE AND created_at >= $3
	`, storeID, actorID, since).Scan(&threshold, &foundSuccess); err != nil {
		return 0, fmt.Errorf("find last successful auth attempt: %w", err)
	}
	comparison := ">="
	if foundSuccess {
		comparison = ">"
	}
	var count int
	query := fmt.Sprintf(`
		SELECT COUNT(*) FROM auth_attempts
		WHERE store_id = $1 AND actor_id = $2 AND successful = FALSE
			AND failure_reason <> 'locked' AND created_at %s $3
	`, comparison)
	if err := s.pool.QueryRow(ctx, query, storeID, actorID, threshold).Scan(&count); err != nil {
		return 0, fmt.Errorf("count failed auth attempts: %w", err)
	}
	return count, nil
}

func (s *Store) SaveReturn(ctx context.Context, ret domain.Return) error {
	lines, err := json.Marshal(ret.Lines)
	if err != nil {
		return fmt.Errorf("marshal return lines: %w", err)
	}
	_, err = s.conn(ctx).Exec(ctx, `
		INSERT INTO returns (
			id, store_id, receipt_id, kind, lines, reason, actor_id, approved_by_id,
			total_minor, status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			store_id = EXCLUDED.store_id,
			receipt_id = EXCLUDED.receipt_id,
			kind = EXCLUDED.kind,
			lines = EXCLUDED.lines,
			reason = EXCLUDED.reason,
			actor_id = EXCLUDED.actor_id,
			approved_by_id = EXCLUDED.approved_by_id,
			total_minor = EXCLUDED.total_minor,
			status = EXCLUDED.status,
			created_at = EXCLUDED.created_at
	`, ret.ID, ret.StoreID, ret.ReceiptID, ret.Kind, lines, ret.Reason, ret.ActorID,
		ret.ApprovedByID, ret.TotalMinor, ret.Status, ret.CreatedAt)
	if err != nil {
		return fmt.Errorf("save return: %w", err)
	}
	return nil
}

func (s *Store) FindReturn(ctx context.Context, returnID string) (domain.Return, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, store_id, receipt_id, kind, lines, reason, actor_id, approved_by_id,
			total_minor, status, created_at
		FROM returns WHERE id = $1
	`, returnID)
	return scanReturn(row)
}

func (s *Store) ListReturnsByReceipt(ctx context.Context, receiptID string) ([]domain.Return, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, store_id, receipt_id, kind, lines, reason, actor_id, approved_by_id,
			total_minor, status, created_at
		FROM returns WHERE receipt_id = $1
		ORDER BY created_at
	`, receiptID)
	if err != nil {
		return nil, fmt.Errorf("list returns by receipt: %w", err)
	}
	defer rows.Close()

	result := make([]domain.Return, 0)
	for rows.Next() {
		ret, err := scanReturn(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, ret)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list returns by receipt: %w", err)
	}
	return result, nil
}

func (s *Store) ListReturnsByStore(ctx context.Context, storeID string) ([]domain.Return, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, store_id, receipt_id, kind, lines, reason, actor_id, approved_by_id,
			total_minor, status, created_at
		FROM returns WHERE store_id = $1
		ORDER BY created_at DESC, id DESC
	`, storeID)
	if err != nil {
		return nil, fmt.Errorf("list returns by store: %w", err)
	}
	defer rows.Close()

	result := make([]domain.Return, 0)
	for rows.Next() {
		ret, err := scanReturn(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, ret)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list returns by store: %w", err)
	}
	return result, nil
}

func (s *Store) SaveOperationJournalEntry(ctx context.Context, entry domain.OperationJournalEntry) error {
	_, err := s.conn(ctx).Exec(ctx, `
		INSERT INTO operation_journal_entries (
			id, store_id, operation_type, actor_id, reference_id, summary, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			store_id = EXCLUDED.store_id,
			operation_type = EXCLUDED.operation_type,
			actor_id = EXCLUDED.actor_id,
			reference_id = EXCLUDED.reference_id,
			summary = EXCLUDED.summary,
			created_at = EXCLUDED.created_at
	`, entry.ID, entry.StoreID, entry.OperationType, entry.ActorID, entry.ReferenceID, entry.Summary, entry.CreatedAt)
	if err != nil {
		return fmt.Errorf("save operation journal entry: %w", err)
	}
	return nil
}

func (s *Store) ListOperationJournalEntries(ctx context.Context, storeID string, params app.PageParams) (app.PageResult[domain.OperationJournalEntry], error) {
	var total int
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM operation_journal_entries WHERE store_id = $1
	`, storeID).Scan(&total); err != nil {
		return app.PageResult[domain.OperationJournalEntry]{}, fmt.Errorf("count operation journal entries: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, store_id, operation_type, actor_id, reference_id, summary, created_at
		FROM operation_journal_entries
		WHERE store_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, storeID, params.Limit, params.Offset)
	if err != nil {
		return app.PageResult[domain.OperationJournalEntry]{}, fmt.Errorf("list operation journal entries: %w", err)
	}
	defer rows.Close()

	entries := []domain.OperationJournalEntry{}
	for rows.Next() {
		entry, err := scanOperationJournalEntry(rows)
		if err != nil {
			return app.PageResult[domain.OperationJournalEntry]{}, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return app.PageResult[domain.OperationJournalEntry]{}, fmt.Errorf("list operation journal entries: %w", err)
	}
	return app.PageResult[domain.OperationJournalEntry]{Items: entries, TotalCount: total}, nil
}

func scanReturn(row rowScanner) (domain.Return, error) {
	var ret domain.Return
	var linesJSON []byte
	if err := row.Scan(
		&ret.ID, &ret.StoreID, &ret.ReceiptID, &ret.Kind, &linesJSON, &ret.Reason,
		&ret.ActorID, &ret.ApprovedByID, &ret.TotalMinor, &ret.Status, &ret.CreatedAt,
	); err != nil {
		return domain.Return{}, err
	}
	if err := json.Unmarshal(linesJSON, &ret.Lines); err != nil {
		return domain.Return{}, fmt.Errorf("decode return lines: %w", err)
	}
	return ret, nil
}

func scanOperationJournalEntry(row rowScanner) (domain.OperationJournalEntry, error) {
	var entry domain.OperationJournalEntry
	if err := row.Scan(
		&entry.ID, &entry.StoreID, &entry.OperationType, &entry.ActorID,
		&entry.ReferenceID, &entry.Summary, &entry.CreatedAt,
	); err != nil {
		return domain.OperationJournalEntry{}, err
	}
	return entry, nil
}
