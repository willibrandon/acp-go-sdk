# Colosseum

Multi-agent consultation tool for design and implementation decisions.

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
┌─────────────────────────────────────────────────────────────────────┐
│                            Colosseum                                 │
│                                                                      │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │                      Orchestrator                             │   │
│  │                                                               │   │
│  │  - Parses user input and routing directives (@claude, etc)   │   │
│  │  - Maintains shared conversation history                      │   │
│  │  - Formats context for each agent                            │   │
│  │  - Coordinates response ordering                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                              │                                       │
│              ┌───────────────┴───────────────┐                      │
│              ▼                               ▼                      │
│  ┌─────────────────────┐         ┌─────────────────────┐           │
│  │   Claude Agent      │         │   Gemini Agent      │           │
│  │                     │         │                     │           │
│  │  ACP Connection     │         │  ACP Connection     │           │
│  │  Response Buffer    │         │  Response Buffer    │           │
│  │  Session State      │         │  Session State      │           │
│  └──────────┬──────────┘         └──────────┬──────────┘           │
│             │                               │                       │
│             ▼                               ▼                       │
│  ┌─────────────────────┐         ┌─────────────────────┐           │
│  │ claude-code-acp     │         │ gemini              │           │
│  │ (subprocess)        │         │ --experimental-acp  │           │
│  └─────────────────────┘         └─────────────────────┘           │
│                                                                      │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │                    Conversation History                       │   │
│  │                                                               │   │
│  │  []Message{ Role, Content, Timestamp }                       │   │
│  │                                                               │   │
│  │  Injected into each prompt so agents have full context       │   │
│  └──────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
```

## Components

### Agent

Manages a single ACP connection to an AI agent.

```go
type Agent struct {
    Name      string                    // "claude" or "gemini"
    conn      *acp.ClientSideConnection
    sessionID string
    cmd       *exec.Cmd
    mu        sync.Mutex
    buffer    strings.Builder           // Accumulates streaming response
}

// Prompt sends a message and returns the complete response.
// Streams chunks to the provided writer as they arrive.
func (a *Agent) Prompt(ctx context.Context, message string, stream io.Writer) (string, error)

// Close terminates the agent subprocess.
func (a *Agent) Close() error
```

### Message

Represents a single conversation turn.

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
}
```

### Orchestrator

Coordinates multi-agent conversations.

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

// Send processes user input, routes to appropriate agent(s), and updates history.
func (o *Orchestrator) Send(ctx context.Context, input string) error

// parseTarget extracts routing directive from input.
// "@claude foo" -> TargetClaude, "foo"
// "@gemini foo" -> TargetGemini, "foo"
// "foo"         -> TargetBoth, "foo"
func parseTarget(input string) (Target, string)

// buildContext formats conversation history for agent consumption.
func (o *Orchestrator) buildContext(target Role) string
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

### Output Format

```
Claude: Response text from Claude streams here as it arrives.
        Multi-line responses are indented for readability.

Gemini: Response text from Gemini follows after Claude completes.
        Each agent's response is clearly labeled.
```

Terminal colors (when supported):
- Claude: Purple (ANSI 35)
- Gemini: Blue (ANSI 34)
- System messages: Yellow (ANSI 33)
- Errors: Red (ANSI 31)

## Execution Flow

### Startup

```
1. Parse CLI flags
2. Spawn Claude subprocess (claude-code-acp)
3. Spawn Gemini subprocess (gemini --experimental-acp)
4. Initialize ACP connections for both (parallel)
5. Create sessions for both agents with working directory context
6. Display connection status
7. Enter input loop
```

### Message Processing

```
1. Read user input
2. Parse target directive (@claude, @gemini, or both)
3. Append user message to history
4. For each target agent:
   a. Build context string with full history
   b. Send prompt via ACP
   c. Stream response chunks to terminal
   d. Buffer complete response
   e. Append agent response to history
5. Return to input loop
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
├── main.go          # Entry point, CLI parsing, REPL loop
├── agent.go         # Agent connection and lifecycle management
├── orchestrator.go  # Multi-agent coordination and history
├── ui.go            # Terminal output formatting
└── README.md        # User-facing documentation
```

## Dependencies

- `github.com/coder/acp-go-sdk` - ACP protocol implementation
- Go standard library only for all other functionality

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
- Sequential responses when addressing both agents (not parallel)
- No web UI (terminal only)
- Two agents maximum (Claude and Gemini)
- Requires both agents to be locally installed and authenticated

## Future Considerations

These are explicitly out of scope for the initial implementation:

- Additional agents (GPT, local models)
- Parallel response streaming
- Conversation export/import
- TUI with scrollback and panes
- Voting/ranking mode for responses
- Integration with spec-kit or other workflows
