package hardwareagent

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type CommandStatus string

const (
	CommandStatusAccepted  CommandStatus = "accepted"
	CommandStatusRunning   CommandStatus = "running"
	CommandStatusCompleted CommandStatus = "completed"
	CommandStatusFailed    CommandStatus = "failed"
)

type Command struct {
	ID          string         `json:"id"`
	DeviceID    string         `json:"deviceId"`
	Type        string         `json:"type"`
	Payload     map[string]any `json:"payload,omitempty"`
	Status      CommandStatus  `json:"status"`
	Result      map[string]any `json:"result,omitempty"`
	Error       string         `json:"error,omitempty"`
	CompletedAt *time.Time     `json:"completedAt,omitempty"`
}

type commandAcceptedResponse struct {
	Command Command `json:"command"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		baseURL:    strings.TrimRight(normalizeBaseURL(baseURL), "/"),
		httpClient: httpClient,
	}
}

func DefaultBaseURL() string {
	return "http://127.0.0.1:8083"
}

func normalizeBaseURL(baseURL string) string {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return DefaultBaseURL()
	}
	if strings.HasPrefix(baseURL, "http://") || strings.HasPrefix(baseURL, "https://") {
		return baseURL
	}
	return "http://" + baseURL
}

func (c *Client) HealthCheck(ctx context.Context) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/healthz", nil)
	if err != nil {
		return fmt.Errorf("create hardware agent health request: %w", err)
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("hardware agent health check: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("hardware agent health status %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

func (c *Client) SendCommand(ctx context.Context, deviceID, commandType string, payload map[string]any) (Command, error) {
	body, err := json.Marshal(map[string]any{
		"type":    commandType,
		"payload": payload,
	})
	if err != nil {
		return Command{}, fmt.Errorf("encode device command: %w", err)
	}

	endpoint := fmt.Sprintf("%s/v1/devices/%s/commands", c.baseURL, deviceID)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return Command{}, fmt.Errorf("create device command request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", newIdempotencyKey("ha"))

	response, err := c.httpClient.Do(request)
	if err != nil {
		return Command{}, fmt.Errorf("send device command: %w", err)
	}
	defer func() { _ = response.Body.Close() }()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return Command{}, fmt.Errorf("read device command response: %w", err)
	}
	if response.StatusCode != http.StatusAccepted {
		return Command{}, fmt.Errorf("device command status %d: %s", response.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	var accepted commandAcceptedResponse
	if err := json.Unmarshal(responseBody, &accepted); err != nil {
		return Command{}, fmt.Errorf("decode device command response: %w", err)
	}
	return accepted.Command, nil
}

func (c *Client) GetCommand(ctx context.Context, deviceID, commandID string) (Command, error) {
	endpoint := fmt.Sprintf("%s/v1/devices/%s/commands/%s", c.baseURL, deviceID, commandID)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Command{}, fmt.Errorf("create get command request: %w", err)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return Command{}, fmt.Errorf("get device command: %w", err)
	}
	defer func() { _ = response.Body.Close() }()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return Command{}, fmt.Errorf("read get command response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return Command{}, fmt.Errorf("get command status %d: %s", response.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	var command Command
	if err := json.Unmarshal(responseBody, &command); err != nil {
		return Command{}, fmt.Errorf("decode get command response: %w", err)
	}
	return command, nil
}

func (c *Client) WaitCommand(ctx context.Context, deviceID, commandID string, timeout time.Duration) (Command, error) {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	deadline := time.Now().Add(timeout)
	for {
		command, err := c.GetCommand(ctx, deviceID, commandID)
		if err != nil {
			return Command{}, err
		}
		switch command.Status {
		case CommandStatusCompleted:
			return command, nil
		case CommandStatusFailed:
			if command.Error != "" {
				return Command{}, fmt.Errorf("device command failed: %s", command.Error)
			}
			return Command{}, fmt.Errorf("device command failed")
		}
		if time.Now().After(deadline) {
			return Command{}, fmt.Errorf("device command timed out")
		}
		select {
		case <-ctx.Done():
			return Command{}, ctx.Err()
		case <-time.After(25 * time.Millisecond):
		}
	}
}

func (c *Client) AuthorizeAndCapture(ctx context.Context, deviceID string, amountMinor int64, currency, reference string) (string, error) {
	if currency == "" {
		currency = "RUB"
	}
	authorize, err := c.SendCommand(ctx, deviceID, "authorize", map[string]any{
		"amountMinor": amountMinor,
		"currency":    currency,
		"reference":   reference,
	})
	if err != nil {
		return "", err
	}
	authorize, err = c.WaitCommand(ctx, deviceID, authorize.ID, 5*time.Second)
	if err != nil {
		return "", err
	}

	capture, err := c.SendCommand(ctx, deviceID, "capture", map[string]any{
		"reference": reference,
	})
	if err != nil {
		return "", err
	}
	capture, err = c.WaitCommand(ctx, deviceID, capture.ID, 5*time.Second)
	if err != nil {
		return "", err
	}

	providerRef := reference
	if capture.Result != nil {
		if rrn, ok := capture.Result["rrn"].(string); ok && rrn != "" {
			providerRef = rrn
		} else if authCode, ok := authorize.Result["authCode"].(string); ok && authCode != "" {
			providerRef = authCode
		}
	}
	return providerRef, nil
}

func (c *Client) CancelCardPayment(ctx context.Context, deviceID, reference string) error {
	command, err := c.SendCommand(ctx, deviceID, "cancel", map[string]any{
		"reference": reference,
	})
	if err != nil {
		return err
	}
	command, err = c.WaitCommand(ctx, deviceID, command.ID, 5*time.Second)
	if err != nil {
		return err
	}
	if command.Result != nil {
		if status, ok := command.Result["status"].(string); ok && status == "cancelled" {
			return nil
		}
	}
	return fmt.Errorf("card payment cancel did not complete")
}

func (c *Client) RefundCardPayment(ctx context.Context, deviceID, reference string, amountMinor int64) error {
	command, err := c.SendCommand(ctx, deviceID, "refund", map[string]any{
		"reference":   reference,
		"amountMinor": amountMinor,
	})
	if err != nil {
		return err
	}
	command, err = c.WaitCommand(ctx, deviceID, command.ID, 5*time.Second)
	if err != nil {
		return err
	}
	if command.Result != nil {
		if status, ok := command.Result["status"].(string); ok && status == "refunded" {
			return nil
		}
	}
	return fmt.Errorf("card payment refund did not complete")
}

func (c *Client) PrintReceipt(ctx context.Context, deviceID string, totalMinor int64) (string, error) {
	command, err := c.SendCommand(ctx, deviceID, "print_receipt", map[string]any{
		"totalMinor": totalMinor,
	})
	if err != nil {
		return "", err
	}
	command, err = c.WaitCommand(ctx, deviceID, command.ID, 5*time.Second)
	if err != nil {
		return "", err
	}
	if command.Result == nil {
		return "", fmt.Errorf("print receipt returned no result")
	}
	fiscalSign, _ := command.Result["fiscalSign"].(string)
	if fiscalSign == "" {
		return "", fmt.Errorf("print receipt returned empty fiscal sign")
	}
	return fiscalSign, nil
}

func newIdempotencyKey(prefix string) string {
	var randomBytes [12]byte
	if _, err := rand.Read(randomBytes[:]); err != nil {
		panic(fmt.Sprintf("generate idempotency key: %v", err))
	}
	return prefix + "_" + hex.EncodeToString(randomBytes[:])
}
