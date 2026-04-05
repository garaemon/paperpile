package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetAttachmentUploadURL_success(t *testing.T) {
	expectedUploadData := &S3UploadData{
		URL: "https://s3.example.com/bucket",
		Fields: map[string]string{
			"key":    "attachments/abc",
			"policy": "base64policy",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if !strings.HasPrefix(r.URL.Path, "/attachments/att-1/file") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("client_md5") != "abc123" {
			t.Errorf("client_md5 = %q, want %q", r.URL.Query().Get("client_md5"), "abc123")
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}

		var reqBody map[string]any
		if err := json.Unmarshal(body, &reqBody); err != nil {
			t.Fatalf("failed to unmarshal body: %v", err)
		}

		attachment := reqBody["attachment"].(map[string]any)
		if attachment["_id"] != "att-1" {
			t.Errorf("_id = %v, want %q", attachment["_id"], "att-1")
		}

		resp := fileUploadResponse{UploadData: expectedUploadData}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(server)
	uploadData, err := client.getAttachmentUploadURL("att-1", "abc123", "test.pdf", 1024)
	if err != nil {
		t.Fatalf("getAttachmentUploadURL() error: %v", err)
	}
	if uploadData.URL != expectedUploadData.URL {
		t.Errorf("URL = %q, want %q", uploadData.URL, expectedUploadData.URL)
	}
	if uploadData.Fields["key"] != "attachments/abc" {
		t.Errorf("Fields[key] = %q, want %q", uploadData.Fields["key"], "attachments/abc")
	}
}

func TestGetAttachmentUploadURL_serverError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.getAttachmentUploadURL("att-1", "abc123", "test.pdf", 1024)
	if err == nil {
		t.Fatal("getAttachmentUploadURL() expected error for 500 response")
	}
}

func TestGetAttachmentUploadURL_noUploadData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := fileUploadResponse{UploadData: nil}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.getAttachmentUploadURL("att-1", "abc123", "test.pdf", 1024)
	if err == nil {
		t.Fatal("getAttachmentUploadURL() expected error when upload_data is nil")
	}
}

func TestConfirmAttachmentUpload_success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/attachments/att-1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}

		var reqBody attachmentPatchRequest
		if err := json.Unmarshal(body, &reqBody); err != nil {
			t.Fatalf("failed to unmarshal body: %v", err)
		}

		if reqBody.Updates["s3_md5"] != "abc123" {
			t.Errorf("s3_md5 = %q, want %q", reqBody.Updates["s3_md5"], "abc123")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)
	err := client.confirmAttachmentUpload("att-1", "abc123")
	if err != nil {
		t.Fatalf("confirmAttachmentUpload() error: %v", err)
	}
}

func TestConfirmAttachmentUpload_serverError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("forbidden"))
	}))
	defer server.Close()

	client := newTestClient(server)
	err := client.confirmAttachmentUpload("att-1", "abc123")
	if err == nil {
		t.Fatal("confirmAttachmentUpload() expected error for 403 response")
	}
}

func TestCreateAttachmentRecord_success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}

		var reqBody map[string]any
		if err := json.Unmarshal(body, &reqBody); err != nil {
			t.Fatalf("failed to unmarshal body: %v", err)
		}

		resp := SyncResponse{SyncStartTime: 1234567890.0, SyncSession: "s-1"}
		w.Header().Set("Content-Type", "application/json")

		changes, ok := reqBody["clientChanges"].([]any)
		if !ok || len(changes) == 0 {
			json.NewEncoder(w).Encode(resp)
			return
		}

		change := changes[0].(map[string]any)
		if change["mcollection"] != "Attachments" {
			t.Errorf("mcollection = %v, want %q", change["mcollection"], "Attachments")
		}
		if change["action"] != "insert" {
			t.Errorf("action = %v, want %q", change["action"], "insert")
		}

		data := change["data"].(map[string]any)
		if data["pub_id"] != "pub-1" {
			t.Errorf("pub_id = %v, want %q", data["pub_id"], "pub-1")
		}
		if data["filename"] != "paper.pdf" {
			t.Errorf("filename = %v, want %q", data["filename"], "paper.pdf")
		}

		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(server)
	err := client.createAttachmentRecord("att-1", "pub-1", "owner-1", "paper.pdf", "md5hash", 2048)
	if err != nil {
		t.Fatalf("createAttachmentRecord() error: %v", err)
	}
}

func TestCreateAttachmentRecord_serverError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
	}))
	defer server.Close()

	client := newTestClient(server)
	err := client.createAttachmentRecord("att-1", "pub-1", "owner-1", "paper.pdf", "md5hash", 2048)
	if err == nil {
		t.Fatal("createAttachmentRecord() expected error for 500 response")
	}
}

func TestUploadToS3WithForm_success(t *testing.T) {
	s3Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}

		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Fatalf("failed to parse multipart form: %v", err)
		}

		if r.FormValue("key") != "attachments/abc" {
			t.Errorf("key = %q, want %q", r.FormValue("key"), "attachments/abc")
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("failed to get file from form: %v", err)
		}
		defer file.Close()

		w.WriteHeader(http.StatusNoContent)
	}))
	defer s3Server.Close()

	uploadData := &S3UploadData{
		URL: s3Server.URL,
		Fields: map[string]string{
			"key":    "attachments/abc",
			"policy": "base64policy",
		},
	}

	err := uploadToS3WithForm(uploadData, []byte("pdf content"))
	if err != nil {
		t.Fatalf("uploadToS3WithForm() error: %v", err)
	}
}

func TestUploadToS3WithForm_serverError(t *testing.T) {
	s3Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("access denied"))
	}))
	defer s3Server.Close()

	uploadData := &S3UploadData{
		URL:    s3Server.URL,
		Fields: map[string]string{},
	}

	err := uploadToS3WithForm(uploadData, []byte("pdf content"))
	if err == nil {
		t.Fatal("uploadToS3WithForm() expected error for 403 response")
	}
}
