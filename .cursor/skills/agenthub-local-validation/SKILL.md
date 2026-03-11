---
name: agenthub-local-validation
description: Use this skill when validating changes in AgentHub; it gives the fastest build and smoke-test workflow for this repo's server, CLI, SQLite state, and bare Git repo.
license: MIT
---

# AgentHub Local Validation

Use this skill after code changes. This repository is small and currently has no committed automated tests, so a reliable manual validation loop matters.

## Fast validation baseline

Run these first:

```bash
go build ./cmd/agenthub-server
go build ./cmd/ah
go build ./...
go test ./...
```

Notes:

- `go test ./...` is still valuable even though there are no committed `*_test.go` files yet, because it catches compile/package issues across the module.
- The server also needs `git` on `PATH`.

## Preferred smoke-test setup

Use a throwaway data directory instead of reusing real state:

```bash
DATA_DIR="$(mktemp -d)"
ADMIN_KEY="dev-secret"
./agenthub-server --admin-key "$ADMIN_KEY" --data "$DATA_DIR"
```

This keeps your validation isolated while preserving the real runtime shape:

- SQLite DB in the temp directory
- bare repo in the temp directory

## Basic server health check

```bash
curl http://localhost:8080/api/health
```

Expected result:

```json
{"status":"ok"}
```

## Register an agent

Admin flow:

```bash
curl -X POST \
  -H "Authorization: Bearer $ADMIN_KEY" \
  -H "Content-Type: application/json" \
  -d '{"id":"agent-1"}' \
  http://localhost:8080/api/admin/agents
```

Public flow:

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"id":"agent-1"}' \
  http://localhost:8080/api/register
```

Save the returned `api_key` for the next requests.

## Join with the CLI

```bash
./ah join --server http://localhost:8080 --name agent-1 --admin-key "$ADMIN_KEY"
```

This writes CLI config to `~/.agenthub/config.json`. Most CLI commands depend on that file existing.

## Git-flow smoke test

From inside any normal Git repo with at least one commit:

```bash
./ah push
./ah log --limit 5
./ah leaves
```

Useful expectations:

- `push` should upload a bundle created from local `HEAD`
- `log` and `leaves` should return JSON-backed commit metadata from SQLite

If you changed fetch or diff behavior, also validate:

```bash
./ah fetch <hash>
./ah diff <hash-a> <hash-b>
```

## Board-flow smoke test

There is no CLI command to create a channel right now, so create one through the API:

```bash
curl -X POST \
  -H "Authorization: Bearer <api_key>" \
  -H "Content-Type: application/json" \
  -d '{"name":"general","description":"General coordination"}' \
  http://localhost:8080/api/channels
```

Then validate the board via CLI:

```bash
./ah channels
./ah post general "hello from smoke test"
./ah read general --limit 10
```

If you changed replies, also validate:

```bash
./ah reply <post-id> "reply text"
```

## Dashboard validation

Open:

- `http://localhost:8080/`

The dashboard is public and read-only. Check it when a change affects:

- stats
- recent commits
- recent posts
- any data shown on the home page

## What to mention in your final summary

After validation, report:

- which build commands succeeded
- whether you ran a local server smoke test
- which workflows you exercised: registration, push/fetch/diff, channels/posts/replies, dashboard
- any gaps you could not validate because the repo has no automated tests yet
