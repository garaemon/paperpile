package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const baseURL = "https://api.paperpile.com/api"

// Client is the Paperpile API client.
type Client struct {
	session    string
	httpClient *http.Client
}

// UserInfo represents the response from /api/users/me.
type UserInfo struct {
	ID          string `json:"_id"`
	GoogleName  string `json:"google_name"`
	GoogleEmail string `json:"google_email"`
}

// NewClient creates a new API client with the given plack_session.
func NewClient(session string) *Client {
	return &Client{
		session:    session,
		httpClient: &http.Client{},
	}
}

// FetchCurrentUser calls /api/users/me to verify the session and return user info.
func (c *Client) FetchCurrentUser() (*UserInfo, error) {
	req, err := http.NewRequest(http.MethodGet, baseURL+"/users/me", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
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

	var user UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &user, nil
}

func (c *Client) setCommonHeaders(req *http.Request) {
	req.Header.Set("Origin", "https://app.paperpile.com")
	req.Header.Set("Referer", "https://app.paperpile.com/")
	req.AddCookie(&http.Cookie{Name: "plack_session", Value: c.session})
}
