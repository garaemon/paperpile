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
	return execUpload(client, os.Stdout, filePath, uploadDuplicates)
}

func execUpload(uploader PDFUploader, out io.Writer, filePath string, allowDuplicates bool) error {
	fmt.Fprintf(out, "Uploading %s ...\n", filepath.Base(filePath))

	task, err := uploader.UploadPDF(filePath, allowDuplicates)
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	fmt.Fprintf(out, "Done! Task ID: %s\n", task.ID)
	return nil
}
