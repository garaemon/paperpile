package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/garaemon/paperpile/internal/api"
	"github.com/garaemon/paperpile/internal/convert"
)

// View renders the current TUI state to a string.
func (m tuiModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress any key to exit.", m.err)
	}
	if m.loading {
		return "\n  Loading library..."
	}

	switch m.activeView {
	case detailView:
		return renderDetailView(m)
	default:
		return m.list.View()
	}
}

func renderDetailView(m tuiModel) string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("69")).
		Padding(0, 1)

	header := headerStyle.Render("Paper Details (q/esc to go back, arrows to scroll)")

	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243"))
	scrollPercent := fmt.Sprintf(" %3.f%%", m.viewport.ScrollPercent()*100)
	footer := footerStyle.Render(scrollPercent)

	return header + "\n" + m.viewport.View() + "\n" + footer
}

func buildListModel(items []api.LibraryItem, width, height int) list.Model {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = libraryItemEntry{item: item}
	}

	paperList := list.New(listItems, list.NewDefaultDelegate(), width, height)
	paperList.Title = "Paperpile Library"
	paperList.SetShowStatusBar(true)
	paperList.SetFilteringEnabled(true)
	paperList.Styles.Title = titleStyle
	return paperList
}

func renderDetailContent(item *api.LibraryItem, noteGetter NoteGetter, width int) string {
	var sb strings.Builder
	contentWidth := width - 4 // account for padding
	if contentWidth < 20 {
		contentWidth = 20
	}

	appendDetailField(&sb, "Title", item.Title)
	appendDetailField(&sb, "Authors", formatAllAuthors(item.Author))
	appendDetailField(&sb, "Year", valueOrDash(item.Year))
	appendDetailField(&sb, "Journal", valueOrDash(item.Journal))
	appendDetailField(&sb, "Type", valueOrDash(item.Pubtype))
	appendDetailField(&sb, "Citation Key", valueOrDash(item.Citekey))
	appendDetailField(&sb, "ID", item.ID)

	noteContent := resolveNoteContent(item, noteGetter)
	appendDetailField(&sb, "Note", noteContent)

	return detailBodyStyle.Render(sb.String())
}

func appendDetailField(sb *strings.Builder, label, value string) {
	styledLabel := detailKeyStyle.Render(label + ":")
	sb.WriteString(styledLabel + " " + value + "\n\n")
}

func formatAllAuthors(authors []api.Author) string {
	if len(authors) == 0 {
		return "-"
	}
	names := make([]string, len(authors))
	for i, a := range authors {
		names[i] = formatSingleAuthor(a)
	}
	return strings.Join(names, ", ")
}

func formatSingleAuthor(a api.Author) string {
	if a.First != "" && a.Last != "" {
		return a.First + " " + a.Last
	}
	if a.Formatted != "" {
		return a.Formatted
	}
	if a.Last != "" {
		return a.Last
	}
	return a.First
}

func resolveNoteContent(item *api.LibraryItem, noteGetter NoteGetter) string {
	// Prefer fetching the note via API for the latest content.
	noteHTML := item.Notes
	if noteGetter != nil {
		fetched, err := noteGetter.GetNote(item.ID)
		if err == nil && fetched != "" {
			noteHTML = fetched
		}
	}

	if noteHTML == "" {
		return "(no note)"
	}

	md, err := convert.HTMLToMarkdown(noteHTML)
	if err != nil {
		return noteHTML
	}
	return md
}

func valueOrDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
