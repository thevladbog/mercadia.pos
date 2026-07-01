package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOpenAPIExposesStoreEdgeOperations(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)

	NewServer().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}

	var document map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &document); err != nil {
		t.Fatalf("decode OpenAPI: %v", err)
	}

	paths := document["paths"].(map[string]any)
	if _, ok := paths["/v1/receipts"]; !ok {
		t.Fatal("expected /v1/receipts path")
	}
	if _, ok := paths["/v1/receipts/{receiptId}/lines"]; !ok {
		t.Fatal("expected /v1/receipts/{receiptId}/lines path")
	}
	if _, ok := paths["/v1/receipts/{receiptId}/scan"]; !ok {
		t.Fatal("expected /v1/receipts/{receiptId}/scan path")
	}
	if _, ok := paths["/v1/receipts/{receiptId}/cancel"]; !ok {
		t.Fatal("expected /v1/receipts/{receiptId}/cancel path")
	}
	if _, ok := paths["/v1/catalog/products/by-barcode/{barcode}"]; !ok {
		t.Fatal("expected /v1/catalog/products/by-barcode/{barcode} path")
	}
	if _, ok := paths["/v1/terminals/{terminalId}"]; !ok {
		t.Fatal("expected /v1/terminals/{terminalId} path")
	}
	if _, ok := paths["/v1/stores/{storeId}/terminals"]; !ok {
		t.Fatal("expected /v1/stores/{storeId}/terminals path")
	}
	if _, ok := paths["/v1/stores/{storeId}/monitoring/terminals"]; !ok {
		t.Fatal("expected /v1/stores/{storeId}/monitoring/terminals path")
	}
	if _, ok := paths["/v1/stores/{storeId}/monitoring/summary"]; !ok {
		t.Fatal("expected /v1/stores/{storeId}/monitoring/summary path")
	}
	if _, ok := paths["/v1/stores/{storeId}/cash-movements"]; !ok {
		t.Fatal("expected /v1/stores/{storeId}/cash-movements path")
	}
	if _, ok := paths["/v1/stores/{storeId}/bank-collections"]; !ok {
		t.Fatal("expected /v1/stores/{storeId}/bank-collections path")
	}
	if _, ok := paths["/v1/stores/{storeId}/business-expenses"]; !ok {
		t.Fatal("expected /v1/stores/{storeId}/business-expenses path")
	}
	if _, ok := paths["/v1/stores/{storeId}/cash-balances"]; !ok {
		t.Fatal("expected /v1/stores/{storeId}/cash-balances path")
	}
	if _, ok := paths["/v1/stores/{storeId}/cash-recounts"]; !ok {
		t.Fatal("expected /v1/stores/{storeId}/cash-recounts path")
	}
	if _, ok := paths["/v1/stores/{storeId}/cash-recounts/{recountId}/resolve"]; !ok {
		t.Fatal("expected /v1/stores/{storeId}/cash-recounts/{recountId}/resolve path")
	}
	if _, ok := paths["/v1/store-edge/sync/outbox-status"]; !ok {
		t.Fatal("expected /v1/store-edge/sync/outbox-status path")
	}
	if _, ok := paths["/v1/operational-days"]; !ok {
		t.Fatal("expected /v1/operational-days path")
	}
	if _, ok := paths["/v1/operational-days/{operationalDayId}/close-check"]; !ok {
		t.Fatal("expected /v1/operational-days/{operationalDayId}/close-check path")
	}
	if _, ok := paths["/v1/operational-days/{operationalDayId}/receipts"]; !ok {
		t.Fatal("expected /v1/operational-days/{operationalDayId}/receipts path")
	}
	if _, ok := paths["/v1/operational-days/{operationalDayId}/shifts"]; !ok {
		t.Fatal("expected /v1/operational-days/{operationalDayId}/shifts path")
	}
	if _, ok := paths["/v1/operational-days/{operationalDayId}/summary"]; !ok {
		t.Fatal("expected /v1/operational-days/{operationalDayId}/summary path")
	}
	if _, ok := paths["/v1/operational-days/{operationalDayId}/close"]; !ok {
		t.Fatal("expected /v1/operational-days/{operationalDayId}/close path")
	}
	if _, ok := paths["/v1/shifts"]; !ok {
		t.Fatal("expected /v1/shifts path")
	}
	if _, ok := paths["/v1/shifts/{shiftId}/close"]; !ok {
		t.Fatal("expected /v1/shifts/{shiftId}/close path")
	}
	if _, ok := paths["/v1/shifts/{shiftId}/cash-in"]; !ok {
		t.Fatal("expected /v1/shifts/{shiftId}/cash-in path")
	}
	if _, ok := paths["/v1/shifts/{shiftId}/cash-out"]; !ok {
		t.Fatal("expected /v1/shifts/{shiftId}/cash-out path")
	}
	if _, ok := paths["/v1/shifts/{shiftId}/receipts"]; !ok {
		t.Fatal("expected /v1/shifts/{shiftId}/receipts path")
	}
	if _, ok := paths["/v1/stores/{storeId}/shifts/open"]; !ok {
		t.Fatal("expected /v1/stores/{storeId}/shifts/open path")
	}
	if _, ok := paths["/v1/auth/sessions"]; !ok {
		t.Fatal("expected /v1/auth/sessions path")
	}
	if _, ok := paths["/v1/receipts/{receiptId}/returns"]; !ok {
		t.Fatal("expected /v1/receipts/{receiptId}/returns path")
	}
	if _, ok := paths["/v1/returns/{returnId}"]; !ok {
		t.Fatal("expected /v1/returns/{returnId} path")
	}
	if _, ok := paths["/v1/stores/{storeId}/returns/no-receipt"]; !ok {
		t.Fatal("expected /v1/stores/{storeId}/returns/no-receipt path")
	}
	if _, ok := paths["/v1/stores/{storeId}/returns"]; !ok {
		t.Fatal("expected /v1/stores/{storeId}/returns path")
	}
	if _, ok := paths["/v1/receipts/{receiptId}/lines/{lineId}/discount"]; !ok {
		t.Fatal("expected /v1/receipts/{receiptId}/lines/{lineId}/discount path")
	}
	if _, ok := paths["/v1/receipts/{receiptId}/marking/validate"]; !ok {
		t.Fatal("expected /v1/receipts/{receiptId}/marking/validate path")
	}
	if _, ok := paths["/v1/stores/{storeId}/operation-journal"]; !ok {
		t.Fatal("expected /v1/stores/{storeId}/operation-journal path")
	}
	if _, ok := paths["/v1/stores/{storeId}/auth-settings"]; !ok {
		t.Fatal("expected /v1/stores/{storeId}/auth-settings path")
	}
	if _, ok := paths["/v1/stores/{storeId}/auth-attempts"]; !ok {
		t.Fatal("expected /v1/stores/{storeId}/auth-attempts path")
	}
	if _, ok := paths["/v1/stores/{storeId}/auth-lockouts/{actorId}/reset"]; !ok {
		t.Fatal("expected /v1/stores/{storeId}/auth-lockouts/{actorId}/reset path")
	}
	if _, ok := paths["/v1/stores/{storeId}/catalog/sync"]; !ok {
		t.Fatal("expected /v1/stores/{storeId}/catalog/sync path")
	}
	if _, ok := paths["/v1/receipts/{receiptId}/payments/{paymentId}/cancel"]; !ok {
		t.Fatal("expected /v1/receipts/{receiptId}/payments/{paymentId}/cancel path")
	}
	if _, ok := paths["/v1/receipts/{receiptId}/payments/{paymentId}/refund"]; !ok {
		t.Fatal("expected /v1/receipts/{receiptId}/payments/{paymentId}/refund path")
	}
	if _, ok := paths["/v1/returns/{returnId}/settle"]; !ok {
		t.Fatal("expected /v1/returns/{returnId}/settle path")
	}
}

func TestCredentialManagementRequiresSession(t *testing.T) {
	server := NewServer()

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/credential-management", nil)

	server.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("credential management without session status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestCredentialManagementUsesSessionActorForPermission(t *testing.T) {
	server := NewServer()
	cashierToken := createAuthSessionForTest(t, server, "cashier-1", "1234")

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/credential-management?actorId=admin-1", nil)
	request.Header.Set(sessionTokenHeader, cashierToken)

	server.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("credential management with spoofed actor status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestCredentialManagementRejectsBodyActorID(t *testing.T) {
	server := NewServer()
	adminToken := createAuthSessionForTest(t, server, "admin-1", "9999")

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/v1/stores/store-1/credential-policy", bytes.NewBufferString(`{
		"actorId": "cashier-1",
		"required": true,
		"allowedKinds": ["ibutton"]
	}`))
	request.Header.Set(sessionTokenHeader, adminToken)
	request.Header.Set("Idempotency-Key", "credential-policy-ignore-body-actor")

	server.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("credential policy with body actor status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestCredentialManagementReplaysPolicyIdempotency(t *testing.T) {
	server := NewServer()
	adminToken := createAuthSessionForTest(t, server, "admin-1", "9999")
	body := `{"required":true,"allowedKinds":["ibutton"]}`

	first := httptest.NewRecorder()
	firstRequest := httptest.NewRequest(http.MethodPut, "/v1/stores/store-1/credential-policy", bytes.NewBufferString(body))
	firstRequest.Header.Set(sessionTokenHeader, adminToken)
	firstRequest.Header.Set("Idempotency-Key", "credential-policy-replay")
	server.ServeHTTP(first, firstRequest)
	if first.Code != http.StatusOK {
		t.Fatalf("credential policy status = %d, body = %s", first.Code, first.Body.String())
	}

	replay := httptest.NewRecorder()
	replayRequest := httptest.NewRequest(http.MethodPut, "/v1/stores/store-1/credential-policy", bytes.NewBufferString(body))
	replayRequest.Header.Set(sessionTokenHeader, adminToken)
	replayRequest.Header.Set("Idempotency-Key", "credential-policy-replay")
	server.ServeHTTP(replay, replayRequest)
	if replay.Code != http.StatusOK {
		t.Fatalf("credential policy replay status = %d, body = %s", replay.Code, replay.Body.String())
	}
	if replay.Body.String() != first.Body.String() {
		t.Fatalf("credential policy replay body = %s, want %s", replay.Body.String(), first.Body.String())
	}

	reused := httptest.NewRecorder()
	reusedRequest := httptest.NewRequest(http.MethodPut, "/v1/stores/store-1/credential-policy", bytes.NewBufferString(`{"required":true,"allowedKinds":["msr_card"]}`))
	reusedRequest.Header.Set(sessionTokenHeader, adminToken)
	reusedRequest.Header.Set("Idempotency-Key", "credential-policy-replay")
	server.ServeHTTP(reused, reusedRequest)
	if reused.Code != http.StatusConflict {
		t.Fatalf("credential policy reused key status = %d, body = %s", reused.Code, reused.Body.String())
	}

	actorBody := `{"inheritStorePolicy":false,"required":true,"allowedKinds":["ibutton","msr_card"]}`
	actorFirst := httptest.NewRecorder()
	actorFirstRequest := httptest.NewRequest(http.MethodPut, "/v1/stores/store-1/actors/cashier-1/credential-policy", bytes.NewBufferString(actorBody))
	actorFirstRequest.Header.Set(sessionTokenHeader, adminToken)
	actorFirstRequest.Header.Set("Idempotency-Key", "actor-credential-policy-replay")
	server.ServeHTTP(actorFirst, actorFirstRequest)
	if actorFirst.Code != http.StatusOK {
		t.Fatalf("actor credential policy status = %d, body = %s", actorFirst.Code, actorFirst.Body.String())
	}

	actorReplay := httptest.NewRecorder()
	actorReplayRequest := httptest.NewRequest(http.MethodPut, "/v1/stores/store-1/actors/cashier-1/credential-policy", bytes.NewBufferString(`{"inheritStorePolicy":false,"required":true,"allowedKinds":["msr_card","ibutton"]}`))
	actorReplayRequest.Header.Set(sessionTokenHeader, adminToken)
	actorReplayRequest.Header.Set("Idempotency-Key", "actor-credential-policy-replay")
	server.ServeHTTP(actorReplay, actorReplayRequest)
	if actorReplay.Code != http.StatusOK {
		t.Fatalf("actor credential policy replay status = %d, body = %s", actorReplay.Code, actorReplay.Body.String())
	}
	if actorReplay.Body.String() != actorFirst.Body.String() {
		t.Fatalf("actor credential policy replay body = %s, want %s", actorReplay.Body.String(), actorFirst.Body.String())
	}

	actorReused := httptest.NewRecorder()
	actorReusedRequest := httptest.NewRequest(http.MethodPut, "/v1/stores/store-1/actors/cashier-1/credential-policy", bytes.NewBufferString(`{"inheritStorePolicy":false,"required":true,"allowedKinds":["barcode_card"]}`))
	actorReusedRequest.Header.Set(sessionTokenHeader, adminToken)
	actorReusedRequest.Header.Set("Idempotency-Key", "actor-credential-policy-replay")
	server.ServeHTTP(actorReused, actorReusedRequest)
	if actorReused.Code != http.StatusConflict {
		t.Fatalf("actor credential policy reused key status = %d, body = %s", actorReused.Code, actorReused.Body.String())
	}
}

func TestCredentialManagementReplaysBindingIdempotency(t *testing.T) {
	server := NewServer()
	adminToken := createAuthSessionForTest(t, server, "admin-1", "9999")
	body := `{"kind":"ibutton","token":"cashier-demo-ibutton","maskedToken":"iButton ****0001"}`

	first := httptest.NewRecorder()
	firstRequest := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/actors/cashier-1/credential-bindings", bytes.NewBufferString(body))
	firstRequest.Header.Set(sessionTokenHeader, adminToken)
	firstRequest.Header.Set("Idempotency-Key", "credential-binding-add-replay")
	server.ServeHTTP(first, firstRequest)
	if first.Code != http.StatusOK {
		t.Fatalf("add credential binding status = %d, body = %s", first.Code, first.Body.String())
	}
	var addResponse ActorCredentialAcceptedResponse
	if err := json.Unmarshal(first.Body.Bytes(), &addResponse); err != nil {
		t.Fatalf("decode add credential binding: %v", err)
	}
	if len(addResponse.Actor.CredentialBindings) != 1 {
		t.Fatalf("credential bindings = %+v", addResponse.Actor.CredentialBindings)
	}
	fingerprint := addResponse.Actor.CredentialBindings[0].TokenFingerprint

	replay := httptest.NewRecorder()
	replayRequest := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/actors/cashier-1/credential-bindings", bytes.NewBufferString(body))
	replayRequest.Header.Set(sessionTokenHeader, adminToken)
	replayRequest.Header.Set("Idempotency-Key", "credential-binding-add-replay")
	server.ServeHTTP(replay, replayRequest)
	if replay.Code != http.StatusOK {
		t.Fatalf("add credential binding replay status = %d, body = %s", replay.Code, replay.Body.String())
	}
	if replay.Body.String() != first.Body.String() {
		t.Fatalf("add credential binding replay body = %s, want %s", replay.Body.String(), first.Body.String())
	}

	reused := httptest.NewRecorder()
	reusedRequest := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/actors/cashier-1/credential-bindings", bytes.NewBufferString(`{"kind":"ibutton","token":"other-token"}`))
	reusedRequest.Header.Set(sessionTokenHeader, adminToken)
	reusedRequest.Header.Set("Idempotency-Key", "credential-binding-add-replay")
	server.ServeHTTP(reused, reusedRequest)
	if reused.Code != http.StatusConflict {
		t.Fatalf("add credential binding reused key status = %d, body = %s", reused.Code, reused.Body.String())
	}

	revokeBody := fmt.Sprintf(`{"kind":"ibutton","tokenFingerprint":"%s"}`, fingerprint)
	revoke := httptest.NewRecorder()
	revokeRequest := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/actors/cashier-1/credential-bindings/revoke", bytes.NewBufferString(revokeBody))
	revokeRequest.Header.Set(sessionTokenHeader, adminToken)
	revokeRequest.Header.Set("Idempotency-Key", "credential-binding-revoke-replay")
	server.ServeHTTP(revoke, revokeRequest)
	if revoke.Code != http.StatusOK {
		t.Fatalf("revoke credential binding status = %d, body = %s", revoke.Code, revoke.Body.String())
	}

	revokeReplay := httptest.NewRecorder()
	revokeReplayRequest := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/actors/cashier-1/credential-bindings/revoke", bytes.NewBufferString(revokeBody))
	revokeReplayRequest.Header.Set(sessionTokenHeader, adminToken)
	revokeReplayRequest.Header.Set("Idempotency-Key", "credential-binding-revoke-replay")
	server.ServeHTTP(revokeReplay, revokeReplayRequest)
	if revokeReplay.Code != http.StatusOK {
		t.Fatalf("revoke credential binding replay status = %d, body = %s", revokeReplay.Code, revokeReplay.Body.String())
	}
	if revokeReplay.Body.String() != revoke.Body.String() {
		t.Fatalf("revoke credential binding replay body = %s, want %s", revokeReplay.Body.String(), revoke.Body.String())
	}
}

func TestStoreAuthSettingsCanBeReadAndUpdatedByManagerSession(t *testing.T) {
	server := NewServer()

	readDefaults := httptest.NewRecorder()
	readDefaultsRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/auth-settings", nil)
	server.ServeHTTP(readDefaults, readDefaultsRequest)
	if readDefaults.Code != http.StatusOK {
		t.Fatalf("read default auth settings status = %d, body = %s", readDefaults.Code, readDefaults.Body.String())
	}
	var defaults StoreAuthSettingsAcceptedResponse
	if err := json.Unmarshal(readDefaults.Body.Bytes(), &defaults); err != nil {
		t.Fatalf("decode default auth settings: %v", err)
	}
	if defaults.Settings.FailedAttemptLimit != 5 || defaults.Settings.LockoutDurationSeconds != 900 || defaults.Settings.POSAutoLockSeconds != 300 {
		t.Fatalf("default auth settings = %+v", defaults.Settings)
	}

	unauthorized := httptest.NewRecorder()
	unauthorizedRequest := httptest.NewRequest(http.MethodPut, "/v1/stores/store-1/auth-settings", bytes.NewBufferString(`{
		"failedAttemptLimit": 3,
		"lockoutDurationSeconds": 600,
		"posAutoLockSeconds": 120
	}`))
	unauthorizedRequest.Header.Set("Idempotency-Key", "auth-settings-no-session")
	server.ServeHTTP(unauthorized, unauthorizedRequest)
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("auth settings without session status = %d, body = %s", unauthorized.Code, unauthorized.Body.String())
	}

	cashierToken := createAuthSessionForTest(t, server, "cashier-1", "1234")
	forbidden := httptest.NewRecorder()
	forbiddenRequest := httptest.NewRequest(http.MethodPut, "/v1/stores/store-1/auth-settings", bytes.NewBufferString(`{
		"failedAttemptLimit": 3,
		"lockoutDurationSeconds": 600,
		"posAutoLockSeconds": 120
	}`))
	forbiddenRequest.Header.Set(sessionTokenHeader, cashierToken)
	forbiddenRequest.Header.Set("Idempotency-Key", "auth-settings-cashier")
	server.ServeHTTP(forbidden, forbiddenRequest)
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("auth settings cashier status = %d, body = %s", forbidden.Code, forbidden.Body.String())
	}

	adminToken := createAuthSessionForTest(t, server, "admin-1", "9999")
	updated := httptest.NewRecorder()
	updatedRequest := httptest.NewRequest(http.MethodPut, "/v1/stores/store-1/auth-settings", bytes.NewBufferString(`{
		"failedAttemptLimit": 3,
		"lockoutDurationSeconds": 600,
		"posAutoLockSeconds": 120
	}`))
	updatedRequest.Header.Set(sessionTokenHeader, adminToken)
	updatedRequest.Header.Set("Idempotency-Key", "auth-settings-admin")
	server.ServeHTTP(updated, updatedRequest)
	if updated.Code != http.StatusOK {
		t.Fatalf("auth settings admin status = %d, body = %s", updated.Code, updated.Body.String())
	}
	var accepted StoreAuthSettingsAcceptedResponse
	if err := json.Unmarshal(updated.Body.Bytes(), &accepted); err != nil {
		t.Fatalf("decode updated auth settings: %v", err)
	}
	if accepted.Settings.FailedAttemptLimit != 3 || accepted.Settings.LockoutDurationSeconds != 600 || accepted.Settings.POSAutoLockSeconds != 120 || accepted.Settings.UpdatedByID != "admin-1" {
		t.Fatalf("updated auth settings = %+v", accepted.Settings)
	}
	if accepted.Settings.UpdatedAt == nil {
		t.Fatal("updated auth settings missing updatedAt")
	}

	replay := httptest.NewRecorder()
	replayRequest := httptest.NewRequest(http.MethodPut, "/v1/stores/store-1/auth-settings", bytes.NewBufferString(`{
		"failedAttemptLimit": 3,
		"lockoutDurationSeconds": 600,
		"posAutoLockSeconds": 120
	}`))
	replayRequest.Header.Set(sessionTokenHeader, adminToken)
	replayRequest.Header.Set("Idempotency-Key", "auth-settings-admin")
	server.ServeHTTP(replay, replayRequest)
	if replay.Code != http.StatusOK {
		t.Fatalf("auth settings replay status = %d, body = %s", replay.Code, replay.Body.String())
	}
	var replayed StoreAuthSettingsAcceptedResponse
	if err := json.Unmarshal(replay.Body.Bytes(), &replayed); err != nil {
		t.Fatalf("decode replayed auth settings: %v", err)
	}
	if replayed.Settings.UpdatedAt == nil || !replayed.Settings.UpdatedAt.Equal(*accepted.Settings.UpdatedAt) {
		t.Fatalf("replayed updatedAt = %v, want %v", replayed.Settings.UpdatedAt, accepted.Settings.UpdatedAt)
	}

	reused := httptest.NewRecorder()
	reusedRequest := httptest.NewRequest(http.MethodPut, "/v1/stores/store-1/auth-settings", bytes.NewBufferString(`{
		"failedAttemptLimit": 4,
		"lockoutDurationSeconds": 600,
		"posAutoLockSeconds": 120
	}`))
	reusedRequest.Header.Set(sessionTokenHeader, adminToken)
	reusedRequest.Header.Set("Idempotency-Key", "auth-settings-admin")
	server.ServeHTTP(reused, reusedRequest)
	if reused.Code != http.StatusConflict {
		t.Fatalf("auth settings reused key status = %d, body = %s", reused.Code, reused.Body.String())
	}
}

func TestAuthSessionLocksAfterConfiguredFailedAttempts(t *testing.T) {
	server := NewServer()
	adminToken := createAuthSessionForTest(t, server, "admin-1", "9999")

	settingsResponse := httptest.NewRecorder()
	settingsRequest := httptest.NewRequest(http.MethodPut, "/v1/stores/store-1/auth-settings", bytes.NewBufferString(`{
		"failedAttemptLimit": 2,
		"lockoutDurationSeconds": 600,
		"posAutoLockSeconds": 120
	}`))
	settingsRequest.Header.Set(sessionTokenHeader, adminToken)
	settingsRequest.Header.Set("Idempotency-Key", "auth-settings-lockout")
	server.ServeHTTP(settingsResponse, settingsRequest)
	if settingsResponse.Code != http.StatusOK {
		t.Fatalf("configure auth settings status = %d, body = %s", settingsResponse.Code, settingsResponse.Body.String())
	}

	first := httptest.NewRecorder()
	firstRequest := httptest.NewRequest(http.MethodPost, "/v1/auth/sessions", bytes.NewBufferString(`{
		"actorId": "cashier-1",
		"pin": "bad-pin",
		"storeId": "store-1",
		"terminalId": "pos-1"
	}`))
	server.ServeHTTP(first, firstRequest)
	if first.Code != http.StatusUnauthorized {
		t.Fatalf("first invalid login status = %d, body = %s", first.Code, first.Body.String())
	}

	second := httptest.NewRecorder()
	secondRequest := httptest.NewRequest(http.MethodPost, "/v1/auth/sessions", bytes.NewBufferString(`{
		"actorId": "cashier-1",
		"pin": "bad-pin",
		"storeId": "store-1",
		"terminalId": "pos-1"
	}`))
	server.ServeHTTP(second, secondRequest)
	if second.Code != http.StatusLocked {
		t.Fatalf("second invalid login status = %d, body = %s", second.Code, second.Body.String())
	}

	correctWhileLocked := httptest.NewRecorder()
	correctWhileLockedRequest := httptest.NewRequest(http.MethodPost, "/v1/auth/sessions", bytes.NewBufferString(`{
		"actorId": "cashier-1",
		"pin": "1234",
		"storeId": "store-1",
		"terminalId": "pos-1"
	}`))
	server.ServeHTTP(correctWhileLocked, correctWhileLockedRequest)
	if correctWhileLocked.Code != http.StatusLocked {
		t.Fatalf("correct login while locked status = %d, body = %s", correctWhileLocked.Code, correctWhileLocked.Body.String())
	}

	audit := httptest.NewRecorder()
	auditRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/auth-attempts?actorId=cashier-1&limit=10", nil)
	auditRequest.Header.Set(sessionTokenHeader, adminToken)
	server.ServeHTTP(audit, auditRequest)
	if audit.Code != http.StatusOK {
		t.Fatalf("auth attempts status = %d, body = %s", audit.Code, audit.Body.String())
	}
	var auditResponse AuthAttemptsResponse
	if err := json.Unmarshal(audit.Body.Bytes(), &auditResponse); err != nil {
		t.Fatalf("decode auth attempts: %v", err)
	}
	if auditResponse.TotalCount < 3 || len(auditResponse.Items) == 0 {
		t.Fatalf("auth attempts response = %+v", auditResponse)
	}
	for _, item := range auditResponse.Items {
		if item.ActorID != "cashier-1" || item.StoreID != "store-1" {
			t.Fatalf("unexpected auth attempt item = %+v", item)
		}
	}

	reset := httptest.NewRecorder()
	resetRequest := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/auth-lockouts/cashier-1/reset", bytes.NewBufferString(`{"reason":"manager verified cashier"}`))
	resetRequest.Header.Set(sessionTokenHeader, adminToken)
	resetRequest.Header.Set("Idempotency-Key", "auth-lockout-reset-cashier")
	server.ServeHTTP(reset, resetRequest)
	if reset.Code != http.StatusOK {
		t.Fatalf("reset auth lockout status = %d, body = %s", reset.Code, reset.Body.String())
	}

	resetReplay := httptest.NewRecorder()
	resetReplayRequest := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/auth-lockouts/cashier-1/reset", bytes.NewBufferString(`{"reason":"manager verified cashier"}`))
	resetReplayRequest.Header.Set(sessionTokenHeader, adminToken)
	resetReplayRequest.Header.Set("Idempotency-Key", "auth-lockout-reset-cashier")
	server.ServeHTTP(resetReplay, resetReplayRequest)
	if resetReplay.Code != http.StatusOK {
		t.Fatalf("reset auth lockout replay status = %d, body = %s", resetReplay.Code, resetReplay.Body.String())
	}
	if resetReplay.Body.String() != reset.Body.String() {
		t.Fatalf("reset auth lockout replay body = %s, want %s", resetReplay.Body.String(), reset.Body.String())
	}

	correctAfterReset := httptest.NewRecorder()
	correctAfterResetRequest := httptest.NewRequest(http.MethodPost, "/v1/auth/sessions", bytes.NewBufferString(`{
		"actorId": "cashier-1",
		"pin": "1234",
		"storeId": "store-1",
		"terminalId": "pos-1"
	}`))
	server.ServeHTTP(correctAfterReset, correctAfterResetRequest)
	if correctAfterReset.Code != http.StatusCreated {
		t.Fatalf("correct login after reset status = %d, body = %s", correctAfterReset.Code, correctAfterReset.Body.String())
	}
}

func TestAuthAttemptsAuditRequiresPermissionAndDoesNotExposeRawCredential(t *testing.T) {
	server := NewServer()

	failed := httptest.NewRecorder()
	failedRequest := httptest.NewRequest(http.MethodPost, "/v1/auth/sessions", bytes.NewBufferString(`{
		"actorId": "cashier-1",
		"pin": "1234",
		"storeId": "store-1",
		"terminalId": "pos-1",
		"credentialFactor": {"kind": "barcode_card", "token": "raw-secret-token"}
	}`))
	server.ServeHTTP(failed, failedRequest)
	if failed.Code != http.StatusUnauthorized {
		t.Fatalf("failed credential login status = %d, body = %s", failed.Code, failed.Body.String())
	}

	cashierToken := createAuthSessionForTest(t, server, "cashier-1", "1234")
	forbidden := httptest.NewRecorder()
	forbiddenRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/auth-attempts", nil)
	forbiddenRequest.Header.Set(sessionTokenHeader, cashierToken)
	server.ServeHTTP(forbidden, forbiddenRequest)
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("auth attempts cashier status = %d, body = %s", forbidden.Code, forbidden.Body.String())
	}

	adminToken := createAuthSessionForTest(t, server, "admin-1", "9999")
	audit := httptest.NewRecorder()
	auditRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/auth-attempts?actorId=cashier-1&successful=false", nil)
	auditRequest.Header.Set(sessionTokenHeader, adminToken)
	server.ServeHTTP(audit, auditRequest)
	if audit.Code != http.StatusOK {
		t.Fatalf("auth attempts admin status = %d, body = %s", audit.Code, audit.Body.String())
	}
	if bytes.Contains(audit.Body.Bytes(), []byte("raw-secret-token")) {
		t.Fatalf("auth attempt audit exposed raw credential: %s", audit.Body.String())
	}
	var response AuthAttemptsResponse
	if err := json.Unmarshal(audit.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode auth attempts: %v", err)
	}
	if response.TotalCount == 0 || response.Items[0].CredentialFingerprint == "" || response.Items[0].CredentialKind != "barcode_card" {
		t.Fatalf("auth attempts response = %+v", response)
	}
}

func TestReceiptWorkflow(t *testing.T) {
	server := NewServer()
	openStoreDayAndShift(t, server, "receipt")

	openResponse := httptest.NewRecorder()
	openRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts", bytes.NewBufferString(`{
		"storeId": "store-1",
		"terminalId": "pos-1",
		"cashierId": "cashier-1",
		"channel": "pos"
	}`))
	openRequest.Header.Set("Idempotency-Key", "open-1")

	server.ServeHTTP(openResponse, openRequest)

	if openResponse.Code != http.StatusAccepted {
		t.Fatalf("open receipt status = %d, body = %s", openResponse.Code, openResponse.Body.String())
	}

	var opened ReceiptAcceptedResponse
	if err := json.Unmarshal(openResponse.Body.Bytes(), &opened); err != nil {
		t.Fatalf("decode open response: %v", err)
	}
	if opened.Receipt.ID == "" {
		t.Fatal("expected receipt id")
	}
	if opened.Receipt.OperationalDayID == "" || opened.Receipt.ShiftID == "" || opened.Receipt.DrawerID != "drawer-1" {
		t.Fatalf("expected receipt store operation links, got %+v", opened.Receipt)
	}

	addLineResponse := httptest.NewRecorder()
	addLineRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+opened.Receipt.ID+"/lines", bytes.NewBufferString(`{
		"productId": "sku-1",
		"barcode": "4600000000000",
		"name": "Milk",
		"quantity": 2,
		"unitPriceMinor": 19999
	}`))
	addLineRequest.Header.Set("Idempotency-Key", "line-1")

	server.ServeHTTP(addLineResponse, addLineRequest)

	if addLineResponse.Code != http.StatusAccepted {
		t.Fatalf("add line status = %d, body = %s", addLineResponse.Code, addLineResponse.Body.String())
	}

	var updated ReceiptAcceptedResponse
	if err := json.Unmarshal(addLineResponse.Body.Bytes(), &updated); err != nil {
		t.Fatalf("decode add line response: %v", err)
	}
	if updated.Receipt.TotalMinor != 39998 {
		t.Fatalf("total minor = %d", updated.Receipt.TotalMinor)
	}

	getResponse := httptest.NewRecorder()
	getRequest := httptest.NewRequest(http.MethodGet, "/v1/receipts/"+opened.Receipt.ID, nil)

	server.ServeHTTP(getResponse, getRequest)

	if getResponse.Code != http.StatusOK {
		t.Fatalf("get receipt status = %d, body = %s", getResponse.Code, getResponse.Body.String())
	}

	var receipt ReceiptResponse
	if err := json.Unmarshal(getResponse.Body.Bytes(), &receipt); err != nil {
		t.Fatalf("decode get response: %v", err)
	}
	if receipt.TotalMinor != 39998 || len(receipt.Lines) != 1 {
		t.Fatalf("receipt total = %d, lines = %d", receipt.TotalMinor, len(receipt.Lines))
	}
}

func TestOpenReceiptRequiresOpenOperationalDayAndShift(t *testing.T) {
	server := NewServer()

	openResponse := httptest.NewRecorder()
	openRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts", bytes.NewBufferString(`{
		"storeId": "store-1",
		"terminalId": "pos-1",
		"cashierId": "cashier-1",
		"channel": "pos"
	}`))
	openRequest.Header.Set("Idempotency-Key", "open-1")

	server.ServeHTTP(openResponse, openRequest)

	if openResponse.Code != http.StatusConflict {
		t.Fatalf("open receipt without shift status = %d, body = %s", openResponse.Code, openResponse.Body.String())
	}
}

func TestCancelReceiptWorkflow(t *testing.T) {
	server := NewServer()
	openStoreDayAndShift(t, server, "cancel")

	openResponse := httptest.NewRecorder()
	openRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts", bytes.NewBufferString(`{
		"storeId": "store-1",
		"terminalId": "pos-1",
		"cashierId": "cashier-1",
		"channel": "pos"
	}`))
	openRequest.Header.Set("Idempotency-Key", "cancel-open-1")

	server.ServeHTTP(openResponse, openRequest)

	if openResponse.Code != http.StatusAccepted {
		t.Fatalf("open receipt status = %d, body = %s", openResponse.Code, openResponse.Body.String())
	}

	var opened ReceiptAcceptedResponse
	if err := json.Unmarshal(openResponse.Body.Bytes(), &opened); err != nil {
		t.Fatalf("decode open response: %v", err)
	}

	cancelResponse := httptest.NewRecorder()
	cancelRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+opened.Receipt.ID+"/cancel", bytes.NewBufferString(`{
		"reason": "Customer changed mind",
		"actorId": "cashier-1"
	}`))
	cancelRequest.Header.Set("Idempotency-Key", "cancel-1")

	server.ServeHTTP(cancelResponse, cancelRequest)

	if cancelResponse.Code != http.StatusAccepted {
		t.Fatalf("cancel receipt status = %d, body = %s", cancelResponse.Code, cancelResponse.Body.String())
	}

	var cancelled ReceiptAcceptedResponse
	if err := json.Unmarshal(cancelResponse.Body.Bytes(), &cancelled); err != nil {
		t.Fatalf("decode cancel response: %v", err)
	}
	if cancelled.Receipt.Status != "cancelled" ||
		cancelled.Receipt.CancelReason != "Customer changed mind" ||
		cancelled.Receipt.CancelledByID != "cashier-1" {
		t.Fatalf("cancelled receipt = %+v", cancelled.Receipt)
	}
}

func TestTerminalHeartbeatWorkflow(t *testing.T) {
	server := NewServer()

	heartbeatResponse := httptest.NewRecorder()
	heartbeatRequest := httptest.NewRequest(http.MethodPost, "/v1/terminals/pos-1/heartbeat", bytes.NewBufferString(`{
		"storeId": "store-1",
		"kind": "pos",
		"softwareVersion": "0.1.0"
	}`))
	heartbeatRequest.Header.Set("Idempotency-Key", "heartbeat-1")

	server.ServeHTTP(heartbeatResponse, heartbeatRequest)

	if heartbeatResponse.Code != http.StatusAccepted {
		t.Fatalf("heartbeat status = %d, body = %s", heartbeatResponse.Code, heartbeatResponse.Body.String())
	}

	var accepted HeartbeatResponse
	if err := json.Unmarshal(heartbeatResponse.Body.Bytes(), &accepted); err != nil {
		t.Fatalf("decode heartbeat response: %v", err)
	}
	if accepted.Terminal.ID != "pos-1" || accepted.Terminal.Status != "online" {
		t.Fatalf("terminal response = %+v", accepted.Terminal)
	}

	getResponse := httptest.NewRecorder()
	getRequest := httptest.NewRequest(http.MethodGet, "/v1/terminals/pos-1", nil)

	server.ServeHTTP(getResponse, getRequest)

	if getResponse.Code != http.StatusOK {
		t.Fatalf("get terminal status = %d, body = %s", getResponse.Code, getResponse.Body.String())
	}

	var terminal TerminalResponse
	if err := json.Unmarshal(getResponse.Body.Bytes(), &terminal); err != nil {
		t.Fatalf("decode terminal response: %v", err)
	}
	if terminal.ID != "pos-1" || terminal.Kind != "pos" || terminal.SoftwareVersion != "0.1.0" {
		t.Fatalf("terminal = %+v", terminal)
	}
}

func TestListStoreTerminalsReturnsHeartbeats(t *testing.T) {
	server := NewServer()

	for _, spec := range []struct {
		terminalID      string
		idempotencyKey  string
		softwareVersion string
	}{
		{terminalID: "pos-1", idempotencyKey: "heartbeat-1", softwareVersion: "0.1.0"},
		{terminalID: "pos-2", idempotencyKey: "heartbeat-2", softwareVersion: "0.2.0"},
	} {
		request := httptest.NewRequest(http.MethodPost, "/v1/terminals/"+spec.terminalID+"/heartbeat", bytes.NewBufferString(`{
			"storeId": "store-1",
			"kind": "pos",
			"softwareVersion": "`+spec.softwareVersion+`"
		}`))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Idempotency-Key", spec.idempotencyKey)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		if response.Code != http.StatusAccepted {
			t.Fatalf("heartbeat %s status = %d body = %s", spec.terminalID, response.Code, response.Body.String())
		}
	}

	listResponse := httptest.NewRecorder()
	listRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/terminals", nil)
	server.ServeHTTP(listResponse, listRequest)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("list terminals status = %d body = %s", listResponse.Code, listResponse.Body.String())
	}

	var listed PaginatedTerminalsResponse
	if err := json.Unmarshal(listResponse.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode list terminals response: %v", err)
	}
	if listed.TotalCount != 2 || len(listed.Items) != 2 {
		t.Fatalf("listed terminals = %+v", listed)
	}
	if listed.Items[0].ID != "pos-1" || listed.Items[1].ID != "pos-2" {
		t.Fatalf("terminal order = %+v", listed.Items)
	}
}

func TestStoreMonitoringEndpoints(t *testing.T) {
	server := NewServer()

	heartbeatRequest := httptest.NewRequest(http.MethodPost, "/v1/terminals/pos-1/heartbeat", bytes.NewBufferString(`{
		"storeId": "store-1",
		"kind": "pos",
		"softwareVersion": "0.1.0"
	}`))
	heartbeatRequest.Header.Set("Content-Type", "application/json")
	heartbeatRequest.Header.Set("Idempotency-Key", "monitoring-heartbeat-1")
	heartbeatResponse := httptest.NewRecorder()
	server.ServeHTTP(heartbeatResponse, heartbeatRequest)
	if heartbeatResponse.Code != http.StatusAccepted {
		t.Fatalf("heartbeat status = %d body = %s", heartbeatResponse.Code, heartbeatResponse.Body.String())
	}

	openStoreDayAndShift(t, server, "monitoring")

	openReceiptResponse := httptest.NewRecorder()
	openReceiptRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts", bytes.NewBufferString(`{
		"storeId": "store-1",
		"terminalId": "pos-1",
		"cashierId": "cashier-1",
		"channel": "pos"
	}`))
	openReceiptRequest.Header.Set("Content-Type", "application/json")
	openReceiptRequest.Header.Set("Idempotency-Key", "monitoring-receipt-open-1")
	server.ServeHTTP(openReceiptResponse, openReceiptRequest)
	if openReceiptResponse.Code != http.StatusAccepted {
		t.Fatalf("open receipt status = %d body = %s", openReceiptResponse.Code, openReceiptResponse.Body.String())
	}

	var openedReceipt ReceiptAcceptedResponse
	if err := json.Unmarshal(openReceiptResponse.Body.Bytes(), &openedReceipt); err != nil {
		t.Fatalf("decode open receipt response: %v", err)
	}

	addLineResponse := httptest.NewRecorder()
	addLineRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/lines", bytes.NewBufferString(`{
		"productId": "sku-1",
		"name": "Milk",
		"quantity": 1,
		"unitPriceMinor": 50000
	}`))
	addLineRequest.Header.Set("Content-Type", "application/json")
	addLineRequest.Header.Set("Idempotency-Key", "monitoring-receipt-line-1")
	server.ServeHTTP(addLineResponse, addLineRequest)
	if addLineResponse.Code != http.StatusAccepted {
		t.Fatalf("add line status = %d body = %s", addLineResponse.Code, addLineResponse.Body.String())
	}

	paymentResponse := httptest.NewRecorder()
	paymentRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/payments", bytes.NewBufferString(`{
		"method": "card_mock",
		"amountMinor": 50000,
		"providerReference": "monitoring-card-1"
	}`))
	paymentRequest.Header.Set("Content-Type", "application/json")
	paymentRequest.Header.Set("Idempotency-Key", "monitoring-payment-1")
	server.ServeHTTP(paymentResponse, paymentRequest)
	if paymentResponse.Code != http.StatusAccepted {
		t.Fatalf("create payment status = %d body = %s", paymentResponse.Code, paymentResponse.Body.String())
	}

	fiscalResponse := httptest.NewRecorder()
	fiscalRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/fiscal-documents", bytes.NewBufferString(`{
		"deviceId": "mock-atol-1"
	}`))
	fiscalRequest.Header.Set("Content-Type", "application/json")
	fiscalRequest.Header.Set("Idempotency-Key", "monitoring-fiscal-1")
	server.ServeHTTP(fiscalResponse, fiscalRequest)
	if fiscalResponse.Code != http.StatusAccepted {
		t.Fatalf("create fiscal document status = %d body = %s", fiscalResponse.Code, fiscalResponse.Body.String())
	}

	terminalsResponse := httptest.NewRecorder()
	terminalsRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/monitoring/terminals", nil)
	server.ServeHTTP(terminalsResponse, terminalsRequest)
	if terminalsResponse.Code != http.StatusOK {
		t.Fatalf("monitoring terminals status = %d body = %s", terminalsResponse.Code, terminalsResponse.Body.String())
	}

	var terminals PaginatedTerminalOverviewResponse
	if err := json.Unmarshal(terminalsResponse.Body.Bytes(), &terminals); err != nil {
		t.Fatalf("decode monitoring terminals: %v", err)
	}
	if terminals.TotalCount != 1 || len(terminals.Items) != 1 {
		t.Fatalf("monitoring terminals = %+v", terminals)
	}
	item := terminals.Items[0]
	if item.ShiftID == "" || item.CashierID != "cashier-1" || item.ReceiptCount != 1 || item.RevenueMinor != 50000 {
		t.Fatalf("monitoring terminal card = %+v", item)
	}

	summaryResponse := httptest.NewRecorder()
	summaryRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/monitoring/summary", nil)
	server.ServeHTTP(summaryResponse, summaryRequest)
	if summaryResponse.Code != http.StatusOK {
		t.Fatalf("monitoring summary status = %d body = %s", summaryResponse.Code, summaryResponse.Body.String())
	}

	var summary StoreMonitoringSummaryResponse
	if err := json.Unmarshal(summaryResponse.Body.Bytes(), &summary); err != nil {
		t.Fatalf("decode monitoring summary: %v", err)
	}
	if summary.ActiveTerminalCount != 1 || summary.ReceiptCountToday != 1 || summary.RevenueMinorToday != 50000 {
		t.Fatalf("monitoring summary = %+v", summary)
	}
}

func TestCancelReceiptPayment(t *testing.T) {
	server := NewServer()
	openStoreDayAndShiftForDate(t, server, "payment-cancel", time.Now().UTC().Format("2006-01-02"))

	openReceiptResponse := httptest.NewRecorder()
	openReceiptRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts", bytes.NewBufferString(`{
		"storeId": "store-1",
		"terminalId": "pos-1",
		"cashierId": "cashier-1",
		"channel": "pos"
	}`))
	openReceiptRequest.Header.Set("Content-Type", "application/json")
	openReceiptRequest.Header.Set("Idempotency-Key", "payment-cancel-receipt-open-1")
	server.ServeHTTP(openReceiptResponse, openReceiptRequest)
	if openReceiptResponse.Code != http.StatusAccepted {
		t.Fatalf("open receipt status = %d body = %s", openReceiptResponse.Code, openReceiptResponse.Body.String())
	}

	var openedReceipt ReceiptAcceptedResponse
	if err := json.Unmarshal(openReceiptResponse.Body.Bytes(), &openedReceipt); err != nil {
		t.Fatalf("decode open receipt response: %v", err)
	}

	addLineResponse := httptest.NewRecorder()
	addLineRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/lines", bytes.NewBufferString(`{
		"productId": "sku-1",
		"name": "Milk",
		"quantity": 1,
		"unitPriceMinor": 50000
	}`))
	addLineRequest.Header.Set("Content-Type", "application/json")
	addLineRequest.Header.Set("Idempotency-Key", "payment-cancel-line-1")
	server.ServeHTTP(addLineResponse, addLineRequest)
	if addLineResponse.Code != http.StatusAccepted {
		t.Fatalf("add line status = %d body = %s", addLineResponse.Code, addLineResponse.Body.String())
	}

	paymentResponse := httptest.NewRecorder()
	paymentRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/payments", bytes.NewBufferString(`{
		"method": "card_mock",
		"amountMinor": 50000
	}`))
	paymentRequest.Header.Set("Content-Type", "application/json")
	paymentRequest.Header.Set("Idempotency-Key", "payment-cancel-payment-1")
	server.ServeHTTP(paymentResponse, paymentRequest)
	if paymentResponse.Code != http.StatusAccepted {
		t.Fatalf("create payment status = %d body = %s", paymentResponse.Code, paymentResponse.Body.String())
	}

	var accepted PaymentAcceptedResponse
	if err := json.Unmarshal(paymentResponse.Body.Bytes(), &accepted); err != nil {
		t.Fatalf("decode payment response: %v", err)
	}

	cancelResponse := httptest.NewRecorder()
	cancelRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/payments/"+accepted.Payment.ID+"/cancel", bytes.NewBufferString(`{
		"actorId": "cashier-1",
		"reason": "Customer changed mind"
	}`))
	cancelRequest.Header.Set("Content-Type", "application/json")
	cancelRequest.Header.Set("Idempotency-Key", "payment-cancel-cancel-1")
	server.ServeHTTP(cancelResponse, cancelRequest)
	if cancelResponse.Code != http.StatusAccepted {
		t.Fatalf("cancel payment status = %d body = %s", cancelResponse.Code, cancelResponse.Body.String())
	}

	var cancelled PaymentAcceptedResponse
	if err := json.Unmarshal(cancelResponse.Body.Bytes(), &cancelled); err != nil {
		t.Fatalf("decode cancel payment response: %v", err)
	}
	if cancelled.Payment.Status != "cancelled" {
		t.Fatalf("payment status = %s", cancelled.Payment.Status)
	}

	getReceiptResponse := httptest.NewRecorder()
	getReceiptRequest := httptest.NewRequest(http.MethodGet, "/v1/receipts/"+openedReceipt.Receipt.ID, nil)
	server.ServeHTTP(getReceiptResponse, getReceiptRequest)
	if getReceiptResponse.Code != http.StatusOK {
		t.Fatalf("get receipt status = %d body = %s", getReceiptResponse.Code, getReceiptResponse.Body.String())
	}

	var receipt ReceiptResponse
	if err := json.Unmarshal(getReceiptResponse.Body.Bytes(), &receipt); err != nil {
		t.Fatalf("decode receipt response: %v", err)
	}
	if receipt.Status != "draft" {
		t.Fatalf("receipt status = %s", receipt.Status)
	}
}

func TestCancelReceiptCashPayment(t *testing.T) {
	server := NewServer()
	openStoreDayAndShiftForDate(t, server, "payment-cancel-cash", time.Now().UTC().Format("2006-01-02"))

	openReceiptResponse := httptest.NewRecorder()
	openReceiptRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts", bytes.NewBufferString(`{
		"storeId": "store-1",
		"terminalId": "pos-1",
		"cashierId": "cashier-1",
		"channel": "pos"
	}`))
	openReceiptRequest.Header.Set("Content-Type", "application/json")
	openReceiptRequest.Header.Set("Idempotency-Key", "payment-cancel-cash-receipt-open-1")
	server.ServeHTTP(openReceiptResponse, openReceiptRequest)
	if openReceiptResponse.Code != http.StatusAccepted {
		t.Fatalf("open receipt status = %d body = %s", openReceiptResponse.Code, openReceiptResponse.Body.String())
	}

	var openedReceipt ReceiptAcceptedResponse
	if err := json.Unmarshal(openReceiptResponse.Body.Bytes(), &openedReceipt); err != nil {
		t.Fatalf("decode open receipt response: %v", err)
	}

	addLineResponse := httptest.NewRecorder()
	addLineRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/lines", bytes.NewBufferString(`{
		"productId": "sku-1",
		"name": "Milk",
		"quantity": 1,
		"unitPriceMinor": 50000
	}`))
	addLineRequest.Header.Set("Content-Type", "application/json")
	addLineRequest.Header.Set("Idempotency-Key", "payment-cancel-cash-line-1")
	server.ServeHTTP(addLineResponse, addLineRequest)
	if addLineResponse.Code != http.StatusAccepted {
		t.Fatalf("add line status = %d body = %s", addLineResponse.Code, addLineResponse.Body.String())
	}

	paymentResponse := httptest.NewRecorder()
	paymentRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/payments", bytes.NewBufferString(`{
		"method": "cash",
		"amountMinor": 50000
	}`))
	paymentRequest.Header.Set("Content-Type", "application/json")
	paymentRequest.Header.Set("Idempotency-Key", "payment-cancel-cash-payment-1")
	server.ServeHTTP(paymentResponse, paymentRequest)
	if paymentResponse.Code != http.StatusAccepted {
		t.Fatalf("create payment status = %d body = %s", paymentResponse.Code, paymentResponse.Body.String())
	}

	var accepted PaymentAcceptedResponse
	if err := json.Unmarshal(paymentResponse.Body.Bytes(), &accepted); err != nil {
		t.Fatalf("decode payment response: %v", err)
	}

	cancelResponse := httptest.NewRecorder()
	cancelRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/payments/"+accepted.Payment.ID+"/cancel", bytes.NewBufferString(`{
		"actorId": "cashier-1",
		"reason": "Customer changed mind"
	}`))
	cancelRequest.Header.Set("Content-Type", "application/json")
	cancelRequest.Header.Set("Idempotency-Key", "payment-cancel-cash-cancel-1")
	server.ServeHTTP(cancelResponse, cancelRequest)
	if cancelResponse.Code != http.StatusAccepted {
		t.Fatalf("cancel payment status = %d body = %s", cancelResponse.Code, cancelResponse.Body.String())
	}

	var cancelled PaymentAcceptedResponse
	if err := json.Unmarshal(cancelResponse.Body.Bytes(), &cancelled); err != nil {
		t.Fatalf("decode cancel payment response: %v", err)
	}
	if cancelled.Payment.Status != "cancelled" {
		t.Fatalf("payment status = %s", cancelled.Payment.Status)
	}

	getReceiptResponse := httptest.NewRecorder()
	getReceiptRequest := httptest.NewRequest(http.MethodGet, "/v1/receipts/"+openedReceipt.Receipt.ID, nil)
	server.ServeHTTP(getReceiptResponse, getReceiptRequest)
	if getReceiptResponse.Code != http.StatusOK {
		t.Fatalf("get receipt status = %d body = %s", getReceiptResponse.Code, getReceiptResponse.Body.String())
	}

	var receipt ReceiptResponse
	if err := json.Unmarshal(getReceiptResponse.Body.Bytes(), &receipt); err != nil {
		t.Fatalf("decode receipt response: %v", err)
	}
	if receipt.Status != "draft" {
		t.Fatalf("receipt status = %s", receipt.Status)
	}
}

func TestRefundReceiptPayment(t *testing.T) {
	server := NewServer()
	openStoreDayAndShiftForDate(t, server, "payment-refund", time.Now().UTC().Format("2006-01-02"))

	openReceiptResponse := httptest.NewRecorder()
	openReceiptRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts", bytes.NewBufferString(`{
		"storeId": "store-1",
		"terminalId": "pos-1",
		"cashierId": "cashier-1",
		"channel": "pos"
	}`))
	openReceiptRequest.Header.Set("Content-Type", "application/json")
	openReceiptRequest.Header.Set("Idempotency-Key", "payment-refund-receipt-open-1")
	server.ServeHTTP(openReceiptResponse, openReceiptRequest)
	if openReceiptResponse.Code != http.StatusAccepted {
		t.Fatalf("open receipt status = %d body = %s", openReceiptResponse.Code, openReceiptResponse.Body.String())
	}

	var openedReceipt ReceiptAcceptedResponse
	if err := json.Unmarshal(openReceiptResponse.Body.Bytes(), &openedReceipt); err != nil {
		t.Fatalf("decode open receipt response: %v", err)
	}

	addLineResponse := httptest.NewRecorder()
	addLineRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/lines", bytes.NewBufferString(`{
		"productId": "sku-1",
		"name": "Milk",
		"quantity": 1,
		"unitPriceMinor": 50000
	}`))
	addLineRequest.Header.Set("Content-Type", "application/json")
	addLineRequest.Header.Set("Idempotency-Key", "payment-refund-line-1")
	server.ServeHTTP(addLineResponse, addLineRequest)
	if addLineResponse.Code != http.StatusAccepted {
		t.Fatalf("add line status = %d body = %s", addLineResponse.Code, addLineResponse.Body.String())
	}

	paymentResponse := httptest.NewRecorder()
	paymentRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/payments", bytes.NewBufferString(`{
		"method": "card_mock",
		"amountMinor": 50000
	}`))
	paymentRequest.Header.Set("Content-Type", "application/json")
	paymentRequest.Header.Set("Idempotency-Key", "payment-refund-payment-1")
	server.ServeHTTP(paymentResponse, paymentRequest)
	if paymentResponse.Code != http.StatusAccepted {
		t.Fatalf("create payment status = %d body = %s", paymentResponse.Code, paymentResponse.Body.String())
	}

	var accepted PaymentAcceptedResponse
	if err := json.Unmarshal(paymentResponse.Body.Bytes(), &accepted); err != nil {
		t.Fatalf("decode payment response: %v", err)
	}

	fiscalResponse := httptest.NewRecorder()
	fiscalRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/fiscal-documents", bytes.NewBufferString(`{
		"deviceId": "mock-atol-1"
	}`))
	fiscalRequest.Header.Set("Content-Type", "application/json")
	fiscalRequest.Header.Set("Idempotency-Key", "payment-refund-fiscal-1")
	server.ServeHTTP(fiscalResponse, fiscalRequest)
	if fiscalResponse.Code != http.StatusAccepted {
		t.Fatalf("fiscal status = %d body = %s", fiscalResponse.Code, fiscalResponse.Body.String())
	}

	refundResponse := httptest.NewRecorder()
	refundRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/payments/"+accepted.Payment.ID+"/refund", bytes.NewBufferString(`{
		"actorId": "cashier-1",
		"reason": "Customer return"
	}`))
	refundRequest.Header.Set("Content-Type", "application/json")
	refundRequest.Header.Set("Idempotency-Key", "payment-refund-refund-1")
	server.ServeHTTP(refundResponse, refundRequest)
	if refundResponse.Code != http.StatusAccepted {
		t.Fatalf("refund payment status = %d body = %s", refundResponse.Code, refundResponse.Body.String())
	}

	var refunded PaymentAcceptedResponse
	if err := json.Unmarshal(refundResponse.Body.Bytes(), &refunded); err != nil {
		t.Fatalf("decode refund payment response: %v", err)
	}
	if refunded.Payment.Status != "refunded" {
		t.Fatalf("payment status = %s", refunded.Payment.Status)
	}

	getReceiptResponse := httptest.NewRecorder()
	getReceiptRequest := httptest.NewRequest(http.MethodGet, "/v1/receipts/"+openedReceipt.Receipt.ID, nil)
	server.ServeHTTP(getReceiptResponse, getReceiptRequest)
	if getReceiptResponse.Code != http.StatusOK {
		t.Fatalf("get receipt status = %d body = %s", getReceiptResponse.Code, getReceiptResponse.Body.String())
	}

	var receipt ReceiptResponse
	if err := json.Unmarshal(getReceiptResponse.Body.Bytes(), &receipt); err != nil {
		t.Fatalf("decode receipt response: %v", err)
	}
	if receipt.Status != "fiscalized" {
		t.Fatalf("receipt status = %s", receipt.Status)
	}
}

func TestRefundReceiptCashPayment(t *testing.T) {
	server := NewServer()
	openStoreDayAndShiftForDate(t, server, "payment-refund-cash", time.Now().UTC().Format("2006-01-02"))

	openReceiptResponse := httptest.NewRecorder()
	openReceiptRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts", bytes.NewBufferString(`{
		"storeId": "store-1",
		"terminalId": "pos-1",
		"cashierId": "cashier-1",
		"channel": "pos"
	}`))
	openReceiptRequest.Header.Set("Content-Type", "application/json")
	openReceiptRequest.Header.Set("Idempotency-Key", "payment-refund-cash-receipt-open-1")
	server.ServeHTTP(openReceiptResponse, openReceiptRequest)
	if openReceiptResponse.Code != http.StatusAccepted {
		t.Fatalf("open receipt status = %d body = %s", openReceiptResponse.Code, openReceiptResponse.Body.String())
	}

	var openedReceipt ReceiptAcceptedResponse
	if err := json.Unmarshal(openReceiptResponse.Body.Bytes(), &openedReceipt); err != nil {
		t.Fatalf("decode open receipt response: %v", err)
	}

	addLineResponse := httptest.NewRecorder()
	addLineRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/lines", bytes.NewBufferString(`{
		"productId": "sku-1",
		"name": "Milk",
		"quantity": 1,
		"unitPriceMinor": 50000
	}`))
	addLineRequest.Header.Set("Content-Type", "application/json")
	addLineRequest.Header.Set("Idempotency-Key", "payment-refund-cash-line-1")
	server.ServeHTTP(addLineResponse, addLineRequest)
	if addLineResponse.Code != http.StatusAccepted {
		t.Fatalf("add line status = %d body = %s", addLineResponse.Code, addLineResponse.Body.String())
	}

	paymentResponse := httptest.NewRecorder()
	paymentRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/payments", bytes.NewBufferString(`{
		"method": "cash",
		"amountMinor": 50000
	}`))
	paymentRequest.Header.Set("Content-Type", "application/json")
	paymentRequest.Header.Set("Idempotency-Key", "payment-refund-cash-payment-1")
	server.ServeHTTP(paymentResponse, paymentRequest)
	if paymentResponse.Code != http.StatusAccepted {
		t.Fatalf("create payment status = %d body = %s", paymentResponse.Code, paymentResponse.Body.String())
	}

	var accepted PaymentAcceptedResponse
	if err := json.Unmarshal(paymentResponse.Body.Bytes(), &accepted); err != nil {
		t.Fatalf("decode payment response: %v", err)
	}

	fiscalResponse := httptest.NewRecorder()
	fiscalRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/fiscal-documents", bytes.NewBufferString(`{
		"deviceId": "mock-atol-1"
	}`))
	fiscalRequest.Header.Set("Content-Type", "application/json")
	fiscalRequest.Header.Set("Idempotency-Key", "payment-refund-cash-fiscal-1")
	server.ServeHTTP(fiscalResponse, fiscalRequest)
	if fiscalResponse.Code != http.StatusAccepted {
		t.Fatalf("fiscal status = %d body = %s", fiscalResponse.Code, fiscalResponse.Body.String())
	}

	refundResponse := httptest.NewRecorder()
	refundRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/payments/"+accepted.Payment.ID+"/refund", bytes.NewBufferString(`{
		"actorId": "cashier-1",
		"reason": "Customer return"
	}`))
	refundRequest.Header.Set("Content-Type", "application/json")
	refundRequest.Header.Set("Idempotency-Key", "payment-refund-cash-refund-1")
	server.ServeHTTP(refundResponse, refundRequest)
	if refundResponse.Code != http.StatusAccepted {
		t.Fatalf("refund payment status = %d body = %s", refundResponse.Code, refundResponse.Body.String())
	}

	var refunded PaymentAcceptedResponse
	if err := json.Unmarshal(refundResponse.Body.Bytes(), &refunded); err != nil {
		t.Fatalf("decode refund payment response: %v", err)
	}
	if refunded.Payment.Status != "refunded" {
		t.Fatalf("payment status = %s", refunded.Payment.Status)
	}

	getReceiptResponse := httptest.NewRecorder()
	getReceiptRequest := httptest.NewRequest(http.MethodGet, "/v1/receipts/"+openedReceipt.Receipt.ID, nil)
	server.ServeHTTP(getReceiptResponse, getReceiptRequest)
	if getReceiptResponse.Code != http.StatusOK {
		t.Fatalf("get receipt status = %d body = %s", getReceiptResponse.Code, getReceiptResponse.Body.String())
	}

	var receipt ReceiptResponse
	if err := json.Unmarshal(getReceiptResponse.Body.Bytes(), &receipt); err != nil {
		t.Fatalf("decode receipt response: %v", err)
	}
	if receipt.Status != "fiscalized" {
		t.Fatalf("receipt status = %s", receipt.Status)
	}
}

func TestSettleReturnWorkflow(t *testing.T) {
	server := NewServer()
	openStoreDayAndShiftForDate(t, server, "return-settle", time.Now().UTC().Format("2006-01-02"))

	openReceiptResponse := httptest.NewRecorder()
	openReceiptRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts", bytes.NewBufferString(`{
		"storeId": "store-1",
		"terminalId": "pos-1",
		"cashierId": "cashier-1",
		"channel": "pos"
	}`))
	openReceiptRequest.Header.Set("Content-Type", "application/json")
	openReceiptRequest.Header.Set("Idempotency-Key", "return-settle-receipt-open-1")
	server.ServeHTTP(openReceiptResponse, openReceiptRequest)
	if openReceiptResponse.Code != http.StatusAccepted {
		t.Fatalf("open receipt status = %d body = %s", openReceiptResponse.Code, openReceiptResponse.Body.String())
	}

	var openedReceipt ReceiptAcceptedResponse
	if err := json.Unmarshal(openReceiptResponse.Body.Bytes(), &openedReceipt); err != nil {
		t.Fatalf("decode open receipt response: %v", err)
	}

	addLineResponse := httptest.NewRecorder()
	addLineRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/lines", bytes.NewBufferString(`{
		"productId": "sku-1",
		"name": "Milk",
		"quantity": 1,
		"unitPriceMinor": 50000
	}`))
	addLineRequest.Header.Set("Content-Type", "application/json")
	addLineRequest.Header.Set("Idempotency-Key", "return-settle-line-1")
	server.ServeHTTP(addLineResponse, addLineRequest)
	if addLineResponse.Code != http.StatusAccepted {
		t.Fatalf("add line status = %d body = %s", addLineResponse.Code, addLineResponse.Body.String())
	}

	paymentResponse := httptest.NewRecorder()
	paymentRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/payments", bytes.NewBufferString(`{
		"method": "cash",
		"amountMinor": 50000
	}`))
	paymentRequest.Header.Set("Content-Type", "application/json")
	paymentRequest.Header.Set("Idempotency-Key", "return-settle-payment-1")
	server.ServeHTTP(paymentResponse, paymentRequest)
	if paymentResponse.Code != http.StatusAccepted {
		t.Fatalf("create payment status = %d body = %s", paymentResponse.Code, paymentResponse.Body.String())
	}

	fiscalResponse := httptest.NewRecorder()
	fiscalRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/fiscal-documents", bytes.NewBufferString(`{
		"deviceId": "mock-atol-1"
	}`))
	fiscalRequest.Header.Set("Content-Type", "application/json")
	fiscalRequest.Header.Set("Idempotency-Key", "return-settle-fiscal-1")
	server.ServeHTTP(fiscalResponse, fiscalRequest)
	if fiscalResponse.Code != http.StatusAccepted {
		t.Fatalf("fiscal status = %d body = %s", fiscalResponse.Code, fiscalResponse.Body.String())
	}

	getReceiptResponse := httptest.NewRecorder()
	getReceiptRequest := httptest.NewRequest(http.MethodGet, "/v1/receipts/"+openedReceipt.Receipt.ID, nil)
	server.ServeHTTP(getReceiptResponse, getReceiptRequest)
	if getReceiptResponse.Code != http.StatusOK {
		t.Fatalf("get receipt status = %d body = %s", getReceiptResponse.Code, getReceiptResponse.Body.String())
	}

	var receipt ReceiptResponse
	if err := json.Unmarshal(getReceiptResponse.Body.Bytes(), &receipt); err != nil {
		t.Fatalf("decode receipt response: %v", err)
	}
	if len(receipt.Lines) != 1 {
		t.Fatalf("receipt lines = %d", len(receipt.Lines))
	}

	returnResponseRecorder := httptest.NewRecorder()
	returnRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/returns", bytes.NewBufferString(fmt.Sprintf(`{
		"lines": [{"lineId": %q, "quantity": 1}],
		"reason": "Customer return",
		"actorId": "senior-1"
	}`, receipt.Lines[0].ID)))
	returnRequest.Header.Set("Content-Type", "application/json")
	returnRequest.Header.Set("Idempotency-Key", "return-settle-return-1")
	server.ServeHTTP(returnResponseRecorder, returnRequest)
	if returnResponseRecorder.Code != http.StatusAccepted {
		t.Fatalf("create return status = %d body = %s", returnResponseRecorder.Code, returnResponseRecorder.Body.String())
	}

	var acceptedReturn ReturnAcceptedResponse
	if err := json.Unmarshal(returnResponseRecorder.Body.Bytes(), &acceptedReturn); err != nil {
		t.Fatalf("decode return response: %v", err)
	}

	settleResponse := httptest.NewRecorder()
	settleRequest := httptest.NewRequest(http.MethodPost, "/v1/returns/"+acceptedReturn.Return.ID+"/settle", bytes.NewBufferString(`{
		"actorId": "senior-1",
		"reason": "Customer return"
	}`))
	settleRequest.Header.Set("Content-Type", "application/json")
	settleRequest.Header.Set("Idempotency-Key", "return-settle-settle-1")
	server.ServeHTTP(settleResponse, settleRequest)
	if settleResponse.Code != http.StatusAccepted {
		t.Fatalf("settle return status = %d body = %s", settleResponse.Code, settleResponse.Body.String())
	}

	var settled ReturnSettledResponse
	if err := json.Unmarshal(settleResponse.Body.Bytes(), &settled); err != nil {
		t.Fatalf("decode settle response: %v", err)
	}
	if settled.Return.Status != "settled" {
		t.Fatalf("return status = %s", settled.Return.Status)
	}
	if len(settled.Payments) != 1 || settled.Payments[0].Status != "refunded" {
		t.Fatalf("payments = %+v", settled.Payments)
	}
}

func TestGetReturnAndListReceiptReturnsWorkflow(t *testing.T) {
	server := NewServer()
	openStoreDayAndShiftForDate(t, server, "return-read", time.Now().UTC().Format("2006-01-02"))

	openReceiptResponse := httptest.NewRecorder()
	openReceiptRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts", bytes.NewBufferString(`{
		"storeId": "store-1",
		"terminalId": "pos-1",
		"cashierId": "cashier-1",
		"channel": "pos"
	}`))
	openReceiptRequest.Header.Set("Content-Type", "application/json")
	openReceiptRequest.Header.Set("Idempotency-Key", "return-read-receipt-open-1")
	server.ServeHTTP(openReceiptResponse, openReceiptRequest)
	if openReceiptResponse.Code != http.StatusAccepted {
		t.Fatalf("open receipt status = %d body = %s", openReceiptResponse.Code, openReceiptResponse.Body.String())
	}

	var openedReceipt ReceiptAcceptedResponse
	if err := json.Unmarshal(openReceiptResponse.Body.Bytes(), &openedReceipt); err != nil {
		t.Fatalf("decode open receipt response: %v", err)
	}

	addLineResponse := httptest.NewRecorder()
	addLineRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/lines", bytes.NewBufferString(`{
		"productId": "sku-1",
		"name": "Milk",
		"quantity": 1,
		"unitPriceMinor": 50000
	}`))
	addLineRequest.Header.Set("Content-Type", "application/json")
	addLineRequest.Header.Set("Idempotency-Key", "return-read-line-1")
	server.ServeHTTP(addLineResponse, addLineRequest)
	if addLineResponse.Code != http.StatusAccepted {
		t.Fatalf("add line status = %d body = %s", addLineResponse.Code, addLineResponse.Body.String())
	}

	paymentResponse := httptest.NewRecorder()
	paymentRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/payments", bytes.NewBufferString(`{
		"method": "cash",
		"amountMinor": 50000
	}`))
	paymentRequest.Header.Set("Content-Type", "application/json")
	paymentRequest.Header.Set("Idempotency-Key", "return-read-payment-1")
	server.ServeHTTP(paymentResponse, paymentRequest)
	if paymentResponse.Code != http.StatusAccepted {
		t.Fatalf("create payment status = %d body = %s", paymentResponse.Code, paymentResponse.Body.String())
	}

	fiscalResponse := httptest.NewRecorder()
	fiscalRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/fiscal-documents", bytes.NewBufferString(`{
		"deviceId": "mock-atol-1"
	}`))
	fiscalRequest.Header.Set("Content-Type", "application/json")
	fiscalRequest.Header.Set("Idempotency-Key", "return-read-fiscal-1")
	server.ServeHTTP(fiscalResponse, fiscalRequest)
	if fiscalResponse.Code != http.StatusAccepted {
		t.Fatalf("fiscal status = %d body = %s", fiscalResponse.Code, fiscalResponse.Body.String())
	}

	getReceiptResponse := httptest.NewRecorder()
	getReceiptRequest := httptest.NewRequest(http.MethodGet, "/v1/receipts/"+openedReceipt.Receipt.ID, nil)
	server.ServeHTTP(getReceiptResponse, getReceiptRequest)
	if getReceiptResponse.Code != http.StatusOK {
		t.Fatalf("get receipt status = %d body = %s", getReceiptResponse.Code, getReceiptResponse.Body.String())
	}

	var receipt ReceiptResponse
	if err := json.Unmarshal(getReceiptResponse.Body.Bytes(), &receipt); err != nil {
		t.Fatalf("decode receipt response: %v", err)
	}

	createReturnResponse := httptest.NewRecorder()
	createReturnRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/returns", bytes.NewBufferString(fmt.Sprintf(`{
		"lines": [{"lineId": %q, "quantity": 1}],
		"reason": "Customer return",
		"actorId": "senior-1"
	}`, receipt.Lines[0].ID)))
	createReturnRequest.Header.Set("Content-Type", "application/json")
	createReturnRequest.Header.Set("Idempotency-Key", "return-read-create-1")
	server.ServeHTTP(createReturnResponse, createReturnRequest)
	if createReturnResponse.Code != http.StatusAccepted {
		t.Fatalf("create return status = %d body = %s", createReturnResponse.Code, createReturnResponse.Body.String())
	}

	var createdReturn ReturnAcceptedResponse
	if err := json.Unmarshal(createReturnResponse.Body.Bytes(), &createdReturn); err != nil {
		t.Fatalf("decode create return response: %v", err)
	}

	getReturnResponse := httptest.NewRecorder()
	getReturnRequest := httptest.NewRequest(http.MethodGet, "/v1/returns/"+createdReturn.Return.ID, nil)
	server.ServeHTTP(getReturnResponse, getReturnRequest)
	if getReturnResponse.Code != http.StatusOK {
		t.Fatalf("get return status = %d body = %s", getReturnResponse.Code, getReturnResponse.Body.String())
	}

	var fetchedReturn ReturnAcceptedResponse
	if err := json.Unmarshal(getReturnResponse.Body.Bytes(), &fetchedReturn); err != nil {
		t.Fatalf("decode get return response: %v", err)
	}
	if fetchedReturn.Return.ID != createdReturn.Return.ID || fetchedReturn.Return.Status != "completed" {
		t.Fatalf("fetched return = %+v", fetchedReturn.Return)
	}

	listReturnsResponse := httptest.NewRecorder()
	listReturnsRequest := httptest.NewRequest(http.MethodGet, "/v1/receipts/"+openedReceipt.Receipt.ID+"/returns", nil)
	server.ServeHTTP(listReturnsResponse, listReturnsRequest)
	if listReturnsResponse.Code != http.StatusOK {
		t.Fatalf("list returns status = %d body = %s", listReturnsResponse.Code, listReturnsResponse.Body.String())
	}

	var listed PaginatedReturnsResponse
	if err := json.Unmarshal(listReturnsResponse.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode list returns response: %v", err)
	}
	if listed.TotalCount != 1 || len(listed.Items) != 1 || listed.Items[0].ID != createdReturn.Return.ID {
		t.Fatalf("listed returns = %+v", listed)
	}

	listStoreReturnsResponse := httptest.NewRecorder()
	listStoreReturnsRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/returns", nil)
	server.ServeHTTP(listStoreReturnsResponse, listStoreReturnsRequest)
	if listStoreReturnsResponse.Code != http.StatusOK {
		t.Fatalf("list store returns status = %d body = %s", listStoreReturnsResponse.Code, listStoreReturnsResponse.Body.String())
	}

	var storeListed PaginatedReturnsResponse
	if err := json.Unmarshal(listStoreReturnsResponse.Body.Bytes(), &storeListed); err != nil {
		t.Fatalf("decode list store returns response: %v", err)
	}
	if storeListed.TotalCount != 1 || len(storeListed.Items) != 1 || storeListed.Items[0].ID != createdReturn.Return.ID {
		t.Fatalf("store listed returns = %+v", storeListed)
	}
}

func TestCreateReturnFiscalDocumentWorkflow(t *testing.T) {
	server := NewServer()
	openStoreDayAndShiftForDate(t, server, "return-fiscal", time.Now().UTC().Format("2006-01-02"))

	openReceiptResponse := httptest.NewRecorder()
	openReceiptRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts", bytes.NewBufferString(`{
		"storeId": "store-1",
		"terminalId": "pos-1",
		"cashierId": "cashier-1",
		"channel": "pos"
	}`))
	openReceiptRequest.Header.Set("Content-Type", "application/json")
	openReceiptRequest.Header.Set("Idempotency-Key", "return-fiscal-receipt-open-1")
	server.ServeHTTP(openReceiptResponse, openReceiptRequest)
	if openReceiptResponse.Code != http.StatusAccepted {
		t.Fatalf("open receipt status = %d body = %s", openReceiptResponse.Code, openReceiptResponse.Body.String())
	}

	var openedReceipt ReceiptAcceptedResponse
	if err := json.Unmarshal(openReceiptResponse.Body.Bytes(), &openedReceipt); err != nil {
		t.Fatalf("decode open receipt response: %v", err)
	}

	addLineResponse := httptest.NewRecorder()
	addLineRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/lines", bytes.NewBufferString(`{
		"productId": "sku-1",
		"name": "Milk",
		"quantity": 1,
		"unitPriceMinor": 50000
	}`))
	addLineRequest.Header.Set("Content-Type", "application/json")
	addLineRequest.Header.Set("Idempotency-Key", "return-fiscal-line-1")
	server.ServeHTTP(addLineResponse, addLineRequest)
	if addLineResponse.Code != http.StatusAccepted {
		t.Fatalf("add line status = %d body = %s", addLineResponse.Code, addLineResponse.Body.String())
	}

	paymentResponse := httptest.NewRecorder()
	paymentRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/payments", bytes.NewBufferString(`{
		"method": "cash",
		"amountMinor": 50000
	}`))
	paymentRequest.Header.Set("Content-Type", "application/json")
	paymentRequest.Header.Set("Idempotency-Key", "return-fiscal-payment-1")
	server.ServeHTTP(paymentResponse, paymentRequest)
	if paymentResponse.Code != http.StatusAccepted {
		t.Fatalf("create payment status = %d body = %s", paymentResponse.Code, paymentResponse.Body.String())
	}

	fiscalResponse := httptest.NewRecorder()
	fiscalRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/fiscal-documents", bytes.NewBufferString(`{
		"deviceId": "mock-atol-1"
	}`))
	fiscalRequest.Header.Set("Content-Type", "application/json")
	fiscalRequest.Header.Set("Idempotency-Key", "return-fiscal-fiscal-1")
	server.ServeHTTP(fiscalResponse, fiscalRequest)
	if fiscalResponse.Code != http.StatusAccepted {
		t.Fatalf("fiscal status = %d body = %s", fiscalResponse.Code, fiscalResponse.Body.String())
	}

	getReceiptResponse := httptest.NewRecorder()
	getReceiptRequest := httptest.NewRequest(http.MethodGet, "/v1/receipts/"+openedReceipt.Receipt.ID, nil)
	server.ServeHTTP(getReceiptResponse, getReceiptRequest)
	if getReceiptResponse.Code != http.StatusOK {
		t.Fatalf("get receipt status = %d body = %s", getReceiptResponse.Code, getReceiptResponse.Body.String())
	}

	var receipt ReceiptResponse
	if err := json.Unmarshal(getReceiptResponse.Body.Bytes(), &receipt); err != nil {
		t.Fatalf("decode receipt response: %v", err)
	}

	returnResponseRecorder := httptest.NewRecorder()
	returnRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/returns", bytes.NewBufferString(fmt.Sprintf(`{
		"lines": [{"lineId": %q, "quantity": 1}],
		"reason": "Customer return",
		"actorId": "senior-1"
	}`, receipt.Lines[0].ID)))
	returnRequest.Header.Set("Content-Type", "application/json")
	returnRequest.Header.Set("Idempotency-Key", "return-fiscal-return-1")
	server.ServeHTTP(returnResponseRecorder, returnRequest)
	if returnResponseRecorder.Code != http.StatusAccepted {
		t.Fatalf("create return status = %d body = %s", returnResponseRecorder.Code, returnResponseRecorder.Body.String())
	}

	var acceptedReturn ReturnAcceptedResponse
	if err := json.Unmarshal(returnResponseRecorder.Body.Bytes(), &acceptedReturn); err != nil {
		t.Fatalf("decode return response: %v", err)
	}

	settleResponse := httptest.NewRecorder()
	settleRequest := httptest.NewRequest(http.MethodPost, "/v1/returns/"+acceptedReturn.Return.ID+"/settle", bytes.NewBufferString(`{
		"actorId": "senior-1",
		"reason": "Customer return"
	}`))
	settleRequest.Header.Set("Content-Type", "application/json")
	settleRequest.Header.Set("Idempotency-Key", "return-fiscal-settle-1")
	server.ServeHTTP(settleResponse, settleRequest)
	if settleResponse.Code != http.StatusAccepted {
		t.Fatalf("settle return status = %d body = %s", settleResponse.Code, settleResponse.Body.String())
	}

	returnFiscalResponse := httptest.NewRecorder()
	returnFiscalRequest := httptest.NewRequest(http.MethodPost, "/v1/returns/"+acceptedReturn.Return.ID+"/fiscal-documents", bytes.NewBufferString(`{
		"deviceId": "mock-atol-1"
	}`))
	returnFiscalRequest.Header.Set("Content-Type", "application/json")
	returnFiscalRequest.Header.Set("Idempotency-Key", "return-fiscal-doc-1")
	server.ServeHTTP(returnFiscalResponse, returnFiscalRequest)
	if returnFiscalResponse.Code != http.StatusAccepted {
		t.Fatalf("create return fiscal status = %d body = %s", returnFiscalResponse.Code, returnFiscalResponse.Body.String())
	}

	var returnFiscal FiscalDocumentAcceptedResponse
	if err := json.Unmarshal(returnFiscalResponse.Body.Bytes(), &returnFiscal); err != nil {
		t.Fatalf("decode return fiscal response: %v", err)
	}
	if returnFiscal.Document.Kind != "return" {
		t.Fatalf("document kind = %s", returnFiscal.Document.Kind)
	}
	if returnFiscal.Document.ReturnID != acceptedReturn.Return.ID {
		t.Fatalf("return id = %s", returnFiscal.Document.ReturnID)
	}
	if returnFiscal.Document.AmountMinor != acceptedReturn.Return.TotalMinor {
		t.Fatalf("amount = %d", returnFiscal.Document.AmountMinor)
	}

	returnFiscalListResponse := httptest.NewRecorder()
	returnFiscalListRequest := httptest.NewRequest(http.MethodGet, "/v1/returns/"+acceptedReturn.Return.ID+"/fiscal-documents", nil)
	server.ServeHTTP(returnFiscalListResponse, returnFiscalListRequest)
	if returnFiscalListResponse.Code != http.StatusOK {
		t.Fatalf("list return fiscal status = %d body = %s", returnFiscalListResponse.Code, returnFiscalListResponse.Body.String())
	}

	var returnFiscalDocuments FiscalDocumentsResponse
	if err := json.Unmarshal(returnFiscalListResponse.Body.Bytes(), &returnFiscalDocuments); err != nil {
		t.Fatalf("decode return fiscal list response: %v", err)
	}
	if len(returnFiscalDocuments.Documents) != 1 {
		t.Fatalf("return fiscal documents count = %d", len(returnFiscalDocuments.Documents))
	}
}

func TestSettlePartialReturnWorkflow(t *testing.T) {
	server := NewServer()
	openStoreDayAndShiftForDate(t, server, "return-settle-partial", time.Now().UTC().Format("2006-01-02"))

	openReceiptResponse := httptest.NewRecorder()
	openReceiptRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts", bytes.NewBufferString(`{
		"storeId": "store-1",
		"terminalId": "pos-1",
		"cashierId": "cashier-1",
		"channel": "pos"
	}`))
	openReceiptRequest.Header.Set("Content-Type", "application/json")
	openReceiptRequest.Header.Set("Idempotency-Key", "return-settle-partial-receipt-open-1")
	server.ServeHTTP(openReceiptResponse, openReceiptRequest)
	if openReceiptResponse.Code != http.StatusAccepted {
		t.Fatalf("open receipt status = %d body = %s", openReceiptResponse.Code, openReceiptResponse.Body.String())
	}

	var openedReceipt ReceiptAcceptedResponse
	if err := json.Unmarshal(openReceiptResponse.Body.Bytes(), &openedReceipt); err != nil {
		t.Fatalf("decode open receipt response: %v", err)
	}

	addLineResponse := httptest.NewRecorder()
	addLineRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/lines", bytes.NewBufferString(`{
		"productId": "sku-1",
		"name": "Milk",
		"quantity": 2,
		"unitPriceMinor": 50000
	}`))
	addLineRequest.Header.Set("Content-Type", "application/json")
	addLineRequest.Header.Set("Idempotency-Key", "return-settle-partial-line-1")
	server.ServeHTTP(addLineResponse, addLineRequest)
	if addLineResponse.Code != http.StatusAccepted {
		t.Fatalf("add line status = %d body = %s", addLineResponse.Code, addLineResponse.Body.String())
	}

	paymentResponse := httptest.NewRecorder()
	paymentRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/payments", bytes.NewBufferString(`{
		"method": "cash",
		"amountMinor": 100000
	}`))
	paymentRequest.Header.Set("Content-Type", "application/json")
	paymentRequest.Header.Set("Idempotency-Key", "return-settle-partial-payment-1")
	server.ServeHTTP(paymentResponse, paymentRequest)
	if paymentResponse.Code != http.StatusAccepted {
		t.Fatalf("create payment status = %d body = %s", paymentResponse.Code, paymentResponse.Body.String())
	}

	fiscalResponse := httptest.NewRecorder()
	fiscalRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/fiscal-documents", bytes.NewBufferString(`{
		"deviceId": "mock-atol-1"
	}`))
	fiscalRequest.Header.Set("Content-Type", "application/json")
	fiscalRequest.Header.Set("Idempotency-Key", "return-settle-partial-fiscal-1")
	server.ServeHTTP(fiscalResponse, fiscalRequest)
	if fiscalResponse.Code != http.StatusAccepted {
		t.Fatalf("fiscal status = %d body = %s", fiscalResponse.Code, fiscalResponse.Body.String())
	}

	getReceiptResponse := httptest.NewRecorder()
	getReceiptRequest := httptest.NewRequest(http.MethodGet, "/v1/receipts/"+openedReceipt.Receipt.ID, nil)
	server.ServeHTTP(getReceiptResponse, getReceiptRequest)
	if getReceiptResponse.Code != http.StatusOK {
		t.Fatalf("get receipt status = %d body = %s", getReceiptResponse.Code, getReceiptResponse.Body.String())
	}

	var receipt ReceiptResponse
	if err := json.Unmarshal(getReceiptResponse.Body.Bytes(), &receipt); err != nil {
		t.Fatalf("decode receipt response: %v", err)
	}

	returnResponseRecorder := httptest.NewRecorder()
	returnRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/returns", bytes.NewBufferString(fmt.Sprintf(`{
		"lines": [{"lineId": %q, "quantity": 1}],
		"reason": "Partial customer return",
		"actorId": "senior-1"
	}`, receipt.Lines[0].ID)))
	returnRequest.Header.Set("Content-Type", "application/json")
	returnRequest.Header.Set("Idempotency-Key", "return-settle-partial-return-1")
	server.ServeHTTP(returnResponseRecorder, returnRequest)
	if returnResponseRecorder.Code != http.StatusAccepted {
		t.Fatalf("create return status = %d body = %s", returnResponseRecorder.Code, returnResponseRecorder.Body.String())
	}

	var acceptedReturn ReturnAcceptedResponse
	if err := json.Unmarshal(returnResponseRecorder.Body.Bytes(), &acceptedReturn); err != nil {
		t.Fatalf("decode return response: %v", err)
	}

	settleResponse := httptest.NewRecorder()
	settleRequest := httptest.NewRequest(http.MethodPost, "/v1/returns/"+acceptedReturn.Return.ID+"/settle", bytes.NewBufferString(`{
		"actorId": "senior-1",
		"reason": "Partial customer return"
	}`))
	settleRequest.Header.Set("Content-Type", "application/json")
	settleRequest.Header.Set("Idempotency-Key", "return-settle-partial-settle-1")
	server.ServeHTTP(settleResponse, settleRequest)
	if settleResponse.Code != http.StatusAccepted {
		t.Fatalf("settle return status = %d body = %s", settleResponse.Code, settleResponse.Body.String())
	}

	var settled ReturnSettledResponse
	if err := json.Unmarshal(settleResponse.Body.Bytes(), &settled); err != nil {
		t.Fatalf("decode settle response: %v", err)
	}
	if settled.Return.Status != "settled" {
		t.Fatalf("return status = %s", settled.Return.Status)
	}
	if len(settled.Payments) != 1 || settled.Payments[0].Status != "partially_refunded" {
		t.Fatalf("payments = %+v", settled.Payments)
	}
	if settled.Payments[0].RefundedAmountMinor != 50000 {
		t.Fatalf("refunded amount = %d", settled.Payments[0].RefundedAmountMinor)
	}
}

func TestCreateReceiptReturnBlocksCumulativeLineQuantityWorkflow(t *testing.T) {
	server := NewServer()
	openStoreDayAndShiftForDate(t, server, "return-cumulative-line", time.Now().UTC().Format("2006-01-02"))

	openReceiptResponse := httptest.NewRecorder()
	openReceiptRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts", bytes.NewBufferString(`{
		"storeId": "store-1",
		"terminalId": "pos-1",
		"cashierId": "cashier-1",
		"channel": "pos"
	}`))
	openReceiptRequest.Header.Set("Content-Type", "application/json")
	openReceiptRequest.Header.Set("Idempotency-Key", "return-cumulative-line-receipt-open-1")
	server.ServeHTTP(openReceiptResponse, openReceiptRequest)
	if openReceiptResponse.Code != http.StatusAccepted {
		t.Fatalf("open receipt status = %d body = %s", openReceiptResponse.Code, openReceiptResponse.Body.String())
	}

	var openedReceipt ReceiptAcceptedResponse
	if err := json.Unmarshal(openReceiptResponse.Body.Bytes(), &openedReceipt); err != nil {
		t.Fatalf("decode open receipt response: %v", err)
	}

	addLineResponse := httptest.NewRecorder()
	addLineRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/lines", bytes.NewBufferString(`{
		"productId": "sku-1",
		"name": "Milk",
		"quantity": 2,
		"unitPriceMinor": 50000
	}`))
	addLineRequest.Header.Set("Content-Type", "application/json")
	addLineRequest.Header.Set("Idempotency-Key", "return-cumulative-line-line-1")
	server.ServeHTTP(addLineResponse, addLineRequest)
	if addLineResponse.Code != http.StatusAccepted {
		t.Fatalf("add line status = %d body = %s", addLineResponse.Code, addLineResponse.Body.String())
	}

	paymentResponse := httptest.NewRecorder()
	paymentRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/payments", bytes.NewBufferString(`{
		"method": "cash",
		"amountMinor": 100000
	}`))
	paymentRequest.Header.Set("Content-Type", "application/json")
	paymentRequest.Header.Set("Idempotency-Key", "return-cumulative-line-payment-1")
	server.ServeHTTP(paymentResponse, paymentRequest)
	if paymentResponse.Code != http.StatusAccepted {
		t.Fatalf("create payment status = %d body = %s", paymentResponse.Code, paymentResponse.Body.String())
	}

	fiscalResponse := httptest.NewRecorder()
	fiscalRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/fiscal-documents", bytes.NewBufferString(`{
		"deviceId": "mock-atol-1"
	}`))
	fiscalRequest.Header.Set("Content-Type", "application/json")
	fiscalRequest.Header.Set("Idempotency-Key", "return-cumulative-line-fiscal-1")
	server.ServeHTTP(fiscalResponse, fiscalRequest)
	if fiscalResponse.Code != http.StatusAccepted {
		t.Fatalf("fiscal status = %d body = %s", fiscalResponse.Code, fiscalResponse.Body.String())
	}

	getReceiptResponse := httptest.NewRecorder()
	getReceiptRequest := httptest.NewRequest(http.MethodGet, "/v1/receipts/"+openedReceipt.Receipt.ID, nil)
	server.ServeHTTP(getReceiptResponse, getReceiptRequest)
	if getReceiptResponse.Code != http.StatusOK {
		t.Fatalf("get receipt status = %d body = %s", getReceiptResponse.Code, getReceiptResponse.Body.String())
	}

	var receipt ReceiptResponse
	if err := json.Unmarshal(getReceiptResponse.Body.Bytes(), &receipt); err != nil {
		t.Fatalf("decode receipt response: %v", err)
	}

	firstReturnResponse := httptest.NewRecorder()
	firstReturnRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/returns", bytes.NewBufferString(fmt.Sprintf(`{
		"lines": [{"lineId": %q, "quantity": 1}],
		"reason": "First partial return",
		"actorId": "senior-1"
	}`, receipt.Lines[0].ID)))
	firstReturnRequest.Header.Set("Content-Type", "application/json")
	firstReturnRequest.Header.Set("Idempotency-Key", "return-cumulative-line-return-1")
	server.ServeHTTP(firstReturnResponse, firstReturnRequest)
	if firstReturnResponse.Code != http.StatusAccepted {
		t.Fatalf("create first return status = %d body = %s", firstReturnResponse.Code, firstReturnResponse.Body.String())
	}

	var firstReturn ReturnAcceptedResponse
	if err := json.Unmarshal(firstReturnResponse.Body.Bytes(), &firstReturn); err != nil {
		t.Fatalf("decode first return response: %v", err)
	}

	settleResponse := httptest.NewRecorder()
	settleRequest := httptest.NewRequest(http.MethodPost, "/v1/returns/"+firstReturn.Return.ID+"/settle", bytes.NewBufferString(`{
		"actorId": "senior-1",
		"reason": "First partial return"
	}`))
	settleRequest.Header.Set("Content-Type", "application/json")
	settleRequest.Header.Set("Idempotency-Key", "return-cumulative-line-settle-1")
	server.ServeHTTP(settleResponse, settleRequest)
	if settleResponse.Code != http.StatusAccepted {
		t.Fatalf("settle return status = %d body = %s", settleResponse.Code, settleResponse.Body.String())
	}

	secondReturnResponse := httptest.NewRecorder()
	secondReturnRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/returns", bytes.NewBufferString(fmt.Sprintf(`{
		"lines": [{"lineId": %q, "quantity": 2}],
		"reason": "Second return exceeds line cap",
		"actorId": "senior-1"
	}`, receipt.Lines[0].ID)))
	secondReturnRequest.Header.Set("Content-Type", "application/json")
	secondReturnRequest.Header.Set("Idempotency-Key", "return-cumulative-line-return-2")
	server.ServeHTTP(secondReturnResponse, secondReturnRequest)
	if secondReturnResponse.Code != http.StatusBadRequest {
		t.Fatalf("create second return status = %d body = %s", secondReturnResponse.Code, secondReturnResponse.Body.String())
	}
}

func TestSettleNoReceiptReturnWorkflow(t *testing.T) {
	server := NewServer()
	openStoreDayAndShiftForDate(t, server, "return-settle-no-receipt", time.Now().UTC().Format("2006-01-02"))

	fundResponse := httptest.NewRecorder()
	fundRequest := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/cash-movements", bytes.NewBufferString(`{
		"type": "change_fund",
		"fromContainerId": "safe-1",
		"fromContainerType": "safe",
		"toContainerId": "drawer-1",
		"toContainerType": "drawer",
		"amountMinor": 500000,
		"currency": "RUB",
		"reason": "Opening change fund",
		"actorId": "senior-1",
		"approvedById": "cashier-1"
	}`))
	fundRequest.Header.Set("Content-Type", "application/json")
	fundRequest.Header.Set("Idempotency-Key", "return-settle-no-receipt-fund-1")
	server.ServeHTTP(fundResponse, fundRequest)
	if fundResponse.Code != http.StatusAccepted {
		t.Fatalf("fund drawer status = %d body = %s", fundResponse.Code, fundResponse.Body.String())
	}

	returnResponseRecorder := httptest.NewRecorder()
	returnRequest := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/returns/no-receipt", bytes.NewBufferString(`{
		"lines": [{"productId": "sku-1", "name": "Milk", "quantity": 1, "unitPriceMinor": 2500}],
		"reason": "No receipt return",
		"actorId": "senior-1",
		"approvedById": "admin-1"
	}`))
	returnRequest.Header.Set("Content-Type", "application/json")
	returnRequest.Header.Set("Idempotency-Key", "return-settle-no-receipt-return-1")
	server.ServeHTTP(returnResponseRecorder, returnRequest)
	if returnResponseRecorder.Code != http.StatusAccepted {
		t.Fatalf("create return status = %d body = %s", returnResponseRecorder.Code, returnResponseRecorder.Body.String())
	}

	var acceptedReturn ReturnAcceptedResponse
	if err := json.Unmarshal(returnResponseRecorder.Body.Bytes(), &acceptedReturn); err != nil {
		t.Fatalf("decode return response: %v", err)
	}

	settleResponse := httptest.NewRecorder()
	settleRequest := httptest.NewRequest(http.MethodPost, "/v1/returns/"+acceptedReturn.Return.ID+"/settle", bytes.NewBufferString(`{
		"actorId": "senior-1",
		"reason": "No receipt payout",
		"drawerId": "drawer-1"
	}`))
	settleRequest.Header.Set("Content-Type", "application/json")
	settleRequest.Header.Set("Idempotency-Key", "return-settle-no-receipt-settle-1")
	server.ServeHTTP(settleResponse, settleRequest)
	if settleResponse.Code != http.StatusAccepted {
		t.Fatalf("settle return status = %d body = %s", settleResponse.Code, settleResponse.Body.String())
	}

	var settled ReturnSettledResponse
	if err := json.Unmarshal(settleResponse.Body.Bytes(), &settled); err != nil {
		t.Fatalf("decode settle response: %v", err)
	}
	if settled.Return.Status != "settled" {
		t.Fatalf("return status = %s", settled.Return.Status)
	}
	if len(settled.Payments) != 0 {
		t.Fatalf("payments = %+v", settled.Payments)
	}
}

func TestCashMovementWorkflow(t *testing.T) {
	server := NewServer()

	createResponse := httptest.NewRecorder()
	createRequest := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/cash-movements", bytes.NewBufferString(`{
		"type": "change_fund",
		"fromContainerId": "safe-1",
		"fromContainerType": "safe",
		"toContainerId": "drawer-1",
		"toContainerType": "drawer",
		"amountMinor": 500000,
		"currency": "RUB",
		"reason": "Opening change fund",
		"actorId": "senior-1",
		"approvedById": "cashier-1"
	}`))
	createRequest.Header.Set("Idempotency-Key", "cash-1")

	server.ServeHTTP(createResponse, createRequest)

	if createResponse.Code != http.StatusAccepted {
		t.Fatalf("create cash movement status = %d, body = %s", createResponse.Code, createResponse.Body.String())
	}

	var accepted CashMovementAcceptedResponse
	if err := json.Unmarshal(createResponse.Body.Bytes(), &accepted); err != nil {
		t.Fatalf("decode cash movement response: %v", err)
	}
	if accepted.Movement.Status != "posted" || accepted.Movement.AmountMinor != 500000 {
		t.Fatalf("cash movement = %+v", accepted.Movement)
	}

	listResponse := httptest.NewRecorder()
	listRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/cash-movements", nil)

	server.ServeHTTP(listResponse, listRequest)

	if listResponse.Code != http.StatusOK {
		t.Fatalf("list cash movements status = %d, body = %s", listResponse.Code, listResponse.Body.String())
	}

	var listed PaginatedCashMovementsResponse
	if err := json.Unmarshal(listResponse.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode cash movements response: %v", err)
	}
	if len(listed.Items) != 1 {
		t.Fatalf("cash movements count = %d", len(listed.Items))
	}

	balancesResponse := httptest.NewRecorder()
	balancesRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/cash-balances", nil)

	server.ServeHTTP(balancesResponse, balancesRequest)

	if balancesResponse.Code != http.StatusOK {
		t.Fatalf("list cash balances status = %d, body = %s", balancesResponse.Code, balancesResponse.Body.String())
	}

	var balances CashBalancesResponse
	if err := json.Unmarshal(balancesResponse.Body.Bytes(), &balances); err != nil {
		t.Fatalf("decode cash balances response: %v", err)
	}
	if len(balances.Balances) != 2 {
		t.Fatalf("cash balances count = %d", len(balances.Balances))
	}

	recountResponse := httptest.NewRecorder()
	recountRequest := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/cash-recounts", bytes.NewBufferString(`{
		"containerId": "drawer-1",
		"containerType": "drawer",
		"currency": "RUB",
		"countedMinor": 500000,
		"reason": "Drawer recount",
		"actorId": "senior-1"
	}`))
	recountRequest.Header.Set("Idempotency-Key", "recount-1")

	server.ServeHTTP(recountResponse, recountRequest)

	if recountResponse.Code != http.StatusAccepted {
		t.Fatalf("create cash recount status = %d, body = %s", recountResponse.Code, recountResponse.Body.String())
	}

	var recount CashRecountAcceptedResponse
	if err := json.Unmarshal(recountResponse.Body.Bytes(), &recount); err != nil {
		t.Fatalf("decode cash recount response: %v", err)
	}
	if recount.Recount.Status != "balanced" || recount.Recount.ExpectedMinor != 500000 {
		t.Fatalf("cash recount = %+v", recount.Recount)
	}

	recountsResponse := httptest.NewRecorder()
	recountsRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/cash-recounts", nil)

	server.ServeHTTP(recountsResponse, recountsRequest)

	if recountsResponse.Code != http.StatusOK {
		t.Fatalf("list cash recounts status = %d, body = %s", recountsResponse.Code, recountsResponse.Body.String())
	}

	var recounts PaginatedCashRecountsResponse
	if err := json.Unmarshal(recountsResponse.Body.Bytes(), &recounts); err != nil {
		t.Fatalf("decode cash recounts response: %v", err)
	}
	if len(recounts.Items) != 1 {
		t.Fatalf("cash recounts count = %d", len(recounts.Items))
	}
}

func TestCashMovementRejectsSelfApproval(t *testing.T) {
	server := NewServer()

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/cash-movements", bytes.NewBufferString(`{
		"type": "change_fund",
		"fromContainerId": "safe-1",
		"fromContainerType": "safe",
		"toContainerId": "drawer-1",
		"toContainerType": "drawer",
		"amountMinor": 500000,
		"actorId": "senior-1",
		"approvedById": "senior-1"
	}`))
	request.Header.Set("Idempotency-Key", "cash-1")

	server.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("self approval status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestBankCollectionAndBusinessExpenseWorkflow(t *testing.T) {
	server := NewServer()

	seedResponse := httptest.NewRecorder()
	seedRequest := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/cash-movements", bytes.NewBufferString(`{
		"type": "cash_in",
		"fromContainerId": "external-customer",
		"fromContainerType": "external",
		"toContainerId": "safe-1",
		"toContainerType": "safe",
		"amountMinor": 500000,
		"reason": "Seed safe balance",
		"actorId": "senior-1"
	}`))
	seedRequest.Header.Set("Idempotency-Key", "seed-safe-1")
	server.ServeHTTP(seedResponse, seedRequest)
	if seedResponse.Code != http.StatusAccepted {
		t.Fatalf("seed safe status = %d, body = %s", seedResponse.Code, seedResponse.Body.String())
	}

	bankResponse := httptest.NewRecorder()
	bankRequest := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/bank-collections", bytes.NewBufferString(`{
		"safeId": "safe-1",
		"bankContainerId": "bank-collection-1",
		"amountMinor": 200000,
		"reason": "Scheduled collection",
		"actorId": "senior-1",
		"approvedById": "admin-1"
	}`))
	bankRequest.Header.Set("Idempotency-Key", "bank-collection-1")
	server.ServeHTTP(bankResponse, bankRequest)
	if bankResponse.Code != http.StatusAccepted {
		t.Fatalf("bank collection status = %d, body = %s", bankResponse.Code, bankResponse.Body.String())
	}

	expenseResponse := httptest.NewRecorder()
	expenseRequest := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/business-expenses", bytes.NewBufferString(`{
		"safeId": "safe-1",
		"payeeId": "vendor-supplies",
		"amountMinor": 50000,
		"reason": "Office supplies",
		"actorId": "senior-1",
		"approvedById": "admin-1"
	}`))
	expenseRequest.Header.Set("Idempotency-Key", "expense-1")
	server.ServeHTTP(expenseResponse, expenseRequest)
	if expenseResponse.Code != http.StatusAccepted {
		t.Fatalf("business expense status = %d, body = %s", expenseResponse.Code, expenseResponse.Body.String())
	}

	balancesResponse := httptest.NewRecorder()
	balancesRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/cash-balances", nil)
	server.ServeHTTP(balancesResponse, balancesRequest)
	if balancesResponse.Code != http.StatusOK {
		t.Fatalf("cash balances status = %d, body = %s", balancesResponse.Code, balancesResponse.Body.String())
	}

	var balances CashBalancesResponse
	if err := json.Unmarshal(balancesResponse.Body.Bytes(), &balances); err != nil {
		t.Fatalf("decode cash balances response: %v", err)
	}
	byContainer := map[string]int64{}
	for _, balance := range balances.Balances {
		byContainer[balance.ContainerID] = balance.BalanceMinor
	}
	if byContainer["safe-1"] != 250000 || byContainer["bank-collection-1"] != 200000 || byContainer["vendor-supplies"] != 50000 {
		t.Fatalf("cash balances = %+v", balances.Balances)
	}
}

func TestCashRecountDiscrepancyRequiresApproval(t *testing.T) {
	server := NewServer()

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/cash-recounts", bytes.NewBufferString(`{
		"containerId": "safe-1",
		"containerType": "safe",
		"countedMinor": 100000,
		"reason": "Safe recount",
		"actorId": "senior-1"
	}`))
	request.Header.Set("Idempotency-Key", "recount-1")

	server.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("cash recount approval status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestCashRecountDiscrepancyCanBeResolved(t *testing.T) {
	server := NewServer()

	createResponse := httptest.NewRecorder()
	createRequest := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/cash-recounts", bytes.NewBufferString(`{
		"containerId": "safe-1",
		"containerType": "safe",
		"countedMinor": 100000,
		"reason": "Safe recount",
		"actorId": "senior-1",
		"approvedById": "cashier-1"
	}`))
	createRequest.Header.Set("Idempotency-Key", "recount-1")

	server.ServeHTTP(createResponse, createRequest)

	if createResponse.Code != http.StatusAccepted {
		t.Fatalf("create cash recount status = %d, body = %s", createResponse.Code, createResponse.Body.String())
	}

	var created CashRecountAcceptedResponse
	if err := json.Unmarshal(createResponse.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode cash recount response: %v", err)
	}
	if created.Recount.ResolutionStatus != "open" {
		t.Fatalf("created cash recount = %+v", created.Recount)
	}

	resolveResponse := httptest.NewRecorder()
	resolveRequest := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/cash-recounts/"+created.Recount.ID+"/resolve", bytes.NewBufferString(`{
		"resolutionNote": "Adjustment movement posted",
		"actorId": "senior-1",
		"approvedById": "admin-1"
	}`))
	resolveRequest.Header.Set("Idempotency-Key", "recount-resolve-1")

	server.ServeHTTP(resolveResponse, resolveRequest)

	if resolveResponse.Code != http.StatusAccepted {
		t.Fatalf("resolve cash recount status = %d, body = %s", resolveResponse.Code, resolveResponse.Body.String())
	}

	var resolved CashRecountAcceptedResponse
	if err := json.Unmarshal(resolveResponse.Body.Bytes(), &resolved); err != nil {
		t.Fatalf("decode resolved cash recount response: %v", err)
	}
	if resolved.Recount.ResolutionStatus != "resolved" || resolved.Recount.ResolvedByID != "senior-1" {
		t.Fatalf("resolved cash recount = %+v", resolved.Recount)
	}
}

func TestShiftWorkflow(t *testing.T) {
	server := NewServer()

	openDay(t, server, "shift-workflow")

	openResponse := httptest.NewRecorder()
	openRequest := httptest.NewRequest(http.MethodPost, "/v1/shifts", bytes.NewBufferString(`{
		"storeId": "store-1",
		"terminalId": "pos-1",
		"cashierId": "cashier-1",
		"drawerId": "drawer-1",
		"sourceSafeId": "safe-1",
		"openingCashMinor": 100000
	}`))
	openRequest.Header.Set("Idempotency-Key", "shift-open-1")

	server.ServeHTTP(openResponse, openRequest)

	if openResponse.Code != http.StatusAccepted {
		t.Fatalf("open shift status = %d, body = %s", openResponse.Code, openResponse.Body.String())
	}

	var opened ShiftAcceptedResponse
	if err := json.Unmarshal(openResponse.Body.Bytes(), &opened); err != nil {
		t.Fatalf("decode open shift response: %v", err)
	}
	if opened.Shift.ID == "" || opened.Shift.Status != "open" {
		t.Fatalf("opened shift = %+v", opened.Shift)
	}
	if opened.Shift.OperationalDayID == "" || opened.Shift.BusinessDate != "2026-06-18" {
		t.Fatalf("opened shift operational day links = %+v", opened.Shift)
	}

	listOpenResponse := httptest.NewRecorder()
	listOpenRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/shifts/open", nil)

	server.ServeHTTP(listOpenResponse, listOpenRequest)

	if listOpenResponse.Code != http.StatusOK {
		t.Fatalf("list open shifts status = %d, body = %s", listOpenResponse.Code, listOpenResponse.Body.String())
	}

	var openShifts ShiftsResponse
	if err := json.Unmarshal(listOpenResponse.Body.Bytes(), &openShifts); err != nil {
		t.Fatalf("decode open shifts response: %v", err)
	}
	if len(openShifts.Shifts) != 1 {
		t.Fatalf("open shifts count = %d", len(openShifts.Shifts))
	}

	closeResponse := httptest.NewRecorder()
	closeRequest := httptest.NewRequest(http.MethodPost, "/v1/shifts/"+opened.Shift.ID+"/close", bytes.NewBufferString(`{
		"closingCashMinor": 125000,
		"safeId": "safe-1",
		"actorId": "cashier-1",
		"approvedById": "senior-1"
	}`))
	closeRequest.Header.Set("Idempotency-Key", "shift-close-1")

	server.ServeHTTP(closeResponse, closeRequest)

	if closeResponse.Code != http.StatusAccepted {
		t.Fatalf("close shift status = %d, body = %s", closeResponse.Code, closeResponse.Body.String())
	}

	var closed ShiftAcceptedResponse
	if err := json.Unmarshal(closeResponse.Body.Bytes(), &closed); err != nil {
		t.Fatalf("decode close shift response: %v", err)
	}
	if closed.Shift.Status != "closed" || closed.Shift.ClosingCashMinor != 125000 {
		t.Fatalf("closed shift = %+v", closed.Shift)
	}

	cashBalanceResponse := httptest.NewRecorder()
	cashBalanceRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/cash-balances", nil)

	server.ServeHTTP(cashBalanceResponse, cashBalanceRequest)

	if cashBalanceResponse.Code != http.StatusOK {
		t.Fatalf("cash balances status = %d, body = %s", cashBalanceResponse.Code, cashBalanceResponse.Body.String())
	}

	var cashBalances CashBalancesResponse
	if err := json.Unmarshal(cashBalanceResponse.Body.Bytes(), &cashBalances); err != nil {
		t.Fatalf("decode cash balances response: %v", err)
	}
	shiftBalances := map[string]int64{}
	for _, balance := range cashBalances.Balances {
		shiftBalances[balance.ContainerID] = balance.BalanceMinor
	}
	if shiftBalances["drawer-1"] != -25000 || shiftBalances["safe-1"] != 25000 {
		t.Fatalf("cash balances = %+v", cashBalances.Balances)
	}

	finalOpenResponse := httptest.NewRecorder()
	finalOpenRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/shifts/open", nil)

	server.ServeHTTP(finalOpenResponse, finalOpenRequest)

	if finalOpenResponse.Code != http.StatusOK {
		t.Fatalf("final list open shifts status = %d, body = %s", finalOpenResponse.Code, finalOpenResponse.Body.String())
	}

	var finalOpenShifts ShiftsResponse
	if err := json.Unmarshal(finalOpenResponse.Body.Bytes(), &finalOpenShifts); err != nil {
		t.Fatalf("decode final open shifts response: %v", err)
	}
	if len(finalOpenShifts.Shifts) != 0 {
		t.Fatalf("final open shifts count = %d", len(finalOpenShifts.Shifts))
	}
}

func TestShiftCashInOutWorkflow(t *testing.T) {
	server := NewServer()
	openStoreDayAndShift(t, server, "shift-cash-io")

	openShiftsResponse := httptest.NewRecorder()
	openShiftsRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/shifts/open", nil)
	server.ServeHTTP(openShiftsResponse, openShiftsRequest)
	if openShiftsResponse.Code != http.StatusOK {
		t.Fatalf("open shifts status = %d, body = %s", openShiftsResponse.Code, openShiftsResponse.Body.String())
	}

	var openShifts ShiftsResponse
	if err := json.Unmarshal(openShiftsResponse.Body.Bytes(), &openShifts); err != nil {
		t.Fatalf("decode open shifts response: %v", err)
	}
	if len(openShifts.Shifts) != 1 {
		t.Fatalf("open shifts count = %d", len(openShifts.Shifts))
	}
	shiftID := openShifts.Shifts[0].ID

	cashInResponse := httptest.NewRecorder()
	cashInRequest := httptest.NewRequest(http.MethodPost, "/v1/shifts/"+shiftID+"/cash-in", bytes.NewBufferString(`{
		"amountMinor": 25000,
		"reason": "Top-up",
		"actorId": "cashier-1"
	}`))
	cashInRequest.Header.Set("Idempotency-Key", "shift-cash-in-1")
	server.ServeHTTP(cashInResponse, cashInRequest)
	if cashInResponse.Code != http.StatusAccepted {
		t.Fatalf("cash in status = %d, body = %s", cashInResponse.Code, cashInResponse.Body.String())
	}

	cashOutResponse := httptest.NewRecorder()
	cashOutRequest := httptest.NewRequest(http.MethodPost, "/v1/shifts/"+shiftID+"/cash-out", bytes.NewBufferString(`{
		"amountMinor": 50000,
		"safeId": "safe-1",
		"reason": "Revenue collection",
		"actorId": "cashier-1",
		"approvedById": "senior-1"
	}`))
	cashOutRequest.Header.Set("Idempotency-Key", "shift-cash-out-1")
	server.ServeHTTP(cashOutResponse, cashOutRequest)
	if cashOutResponse.Code != http.StatusAccepted {
		t.Fatalf("cash out status = %d, body = %s", cashOutResponse.Code, cashOutResponse.Body.String())
	}

	cashBalanceResponse := httptest.NewRecorder()
	cashBalanceRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/cash-balances", nil)
	server.ServeHTTP(cashBalanceResponse, cashBalanceRequest)
	if cashBalanceResponse.Code != http.StatusOK {
		t.Fatalf("cash balances status = %d, body = %s", cashBalanceResponse.Code, cashBalanceResponse.Body.String())
	}

	var cashBalances CashBalancesResponse
	if err := json.Unmarshal(cashBalanceResponse.Body.Bytes(), &cashBalances); err != nil {
		t.Fatalf("decode cash balances response: %v", err)
	}
	shiftBalances := map[string]int64{}
	for _, balance := range cashBalances.Balances {
		shiftBalances[balance.ContainerID] = balance.BalanceMinor
	}
	if shiftBalances["drawer-1"] != 75000 || shiftBalances["safe-1"] != -50000 {
		t.Fatalf("cash balances = %+v", cashBalances.Balances)
	}
}

func TestCloseShiftBlocksUnresolvedReceipt(t *testing.T) {
	server := NewServer()
	openStoreDayAndShift(t, server, "shift-unresolved")

	openShiftsResponse := httptest.NewRecorder()
	openShiftsRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/shifts/open", nil)
	server.ServeHTTP(openShiftsResponse, openShiftsRequest)

	if openShiftsResponse.Code != http.StatusOK {
		t.Fatalf("open shifts status = %d, body = %s", openShiftsResponse.Code, openShiftsResponse.Body.String())
	}

	var openShifts ShiftsResponse
	if err := json.Unmarshal(openShiftsResponse.Body.Bytes(), &openShifts); err != nil {
		t.Fatalf("decode open shifts response: %v", err)
	}
	if len(openShifts.Shifts) != 1 {
		t.Fatalf("open shifts count = %d", len(openShifts.Shifts))
	}

	openReceiptResponse := httptest.NewRecorder()
	openReceiptRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts", bytes.NewBufferString(`{
		"storeId": "store-1",
		"terminalId": "pos-1",
		"cashierId": "cashier-1",
		"channel": "pos"
	}`))
	openReceiptRequest.Header.Set("Idempotency-Key", "shift-unresolved-receipt-open-1")
	server.ServeHTTP(openReceiptResponse, openReceiptRequest)

	if openReceiptResponse.Code != http.StatusAccepted {
		t.Fatalf("open receipt status = %d, body = %s", openReceiptResponse.Code, openReceiptResponse.Body.String())
	}

	var openedReceipt ReceiptAcceptedResponse
	if err := json.Unmarshal(openReceiptResponse.Body.Bytes(), &openedReceipt); err != nil {
		t.Fatalf("decode open receipt response: %v", err)
	}

	blockedCloseResponse := httptest.NewRecorder()
	blockedCloseRequest := httptest.NewRequest(http.MethodPost, "/v1/shifts/"+openShifts.Shifts[0].ID+"/close", bytes.NewBufferString(`{
		"closingCashMinor": 0
	}`))
	blockedCloseRequest.Header.Set("Idempotency-Key", "shift-unresolved-close-1")
	server.ServeHTTP(blockedCloseResponse, blockedCloseRequest)

	if blockedCloseResponse.Code != http.StatusConflict {
		t.Fatalf("blocked close shift status = %d, body = %s", blockedCloseResponse.Code, blockedCloseResponse.Body.String())
	}

	var problem map[string]any
	if err := json.Unmarshal(blockedCloseResponse.Body.Bytes(), &problem); err != nil {
		t.Fatalf("decode blocked close problem: %v", err)
	}
	if problem["code"] != "shift_close_blocked" {
		t.Fatalf("problem = %+v", problem)
	}

	cancelResponse := httptest.NewRecorder()
	cancelRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/cancel", bytes.NewBufferString(`{
		"reason": "Customer changed mind",
		"actorId": "cashier-1"
	}`))
	cancelRequest.Header.Set("Idempotency-Key", "shift-unresolved-receipt-cancel-1")
	server.ServeHTTP(cancelResponse, cancelRequest)

	if cancelResponse.Code != http.StatusAccepted {
		t.Fatalf("cancel receipt status = %d, body = %s", cancelResponse.Code, cancelResponse.Body.String())
	}

	closeResponse := httptest.NewRecorder()
	closeRequest := httptest.NewRequest(http.MethodPost, "/v1/shifts/"+openShifts.Shifts[0].ID+"/close", bytes.NewBufferString(`{
		"closingCashMinor": 0
	}`))
	closeRequest.Header.Set("Idempotency-Key", "shift-unresolved-close-2")
	server.ServeHTTP(closeResponse, closeRequest)

	if closeResponse.Code != http.StatusAccepted {
		t.Fatalf("close shift after cancel status = %d, body = %s", closeResponse.Code, closeResponse.Body.String())
	}
}

func TestListShiftReceipts(t *testing.T) {
	server := NewServer()
	openStoreDayAndShift(t, server, "shift-receipts")

	openShiftsResponse := httptest.NewRecorder()
	openShiftsRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/shifts/open", nil)
	server.ServeHTTP(openShiftsResponse, openShiftsRequest)

	if openShiftsResponse.Code != http.StatusOK {
		t.Fatalf("open shifts status = %d, body = %s", openShiftsResponse.Code, openShiftsResponse.Body.String())
	}

	var openShifts ShiftsResponse
	if err := json.Unmarshal(openShiftsResponse.Body.Bytes(), &openShifts); err != nil {
		t.Fatalf("decode open shifts response: %v", err)
	}
	if len(openShifts.Shifts) != 1 {
		t.Fatalf("open shifts count = %d", len(openShifts.Shifts))
	}

	openReceiptResponse := httptest.NewRecorder()
	openReceiptRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts", bytes.NewBufferString(`{
		"storeId": "store-1",
		"terminalId": "pos-1",
		"cashierId": "cashier-1",
		"channel": "pos"
	}`))
	openReceiptRequest.Header.Set("Idempotency-Key", "shift-receipts-open-1")
	server.ServeHTTP(openReceiptResponse, openReceiptRequest)

	if openReceiptResponse.Code != http.StatusAccepted {
		t.Fatalf("open receipt status = %d, body = %s", openReceiptResponse.Code, openReceiptResponse.Body.String())
	}

	var openedReceipt ReceiptAcceptedResponse
	if err := json.Unmarshal(openReceiptResponse.Body.Bytes(), &openedReceipt); err != nil {
		t.Fatalf("decode open receipt response: %v", err)
	}

	listResponse := httptest.NewRecorder()
	listRequest := httptest.NewRequest(http.MethodGet, "/v1/shifts/"+openShifts.Shifts[0].ID+"/receipts", nil)
	server.ServeHTTP(listResponse, listRequest)

	if listResponse.Code != http.StatusOK {
		t.Fatalf("list shift receipts status = %d, body = %s", listResponse.Code, listResponse.Body.String())
	}

	var listed ReceiptsResponse
	if err := json.Unmarshal(listResponse.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode shift receipts response: %v", err)
	}
	if len(listed.Receipts) != 1 || listed.Receipts[0].ID != openedReceipt.Receipt.ID || listed.Receipts[0].Status != "draft" {
		t.Fatalf("listed receipts = %+v", listed.Receipts)
	}

	cancelResponse := httptest.NewRecorder()
	cancelRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/cancel", bytes.NewBufferString(`{
		"reason": "Customer changed mind",
		"actorId": "cashier-1"
	}`))
	cancelRequest.Header.Set("Idempotency-Key", "shift-receipts-cancel-1")
	server.ServeHTTP(cancelResponse, cancelRequest)

	if cancelResponse.Code != http.StatusAccepted {
		t.Fatalf("cancel receipt status = %d, body = %s", cancelResponse.Code, cancelResponse.Body.String())
	}

	listAfterCancelResponse := httptest.NewRecorder()
	listAfterCancelRequest := httptest.NewRequest(http.MethodGet, "/v1/shifts/"+openShifts.Shifts[0].ID+"/receipts", nil)
	server.ServeHTTP(listAfterCancelResponse, listAfterCancelRequest)

	if listAfterCancelResponse.Code != http.StatusOK {
		t.Fatalf("list shift receipts after cancel status = %d, body = %s", listAfterCancelResponse.Code, listAfterCancelResponse.Body.String())
	}

	var listedAfterCancel ReceiptsResponse
	if err := json.Unmarshal(listAfterCancelResponse.Body.Bytes(), &listedAfterCancel); err != nil {
		t.Fatalf("decode shift receipts after cancel response: %v", err)
	}
	if len(listedAfterCancel.Receipts) != 1 || listedAfterCancel.Receipts[0].Status != "cancelled" {
		t.Fatalf("listed receipts after cancel = %+v", listedAfterCancel.Receipts)
	}
}

func TestShiftRejectsDuplicateOpenTerminal(t *testing.T) {
	server := NewServer()
	openDay(t, server, "shift-duplicate-terminal")

	firstResponse := httptest.NewRecorder()
	firstRequest := httptest.NewRequest(http.MethodPost, "/v1/shifts", bytes.NewBufferString(`{
		"storeId": "store-1",
		"terminalId": "pos-1",
		"cashierId": "cashier-1",
		"drawerId": "drawer-1",
		"sourceSafeId": "safe-1",
		"openingCashMinor": 100000
	}`))
	firstRequest.Header.Set("Idempotency-Key", "shift-open-1")
	server.ServeHTTP(firstResponse, firstRequest)

	if firstResponse.Code != http.StatusAccepted {
		t.Fatalf("first open shift status = %d, body = %s", firstResponse.Code, firstResponse.Body.String())
	}

	secondResponse := httptest.NewRecorder()
	secondRequest := httptest.NewRequest(http.MethodPost, "/v1/shifts", bytes.NewBufferString(`{
		"storeId": "store-1",
		"terminalId": "pos-1",
		"cashierId": "cashier-2",
		"drawerId": "drawer-2",
		"sourceSafeId": "safe-1",
		"openingCashMinor": 100000
	}`))
	secondRequest.Header.Set("Idempotency-Key", "shift-open-2")
	server.ServeHTTP(secondResponse, secondRequest)

	if secondResponse.Code != http.StatusConflict {
		t.Fatalf("duplicate terminal shift status = %d, body = %s", secondResponse.Code, secondResponse.Body.String())
	}
}

func TestOpenShiftRequiresOpenOperationalDay(t *testing.T) {
	server := NewServer()

	openResponse := httptest.NewRecorder()
	openRequest := httptest.NewRequest(http.MethodPost, "/v1/shifts", bytes.NewBufferString(`{
		"storeId": "store-1",
		"terminalId": "pos-1",
		"cashierId": "cashier-1",
		"drawerId": "drawer-1",
		"sourceSafeId": "safe-1",
		"openingCashMinor": 100000
	}`))
	openRequest.Header.Set("Idempotency-Key", "shift-open-without-day-1")

	server.ServeHTTP(openResponse, openRequest)

	if openResponse.Code != http.StatusConflict {
		t.Fatalf("open shift without day status = %d, body = %s", openResponse.Code, openResponse.Body.String())
	}

	var problem map[string]any
	if err := json.Unmarshal(openResponse.Body.Bytes(), &problem); err != nil {
		t.Fatalf("decode open shift problem: %v", err)
	}
	if problem["code"] != "open_operational_day_required" {
		t.Fatalf("problem = %+v", problem)
	}
}

func TestOperationalDayWorkflowRequiresNoSalesOverride(t *testing.T) {
	server := NewServer()

	openResponse := httptest.NewRecorder()
	openRequest := httptest.NewRequest(http.MethodPost, "/v1/operational-days", bytes.NewBufferString(`{
		"storeId": "store-1",
		"businessDate": "2026-06-18",
		"openedById": "senior-1"
	}`))
	openRequest.Header.Set("Idempotency-Key", "oday-open-1")

	server.ServeHTTP(openResponse, openRequest)

	if openResponse.Code != http.StatusAccepted {
		t.Fatalf("open operational day status = %d, body = %s", openResponse.Code, openResponse.Body.String())
	}

	var opened OperationalDayAcceptedResponse
	if err := json.Unmarshal(openResponse.Body.Bytes(), &opened); err != nil {
		t.Fatalf("decode open operational day response: %v", err)
	}
	if opened.OperationalDay.ID == "" || opened.OperationalDay.Status != "open" {
		t.Fatalf("opened operational day = %+v", opened.OperationalDay)
	}

	currentResponse := httptest.NewRecorder()
	currentRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/operational-days/current", nil)

	server.ServeHTTP(currentResponse, currentRequest)

	if currentResponse.Code != http.StatusOK {
		t.Fatalf("current operational day status = %d, body = %s", currentResponse.Code, currentResponse.Body.String())
	}

	readinessResponse := httptest.NewRecorder()
	readinessRequest := httptest.NewRequest(http.MethodPost, "/v1/operational-days/"+opened.OperationalDay.ID+"/close-check", nil)

	server.ServeHTTP(readinessResponse, readinessRequest)

	if readinessResponse.Code != http.StatusOK {
		t.Fatalf("close readiness status = %d, body = %s", readinessResponse.Code, readinessResponse.Body.String())
	}

	var readiness OperationalDayCloseReadinessResponse
	if err := json.Unmarshal(readinessResponse.Body.Bytes(), &readiness); err != nil {
		t.Fatalf("decode readiness response: %v", err)
	}
	if readiness.CanClose || len(readiness.Blockers) != 1 || readiness.Blockers[0].Code != "no_sales_receipts" {
		t.Fatalf("readiness = %+v", readiness)
	}

	blockedCloseResponse := httptest.NewRecorder()
	blockedCloseRequest := httptest.NewRequest(http.MethodPost, "/v1/operational-days/"+opened.OperationalDay.ID+"/close", bytes.NewBufferString(`{
		"closedById": "senior-1"
	}`))
	blockedCloseRequest.Header.Set("Idempotency-Key", "oday-close-1")

	server.ServeHTTP(blockedCloseResponse, blockedCloseRequest)

	if blockedCloseResponse.Code != http.StatusConflict {
		t.Fatalf("blocked close status = %d, body = %s", blockedCloseResponse.Code, blockedCloseResponse.Body.String())
	}

	closeResponse := httptest.NewRecorder()
	closeRequest := httptest.NewRequest(http.MethodPost, "/v1/operational-days/"+opened.OperationalDay.ID+"/close", bytes.NewBufferString(`{
		"closedById": "senior-1",
		"overrideNoSales": true,
		"overrideActorId": "admin-1"
	}`))
	closeRequest.Header.Set("Idempotency-Key", "oday-close-2")

	server.ServeHTTP(closeResponse, closeRequest)

	if closeResponse.Code != http.StatusAccepted {
		t.Fatalf("override close status = %d, body = %s", closeResponse.Code, closeResponse.Body.String())
	}

	var closed OperationalDayAcceptedResponse
	if err := json.Unmarshal(closeResponse.Body.Bytes(), &closed); err != nil {
		t.Fatalf("decode close operational day response: %v", err)
	}
	if closed.OperationalDay.Status != "closed" || closed.OperationalDay.ClosedByID != "senior-1" {
		t.Fatalf("closed operational day = %+v", closed.OperationalDay)
	}
}

func TestOperationalDayCloseCheckBlocksUnresolvedReceipt(t *testing.T) {
	server := NewServer()
	openStoreDayAndShift(t, server, "unresolved")

	dayResponse := httptest.NewRecorder()
	dayRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/operational-days/current", nil)
	server.ServeHTTP(dayResponse, dayRequest)

	if dayResponse.Code != http.StatusOK {
		t.Fatalf("current operational day status = %d, body = %s", dayResponse.Code, dayResponse.Body.String())
	}

	var day OperationalDayResponse
	if err := json.Unmarshal(dayResponse.Body.Bytes(), &day); err != nil {
		t.Fatalf("decode current operational day response: %v", err)
	}

	openReceiptResponse := httptest.NewRecorder()
	openReceiptRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts", bytes.NewBufferString(`{
		"storeId": "store-1",
		"terminalId": "pos-1",
		"cashierId": "cashier-1",
		"channel": "pos"
	}`))
	openReceiptRequest.Header.Set("Idempotency-Key", "unresolved-receipt-open-1")
	server.ServeHTTP(openReceiptResponse, openReceiptRequest)

	if openReceiptResponse.Code != http.StatusAccepted {
		t.Fatalf("open receipt status = %d, body = %s", openReceiptResponse.Code, openReceiptResponse.Body.String())
	}

	var opened ReceiptAcceptedResponse
	if err := json.Unmarshal(openReceiptResponse.Body.Bytes(), &opened); err != nil {
		t.Fatalf("decode open receipt response: %v", err)
	}

	readinessResponse := httptest.NewRecorder()
	readinessRequest := httptest.NewRequest(http.MethodPost, "/v1/operational-days/"+day.ID+"/close-check", nil)
	server.ServeHTTP(readinessResponse, readinessRequest)

	if readinessResponse.Code != http.StatusOK {
		t.Fatalf("close readiness status = %d, body = %s", readinessResponse.Code, readinessResponse.Body.String())
	}

	var readiness OperationalDayCloseReadinessResponse
	if err := json.Unmarshal(readinessResponse.Body.Bytes(), &readiness); err != nil {
		t.Fatalf("decode readiness response: %v", err)
	}
	if !hasOperationalDayBlocker(readiness.Blockers, "unresolved_receipt") {
		t.Fatalf("expected unresolved receipt blocker, got %+v", readiness.Blockers)
	}

	cancelResponse := httptest.NewRecorder()
	cancelRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+opened.Receipt.ID+"/cancel", bytes.NewBufferString(`{
		"reason": "Customer changed mind",
		"actorId": "cashier-1"
	}`))
	cancelRequest.Header.Set("Idempotency-Key", "unresolved-receipt-cancel-1")
	server.ServeHTTP(cancelResponse, cancelRequest)

	if cancelResponse.Code != http.StatusAccepted {
		t.Fatalf("cancel receipt status = %d, body = %s", cancelResponse.Code, cancelResponse.Body.String())
	}

	readinessAfterCancelResponse := httptest.NewRecorder()
	readinessAfterCancelRequest := httptest.NewRequest(http.MethodPost, "/v1/operational-days/"+day.ID+"/close-check", nil)
	server.ServeHTTP(readinessAfterCancelResponse, readinessAfterCancelRequest)

	if readinessAfterCancelResponse.Code != http.StatusOK {
		t.Fatalf("close readiness after cancel status = %d, body = %s", readinessAfterCancelResponse.Code, readinessAfterCancelResponse.Body.String())
	}

	var readinessAfterCancel OperationalDayCloseReadinessResponse
	if err := json.Unmarshal(readinessAfterCancelResponse.Body.Bytes(), &readinessAfterCancel); err != nil {
		t.Fatalf("decode readiness after cancel response: %v", err)
	}
	if hasOperationalDayBlocker(readinessAfterCancel.Blockers, "unresolved_receipt") {
		t.Fatalf("unexpected unresolved receipt blocker after cancel: %+v", readinessAfterCancel.Blockers)
	}
}

func TestListOperationalDayReceipts(t *testing.T) {
	server := NewServer()
	openStoreDayAndShift(t, server, "oday-receipts")

	dayResponse := httptest.NewRecorder()
	dayRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/operational-days/current", nil)
	server.ServeHTTP(dayResponse, dayRequest)

	if dayResponse.Code != http.StatusOK {
		t.Fatalf("current operational day status = %d, body = %s", dayResponse.Code, dayResponse.Body.String())
	}

	var day OperationalDayResponse
	if err := json.Unmarshal(dayResponse.Body.Bytes(), &day); err != nil {
		t.Fatalf("decode current operational day response: %v", err)
	}

	openReceiptResponse := httptest.NewRecorder()
	openReceiptRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts", bytes.NewBufferString(`{
		"storeId": "store-1",
		"terminalId": "pos-1",
		"cashierId": "cashier-1",
		"channel": "pos"
	}`))
	openReceiptRequest.Header.Set("Idempotency-Key", "oday-receipts-open-1")
	server.ServeHTTP(openReceiptResponse, openReceiptRequest)

	if openReceiptResponse.Code != http.StatusAccepted {
		t.Fatalf("open receipt status = %d, body = %s", openReceiptResponse.Code, openReceiptResponse.Body.String())
	}

	var openedReceipt ReceiptAcceptedResponse
	if err := json.Unmarshal(openReceiptResponse.Body.Bytes(), &openedReceipt); err != nil {
		t.Fatalf("decode open receipt response: %v", err)
	}

	listResponse := httptest.NewRecorder()
	listRequest := httptest.NewRequest(http.MethodGet, "/v1/operational-days/"+day.ID+"/receipts", nil)
	server.ServeHTTP(listResponse, listRequest)

	if listResponse.Code != http.StatusOK {
		t.Fatalf("list operational day receipts status = %d, body = %s", listResponse.Code, listResponse.Body.String())
	}

	var listed PaginatedReceiptsResponse
	if err := json.Unmarshal(listResponse.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode operational day receipts response: %v", err)
	}
	if len(listed.Items) != 1 ||
		listed.Items[0].ID != openedReceipt.Receipt.ID ||
		listed.Items[0].OperationalDayID != day.ID {
		t.Fatalf("listed receipts = %+v", listed.Items)
	}
}

func TestListOperationalDayShifts(t *testing.T) {
	server := NewServer()
	openStoreDayAndShift(t, server, "oday-shifts")

	dayResponse := httptest.NewRecorder()
	dayRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/operational-days/current", nil)
	server.ServeHTTP(dayResponse, dayRequest)

	if dayResponse.Code != http.StatusOK {
		t.Fatalf("current operational day status = %d, body = %s", dayResponse.Code, dayResponse.Body.String())
	}

	var day OperationalDayResponse
	if err := json.Unmarshal(dayResponse.Body.Bytes(), &day); err != nil {
		t.Fatalf("decode current operational day response: %v", err)
	}

	listResponse := httptest.NewRecorder()
	listRequest := httptest.NewRequest(http.MethodGet, "/v1/operational-days/"+day.ID+"/shifts", nil)
	server.ServeHTTP(listResponse, listRequest)

	if listResponse.Code != http.StatusOK {
		t.Fatalf("list operational day shifts status = %d, body = %s", listResponse.Code, listResponse.Body.String())
	}

	var listed PaginatedShiftsResponse
	if err := json.Unmarshal(listResponse.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode operational day shifts response: %v", err)
	}
	if len(listed.Items) != 1 ||
		listed.Items[0].OperationalDayID != day.ID ||
		listed.Items[0].BusinessDate != "2026-06-18" {
		t.Fatalf("listed shifts = %+v", listed.Items)
	}
}

func TestOperationalDaySummary(t *testing.T) {
	server := NewServer()
	openStoreDayAndShift(t, server, "oday-summary")

	dayResponse := httptest.NewRecorder()
	dayRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/operational-days/current", nil)
	server.ServeHTTP(dayResponse, dayRequest)

	if dayResponse.Code != http.StatusOK {
		t.Fatalf("current operational day status = %d, body = %s", dayResponse.Code, dayResponse.Body.String())
	}

	var day OperationalDayResponse
	if err := json.Unmarshal(dayResponse.Body.Bytes(), &day); err != nil {
		t.Fatalf("decode current operational day response: %v", err)
	}

	openReceiptResponse := httptest.NewRecorder()
	openReceiptRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts", bytes.NewBufferString(`{
		"storeId": "store-1",
		"terminalId": "pos-1",
		"cashierId": "cashier-1",
		"channel": "pos"
	}`))
	openReceiptRequest.Header.Set("Idempotency-Key", "oday-summary-open-1")
	server.ServeHTTP(openReceiptResponse, openReceiptRequest)

	if openReceiptResponse.Code != http.StatusAccepted {
		t.Fatalf("open receipt status = %d, body = %s", openReceiptResponse.Code, openReceiptResponse.Body.String())
	}

	var openedReceipt ReceiptAcceptedResponse
	if err := json.Unmarshal(openReceiptResponse.Body.Bytes(), &openedReceipt); err != nil {
		t.Fatalf("decode open receipt response: %v", err)
	}

	addLineResponse := httptest.NewRecorder()
	addLineRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/lines", bytes.NewBufferString(`{
		"productId": "sku-1",
		"name": "Milk",
		"quantity": 1,
		"unitPriceMinor": 50000
	}`))
	addLineRequest.Header.Set("Idempotency-Key", "oday-summary-line-1")
	server.ServeHTTP(addLineResponse, addLineRequest)

	if addLineResponse.Code != http.StatusAccepted {
		t.Fatalf("add line status = %d, body = %s", addLineResponse.Code, addLineResponse.Body.String())
	}

	paymentResponse := httptest.NewRecorder()
	paymentRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/payments", bytes.NewBufferString(`{
		"method": "card_mock",
		"amountMinor": 50000,
		"providerReference": "card-summary-1"
	}`))
	paymentRequest.Header.Set("Idempotency-Key", "oday-summary-payment-1")
	server.ServeHTTP(paymentResponse, paymentRequest)

	if paymentResponse.Code != http.StatusAccepted {
		t.Fatalf("create payment status = %d, body = %s", paymentResponse.Code, paymentResponse.Body.String())
	}

	fiscalResponse := httptest.NewRecorder()
	fiscalRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+openedReceipt.Receipt.ID+"/fiscal-documents", bytes.NewBufferString(`{
		"deviceId": "mock-atol-1"
	}`))
	fiscalRequest.Header.Set("Idempotency-Key", "oday-summary-fiscal-1")
	server.ServeHTTP(fiscalResponse, fiscalRequest)

	if fiscalResponse.Code != http.StatusAccepted {
		t.Fatalf("create fiscal document status = %d, body = %s", fiscalResponse.Code, fiscalResponse.Body.String())
	}

	movementResponse := httptest.NewRecorder()
	movementRequest := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/cash-movements", bytes.NewBufferString(`{
		"type": "cash_sale",
		"fromContainerId": "external-customer",
		"fromContainerType": "external",
		"toContainerId": "drawer-1",
		"toContainerType": "drawer",
		"amountMinor": 50000,
		"currency": "RUB",
		"reason": "Cash sale",
		"actorId": "cashier-1"
	}`))
	movementRequest.Header.Set("Idempotency-Key", "oday-summary-cash-1")
	server.ServeHTTP(movementResponse, movementRequest)

	if movementResponse.Code != http.StatusAccepted {
		t.Fatalf("create cash movement status = %d, body = %s", movementResponse.Code, movementResponse.Body.String())
	}

	recountResponse := httptest.NewRecorder()
	recountRequest := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/cash-recounts", bytes.NewBufferString(`{
		"containerId": "drawer-1",
		"containerType": "drawer",
		"currency": "RUB",
		"countedMinor": 40000,
		"reason": "Drawer recount",
		"actorId": "cashier-1",
		"approvedById": "senior-1"
	}`))
	recountRequest.Header.Set("Idempotency-Key", "oday-summary-recount-1")
	server.ServeHTTP(recountResponse, recountRequest)

	if recountResponse.Code != http.StatusAccepted {
		t.Fatalf("create cash recount status = %d, body = %s", recountResponse.Code, recountResponse.Body.String())
	}

	summaryResponse := httptest.NewRecorder()
	summaryRequest := httptest.NewRequest(http.MethodGet, "/v1/operational-days/"+day.ID+"/summary", nil)
	server.ServeHTTP(summaryResponse, summaryRequest)

	if summaryResponse.Code != http.StatusOK {
		t.Fatalf("operational day summary status = %d, body = %s", summaryResponse.Code, summaryResponse.Body.String())
	}

	var summary OperationalDaySummaryResponse
	if err := json.Unmarshal(summaryResponse.Body.Bytes(), &summary); err != nil {
		t.Fatalf("decode operational day summary response: %v", err)
	}
	if summary.CanClose ||
		summary.Shifts.TotalCount != 1 ||
		summary.Shifts.OpenCount != 1 ||
		summary.Receipts.TotalCount != 1 ||
		summary.Receipts.FiscalizedCount != 1 ||
		summary.Receipts.UnresolvedCount != 0 {
		t.Fatalf("summary = %+v", summary)
	}
	if summary.Payments.TotalCount != 1 ||
		summary.Payments.CapturedCount != 1 ||
		summary.Payments.CapturedTotalMinor != 50000 ||
		len(summary.Payments.Methods) != 1 ||
		summary.Payments.Methods[0].Method != "card_mock" ||
		summary.Payments.Methods[0].CapturedTotalMinor != 50000 {
		t.Fatalf("payment summary = %+v", summary.Payments)
	}
	if summary.Fiscal.TotalCount != 1 ||
		summary.Fiscal.FiscalizedCount != 1 ||
		summary.Fiscal.FiscalizedTotalMinor != 50000 {
		t.Fatalf("fiscal summary = %+v", summary.Fiscal)
	}
	cashByContainer := map[string]int64{}
	for _, balance := range summary.Cash.Balances {
		cashByContainer[balance.ContainerID] = balance.BalanceMinor
	}
	if len(summary.Cash.Balances) != 2 ||
		cashByContainer["drawer-1"] != 150000 ||
		cashByContainer["safe-1"] != -100000 ||
		summary.Cash.NonZeroDrawerCount != 1 {
		t.Fatalf("cash summary = %+v", summary.Cash)
	}
	if summary.Cash.Recounts.TotalCount != 1 ||
		summary.Cash.Recounts.DiscrepancyCount != 1 ||
		summary.Cash.Recounts.OpenDiscrepancyCount != 1 {
		t.Fatalf("cash recount summary = %+v", summary.Cash.Recounts)
	}
	if hasOperationalDayBlocker(summary.Blockers, "unresolved_receipt") {
		t.Fatalf("unexpected unresolved receipt blocker, got %+v", summary.Blockers)
	}
}

func TestOperationalDayCloseCheckBlocksNonZeroDrawerBalance(t *testing.T) {
	server := NewServer()

	openResponse := httptest.NewRecorder()
	openRequest := httptest.NewRequest(http.MethodPost, "/v1/operational-days", bytes.NewBufferString(`{
		"storeId": "store-1",
		"businessDate": "2026-06-18",
		"openedById": "senior-1"
	}`))
	openRequest.Header.Set("Idempotency-Key", "oday-open-1")
	server.ServeHTTP(openResponse, openRequest)

	if openResponse.Code != http.StatusAccepted {
		t.Fatalf("open operational day status = %d, body = %s", openResponse.Code, openResponse.Body.String())
	}

	var opened OperationalDayAcceptedResponse
	if err := json.Unmarshal(openResponse.Body.Bytes(), &opened); err != nil {
		t.Fatalf("decode open operational day response: %v", err)
	}

	movementResponse := httptest.NewRecorder()
	movementRequest := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/cash-movements", bytes.NewBufferString(`{
		"type": "cash_sale",
		"fromContainerId": "external-customer",
		"fromContainerType": "external",
		"toContainerId": "drawer-1",
		"toContainerType": "drawer",
		"amountMinor": 50000,
		"currency": "RUB",
		"reason": "Cash sale",
		"actorId": "cashier-1"
	}`))
	movementRequest.Header.Set("Idempotency-Key", "cash-sale-1")
	server.ServeHTTP(movementResponse, movementRequest)

	if movementResponse.Code != http.StatusAccepted {
		t.Fatalf("create cash movement status = %d, body = %s", movementResponse.Code, movementResponse.Body.String())
	}

	readinessResponse := httptest.NewRecorder()
	readinessRequest := httptest.NewRequest(http.MethodPost, "/v1/operational-days/"+opened.OperationalDay.ID+"/close-check", nil)

	server.ServeHTTP(readinessResponse, readinessRequest)

	if readinessResponse.Code != http.StatusOK {
		t.Fatalf("close readiness status = %d, body = %s", readinessResponse.Code, readinessResponse.Body.String())
	}

	var readiness OperationalDayCloseReadinessResponse
	if err := json.Unmarshal(readinessResponse.Body.Bytes(), &readiness); err != nil {
		t.Fatalf("decode readiness response: %v", err)
	}
	if !hasOperationalDayBlocker(readiness.Blockers, "nonzero_drawer_balance") {
		t.Fatalf("expected nonzero drawer blocker, got %+v", readiness.Blockers)
	}
}

func TestScanReceiptWorkflow(t *testing.T) {
	server := NewServer()
	openStoreDayAndShift(t, server, "scan")

	productResponse := httptest.NewRecorder()
	productRequest := httptest.NewRequest(http.MethodGet, "/v1/catalog/products/by-barcode/4600000000000", nil)

	server.ServeHTTP(productResponse, productRequest)

	if productResponse.Code != http.StatusOK {
		t.Fatalf("product lookup status = %d, body = %s", productResponse.Code, productResponse.Body.String())
	}

	openResponse := httptest.NewRecorder()
	openRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts", bytes.NewBufferString(`{
		"storeId": "store-1",
		"terminalId": "pos-1",
		"cashierId": "cashier-1"
	}`))
	openRequest.Header.Set("Idempotency-Key", "open-scan-1")
	server.ServeHTTP(openResponse, openRequest)

	if openResponse.Code != http.StatusAccepted {
		t.Fatalf("open receipt status = %d, body = %s", openResponse.Code, openResponse.Body.String())
	}

	var opened ReceiptAcceptedResponse
	if err := json.Unmarshal(openResponse.Body.Bytes(), &opened); err != nil {
		t.Fatalf("decode open response: %v", err)
	}

	scanResponse := httptest.NewRecorder()
	scanRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+opened.Receipt.ID+"/scan", bytes.NewBufferString(`{
		"barcode": "4600000000000",
		"quantity": 2
	}`))
	scanRequest.Header.Set("Idempotency-Key", "scan-1")

	server.ServeHTTP(scanResponse, scanRequest)

	if scanResponse.Code != http.StatusAccepted {
		t.Fatalf("scan status = %d, body = %s", scanResponse.Code, scanResponse.Body.String())
	}

	var scanned ReceiptAcceptedResponse
	if err := json.Unmarshal(scanResponse.Body.Bytes(), &scanned); err != nil {
		t.Fatalf("decode scan response: %v", err)
	}
	if scanned.Receipt.TotalMinor != 39998 || scanned.Receipt.Lines[0].ProductID != "demo-milk-1" {
		t.Fatalf("scanned receipt = %+v", scanned.Receipt)
	}

	paymentResponse := httptest.NewRecorder()
	paymentRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+opened.Receipt.ID+"/payments", bytes.NewBufferString(`{
		"method": "cash",
		"amountMinor": 39998
	}`))
	paymentRequest.Header.Set("Idempotency-Key", "payment-1")

	server.ServeHTTP(paymentResponse, paymentRequest)

	if paymentResponse.Code != http.StatusAccepted {
		t.Fatalf("payment status = %d, body = %s", paymentResponse.Code, paymentResponse.Body.String())
	}

	var paid PaymentAcceptedResponse
	if err := json.Unmarshal(paymentResponse.Body.Bytes(), &paid); err != nil {
		t.Fatalf("decode payment response: %v", err)
	}
	if paid.Payment.Status != "captured" || paid.Payment.AmountMinor != 39998 {
		t.Fatalf("payment = %+v", paid.Payment)
	}

	cashBalanceResponse := httptest.NewRecorder()
	cashBalanceRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/cash-balances", nil)

	server.ServeHTTP(cashBalanceResponse, cashBalanceRequest)

	if cashBalanceResponse.Code != http.StatusOK {
		t.Fatalf("cash balances status = %d, body = %s", cashBalanceResponse.Code, cashBalanceResponse.Body.String())
	}

	var cashBalances CashBalancesResponse
	if err := json.Unmarshal(cashBalanceResponse.Body.Bytes(), &cashBalances); err != nil {
		t.Fatalf("decode cash balances response: %v", err)
	}
	cashByContainer := map[string]int64{}
	for _, balance := range cashBalances.Balances {
		cashByContainer[balance.ContainerID] = balance.BalanceMinor
	}
	if len(cashBalances.Balances) != 2 ||
		cashByContainer["drawer-1"] != 139998 ||
		cashByContainer["safe-1"] != -100000 {
		t.Fatalf("cash balances = %+v", cashBalances.Balances)
	}

	listResponse := httptest.NewRecorder()
	listRequest := httptest.NewRequest(http.MethodGet, "/v1/receipts/"+opened.Receipt.ID+"/payments", nil)

	server.ServeHTTP(listResponse, listRequest)

	if listResponse.Code != http.StatusOK {
		t.Fatalf("list payments status = %d, body = %s", listResponse.Code, listResponse.Body.String())
	}

	var listed PaymentsResponse
	if err := json.Unmarshal(listResponse.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode payments response: %v", err)
	}
	if len(listed.Payments) != 1 {
		t.Fatalf("payments count = %d", len(listed.Payments))
	}

	fiscalResponse := httptest.NewRecorder()
	fiscalRequest := httptest.NewRequest(http.MethodPost, "/v1/receipts/"+opened.Receipt.ID+"/fiscal-documents", bytes.NewBufferString(`{
		"deviceId": "mock-atol-1"
	}`))
	fiscalRequest.Header.Set("Idempotency-Key", "fiscal-1")

	server.ServeHTTP(fiscalResponse, fiscalRequest)

	if fiscalResponse.Code != http.StatusAccepted {
		t.Fatalf("fiscal status = %d, body = %s", fiscalResponse.Code, fiscalResponse.Body.String())
	}

	var fiscalized FiscalDocumentAcceptedResponse
	if err := json.Unmarshal(fiscalResponse.Body.Bytes(), &fiscalized); err != nil {
		t.Fatalf("decode fiscal response: %v", err)
	}
	if fiscalized.Document.Status != "fiscalized" || fiscalized.Document.AmountMinor != 39998 {
		t.Fatalf("fiscal document = %+v", fiscalized.Document)
	}

	fiscalListResponse := httptest.NewRecorder()
	fiscalListRequest := httptest.NewRequest(http.MethodGet, "/v1/receipts/"+opened.Receipt.ID+"/fiscal-documents", nil)

	server.ServeHTTP(fiscalListResponse, fiscalListRequest)

	if fiscalListResponse.Code != http.StatusOK {
		t.Fatalf("fiscal list status = %d, body = %s", fiscalListResponse.Code, fiscalListResponse.Body.String())
	}

	var fiscalDocuments FiscalDocumentsResponse
	if err := json.Unmarshal(fiscalListResponse.Body.Bytes(), &fiscalDocuments); err != nil {
		t.Fatalf("decode fiscal list response: %v", err)
	}
	if len(fiscalDocuments.Documents) != 1 {
		t.Fatalf("fiscal documents count = %d", len(fiscalDocuments.Documents))
	}

	finalReceiptResponse := httptest.NewRecorder()
	finalReceiptRequest := httptest.NewRequest(http.MethodGet, "/v1/receipts/"+opened.Receipt.ID, nil)

	server.ServeHTTP(finalReceiptResponse, finalReceiptRequest)

	if finalReceiptResponse.Code != http.StatusOK {
		t.Fatalf("final receipt status = %d, body = %s", finalReceiptResponse.Code, finalReceiptResponse.Body.String())
	}

	var finalReceipt ReceiptResponse
	if err := json.Unmarshal(finalReceiptResponse.Body.Bytes(), &finalReceipt); err != nil {
		t.Fatalf("decode final receipt response: %v", err)
	}
	if finalReceipt.Status != "fiscalized" {
		t.Fatalf("final receipt lifecycle status = %s", finalReceipt.Status)
	}
}

func openStoreDayAndShift(t *testing.T, server http.Handler, keyPrefix string) {
	t.Helper()
	openStoreDayAndShiftForDate(t, server, keyPrefix, "2026-06-18")
}

func createAuthSessionForTest(t *testing.T, server http.Handler, actorID string, pin string) string {
	t.Helper()

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/v1/auth/sessions", bytes.NewBufferString(fmt.Sprintf(`{
		"actorId": %q,
		"pin": %q,
		"storeId": "store-1"
	}`, actorID, pin)))

	server.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("create auth session status = %d, body = %s", response.Code, response.Body.String())
	}
	var accepted SessionAcceptedResponse
	if err := json.Unmarshal(response.Body.Bytes(), &accepted); err != nil {
		t.Fatalf("decode auth session: %v", err)
	}
	return accepted.Session.Token
}

func openStoreDayAndShiftForDate(t *testing.T, server http.Handler, keyPrefix string, businessDate string) {
	t.Helper()

	openDayForDate(t, server, keyPrefix, businessDate)

	shiftResponse := httptest.NewRecorder()
	shiftRequest := httptest.NewRequest(http.MethodPost, "/v1/shifts", bytes.NewBufferString(`{
		"storeId": "store-1",
		"terminalId": "pos-1",
		"cashierId": "cashier-1",
		"drawerId": "drawer-1",
		"sourceSafeId": "safe-1",
		"openingCashMinor": 100000
	}`))
	shiftRequest.Header.Set("Idempotency-Key", keyPrefix+"-shift-open-1")

	server.ServeHTTP(shiftResponse, shiftRequest)

	if shiftResponse.Code != http.StatusAccepted {
		t.Fatalf("open shift status = %d, body = %s", shiftResponse.Code, shiftResponse.Body.String())
	}
}

func openDay(t *testing.T, server http.Handler, keyPrefix string) {
	t.Helper()
	openDayForDate(t, server, keyPrefix, "2026-06-18")
}

func openDayForDate(t *testing.T, server http.Handler, keyPrefix string, businessDate string) {
	t.Helper()

	dayResponse := httptest.NewRecorder()
	dayRequest := httptest.NewRequest(http.MethodPost, "/v1/operational-days", bytes.NewBufferString(fmt.Sprintf(`{
		"storeId": "store-1",
		"businessDate": %q,
		"openedById": "senior-1"
	}`, businessDate)))
	dayRequest.Header.Set("Idempotency-Key", keyPrefix+"-oday-open-1")

	server.ServeHTTP(dayResponse, dayRequest)

	if dayResponse.Code != http.StatusAccepted {
		t.Fatalf("open operational day status = %d, body = %s", dayResponse.Code, dayResponse.Body.String())
	}
}

func hasOperationalDayBlocker(blockers []OperationalDayBlocker, code string) bool {
	for _, blocker := range blockers {
		if blocker.Code == code {
			return true
		}
	}
	return false
}
