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
