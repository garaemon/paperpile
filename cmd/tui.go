package cmd

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/garaemon/paperpile/internal/api"
	"github.com/garaemon/paperpile/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(tuiCmd)
}

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive TUI mode",
	RunE:  runTUI,
}

func runTUI(cmd *cobra.Command, args []string) error {
	client := api.NewClient(config.GetSession())
	model := createInitialModel(client)
	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err := program.Run()
	return err
}

type viewType int

const (
	listView viewType = iota
	detailView
)

// tuiModel holds the state for the TUI application.
type tuiModel struct {
	fetcher      LibraryFetcher
	noteGetter   NoteGetter
	items        []api.LibraryItem
	list         list.Model
	viewport     viewport.Model
	activeView   viewType
	selectedItem *api.LibraryItem
	width        int
	height       int
	loading      bool
	err          error
}

// libraryItemEntry wraps a LibraryItem for the bubbles list interface.
type libraryItemEntry struct {
	item api.LibraryItem
}

// FilterValue returns the string used for filtering in the list.
func (e libraryItemEntry) FilterValue() string {
	return e.item.Title + " " + e.item.FormatFirstAuthor()
}

// Title returns the display title for the list item.
func (e libraryItemEntry) Title() string {
	return e.item.Title
}

// Description returns the description line for the list item.
func (e libraryItemEntry) Description() string {
	year := e.item.Year
	if year == "" {
		year = "-"
	}
	author := e.item.FormatFirstAuthor()
	if author == "" {
		author = "-"
	}
	return fmt.Sprintf("%s | %s", year, author)
}

// libraryFetchedMsg carries fetched library items after async loading.
type libraryFetchedMsg struct {
	items []api.LibraryItem
	err   error
}

func createInitialModel(client *api.Client) tuiModel {
	return tuiModel{
		fetcher:    client,
		noteGetter: client,
		loading:    true,
		activeView: listView,
	}
}

func fetchLibraryCmd(fetcher LibraryFetcher) tea.Cmd {
	return func() tea.Msg {
		items, err := fetcher.FetchLibrary()
		return libraryFetchedMsg{items: items, err: err}
	}
}

// Init starts the async library fetch.
func (m tuiModel) Init() tea.Cmd {
	return fetchLibraryCmd(m.fetcher)
}

var (
	titleStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69"))
	detailKeyStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("243"))
	detailBodyStyle = lipgloss.NewStyle().Padding(1, 2)
)
