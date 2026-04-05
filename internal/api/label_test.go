package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetItemLabels_success(t *testing.T) {
	items := []LibraryItem{
		{ID: "item-1", Title: "Paper 1", LabelIDs: []string{"label-a", "label-b"}},
		{ID: "item-2", Title: "Paper 2", LabelIDs: nil},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(items)
	}))
	defer server.Close()

	client := newTestClient(server)
	labels, err := client.GetItemLabels("item-1")
	if err != nil {
		t.Fatalf("GetItemLabels() error: %v", err)
	}
	if len(labels) != 2 {
		t.Fatalf("got %d labels, want 2", len(labels))
	}
	if labels[0] != "label-a" {
		t.Errorf("labels[0] = %q, want %q", labels[0], "label-a")
	}
}

func TestGetItemLabels_notFound(t *testing.T) {
	items := []LibraryItem{
		{ID: "item-1", Title: "Paper 1"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(items)
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetItemLabels("nonexistent")
	if err == nil {
		t.Fatal("GetItemLabels() expected error for nonexistent item")
	}
}

func TestGetItemLabelNames_success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/library" {
			items := []LibraryItem{
				{ID: "item-1", LabelIDs: []string{"id-1", "id-2"}},
			}
			json.NewEncoder(w).Encode(items)
			return
		}
		if r.URL.Path == "/collections" {
			collections := []Collection{
				{ID: "id-1", Name: "ML", CollectionType: "label"},
				{ID: "id-2", Name: "Robotics", CollectionType: "label"},
				{ID: "id-3", Name: "Unread", CollectionType: "folder"},
			}
			json.NewEncoder(w).Encode(collections)
			return
		}
	}))
	defer server.Close()

	client := newTestClient(server)
	names, err := client.GetItemLabelNames("item-1")
	if err != nil {
		t.Fatalf("GetItemLabelNames() error: %v", err)
	}
	if len(names) != 2 {
		t.Fatalf("got %d names, want 2", len(names))
	}
	if names[0] != "ML" {
		t.Errorf("names[0] = %q, want %q", names[0], "ML")
	}
	if names[1] != "Robotics" {
		t.Errorf("names[1] = %q, want %q", names[1], "Robotics")
	}
}

func TestResolveLabelName_success(t *testing.T) {
	collections := []Collection{
		{ID: "id-1", Name: "ML", CollectionType: "label"},
		{ID: "id-2", Name: "Robotics", CollectionType: "label"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(collections)
	}))
	defer server.Close()

	client := newTestClient(server)
	id, err := client.ResolveLabelName("Robotics")
	if err != nil {
		t.Fatalf("ResolveLabelName() error: %v", err)
	}
	if id != "id-2" {
		t.Errorf("id = %q, want %q", id, "id-2")
	}
}

func TestResolveLabelName_notFound(t *testing.T) {
	collections := []Collection{
		{ID: "id-1", Name: "ML", CollectionType: "label"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(collections)
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.ResolveLabelName("Nonexistent")
	if err == nil {
		t.Fatal("ResolveLabelName() expected error for nonexistent label")
	}
}

func TestCreateLabel_success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sync" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			return
		}

		body, _ := io.ReadAll(r.Body)
		var reqBody map[string]any
		json.Unmarshal(body, &reqBody)

		resp := SyncResponse{SyncStartTime: 1234567890.0}
		w.Header().Set("Content-Type", "application/json")

		changes, ok := reqBody["clientChanges"].([]any)
		if !ok || len(changes) == 0 {
			json.NewEncoder(w).Encode(resp)
			return
		}

		change := changes[0].(map[string]any)

		if change["mcollection"] != "Collections" {
			t.Errorf("mcollection = %v, want %q", change["mcollection"], "Collections")
		}
		if change["action"] != "insert" {
			t.Errorf("action = %v, want %q", change["action"], "insert")
		}

		data := change["data"].(map[string]any)
		if data["cName"] != "NewTag" {
			t.Errorf("cName = %v, want %q", data["cName"], "NewTag")
		}
		if data["collectionType"] != "label" {
			t.Errorf("collectionType = %v, want %q", data["collectionType"], "label")
		}
		if data["cParent"] != "ROOT" {
			t.Errorf("cParent = %v, want %q", data["cParent"], "ROOT")
		}

		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(server)
	id, err := client.CreateLabel("NewTag")
	if err != nil {
		t.Fatalf("CreateLabel() error: %v", err)
	}
	if id == "" {
		t.Error("CreateLabel() returned empty ID")
	}
}

func TestCreateLabel_serverError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.CreateLabel("NewTag")
	if err == nil {
		t.Fatal("CreateLabel() expected error for 500 response")
	}
}
