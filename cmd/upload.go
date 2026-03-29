package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/garaemon/paperpile-cli/internal/api"
	"github.com/garaemon/paperpile-cli/internal/config"
	"github.com/spf13/cobra"
)

var uploadDuplicates bool

func init() {
	uploadCmd.Flags().BoolVar(&uploadDuplicates, "allow-duplicates", false, "Import even if a duplicate exists")
	rootCmd.AddCommand(uploadCmd)
}

var uploadCmd = &cobra.Command{
	Use:   "upload <file>",
	Short: "Upload a PDF to Paperpile",
	Args:  cobra.ExactArgs(1),
	RunE:  runUpload,
}

func runUpload(cmd *cobra.Command, args []string) error {
	filePath, err := filepath.Abs(args[0])
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	if _, err := os.Stat(filePath); err != nil {
		return fmt.Errorf("file not found: %s", filePath)
	}

	client := api.NewClient(config.GetSession())

	fmt.Printf("Uploading %s ...\n", filepath.Base(filePath))

	task, err := client.UploadPDF(filePath, uploadDuplicates)
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	fmt.Printf("Done! Task ID: %s\n", task.ID)
	return nil
}
