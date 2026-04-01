package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPushSyncChanges_success(t *testing.T) {
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

		if reqBody["syncClientId"] != "paperpile" {
			t.Errorf("syncClientId = %v, want %q", reqBody["syncClientId"], "paperpile")
		}

		resp := SyncResponse{
			SyncStartTime:  1234567890.0,
			SyncSession:    "session-1",
			TotalChanges:   0,
			LastClientSync: 1234567890.0,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(server)
	changes := []map[string]any{
		{"action": "update", "id": "item1"},
	}
	resp, err := client.pushSyncChanges(changes)
	if err != nil {
		t.Fatalf("pushSyncChanges() error: %v", err)
	}
	if resp.SyncSession != "session-1" {
		t.Errorf("SyncSession = %q, want %q", resp.SyncSession, "session-1")
	}
}

func TestPushSyncChanges_serverError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.pushSyncChanges([]map[string]any{})
	if err == nil {
		t.Fatal("pushSyncChanges() expected error for 500 response")
	}
}
