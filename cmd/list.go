package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/garaemon/paperpile-cli/internal/api"
	"github.com/garaemon/paperpile-cli/internal/config"
	"github.com/spf13/cobra"
)

var listTrashed bool

func init() {
	listCmd.Flags().BoolVar(&listTrashed, "trashed", false, "Include trashed items")
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List library items",
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	client := api.NewClient(config.GetSession())
	items, err := client.FetchLibrary()
	if err != nil {
		return fmt.Errorf("failed to fetch library: %w", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tYEAR\tFIRST AUTHOR\tTITLE")

	for _, item := range items {
		if item.Trashed != 0 && !listTrashed {
			continue
		}

		title := item.Title
		if len(title) > 80 {
			title = title[:77] + "..."
		}

		year := item.Year
		if year == "" {
			year = "-"
		}

		author := item.FormatFirstAuthor()
		if author == "" {
			author = "-"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", item.ID, year, author, title)
	}

	w.Flush()
	return nil
}
