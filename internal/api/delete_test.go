package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTrashItem_success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path+"/"+r.URL.RawQuery != "/sync/"+"v=3" {
			// Verify it hits the sync endpoint
			if r.URL.Path != "/sync" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
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
		if change["id"] != "item-123" {
			t.Errorf("id = %v, want %q", change["id"], "item-123")
		}

		data, ok := change["data"].(map[string]any)
		if !ok {
			t.Fatal("expected data to be a map")
		}
		if data["trashed"] != float64(1) {
			t.Errorf("trashed = %v, want 1", data["trashed"])
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
	err := client.TrashItem("item-123")
	if err != nil {
		t.Fatalf("TrashItem() error: %v", err)
	}
}

func TestTrashItem_serverError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	client := newTestClient(server)
	err := client.TrashItem("item-123")
	if err == nil {
		t.Fatal("TrashItem() expected error for 500 response")
	}
}
