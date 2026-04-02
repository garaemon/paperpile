package cmd

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/garaemon/paperpile/internal/api"
)

// Update handles all TUI messages and returns the updated model.
func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return handleKeyMsg(m, msg)
	case tea.WindowSizeMsg:
		return handleWindowSizeMsg(m, msg)
	case libraryFetchedMsg:
		return handleLibraryFetchedMsg(m, msg)
	}
	return delegateUpdate(m, msg)
}

func handleKeyMsg(m tuiModel, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.activeView {
	case listView:
		return handleListKeyMsg(m, msg)
	case detailView:
		return handleDetailKeyMsg(m, msg)
	}
	return m, nil
}

func handleListKeyMsg(m tuiModel, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// When the list is filtering, let it handle all keys except ctrl+c.
	if m.list.FilterState() == list.Filtering {
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "enter":
		return switchToDetailView(m)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func handleDetailKeyMsg(m tuiModel, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.activeView = listView
		m.selectedItem = nil
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func handleWindowSizeMsg(m tuiModel, msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height

	if m.list.Items() != nil {
		m.list.SetSize(msg.Width, msg.Height)
	}
	if m.activeView == detailView {
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 2 // leave room for header/footer
	}
	return m, nil
}

func handleLibraryFetchedMsg(m tuiModel, msg libraryFetchedMsg) (tea.Model, tea.Cmd) {
	m.loading = false
	if msg.err != nil {
		m.err = msg.err
		return m, tea.Quit
	}

	m.items = filterActiveItems(msg.items)
	m.list = buildListModel(m.items, m.width, m.height)
	return m, nil
}

func delegateUpdate(m tuiModel, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.activeView {
	case listView:
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	case detailView:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
	return m, nil
}

func switchToDetailView(m tuiModel) (tea.Model, tea.Cmd) {
	selected, ok := m.list.SelectedItem().(libraryItemEntry)
	if !ok {
		return m, nil
	}

	item := selected.item
	m.selectedItem = &item
	m.activeView = detailView

	detailContent := renderDetailContent(&item, m.noteGetter, m.width)
	m.viewport = viewport.New(m.width, m.height-2)
	m.viewport.SetContent(detailContent)
	return m, nil
}

func filterActiveItems(items []api.LibraryItem) []api.LibraryItem {
	var activeItems []api.LibraryItem
	for _, item := range items {
		if item.Trashed == 0 {
			activeItems = append(activeItems, item)
		}
	}
	return activeItems
}
