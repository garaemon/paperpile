package cmd

import (
	"fmt"

	"github.com/garaemon/paperpile-cli/internal/api"
	"github.com/garaemon/paperpile-cli/internal/config"
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
	itemID := args[0]
	client := api.NewClient(config.GetSession())

	fmt.Printf("Trashing item %s ...\n", itemID)

	if err := client.TrashItem(itemID); err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	fmt.Println("Done! Item moved to trash.")
	return nil
}
