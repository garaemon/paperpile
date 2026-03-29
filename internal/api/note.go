package api

import (
	"fmt"
	"time"
)

// UpdateNote sets the note field on a library item via the Sync API.
func (c *Client) UpdateNote(itemID, note string) error {
	now := time.Now().Unix()

	changes := []map[string]any{
		{
			"mcollection": "Library",
			"action":      "update",
			"id":          itemID,
			"timestamp":   float64(now),
			"fields":      []string{"note", "updated"},
			"data":        map[string]any{"note": note, "updated": float64(now)},
		},
	}

	_, err := c.pushSyncChanges(changes)
	if err != nil {
		return fmt.Errorf("failed to sync note change: %w", err)
	}
	return nil
}

// GetNote fetches a single item's note from the library.
func (c *Client) GetNote(itemID string) (string, error) {
	items, err := c.FetchLibrary()
	if err != nil {
		return "", fmt.Errorf("failed to fetch library: %w", err)
	}

	for _, item := range items {
		if item.ID == itemID {
			return item.Notes, nil
		}
	}
	return "", fmt.Errorf("item %s not found", itemID)
}
