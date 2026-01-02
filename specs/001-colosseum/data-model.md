# Data Model: Colosseum Multi-Agent Consultation TUI

**Date**: 2026-01-01
**Branch**: `001-colosseum`

## Overview

Colosseum is a session-based CLI application with no persistence. All entities exist only in memory during a session.

---

## Entities

### Message

Represents a single conversation turn for display in the viewport.

```go
type Role string

const (
    RoleUser   Role = "user"
    RoleClaude Role = "claude"
    RoleGemini Role = "gemini"
    RoleSystem Role = "system"
)

type Message struct {
    Role      Role      // Who sent this message
    Content   string    // Message text (may be partial during streaming)
    Timestamp time.Time // When the message was created
    Streaming bool      // True while content is still arriving
}
```

**Validation Rules**:
- `Role` must be one of the defined constants
- `Content` may be empty only during initial streaming state
- `Timestamp` must be set on creation

**Relationships**:
- Ordered collection maintained by `Orchestrator.history`
- Displayed in `Model.messages` viewport

---

### Agent

Represents a connected AI agent with ACP connection and streaming state.

```go
type Agent struct {
    Name      string                    // "claude" or "gemini"
    conn      *acp.ClientSideConnection // ACP protocol connection
    sessionID string                    // Active session identifier
    cmd       *exec.Cmd                 // Subprocess handle

    mu        sync.Mutex                // Protects buffer and state
    buffer    strings.Builder           // Current response accumulator
    chunks    chan string               // Stream chunks for TUI updates
    connected bool                      // Connection status

    autoApprove bool                    // --yolo flag behavior
}
```

**Validation Rules**:
- `Name` must be "claude" or "gemini"
- `conn` must be initialized before `sessionID` is set
- `chunks` channel must be created before streaming begins

**State Transitions**:
```
Disconnected → Connecting → Connected → Prompting → Streaming → Connected
                    ↓                                    ↓
                 Failed                              Failed
```

**Relationships**:
- Managed by `Orchestrator`
- Receives commands from `Model` via `tea.Cmd`
- Sends updates to `Model` via message channel

---

### Orchestrator

Coordinates multi-agent conversations and produces Bubble Tea commands.

```go
type Target int

const (
    TargetBoth Target = iota
    TargetClaude
    TargetGemini
)

type Orchestrator struct {
    claude   *Agent           // Claude Code agent
    gemini   *Agent           // Gemini CLI agent
    history  []Message        // Full conversation history
    pending  []string         // Agents waiting to respond
    mu       sync.Mutex       // Protects history and pending
}
```

**Validation Rules**:
- At least one agent must be connected for the orchestrator to function
- `history` is append-only during a session
- `pending` is cleared when all agents have responded

**State Transitions**:
```
Idle → Prompting(pending=[claude,gemini]) → Streaming(claude) →
     → Prompting(pending=[gemini]) → Streaming(gemini) → Idle
```

---

### Model (Bubble Tea)

Main application state following the Elm architecture.

```go
type AppState int

const (
    StateIdle      AppState = iota  // Ready for input
    StateWaiting                     // Waiting for agent response
    StateStreaming                   // Receiving streaming response
    StatePermission                  // Showing permission dialog
)

type Model struct {
    // TUI Components (from bubbles)
    viewport viewport.Model      // Scrollable message history
    input    textinput.Model     // User input field
    spinner  spinner.Model       // Loading indicator

    // Application State
    orchestrator *Orchestrator   // Agent coordination
    messages     []Message       // Conversation history (for display)
    state        AppState        // Current interaction state
    err          error           // Last error, if any

    // Permission Dialog State (when StatePermission)
    permRequest  *PermissionRequest // Current permission request

    // Layout
    width  int
    height int
    ready  bool                  // Terminal size received
}
```

**State Transitions**:
```
Idle --(Enter with input)--> Waiting
Waiting --(AgentChunkMsg)--> Streaming
Streaming --(AgentChunkMsg)--> Streaming
Streaming --(AgentCompleteMsg with pending)--> Waiting
Streaming --(AgentCompleteMsg no pending)--> Idle
Any --(PermissionRequestMsg)--> Permission
Permission --(user selects)--> Previous State
Any --(Escape)--> Idle (cancel)
Any --(Ctrl+C)--> Quit
```

---

### PermissionRequest

Represents an agent's request for user approval.

```go
type PermissionOption struct {
    ID   string // Option identifier to return
    Name string // Display name
    Kind string // "allow_once", "allow_always", "deny"
}

type PermissionRequest struct {
    Agent    string             // Which agent is asking
    Title    string             // What action needs permission
    Options  []PermissionOption // Available choices
    Response chan<- int         // Channel to send selected index
}
```

**Validation Rules**:
- `Options` must have at least one entry
- `Response` must be non-nil

---

## Message Types (tea.Msg)

### Streaming Messages

```go
// AgentChunkMsg - Partial response from an agent
type AgentChunkMsg struct {
    Agent   string // "claude" or "gemini"
    Content string // Text chunk to append
}

// AgentCompleteMsg - Agent finished responding
type AgentCompleteMsg struct {
    Agent   string // "claude" or "gemini"
    Content string // Final complete content
    Err     error  // Error if failed
}

// AgentStatusMsg - Connection status update
type AgentStatusMsg struct {
    Agent     string // "claude" or "gemini"
    Connected bool   // Connection state
    Err       error  // Error if failed
}
```

### Permission Messages

```go
// PermissionRequestMsg - Agent needs permission
type PermissionRequestMsg struct {
    Request *PermissionRequest
}

// PermissionResponseMsg - User responded to permission
type PermissionResponseMsg struct {
    SelectedIndex int
}
```

### System Messages

```go
// InitCompleteMsg - All agents initialized
type InitCompleteMsg struct {
    ClaudeConnected bool
    GeminiConnected bool
}

// ErrorMsg - General error
type ErrorMsg struct {
    Err error
}
```

---

## CLI Configuration

```go
type Config struct {
    ClaudeCmd   string        // Command to spawn Claude (default: "npx -y @zed-industries/claude-code-acp@latest")
    GeminiCmd   string        // Command to spawn Gemini (default: "gemini")
    Yolo        bool          // Auto-approve permissions
    Timeout     time.Duration // Response timeout (default: 120s)
    NoColor     bool          // Disable colors
    Debug       bool          // Enable debug logging
}
```

---

## Data Flow

```
User Input
    │
    ▼
┌─────────────────┐
│ Model.Update    │ ─── KeyMsg (Enter) ───▶ parseTarget(input)
└────────┬────────┘                              │
         │                                       ▼
         │                              ┌─────────────────┐
         │                              │ Orchestrator    │
         │                              │ .Send()         │
         │                              └────────┬────────┘
         │                                       │
         ▼                                       ▼
┌─────────────────┐                     ┌─────────────────┐
│ state=Waiting   │                     │ Agent.Prompt()  │
│ spinner.Tick    │                     │ (goroutine)     │
└─────────────────┘                     └────────┬────────┘
                                                 │
    ┌────────────────────────────────────────────┘
    │
    ▼
┌─────────────────┐     SessionUpdate callback
│ ACP Connection  │ ────────────────────────────▶ chunks channel
└─────────────────┘                                    │
                                                       ▼
                                              ┌─────────────────┐
                                              │ waitForChunk()  │
                                              │ tea.Cmd         │
                                              └────────┬────────┘
                                                       │
                                              AgentChunkMsg
                                                       │
                                                       ▼
                                              ┌─────────────────┐
                                              │ Model.Update    │
                                              │ append to msg   │
                                              │ viewport.Set    │
                                              └─────────────────┘
```

---

## Session Lifecycle

1. **Startup**
   - Parse CLI flags into `Config`
   - Create `Orchestrator` with `Agent` instances
   - Initialize Bubble Tea program
   - Spawn agent subprocesses
   - Initialize ACP connections (parallel)
   - Create sessions with cwd context

2. **Running**
   - User types message → `Model.Update` → `Orchestrator.Send`
   - Agents stream responses → `AgentChunkMsg` → viewport update
   - Permission requests → `PermissionRequestMsg` → dialog → response

3. **Shutdown**
   - User sends `:exit` or Ctrl+C
   - Cancel active contexts
   - Close agent connections
   - Terminate subprocesses
   - Exit program

---

## Invariants

1. **At most one agent streaming at a time** (sequential responses)
2. **History is append-only** (no editing past messages)
3. **Viewport reflects history** (kept in sync on every update)
4. **Permission responses unblock exactly one waiting agent**
5. **State machine is never in an invalid state** (all transitions defined)
