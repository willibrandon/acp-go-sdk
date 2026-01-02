# Contract: Agent Interface

**Purpose**: Defines the interface for AI agents that can participate in Colosseum conversations.

## Interface Definition

```go
package colosseum

import (
    "context"
    tea "github.com/charmbracelet/bubbletea"
)

// Agent represents an AI agent that can participate in conversations.
// Both Claude and Gemini agents implement this interface.
type Agent interface {
    // Name returns the agent identifier ("claude" or "gemini")
    Name() string

    // Connected returns true if the agent has an active ACP connection
    Connected() bool

    // Connect initializes the agent subprocess and ACP connection.
    // Returns a tea.Cmd that sends AgentStatusMsg on completion.
    Connect(ctx context.Context) tea.Cmd

    // Prompt sends a message to the agent and streams the response.
    // The message includes formatted conversation context.
    // Returns a tea.Cmd that sends AgentChunkMsg for each chunk
    // and AgentCompleteMsg when finished.
    Prompt(ctx context.Context, message string) tea.Cmd

    // Cancel cancels any in-progress prompt.
    // Returns a tea.Cmd that sends AgentCompleteMsg with cancellation.
    Cancel(ctx context.Context) tea.Cmd

    // Close terminates the agent subprocess and releases resources.
    Close() error
}
```

## Message Types

```go
// AgentChunkMsg is sent for each streaming chunk from an agent.
type AgentChunkMsg struct {
    Agent   string // Agent name
    Content string // Text chunk
}

// AgentCompleteMsg is sent when an agent finishes responding.
type AgentCompleteMsg struct {
    Agent   string // Agent name
    Content string // Complete response text
    Err     error  // Non-nil if response failed or was cancelled
}

// AgentStatusMsg is sent when agent connection status changes.
type AgentStatusMsg struct {
    Agent     string
    Connected bool
    Err       error
}
```

## Implementation Requirements

1. **Non-blocking**: All methods must return `tea.Cmd`, never block the caller
2. **Streaming**: `Prompt` must stream via `AgentChunkMsg`, not wait for full response
3. **Cancellation**: `Cancel` must be honored within reasonable time (< 1s)
4. **Thread Safety**: All methods must be safe for concurrent calls
5. **Resource Cleanup**: `Close` must terminate subprocess and close connections

## Usage Example

```go
// Connect agents during init
func (m Model) Init() tea.Cmd {
    return tea.Batch(
        m.claude.Connect(context.Background()),
        m.gemini.Connect(context.Background()),
    )
}

// Handle streaming updates
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case AgentChunkMsg:
        m.appendChunk(msg.Agent, msg.Content)
        return m, m.waitForNextChunk(msg.Agent)
    case AgentCompleteMsg:
        m.finalizeMessage(msg.Agent)
        return m, m.orchestrator.NextAgent()
    }
}
```
