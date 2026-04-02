package cmd

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/garaemon/paperpile/internal/api"
	"github.com/garaemon/paperpile/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	tagCmd.AddCommand(tagListCmd)
	tagCmd.AddCommand(tagGetCmd)
	tagCmd.AddCommand(tagAddCmd)
	tagCmd.AddCommand(tagRemoveCmd)
	tagCmd.AddCommand(tagCreateCmd)
	tagCmd.AddCommand(tagDeleteCmd)
	rootCmd.AddCommand(tagCmd)
}

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Manage tags (labels) on library items",
}

var tagListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available tags",
	Args:  cobra.NoArgs,
	RunE:  runTagList,
}

var tagGetCmd = &cobra.Command{
	Use:   "get <item_id>",
	Short: "Get tags of a library item",
	Args:  cobra.ExactArgs(1),
	RunE:  runTagGet,
}

var tagAddCmd = &cobra.Command{
	Use:   "add <item_id> <tag_name>",
	Short: "Add a tag to a library item",
	Args:  cobra.ExactArgs(2),
	RunE:  runTagAdd,
}

var tagRemoveCmd = &cobra.Command{
	Use:   "remove <item_id> <tag_name>",
	Short: "Remove a tag from a library item",
	Args:  cobra.ExactArgs(2),
	RunE:  runTagRemove,
}

var tagCreateCmd = &cobra.Command{
	Use:   "create <tag_name>",
	Short: "Create a new tag",
	Args:  cobra.ExactArgs(1),
	RunE:  runTagCreate,
}

var tagDeleteCmd = &cobra.Command{
	Use:   "delete <tag_name>",
	Short: "Delete a tag",
	Args:  cobra.ExactArgs(1),
	RunE:  runTagDelete,
}

func runTagList(cmd *cobra.Command, args []string) error {
	client := api.NewClient(config.GetSession())
	return execTagList(client, os.Stdout)
}

func runTagGet(cmd *cobra.Command, args []string) error {
	client := api.NewClient(config.GetSession())
	return execTagGet(client, os.Stdout, args[0])
}

func runTagAdd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(config.GetSession())
	return execTagAdd(client, os.Stdout, args[0], args[1])
}

func runTagRemove(cmd *cobra.Command, args []string) error {
	client := api.NewClient(config.GetSession())
	return execTagRemove(client, os.Stdout, args[0], args[1])
}

func runTagCreate(cmd *cobra.Command, args []string) error {
	client := api.NewClient(config.GetSession())
	return execTagCreate(client, os.Stdout, args[0])
}

func runTagDelete(cmd *cobra.Command, args []string) error {
	client := api.NewClient(config.GetSession())
	return execTagDelete(client, os.Stdout, args[0])
}

func execTagList(fetcher LabelFetcher, out io.Writer) error {
	labels, err := fetcher.FetchLabels()
	if err != nil {
		return fmt.Errorf("failed to fetch labels: %w", err)
	}

	if len(labels) == 0 {
		fmt.Fprintln(out, "(no tags)")
		return nil
	}

	w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tCOUNT")
	for _, label := range labels {
		fmt.Fprintf(w, "%s\t%s\t%d\n", label.ID, label.Name, label.Count)
	}
	w.Flush()
	return nil
}

func execTagGet(getter ItemLabelGetter, out io.Writer, itemID string) error {
	labels, err := getter.GetItemLabelNames(itemID)
	if err != nil {
		return fmt.Errorf("failed to get tags: %w", err)
	}

	if len(labels) == 0 {
		fmt.Fprintln(out, "(no tags)")
		return nil
	}

	for _, name := range labels {
		fmt.Fprintln(out, name)
	}
	return nil
}

func execTagAdd(adder TagAdder, out io.Writer, itemID, tagName string) error {
	if err := adder.AddLabelByName(itemID, tagName); err != nil {
		return fmt.Errorf("failed to add tag: %w", err)
	}
	fmt.Fprintf(out, "Tag %q added to item %s\n", tagName, itemID)
	return nil
}

func execTagDelete(deleter TagDeleter, out io.Writer, tagName string) error {
	if err := deleter.DeleteLabel(tagName); err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}
	fmt.Fprintf(out, "Tag %q deleted\n", tagName)
	return nil
}

func execTagCreate(creator TagCreator, out io.Writer, tagName string) error {
	id, err := creator.CreateLabel(tagName)
	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}
	fmt.Fprintf(out, "Tag %q created (ID: %s)\n", tagName, id)
	return nil
}

func execTagRemove(remover TagRemover, out io.Writer, itemID, tagName string) error {
	if err := remover.RemoveLabelByName(itemID, tagName); err != nil {
		return fmt.Errorf("failed to remove tag: %w", err)
	}
	fmt.Fprintf(out, "Tag %q removed from item %s\n", tagName, itemID)
	return nil
}
