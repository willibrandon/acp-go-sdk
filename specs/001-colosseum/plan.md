# Implementation Plan: Colosseum Multi-Agent Consultation TUI

**Branch**: `001-colosseum` | **Date**: 2026-01-01 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-colosseum/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Build a terminal-based multi-agent consultation tool that connects Claude Code and Gemini via ACP, allowing developers to consult both AI agents simultaneously with shared conversation context. Uses Bubble Tea (Elm architecture) for the TUI with streaming responses, message routing via @directives, and graceful degradation when one agent is unavailable.

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: acp-go-sdk, bubbletea, bubbles (viewport, textinput, spinner), lipgloss
**Storage**: N/A (session-only, no persistence)
**Testing**: Manual testing scenarios per DESIGN.md, go test for unit tests
**Target Platform**: macOS, Linux (terminal with 256-color support)
**Project Type**: single (CLI application)
**Performance Goals**: Agent connection <10s, response streaming <5s to first chunk, smooth scroll with 100+ messages
**Constraints**: 120s response timeout (configurable), non-blocking event loop (per Constitution)
**Scale/Scope**: 2 agents (Claude, Gemini), single user, session-based

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Elm Architecture | ✅ PASS | Model/Update/View pattern enforced via Bubble Tea |
| II. Non-Blocking Event Loop | ✅ PASS | All I/O via tea.Cmd, context cancellation supported |
| III. Streaming First | ✅ PASS | AgentChunkMsg for incremental updates, spinner during streaming |
| IV. Graceful Degradation | ✅ PASS | Continue with one agent if other fails (FR-013) |
| V. Centralized Styling | ✅ PASS | All styles in styles.go per file structure |
| VI. Reference-Driven Development | ✅ PASS | Local Charm libs available at ../bubbletea, ../bubbles, etc. |

**Technology Constraints:**
| Constraint | Status | Notes |
|-----------|--------|-------|
| Go 1.21+ | ✅ PASS | go.mod specifies go 1.21 |
| Required deps (acp, bubbletea, bubbles, lipgloss) | ✅ PASS | All specified in dependencies |
| Prohibited: huh forms | ✅ PASS | Not using huh forms |
| Prohibited: Agents beyond Claude/Gemini | ✅ PASS | Only Claude and Gemini |
| Prohibited: Conversation persistence | ✅ PASS | Session-only storage |
| Prohibited: Parallel streaming | ✅ PASS | Sequential responses only |

**GATE RESULT: ✅ PASS** — Proceed to Phase 0.

## Project Structure

### Documentation (this feature)

```text
specs/001-colosseum/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
example/colosseum/
├── DESIGN.md            # Design document (already exists)
├── main.go              # Entry point, CLI parsing, Bubble Tea program
├── model.go             # Bubble Tea model, Init, Update, View
├── agent.go             # Agent connection, ACP client, streaming
├── orchestrator.go      # Multi-agent coordination and history
├── styles.go            # Lip Gloss styles and theme
├── keys.go              # Key bindings
└── README.md            # User-facing documentation
```

**Structure Decision**: Following the existing example pattern in the repo (example/claude-code/, example/gemini/), Colosseum will be placed at `example/colosseum/` with flat file structure per DESIGN.md specification. This matches the Constitution's Development Workflow section.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations — all principles and constraints pass.

---

## Constitution Check (Post-Design)

*Re-evaluation after Phase 1 design artifacts completed.*

| Principle | Status | Verification |
|-----------|--------|--------------|
| I. Elm Architecture | ✅ PASS | data-model.md defines Model/Update/View separation, state transitions documented |
| II. Non-Blocking Event Loop | ✅ PASS | All I/O via tea.Cmd per contracts/agent.go.md, channel-based streaming per research.md |
| III. Streaming First | ✅ PASS | AgentChunkMsg pattern documented in research.md and data-model.md |
| IV. Graceful Degradation | ✅ PASS | Status tracking per data-model.md, degradation flow per orchestrator contract |
| V. Centralized Styling | ✅ PASS | styles.go specified in project structure, patterns in research.md |
| VI. Reference-Driven Development | ✅ PASS | Reference libs used in research.md, documented in quickstart.md |

**POST-DESIGN GATE: ✅ PASS** — Ready for `/speckit.tasks`.

---

## Generated Artifacts

| Artifact | Path | Status |
|----------|------|--------|
| Implementation Plan | `specs/001-colosseum/plan.md` | ✅ Complete |
| Research | `specs/001-colosseum/research.md` | ✅ Complete |
| Data Model | `specs/001-colosseum/data-model.md` | ✅ Complete |
| Contracts | `specs/001-colosseum/contracts/` | ✅ Complete |
| Quickstart | `specs/001-colosseum/quickstart.md` | ✅ Complete |
| Agent Context | `CLAUDE.md` | ✅ Updated |

---

## Next Steps

Run `/speckit.tasks` to generate the implementation task list.
