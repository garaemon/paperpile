# Paperpile CLI Development Plan

A CLI tool for uploading and deleting PDFs to/from Paperpile.

## 1. Goal

Automate adding (PDF upload) and deleting references in Paperpile to streamline workflow.

## 2. Tech Stack

- **Language**: Go (static binary distribution, suitable for CLI)
- **CLI framework**: `spf13/cobra`
- **Config management**: `spf13/viper` (for storing auth credentials)
- **HTTP client**: `net/http` (standard library)

## 3. API Analysis

### 3.1 Authentication

Paperpile uses a Perl/Plack-based backend. Authentication relies on a single session cookie:

- **`plack_session`**: The only required cookie for API authentication.
- Other cookies (`AWSALB`, `AWSALBCORS`, `statsiguuid`, `_ga`, `mp_*`, `intercom-*`) are for load balancing or analytics and are not required.
- Session lifetime is unknown (needs testing — may last hours to days).

### 3.2 Upload Flow

Upload consists of 3 steps (+ 1 optional analytics step):

| Step | Endpoint | Method | Description |
|------|----------|--------|-------------|
| 1 | `api.paperpile.com/api/import/files` | POST | Initiate upload. Send file name(s) → receive S3 presigned URL and task ID |
| 2 | `*.s3-global.amazonaws.com/import/{userId}/{taskId}/...` | PUT | Upload PDF binary to S3 presigned URL (no cookie needed; URL itself is auth) |
| 3 | `api.paperpile.com/api/tasks/{taskId}/subtasks` | PATCH | Notify upload completion (`status: "uploaded"`) |
| 4 | `e.paperpile.com/track/` | POST | Mixpanel analytics (can be skipped) |

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

**Step 1 response:** Returns S3 presigned URL (with `X-Amz-*` query params, expires in 3600s) and task/subtask UUIDs.

**Step 3 request body:**
```json
{
  "subtasks": ["<subtask-uuid>"],
  "status": "uploaded"
}
```

**Required headers for API calls (Step 1 & 3):**
- `Content-Type: application/json`
- `Cookie: plack_session=<session_value>`
- `Origin: https://app.paperpile.com`
- `Referer: https://app.paperpile.com/`

**S3 upload (Step 2):**
- No cookies required (presigned URL contains auth)
- `Content-Type: application/x-www-form-urlencoded`
- Request body is the raw PDF binary

### 3.3 User ID

The user ID (`68CE82F6807411EA9B68A87FDE8EC746` format) appears in the S3 path. It is likely returned in Step 1 response or can be derived from the session.

### 3.4 Endpoints Not Yet Analyzed

- **Delete**: Need to capture delete request to identify endpoint and item ID format.
- **List**: Need to identify how to fetch library items.

## 4. Authentication Design

### Approach: Bookmarklet + Local HTTP Server

1. User runs `paperpile-cli login`.
2. CLI starts a local HTTP server on `localhost:18080`.
3. CLI opens the user's browser to `https://app.paperpile.com`.
4. User clicks a bookmarklet (one-time setup) that extracts `plack_session` from `document.cookie` and sends it to `localhost:18080/callback`.
5. CLI receives the session, saves it to `~/.config/paperpile-cli/config.yaml`, and shuts down the server.

**Bookmarklet:**
```javascript
javascript:void(fetch('http://localhost:18080/callback',{method:'POST',body:document.cookie.match(/plack_session=([^;]+)/)[1]}))
```

**Session refresh:** When the session expires, user re-runs `paperpile-cli login` and clicks the bookmarklet again.

## 5. Development Phases

### Phase 1: Project Setup & Auth
- [x] API research (upload flow)
- [ ] `go mod init`, `cobra-cli init`
- [ ] Implement `login` command (local HTTP server + bookmarklet flow)
- [ ] Config file management (`~/.config/paperpile-cli/config.yaml`)

### Phase 2: Upload Command
- [ ] Implement Step 1: `POST /api/import/files` (get presigned URL)
- [ ] Implement Step 2: `PUT` PDF to S3
- [ ] Implement Step 3: `PATCH /api/tasks/{taskId}/subtasks` (notify completion)
- [ ] End-to-end upload test

### Phase 3: List & Delete
- [ ] Capture and analyze list/delete API requests
- [ ] Implement `list` command
- [ ] Implement `delete` command

### Phase 4: Polish
- [ ] Error handling and retry logic
- [ ] Session expiry detection and re-login prompt
- [ ] CI/CD setup

## 6. Command Design

```
paperpile-cli login                  # Start auth flow (bookmarklet)
paperpile-cli upload <file_path>     # Upload a PDF
paperpile-cli list                   # List library items
paperpile-cli delete <item_id>      # Delete an item
```
