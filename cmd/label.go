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
	labelCmd.AddCommand(labelListCmd)
	labelCmd.AddCommand(labelGetCmd)
	labelCmd.AddCommand(labelAddCmd)
	labelCmd.AddCommand(labelRemoveCmd)
	labelCmd.AddCommand(labelCreateCmd)
	labelCmd.AddCommand(labelDeleteCmd)
	rootCmd.AddCommand(labelCmd)
}

var labelCmd = &cobra.Command{
	Use:   "label",
	Short: "Manage labels on library items",
}

var labelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available labels",
	Args:  cobra.NoArgs,
	RunE:  runLabelList,
}

var labelGetCmd = &cobra.Command{
	Use:   "get <item_id>",
	Short: "Get labels of a library item",
	Args:  cobra.ExactArgs(1),
	RunE:  runLabelGet,
}

var labelAddCmd = &cobra.Command{
	Use:   "add <item_id> <label_name>",
	Short: "Add a label to a library item",
	Args:  cobra.ExactArgs(2),
	RunE:  runLabelAdd,
}

var labelRemoveCmd = &cobra.Command{
	Use:   "remove <item_id> <label_name>",
	Short: "Remove a label from a library item",
	Args:  cobra.ExactArgs(2),
	RunE:  runLabelRemove,
}

var labelCreateCmd = &cobra.Command{
	Use:   "create <label_name>",
	Short: "Create a new label",
	Args:  cobra.ExactArgs(1),
	RunE:  runLabelCreate,
}

var labelDeleteCmd = &cobra.Command{
	Use:   "delete <label_name>",
	Short: "Delete a label",
	Args:  cobra.ExactArgs(1),
	RunE:  runLabelDelete,
}

func runLabelList(cmd *cobra.Command, args []string) error {
	client := api.NewClient(config.GetSession())
	return execLabelList(client, os.Stdout)
}

func runLabelGet(cmd *cobra.Command, args []string) error {
	client := api.NewClient(config.GetSession())
	return execLabelGet(client, os.Stdout, args[0])
}

func runLabelAdd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(config.GetSession())
	return execLabelAdd(client, os.Stdout, args[0], args[1])
}

func runLabelRemove(cmd *cobra.Command, args []string) error {
	client := api.NewClient(config.GetSession())
	return execLabelRemove(client, os.Stdout, args[0], args[1])
}

func runLabelCreate(cmd *cobra.Command, args []string) error {
	client := api.NewClient(config.GetSession())
	return execLabelCreate(client, os.Stdout, args[0])
}

func runLabelDelete(cmd *cobra.Command, args []string) error {
	client := api.NewClient(config.GetSession())
	return execLabelDelete(client, os.Stdout, args[0])
}

func execLabelList(fetcher LabelFetcher, out io.Writer) error {
	labels, err := fetcher.FetchLabels()
	if err != nil {
		return fmt.Errorf("failed to fetch labels: %w", err)
	}

	if len(labels) == 0 {
		fmt.Fprintln(out, "(no labels)")
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

func execLabelGet(getter ItemLabelGetter, out io.Writer, itemID string) error {
	labels, err := getter.GetItemLabelNames(itemID)
	if err != nil {
		return fmt.Errorf("failed to get labels: %w", err)
	}

	if len(labels) == 0 {
		fmt.Fprintln(out, "(no labels)")
		return nil
	}

	for _, name := range labels {
		fmt.Fprintln(out, name)
	}
	return nil
}

func execLabelAdd(adder LabelAdder, out io.Writer, itemID, labelName string) error {
	if err := adder.AddLabelByName(itemID, labelName); err != nil {
		return fmt.Errorf("failed to add label: %w", err)
	}
	fmt.Fprintf(out, "Label %q added to item %s\n", labelName, itemID)
	return nil
}

func execLabelDelete(deleter LabelDeleter, out io.Writer, labelName string) error {
	if err := deleter.DeleteLabel(labelName); err != nil {
		return fmt.Errorf("failed to delete label: %w", err)
	}
	fmt.Fprintf(out, "Label %q deleted\n", labelName)
	return nil
}

func execLabelCreate(creator LabelCreator, out io.Writer, labelName string) error {
	id, err := creator.CreateLabel(labelName)
	if err != nil {
		return fmt.Errorf("failed to create label: %w", err)
	}
	fmt.Fprintf(out, "Label %q created (ID: %s)\n", labelName, id)
	return nil
}

func execLabelRemove(remover LabelRemover, out io.Writer, itemID, labelName string) error {
	if err := remover.RemoveLabelByName(itemID, labelName); err != nil {
		return fmt.Errorf("failed to remove label: %w", err)
	}
	fmt.Fprintf(out, "Label %q removed from item %s\n", labelName, itemID)
	return nil
}
