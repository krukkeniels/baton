# PRD — CLI Orchestrator for LLM-Driven Task Execution (Requirements-Only, Cycle-Based)

**Purpose**: A single CLI that wraps external LLM CLIs to advance work one task state at a time. Execution is organized into cycles. Each cycle advances exactly one task by one valid state transition, with context cleared between cycles and a formal handover artifact to bridge cycles. Tasks embed detailed, LLM-readable documents (e.g., implementation plans).

## 1. Scope & Objectives

### In scope (v1)
- Deterministic state machine with single-step advancement per cycle
- Read a Markdown plan file (vision, requirements, roadmap)
- Maintain a local persistent store for tasks, requirements, agents, artifacts, and audits
- Expose MCP operations so an LLM can read/update tasks & requirements and select the next task
- Enforce completion handshake after every cycle
- Context reset at cycle boundaries with rehydration only from stored sources
- Handover artifacts (e.g., implementation plan) embedded in the task record

### Out of scope (v1)
- Multi-user sync, remote PM, cross-repo orchestration, auto PR creation (may be later)

### Primary objective
`start` performs one cycle: select → transition → analyze/execute → handover → completion handshake → audit → stop.

## 2. Definitions

- **Cycle**: One atomic execution that advances a single task by one valid state transition and then stops. Begins with context reset; rehydrates context strictly from stored sources (plan file, task, artifacts, requirements, config).
- **Plan file**: Markdown source of truth (vision, scope, product requirements, roadmap).
- **Task**: Unit of work with state, priority, dependencies, owner, embedded artifacts (e.g., implementation plan), and audit history.
- **Agent**: Role (e.g., Architect, Developer) that executes state logic subject to permissions.
- **Handover**: Structured artifact produced at the end of a cycle that the next cycle consumes.
- **Local store**: Tech-agnostic persistent storage for all records.
- **MCP tools**: LLM-callable operations to query/update tasks, requirements, plan, and artifacts.

## 3. Users & Goals

- **Developer**: Run a single command to advance one step; receive precise handovers (what to build, how to test).
- **Architect**: Keep plan/requirements in Markdown; ensure traceability and high-signal handovers.
- **Operator**: Inspect selection rationale, override states, review cycle-level audits.

## 4. State Machine

### 4.1 Canonical states
```
ready_for_plan
→ planning
→ ready_for_implementation
→ implementing
→ ready_for_code_review
→ reviewing
→ (ready_for_commit | needs_fixes)
→ (committing | fixing)
→ (DONE | ready_for_code_review)
```

### 4.2 Input aliases → normalized
- `ready_for_implmentation` → `ready_for_implementation`
- `ready_for_code_revie` → `ready_for_code_review`
- `need_fixes` → `needs_fixes`
- `commiting` → `committing`

### 4.3 Allowed transitions
As listed above; invalid transitions must be rejected with a precise explanation.

**Invariant**: One cycle advances one task by one valid transition.

## 5. Functional Requirements

### 5.1 Plan & Requirements

**FR-P1**: Read a configured Markdown plan file path.

**FR-P2**: Extract "Product Requirements" (functional, non-functional, constraints, risks) into the store with stable keys.

**FR-P3**: Re-ingest updates without duplicating keys; report diffs and parse errors with locations.

**FR-P4**: Link requirements to tasks (traceability).

### 5.2 Local Persistent Store

**FR-S1**: Persist tasks, requirements, agents, artifacts, audit logs.

**FR-S2**: Guarantee durable, atomic state changes per cycle.

**FR-S3**: Prevent simultaneous advancement of the same task (locking).

### 5.3 Tasks (with embedded documents)

**FR-T1**: Task fields include: id, title, description, state, priority, owner, tags[], dependencies[], blocked_by[], created_at, updated_at.

**FR-T2**: Tasks can embed artifacts as named documents (markdown or structured text) that are easy for LLMs to consume.

**FR-T3**: The implementation plan (from planning) must be stored as a task artifact and versioned on each update.

**FR-T4**: List, filter (state/priority/tag/owner), import/export tasks; validate dependencies before advancement.

### 5.4 Agents

**FR-A1**: Provide at least two roles: architect, developer.

**FR-A2**: Each agent has: id, name, role description, routing policy, permissions.

**FR-A3**: Agent policies configurable per project.

### 5.5 Task Selection

**FR-SEL1**: Default: (1) highest priority first, (2) dependencies satisfied, (3) prefer leaf subtasks, (4) oldest updated_at tie-break.

**FR-SEL2**: Configurable selection policy.

**FR-SEL3**: `tasks next` shows what would be picked and why.

### 5.6 Cycle Lifecycle

**FR-CYC1** (Context reset): At cycle start, clear all conversational/contextual memory. Rehydrate only from stored sources (plan file, selected task & artifacts, requirements, config).

**FR-CYC2** (Single transition): Perform one valid transition for the selected task.

**FR-CYC3** (Analysis & execution): Execute state logic per agent policy, operating solely on the rehydrated context.

**FR-CYC4** (Handover creation): Produce/update handover artifact(s) required for the next state; attach to task.

**FR-CYC5** (Completion handshake): Enforce explicit completion (see §5.7).

**FR-CYC6** (Audit): Record full cycle inputs, outputs, handovers, decisions.

**FR-CYC7** (Stop): End cycle; do not chain multiple transitions.

### 5.7 Completion Handshake & Enforcement

**FR-CH1**: After execution, the agent must either update the state to the correct next state or return a structured not finished/blocked outcome with reason and recommended next state.

**FR-CH2**: If neither is provided, issue follow-up: "Are you finished? The state is not updated."

**FR-CH3**: Bounded retries (configurable; default 1–2). On failure, set `needs_fixes` with an actionable note and audit the exchange.

**FR-CH4**: Validate any proposed next state against allowed transitions; reject invalid proposals with one corrective follow-up; persistent invalidity → `needs_fixes`.

**FR-CH5**: A cycle is complete only when the state change or explicit outcome is persisted and audited.

**FR-CH6**: Finalization must use MCP ops (no bypass).

**FR-CH7**: `--dry-run`: perform the same checks without persisting; show the path and expected result.

### 5.8 Handover Artifacts (Between-Cycle Contracts)

**FR-H1**: Each state that precedes another must emit a handover artifact the next state can consume.

**FR-H2**: Required handovers:

- **planning → implementing**: Implementation Plan containing at minimum:
  - Goal & scope (task-scoped)
  - Acceptance criteria
  - Constraints & dependencies (requirements & code)
  - Step-by-step implementation outline
  - Test plan (unit/integration) and data assumptions
  - File/Module touch list (anticipated)
  - Risks & mitigations

- **implementing → reviewing**: Change Summary including:
  - What changed & why (linked to acceptance criteria)
  - Diff summary (human-readable)
  - Commands executed & results (tests/lints)
  - Known limitations/TODOs

- **reviewing → (ready_for_commit | needs_fixes)**: Review Findings with:
  - Decision (approve / request changes)
  - Findings grouped by severity
  - Concrete actions (if needs_fixes)

- **fixing → ready_for_code_review**: Fix Plan & Results:
  - Issues addressed
  - Minimal deltas applied
  - Verification performed

- **committing → DONE**: Commit Summary:
  - Commit reference(s)
  - Message(s) and scope
  - Any post-commit notes

**FR-H3**: All handovers must be LLM-readable (markdown or simple structured text) and attached to the task as artifacts.

**FR-H4**: Handovers must be versioned per cycle and referenced in audits.

### 5.9 External LLM CLI Integration (tool-agnostic)

**FR-LLM1**: Support at least two external LLM CLIs; allow per-state/agent routing.

**FR-LLM2**: One-time fallback to an alternate provider on failure/limits.

**FR-LLM3**: Redact secrets in prompts/logs; bound prompt size; deterministic chunking.

### 5.10 Version Control (VCS)

**FR-V1**: On committing: stage changes, create a conventional commit message, record commit ref in audit.

**FR-V2**: No-op in `--dry-run`; never log secrets.

**FR-V3**: Failures → `needs_fixes` with reason.

### 5.11 MCP Operations (LLM-facing)

**FR-M1**: Namespace provides:

- **Tasks**: get_next (with rationale), get(id), update_state(id, state, note?), append_note(id, note), list(filters), export/import.
- **Requirements**: list(filter), get(id/key), update(id/key, patch).
- **Plan**: read raw content.
- **Artifacts** (task-scoped): list(task_id), get(task_id, name/version), upsert(task_id, name, content, meta), history(task_id, name).

**FR-M2**: Machine-friendly errors with remediation hints.

**FR-M3**: Enforce agent permissions per operation.

### 5.12 Auditability & Observability

**FR-AU1**: Every cycle writes an audit entry: task id, selected rationale, prev/next state, actor, inputs summary (plan/requirements/artifacts hashes or excerpts), outputs summary (handover names), external commands invoked, outcome.

**FR-AU2**: If a follow-up was required, include exact follow-up text and agent responses.

**FR-AU3**: Provide status view: counts by state, recent cycles, blockers, pending follow-ups.

### 5.13 Configuration & Safety

**FR-C1**: Config includes: plan file path, selection policy, routing policy, allowed local commands, VCS behavior, retry counts for completion handshake, cycle timebox.

**FR-C2**: Enforce allow-list for local execution.

**FR-C3**: Restrict file writes to project workspace unless explicitly permitted.

**FR-C4**: Redact env secrets and known patterns in logs/audits.

### 5.14 Error Handling & Recovery

**FR-E1**: Categorize errors (selection, transition, execution, VCS, storage, configuration) with actionable guidance.

**FR-E2**: Idempotent cycles: resume after interruption without duplicate effects.

**FR-E3**: No partial advancement without an audit entry.

## 6. Non-Functional Requirements

**NFR-1**: **Determinism**: Given same inputs/config, selection and outcomes (excluding external LLM variability) are reproducible; handovers make next cycles predictable.

**NFR-2**: **Performance**: Typical cycle completes promptly (excluding external waits); plan ingestion within practical bounds.

**NFR-3**: **Reliability**: Crash-safe; audits never lost.

**NFR-4**: **Security/Privacy**: No secrets in artifacts/logs; minimal MCP exposure.

**NFR-5**: **Portability**: No OS-specific dependencies beyond standard shell/fs.

**NFR-6**: **Usability**: Clear CLI help; precise errors; visible "why this task".

**NFR-7**: **Extensibility**: Add states/agents/selection rules via config when feasible.

**NFR-8**: **Offline-first**: All but external LLM/VCS network actions work offline.

**NFR-9**: **Concurrency**: Prevent concurrent advancement of same task; document chosen locking semantics.

**NFR-10**: **Traceability**: Every handover/artifact and state change is linked across cycles.

## 7. Data Requirements (information model; tech-agnostic)

### Task
- **Core fields**: id, title, description, state, priority, owner, tags[], dependencies[], blocked_by[], created_at, updated_at
- **artifacts[]**: array of { name, version, content, meta, created_at }
  - Required names include: implementation_plan, change_summary, review_findings, fix_plan, commit_summary (as applicable per state)
- **links**: requirement keys covered; related tasks
- **cycle_history[]**: references to audit entries for each cycle

### Requirement
- **Fields**: id and/or key (e.g., "FR-12"), title, text, type (functional|nonfunctional|constraint|risk), timestamps

### Agent
- **Fields**: id, name, role description, routing policy, permissions

### Audit entry (per cycle)
- **Fields**: id, task_id, prev_state, next_state, actor, selection_reason, inputs_summary (hashes/refs), outputs_summary (handover names/versions), commands, result, note, follow_ups[], created_at

## 8. CLI Command Requirements (behavior only)

- **init**: Create config, local store, sample plan; idempotent.
- **ingest** `<plan-file>`: Parse plan; upsert requirements; report changes/errors with locations.
- **start** `[--dry-run]`: Execute one cycle: context reset → rehydrate → select → transition → execute → handover → completion handshake → audit → stop.
- **status**: Summaries by state; recent cycles; blockers; pending follow-ups.
- **tasks list** `[filters]`: List/filter tasks.
- **tasks next**: Show next selection and rationale.
- **tasks update** `--id … --state … [--note …]`: Manual override with validation & audit.
- **artifacts** `list|get|upsert --task …`: Manage task-scoped artifacts (human-readable content).
- **cycles show** `--task …`: Display cycle history and associated handovers (read-only).

## 9. Acceptance Criteria (MVP)

**AC-1**: `start` advances exactly one task by one valid transition, writes an audit entry, and stops.

**AC-2**: Every planning completion stores an Implementation Plan artifact attached to the task; implementing cycles must consume it.

**AC-3**: Context is cleared at the start of each cycle; rehydration sources are limited to plan, task (incl. artifacts), requirements, config (verified in audit).

**AC-4**: If an agent completes work without updating state, the system issues: "Are you finished? The state is not updated."; retries are bounded; unresolved → `needs_fixes` with a clear note.

**AC-5**: `tasks next` explains selection per policy and identifies any blocking dependencies.

**AC-6**: `reviewing` produces Review Findings and transitions to `ready_for_commit` or `needs_fixes`.

**AC-7**: `committing` records a commit reference in the audit (or `needs_fixes` on failure).

**AC-8**: Re-ingesting the plan updates requirement records without key duplication; task→requirement links remain intact.

**AC-9**: Two concurrent `start` invocations cannot advance the same task.

**AC-10**: All artifacts are versioned; audits reference artifact versions used/produced.

## 10. Success Metrics

- ≥ 85% cycles complete without manual intervention
- ≥ 95% of cycles that enter planning produce a valid Implementation Plan consumed in the next implementing cycle
- 0 audit/hand-over data loss across crashes
- 100% requirements traceable to tasks or flagged as unmapped

## 11. Risks & Mitigations

- **Context loss causing regressions** → Strict rehydration rules; required handovers; audit of inputs used
- **LLM variability** → Tight output contracts; fallback; needs_fixes path; deterministic chunking
- **Handovers too verbose/too sparse** → Minimum fields per state; validation checks
- **Selection surprises** → `tasks next` rationale and audit selection_reason
- **Hanging steps** → Completion handshake & bounded retries

## 12. Open Questions

1. Should cycles have a configurable timebox that converts to `needs_fixes` if exceeded?
2. Do we allow custom handover templates per project/state in v1, or fix the minimal set above?
3. Should selection policy consider owner/agent affinity by default?
4. What is the minimal artifact diff (e.g., store content hashes) required in audits for reproducibility?

---

*End of PRD (requirements-only, cycle-based).*