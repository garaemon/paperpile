package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDownloadFile_success(t *testing.T) {
	expectedContent := []byte("%PDF-1.4 fake pdf content")
	expectedFilename := "paper.pdf"

	// S3 server that serves the actual file
	s3Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Disposition", `attachment; filename="paper.pdf"`)
		w.Header().Set("Content-Type", "application/pdf")
		w.Write(expectedContent)
	}))
	defer s3Server.Close()

	// API server that returns the presigned URL
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/items/item-123/file" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"url":"` + s3Server.URL + `/paper.pdf"}`))
	}))
	defer apiServer.Close()

	client := newTestClient(apiServer)
	result, err := client.DownloadFile("item-123")
	if err != nil {
		t.Fatalf("DownloadFile() error: %v", err)
	}

	if string(result.Data) != string(expectedContent) {
		t.Errorf("Data = %q, want %q", string(result.Data), string(expectedContent))
	}
	if result.Filename != expectedFilename {
		t.Errorf("Filename = %q, want %q", result.Filename, expectedFilename)
	}
}

func TestDownloadFile_noContentDisposition(t *testing.T) {
	s3Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pdf data"))
	}))
	defer s3Server.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"url":"` + s3Server.URL + `"}`))
	}))
	defer apiServer.Close()

	client := newTestClient(apiServer)
	result, err := client.DownloadFile("item-456")
	if err != nil {
		t.Fatalf("DownloadFile() error: %v", err)
	}

	if result.Filename != "item-456.pdf" {
		t.Errorf("Filename = %q, want %q", result.Filename, "item-456.pdf")
	}
}

func TestDownloadFile_apiError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.DownloadFile("item-missing")
	if err == nil {
		t.Fatal("DownloadFile() expected error for 404 response")
	}
}

func TestDownloadFile_s3Error(t *testing.T) {
	s3Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("access denied"))
	}))
	defer s3Server.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"url":"` + s3Server.URL + `"}`))
	}))
	defer apiServer.Close()

	client := newTestClient(apiServer)
	_, err := client.DownloadFile("item-123")
	if err == nil {
		t.Fatal("DownloadFile() expected error for S3 403 response")
	}
}
