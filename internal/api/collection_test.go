package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchCollections_success(t *testing.T) {
	collections := []Collection{
		{ID: "label-1", Name: "ML", CollectionType: "label", Count: 5},
		{ID: "folder-1", Name: "Unread", CollectionType: "folder", Count: 10},
		{ID: "label-2", Name: "Robotics", CollectionType: "label", Count: 3},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/collections" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(collections)
	}))
	defer server.Close()

	client := newTestClient(server)
	result, err := client.FetchCollections()
	if err != nil {
		t.Fatalf("FetchCollections() error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("got %d collections, want 3", len(result))
	}
	if result[0].Name != "ML" {
		t.Errorf("result[0].Name = %q, want %q", result[0].Name, "ML")
	}
}

func TestFetchCollections_serverError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.FetchCollections()
	if err == nil {
		t.Fatal("FetchCollections() expected error for 500 response")
	}
}

func TestFetchLabels_filtersLabelsOnly(t *testing.T) {
	collections := []Collection{
		{ID: "label-1", Name: "ML", CollectionType: "label"},
		{ID: "folder-1", Name: "Unread", CollectionType: "folder"},
		{ID: "label-2", Name: "Robotics", CollectionType: "label"},
		{ID: "label-3", Name: "Trashed", CollectionType: "label", Trashed: 1},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(collections)
	}))
	defer server.Close()

	client := newTestClient(server)
	labels, err := client.FetchLabels()
	if err != nil {
		t.Fatalf("FetchLabels() error: %v", err)
	}
	if len(labels) != 2 {
		t.Fatalf("got %d labels, want 2", len(labels))
	}
	if labels[0].Name != "ML" {
		t.Errorf("labels[0].Name = %q, want %q", labels[0].Name, "ML")
	}
	if labels[1].Name != "Robotics" {
		t.Errorf("labels[1].Name = %q, want %q", labels[1].Name, "Robotics")
	}
}

func TestIsLabel(t *testing.T) {
	label := Collection{CollectionType: "label"}
	folder := Collection{CollectionType: "folder"}

	if !label.IsLabel() {
		t.Error("expected label.IsLabel() to be true")
	}
	if folder.IsLabel() {
		t.Error("expected folder.IsLabel() to be false")
	}
}
