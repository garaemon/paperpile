package auth

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

const callbackPath = "/callback"

// StartCallbackServer starts a local HTTP server that waits for the bookmarklet
// to POST the plack_session value. It returns the session string once received.
func StartCallbackServer(port int) (string, error) {
	sessionCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleSetupPage(w, port)
	})
	mux.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
		handleCallback(w, r, sessionCh)
	})

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return "", fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	server := &http.Server{Handler: mux}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	select {
	case session := <-sessionCh:
		return session, nil
	case err := <-errCh:
		return "", err
	case <-time.After(5 * time.Minute):
		return "", fmt.Errorf("login timed out (5 minutes)")
	}
}

func handleSetupPage(w http.ResponseWriter, port int) {
	bookmarklet := fmt.Sprintf(
		`javascript:void(fetch('http://localhost:%d/callback',{method:'POST',body:document.cookie.match(/plack_session=([^;]+)/)[1]}))`,
		port,
	)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html><head><title>Paperpile CLI Login</title>
<style>
  body { font-family: -apple-system, sans-serif; max-width: 600px; margin: 80px auto; padding: 0 20px; color: #333; }
  h1 { font-size: 1.4em; }
  .bookmarklet { display: inline-block; background: #2563eb; color: white; padding: 12px 24px;
    border-radius: 8px; text-decoration: none; font-size: 1.1em; font-weight: bold; cursor: grab; }
  .bookmarklet:hover { background: #1d4ed8; }
  .step { margin: 20px 0; padding: 16px; background: #f8f9fa; border-radius: 8px; }
  .step-num { font-weight: bold; color: #2563eb; }
  .done { display: none; text-align: center; margin-top: 40px; }
  .done.show { display: block; }
  .done h2 { color: #16a34a; }
</style></head>
<body>
  <h1>Paperpile CLI Login</h1>

  <div class="step">
    <span class="step-num">Step 1:</span>
    Drag this button to your bookmarks bar (one-time only):<br><br>
    <a class="bookmarklet" href="%s">Paperpile: Send to CLI</a>
  </div>

  <div class="step">
    <span class="step-num">Step 2:</span>
    Go to <a href="https://app.paperpile.com" target="_blank">app.paperpile.com</a>
    (make sure you are logged in).
  </div>

  <div class="step">
    <span class="step-num">Step 3:</span>
    Click the <strong>"Paperpile: Send to CLI"</strong> bookmarklet in your bookmarks bar.
  </div>

  <div class="done" id="done">
    <h2>Login successful!</h2>
    <p>You can close this tab and return to your terminal.</p>
  </div>

  <script>
    // Poll to check if login succeeded, then show success message.
    let check = setInterval(async () => {
      try {
        let r = await fetch('/callback', {method: 'OPTIONS'});
        // Server shuts down after receiving session, so a network error means success.
      } catch(e) {
        clearInterval(check);
        document.getElementById('done').classList.add('show');
      }
    }, 2000);
  </script>
</body></html>`, bookmarklet)
}

func handleCallback(w http.ResponseWriter, r *http.Request, sessionCh chan<- string) {
	// Allow CORS from any origin so the bookmarklet works from app.paperpile.com.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	session := strings.TrimSpace(string(body))
	if session == "" {
		http.Error(w, "empty session", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, "ok")

	sessionCh <- session
}
