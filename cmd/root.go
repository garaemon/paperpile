package cmd

import (
	"fmt"

	"github.com/garaemon/paperpile/internal/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "paperpile",
	Short: "CLI tool for Paperpile",
	Long:  "A command-line tool to upload, list, and delete references in Paperpile.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "login" {
			return nil
		}
		if err := config.Load(); err != nil {
			return err
		}
		if config.GetSession() == "" {
			return fmt.Errorf("not logged in. Run 'paperpile login' first")
		}
		return nil
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
