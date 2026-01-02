# Quickstart: Colosseum Development

**Branch**: `001-colosseum`
**Location**: `example/colosseum/`

## Prerequisites

1. **Go 1.21+** installed
2. **Node.js** installed (for Claude Code via npx)
3. **Gemini CLI** installed and authenticated
4. **Claude Code** authenticated (`~/.claude/settings.json` exists)

## Development Setup

### 1. Navigate to project

```bash
cd /Users/brandon/src/acp-go-sdk/example/colosseum
```

### 2. Initialize Go module (if not exists)

```bash
go mod init github.com/coder/acp-go-sdk/example/colosseum
```

### 3. Add dependencies

```bash
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/bubbles
go get github.com/charmbracelet/lipgloss
```

### 4. Run the application

```bash
go run .
```

## File Creation Order

Based on dependencies, create files in this order:

1. **`styles.go`** - No dependencies (Constitution Principle V: centralized styling)
2. **`keys.go`** - No dependencies
3. **`agent.go`** - Depends on styles.go (for logging colors)
4. **`orchestrator.go`** - Depends on agent.go
5. **`model.go`** - Depends on orchestrator.go, styles.go
6. **`main.go`** - Depends on model.go

## Key Implementation Patterns

### Streaming with Channels (from research.md)

```go
// In agent.go - Wait for next chunk
func waitForAgentChunk(agent string, chunks <-chan string) tea.Cmd {
    return func() tea.Msg {
        content, ok := <-chunks
        if !ok {
            return AgentCompleteMsg{Agent: agent}
        }
        return AgentChunkMsg{Agent: agent, Content: content}
    }
}
```

### Viewport + Input Layout (from research.md)

```go
// In model.go - Handle window resize
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
    m.viewport.Width = msg.Width
    m.viewport.Height = msg.Height - 4  // header + input
    m.input.Width = msg.Width - 4
```

### ACP Client Implementation (from research.md)

```go
// In agent.go - Implement acp.Client interface
func (a *Agent) SessionUpdate(ctx context.Context, params acp.SessionNotification) error {
    if chunk := params.Update.AgentMessageChunk; chunk != nil {
        if chunk.Content.Text != nil {
            a.chunks <- chunk.Content.Text.Text
        }
    }
    return nil
}
```

## Testing

### Manual Test Scenarios

1. **Start with both agents**
   ```bash
   go run .
   # Verify: Both Claude and Gemini show connected (●)
   ```

2. **Basic consultation**
   ```
   > What are the trade-offs between REST and GraphQL?
   # Verify: Claude responds, then Gemini responds
   ```

3. **Directed question**
   ```
   > @claude Can you elaborate on that?
   # Verify: Only Claude responds
   ```

4. **Graceful degradation**
   ```bash
   go run . --gemini /nonexistent/path
   # Verify: Warning shown, Claude-only mode works
   ```

5. **Permission handling**
   ```
   > Read the contents of go.mod
   # Verify: Permission prompt appears (unless --yolo)
   ```

### Run with debug logging

```bash
go run . --debug 2>debug.log
```

## Reference Libraries

Local reference implementations (see Constitution Principle VI):

| Library | Path | Use For |
|---------|------|---------|
| Bubble Tea | `../bubbletea` | Framework patterns, examples |
| Bubbles | `../bubbles` | Component usage |
| Lip Gloss | `../lipgloss` | Styling patterns |
| ACP SDK Examples | `../claude-code`, `../gemini` | ACP connection patterns |

## Troubleshooting

### Claude fails to connect

```
Error: initialize error: ...
```

1. Check Node.js is installed: `node --version`
2. Check Claude Code auth: `ls ~/.claude/settings.json`
3. Test manually: `npx -y @zed-industries/claude-code-acp@latest`

### Gemini fails to connect

```
Error: failed to start Gemini: ...
```

1. Check Gemini is installed: `which gemini`
2. Check auth: `gemini --version`
3. Test ACP mode: `gemini --experimental-acp`

### Spinner not animating

Ensure `spinner.Tick` is returned in batch commands when in streaming state.

### Viewport not scrolling

Ensure `viewport.Update(msg)` is called and its command is returned.
