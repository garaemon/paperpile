package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const syncURL = baseURL + "/sync?v=3"

type syncRequest struct {
	SyncClientID   string           `json:"syncClientId"`
	LastServerSync float64          `json:"last_server_sync,omitempty"`
	ClientChanges  []map[string]any `json:"clientChanges,omitempty"`
}

// SyncResponse represents the response from POST /api/sync?v=3.
type SyncResponse struct {
	SyncStartTime    float64 `json:"syncStartTime"`
	SyncSession      string  `json:"syncSession"`
	TotalChanges     int     `json:"totalServerChanges"`
	LastClientSync   float64 `json:"lastClientSync"`
}

// pushSyncChanges sends local changes to the server via the Sync API.
func (c *Client) pushSyncChanges(changes []map[string]any) (*SyncResponse, error) {
	reqBody := syncRequest{
		SyncClientID:   "paperpile-cli",
		LastServerSync: float64(time.Now().Unix()),
		ClientChanges:  changes,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sync request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, syncURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	c.setCommonHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var syncResp SyncResponse
	if err := json.NewDecoder(resp.Body).Decode(&syncResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &syncResp, nil
}
