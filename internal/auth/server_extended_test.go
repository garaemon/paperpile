package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newPostRequest(path, body string) *http.Request {
	return httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
}

func newRequest(method, path, body string) *http.Request {
	return httptest.NewRequest(method, path, strings.NewReader(body))
}

func newRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

func TestStartCallbackServer_receivesSession(t *testing.T) {
	port := 19876

	resultCh := make(chan string, 1)
	errCh := make(chan error, 1)

	go func() {
		session, err := StartCallbackServer(port)
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- session
	}()

	// Wait for the server to start
	serverURL := fmt.Sprintf("http://localhost:%d", port)
	var connected bool
	for i := 0; i < 20; i++ {
		resp, err := http.Get(serverURL)
		if err == nil {
			resp.Body.Close()
			connected = true
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !connected {
		t.Fatal("server did not start in time")
	}

	// POST the session
	resp, err := http.Post(
		serverURL+"/callback",
		"text/plain",
		strings.NewReader("test-session-abc"),
	)
	if err != nil {
		t.Fatalf("failed to POST session: %v", err)
	}
	resp.Body.Close()

	select {
	case session := <-resultCh:
		if session != "test-session-abc" {
			t.Errorf("session = %q, want %q", session, "test-session-abc")
		}
	case err := <-errCh:
		t.Fatalf("StartCallbackServer() error: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for session")
	}
}

func TestHandleCallback_whitespaceSession(t *testing.T) {
	sessionCh := make(chan string, 1)

	req := newPostRequest("/callback", "  my-session  ")
	w := newRecorder()

	handleCallback(w, req, sessionCh)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	select {
	case session := <-sessionCh:
		if session != "my-session" {
			t.Errorf("session = %q, want %q (trimmed)", session, "my-session")
		}
	default:
		t.Fatal("expected session to be sent to channel")
	}
}

func TestHandleCallback_putMethodNotAllowed(t *testing.T) {
	sessionCh := make(chan string, 1)

	req := newRequest(http.MethodPut, "/callback", "")
	w := newRecorder()

	handleCallback(w, req, sessionCh)

	resp := w.Result()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
	}
}

func TestHandleCallback_deleteMethodNotAllowed(t *testing.T) {
	sessionCh := make(chan string, 1)

	req := newRequest(http.MethodDelete, "/callback", "")
	w := newRecorder()

	handleCallback(w, req, sessionCh)

	resp := w.Result()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
	}
}

func TestHandleCallback_whitespaceOnlySession(t *testing.T) {
	sessionCh := make(chan string, 1)

	req := newPostRequest("/callback", "   ")
	w := newRecorder()

	handleCallback(w, req, sessionCh)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestHandleSetupPage_containsSteps(t *testing.T) {
	w := newRecorder()
	handleSetupPage(w, 12345)

	body := w.Body.String()
	if !strings.Contains(body, "Step 1") {
		t.Error("setup page should contain Step 1")
	}
	if !strings.Contains(body, "Step 2") {
		t.Error("setup page should contain Step 2")
	}
	if !strings.Contains(body, "Step 3") {
		t.Error("setup page should contain Step 3")
	}
	if !strings.Contains(body, "localhost:12345") {
		t.Error("setup page should contain correct port in bookmarklet")
	}
	if !strings.Contains(body, "app.paperpile.com") {
		t.Error("setup page should contain link to paperpile")
	}
}
