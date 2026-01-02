package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Target specifies which agent(s) should respond to a message
type Target int

const (
	TargetBoth   Target = iota // Both agents respond (Claude first)
	TargetClaude               // Only Claude responds
	TargetGemini               // Only Gemini responds
)

// Orchestrator coordinates multi-agent conversations
type Orchestrator struct {
	claude  *Agent
	gemini  *Agent
	history []Message
	pending []string // Agents waiting to respond
	mu      sync.Mutex
	ctx     context.Context
}

// NewOrchestrator creates a new orchestrator with the given agents
func NewOrchestrator(ctx context.Context, claude, gemini *Agent) *Orchestrator {
	return &Orchestrator{
		claude:  claude,
		gemini:  gemini,
		history: make([]Message, 0),
		pending: make([]string, 0),
		ctx:     ctx,
	}
}

// ParseTarget extracts routing directive from user input
func ParseTarget(input string) (Target, string) {
	input = strings.TrimSpace(input)

	switch {
	case strings.HasPrefix(strings.ToLower(input), "@claude "):
		return TargetClaude, strings.TrimSpace(input[8:])
	case strings.HasPrefix(strings.ToLower(input), "@gemini "):
		return TargetGemini, strings.TrimSpace(input[8:])
	case strings.HasPrefix(strings.ToLower(input), "@both "):
		return TargetBoth, strings.TrimSpace(input[6:])
	default:
		return TargetBoth, input
	}
}

// Send processes user input and initiates agent responses
func (o *Orchestrator) Send(input string) tea.Cmd {
	target, cleanedInput := ParseTarget(input)

	// Add user message to history
	o.mu.Lock()
	o.history = append(o.history, Message{
		Role:      RoleUser,
		Content:   cleanedInput,
		Timestamp: time.Now(),
	})

	// Determine which agents should respond
	o.pending = nil
	switch target {
	case TargetBoth:
		if o.claude.Connected() {
			o.pending = append(o.pending, "claude")
		}
		if o.gemini.Connected() {
			o.pending = append(o.pending, "gemini")
		}
	case TargetClaude:
		if o.claude.Connected() {
			o.pending = append(o.pending, "claude")
		}
	case TargetGemini:
		if o.gemini.Connected() {
			o.pending = append(o.pending, "gemini")
		}
	}
	o.mu.Unlock()

	return o.promptNextAgent()
}

// NextAgent prompts the next pending agent, if any
func (o *Orchestrator) NextAgent() tea.Cmd {
	return o.promptNextAgent()
}

func (o *Orchestrator) promptNextAgent() tea.Cmd {
	o.mu.Lock()
	defer o.mu.Unlock()

	if len(o.pending) == 0 {
		return nil
	}

	agentName := o.pending[0]
	o.pending = o.pending[1:]

	// Build context for the agent
	context := o.buildContext(agentName)

	// Get the agent and start prompting
	agent := o.getAgent(agentName)
	if agent == nil {
		return nil
	}

	// Add placeholder message for streaming
	o.history = append(o.history, Message{
		Role:      Role(agentName),
		Content:   "",
		Timestamp: time.Now(),
		Streaming: true,
	})

	return tea.Batch(
		agent.Prompt(o.ctx, context),
		agent.WaitForChunk(),
	)
}

func (o *Orchestrator) getAgent(name string) *Agent {
	switch name {
	case "claude":
		return o.claude
	case "gemini":
		return o.gemini
	default:
		return nil
	}
}

// buildContext creates the prompt context for an agent
func (o *Orchestrator) buildContext(forAgent string) string {
	var sb strings.Builder

	// Get the other agent's name
	otherAgent := "Claude"
	if forAgent == "claude" {
		otherAgent = "Gemini"
	}

	sb.WriteString(fmt.Sprintf(`You are participating in a collaborative discussion with a user and another AI assistant (%s). The user is consulting both of you to get multiple perspectives on a technical decision.

`, otherAgent))

	// Add conversation history if there is any
	if len(o.history) > 1 { // More than just the current user message
		sb.WriteString("Previous discussion:\n\n")
		for i, msg := range o.history[:len(o.history)-1] { // Exclude latest (which is current user message)
			if msg.Streaming {
				continue // Skip incomplete messages
			}
			label := ""
			switch msg.Role {
			case RoleUser:
				label = "User"
			case RoleClaude:
				label = "Claude"
			case RoleGemini:
				label = "Gemini"
			case RoleSystem:
				label = "System"
			default:
				label = string(msg.Role)
			}
			sb.WriteString(fmt.Sprintf("[%s]: %s\n\n", label, msg.Content))
			if i < len(o.history)-2 {
				sb.WriteString("\n")
			}
		}
		sb.WriteString("---\n\n")
	}

	sb.WriteString("The user is now addressing you. Provide your perspective. You may reference, build upon, or respectfully disagree with the other assistant's points.\n\n")

	// Add the current user message
	if len(o.history) > 0 {
		currentMsg := o.history[len(o.history)-1]
		if currentMsg.Role == RoleUser {
			sb.WriteString(fmt.Sprintf("[User]: %s\n", currentMsg.Content))
		}
	}

	return sb.String()
}

// AppendChunk appends a streaming chunk to the current agent's message
func (o *Orchestrator) AppendChunk(agent, content string) {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Find the streaming message for this agent
	for i := len(o.history) - 1; i >= 0; i-- {
		if o.history[i].Role == Role(agent) && o.history[i].Streaming {
			o.history[i].Content += content
			return
		}
	}
}

// FinalizeMessage marks the current agent's message as complete
func (o *Orchestrator) FinalizeMessage(agent string, finalContent string) {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Find the streaming message for this agent
	for i := len(o.history) - 1; i >= 0; i-- {
		if o.history[i].Role == Role(agent) && o.history[i].Streaming {
			if finalContent != "" {
				o.history[i].Content = finalContent
			}
			o.history[i].Streaming = false
			return
		}
	}
}

// HasPending returns true if there are agents waiting to respond
func (o *Orchestrator) HasPending() bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	return len(o.pending) > 0
}

// History returns a copy of the conversation history
func (o *Orchestrator) History() []Message {
	o.mu.Lock()
	defer o.mu.Unlock()

	result := make([]Message, len(o.history))
	copy(result, o.history)
	return result
}

// ClearHistory clears the conversation history (viewport only)
func (o *Orchestrator) ClearHistory() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.history = make([]Message, 0)
}

// Status returns connection status for all agents
func (o *Orchestrator) Status() map[string]bool {
	return map[string]bool{
		"claude": o.claude.Connected(),
		"gemini": o.gemini.Connected(),
	}
}

// Cancel cancels all pending agent prompts
func (o *Orchestrator) Cancel() tea.Cmd {
	o.mu.Lock()
	o.pending = nil
	o.mu.Unlock()

	var cmds []tea.Cmd
	if o.claude.Connected() {
		cmds = append(cmds, o.claude.Cancel(o.ctx))
	}
	if o.gemini.Connected() {
		cmds = append(cmds, o.gemini.Cancel(o.ctx))
	}

	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

// Close terminates all agent connections
func (o *Orchestrator) Close() error {
	var errs []error
	if err := o.claude.Close(); err != nil {
		errs = append(errs, fmt.Errorf("claude: %w", err))
	}
	if err := o.gemini.Close(); err != nil {
		errs = append(errs, fmt.Errorf("gemini: %w", err))
	}
	if len(errs) > 0 {
		return fmt.Errorf("close errors: %v", errs)
	}
	return nil
}

// AddSystemMessage adds a system message to the history
func (o *Orchestrator) AddSystemMessage(content string) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.history = append(o.history, Message{
		Role:      RoleSystem,
		Content:   content,
		Timestamp: time.Now(),
	})
}
