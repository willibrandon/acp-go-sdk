package main

import "time"

// Role identifies the sender of a message
type Role string

const (
	RoleUser   Role = "user"
	RoleClaude Role = "claude"
	RoleGemini Role = "gemini"
	RoleSystem Role = "system"
)

// Message represents a single conversation turn
type Message struct {
	Role      Role      // Who sent this message
	Content   string    // Message text (may be partial during streaming)
	Timestamp time.Time // When the message was created
	Streaming bool      // True while content is still arriving
}

// AgentChunkMsg is sent for each streaming chunk from an agent
type AgentChunkMsg struct {
	Agent   string // Agent name ("claude" or "gemini")
	Content string // Text chunk to append
}

// AgentCompleteMsg is sent when an agent finishes responding
type AgentCompleteMsg struct {
	Agent   string // Agent name
	Content string // Complete response text
	Err     error  // Non-nil if response failed or was cancelled
}

// AgentStatusMsg is sent when agent connection status changes
type AgentStatusMsg struct {
	Agent     string // Agent name
	Connected bool   // Connection state
	Err       error  // Error if connection failed
}

// PermissionOption represents a single permission choice
type PermissionOption struct {
	ID   string // Option identifier to return to agent
	Name string // Display name
	Kind string // "allow_once", "allow_always", "deny"
}

// PermissionRequestMsg is sent when an agent needs permission
type PermissionRequestMsg struct {
	Agent   string             // Which agent is asking
	Title   string             // What action needs permission
	Options []PermissionOption // Available choices
}

// PermissionResponseMsg is sent when user responds to permission dialog
type PermissionResponseMsg struct {
	SelectedIndex int // Index of selected option
}

// InitCompleteMsg is sent when all agents have finished initialization
type InitCompleteMsg struct {
	ClaudeConnected bool
	GeminiConnected bool
}

// ErrorMsg is sent for general errors
type ErrorMsg struct {
	Err error
}
