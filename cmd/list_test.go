package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/garaemon/paperpile/internal/api"
)

type mockLibraryFetcher struct {
	items []api.LibraryItem
	err   error
}

func (m *mockLibraryFetcher) FetchLibrary() ([]api.LibraryItem, error) {
	return m.items, m.err
}

func TestExecList_success(t *testing.T) {
	fetcher := &mockLibraryFetcher{
		items: []api.LibraryItem{
			{
				ID:    "item-1",
				Title: "Deep Learning",
				Year:  "2016",
				Author: []api.Author{
					{Last: "Goodfellow"},
				},
			},
			{
				ID:    "item-2",
				Title: "Attention Is All You Need",
				Year:  "2017",
				Author: []api.Author{
					{Last: "Vaswani"},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := execList(fetcher, &buf, false, "created")
	if err != nil {
		t.Fatalf("execList() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "item-1") {
		t.Error("output should contain item-1")
	}
	if !strings.Contains(output, "Deep Learning") {
		t.Error("output should contain title")
	}
	if !strings.Contains(output, "Goodfellow") {
		t.Error("output should contain author")
	}
	if !strings.Contains(output, "2017") {
		t.Error("output should contain year")
	}
}

func TestExecList_filtersTrashed(t *testing.T) {
	fetcher := &mockLibraryFetcher{
		items: []api.LibraryItem{
			{ID: "active", Title: "Active Paper", Trashed: 0},
			{ID: "trashed", Title: "Trashed Paper", Trashed: 1},
		},
	}

	var buf bytes.Buffer
	err := execList(fetcher, &buf, false, "created")
	if err != nil {
		t.Fatalf("execList() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "active") {
		t.Error("output should contain active item")
	}
	if strings.Contains(output, "trashed") {
		t.Error("output should not contain trashed item")
	}
}

func TestExecList_includesTrashed(t *testing.T) {
	fetcher := &mockLibraryFetcher{
		items: []api.LibraryItem{
			{ID: "active", Title: "Active Paper", Trashed: 0},
			{ID: "trashed", Title: "Trashed Paper", Trashed: 1},
		},
	}

	var buf bytes.Buffer
	err := execList(fetcher, &buf, true, "created")
	if err != nil {
		t.Fatalf("execList() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "active") {
		t.Error("output should contain active item")
	}
	if !strings.Contains(output, "trashed") {
		t.Error("output should contain trashed item when includeTrashed=true")
	}
}

func TestExecList_truncatesLongTitle(t *testing.T) {
	longTitle := strings.Repeat("A", 100)
	fetcher := &mockLibraryFetcher{
		items: []api.LibraryItem{
			{ID: "item-1", Title: longTitle},
		},
	}

	var buf bytes.Buffer
	err := execList(fetcher, &buf, false, "created")
	if err != nil {
		t.Fatalf("execList() error: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, longTitle) {
		t.Error("long title should be truncated")
	}
	if !strings.Contains(output, "...") {
		t.Error("truncated title should end with ...")
	}
}

func TestExecList_emptyYearAndAuthor(t *testing.T) {
	fetcher := &mockLibraryFetcher{
		items: []api.LibraryItem{
			{ID: "item-1", Title: "No Metadata"},
		},
	}

	var buf bytes.Buffer
	err := execList(fetcher, &buf, false, "created")
	if err != nil {
		t.Fatalf("execList() error: %v", err)
	}

	output := buf.String()
	// Year and author should show "-" for missing values
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "item-1") {
			if !strings.Contains(line, "-") {
				t.Error("missing year/author should show '-'")
			}
			break
		}
	}
}

func TestExecList_fetchError(t *testing.T) {
	fetcher := &mockLibraryFetcher{
		err: errors.New("connection refused"),
	}

	var buf bytes.Buffer
	err := execList(fetcher, &buf, false, "created")
	if err == nil {
		t.Fatal("execList() expected error")
	}
	if !strings.Contains(err.Error(), "failed to fetch library") {
		t.Errorf("error = %q, want to contain 'failed to fetch library'", err.Error())
	}
}

func TestExecList_sortsByCreatedDescending(t *testing.T) {
	fetcher := &mockLibraryFetcher{
		items: []api.LibraryItem{
			{ID: "oldest", Title: "Oldest Paper", Created: 1000000},
			{ID: "newest", Title: "Newest Paper", Created: 3000000},
			{ID: "middle", Title: "Middle Paper", Created: 2000000},
		},
	}

	var buf bytes.Buffer
	err := execList(fetcher, &buf, false, "created")
	if err != nil {
		t.Fatalf("execList() error: %v", err)
	}

	output := buf.String()
	newestIdx := strings.Index(output, "newest")
	middleIdx := strings.Index(output, "middle")
	oldestIdx := strings.Index(output, "oldest")

	if newestIdx > middleIdx || middleIdx > oldestIdx {
		t.Errorf("items should be sorted by created descending: newest=%d, middle=%d, oldest=%d",
			newestIdx, middleIdx, oldestIdx)
	}
}

func TestExecList_sortNonePreservesOrder(t *testing.T) {
	fetcher := &mockLibraryFetcher{
		items: []api.LibraryItem{
			{ID: "first", Title: "First Paper", Created: 3000000},
			{ID: "second", Title: "Second Paper", Created: 1000000},
			{ID: "third", Title: "Third Paper", Created: 2000000},
		},
	}

	var buf bytes.Buffer
	err := execList(fetcher, &buf, false, "none")
	if err != nil {
		t.Fatalf("execList() error: %v", err)
	}

	output := buf.String()
	firstIdx := strings.Index(output, "first")
	secondIdx := strings.Index(output, "second")
	thirdIdx := strings.Index(output, "third")

	if firstIdx > secondIdx || secondIdx > thirdIdx {
		t.Errorf("items should preserve original order with sort=none: first=%d, second=%d, third=%d",
			firstIdx, secondIdx, thirdIdx)
	}
}

func TestExecList_header(t *testing.T) {
	fetcher := &mockLibraryFetcher{items: []api.LibraryItem{}}

	var buf bytes.Buffer
	err := execList(fetcher, &buf, false, "created")
	if err != nil {
		t.Fatalf("execList() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ID") || !strings.Contains(output, "TITLE") {
		t.Error("output should contain header row")
	}
}
