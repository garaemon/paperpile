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
	attachments []api.AttachmentInfo
	findErr     error
	result      *api.DownloadResult
	downloadErr error
}

func (m *mockFileDownloader) FindAttachmentsForItem(itemID string) ([]api.AttachmentInfo, error) {
	return m.attachments, m.findErr
}

func (m *mockFileDownloader) DownloadAttachment(attachment api.AttachmentInfo) (*api.DownloadResult, error) {
	return m.result, m.downloadErr
}

func TestExecDownload_success(t *testing.T) {
	tmpDir := t.TempDir()
	downloader := &mockFileDownloader{
		attachments: []api.AttachmentInfo{
			{ID: "att-1", PubID: "item-abc", Filename: "paper.pdf", MimeType: "application/pdf"},
		},
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

func TestExecDownload_noAttachments(t *testing.T) {
	downloader := &mockFileDownloader{
		findErr: errors.New("no attachments found for item missing-item"),
	}

	var buf bytes.Buffer
	err := execDownload(downloader, &buf, "missing-item", ".")
	if err == nil {
		t.Fatal("execDownload() expected error")
	}
	if !strings.Contains(err.Error(), "failed to find attachments") {
		t.Errorf("error = %q, want to contain 'failed to find attachments'", err.Error())
	}
}

func TestExecDownload_downloadError(t *testing.T) {
	downloader := &mockFileDownloader{
		attachments: []api.AttachmentInfo{
			{ID: "att-1", PubID: "item-abc", Filename: "paper.pdf"},
		},
		downloadErr: errors.New("network error"),
	}

	var buf bytes.Buffer
	err := execDownload(downloader, &buf, "item-abc", ".")
	if err == nil {
		t.Fatal("execDownload() expected error")
	}
	if !strings.Contains(err.Error(), "failed to download") {
		t.Errorf("error = %q, want to contain 'failed to download'", err.Error())
	}
}
