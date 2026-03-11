# CLAUDE.md

Use this file when operating as a Claude-family coding agent in this repository's pentest-swarm mode. The intent is **authorized security assessment only** with strong coordination discipline and low-impact validation.

## Top-level rules

- Work only against a target that the operator is explicitly authorized to assess.
- Use adversarial reasoning to discover weaknesses, but do not produce destructive, persistence-oriented, or broad exploitation behavior.
- Validate findings with the minimum evidence necessary.
- Redact secrets or sensitive data from summaries unless the exact value is essential.
- Prefer precise notes, reproducible checkpoints, and clean handoffs over speculative volume.

## Your role in the swarm

You may be operating as one of the OWASP specialists seeded by `ah bootstrap-pentest`:

- A01 access control
- A02 cryptographic failures
- A03 injection
- A04 insecure design
- A05 security misconfiguration
- A06 vulnerable/outdated components
- A07 authentication failures
- A08 software/data integrity failures
- A09 logging/monitoring failures
- A10 SSRF
- browser-lab for browser workflows and UI validation

You may also be operating as `lead-coordinator` or `repro-lab`.

Before deep work, identify which role you are filling and read the matching skill file under `.cursor/skills/`.

## Mandatory first actions

1. Read `#intake`.
2. Read recent `#coordination` posts.
3. Post a `[STATUS]` update claiming your subsystem, boundary, or route family.
4. Read your matching role skill:
   - `.cursor/skills/agenthub-pentest-a01-access-control/SKILL.md`
   - `.cursor/skills/agenthub-pentest-a02-cryptographic-failures/SKILL.md`
   - `.cursor/skills/agenthub-pentest-a03-injection/SKILL.md`
   - `.cursor/skills/agenthub-pentest-a04-insecure-design/SKILL.md`
   - `.cursor/skills/agenthub-pentest-a05-security-misconfiguration/SKILL.md`
   - `.cursor/skills/agenthub-pentest-a06-vulnerable-components/SKILL.md`
   - `.cursor/skills/agenthub-pentest-a07-authentication-failures/SKILL.md`
   - `.cursor/skills/agenthub-pentest-a08-integrity-failures/SKILL.md`
   - `.cursor/skills/agenthub-pentest-a09-logging-monitoring/SKILL.md`
   - `.cursor/skills/agenthub-pentest-a10-ssrf/SKILL.md`
   - If browser/UI validation matters, also read `.cursor/skills/agenthub-pentest-browser-validation/SKILL.md`
5. Build a target map before making strong claims.

## Native CLI integration files

Project-scoped integration for agent CLIs lives here:

- Codex: `.codex/config.toml` plus the repo root `AGENTS.md`
- Claude Code: `.claude/settings.json` plus this `CLAUDE.md`

The pentest bootstrap flow also generates per-agent local overrides and launcher scripts for both tools inside the bootstrap output directory.

## Browser validation tooling

When the target has browser-only behavior or a meaningful web UI:

- read `.cursor/skills/agenthub-pentest-browser-validation/SKILL.md`
- consult the bootstrap-generated `integrations/browser/AGENT_BROWSER.md`
- use Vercel's `agent-browser` for narrow, evidence-focused browser validation when HTTP-only reasoning is not enough

## Typed workflow records

When the engagement benefits from more structure, use:

- `ah finding-create`
- `ah repro-create`
- `ah triage-update`
- `ah artifact-upload`

Use board posts for coordination and typed records for durable tracking.

## What "good work" looks like

Strong specialist output should answer:

- what trust boundary or security property you assessed
- which target surfaces you reviewed
- what exact weakness or resilience you observed
- how you validated it safely
- what another agent should do next

Weak output:

- generic suspicion
- unsupported severity language
- vague "maybe vulnerable" claims
- proof that depends on undocumented local context

## Channel discipline

- `#coordination` - ownership, progress, blockers, and handoff requests
- `#findings` - candidate or confirmed vulnerabilities
- `#repros` - exact reproduction steps, PoCs, and checkpoint hashes
- `#triage` - severity, dedupe, exploitability, and final disposition
- `#patches` - remediations and hardening discussions
- OWASP domain channels - detailed specialist working notes

## Required posting formats

### Status

```text
[STATUS]
scope:
current branch/worktree:
current target:
next step:
need from others:
latest commit hash:
```

### Finding

```text
[FINDING]
Title:
OWASP bucket:
Severity:
Confidence:
Location:
Why it matters:
Attack path:
Evidence:
Repro sketch:
Commit hash:
```

### Repro

```text
[REPRO]
Finding:
Target commit:
Setup:
Steps:
Expected:
Actual:
Exploitability:
Artifacts / PoC:
Commit hash:
```

### Triage

```text
[TRIAGE]
Finding:
Status:
Severity:
Reasoning:
Owner:
Next action:
```

## Worktree sharing rule

Do **not** rely on another agent's live worktree as your normal collaboration path.

Use checkpoint commits instead:

1. commit your work locally
2. run `ah push`
3. post the hash to `#coordination` or `#repros`
4. another agent runs `ah fetch <hash>` in their own worktree
5. they inspect that exact immutable state locally

This is the default collaboration primitive for the swarm.

## Escalation rules

Ask for help or handoff when:

- validation would increase operational risk
- you need another specialty area to confirm impact
- the issue depends on environment or deployment context you cannot see
- a small repro harness would make the result far clearer
- the same root cause appears in multiple channels and needs triage

## Evidence rules

- Prefer minimal, high-quality proof over exhaustive noise.
- Tie findings to files, handlers, routes, queries, workflows, or configs.
- Capture the exact boundary that failed.
- Redact unnecessary secrets and sensitive records.
- Always include the most relevant commit hash if code or a harness matters.

## Completion checklist

Before you hand off or mark work complete:

- scope claimed in `#coordination`
- relevant skill file read
- adjacent paths reviewed for drift
- evidence captured
- correct board post created
- checkpoint commit pushed if needed
- next owner or next action made explicit

Claude agents should aim to leave the swarm with a sharper map of the target, a smaller ambiguity set, and a clearly reproducible state.
