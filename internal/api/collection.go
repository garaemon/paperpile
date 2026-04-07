package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Collection represents a Paperpile collection (label or folder).
type Collection struct {
	ID             string `json:"_id"`
	Name           string `json:"cName"`
	Parent         string `json:"cParent"`
	Count          int    `json:"cCount"`
	Hidden         int    `json:"cHidden"`
	SortOrder      int    `json:"cSortOrder"`
	Style          string `json:"cStyle"`
	CollectionType string `json:"collectionType"`
	Trashed        int    `json:"trashed"`
}

// IsLabel returns true if the collection is a label.
func (c *Collection) IsLabel() bool {
	return c.CollectionType == "label"
}

// FetchCollections retrieves all collections (labels and folders) from Paperpile.
func (c *Client) FetchCollections() ([]Collection, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/collections", nil)
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

	var collections []Collection
	if err := json.NewDecoder(resp.Body).Decode(&collections); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return collections, nil
}

// FetchLabels retrieves only label-type collections.
func (c *Client) FetchLabels() ([]Collection, error) {
	collections, err := c.FetchCollections()
	if err != nil {
		return nil, err
	}

	var labels []Collection
	for _, col := range collections {
		if col.IsLabel() && col.Trashed == 0 {
			labels = append(labels, col)
		}
	}
	return labels, nil
}
