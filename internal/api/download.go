package api

import (
	"fmt"
	"io"
	"net/http"
)

// AttachmentInfo represents a file attachment linked to a library item.
type AttachmentInfo struct {
	ID       string `json:"_id"`
	PubID    string `json:"pub_id"`
	Filename string `json:"filename"`
	MimeType string `json:"mimeType"`
	MD5      string `json:"md5"`
	Filesize int    `json:"filesize"`
}

// DownloadResult holds the result of a file download.
type DownloadResult struct {
	Data     []byte
	Filename string
}

// FetchAttachments retrieves all attachment records from the server via the Sync API.
func (c *Client) FetchAttachments() ([]AttachmentInfo, error) {
	syncResp, err := c.pullSyncData()
	if err != nil {
		return nil, fmt.Errorf("failed to pull sync data: %w", err)
	}

	var attachments []AttachmentInfo
	for _, change := range syncResp.ServerChanges {
		if change.Collection != "Attachments" || change.Action != "insert" {
			continue
		}
		attachments = append(attachments, change.Attachment)
	}
	return attachments, nil
}

// FindAttachmentsForItem returns all attachments linked to the given library item.
func (c *Client) FindAttachmentsForItem(itemID string) ([]AttachmentInfo, error) {
	allAttachments, err := c.FetchAttachments()
	if err != nil {
		return nil, err
	}

	var matched []AttachmentInfo
	for _, a := range allAttachments {
		if a.PubID == itemID {
			matched = append(matched, a)
		}
	}
	if len(matched) == 0 {
		return nil, fmt.Errorf("no attachments found for item %s", itemID)
	}
	return matched, nil
}

// DownloadAttachment downloads a single attachment by its ID.
// The API returns a 307 redirect to a presigned S3 URL.
func (c *Client) DownloadAttachment(attachment AttachmentInfo) (*DownloadResult, error) {
	endpoint := fmt.Sprintf("%s/attachments/%s/file?type=latest", c.baseURL, attachment.ID)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	c.setCommonHeaders(req)

	// Follow the 307 redirect to S3 automatically.
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download returned status %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	filename := attachment.Filename
	if filename == "" {
		filename = attachment.ID + ".bin"
	}

	return &DownloadResult{Data: data, Filename: filename}, nil
}
