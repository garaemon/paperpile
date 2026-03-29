package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateNote_success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/sync" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}

		var reqBody map[string]any
		if err := json.Unmarshal(body, &reqBody); err != nil {
			t.Fatalf("failed to unmarshal body: %v", err)
		}

		changes, ok := reqBody["clientChanges"].([]any)
		if !ok || len(changes) == 0 {
			t.Fatal("expected clientChanges to be non-empty")
		}

		change := changes[0].(map[string]any)
		if change["mcollection"] != "Library" {
			t.Errorf("mcollection = %v, want %q", change["mcollection"], "Library")
		}
		if change["action"] != "update" {
			t.Errorf("action = %v, want %q", change["action"], "update")
		}
		if change["id"] != "item-456" {
			t.Errorf("id = %v, want %q", change["id"], "item-456")
		}

		data, ok := change["data"].(map[string]any)
		if !ok {
			t.Fatal("expected data to be a map")
		}
		if data["note"] != "This is a test note" {
			t.Errorf("notes = %v, want %q", data["note"], "This is a test note")
		}

		resp := SyncResponse{
			SyncStartTime: 1234567890.0,
			SyncSession:   "session-1",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(server)
	err := client.UpdateNote("item-456", "This is a test note")
	if err != nil {
		t.Fatalf("UpdateNote() error: %v", err)
	}
}

func TestUpdateNote_serverError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	client := newTestClient(server)
	err := client.UpdateNote("item-456", "note text")
	if err == nil {
		t.Fatal("UpdateNote() expected error for 500 response")
	}
}

func TestGetNote_success(t *testing.T) {
	items := []LibraryItem{
		{ID: "item-1", Title: "Paper 1", Notes: ""},
		{ID: "item-2", Title: "Paper 2", Notes: "Important findings"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(items)
	}))
	defer server.Close()

	client := newTestClient(server)
	note, err := client.GetNote("item-2")
	if err != nil {
		t.Fatalf("GetNote() error: %v", err)
	}
	if note != "Important findings" {
		t.Errorf("note = %q, want %q", note, "Important findings")
	}
}

func TestGetNote_notFound(t *testing.T) {
	items := []LibraryItem{
		{ID: "item-1", Title: "Paper 1"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(items)
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetNote("nonexistent")
	if err == nil {
		t.Fatal("GetNote() expected error for nonexistent item")
	}
}
