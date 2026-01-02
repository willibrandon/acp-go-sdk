package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/coder/acp-go-sdk"
)

// Agent represents a connected AI agent with ACP connection
type Agent struct {
	name        string
	command     string
	conn        *acp.ClientSideConnection
	sessionID   acp.SessionId
	cmd         *exec.Cmd
	cancel      context.CancelFunc
	autoApprove bool
	timeout     time.Duration

	mu        sync.Mutex
	buffer    strings.Builder
	chunks    chan string
	connected bool

	// Channel to send permission requests to the TUI
	permChan chan<- PermissionRequestMsg
	// Channel to receive permission responses from TUI
	permRespChan chan int
}

// Ensure Agent implements acp.Client
var _ acp.Client = (*Agent)(nil)

// NewAgent creates a new agent with the given configuration
func NewAgent(name, command string, autoApprove bool, timeout time.Duration, permChan chan<- PermissionRequestMsg) *Agent {
	return &Agent{
		name:         name,
		command:      command,
		autoApprove:  autoApprove,
		timeout:      timeout,
		chunks:       make(chan string, 100),
		permChan:     permChan,
		permRespChan: make(chan int),
	}
}

// Name returns the agent identifier
func (a *Agent) Name() string {
	return a.name
}

// Connected returns true if the agent has an active ACP connection
func (a *Agent) Connected() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.connected
}

// Connect initializes the agent subprocess and ACP connection
func (a *Agent) Connect(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		err := a.connect(ctx)
		if err != nil {
			return AgentStatusMsg{
				Agent:     a.name,
				Connected: false,
				Err:       err,
			}
		}
		return AgentStatusMsg{
			Agent:     a.name,
			Connected: true,
		}
	}
}

func (a *Agent) connect(ctx context.Context) error {
	// Parse command into executable and args
	parts := strings.Fields(a.command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command for agent %s", a.name)
	}

	cmdCtx, cancel := context.WithCancel(ctx)
	a.cancel = cancel

	a.cmd = exec.CommandContext(cmdCtx, parts[0], parts[1:]...)
	// Don't pipe stderr to terminal - it corrupts the TUI
	// In debug mode, users can redirect stderr externally

	stdin, err := a.cmd.StdinPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("stdin pipe: %w", err)
	}

	stdout, err := a.cmd.StdoutPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("stdout pipe: %w", err)
	}

	if err := a.cmd.Start(); err != nil {
		cancel()
		return fmt.Errorf("start command: %w", err)
	}

	// Create ACP connection
	a.conn = acp.NewClientSideConnection(a, stdin, stdout)

	// Initialize
	_, err = a.conn.Initialize(ctx, acp.InitializeRequest{
		ProtocolVersion: acp.ProtocolVersionNumber,
		ClientCapabilities: acp.ClientCapabilities{
			Fs: acp.FileSystemCapability{
				ReadTextFile:  true,
				WriteTextFile: true,
			},
		},
	})
	if err != nil {
		a.Close()
		return fmt.Errorf("initialize: %w", err)
	}

	// Create session
	cwd, _ := os.Getwd()
	resp, err := a.conn.NewSession(ctx, acp.NewSessionRequest{
		Cwd:        cwd,
		McpServers: []acp.McpServer{},
	})
	if err != nil {
		a.Close()
		return fmt.Errorf("new session: %w", err)
	}

	a.mu.Lock()
	a.sessionID = resp.SessionId
	a.connected = true
	a.mu.Unlock()

	return nil
}

// Prompt sends a message to the agent and streams the response
func (a *Agent) Prompt(ctx context.Context, message string) tea.Cmd {
	return func() tea.Msg {
		a.mu.Lock()
		if !a.connected || string(a.sessionID) == "" {
			a.mu.Unlock()
			return AgentCompleteMsg{
				Agent: a.name,
				Err:   fmt.Errorf("agent not connected"),
			}
		}
		sessionID := a.sessionID
		a.buffer.Reset()
		a.mu.Unlock()

		// Set timeout context
		promptCtx := ctx
		if a.timeout > 0 {
			var cancel context.CancelFunc
			promptCtx, cancel = context.WithTimeout(ctx, a.timeout)
			defer cancel()
		}

		// Send prompt - this blocks until response is complete
		// Streaming happens via SessionUpdate callback
		_, err := a.conn.Prompt(promptCtx, acp.PromptRequest{
			SessionId: acp.SessionId(sessionID),
			Prompt:    []acp.ContentBlock{acp.TextBlock(message)},
		})

		// Close chunks channel to signal completion
		close(a.chunks)
		a.chunks = make(chan string, 100)

		a.mu.Lock()
		content := a.buffer.String()
		a.mu.Unlock()

		return AgentCompleteMsg{
			Agent:   a.name,
			Content: content,
			Err:     err,
		}
	}
}

// WaitForChunk returns a command that waits for the next streaming chunk
func (a *Agent) WaitForChunk() tea.Cmd {
	return func() tea.Msg {
		content, ok := <-a.chunks
		if !ok {
			// Channel closed, streaming complete
			return nil
		}
		return AgentChunkMsg{
			Agent:   a.name,
			Content: content,
		}
	}
}

// Cancel cancels any in-progress prompt
func (a *Agent) Cancel(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		a.mu.Lock()
		sessionID := a.sessionID
		a.mu.Unlock()

		if string(sessionID) != "" && a.conn != nil {
			_ = a.conn.Cancel(ctx, acp.CancelNotification{SessionId: acp.SessionId(sessionID)})
		}

		return AgentCompleteMsg{
			Agent: a.name,
			Err:   fmt.Errorf("cancelled"),
		}
	}
}

// Close terminates the agent subprocess and releases resources
func (a *Agent) Close() error {
	a.mu.Lock()
	a.connected = false
	a.mu.Unlock()

	if a.cancel != nil {
		a.cancel()
	}

	if a.cmd != nil && a.cmd.Process != nil {
		return a.cmd.Process.Kill()
	}
	return nil
}

// RespondToPermission sends a permission response
func (a *Agent) RespondToPermission(index int) {
	select {
	case a.permRespChan <- index:
	default:
	}
}

// ---- ACP Client Interface Implementation ----

// SessionUpdate receives streaming updates from the agent
func (a *Agent) SessionUpdate(ctx context.Context, params acp.SessionNotification) error {
	u := params.Update

	switch {
	case u.AgentMessageChunk != nil:
		if u.AgentMessageChunk.Content.Text != nil {
			text := u.AgentMessageChunk.Content.Text.Text
			a.mu.Lock()
			a.buffer.WriteString(text)
			a.mu.Unlock()

			// Send to chunks channel for TUI updates
			select {
			case a.chunks <- text:
			default:
				// Channel full, drop chunk (shouldn't happen with buffered channel)
			}
		}
	case u.ToolCall != nil:
		// Tool call status update - could log for debug
	case u.ToolCallUpdate != nil:
		// Tool call result - could log for debug
	}

	return nil
}

// RequestPermission handles permission requests from the agent
func (a *Agent) RequestPermission(ctx context.Context, params acp.RequestPermissionRequest) (acp.RequestPermissionResponse, error) {
	if a.autoApprove {
		// Prefer allow option
		for _, o := range params.Options {
			if o.Kind == acp.PermissionOptionKindAllowOnce || o.Kind == acp.PermissionOptionKindAllowAlways {
				return acp.RequestPermissionResponse{
					Outcome: acp.RequestPermissionOutcome{
						Selected: &acp.RequestPermissionOutcomeSelected{OptionId: o.OptionId},
					},
				}, nil
			}
		}
		// Fallback to first option
		if len(params.Options) > 0 {
			return acp.RequestPermissionResponse{
				Outcome: acp.RequestPermissionOutcome{
					Selected: &acp.RequestPermissionOutcomeSelected{OptionId: params.Options[0].OptionId},
				},
			}, nil
		}
		return acp.RequestPermissionResponse{
			Outcome: acp.RequestPermissionOutcome{
				Cancelled: &acp.RequestPermissionOutcomeCancelled{},
			},
		}, nil
	}

	// Send permission request to TUI
	title := ""
	if params.ToolCall.Title != nil {
		title = *params.ToolCall.Title
	}

	options := make([]PermissionOption, len(params.Options))
	for i, o := range params.Options {
		options[i] = PermissionOption{
			ID:   string(o.OptionId),
			Name: o.Name,
			Kind: string(o.Kind),
		}
	}

	a.permChan <- PermissionRequestMsg{
		Agent:   a.name,
		Title:   title,
		Options: options,
	}

	// Wait for response from TUI
	select {
	case idx := <-a.permRespChan:
		if idx >= 0 && idx < len(params.Options) {
			return acp.RequestPermissionResponse{
				Outcome: acp.RequestPermissionOutcome{
					Selected: &acp.RequestPermissionOutcomeSelected{OptionId: params.Options[idx].OptionId},
				},
			}, nil
		}
		return acp.RequestPermissionResponse{
			Outcome: acp.RequestPermissionOutcome{
				Cancelled: &acp.RequestPermissionOutcomeCancelled{},
			},
		}, nil
	case <-ctx.Done():
		return acp.RequestPermissionResponse{
			Outcome: acp.RequestPermissionOutcome{
				Cancelled: &acp.RequestPermissionOutcomeCancelled{},
			},
		}, nil
	}
}

// ReadTextFile implements acp.Client
func (a *Agent) ReadTextFile(ctx context.Context, params acp.ReadTextFileRequest) (acp.ReadTextFileResponse, error) {
	if !filepath.IsAbs(params.Path) {
		return acp.ReadTextFileResponse{}, fmt.Errorf("path must be absolute: %s", params.Path)
	}
	b, err := os.ReadFile(params.Path)
	if err != nil {
		return acp.ReadTextFileResponse{}, fmt.Errorf("read %s: %w", params.Path, err)
	}
	content := string(b)
	if params.Line != nil || params.Limit != nil {
		lines := strings.Split(content, "\n")
		start := 0
		if params.Line != nil && *params.Line > 0 {
			start = min(max(*params.Line-1, 0), len(lines))
		}
		end := len(lines)
		if params.Limit != nil && *params.Limit > 0 {
			if start+*params.Limit < end {
				end = start + *params.Limit
			}
		}
		content = strings.Join(lines[start:end], "\n")
	}
	return acp.ReadTextFileResponse{Content: content}, nil
}

// WriteTextFile implements acp.Client
func (a *Agent) WriteTextFile(ctx context.Context, params acp.WriteTextFileRequest) (acp.WriteTextFileResponse, error) {
	if !filepath.IsAbs(params.Path) {
		return acp.WriteTextFileResponse{}, fmt.Errorf("path must be absolute: %s", params.Path)
	}
	dir := filepath.Dir(params.Path)
	if dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return acp.WriteTextFileResponse{}, fmt.Errorf("mkdir %s: %w", dir, err)
		}
	}
	if err := os.WriteFile(params.Path, []byte(params.Content), 0o644); err != nil {
		return acp.WriteTextFileResponse{}, fmt.Errorf("write %s: %w", params.Path, err)
	}
	return acp.WriteTextFileResponse{}, nil
}

// Terminal methods - implement as no-ops for now
func (a *Agent) CreateTerminal(ctx context.Context, params acp.CreateTerminalRequest) (acp.CreateTerminalResponse, error) {
	return acp.CreateTerminalResponse{TerminalId: "term-1"}, nil
}

func (a *Agent) TerminalOutput(ctx context.Context, params acp.TerminalOutputRequest) (acp.TerminalOutputResponse, error) {
	return acp.TerminalOutputResponse{Output: "", Truncated: false}, nil
}

func (a *Agent) ReleaseTerminal(ctx context.Context, params acp.ReleaseTerminalRequest) (acp.ReleaseTerminalResponse, error) {
	return acp.ReleaseTerminalResponse{}, nil
}

func (a *Agent) WaitForTerminalExit(ctx context.Context, params acp.WaitForTerminalExitRequest) (acp.WaitForTerminalExitResponse, error) {
	return acp.WaitForTerminalExitResponse{}, nil
}

func (a *Agent) KillTerminalCommand(ctx context.Context, params acp.KillTerminalCommandRequest) (acp.KillTerminalCommandResponse, error) {
	return acp.KillTerminalCommandResponse{}, nil
}
