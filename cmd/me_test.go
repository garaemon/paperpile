package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/garaemon/paperpile/internal/api"
)

type mockUserFetcher struct {
	user *api.UserInfo
	err  error
}

func (m *mockUserFetcher) FetchCurrentUser() (*api.UserInfo, error) {
	return m.user, m.err
}

func TestExecMe_success(t *testing.T) {
	fetcher := &mockUserFetcher{
		user: &api.UserInfo{
			ID:          "user-123",
			GoogleName:  "Test User",
			GoogleEmail: "test@example.com",
		},
	}

	var buf bytes.Buffer
	err := execMe(fetcher, &buf)
	if err != nil {
		t.Fatalf("execMe() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Test User") {
		t.Error("output should contain user name")
	}
	if !strings.Contains(output, "test@example.com") {
		t.Error("output should contain email")
	}
	if !strings.Contains(output, "user-123") {
		t.Error("output should contain user ID")
	}
}

func TestExecMe_fetchError(t *testing.T) {
	fetcher := &mockUserFetcher{
		err: errors.New("unauthorized"),
	}

	var buf bytes.Buffer
	err := execMe(fetcher, &buf)
	if err == nil {
		t.Fatal("execMe() expected error")
	}
	if !strings.Contains(err.Error(), "failed to fetch user info") {
		t.Errorf("error = %q, want to contain 'failed to fetch user info'", err.Error())
	}
}
