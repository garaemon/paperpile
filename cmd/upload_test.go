package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/garaemon/paperpile/internal/api"
)

type mockPDFUploader struct {
	calledPath       string
	calledDuplicates bool
	task             *api.ImportTask
	err              error
}

func (m *mockPDFUploader) UploadPDF(filePath string, importDuplicates bool) (*api.ImportTask, error) {
	m.calledPath = filePath
	m.calledDuplicates = importDuplicates
	return m.task, m.err
}

func TestExecUpload_success(t *testing.T) {
	uploader := &mockPDFUploader{
		task: &api.ImportTask{ID: "task-999"},
	}

	var buf bytes.Buffer
	err := execUpload(uploader, &buf, "/path/to/paper.pdf", false)
	if err != nil {
		t.Fatalf("execUpload() error: %v", err)
	}

	if uploader.calledPath != "/path/to/paper.pdf" {
		t.Errorf("calledPath = %q, want %q", uploader.calledPath, "/path/to/paper.pdf")
	}
	if uploader.calledDuplicates != false {
		t.Error("calledDuplicates should be false")
	}

	output := buf.String()
	if !strings.Contains(output, "paper.pdf") {
		t.Error("output should contain file name")
	}
	if !strings.Contains(output, "task-999") {
		t.Error("output should contain task ID")
	}
}

func TestExecUpload_withDuplicates(t *testing.T) {
	uploader := &mockPDFUploader{
		task: &api.ImportTask{ID: "task-111"},
	}

	var buf bytes.Buffer
	err := execUpload(uploader, &buf, "/path/to/paper.pdf", true)
	if err != nil {
		t.Fatalf("execUpload() error: %v", err)
	}

	if !uploader.calledDuplicates {
		t.Error("calledDuplicates should be true")
	}
}

func TestExecUpload_uploadError(t *testing.T) {
	uploader := &mockPDFUploader{
		err: errors.New("S3 timeout"),
	}

	var buf bytes.Buffer
	err := execUpload(uploader, &buf, "/path/to/paper.pdf", false)
	if err == nil {
		t.Fatal("execUpload() expected error")
	}
	if !strings.Contains(err.Error(), "upload failed") {
		t.Errorf("error = %q, want to contain 'upload failed'", err.Error())
	}
}
