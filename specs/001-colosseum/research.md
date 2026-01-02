# Research: Colosseum Multi-Agent Consultation TUI

**Date**: 2026-01-01
**Branch**: `001-colosseum`

## Overview

This document consolidates research findings for implementing Colosseum, resolving all technical unknowns identified in the planning phase.

---

## 1. Bubble Tea Streaming Patterns

### Decision: Channel-based async updates with `tea.Cmd`

### Rationale

The `realtime` example in bubbletea demonstrates the canonical pattern for streaming external data:

1. Use a channel to receive events from async operations
2. Return a `tea.Cmd` that blocks on the channel and returns a message
3. Re-subscribe for the next event in the Update handler

### Pattern Implementation

```go
// Message types for streaming
type AgentChunkMsg struct {
    Agent   string
    Content string
}

type AgentCompleteMsg struct {
    Agent   string
    Content string
    Err     error
}

// Wait for next chunk from agent stream
func waitForAgentChunk(agent string, chunks <-chan string) tea.Cmd {
    return func() tea.Msg {
        content, ok := <-chunks
        if !ok {
            return AgentCompleteMsg{Agent: agent}
        }
        return AgentChunkMsg{Agent: agent, Content: content}
    }
}

// In Update handler
case AgentChunkMsg:
    m.appendToStreamingMessage(msg.Agent, msg.Content)
    m.viewport.SetContent(m.renderMessages())
    m.viewport.GotoBottom()
    return m, waitForAgentChunk(msg.Agent, m.agentChunks[msg.Agent])
```

### Alternatives Considered

- **Polling with `tea.Every`**: Rejected because it adds latency and complexity
- **Direct channel reads in Update**: Rejected because it blocks the event loop (violates Constitution Principle II)

### Reference

- `/Users/brandon/src/bubbletea/examples/realtime/main.go` - Channel-based realtime updates
- `/Users/brandon/src/bubbletea/examples/chat/main.go` - Chat interface with viewport

---

## 2. Chat Interface Layout

### Decision: Viewport + TextInput with responsive sizing

### Rationale

The `chat` example in bubbletea provides the exact pattern needed:

1. Viewport for scrollable message history
2. TextInput/TextArea for user input
3. Handle `tea.WindowSizeMsg` for responsive layout
4. Use `viewport.GotoBottom()` after adding messages

### Pattern Implementation

```go
type Model struct {
    viewport viewport.Model
    input    textinput.Model
    messages []Message
    width    int
    height   int
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.viewport.Width = msg.Width
        m.viewport.Height = msg.Height - 4  // Reserve header + input
        m.input.Width = msg.Width - 4
        return m, nil
    }
    // ...
}

func (m Model) View() string {
    header := m.renderHeader()
    content := m.viewport.View()
    input := m.input.View()
    return lipgloss.JoinVertical(lipgloss.Left, header, content, input)
}
```

### Key Findings

- Use `lipgloss.Height()` to calculate component heights for layout math
- Call `viewport.GotoBottom()` after appending new messages
- Set content with `lipgloss.NewStyle().Width(m.viewport.Width).Render()` for proper wrapping

### Reference

- `/Users/brandon/src/bubbletea/examples/chat/main.go` - Complete chat layout pattern

---

## 3. Spinner Integration

### Decision: Spinner component with conditional rendering

### Rationale

The bubbles spinner integrates seamlessly with Bubble Tea's update loop via `spinner.TickMsg`.

### Pattern Implementation

```go
type Model struct {
    spinner  spinner.Model
    state    AppState
}

func (m Model) Init() tea.Cmd {
    return m.spinner.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case spinner.TickMsg:
        if m.state == StateStreaming {
            var cmd tea.Cmd
            m.spinner, cmd = m.spinner.Update(msg)
            return m, cmd
        }
    }
    // ...
}

func (m Model) View() string {
    if m.state == StateStreaming {
        return m.spinner.View() + " Waiting for response..."
    }
    return ""
}
```

### Reference

- `/Users/brandon/src/bubbles/spinner/spinner.go` - Spinner API
- `/Users/brandon/src/bubbletea/examples/realtime/main.go` - Spinner with async

---

## 4. ACP SDK Connection Management

### Decision: Separate `acp.ClientSideConnection` per agent with shared Client interface

### Rationale

Existing examples (`example/claude-code/main.go`, `example/gemini/main.go`) demonstrate:

1. Each agent runs as a subprocess with stdin/stdout pipes
2. `acp.NewClientSideConnection(client, stdin, stdout)` creates the connection
3. The `Client` interface handles callbacks: `SessionUpdate`, `RequestPermission`, etc.
4. Streaming happens via `SessionUpdate` callback with `AgentMessageChunk`

### Pattern Implementation

```go
type Agent struct {
    name      string
    conn      *acp.ClientSideConnection
    sessionID string
    cmd       *exec.Cmd
    chunks    chan string  // Stream to TUI
}

// Client implementation receives streaming updates
func (a *Agent) SessionUpdate(ctx context.Context, params acp.SessionNotification) error {
    u := params.Update
    if u.AgentMessageChunk != nil && u.AgentMessageChunk.Content.Text != nil {
        a.chunks <- u.AgentMessageChunk.Content.Text.Text
    }
    return nil
}

func (a *Agent) RequestPermission(ctx context.Context, params acp.RequestPermissionRequest) (acp.RequestPermissionResponse, error) {
    if a.autoApprove {
        // Select allow option
        for _, o := range params.Options {
            if o.Kind == acp.PermissionOptionKindAllowOnce {
                return acp.RequestPermissionResponse{
                    Outcome: acp.RequestPermissionOutcome{
                        Selected: &acp.RequestPermissionOutcomeSelected{OptionId: o.OptionId},
                    },
                }, nil
            }
        }
    }
    // Prompt user via TUI (send message to Bubble Tea)
    // ...
}
```

### Key Findings

- `conn.Initialize()` must be called before `conn.NewSession()`
- `conn.Prompt()` blocks until response complete; streaming happens via `SessionUpdate` callback
- Each agent needs its own goroutine to avoid blocking
- `conn.Cancel()` sends cancellation notification

### Reference

- `/Users/brandon/src/acp-go-sdk/example/claude-code/main.go` - Claude connection
- `/Users/brandon/src/acp-go-sdk/example/gemini/main.go` - Gemini connection

---

## 5. Multi-Agent Orchestration

### Decision: Sequential prompting with shared history context

### Rationale

Per spec requirements:
- Claude responds first, then Gemini (sequential, not parallel)
- Both agents see the full conversation history
- Context formatting identifies each participant

### Pattern Implementation

```go
type Orchestrator struct {
    claude   *Agent
    gemini   *Agent
    history  []Message
    pending  []string  // Agents waiting to respond
}

func (o *Orchestrator) Send(ctx context.Context, input string, target Target) tea.Cmd {
    // Add user message to history
    o.history = append(o.history, Message{Role: RoleUser, Content: input})

    // Determine agents to prompt
    switch target {
    case TargetBoth:
        o.pending = []string{"claude", "gemini"}
    case TargetClaude:
        o.pending = []string{"claude"}
    case TargetGemini:
        o.pending = []string{"gemini"}
    }

    return o.promptNextAgent(ctx)
}

func (o *Orchestrator) promptNextAgent(ctx context.Context) tea.Cmd {
    if len(o.pending) == 0 {
        return nil
    }
    agent := o.pending[0]
    o.pending = o.pending[1:]

    context := o.buildContext(agent)
    return o.agents[agent].Prompt(ctx, context)
}

func (o *Orchestrator) buildContext(forAgent string) string {
    var sb strings.Builder
    sb.WriteString("You are participating in a collaborative discussion...\n\n")
    for _, msg := range o.history {
        sb.WriteString(fmt.Sprintf("[%s]: %s\n\n", msg.Role, msg.Content))
    }
    return sb.String()
}
```

### Target Parsing

```go
func parseTarget(input string) (Target, string) {
    input = strings.TrimSpace(input)
    switch {
    case strings.HasPrefix(input, "@claude "):
        return TargetClaude, strings.TrimPrefix(input, "@claude ")
    case strings.HasPrefix(input, "@gemini "):
        return TargetGemini, strings.TrimPrefix(input, "@gemini ")
    case strings.HasPrefix(input, "@both "):
        return TargetBoth, strings.TrimPrefix(input, "@both ")
    default:
        return TargetBoth, input
    }
}
```

---

## 6. Lip Gloss Styling

### Decision: Centralized style definitions in `styles.go`

### Rationale

Per Constitution Principle V, all styles must be centralized.

### Pattern Implementation

```go
// styles.go
package main

import "github.com/charmbracelet/lipgloss"

var (
    // Colors
    ClaudeColor = lipgloss.Color("#A855F7")  // Purple
    GeminiColor = lipgloss.Color("#3B82F6")  // Blue
    UserColor   = lipgloss.Color("#22C55E")  // Green
    SystemColor = lipgloss.Color("#6B7280")  // Gray
    BorderColor = lipgloss.Color("#374151")

    // Role styles
    ClaudeNameStyle = lipgloss.NewStyle().
        Foreground(ClaudeColor).
        Bold(true)

    GeminiNameStyle = lipgloss.NewStyle().
        Foreground(GeminiColor).
        Bold(true)

    UserNameStyle = lipgloss.NewStyle().
        Foreground(UserColor).
        Bold(true)

    // Status indicators
    ConnectedStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#22C55E"))

    DisconnectedStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#EF4444"))

    // Layout
    HeaderStyle = lipgloss.NewStyle().
        Bold(true).
        Padding(0, 1).
        Border(lipgloss.NormalBorder(), false, false, true, false).
        BorderForeground(BorderColor)

    InputStyle = lipgloss.NewStyle().
        Border(lipgloss.NormalBorder(), true, false, false, false).
        BorderForeground(BorderColor).
        Padding(0, 1)
)

func RenderStatus(connected bool) string {
    if connected {
        return ConnectedStyle.Render("●")
    }
    return DisconnectedStyle.Render("○")
}
```

### Reference

- `/Users/brandon/src/lipgloss/examples/layout/main.go` - Layout patterns
- DESIGN.md Styles section - Color definitions

---

## 7. Permission Request Handling

### Decision: Message-based permission flow integrated with TUI

### Rationale

Permission requests from agents need to:
1. Display in the TUI with options
2. Wait for user selection
3. Return the selected option to the agent

### Pattern Implementation

```go
type PermissionRequestMsg struct {
    Agent    string
    Title    string
    Options  []PermissionOption
    Response chan acp.RequestPermissionResponse
}

// In Agent's RequestPermission callback
func (a *Agent) RequestPermission(ctx context.Context, params acp.RequestPermissionRequest) (acp.RequestPermissionResponse, error) {
    if a.autoApprove {
        return autoApprovePermission(params)
    }

    // Send to TUI and wait for response
    respChan := make(chan acp.RequestPermissionResponse)
    a.permissionChan <- PermissionRequestMsg{
        Agent:    a.name,
        Title:    *params.ToolCall.Title,
        Options:  convertOptions(params.Options),
        Response: respChan,
    }
    return <-respChan, nil
}

// In Model Update
case PermissionRequestMsg:
    m.showPermissionDialog(msg)
    return m, nil
```

---

## Summary

All technical unknowns have been resolved:

| Area | Decision | Key Pattern |
|------|----------|-------------|
| Streaming | Channel + `tea.Cmd` | `waitForAgentChunk()` resubscription |
| Layout | Viewport + TextInput | WindowSizeMsg handling |
| Spinner | Conditional with TickMsg | Only tick when streaming |
| ACP | Separate connections per agent | `SessionUpdate` callback streaming |
| Orchestration | Sequential with history | `promptNextAgent()` queue |
| Styling | Centralized in styles.go | Role-based color mapping |
| Permissions | Message-based TUI flow | Response channel pattern |

**Next Step**: Proceed to Phase 1 with data-model.md
