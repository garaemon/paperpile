package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/garaemon/paperpile/internal/api"
	"github.com/garaemon/paperpile/internal/auth"
	"github.com/garaemon/paperpile/internal/config"
	"github.com/spf13/cobra"
)

const defaultPort = 18080

var loginPort int

func init() {
	loginCmd.Flags().IntVar(&loginPort, "port", defaultPort, "Local server port for receiving session")
	rootCmd.AddCommand(loginCmd)
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Paperpile via bookmarklet",
	Long: `Start a local HTTP server and open a setup page in your browser.
Drag the bookmarklet to your bookmarks bar, then use it on app.paperpile.com.`,
	RunE: runLogin,
}

func runLogin(cmd *cobra.Command, args []string) error {
	setupURL := fmt.Sprintf("http://localhost:%d", loginPort)

	fmt.Println("=== Paperpile CLI Login ===")
	fmt.Println()
	fmt.Printf("Opening setup page: %s\n", setupURL)
	fmt.Printf("Waiting for session ...\n")

	// Open browser after a short delay so the server is ready.
	go func() {
		if err := openBrowser(setupURL); err != nil {
			fmt.Printf("Could not open browser. Please visit: %s\n", setupURL)
		}
	}()

	session, err := auth.StartCallbackServer(loginPort)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	fmt.Println("Verifying session ...")
	client := api.NewClient(session)
	user, err := client.FetchCurrentUser()
	if err != nil {
		return fmt.Errorf("session verification failed: %w", err)
	}

	if err := config.SaveSession(session); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	fmt.Printf("Login successful! Welcome, %s (%s)\n", user.GoogleName, user.GoogleEmail)
	return nil
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return exec.Command(cmd, args...).Start()
}
