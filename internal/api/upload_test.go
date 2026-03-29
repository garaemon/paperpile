package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestInitiateImport_success(t *testing.T) {
	expectedTask := ImportTask{
		ID:     "task-1",
		Status: "pending",
		Subtasks: []ImportSubtask{
			{
				ID:        "sub-1",
				ParentID:  "task-1",
				Name:      "test.pdf",
				UploadURL: "https://s3.example.com/upload",
				Status:    "pending",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/import/files" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}

		var reqBody importRequest
		if err := json.Unmarshal(body, &reqBody); err != nil {
			t.Fatalf("failed to unmarshal body: %v", err)
		}

		if len(reqBody.Files) != 1 {
			t.Fatalf("expected 1 file, got %d", len(reqBody.Files))
		}
		if reqBody.Files[0].Names[0] != "test.pdf" {
			t.Errorf("file name = %q, want %q", reqBody.Files[0].Names[0], "test.pdf")
		}
		if reqBody.ImportDuplicates != true {
			t.Error("expected ImportDuplicates to be true")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedTask)
	}))
	defer server.Close()

	client := newTestClient(server)
	task, err := client.initiateImport("test.pdf", true)
	if err != nil {
		t.Fatalf("initiateImport() error: %v", err)
	}
	if task.ID != "task-1" {
		t.Errorf("task.ID = %q, want %q", task.ID, "task-1")
	}
	if len(task.Subtasks) != 1 {
		t.Fatalf("expected 1 subtask, got %d", len(task.Subtasks))
	}
	if task.Subtasks[0].UploadURL != "https://s3.example.com/upload" {
		t.Errorf("UploadURL = %q, want %q", task.Subtasks[0].UploadURL, "https://s3.example.com/upload")
	}
}

func TestInitiateImport_serverError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.initiateImport("test.pdf", false)
	if err == nil {
		t.Fatal("initiateImport() expected error for 400 response")
	}
}

func TestNotifyUploadComplete_success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tasks/task-1/subtasks" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPatch {
			t.Errorf("unexpected method: %s", r.Method)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}

		var reqBody subtaskPatch
		if err := json.Unmarshal(body, &reqBody); err != nil {
			t.Fatalf("failed to unmarshal body: %v", err)
		}

		if reqBody.Status != "uploaded" {
			t.Errorf("status = %q, want %q", reqBody.Status, "uploaded")
		}
		if len(reqBody.Subtasks) != 1 || reqBody.Subtasks[0] != "sub-1" {
			t.Errorf("subtasks = %v, want [sub-1]", reqBody.Subtasks)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)
	err := client.notifyUploadComplete("task-1", "sub-1")
	if err != nil {
		t.Fatalf("notifyUploadComplete() error: %v", err)
	}
}

func TestNotifyUploadComplete_serverError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
	}))
	defer server.Close()

	client := newTestClient(server)
	err := client.notifyUploadComplete("task-1", "sub-1")
	if err == nil {
		t.Fatal("notifyUploadComplete() expected error for 500 response")
	}
}

func TestUploadPDF_success(t *testing.T) {
	// Create a temporary PDF file
	tmpDir := t.TempDir()
	pdfPath := filepath.Join(tmpDir, "test.pdf")
	if err := os.WriteFile(pdfPath, []byte("%PDF-1.4 test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// S3 mock server
	s3Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("S3: unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer s3Server.Close()

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/import/files" && r.Method == http.MethodPost:
			callCount++
			task := ImportTask{
				ID:     "task-1",
				Status: "pending",
				Subtasks: []ImportSubtask{
					{
						ID:        "sub-1",
						ParentID:  "task-1",
						Name:      "test.pdf",
						UploadURL: s3Server.URL + "/upload",
						Status:    "pending",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(task)
		case r.URL.Path == "/tasks/task-1/subtasks" && r.Method == http.MethodPatch:
			callCount++
			w.WriteHeader(http.StatusOK)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := newTestClient(server)
	task, err := client.UploadPDF(pdfPath, false)
	if err != nil {
		t.Fatalf("UploadPDF() error: %v", err)
	}
	if task.ID != "task-1" {
		t.Errorf("task.ID = %q, want %q", task.ID, "task-1")
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestUploadPDF_noSubtasks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		task := ImportTask{
			ID:       "task-1",
			Status:   "pending",
			Subtasks: []ImportSubtask{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	pdfPath := filepath.Join(tmpDir, "test.pdf")
	if err := os.WriteFile(pdfPath, []byte("%PDF-1.4"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	client := newTestClient(server)
	_, err := client.UploadPDF(pdfPath, false)
	if err == nil {
		t.Fatal("UploadPDF() expected error when no subtasks returned")
	}
}

func TestUploadToS3_success(t *testing.T) {
	s3Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("unexpected method: %s", r.Method)
		}
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/pdf" {
			t.Errorf("Content-Type = %q, want %q", contentType, "application/pdf")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer s3Server.Close()

	tmpDir := t.TempDir()
	pdfPath := filepath.Join(tmpDir, "test.pdf")
	if err := os.WriteFile(pdfPath, []byte("%PDF-1.4 content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	err := uploadToS3(s3Server.URL+"/upload", pdfPath)
	if err != nil {
		t.Fatalf("uploadToS3() error: %v", err)
	}
}

func TestUploadToS3_fileNotFound(t *testing.T) {
	err := uploadToS3("http://example.com", "/nonexistent/file.pdf")
	if err == nil {
		t.Fatal("uploadToS3() expected error for nonexistent file")
	}
}

func TestUploadToS3_serverError(t *testing.T) {
	s3Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("access denied"))
	}))
	defer s3Server.Close()

	tmpDir := t.TempDir()
	pdfPath := filepath.Join(tmpDir, "test.pdf")
	if err := os.WriteFile(pdfPath, []byte("%PDF-1.4"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	err := uploadToS3(s3Server.URL+"/upload", pdfPath)
	if err == nil {
		t.Fatal("uploadToS3() expected error for 403 response")
	}
}
