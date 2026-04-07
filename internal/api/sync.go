package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// syncPath is the path for the sync endpoint.
const syncPath = "/sync?v=3"

type syncRequest struct {
	SyncClientID   string           `json:"syncClientId"`
	LastServerSync float64          `json:"last_server_sync,omitempty"`
	ClientChanges  []map[string]any `json:"clientChanges,omitempty"`
	SyncSession    string           `json:"syncSession,omitempty"`
	Token          string           `json:"token,omitempty"`
}

// SyncResponse represents the response from POST /api/sync?v=3.
type SyncResponse struct {
	SyncStartTime  float64 `json:"syncStartTime"`
	SyncSession    string  `json:"syncSession"`
	TotalChanges   int     `json:"totalServerChanges"`
	LastClientSync float64 `json:"lastClientSync"`
}

// SyncPullResponse represents the full response including server changes.
type SyncPullResponse struct {
	SyncStartTime float64            `json:"syncStartTime"`
	SyncSession   string             `json:"syncSession"`
	TotalChanges  int                `json:"totalServerChanges"`
	ServerChanges []SyncServerChange `json:"serverChanges"`
	Token         string             `json:"token"`
}

// SyncServerChange represents a single change record from the server.
type SyncServerChange struct {
	Collection string         `json:"mcollection"`
	Action     string         `json:"action"`
	ID         string         `json:"id"`
	Attachment AttachmentInfo `json:"data"`
}

// fetchSyncStartTime retrieves the current server sync timestamp.
func (c *Client) fetchSyncStartTime() (float64, error) {
	resp, err := c.doSyncRequest(syncRequest{
		SyncClientID: "paperpile",
	})
	if err != nil {
		return 0, fmt.Errorf("failed to fetch sync start time: %w", err)
	}
	return resp.SyncStartTime, nil
}

// pushSyncChanges sends local changes to the server via the Sync API.
// It first fetches the current server sync time, then sends changes with
// that timestamp so the server does not overwrite them with stale data.
func (c *Client) pushSyncChanges(changes []map[string]any) (*SyncResponse, error) {
	syncStartTime, err := c.fetchSyncStartTime()
	if err != nil {
		return nil, err
	}

	return c.doSyncRequest(syncRequest{
		SyncClientID:   "paperpile",
		LastServerSync: syncStartTime,
		ClientChanges:  changes,
	})
}

func (c *Client) doSyncRequest(reqBody syncRequest) (*SyncResponse, error) {
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sync request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+syncPath, bytes.NewReader(jsonBody))
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

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var syncResp SyncResponse
	if err := json.Unmarshal(respBody, &syncResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &syncResp, nil
}

// pullSyncData fetches all server data by syncing from timestamp 0,
// paginating with token until all changes are retrieved.
func (c *Client) pullSyncData() (*SyncPullResponse, error) {
	firstResp, err := c.fetchSyncPage("", "")
	if err != nil {
		return nil, err
	}

	allChanges := firstResp.ServerChanges
	syncSession := firstResp.SyncSession
	token := firstResp.Token

	for len(allChanges) < firstResp.TotalChanges && token != "" {
		pageResp, err := c.fetchSyncPage(syncSession, token)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch sync page: %w", err)
		}
		allChanges = append(allChanges, pageResp.ServerChanges...)
		token = pageResp.Token
	}

	firstResp.ServerChanges = allChanges
	return firstResp, nil
}

// fetchSyncPage fetches a single page of sync data from the server.
func (c *Client) fetchSyncPage(syncSession, token string) (*SyncPullResponse, error) {
	reqBody := syncRequest{
		SyncClientID:   "paperpile-cli",
		LastServerSync: 0,
		SyncSession:    syncSession,
		Token:          token,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sync request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+syncPath, bytes.NewReader(jsonBody))
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

	var pullResp SyncPullResponse
	if err := json.NewDecoder(resp.Body).Decode(&pullResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &pullResp, nil
}
