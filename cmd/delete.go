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
	rootCmd.AddCommand(deleteCmd)
}

var deleteCmd = &cobra.Command{
	Use:   "delete <item_id>",
	Short: "Move a library item to trash",
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

func runDelete(cmd *cobra.Command, args []string) error {
	client := api.NewClient(config.GetSession())
	return execDelete(client, os.Stdout, args[0])
}

func execDelete(trasher ItemTrasher, out io.Writer, itemID string) error {
	fmt.Fprintf(out, "Trashing item %s ...\n", itemID)

	if err := trasher.TrashItem(itemID); err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	fmt.Fprintln(out, "Done! Item moved to trash.")
	return nil
}
