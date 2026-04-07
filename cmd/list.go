package cmd

import (
	"fmt"
	"io"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/garaemon/paperpile/internal/api"
	"github.com/garaemon/paperpile/internal/config"
	"github.com/spf13/cobra"
)

var listTrashed bool
var listSort string

func init() {
	listCmd.Flags().BoolVar(&listTrashed, "trashed", false, "Include trashed items")
	listCmd.Flags().StringVar(&listSort, "sort", "created", "Sort order: created (newest first) or none")
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List library items",
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	client := api.NewClient(config.GetSession())
	return execList(client, os.Stdout, listTrashed, listSort)
}

func execList(fetcher LibraryFetcher, out io.Writer, includeTrashed bool, sortBy string) error {
	items, err := fetcher.FetchLibrary()
	if err != nil {
		return fmt.Errorf("failed to fetch library: %w", err)
	}

	if sortBy == "created" {
		sort.Slice(items, func(i, j int) bool {
			return items[i].Created > items[j].Created
		})
	}

	w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tYEAR\tFIRST AUTHOR\tTITLE")

	for _, item := range items {
		if item.Trashed != 0 && !includeTrashed {
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
