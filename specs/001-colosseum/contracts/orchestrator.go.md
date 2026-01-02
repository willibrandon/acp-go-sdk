# Contract: Orchestrator Interface

**Purpose**: Coordinates multi-agent conversations and manages conversation history.

## Interface Definition

```go
package colosseum

import (
    "context"
    tea "github.com/charmbracelet/bubbletea"
)

// Target specifies which agent(s) should respond to a message.
type Target int

const (
    TargetBoth   Target = iota // Both agents respond (Claude first)
    TargetClaude               // Only Claude responds
    TargetGemini               // Only Gemini responds
)

// Orchestrator coordinates multi-agent conversations.
type Orchestrator interface {
    // Send processes user input and initiates agent responses.
    // Parses target directive (@claude, @gemini) from input.
    // Returns tea.Cmd that triggers first agent prompt.
    Send(ctx context.Context, input string) tea.Cmd

    // NextAgent prompts the next pending agent, if any.
    // Returns nil if no agents are pending.
    NextAgent(ctx context.Context) tea.Cmd

    // Cancel cancels all pending agent prompts.
    Cancel(ctx context.Context) tea.Cmd

    // History returns the conversation history for display.
    History() []Message

    // Status returns connection status for all agents.
    Status() map[string]bool

    // Close terminates all agent connections.
    Close() error
}
```

## Functions

```go
// ParseTarget extracts routing directive from user input.
// Returns the target and the cleaned input without the directive.
func ParseTarget(input string) (Target, string)

// Example:
//   ParseTarget("@claude What do you think?")
//   // Returns: TargetClaude, "What do you think?"
//
//   ParseTarget("What do you think?")
//   // Returns: TargetBoth, "What do you think?"
```

## Context Formatting

The orchestrator formats conversation history for each agent:

```go
// BuildContext creates the prompt context for an agent.
// Includes conversation history and multi-agent instructions.
func (o *Orchestrator) BuildContext(forAgent string, userMessage string) string
```

**Context Template**:
```text
You are participating in a collaborative discussion with a user and another
AI assistant ({other_agent}). The user is consulting both of you to get
multiple perspectives on a technical decision.

Previous discussion:

[User]: {message_1}

[Claude]: {response_1}

[Gemini]: {response_2}

---

The user is now addressing you. Provide your perspective. You may reference,
build upon, or respectfully disagree with the other assistant's points.

[User]: {current_message}
```

## Implementation Requirements

1. **Sequential Prompting**: When `TargetBoth`, Claude responds first, then Gemini
2. **Shared History**: All agents see the same conversation history
3. **Participant Labels**: History clearly identifies who said what
4. **Thread Safety**: History access must be synchronized
5. **Graceful Degradation**: Works with one agent if other is unavailable

## Usage Example

```go
// In Model.Update for Enter key
case tea.KeyMsg:
    if msg.Type == tea.KeyEnter && m.input.Value() != "" {
        input := m.input.Value()
        m.input.Reset()
        m.state = StateWaiting
        return m, m.orchestrator.Send(ctx, input)
    }

// When agent completes
case AgentCompleteMsg:
    m.finalizeMessage(msg.Agent)
    if cmd := m.orchestrator.NextAgent(ctx); cmd != nil {
        return m, cmd  // More agents pending
    }
    m.state = StateIdle
    return m, nil
```
