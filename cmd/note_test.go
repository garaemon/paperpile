package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

type mockNoteGetter struct {
	note string
	err  error
}

func (m *mockNoteGetter) GetNote(itemID string) (string, error) {
	return m.note, m.err
}

type mockNoteUpdater struct {
	updatedID   string
	updatedNote string
	err         error
}

func (m *mockNoteUpdater) UpdateNote(itemID, note string) error {
	m.updatedID = itemID
	m.updatedNote = note
	return m.err
}

func TestExecNoteGet_success(t *testing.T) {
	getter := &mockNoteGetter{note: "My research note"}

	var buf bytes.Buffer
	err := execNoteGet(getter, &buf, "item-abc", false)
	if err != nil {
		t.Fatalf("execNoteGet() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "My research note") {
		t.Errorf("output = %q, want to contain 'My research note'", output)
	}
}

func TestExecNoteGet_emptyNote(t *testing.T) {
	getter := &mockNoteGetter{note: ""}

	var buf bytes.Buffer
	err := execNoteGet(getter, &buf, "item-abc", false)
	if err != nil {
		t.Fatalf("execNoteGet() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "(no note)") {
		t.Errorf("output = %q, want to contain '(no note)'", output)
	}
}

func TestExecNoteGet_error(t *testing.T) {
	getter := &mockNoteGetter{err: errors.New("not found")}

	var buf bytes.Buffer
	err := execNoteGet(getter, &buf, "item-abc", false)
	if err == nil {
		t.Fatal("execNoteGet() expected error")
	}
	if !strings.Contains(err.Error(), "failed to get note") {
		t.Errorf("error = %q, want to contain 'failed to get note'", err.Error())
	}
}

func TestExecNoteSet_success(t *testing.T) {
	updater := &mockNoteUpdater{}

	var buf bytes.Buffer
	err := execNoteSet(updater, &buf, "item-abc", "New note text", false)
	if err != nil {
		t.Fatalf("execNoteSet() error: %v", err)
	}

	if updater.updatedID != "item-abc" {
		t.Errorf("updatedID = %q, want %q", updater.updatedID, "item-abc")
	}
	if updater.updatedNote != "New note text" {
		t.Errorf("updatedNote = %q, want %q", updater.updatedNote, "New note text")
	}

	output := buf.String()
	if !strings.Contains(output, "item-abc") {
		t.Error("output should mention the item ID")
	}
}

func TestExecNoteGet_markdown(t *testing.T) {
	getter := &mockNoteGetter{note: "<p>This is <b>bold</b> text</p>"}

	var buf bytes.Buffer
	err := execNoteGet(getter, &buf, "item-abc", true)
	if err != nil {
		t.Fatalf("execNoteGet() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "**bold**") {
		t.Errorf("output = %q, want to contain **bold**", output)
	}
}

func TestExecNoteSet_markdown(t *testing.T) {
	updater := &mockNoteUpdater{}

	var buf bytes.Buffer
	err := execNoteSet(updater, &buf, "item-abc", "This is **bold** text", true)
	if err != nil {
		t.Fatalf("execNoteSet() error: %v", err)
	}

	if !strings.Contains(updater.updatedNote, "<strong>bold</strong>") {
		t.Errorf("updatedNote = %q, want to contain <strong>bold</strong>", updater.updatedNote)
	}
}

func TestExecNoteSet_error(t *testing.T) {
	updater := &mockNoteUpdater{err: errors.New("sync failed")}

	var buf bytes.Buffer
	err := execNoteSet(updater, &buf, "item-abc", "Note text", false)
	if err == nil {
		t.Fatal("execNoteSet() expected error")
	}
	if !strings.Contains(err.Error(), "failed to set note") {
		t.Errorf("error = %q, want to contain 'failed to set note'", err.Error())
	}
}
