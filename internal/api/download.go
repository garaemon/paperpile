package api

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
)

// DownloadResult holds the result of a file download.
type DownloadResult struct {
	Data     []byte
	Filename string
}

type fileResponse struct {
	URL string `json:"url"`
}

// DownloadFile downloads the PDF attachment for the given item.
// It first fetches the presigned S3 URL from the API, then downloads the file content.
func (c *Client) DownloadFile(itemID string) (*DownloadResult, error) {
	presignedURL, err := c.fetchPresignedURL(itemID)
	if err != nil {
		return nil, err
	}

	return c.downloadFromS3(itemID, presignedURL)
}

func (c *Client) fetchPresignedURL(itemID string) (string, error) {
	endpoint := fmt.Sprintf("%s/items/%s/file", c.baseURL, itemID)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	c.setCommonHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var fileResp fileResponse
	if err := json.NewDecoder(resp.Body).Decode(&fileResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if fileResp.URL == "" {
		return "", fmt.Errorf("no download URL returned for item %s", itemID)
	}

	return fileResp.URL, nil
}

func (c *Client) downloadFromS3(itemID, s3URL string) (*DownloadResult, error) {
	resp, err := c.httpClient.Get(s3URL)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("S3 download returned status %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	filename := extractFilename(resp.Header.Get("Content-Disposition"), itemID)

	return &DownloadResult{Data: data, Filename: filename}, nil
}

func extractFilename(contentDisposition, itemID string) string {
	if contentDisposition != "" {
		_, params, err := mime.ParseMediaType(contentDisposition)
		if err == nil {
			if name, ok := params["filename"]; ok && name != "" {
				return name
			}
		}
	}
	return itemID + ".pdf"
}
