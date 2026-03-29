package cmd

import (
	"fmt"

	"github.com/garaemon/paperpile-cli/internal/api"
	"github.com/garaemon/paperpile-cli/internal/config"
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
	user, err := client.FetchCurrentUser()
	if err != nil {
		return fmt.Errorf("failed to fetch user info: %w", err)
	}

	fmt.Printf("Name:  %s\n", user.GoogleName)
	fmt.Printf("Email: %s\n", user.GoogleEmail)
	fmt.Printf("ID:    %s\n", user.ID)
	return nil
}
