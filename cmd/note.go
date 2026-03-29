package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/garaemon/paperpile-cli/internal/api"
	"github.com/garaemon/paperpile-cli/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	noteCmd.AddCommand(noteGetCmd)
	noteCmd.AddCommand(noteSetCmd)
	rootCmd.AddCommand(noteCmd)
}

var noteCmd = &cobra.Command{
	Use:   "note",
	Short: "Manage notes on library items",
}

var noteGetCmd = &cobra.Command{
	Use:   "get <item_id>",
	Short: "Get the note of a library item",
	Args:  cobra.ExactArgs(1),
	RunE:  runNoteGet,
}

var noteSetCmd = &cobra.Command{
	Use:   "set <item_id> <text>...",
	Short: "Set the note of a library item",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runNoteSet,
}

func runNoteGet(cmd *cobra.Command, args []string) error {
	client := api.NewClient(config.GetSession())
	return execNoteGet(client, os.Stdout, args[0])
}

func runNoteSet(cmd *cobra.Command, args []string) error {
	client := api.NewClient(config.GetSession())
	noteText := strings.Join(args[1:], " ")
	return execNoteSet(client, os.Stdout, args[0], noteText)
}

func execNoteGet(getter NoteGetter, out io.Writer, itemID string) error {
	note, err := getter.GetNote(itemID)
	if err != nil {
		return fmt.Errorf("failed to get note: %w", err)
	}
	if note == "" {
		fmt.Fprintln(out, "(no note)")
		return nil
	}
	fmt.Fprintln(out, note)
	return nil
}

func execNoteSet(updater NoteUpdater, out io.Writer, itemID, note string) error {
	if err := updater.UpdateNote(itemID, note); err != nil {
		return fmt.Errorf("failed to set note: %w", err)
	}
	fmt.Fprintf(out, "Note updated for item %s\n", itemID)
	return nil
}
