<!--
  Sync Impact Report
  ==================
  Version change: 0.0.0 → 1.0.0 (initial ratification)

  Added sections:
  - 6 Core Principles for Colosseum TUI development
  - Technology Constraints section (Charm ecosystem requirements)
  - Development Workflow section (Elm architecture, async patterns)
  - Governance rules

  Templates reviewed:
  - .specify/templates/plan-template.md ✅ Compatible (Constitution Check section exists)
  - .specify/templates/spec-template.md ✅ Compatible (no conflicts)
  - .specify/templates/tasks-template.md ✅ Compatible (no conflicts)

  Follow-up TODOs: None
-->

# Colosseum Constitution

## Core Principles

### I. Elm Architecture

All state management MUST follow the Elm architecture pattern: Model, Update, View.

- The Model struct holds all application state
- Update receives messages and returns new state plus commands
- View renders the current state to the terminal
- No state mutations outside the Update function

**Rationale**: Bubble Tea enforces this pattern; violating it causes race conditions and unpredictable UI behavior.

### II. Non-Blocking Event Loop

All I/O operations MUST be performed via `tea.Cmd`; the TUI event loop MUST NOT block.

- Agent prompts, status checks, and subprocess operations return `tea.Cmd`
- Long-running operations communicate progress via messages (e.g., `AgentChunkMsg`)
- Context cancellation MUST be supported for all async operations

**Rationale**: Blocking the event loop freezes the UI and prevents user input handling.

### III. Streaming First

Responses MUST stream incrementally via `AgentChunkMsg` and update the viewport in real-time.

- Each chunk appends to the current message and triggers a viewport refresh
- `AgentCompleteMsg` finalizes the message and transitions state
- The spinner MUST display during streaming to indicate activity

**Rationale**: Users expect real-time feedback; batch responses feel unresponsive.

### IV. Graceful Degradation

Errors MUST be handled explicitly; the application MUST continue when one agent fails.

- If Claude fails to connect, continue with Gemini only (and vice versa)
- Mid-conversation agent errors display a message but don't crash the TUI
- Connection status MUST be visible in the header at all times

**Rationale**: Network and subprocess failures are expected; the tool remains useful with one agent.

### V. Centralized Styling

All Lip Gloss styles MUST be defined in `styles.go`; no inline styling in View functions.

- Color constants for Claude (purple), Gemini (blue), User (green)
- Reusable style variables for headers, borders, messages, status indicators
- Theme changes require only `styles.go` modifications

**Rationale**: Consistent appearance and maintainable theming.

### VI. Reference-Driven Development

Implementation MUST reference local Charm libraries for patterns and best practices.

- `../bubbletea` for framework patterns and examples
- `../bubbles` for viewport, textinput, spinner component usage
- `../lipgloss` for styling patterns
- `../huh` for forms only if settings UI becomes necessary

**Rationale**: Local references ensure consistency with proven patterns and reduce guesswork.

## Technology Constraints

**Language**: Go 1.21+

**Required Dependencies**:
- `github.com/coder/acp-go-sdk` - ACP protocol
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/bubbles` - Components (viewport, textinput, spinner)
- `github.com/charmbracelet/lipgloss` - Styling

**Prohibited**:
- `huh` forms unless explicitly required for settings
- Agents beyond Claude and Gemini
- Conversation persistence
- Parallel response streaming (sequential only)

**Local References**:
- `../bubbletea`, `../bubbles`, `../lipgloss`, `../huh`

## Development Workflow

**File Organization**:
- `main.go` - Entry point, CLI parsing, Bubble Tea program initialization
- `model.go` - Model struct, Init, Update, View functions
- `agent.go` - Agent connection, ACP client implementation, streaming
- `orchestrator.go` - Multi-agent coordination, history, target parsing
- `styles.go` - All Lip Gloss styles and color definitions
- `keys.go` - Key bindings

**Agent Interface**:
- SHOULD be generic to support future agent types
- MUST implement `Prompt(ctx, message) tea.Cmd` for streaming
- MUST implement `Status() tea.Cmd` for connection checks
- MUST implement `Close() error` for cleanup

**Testing**:
- Manual testing scenarios defined in DESIGN.md
- Focus on connection handling, streaming, and graceful degradation

## Governance

This constitution governs all Colosseum development decisions.

- Amendments require documentation update and version increment
- All code reviews MUST verify compliance with these principles
- Complexity beyond these constraints MUST be justified in writing

**Version**: 1.0.0 | **Ratified**: 2026-01-01 | **Last Amended**: 2026-01-01
