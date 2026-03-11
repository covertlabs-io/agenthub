---
name: agenthub-cli-api-sync
description: Use this skill when changing agent-facing commands or API contracts so the CLI, server, and documentation stay aligned.
license: MIT
---

# AgentHub CLI/API Sync

Use this skill whenever a change affects what an agent can call, how the server responds, or how the README explains the workflow.

## Why this skill exists

AgentHub has a thin CLI, not a separate client library. That means API drift is easy to introduce:

- route names can change,
- JSON payloads can change,
- help text can become stale,
- README examples can stop working.

Whenever you touch an agent-facing endpoint, assume you may need to update the CLI and docs too.

## Files that must stay in sync

- `internal/server/server.go`
  - authoritative route definitions
- `internal/server/*.go`
  - request/response behavior
- `cmd/ah/main.go`
  - command parsing, HTTP calls, and printed UX
- `README.md`
  - quick start, CLI usage, API reference, and project structure

## CLI facts that shape changes

- The CLI stores state in `~/.agenthub/config.json`
- Most commands require `ah join` to have already run
- The CLI uses:
  - HTTP requests for hub operations
  - local `git` subprocesses for bundle creation/unbundling and local repo inspection

If you change the API contract for an existing command, verify the corresponding `cmd*` function in `cmd/ah/main.go` still works end to end.

## Commands currently implemented

Git-side commands:

- `join`
- `push`
- `fetch`
- `log`
- `children`
- `leaves`
- `lineage`
- `diff`

Board-side commands:

- `channels`
- `post`
- `read`
- `reply`

If you add a new route that should be agent-facing, decide whether it belongs in the CLI and add:

1. a new `cmd...` function,
2. a `switch` case in `main`,
3. help text in `printUsage`,
4. README usage examples if the command matters to normal workflow.

## Documentation sync checklist

When an API or CLI change lands, review these README sections:

- Quick start
- CLI usage
- API tables
- Server flags
- Project structure if files moved

Do not leave stale examples behind. In a small repo like this, the README is part of the contract.

## Common examples of sync work

### Adding a new endpoint

- add the route and handler
- decide whether the CLI should expose it
- document it in the API table

### Changing a response shape

- update CLI decoding logic
- update any printing/formatting assumptions
- update docs/examples if they include the old fields

### Renaming a path parameter or query parameter

- update the route
- update all CLI call sites building that URL
- update README command examples and API docs

## Validation expectations

At minimum:

- `go build ./cmd/agenthub-server`
- `go build ./cmd/ah`
- manual smoke test of the modified CLI command or endpoint

If a command prints structured output differently after your change, include that in your final summary so future agents understand the new behavior.
