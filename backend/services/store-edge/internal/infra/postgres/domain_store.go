package postgres

import (
	"context"
	"encoding/json"
	"fmt"

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
