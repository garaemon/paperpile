package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchAttachments_success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sync" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := SyncPullResponse{
			ServerChanges: []SyncServerChange{
				{
					Collection: "Library",
					Action:     "insert",
					ID:         "lib-1",
				},
				{
					Collection: "Attachments",
					Action:     "insert",
					ID:         "att-1",
					Attachment: AttachmentInfo{
						ID:       "att-1",
						PubID:    "lib-1",
						Filename: "paper.pdf",
						MimeType: "application/pdf",
						MD5:      "abc123",
						Filesize: 12345,
					},
				},
				{
					Collection: "Attachments",
					Action:     "insert",
					ID:         "att-2",
					Attachment: AttachmentInfo{
						ID:       "att-2",
						PubID:    "lib-2",
						Filename: "notes.txt",
						MimeType: "text/plain",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(server)
	attachments, err := client.FetchAttachments()
	if err != nil {
		t.Fatalf("FetchAttachments() error: %v", err)
	}

	if len(attachments) != 2 {
		t.Fatalf("got %d attachments, want 2", len(attachments))
	}
	if attachments[0].Filename != "paper.pdf" {
		t.Errorf("Filename = %q, want %q", attachments[0].Filename, "paper.pdf")
	}
}

func TestFindAttachmentsForItem_success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := SyncPullResponse{
			ServerChanges: []SyncServerChange{
				{
					Collection: "Attachments",
					Action:     "insert",
					ID:         "att-1",
					Attachment: AttachmentInfo{ID: "att-1", PubID: "item-A", Filename: "a.pdf"},
				},
				{
					Collection: "Attachments",
					Action:     "insert",
					ID:         "att-2",
					Attachment: AttachmentInfo{ID: "att-2", PubID: "item-B", Filename: "b.pdf"},
				},
				{
					Collection: "Attachments",
					Action:     "insert",
					ID:         "att-3",
					Attachment: AttachmentInfo{ID: "att-3", PubID: "item-A", Filename: "supplement.docx"},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(server)
	attachments, err := client.FindAttachmentsForItem("item-A")
	if err != nil {
		t.Fatalf("FindAttachmentsForItem() error: %v", err)
	}

	if len(attachments) != 2 {
		t.Fatalf("got %d attachments, want 2", len(attachments))
	}
	if attachments[0].Filename != "a.pdf" {
		t.Errorf("first attachment = %q, want %q", attachments[0].Filename, "a.pdf")
	}
	if attachments[1].Filename != "supplement.docx" {
		t.Errorf("second attachment = %q, want %q", attachments[1].Filename, "supplement.docx")
	}
}

func TestFindAttachmentsForItem_notFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := SyncPullResponse{ServerChanges: []SyncServerChange{}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.FindAttachmentsForItem("nonexistent")
	if err == nil {
		t.Fatal("FindAttachmentsForItem() expected error for missing item")
	}
}

func TestDownloadAttachment_success(t *testing.T) {
	expectedContent := []byte("%PDF-1.4 fake content")

	// The API returns a 307 redirect, but httptest + http.Client follow it.
	// Simulate by directly serving the file.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/attachments/att-123/file" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("type") != "latest" {
			t.Errorf("missing type=latest query param")
		}

		w.Header().Set("Content-Type", "application/pdf")
		w.Write(expectedContent)
	}))
	defer server.Close()

	client := newTestClient(server)
	attachment := AttachmentInfo{
		ID:       "att-123",
		Filename: "paper.pdf",
		MimeType: "application/pdf",
	}

	result, err := client.DownloadAttachment(attachment)
	if err != nil {
		t.Fatalf("DownloadAttachment() error: %v", err)
	}

	if string(result.Data) != string(expectedContent) {
		t.Errorf("Data = %q, want %q", string(result.Data), string(expectedContent))
	}
	if result.Filename != "paper.pdf" {
		t.Errorf("Filename = %q, want %q", result.Filename, "paper.pdf")
	}
}

func TestDownloadAttachment_apiError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer server.Close()

	client := newTestClient(server)
	attachment := AttachmentInfo{ID: "att-missing", Filename: "gone.pdf"}

	_, err := client.DownloadAttachment(attachment)
	if err == nil {
		t.Fatal("DownloadAttachment() expected error for 404 response")
	}
}
