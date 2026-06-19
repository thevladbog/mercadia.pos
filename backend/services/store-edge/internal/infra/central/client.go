package central

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const syncAPIKeyHeader = "X-Sync-Api-Key"

type CatalogProduct struct {
	ID             string    `json:"id"`
	StoreID        string    `json:"storeId"`
	Name           string    `json:"name"`
	Barcodes       []string  `json:"barcodes"`
	UnitPriceMinor int64     `json:"unitPriceMinor"`
	TaxCategoryID  string    `json:"taxCategoryId"`
	Active         bool      `json:"active"`
	Version        int64     `json:"version"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type catalogDeltaResponse struct {
	Since    time.Time        `json:"since"`
	Products []CatalogProduct `json:"products"`
}

type Client struct {
	baseURL    string
	syncAPIKey string
	httpClient *http.Client
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	return NewClientWithSyncAPIKey(baseURL, "", httpClient)
}

func NewClientWithSyncAPIKey(baseURL, syncAPIKey string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		syncAPIKey: strings.TrimSpace(syncAPIKey),
		httpClient: httpClient,
	}
}

func DefaultBaseURL() string {
	return "http://127.0.0.1:8082"
}

func (c *Client) CatalogDelta(ctx context.Context, storeID string, since time.Time) ([]CatalogProduct, time.Time, error) {
	if storeID == "" {
		return nil, time.Time{}, fmt.Errorf("store id is required")
	}
	if since.IsZero() {
		since = time.Unix(0, 0).UTC()
	}

	endpoint := fmt.Sprintf("%s/v1/stores/%s/catalog/delta?since=%s",
		c.baseURL,
		url.PathEscape(storeID),
		url.QueryEscape(since.UTC().Format(time.RFC3339)),
	)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("create catalog delta request: %w", err)
	}
	if c.syncAPIKey != "" {
		request.Header.Set(syncAPIKeyHeader, c.syncAPIKey)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("fetch catalog delta: %w", err)
	}
	defer func() { _ = response.Body.Close() }()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("read catalog delta response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, time.Time{}, fmt.Errorf("catalog delta status %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload catalogDeltaResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, time.Time{}, fmt.Errorf("decode catalog delta: %w", err)
	}

	return payload.Products, payload.Since, nil
}
