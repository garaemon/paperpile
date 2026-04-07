package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/garaemon/paperpile/internal/api"
	"github.com/garaemon/paperpile/internal/config"
	"github.com/garaemon/paperpile/internal/markup"
	"github.com/spf13/cobra"
)

func init() {
	noteGetCmd.Flags().BoolVar(&noteGetMarkdown, "markdown", false, "Display note as Markdown (converted from HTML)")
	noteSetCmd.Flags().BoolVar(&noteSetMarkdown, "markdown", false, "Accept Markdown input and convert to HTML before saving")
	noteCmd.AddCommand(noteGetCmd)
	noteCmd.AddCommand(noteSetCmd)
	rootCmd.AddCommand(noteCmd)
}

var noteGetMarkdown bool
var noteSetMarkdown bool

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
	return execNoteGet(client, os.Stdout, args[0], noteGetMarkdown)
}

func runNoteSet(cmd *cobra.Command, args []string) error {
	client := api.NewClient(config.GetSession())
	noteText := strings.Join(args[1:], " ")
	return execNoteSet(client, os.Stdout, args[0], noteText, noteSetMarkdown)
}

func execNoteGet(getter NoteGetter, out io.Writer, itemID string, markdown bool) error {
	note, err := getter.GetNote(itemID)
	if err != nil {
		return fmt.Errorf("failed to get note: %w", err)
	}
	if note == "" {
		fmt.Fprintln(out, "(no note)")
		return nil
	}
	if markdown {
		md, err := markup.HTMLToMarkdown(note)
		if err != nil {
			return fmt.Errorf("failed to convert note to markdown: %w", err)
		}
		fmt.Fprintln(out, md)
		return nil
	}
	fmt.Fprintln(out, note)
	return nil
}

func execNoteSet(updater NoteUpdater, out io.Writer, itemID, note string, markdown bool) error {
	if markdown {
		note = markup.MarkdownToHTML(note)
	}
	if err := updater.UpdateNote(itemID, note); err != nil {
		return fmt.Errorf("failed to set note: %w", err)
	}
	fmt.Fprintf(out, "Note updated for item %s\n", itemID)
	return nil
}
