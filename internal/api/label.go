package api

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// GetItemLabels fetches the label IDs for a specific library item.
func (c *Client) GetItemLabels(itemID string) ([]string, error) {
	items, err := c.FetchLibrary()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch library: %w", err)
	}

	for _, item := range items {
		if item.ID == itemID {
			return item.LabelIDs, nil
		}
	}
	return nil, fmt.Errorf("item %s not found", itemID)
}

// GetItemLabelNames fetches the label names for a specific library item.
func (c *Client) GetItemLabelNames(itemID string) ([]string, error) {
	labelIDs, err := c.GetItemLabels(itemID)
	if err != nil {
		return nil, err
	}

	if len(labelIDs) == 0 {
		return nil, nil
	}

	labels, err := c.FetchLabels()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch labels: %w", err)
	}

	labelMap := make(map[string]string, len(labels))
	for _, label := range labels {
		labelMap[label.ID] = label.Name
	}

	var names []string
	for _, id := range labelIDs {
		if name, ok := labelMap[id]; ok {
			names = append(names, name)
		} else {
			names = append(names, id)
		}
	}
	return names, nil
}

// ResolveLabelName finds a label ID by name. Returns the label ID or an error.
func (c *Client) ResolveLabelName(name string) (string, error) {
	labels, err := c.FetchLabels()
	if err != nil {
		return "", err
	}

	for _, label := range labels {
		if label.Name == name {
			return label.ID, nil
		}
	}
	return "", fmt.Errorf("label %q not found", name)
}

// DeleteLabel deletes a label by marking it as trashed via the Sync API.
func (c *Client) DeleteLabel(labelName string) error {
	labelID, err := c.ResolveLabelName(labelName)
	if err != nil {
		return err
	}

	now := float64(time.Now().UnixMilli()) / 1000.0

	changes := []map[string]any{
		{
			"mcollection": "Collections",
			"action":      "update",
			"id":          labelID,
			"timestamp":   now,
			"fields":      []string{"trashed", "updated"},
			"data":        map[string]any{"trashed": 1, "updated": now},
		},
	}

	_, err = c.pushSyncChanges(changes)
	if err != nil {
		return fmt.Errorf("failed to delete label: %w", err)
	}
	return nil
}

// CreateLabel creates a new label via the Sync API.
func (c *Client) CreateLabel(name string) (string, error) {
	now := float64(time.Now().UnixMilli()) / 1000.0
	labelID := uuid.New().String()

	changes := []map[string]any{
		{
			"mcollection": "Collections",
			"action":      "insert",
			"id":          labelID,
			"timestamp":   now,
			"data": map[string]any{
				"_id":            labelID,
				"cName":          name,
				"cParent":        "ROOT",
				"cSortOrder":     -1,
				"cStyle":         "0",
				"cHidden":        0,
				"collectionType": "label",
				"created":        now,
				"updated":        now,
			},
		},
	}

	_, err := c.pushSyncChanges(changes)
	if err != nil {
		return "", fmt.Errorf("failed to create label: %w", err)
	}
	return labelID, nil
}

// AssignLabel assigns a label to a library item by name.
// It resolves the label name to an ID, fetches the item's current labels,
// appends the new label ID, and pushes the update via the Sync API.
func (c *Client) AssignLabel(itemID, labelName string) error {
	labelID, err := c.ResolveLabelName(labelName)
	if err != nil {
		return err
	}

	currentLabelIDs, err := c.GetItemLabels(itemID)
	if err != nil {
		return err
	}

	for _, id := range currentLabelIDs {
		if id == labelID {
			return fmt.Errorf("label %q is already assigned to item %s", labelName, itemID)
		}
	}

	newLabelIDs := append(currentLabelIDs, labelID)
	now := float64(time.Now().UnixMilli()) / 1000.0

	changes := []map[string]any{
		{
			"mcollection": "Library",
			"action":      "update",
			"id":          itemID,
			"timestamp":   now,
			"fields":      []string{"labelIds", "updated"},
			"data":        map[string]any{"labelIds": newLabelIDs, "updated": now},
		},
	}

	_, err = c.pushSyncChanges(changes)
	if err != nil {
		return fmt.Errorf("failed to assign label: %w", err)
	}
	return nil
}
