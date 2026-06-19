package hardwareagent_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/infra/hardwareagent"
)

func TestClientAuthorizeAndCapture(t *testing.T) {
	var mu sync.Mutex
	commands := map[string]hardwareagent.Command{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/devices/sim-payment-1/commands":
			var request struct {
				Type    string         `json:"type"`
				Payload map[string]any `json:"payload"`
			}
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			command := hardwareagent.Command{
				ID:       "cmd-" + request.Type,
				DeviceID: "sim-payment-1",
				Type:     request.Type,
				Payload:  request.Payload,
				Status:   hardwareagent.CommandStatusCompleted,
			}
			switch request.Type {
			case "authorize":
				command.Result = map[string]any{"authCode": "AUTH123", "rrn": "RRN456"}
			case "capture":
				command.Result = map[string]any{"rrn": "RRN456"}
			}
			mu.Lock()
			commands[command.ID] = command
			mu.Unlock()
			w.WriteHeader(http.StatusAccepted)
			_ = json.NewEncoder(w).Encode(map[string]any{"command": command})
		case r.Method == http.MethodGet:
			commandID := r.URL.Path[len("/v1/devices/sim-payment-1/commands/"):]
			mu.Lock()
			command := commands[commandID]
			mu.Unlock()
			_ = json.NewEncoder(w).Encode(command)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := hardwareagent.NewClient(server.URL, server.Client())
	ref, err := client.AuthorizeAndCapture(context.Background(), "sim-payment-1", 19999, "RUB", "receipt-1")
	if err != nil {
		t.Fatalf("authorize and capture: %v", err)
	}
	if ref != "RRN456" {
		t.Fatalf("provider ref = %q", ref)
	}
}

func TestClientPrintReceipt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost:
			command := hardwareagent.Command{
				ID:       "cmd-print",
				DeviceID: "sim-fiscal-1",
				Type:     "print_receipt",
				Status:   hardwareagent.CommandStatusCompleted,
				Result:   map[string]any{"fiscalSign": "SIM-FS-0001042"},
			}
			w.WriteHeader(http.StatusAccepted)
			_ = json.NewEncoder(w).Encode(map[string]any{"command": command})
		case r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(hardwareagent.Command{
				ID:       "cmd-print",
				DeviceID: "sim-fiscal-1",
				Type:     "print_receipt",
				Status:   hardwareagent.CommandStatusCompleted,
				Result:   map[string]any{"fiscalSign": "SIM-FS-0001042"},
			})
		}
	}))
	defer server.Close()

	client := hardwareagent.NewClient(server.URL, server.Client())
	sign, err := client.PrintReceipt(context.Background(), "sim-fiscal-1", 19999)
	if err != nil {
		t.Fatalf("print receipt: %v", err)
	}
	if sign != "SIM-FS-0001042" {
		t.Fatalf("fiscal sign = %q", sign)
	}
}

func TestClientHealthCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/healthz" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := hardwareagent.NewClient(server.URL, server.Client())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := client.HealthCheck(ctx); err != nil {
		t.Fatalf("health check: %v", err)
	}
}
