# paperpile

An **unofficial** command-line tool to upload, list, and delete references in [Paperpile](https://paperpile.com/).

Paperpile has no public API. This tool works by reverse-engineering the web app's internal endpoints.

## Installation

Requires Go 1.26+.

```bash
go install github.com/garaemon/paperpile@latest
```

Or build from source:

```bash
git clone https://github.com/garaemon/paperpile.git
cd paperpile
make build
```

The binary is named `paperpile`.

## Authentication

Paperpile uses session cookies for authentication. This CLI obtains the session via a bookmarklet flow.

```bash
paperpile login
```

1. The CLI starts a local HTTP server and opens a setup page in your browser.
2. Drag the **"Paperpile: Send to CLI"** bookmarklet to your bookmarks bar (one-time setup).
3. Navigate to [app.paperpile.com](https://app.paperpile.com/) and log in.
4. Click the bookmarklet. It extracts the session cookie and sends it to the CLI.
5. The CLI verifies the session and saves it to `~/.config/paperpile/config.yaml`.

When the session expires, run `paperpile login` again and click the bookmarklet.

### Options

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `18080` | Local server port for receiving session |

## Commands

### `me` - Show current user info

```bash
paperpile me
```

Output:
```
Name:  John Doe
Email: john@example.com
ID:    68CE82F6807411EA9B68A87FDE8EC746
```

### `upload` - Upload a PDF

```bash
paperpile upload paper.pdf
```

Paperpile automatically extracts metadata (title, authors, journal, etc.) from the uploaded PDF.

| Flag | Description |
|------|-------------|
| `--allow-duplicates` | Import even if a duplicate already exists |

### `list` - List library items

```bash
paperpile list
```

Output is a tab-separated table with columns: ID, Year, First Author, Title.

| Flag | Description |
|------|-------------|
| `--trashed` | Include trashed items in the output |

### `delete` - Move an item to trash

```bash
paperpile delete <item_id>
```

The `item_id` can be found via `paperpile list`.

### `attach` - Attach a PDF to an existing item

```bash
paperpile attach <item_id> paper.pdf
```

Attaches a PDF file to a library item that does not yet have a PDF, or adds an additional file to an existing item.

### `note get` - Get the note of a library item

```bash
paperpile note get <item_id>
```

Prints the note text attached to the item. If no note exists, prints `(no note)`.

### `note set` - Set the note of a library item

```bash
paperpile note set <item_id> <text>...
```

Sets or replaces the note on a library item. Multiple arguments are joined with spaces.

### `label list` - List all labels

```bash
paperpile label list
```

Lists all labels in your Paperpile library with their IDs, names, and item counts.

### `label get` - Get labels of a library item

```bash
paperpile label get <item_id>
```

Prints the label names assigned to the item. If no labels are assigned, prints `(no labels)`.

### `label create` - Create a new label

```bash
paperpile label create <label_name>
```

Creates a new label in your Paperpile library.

### `label unassign` - Remove a label from a library item

```bash
paperpile label unassign <item_id> <label_name>
```

Removes a label from a library item.

## Configuration

Session credentials are stored in `~/.config/paperpile/config.yaml`.

## License

MIT
