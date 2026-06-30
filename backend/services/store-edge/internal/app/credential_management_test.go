package app_test

import (
	"context"
	"errors"
	"testing"

	"mercadia.dev/pos/services/store-edge/internal/app"
	"mercadia.dev/pos/services/store-edge/internal/domain"
	"mercadia.dev/pos/services/store-edge/internal/infra/memory"
)

func TestCredentialManagementAddsBindingForManagedActor(t *testing.T) {
	store := memory.NewStore(memory.WithDemoActors())
	credentials := app.NewCredentialManagementService(store)
	auth := app.NewAuthService(store, store)

	actor, err := credentials.AddCredentialBinding(context.Background(), app.AddCredentialBindingCommand{
		TargetActorID: "cashier-1",
		ManagerID:     "senior-1",
		Kind:          domain.CredentialKindBarcodeCard,
		Token:         "demo-cashier-barcode-2",
		MaskedToken:   "Barcode staff demo ****0002",
	})
	if err != nil {
		t.Fatalf("add binding: %v", err)
	}
	if len(actor.CredentialBindings) != 1 || actor.CredentialBindings[0].TokenFingerprint == "demo-cashier-barcode-2" {
		t.Fatalf("unsafe credential binding response = %+v", actor.CredentialBindings)
	}

	_, err = credentials.SetActorCredentialPolicy(context.Background(), app.SetActorCredentialPolicyCommand{
		TargetActorID: "cashier-1",
		ManagerID:     "senior-1",
		Required:      true,
		AllowedKinds:  []domain.CredentialKind{domain.CredentialKindBarcodeCard},
	})
	if err != nil {
		t.Fatalf("set actor policy: %v", err)
	}

	result, err := auth.CreateSession(context.Background(), app.CreateSessionCommand{
		ActorID: "cashier-1",
		PIN:     "1234",
		StoreID: "store-1",
		CredentialFactor: &domain.SubmittedCredentialFactor{
			Kind:  domain.CredentialKindBarcodeCard,
			Token: "demo-cashier-barcode-2",
		},
	})
	if err != nil {
		t.Fatalf("create session with managed binding: %v", err)
	}
	if result.CredentialFactor == nil || result.CredentialFactor.TokenFingerprint == "demo-cashier-barcode-2" {
		t.Fatalf("unsafe session credential factor = %+v", result.CredentialFactor)
	}
}

func TestCredentialManagementRejectsUnauthorizedAndSelfManagedChanges(t *testing.T) {
	store := memory.NewStore(memory.WithDemoActors())
	credentials := app.NewCredentialManagementService(store)

	_, err := credentials.AddCredentialBinding(context.Background(), app.AddCredentialBindingCommand{
		TargetActorID: "senior-1",
		ManagerID:     "cashier-1",
		Kind:          domain.CredentialKindIButton,
		Token:         "demo-token",
	})
	if !errors.Is(err, app.ErrPermissionDenied) {
		t.Fatalf("expected permission denied, got %v", err)
	}

	_, err = credentials.AddCredentialBinding(context.Background(), app.AddCredentialBindingCommand{
		TargetActorID: "senior-1",
		ManagerID:     "senior-1",
		Kind:          domain.CredentialKindIButton,
		Token:         "demo-token",
	})
	if !errors.Is(err, app.ErrSeparationOfDutiesViolation) {
		t.Fatalf("expected separation of duties violation, got %v", err)
	}
}

func TestCredentialManagementRevokesBinding(t *testing.T) {
	store := memory.NewStore(memory.WithDemoActors())
	credentials := app.NewCredentialManagementService(store)

	actor, err := credentials.GetCredentialManagement(context.Background(), app.GetCredentialManagementQuery{
		StoreID:   "store-1",
		ManagerID: "admin-1",
	})
	if err != nil {
		t.Fatalf("get credential management: %v", err)
	}
	var fingerprint string
	for _, candidate := range actor.Actors {
		if candidate.ID == "senior-1" {
			for _, binding := range candidate.CredentialBindings {
				if binding.Kind == domain.CredentialKindMSRCard {
					fingerprint = binding.TokenFingerprint
				}
			}
		}
	}
	if fingerprint == "" {
		t.Fatal("senior MSR binding fingerprint not found")
	}

	updated, err := credentials.RevokeCredentialBinding(context.Background(), app.RevokeCredentialBindingCommand{
		TargetActorID:    "senior-1",
		ManagerID:        "admin-1",
		Kind:             domain.CredentialKindMSRCard,
		TokenFingerprint: fingerprint,
	})
	if err != nil {
		t.Fatalf("revoke binding: %v", err)
	}
	for _, binding := range updated.CredentialBindings {
		if binding.Kind == domain.CredentialKindMSRCard && binding.TokenFingerprint == fingerprint && binding.Active {
			t.Fatalf("binding remains active = %+v", binding)
		}
	}
}

func TestCredentialManagementRejectsRequiredPolicyWithoutAllowedKinds(t *testing.T) {
	store := memory.NewStore(memory.WithDemoActors())
	credentials := app.NewCredentialManagementService(store)

	_, err := credentials.SetStoreCredentialPolicy(context.Background(), app.SetStoreCredentialPolicyCommand{
		StoreID:   "store-1",
		ManagerID: "admin-1",
		Required:  true,
	})
	if !errors.Is(err, app.ErrInvalidAuthCommand) {
		t.Fatalf("expected invalid auth command, got %v", err)
	}
}
