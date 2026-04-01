package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/garaemon/paperpile/internal/api"
	"github.com/garaemon/paperpile/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(meCmd)
}

var meCmd = &cobra.Command{
	Use:   "me",
	Short: "Show current user info",
	RunE:  runMe,
}

func runMe(cmd *cobra.Command, args []string) error {
	client := api.NewClient(config.GetSession())
	return execMe(client, os.Stdout)
}

func execMe(fetcher UserFetcher, out io.Writer) error {
	user, err := fetcher.FetchCurrentUser()
	if err != nil {
		return fmt.Errorf("failed to fetch user info: %w", err)
	}

	fmt.Fprintf(out, "Name:  %s\n", user.GoogleName)
	fmt.Fprintf(out, "Email: %s\n", user.GoogleEmail)
	fmt.Fprintf(out, "ID:    %s\n", user.ID)
	return nil
}
