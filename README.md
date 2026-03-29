# paperpile-cli

A command-line tool to upload, list, and delete references in [Paperpile](https://paperpile.com/).

Paperpile has no public API. This tool works by reverse-engineering the web app's internal endpoints.

## Installation

Requires Go 1.26+.

```bash
go install github.com/garaemon/paperpile-cli@latest
```

Or build from source:

```bash
git clone https://github.com/garaemon/paperpile-cli.git
cd paperpile-cli
make build
```

## Authentication

Paperpile uses session cookies for authentication. This CLI obtains the session via a bookmarklet flow.

```bash
paperpile-cli login
```

1. The CLI starts a local HTTP server and opens a setup page in your browser.
2. Drag the **"Paperpile: Send to CLI"** bookmarklet to your bookmarks bar (one-time setup).
3. Navigate to [app.paperpile.com](https://app.paperpile.com/) and log in.
4. Click the bookmarklet. It extracts the session cookie and sends it to the CLI.
5. The CLI verifies the session and saves it to `~/.config/paperpile-cli/config.yaml`.

When the session expires, run `paperpile-cli login` again and click the bookmarklet.

### Options

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `18080` | Local server port for receiving session |

## Commands

### `me` - Show current user info

```bash
paperpile-cli me
```

Output:
```
Name:  John Doe
Email: john@example.com
ID:    68CE82F6807411EA9B68A87FDE8EC746
```

### `upload` - Upload a PDF

```bash
paperpile-cli upload paper.pdf
```

Paperpile automatically extracts metadata (title, authors, journal, etc.) from the uploaded PDF.

| Flag | Description |
|------|-------------|
| `--allow-duplicates` | Import even if a duplicate already exists |

### `list` - List library items

```bash
paperpile-cli list
```

Output is a tab-separated table with columns: ID, Year, First Author, Title.

| Flag | Description |
|------|-------------|
| `--trashed` | Include trashed items in the output |

### `delete` - Move an item to trash

```bash
paperpile-cli delete <item_id>
```

The `item_id` can be found via `paperpile-cli list`.

### `attach` - Attach a PDF to an existing item

```bash
paperpile-cli attach <item_id> paper.pdf
```

Attaches a PDF file to a library item that does not yet have a PDF, or adds an additional file to an existing item.

## Configuration

Session credentials are stored in `~/.config/paperpile-cli/config.yaml`.

## License

MIT
