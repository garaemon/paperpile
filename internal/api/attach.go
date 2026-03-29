package api

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// S3UploadData represents the presigned POST data returned by the server.
type S3UploadData struct {
	URL    string            `json:"url"`
	Fields map[string]string `json:"fields"`
}

type fileUploadResponse struct {
	UploadData *S3UploadData `json:"upload_data"`
}

type attachmentPatchRequest struct {
	Updates map[string]string `json:"updates"`
}

// AttachFile attaches a PDF file to an existing library item.
// The flow discovered from the Paperpile Service Worker:
// 1. Create an attachment record via Sync API (Attachments collection insert)
// 2. POST /api/attachments/{id}/file to get S3 presigned POST data
// 3. POST the file to S3 using multipart form
// 4. PATCH /api/attachments/{id} to confirm upload with s3_md5
func (c *Client) AttachFile(itemID, filePath string) (string, error) {
	fileName := filepath.Base(filePath)

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	fileMD5 := computeMD5(fileData)
	fileSize := len(fileData)
	attachmentID := uuid.New().String()

	userInfo, err := c.FetchCurrentUser()
	if err != nil {
		return "", fmt.Errorf("failed to get user info: %w", err)
	}

	if err := c.createAttachmentRecord(attachmentID, itemID, userInfo.ID, fileName, fileMD5, fileSize); err != nil {
		return "", fmt.Errorf("failed to create attachment record: %w", err)
	}

	uploadData, err := c.getAttachmentUploadURL(attachmentID, fileMD5, fileName, fileSize)
	if err != nil {
		return "", fmt.Errorf("failed to get upload URL: %w", err)
	}

	if err := uploadToS3WithForm(uploadData, fileData); err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	if err := c.confirmAttachmentUpload(attachmentID, fileMD5); err != nil {
		return "", fmt.Errorf("failed to confirm upload: %w", err)
	}

	return attachmentID, nil
}

func (c *Client) createAttachmentRecord(attachmentID, pubID, ownerID, fileName, fileMD5 string, fileSize int) error {
	now := float64(time.Now().UnixMilli()) / 1000.0

	changes := []map[string]any{
		{
			"mcollection": "Attachments",
			"action":      "insert",
			"id":          attachmentID,
			"timestamp":   now,
			"data": map[string]any{
				"_id":         attachmentID,
				"pub_id":      pubID,
				"owner":       ownerID,
				"filename":    fileName,
				"mimeType":    "application/pdf",
				"filesize":    fileSize,
				"md5":         fileMD5,
				"article_pdf": 1,
				"created":     now,
				"updated":     now,
			},
		},
	}

	_, err := c.pushSyncChanges(changes)
	return err
}

func (c *Client) getAttachmentUploadURL(attachmentID, fileMD5, fileName string, fileSize int) (*S3UploadData, error) {
	url := fmt.Sprintf("%s/attachments/%s/file?client_md5=%s", baseURL, attachmentID, fileMD5)

	attachmentData := map[string]any{
		"attachment": map[string]any{
			"_id":      attachmentID,
			"md5":      fileMD5,
			"filename": fileName,
			"filesize": fileSize,
			"mimeType": "application/pdf",
		},
	}

	jsonBody, err := json.Marshal(attachmentData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	c.setCommonHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var uploadResp fileUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if uploadResp.UploadData == nil {
		return nil, fmt.Errorf("no upload_data in response")
	}

	return uploadResp.UploadData, nil
}

func uploadToS3WithForm(uploadData *S3UploadData, fileData []byte) error {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	for key, val := range uploadData.Fields {
		if err := writer.WriteField(key, val); err != nil {
			return fmt.Errorf("failed to write form field %s: %w", key, err)
		}
	}

	part, err := writer.CreateFormFile("file", "upload.pdf")
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := part.Write(fileData); err != nil {
		return fmt.Errorf("failed to write file data: %w", err)
	}
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, uploadData.URL, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("S3 returned status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *Client) confirmAttachmentUpload(attachmentID, fileMD5 string) error {
	url := fmt.Sprintf("%s/attachments/%s", baseURL, attachmentID)

	reqBody := attachmentPatchRequest{
		Updates: map[string]string{"s3_md5": fileMD5},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	c.setCommonHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func computeMD5(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}
