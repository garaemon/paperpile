package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/garaemon/paperpile-cli/internal/api"
	"github.com/garaemon/paperpile-cli/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(attachCmd)
}

var attachCmd = &cobra.Command{
	Use:   "attach <item_id> <file>",
	Short: "Attach a PDF to an existing library item",
	Args:  cobra.ExactArgs(2),
	RunE:  runAttach,
}

func runAttach(cmd *cobra.Command, args []string) error {
	itemID := args[0]

	filePath, err := filepath.Abs(args[1])
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	if _, err := os.Stat(filePath); err != nil {
		return fmt.Errorf("file not found: %s", filePath)
	}

	client := api.NewClient(config.GetSession())

	fmt.Printf("Attaching %s to item %s ...\n", filepath.Base(filePath), itemID)

	attachmentID, err := client.AttachFile(itemID, filePath)
	if err != nil {
		return fmt.Errorf("attach failed: %w", err)
	}

	fmt.Printf("Done! Attachment ID: %s\n", attachmentID)
	return nil
}
