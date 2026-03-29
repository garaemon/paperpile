package api

import (
	"fmt"
	"time"
)

// TrashItem moves a library item to the trash via the Sync API.
func (c *Client) TrashItem(itemID string) error {
	now := time.Now().Unix()

	changes := []map[string]any{
		{
			"mcollection": "Library",
			"action":      "update",
			"id":          itemID,
			"timestamp":   float64(now),
			"fields":      []string{"trashed", "updated"},
			"data":        map[string]any{"trashed": 1, "updated": float64(now)},
		},
	}

	_, err := c.pushSyncChanges(changes)
	if err != nil {
		return fmt.Errorf("failed to sync trash change: %w", err)
	}
	return nil
}
