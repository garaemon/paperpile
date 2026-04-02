package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/garaemon/paperpile/internal/api"
)

type mockLabelFetcher struct {
	labels []api.Collection
	err    error
}

func (m *mockLabelFetcher) FetchLabels() ([]api.Collection, error) {
	return m.labels, m.err
}

type mockItemLabelGetter struct {
	names []string
	err   error
}

func (m *mockItemLabelGetter) GetItemLabelNames(itemID string) ([]string, error) {
	return m.names, m.err
}

type mockTagAdder struct {
	calledItemID  string
	calledTagName string
	err           error
}

func (m *mockTagAdder) AddLabelByName(itemID, tagName string) error {
	m.calledItemID = itemID
	m.calledTagName = tagName
	return m.err
}

type mockTagRemover struct {
	calledItemID  string
	calledTagName string
	err           error
}

func (m *mockTagRemover) RemoveLabelByName(itemID, tagName string) error {
	m.calledItemID = itemID
	m.calledTagName = tagName
	return m.err
}

func TestExecTagList_success(t *testing.T) {
	fetcher := &mockLabelFetcher{
		labels: []api.Collection{
			{ID: "id-1", Name: "ML", Count: 5},
			{ID: "id-2", Name: "Robotics", Count: 3},
		},
	}

	var buf bytes.Buffer
	err := execTagList(fetcher, &buf)
	if err != nil {
		t.Fatalf("execTagList() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ML") {
		t.Errorf("output should contain 'ML', got: %s", output)
	}
	if !strings.Contains(output, "Robotics") {
		t.Errorf("output should contain 'Robotics', got: %s", output)
	}
}

func TestExecTagList_empty(t *testing.T) {
	fetcher := &mockLabelFetcher{labels: nil}

	var buf bytes.Buffer
	err := execTagList(fetcher, &buf)
	if err != nil {
		t.Fatalf("execTagList() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "(no tags)") {
		t.Errorf("output should contain '(no tags)', got: %s", output)
	}
}

func TestExecTagList_error(t *testing.T) {
	fetcher := &mockLabelFetcher{err: errors.New("api error")}

	var buf bytes.Buffer
	err := execTagList(fetcher, &buf)
	if err == nil {
		t.Fatal("execTagList() expected error")
	}
}

func TestExecTagGet_success(t *testing.T) {
	getter := &mockItemLabelGetter{names: []string{"ML", "Robotics"}}

	var buf bytes.Buffer
	err := execTagGet(getter, &buf, "item-1")
	if err != nil {
		t.Fatalf("execTagGet() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ML") {
		t.Errorf("output should contain 'ML', got: %s", output)
	}
	if !strings.Contains(output, "Robotics") {
		t.Errorf("output should contain 'Robotics', got: %s", output)
	}
}

func TestExecTagGet_empty(t *testing.T) {
	getter := &mockItemLabelGetter{names: nil}

	var buf bytes.Buffer
	err := execTagGet(getter, &buf, "item-1")
	if err != nil {
		t.Fatalf("execTagGet() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "(no tags)") {
		t.Errorf("output should contain '(no tags)', got: %s", output)
	}
}

func TestExecTagGet_error(t *testing.T) {
	getter := &mockItemLabelGetter{err: errors.New("not found")}

	var buf bytes.Buffer
	err := execTagGet(getter, &buf, "item-1")
	if err == nil {
		t.Fatal("execTagGet() expected error")
	}
}

func TestExecTagAdd_success(t *testing.T) {
	adder := &mockTagAdder{}

	var buf bytes.Buffer
	err := execTagAdd(adder, &buf, "item-1", "ML")
	if err != nil {
		t.Fatalf("execTagAdd() error: %v", err)
	}

	if adder.calledItemID != "item-1" {
		t.Errorf("calledItemID = %q, want %q", adder.calledItemID, "item-1")
	}
	if adder.calledTagName != "ML" {
		t.Errorf("calledTagName = %q, want %q", adder.calledTagName, "ML")
	}

	output := buf.String()
	if !strings.Contains(output, "ML") || !strings.Contains(output, "item-1") {
		t.Errorf("output should mention tag and item, got: %s", output)
	}
}

func TestExecTagAdd_error(t *testing.T) {
	adder := &mockTagAdder{err: errors.New("label not found")}

	var buf bytes.Buffer
	err := execTagAdd(adder, &buf, "item-1", "Nonexistent")
	if err == nil {
		t.Fatal("execTagAdd() expected error")
	}
}

func TestExecTagRemove_success(t *testing.T) {
	remover := &mockTagRemover{}

	var buf bytes.Buffer
	err := execTagRemove(remover, &buf, "item-1", "ML")
	if err != nil {
		t.Fatalf("execTagRemove() error: %v", err)
	}

	if remover.calledItemID != "item-1" {
		t.Errorf("calledItemID = %q, want %q", remover.calledItemID, "item-1")
	}
	if remover.calledTagName != "ML" {
		t.Errorf("calledTagName = %q, want %q", remover.calledTagName, "ML")
	}
}

type mockTagCreator struct {
	calledName string
	returnedID string
	err        error
}

func (m *mockTagCreator) CreateLabel(name string) (string, error) {
	m.calledName = name
	return m.returnedID, m.err
}

func TestExecTagCreate_success(t *testing.T) {
	creator := &mockTagCreator{returnedID: "new-id-123"}

	var buf bytes.Buffer
	err := execTagCreate(creator, &buf, "NewTag")
	if err != nil {
		t.Fatalf("execTagCreate() error: %v", err)
	}

	if creator.calledName != "NewTag" {
		t.Errorf("calledName = %q, want %q", creator.calledName, "NewTag")
	}

	output := buf.String()
	if !strings.Contains(output, "NewTag") {
		t.Errorf("output should mention tag name, got: %s", output)
	}
	if !strings.Contains(output, "new-id-123") {
		t.Errorf("output should mention tag ID, got: %s", output)
	}
}

func TestExecTagCreate_error(t *testing.T) {
	creator := &mockTagCreator{err: errors.New("sync failed")}

	var buf bytes.Buffer
	err := execTagCreate(creator, &buf, "NewTag")
	if err == nil {
		t.Fatal("execTagCreate() expected error")
	}
}

type mockTagDeleter struct {
	calledName string
	err        error
}

func (m *mockTagDeleter) DeleteLabel(name string) error {
	m.calledName = name
	return m.err
}

func TestExecTagDelete_success(t *testing.T) {
	deleter := &mockTagDeleter{}

	var buf bytes.Buffer
	err := execTagDelete(deleter, &buf, "TestTag")
	if err != nil {
		t.Fatalf("execTagDelete() error: %v", err)
	}

	if deleter.calledName != "TestTag" {
		t.Errorf("calledName = %q, want %q", deleter.calledName, "TestTag")
	}

	output := buf.String()
	if !strings.Contains(output, "TestTag") {
		t.Errorf("output should mention tag name, got: %s", output)
	}
}

func TestExecTagDelete_error(t *testing.T) {
	deleter := &mockTagDeleter{err: errors.New("not found")}

	var buf bytes.Buffer
	err := execTagDelete(deleter, &buf, "TestTag")
	if err == nil {
		t.Fatal("execTagDelete() expected error")
	}
}

func TestExecTagRemove_error(t *testing.T) {
	remover := &mockTagRemover{err: errors.New("not on item")}

	var buf bytes.Buffer
	err := execTagRemove(remover, &buf, "item-1", "ML")
	if err == nil {
		t.Fatal("execTagRemove() expected error")
	}
}
