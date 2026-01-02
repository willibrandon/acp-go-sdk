# Colosseum

Multi-agent consultation TUI for design and implementation decisions.

## Reference Libraries

Local Charm ecosystem libraries for development reference:

| Library | Path | Purpose |
|---------|------|---------|
| Bubble Tea | `../bubbletea` | Core TUI framework (Elm architecture) |
| Bubbles | `../bubbles` | Pre-built components (viewport, text input, spinner) |
| Lip Gloss | `../lipgloss` | Styling, colors, borders, layout |
| Huh | `../huh` | Forms and interactive prompts |

## Problem Statement

During software design and implementation, developers face decisions that benefit from multiple perspectives:

- Architectural trade-offs with no clear "right" answer
- Implementation approaches with different cost/complexity profiles
- Debugging complex issues where fresh eyes help
- API design where consistency and ergonomics matter
- Technology selection with long-term implications

A single AI assistant provides one perspective. Different models have different strengths, training data, and reasoning patterns. Consulting multiple models manually is tedious and loses conversational context.

## Solution

Colosseum provides a unified interface for consulting multiple AI agents (Claude Code and Gemini) simultaneously. Both agents participate in the same conversation, see each other's responses, and can build on or challenge each other's reasoning.

### Example Session

```
$ colosseum

Connected: Claude (claude-code-acp) | Gemini (gemini --experimental-acp)
Session: /Users/dev/myproject
Type a message to consult both agents. Use @claude or @gemini to address one.

> We need to decide between PostgreSQL and SQLite for a desktop app
  that syncs to a cloud backend. What factors should we consider?

Claude: For a desktop app with cloud sync, consider these factors:

  1. Deployment complexity: SQLite is zero-config, embedded. PostgreSQL
     requires a separate process or embedded solution like embedded-postgres.

  2. Concurrent access: If your app has multiple processes or threads
     writing simultaneously, PostgreSQL handles this natively. SQLite
     requires WAL mode and careful connection management.

  3. Sync architecture: SQLite's file-based nature makes it amenable to
     file-level sync (like Dropbox-style), while PostgreSQL typically
     requires logical replication or API-based sync.

Gemini: I'd add a few dimensions to Claude's analysis:

  1. Offline-first capability: SQLite excels here. The entire database
     travels with the app. PostgreSQL requires connection management
     and offline queue strategies.

  2. Query compatibility: If your cloud backend runs PostgreSQL, using
     SQLite locally means maintaining two SQL dialects. Consider if
     your ORM abstracts this sufficiently.

  3. Data volume: SQLite performs well up to ~100GB. Beyond that, or
     with complex queries on large datasets, PostgreSQL's query planner
     and indexing options become valuable.

> @gemini What about using libSQL/Turso for the SQLite path?

Gemini: libSQL is worth serious consideration for your use case...
```

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Colosseum TUI                                   │
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │                         Bubble Tea Application                          │ │
│  │                                                                         │ │
│  │  Model {                                                                │ │
│  │    orchestrator  *Orchestrator     // Agent coordination               │ │
│  │    viewport      viewport.Model    // Scrollable conversation view     │ │
│  │    input         textinput.Model   // User input field                 │ │
│  │    spinner       spinner.Model     // Loading indicator                │ │
│  │    messages      []Message         // Conversation history             │ │
│  │    state         AppState          // idle, waiting, streaming         │ │
│  │  }                                                                      │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│                                      │                                       │
│                                      ▼                                       │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │                           Orchestrator                                  │ │
│  │                                                                         │ │
│  │  - Parses user input and routing directives (@claude, @gemini, @both)  │ │
│  │  - Maintains shared conversation history                                │ │
│  │  - Formats context for each agent                                      │ │
│  │  - Sends streaming updates via tea.Cmd                                 │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│                                      │                                       │
│                  ┌───────────────────┴───────────────────┐                  │
│                  ▼                                       ▼                  │
│  ┌─────────────────────────────┐         ┌─────────────────────────────┐   │
│  │       Claude Agent          │         │       Gemini Agent          │   │
│  │                             │         │                             │   │
│  │  ACP Connection             │         │  ACP Connection             │   │
│  │  Response Buffer            │         │  Response Buffer            │   │
│  │  Session State              │         │  Session State              │   │
│  │  Streaming Channel          │         │  Streaming Channel          │   │
│  └──────────────┬──────────────┘         └──────────────┬──────────────┘   │
│                 │                                       │                   │
│                 ▼                                       ▼                   │
│  ┌─────────────────────────────┐         ┌─────────────────────────────┐   │
│  │ claude-code-acp             │         │ gemini --experimental-acp   │   │
│  │ (subprocess via npx)        │         │ (subprocess)                │   │
│  └─────────────────────────────┘         └─────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
```

### TUI Layout

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  COLOSSEUM                                          Claude ● | Gemini ●     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  You                                                                        │
│  We need to decide between PostgreSQL and SQLite for a desktop app that    │
│  syncs to a cloud backend. What factors should we consider?                │
│                                                                             │
│  ─────────────────────────────────────────────────────────────────────────  │
│                                                                             │
│  Claude                                                                     │
│  For a desktop app with cloud sync, consider these factors:                │
│                                                                             │
│  1. Deployment complexity: SQLite is zero-config, embedded...              │
│  2. Concurrent access: If your app has multiple processes...               │
│  3. Sync architecture: SQLite's file-based nature makes it...              │
│                                                                             │
│  ─────────────────────────────────────────────────────────────────────────  │
│                                                                             │
│  Gemini                                                                     │
│  I'd add a few dimensions to Claude's analysis:                            │
│                                                                             │
│  1. Offline-first capability: SQLite excels here...                        │
│  2. Query compatibility: If your cloud backend runs PostgreSQL...          │
│                                                                      ▼ more │
├─────────────────────────────────────────────────────────────────────────────┤
│  > @gemini what about libSQL?                                          ⏎   │
└─────────────────────────────────────────────────────────────────────────────┘
  @claude | @gemini | @both (default)                    ctrl+c to exit
```

## Components

### Bubble Tea Model

The main application state following the Elm architecture.

```go
type AppState int

const (
    StateIdle      AppState = iota  // Ready for input
    StateWaiting                     // Waiting for agent response
    StateStreaming                   // Receiving streaming response
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

    // Layout
    width  int
    height int
    ready  bool                  // Terminal size received
}

// Messages (Bubble Tea commands)
type (
    AgentChunkMsg    struct { Agent string; Content string }
    AgentCompleteMsg struct { Agent string; Content string; Err error }
    AgentStatusMsg   struct { Agent string; Connected bool }
)
```

### Agent

Manages a single ACP connection with streaming support for Bubble Tea.

```go
type Agent struct {
    Name      string
    conn      *acp.ClientSideConnection
    sessionID string
    cmd       *exec.Cmd

    mu        sync.Mutex
    buffer    strings.Builder
    chunks    chan string        // Stream chunks for TUI updates
    connected bool
}

// Prompt sends a message and streams responses via the chunks channel.
// Returns a tea.Cmd that can be used in the Bubble Tea update loop.
func (a *Agent) Prompt(ctx context.Context, message string) tea.Cmd

// Status returns a tea.Cmd that checks connection status.
func (a *Agent) Status() tea.Cmd

// Close terminates the agent subprocess.
func (a *Agent) Close() error
```

### Message

Represents a single conversation turn for display.

```go
type Role string

const (
    RoleUser   Role = "user"
    RoleClaude Role = "claude"
    RoleGemini Role = "gemini"
    RoleSystem Role = "system"
)

type Message struct {
    Role      Role
    Content   string
    Timestamp time.Time
    Streaming bool              // True while content is still arriving
}
```

### Orchestrator

Coordinates multi-agent conversations and produces Bubble Tea commands.

```go
type Orchestrator struct {
    claude  *Agent
    gemini  *Agent
    history []Message
    mu      sync.Mutex
}

type Target int

const (
    TargetBoth Target = iota
    TargetClaude
    TargetGemini
)

// Send processes user input and returns a tea.Cmd that will stream responses.
func (o *Orchestrator) Send(ctx context.Context, input string) tea.Cmd

// parseTarget extracts routing directive from input.
func parseTarget(input string) (Target, string)

// buildContext formats conversation history for agent consumption.
func (o *Orchestrator) buildContext(target Role) string
```

### Styles (Lip Gloss)

Consistent styling across the TUI.

```go
var (
    // Colors
    ClaudeColor = lipgloss.Color("#A855F7")  // Purple
    GeminiColor = lipgloss.Color("#3B82F6")  // Blue
    UserColor   = lipgloss.Color("#22C55E")  // Green
    BorderColor = lipgloss.Color("#374151")  // Gray

    // Message styles
    ClaudeStyle = lipgloss.NewStyle().
        Foreground(ClaudeColor).
        Bold(true)

    GeminiStyle = lipgloss.NewStyle().
        Foreground(GeminiColor).
        Bold(true)

    UserStyle = lipgloss.NewStyle().
        Foreground(UserColor).
        Bold(true)

    // Layout styles
    HeaderStyle = lipgloss.NewStyle().
        Bold(true).
        Padding(0, 1).
        Border(lipgloss.NormalBorder(), false, false, true, false).
        BorderForeground(BorderColor)

    InputStyle = lipgloss.NewStyle().
        Border(lipgloss.NormalBorder(), true, false, false, false).
        BorderForeground(BorderColor).
        Padding(0, 1)

    StatusConnected    = lipgloss.NewStyle().Foreground(lipgloss.Color("#22C55E"))
    StatusDisconnected = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444"))
)
```

### Context Formatting

Each agent receives the conversation history formatted to clearly identify participants:

```
You are participating in a collaborative discussion with a user and another
AI assistant. The user is consulting both of you to get multiple perspectives
on a technical decision.

Previous discussion:

[User]: We need to decide between PostgreSQL and SQLite for a desktop app...

[Claude]: For a desktop app with cloud sync, consider these factors...

[Gemini]: I'd add a few dimensions to Claude's analysis...

---

The user is now addressing you directly. Provide your perspective. You may
reference, build upon, or respectfully disagree with the other assistant's
points when relevant.

[User]: What about using libSQL/Turso for the SQLite path?
```

## User Interface

### Input Commands

| Input | Behavior |
|-------|----------|
| `<message>` | Both agents respond (Claude first, then Gemini) |
| `@claude <message>` | Only Claude responds |
| `@gemini <message>` | Only Gemini responds |
| `@both <message>` | Explicit form of default behavior |
| `:status` | Display connection status for both agents |
| `:history` | Display conversation history |
| `:clear` | Clear conversation history |
| `:exit` | Terminate session |

### TUI Rendering

The viewport displays messages with role-based styling:

```go
func (m Model) renderMessages() string {
    var sb strings.Builder
    for _, msg := range m.messages {
        switch msg.Role {
        case RoleUser:
            sb.WriteString(UserStyle.Render("You"))
        case RoleClaude:
            sb.WriteString(ClaudeStyle.Render("Claude"))
        case RoleGemini:
            sb.WriteString(GeminiStyle.Render("Gemini"))
        }
        sb.WriteString("\n")
        sb.WriteString(msg.Content)
        if msg.Streaming {
            sb.WriteString(m.spinner.View())
        }
        sb.WriteString("\n\n")
    }
    return sb.String()
}
```

### Keyboard Navigation

| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `Ctrl+C` | Exit application |
| `↑` / `↓` | Scroll conversation history |
| `PgUp` / `PgDn` | Scroll page up/down |
| `Home` / `End` | Jump to start/end of history |
| `Esc` | Cancel current request |

## Execution Flow

### Startup

```
1. Parse CLI flags
2. Initialize Bubble Tea program with alt screen
3. Spawn Claude subprocess (claude-code-acp)
4. Spawn Gemini subprocess (gemini --experimental-acp)
5. Initialize ACP connections for both (parallel, via tea.Cmd)
6. Create sessions for both agents with working directory context
7. Update header with connection status
8. Focus text input, ready for user
```

### Bubble Tea Update Loop

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {

    case tea.KeyMsg:
        switch msg.String() {
        case "enter":
            if m.state == StateIdle && m.input.Value() != "" {
                input := m.input.Value()
                m.input.Reset()
                m.state = StateWaiting
                return m, m.orchestrator.Send(context.Background(), input)
            }
        case "ctrl+c":
            return m, tea.Quit
        case "esc":
            if m.state != StateIdle {
                // Cancel current request
                return m, m.orchestrator.Cancel()
            }
        }

    case AgentChunkMsg:
        // Append chunk to current streaming message
        m.updateStreamingMessage(msg.Agent, msg.Content)
        m.viewport.SetContent(m.renderMessages())
        m.viewport.GotoBottom()
        return m, m.waitForNextChunk(msg.Agent)

    case AgentCompleteMsg:
        // Finalize message, possibly trigger next agent
        m.finalizeMessage(msg.Agent, msg.Content)
        if m.orchestrator.HasPendingAgent() {
            return m, m.orchestrator.PromptNextAgent()
        }
        m.state = StateIdle
        return m, nil

    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.viewport.Width = msg.Width
        m.viewport.Height = msg.Height - 4  // Reserve header + input
        m.ready = true
    }

    // Update sub-components
    var cmds []tea.Cmd
    m.input, cmd = m.input.Update(msg)
    cmds = append(cmds, cmd)
    m.viewport, cmd = m.viewport.Update(msg)
    cmds = append(cmds, cmd)
    if m.state == StateWaiting {
        m.spinner, cmd = m.spinner.Update(msg)
        cmds = append(cmds, cmd)
    }

    return m, tea.Batch(cmds...)
}
```

### Message Processing

```
1. User presses Enter with non-empty input
2. Parse target directive (@claude, @gemini, or both)
3. Append user message to history, update viewport
4. Set state to Waiting, return tea.Cmd to prompt first agent
5. Agent streams chunks via AgentChunkMsg
6. Each chunk updates the streaming message in viewport
7. AgentCompleteMsg finalizes message
8. If @both, trigger next agent; otherwise return to Idle
9. Focus returns to text input
```

### Shutdown

```
1. User enters :exit or sends interrupt (Ctrl+C)
2. Cancel active contexts
3. Terminate agent subprocesses
4. Exit cleanly
```

## Error Handling

| Scenario | Behavior |
|----------|----------|
| Claude fails to connect | Warn user, continue with Gemini only |
| Gemini fails to connect | Warn user, continue with Claude only |
| Both fail to connect | Exit with error |
| Agent errors mid-conversation | Display error, continue with other agent |
| Agent timeout | Cancel after 120s, display timeout message |
| Permission request | Auto-approve if `-yolo` flag, otherwise prompt user |

## CLI Interface

```
colosseum - Multi-agent consultation for design and implementation decisions

Usage:
  colosseum [flags]

Flags:
  --claude string    Claude ACP command (default: npx -y @zed-industries/claude-code-acp@latest)
  --gemini string    Gemini CLI path (default: gemini)
  --yolo             Auto-approve all tool permission requests
  --timeout int      Response timeout in seconds (default: 120)
  --no-color         Disable colored output
  --debug            Enable debug logging
  --help             Display help
```

## File Structure

```
example/colosseum/
├── DESIGN.md        # This document
├── main.go          # Entry point, CLI parsing, Bubble Tea program
├── model.go         # Bubble Tea model, Init, Update, View
├── agent.go         # Agent connection, ACP client, streaming
├── orchestrator.go  # Multi-agent coordination and history
├── styles.go        # Lip Gloss styles and theme
├── keys.go          # Key bindings
└── README.md        # User-facing documentation
```

## Dependencies

Go modules:

```go
require (
    github.com/coder/acp-go-sdk v0.6.3
    github.com/charmbracelet/bubbletea v1.2.4
    github.com/charmbracelet/bubbles v0.20.0
    github.com/charmbracelet/lipgloss v1.0.0
)
```

Reference implementations available locally:
- `../bubbletea` - Framework examples and patterns
- `../bubbles` - Component usage (viewport, textinput, spinner)
- `../lipgloss` - Styling examples
- `../huh` - Form patterns (if needed for settings)

External runtime requirements:
- Node.js (for claude-code-acp via npx)
- Gemini CLI installed and authenticated
- Claude Code authenticated (~/.claude/settings.json)

## Testing

Manual testing scenarios:

1. **Basic consultation**: Ask both agents a design question, verify both respond
2. **Directed question**: Use @claude and @gemini to verify routing works
3. **Context continuity**: Ask follow-up questions, verify agents reference earlier discussion
4. **Disagreement handling**: Ask a question where agents may disagree, verify both perspectives shown
5. **Single agent mode**: Start with one agent unavailable, verify graceful degradation
6. **Long conversation**: 10+ turns to verify history management
7. **Tool use**: Ask questions requiring file reads, verify permission handling

## Limitations

- No persistent conversation history (session only)
- Sequential responses when addressing both agents (not parallel streaming)
- Terminal only (no web UI)
- Two agents maximum (Claude and Gemini)
- Requires both agents to be locally installed and authenticated

## Future Considerations

These are explicitly out of scope for the initial implementation:

- Additional agents (GPT, local models via Ollama)
- Parallel response streaming (both agents responding simultaneously)
- Conversation export/import (save/load sessions)
- Voting/ranking mode for responses
- Integration with spec-kit or other workflows
- Web UI version
- Configuration persistence (remember preferences)
