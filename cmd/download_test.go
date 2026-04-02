package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/garaemon/paperpile/internal/api"
)

type mockFileDownloader struct {
	result *api.DownloadResult
	err    error
}

func (m *mockFileDownloader) DownloadFile(itemID string) (*api.DownloadResult, error) {
	return m.result, m.err
}

func TestExecDownload_success(t *testing.T) {
	tmpDir := t.TempDir()
	downloader := &mockFileDownloader{
		result: &api.DownloadResult{
			Data:     []byte("fake pdf"),
			Filename: "paper.pdf",
		},
	}

	var buf bytes.Buffer
	err := execDownload(downloader, &buf, "item-abc", tmpDir)
	if err != nil {
		t.Fatalf("execDownload() error: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "paper.pdf")
	if !strings.Contains(buf.String(), expectedPath) {
		t.Errorf("output = %q, want to contain %q", buf.String(), expectedPath)
	}

	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if string(content) != "fake pdf" {
		t.Errorf("file content = %q, want %q", string(content), "fake pdf")
	}
}

func TestExecDownload_downloadError(t *testing.T) {
	downloader := &mockFileDownloader{
		err: errors.New("network error"),
	}

	var buf bytes.Buffer
	err := execDownload(downloader, &buf, "item-abc", ".")
	if err == nil {
		t.Fatal("execDownload() expected error")
	}
	if !strings.Contains(err.Error(), "download failed") {
		t.Errorf("error = %q, want to contain 'download failed'", err.Error())
	}
}
