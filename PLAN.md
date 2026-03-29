# Paperpile CLI Development Plan

A CLI tool for uploading, listing, and deleting PDFs in Paperpile.

## 1. Goal

Automate adding (PDF upload), listing, and deleting references in Paperpile to streamline workflow.

## 2. Tech Stack

- **Language**: Go (static binary distribution, suitable for CLI)
- **CLI framework**: `spf13/cobra`
- **Config management**: `spf13/viper` (for storing auth credentials)
- **HTTP client**: `net/http` (standard library)

## 3. API Analysis

Paperpile has no public API. The following endpoints were discovered by reverse-engineering the web app.

### 3.1 Authentication

Paperpile uses a Perl/Plack-based backend. Authentication relies on a single session cookie:

- **`plack_session`**: The only required cookie for API authentication.
- Other cookies (`AWSALB`, `AWSALBCORS`, `statsiguuid`, `_ga`, `mp_*`, `intercom-*`) are for load balancing or analytics and are not required.
- Session lifetime is unknown (needs testing — may last hours to days).

**Required headers for all API calls:**
- `Cookie: plack_session=<session_value>`
- `Origin: https://app.paperpile.com`
- `Referer: https://app.paperpile.com/`

### 3.2 REST Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/users/me` | GET | Get current user info (name, email, ID). Used for session validation. |
| `/api/library` | GET | Get all library items as JSON array. Each item has `_id`, `title`, `author`, `year`, `journal`, `trashed`, etc. |
| `/api/import/files` | POST | Initiate a file upload. Returns S3 presigned URL and task/subtask UUIDs. |
| `/api/tasks/{taskId}/subtasks` | PATCH | Notify upload completion (`status: "uploaded"`). |

### 3.3 Upload Flow

Upload consists of 3 steps:

| Step | Endpoint | Method | Description |
|------|----------|--------|-------------|
| 1 | `api.paperpile.com/api/import/files` | POST | Initiate upload. Send file name(s) → receive S3 presigned URL and task ID |
| 2 | `*.s3-global.amazonaws.com/import/{userId}/{taskId}/...` | PUT | Upload PDF binary to S3 presigned URL (no cookie needed; URL itself is auth) |
| 3 | `api.paperpile.com/api/tasks/{taskId}/subtasks` | PATCH | Notify upload completion (`status: "uploaded"`) |

**Step 1 request body:**
```json
{
  "files": [{"names": ["example.pdf"], "type": "file_upload"}],
  "isPartialImport": false,
  "collections": [],
  "keepFolderOrganization": false,
  "preserveCitationKey": false,
  "importDuplicates": false
}
```

**Step 1 response:** Returns S3 presigned URL (with `X-Amz-*` query params, expires in 3600s), task/subtask UUIDs, and `uploadUrl` field.

**Step 3 request body:**
```json
{
  "subtasks": ["<subtask-uuid>"],
  "status": "uploaded"
}
```

**S3 upload (Step 2):**
- No cookies required (presigned URL contains auth)
- `Content-Type: application/pdf`
- `Content-Length` header required
- Request body is the raw PDF binary

### 3.4 Sync API (for mutations: delete, update, etc.)

The web app uses an **offline-first architecture** with Dexie (IndexedDB) and a Service Worker.
Mutations (trash, star, edit metadata, etc.) are NOT done via REST — they go through the **Sync API**.

**Endpoint:** `POST /api/sync?v=3`

**Request body:**
```json
{
  "syncClientId": "paperpile-cli",
  "last_server_sync": 1774737619.0,
  "clientChanges": [
    {
      "mcollection": "Library",
      "action": "update",
      "id": "<item-uuid>",
      "timestamp": 1774737619.0,
      "fields": ["trashed", "updated"],
      "data": {"trashed": 1, "updated": 1774737619.0}
    }
  ]
}
```

**Response:**
```json
{
  "syncStartTime": 1774737619.691,
  "syncSession": "<uuid>",
  "totalServerChanges": 0,
  "lastClientSync": 1774737619.466
}
```

**Key details:**
- `mcollection` (not `collection`) specifies the data collection: `Library`, `Attachments`, etc.
- `action`: `update`, `insert`, `remove`
- `fields`: array of field names being changed
- `data`: the actual new values for those fields
- `syncClientId`: arbitrary client identifier
- `last_server_sync`: timestamp to avoid receiving full server history
- The same endpoint can be used without `clientChanges` (just `syncClientId`) to pull server changes (initial sync).

The WebSocket (`/socket.io/`) is used only for server-to-client push notifications, not for client mutations.

### 3.5 User ID

The user ID (`68CE82F6807411EA9B68A87FDE8EC746` format) is returned by `/api/users/me` and appears in S3 paths.

### 3.6 Endpoints Not Yet Analyzed

- **Attach file to existing item**: Possibly via `/api/attachments` endpoint.

## 4. Authentication Design

### Approach: Bookmarklet + Local HTTP Server

1. User runs `paperpile-cli login`.
2. CLI starts a local HTTP server on `localhost:18080` with a setup page.
3. CLI opens the setup page in the browser.
4. User drags a bookmarklet ("Paperpile: Send to CLI") to their bookmarks bar (one-time setup).
5. User navigates to `app.paperpile.com` and clicks the bookmarklet.
6. Bookmarklet extracts `plack_session` from `document.cookie` and POSTs it to `localhost:18080/callback`.
7. CLI verifies the session via `/api/users/me`, saves it to `~/.config/paperpile-cli/config.yaml`, and shuts down.

**Session refresh:** When the session expires, user re-runs `paperpile-cli login` and clicks the bookmarklet again.

## 5. Development Phases

### Phase 1: Project Setup & Auth
- [x] API research (upload flow)
- [x] `go mod init`, Cobra setup
- [x] Implement `login` command (local HTTP server + bookmarklet flow)
- [x] Config file management (`~/.config/paperpile-cli/config.yaml`)
- [x] Implement `me` command (session validation)

### Phase 2: Upload Command
- [x] Implement Step 1: `POST /api/import/files` (get presigned URL)
- [x] Implement Step 2: `PUT` PDF to S3
- [x] Implement Step 3: `PATCH /api/tasks/{taskId}/subtasks` (notify completion)
- [x] End-to-end upload test
- [x] `--allow-duplicates` flag

### Phase 3: List & Delete
- [x] Discover list endpoint (`GET /api/library`)
- [x] Implement `list` command (with `--trashed` filter)
- [x] Discover Sync API (`POST /api/sync?v=3`)
- [x] Implement `delete` command via Sync API

### Phase 4: Attach & Polish
- [x] Analyze file attachment endpoint
- [x] Implement `attach` command
- [ ] Error handling and retry logic
- [ ] Session expiry detection and re-login prompt

## 6. Command Design

```
paperpile-cli login                              # Start auth flow (bookmarklet)
paperpile-cli me                                 # Show current user info
paperpile-cli upload <file_path>                 # Upload a PDF
paperpile-cli upload --allow-duplicates <file>   # Upload even if duplicate exists
paperpile-cli list                               # List library items
paperpile-cli list --trashed                     # Include trashed items
paperpile-cli delete <item_id>                   # Move item to trash
```
