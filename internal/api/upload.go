package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// ImportTask represents the response from POST /api/import/files.
type ImportTask struct {
	ID       string          `json:"_id"`
	Status   string          `json:"status"`
	Subtasks []ImportSubtask `json:"subtasks"`
}

// ImportSubtask represents a single file subtask within an import task.
type ImportSubtask struct {
	ID        string `json:"_id"`
	ParentID  string `json:"parentId"`
	Name      string `json:"name"`
	UploadURL string `json:"uploadUrl"`
	Status    string `json:"status"`
}

type importRequest struct {
	Files                 []importFile `json:"files"`
	IsPartialImport       bool         `json:"isPartialImport"`
	Collections           []any        `json:"collections"`
	KeepFolderOrganization bool        `json:"keepFolderOrganization"`
	PreserveCitationKey   bool         `json:"preserveCitationKey"`
	ImportDuplicates      bool         `json:"importDuplicates"`
}

type importFile struct {
	Names []string `json:"names"`
	Type  string   `json:"type"`
}

type subtaskPatch struct {
	Subtasks []string `json:"subtasks"`
	Status   string   `json:"status"`
}

// UploadPDF uploads a PDF file to Paperpile in 3 steps:
// 1. POST /api/import/files to get a presigned S3 URL
// 2. PUT the PDF binary to S3
// 3. PATCH /api/tasks/{taskId}/subtasks to notify completion
// UploadPDF uploads a PDF file to Paperpile in 3 steps:
// 1. POST /api/import/files to get a presigned S3 URL
// 2. PUT the PDF binary to S3
// 3. PATCH /api/tasks/{taskId}/subtasks to notify completion
func (c *Client) UploadPDF(filePath string, importDuplicates bool) (*ImportTask, error) {
	fileName := filepath.Base(filePath)

	task, err := c.initiateImport(fileName, importDuplicates)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate import: %w", err)
	}

	if len(task.Subtasks) == 0 {
		return nil, fmt.Errorf("no subtasks returned from import")
	}
	subtask := task.Subtasks[0]

	if err := uploadToS3(subtask.UploadURL, filePath); err != nil {
		return nil, fmt.Errorf("failed to upload to S3: %w", err)
	}

	if err := c.notifyUploadComplete(task.ID, subtask.ID); err != nil {
		return nil, fmt.Errorf("failed to notify upload completion: %w", err)
	}

	return task, nil
}

func (c *Client) initiateImport(fileName string, importDuplicates bool) (*ImportTask, error) {
	reqBody := importRequest{
		Files: []importFile{
			{Names: []string{fileName}, Type: "file_upload"},
		},
		Collections:      make([]any, 0),
		ImportDuplicates: importDuplicates,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, baseURL+"/import/files", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
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

	var task ImportTask
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return nil, err
	}
	return &task, nil
}

func uploadToS3(uploadURL, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, uploadURL, f)
	if err != nil {
		return err
	}
	req.ContentLength = stat.Size()
	req.Header.Set("Content-Type", "application/pdf")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("S3 returned status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *Client) notifyUploadComplete(taskID, subtaskID string) error {
	reqBody := subtaskPatch{
		Subtasks: []string{subtaskID},
		Status:   "uploaded",
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/tasks/%s/subtasks", baseURL, taskID)
	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	c.setCommonHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
