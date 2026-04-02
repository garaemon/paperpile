package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Author represents a paper author.
type Author struct {
	First     string `json:"first"`
	Last      string `json:"last"`
	Formatted string `json:"formatted"`
}

// LibraryItem represents a single item in the Paperpile library.
type LibraryItem struct {
	ID      string   `json:"_id"`
	Title   string   `json:"title"`
	Author  []Author `json:"author"`
	Year    string   `json:"year"`
	Journal string   `json:"journal"`
	Pubtype string   `json:"pubtype"`
	Citekey string   `json:"citekey"`
	Trashed  int      `json:"trashed"`
	Notes    string   `json:"note"`
	LabelIDs []string `json:"labels"`
}

// FormatFirstAuthor returns a readable first author string.
func (item *LibraryItem) FormatFirstAuthor() string {
	if len(item.Author) == 0 {
		return ""
	}
	a := item.Author[0]
	if a.Last != "" {
		return a.Last
	}
	return a.Formatted
}

// FetchLibrary retrieves all items from the user's Paperpile library.
func (c *Client) FetchLibrary() ([]LibraryItem, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/library", nil)
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

	var items []LibraryItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return items, nil
}
