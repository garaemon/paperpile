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

// FileDownloader downloads attachments for a library item.
type FileDownloader interface {
	FindAttachmentsForItem(itemID string) ([]api.AttachmentInfo, error)
	DownloadAttachment(attachment api.AttachmentInfo) (*api.DownloadResult, error)
}

// LabelFetcher fetches all labels from Paperpile.
type LabelFetcher interface {
	FetchLabels() ([]api.Collection, error)
}

// ItemLabelGetter retrieves label names for a library item.
type ItemLabelGetter interface {
	GetItemLabelNames(itemID string) ([]string, error)
}

// LabelCreator creates a new label.
type LabelCreator interface {
	CreateLabel(name string) (string, error)
}

// LabelUnassigner removes a label from a library item.
type LabelUnassigner interface {
	UnassignLabel(itemID, labelName string) error
}

// LabelAssigner assigns a label to a library item.
type LabelAssigner interface {
	AssignLabel(itemID, labelName string) error
}

// LabelDeleter deletes a label.
type LabelDeleter interface {
	DeleteLabel(labelName string) error
}
