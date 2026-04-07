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
	labelCmd.AddCommand(labelCreateCmd)
	labelCmd.AddCommand(labelUnassignCmd)
	labelCmd.AddCommand(labelAssignCmd)
	labelCmd.AddCommand(labelDeleteCmd)
	rootCmd.AddCommand(labelCmd)
}

var labelCmd = &cobra.Command{
	Use:   "label",
	Short: "Manage labels on library items",
}

var labelGetCmd = &cobra.Command{
	Use:   "get <item_id>",
	Short: "Get labels of a library item",
	Args:  cobra.ExactArgs(1),
	RunE:  runLabelGet,
}

var labelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available labels",
	Args:  cobra.NoArgs,
	RunE:  runLabelList,
}

func runLabelGet(cmd *cobra.Command, args []string) error {
	client := api.NewClient(config.GetSession())
	return execLabelGet(client, os.Stdout, args[0])
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

func runLabelList(cmd *cobra.Command, args []string) error {
	client := api.NewClient(config.GetSession())
	return execLabelList(client, os.Stdout)
}

var labelCreateCmd = &cobra.Command{
	Use:   "create <label_name>",
	Short: "Create a new label",
	Args:  cobra.ExactArgs(1),
	RunE:  runLabelCreate,
}

func runLabelCreate(cmd *cobra.Command, args []string) error {
	client := api.NewClient(config.GetSession())
	return execLabelCreate(client, os.Stdout, args[0])
}

func execLabelCreate(creator LabelCreator, out io.Writer, labelName string) error {
	id, err := creator.CreateLabel(labelName)
	if err != nil {
		return fmt.Errorf("failed to create label: %w", err)
	}
	fmt.Fprintf(out, "Label %q created (ID: %s)\n", labelName, id)
	return nil
}

var labelUnassignCmd = &cobra.Command{
	Use:   "unassign <item_id> <label_name>",
	Short: "Remove a label from a library item",
	Args:  cobra.ExactArgs(2),
	RunE:  runLabelUnassign,
}

func runLabelUnassign(cmd *cobra.Command, args []string) error {
	client := api.NewClient(config.GetSession())
	return execLabelUnassign(client, os.Stdout, args[0], args[1])
}

func execLabelUnassign(unassigner LabelUnassigner, out io.Writer, itemID, labelName string) error {
	if err := unassigner.UnassignLabel(itemID, labelName); err != nil {
		return fmt.Errorf("failed to unassign label: %w", err)
	}
	fmt.Fprintf(out, "Label %q unassigned from item %s\n", labelName, itemID)
	return nil
}

var labelAssignCmd = &cobra.Command{
	Use:   "assign <item_id> <label_name>",
	Short: "Assign a label to a library item",
	Args:  cobra.ExactArgs(2),
	RunE:  runLabelAssign,
}

func runLabelAssign(cmd *cobra.Command, args []string) error {
	client := api.NewClient(config.GetSession())
	return execLabelAssign(client, os.Stdout, args[0], args[1])
}

func execLabelAssign(assigner LabelAssigner, out io.Writer, itemID, labelName string) error {
	if err := assigner.AssignLabel(itemID, labelName); err != nil {
		return fmt.Errorf("failed to assign label: %w", err)
	}
	fmt.Fprintf(out, "Label %q assigned to item %s\n", labelName, itemID)
	return nil
}

var labelDeleteCmd = &cobra.Command{
	Use:   "delete <label_name>",
	Short: "Delete a label",
	Args:  cobra.ExactArgs(1),
	RunE:  runLabelDelete,
}

func runLabelDelete(cmd *cobra.Command, args []string) error {
	client := api.NewClient(config.GetSession())
	return execLabelDelete(client, os.Stdout, args[0])
}

func execLabelDelete(deleter LabelDeleter, out io.Writer, labelName string) error {
	if err := deleter.DeleteLabel(labelName); err != nil {
		return fmt.Errorf("failed to delete label: %w", err)
	}
	fmt.Fprintf(out, "Label %q deleted\n", labelName)
	return nil
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
