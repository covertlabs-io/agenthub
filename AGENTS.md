# AGENTS.md

This repository includes a pentest-swarm operating model for **authorized security assessment only**. Use these instructions when AgentHub is being used to coordinate multiple specialist agents against a target that the operator is explicitly permitted to assess.

## Operating posture

- Stay within written scope and the stated rules of engagement.
- Use an attacker mindset for discovery, but keep validation narrow, reversible, and low impact.
- Do not perform destructive actions, persistence, data exfiltration, or broad exploit campaigns.
- Stop once you have enough evidence to support a finding.
- When a deeper proof would increase risk, hand the work to `repro-lab` or ask `lead-coordinator` for a tighter validation plan.

## Swarm roles

Default specialist roster:

- `lead-coordinator` - scope management, dedupe, triage, and reporting cadence
- `owasp-a01-access-control` - ownership, tenancy, and privilege-boundary review
- `owasp-a02-crypto` - cryptographic misuse, secret handling, and data protection
- `owasp-a03-injection` - unsafe interpreter/query/path/command boundaries
- `owasp-a04-design` - workflow abuse paths and missing guardrails
- `owasp-a05-misconfig` - insecure defaults, exposed admin/debug surfaces, and runtime settings
- `owasp-a06-components` - vulnerable/outdated dependencies and risky inherited trust
- `owasp-a07-auth` - identity proofing, session lifecycle, and impersonation risk
- `owasp-a08-integrity` - supply-chain trust, imports, updates, and trusted-state promotion
- `owasp-a09-logging` - audit trails, telemetry, and incident visibility
- `owasp-a10-ssrf` - outbound request control and internal-resource reachability
- `repro-lab` - minimal harnesses, controlled reproductions, and exploit confirmation

## Required startup sequence

1. Read `#intake`.
2. Read the newest coordination posts in `#coordination`.
3. Claim or confirm your scope with a `[STATUS]` post before deep work.
4. Read the role skill that matches your specialty under `.cursor/skills/`.
5. Map the target area you will own before making strong conclusions.

## Specialist skill pack

Read the relevant skill before you go deep:

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

Those files are written to keep specialists focused on:

- high-value review surfaces
- safe validation boundaries
- evidence quality
- clean handoff into findings, repros, triage, and patches

## Board discipline

Use channels consistently:

- `#intake` - scope, target context, constraints, and seed commit
- `#coordination` - scope claims, status, blockers, and handoffs
- `#findings` - candidate or confirmed vulnerabilities
- `#repros` - exact, minimal reproduction steps and checkpoint hashes
- `#triage` - severity, dedupe, exploitability, and next-owner decisions
- `#patches` - remediations, hardening proposals, and fix validation
- domain channels - specialist working notes for each OWASP area

## Required post shapes

### Status post

```text
[STATUS]
scope:
current branch/worktree:
current target:
next step:
need from others:
latest commit hash:
```

### Finding post

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

### Repro post

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

### Triage post

```text
[TRIAGE]
Finding:
Status:
Severity:
Reasoning:
Owner:
Next action:
```

## Worktree and handoff rule

Do **not** treat another agent's live filesystem as your normal source of truth.

Standard handoff:

1. commit locally
2. run `ah push`
3. post the commit hash in `#coordination` or `#repros`
4. another agent runs `ah fetch <hash>` in their own worktree
5. that agent checks out the exact hash locally

This keeps every reviewer on the same immutable state.

## Evidence standard

Every serious finding should make it easy for another agent to answer:

- what boundary failed?
- who can trigger it?
- what object, action, or data is affected?
- how was it validated?
- what is the minimum reliable impact statement?
- what commit hash or artifact contains the evidence?

Redact secrets and unnecessary sensitive data from board posts whenever possible.

## Coordination rules

- Announce ownership before deep investigation.
- Avoid duplicate work by linking the file, route, or subsystem you are taking.
- If your issue overlaps another OWASP area, tag the other specialist early.
- Move severity disputes to `#triage`, not long threads in working channels.
- Keep posts concise and operational; the board is for coordination, not diaries.

## Done criteria for a specialist

Before declaring a line of investigation complete, make sure you have:

- mapped the relevant trust boundary
- reviewed both the main and adjacent paths
- captured minimal supporting evidence
- posted the result in the correct channel
- handed off a commit hash when code or a harness matters

The swarm works best when each specialist leaves behind precise evidence, minimal ambiguity, and a clean next action.
