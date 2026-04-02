package cmd

import "github.com/garaemon/paperpile/internal/api"

// LibraryFetcher fetches library items from Paperpile.
type LibraryFetcher interface {
	FetchLibrary() ([]api.LibraryItem, error)
}

// UserFetcher fetches the current user info.
type UserFetcher interface {
	FetchCurrentUser() (*api.UserInfo, error)
}

// ItemTrasher moves an item to the trash.
type ItemTrasher interface {
	TrashItem(itemID string) error
}

// PDFUploader uploads a PDF file to Paperpile.
type PDFUploader interface {
	UploadPDF(filePath string, importDuplicates bool) (*api.ImportTask, error)
}

// FileAttacher attaches a PDF to an existing library item.
type FileAttacher interface {
	AttachFile(itemID, filePath string) (string, error)
}

// NoteGetter retrieves a note from a library item.
type NoteGetter interface {
	GetNote(itemID string) (string, error)
}

// NoteUpdater updates a note on a library item.
type NoteUpdater interface {
	UpdateNote(itemID, note string) error
}

// LabelFetcher fetches all labels from Paperpile.
type LabelFetcher interface {
	FetchLabels() ([]api.Collection, error)
}

// ItemLabelGetter retrieves label names for a library item.
type ItemLabelGetter interface {
	GetItemLabelNames(itemID string) ([]string, error)
}

// TagAdder adds a tag to a library item by name.
type TagAdder interface {
	AddLabelByName(itemID, tagName string) error
}

// TagRemover removes a tag from a library item by name.
type TagRemover interface {
	RemoveLabelByName(itemID, tagName string) error
}

// TagCreator creates a new tag (label).
type TagCreator interface {
	CreateLabel(name string) (string, error)
}

// TagDeleter deletes a tag (label) by name.
type TagDeleter interface {
	DeleteLabel(name string) error
}
