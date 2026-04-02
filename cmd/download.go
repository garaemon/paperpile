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

var downloadOutputDir string

func init() {
	downloadCmd.Flags().StringVarP(&downloadOutputDir, "output", "o", ".", "Output directory for the downloaded file")
	rootCmd.AddCommand(downloadCmd)
}

var downloadCmd = &cobra.Command{
	Use:   "download <item_id>",
	Short: "Download the PDF for a library item",
	Args:  cobra.ExactArgs(1),
	RunE:  runDownload,
}

func runDownload(cmd *cobra.Command, args []string) error {
	client := api.NewClient(config.GetSession())
	return execDownload(client, os.Stdout, args[0], downloadOutputDir)
}

func execDownload(downloader FileDownloader, out io.Writer, itemID, outputDir string) error {
	result, err := downloader.DownloadFile(itemID)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	outputPath := filepath.Join(outputDir, result.Filename)
	if err := os.WriteFile(outputPath, result.Data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Fprintf(out, "Downloaded: %s\n", outputPath)
	return nil
}
