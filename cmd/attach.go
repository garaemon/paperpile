package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/garaemon/paperpile/internal/api"
	"github.com/garaemon/paperpile/internal/config"
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
	return execAttach(client, os.Stdout, itemID, filePath)
}

func execAttach(attacher FileAttacher, out io.Writer, itemID, filePath string) error {
	fmt.Fprintf(out, "Attaching %s to item %s ...\n", filepath.Base(filePath), itemID)

	attachmentID, err := attacher.AttachFile(itemID, filePath)
	if err != nil {
		return fmt.Errorf("attach failed: %w", err)
	}

	fmt.Fprintf(out, "Done! Attachment ID: %s\n", attachmentID)
	return nil
}
