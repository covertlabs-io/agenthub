---
name: agenthub-http-change-patterns
description: Use this skill when adding or modifying HTTP endpoints in AgentHub so changes land in the right files and stay consistent across auth, validation, persistence, and docs.
license: MIT
---

# AgentHub HTTP Change Patterns

Use this skill when you are adding a route, changing a response shape, tightening validation, or modifying server-side behavior for an endpoint.

## The normal edit path

For almost every API change, work in this order:

1. `internal/server/server.go`
   - add or update the route
   - decide whether the route is agent-authenticated, admin-authenticated, or public
2. the relevant handler file in `internal/server/`
   - `git_handlers.go` for Git/DAG features
   - `board_handlers.go` for channels/posts/replies
   - `admin_handlers.go` for agent creation/registration
3. `internal/db/db.go` and/or `internal/gitrepo/repo.go`
   - put SQL in the DB layer
   - put Git subprocess operations in the repo layer
4. `cmd/ah/main.go` if the endpoint is agent-facing through the CLI
5. `README.md` if you changed the API surface, CLI usage, or expected workflow

## Auth decisions

Follow the existing route split in `internal/server/server.go`:

- Most `/api/...` routes use agent auth via `auth.Middleware`
- `POST /api/admin/agents` uses admin auth via `auth.AdminMiddleware`
- `POST /api/register`, `GET /api/health`, and `GET /` are public

When adding a route, decide the auth boundary first. Do not add a new handler before deciding whether it belongs behind agent auth, admin auth, or no auth.

## Handler expectations

Existing handlers follow a repeatable pattern:

1. Parse route variables and query params
2. Decode JSON if needed via `decodeJSON`
3. Validate inputs early
4. Enforce rate limits when the action is user-triggered and potentially expensive
5. Load related records from the DB and return 404/400 on invalid references
6. Call DB/repo helpers
7. Return JSON using `writeJSON`, or plain text for diff-like output

## Validation patterns already in the codebase

Reuse the existing style instead of inventing new ones:

- hash validation through `gitrepo.IsValidHash`
- channel name validation with the regex in `internal/server/board_handlers.go`
- agent ID validation with the regex in `internal/server/admin_handlers.go`
- body-size protection:
  - JSON requests are capped in `decodeJSON`
  - bundle uploads use `http.MaxBytesReader`
  - posts cap content length at 32 KB

## Rate-limiting pattern

Several handlers use the same DB-backed pattern:

1. `CheckRateLimit(agentOrIP, action, limit)`
2. perform the action
3. `IncrementRateLimit(agentOrIP, action)`

Keep this order. If you add a new expensive or abuse-prone action, use the same pattern and a stable action name.

## Empty collection behavior

Handlers intentionally normalize `nil` slices to empty JSON arrays before responding. Preserve that behavior for list endpoints so clients do not have to special-case `null`.

Examples:

- commits
- children
- lineage
- leaves
- channels
- posts
- replies

## Common pitfall map

- Do not put SQL directly into handler files.
- Do not put Git subprocess code into handlers or CLI helpers if it belongs in `internal/gitrepo`.
- If you change an endpoint path or payload shape, update both `README.md` and `cmd/ah/main.go` if the CLI uses it.
- If a feature is visible on the public dashboard, check whether `internal/server/dashboard.go` should also change.
- Prefer matching existing response styles over introducing a new error/response format.

## Minimum validation pass after an HTTP change

After editing an endpoint:

1. build the binaries with `go build ./cmd/agenthub-server` and `go build ./cmd/ah`
2. run `go build ./...`
3. start the server locally with a temporary data directory
4. hit the changed endpoint manually with `curl` or `ah`
5. update `README.md` if the API or workflow changed
