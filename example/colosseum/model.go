package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AppState represents the current state of the application
type AppState int

const (
	StateIdle       AppState = iota // Ready for input
	StateWaiting                    // Waiting for agent response
	StateStreaming                  // Receiving streaming response
	StatePermission                 // Showing permission dialog
)

// Model is the main Bubble Tea model
type Model struct {
	// TUI Components
	viewport viewport.Model
	input    textinput.Model
	spinner  spinner.Model
	keys     KeyMap

	// Application State
	orchestrator   *Orchestrator
	state          AppState
	currentAgent   string // Agent currently streaming
	err            error
	claudeStatus   bool
	geminiStatus   bool
	initInProgress int // Number of agents still initializing

	// Permission Dialog
	permRequest   *PermissionRequestMsg
	permSelection int

	// Layout
	width  int
	height int
	ready  bool
}

// NewModel creates a new Model with the given orchestrator
func NewModel(orchestrator *Orchestrator) Model {
	ti := textinput.New()
	ti.Placeholder = "Type a message..."
	ti.Focus()
	ti.CharLimit = 4096
	ti.Width = 80

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return Model{
		input:          ti,
		spinner:        s,
		keys:           DefaultKeyMap(),
		orchestrator:   orchestrator,
		state:          StateIdle,
		initInProgress: 2, // Claude and Gemini
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.spinner.Tick,
		m.orchestrator.claude.Connect(m.orchestrator.ctx),
		m.orchestrator.gemini.Connect(m.orchestrator.ctx),
	)
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 2
		inputHeight := 3
		viewportHeight := msg.Height - headerHeight - inputHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, viewportHeight)
			m.viewport.SetContent(m.renderMessages())
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = viewportHeight
		}

		m.input.Width = msg.Width - 4
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case AgentStatusMsg:
		return m.handleAgentStatus(msg)

	case AgentChunkMsg:
		return m.handleAgentChunk(msg)

	case AgentCompleteMsg:
		return m.handleAgentComplete(msg)

	case PermissionRequestMsg:
		m.permRequest = &msg
		m.permSelection = 0
		m.state = StatePermission
		return m, nil

	case spinner.TickMsg:
		if m.state == StateWaiting || m.state == StateStreaming {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	// Update viewport
	var vpCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)
	cmds = append(cmds, vpCmd)

	// Update input
	var inputCmd tea.Cmd
	m.input, inputCmd = m.input.Update(msg)
	cmds = append(cmds, inputCmd)

	return m, tea.Batch(cmds...)
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle permission dialog keys
	if m.state == StatePermission && m.permRequest != nil {
		switch msg.String() {
		case "up", "k":
			if m.permSelection > 0 {
				m.permSelection--
			}
			return m, nil
		case "down", "j":
			if m.permSelection < len(m.permRequest.Options)-1 {
				m.permSelection++
			}
			return m, nil
		case "enter":
			agent := m.orchestrator.getAgent(m.permRequest.Agent)
			if agent != nil {
				agent.RespondToPermission(m.permSelection)
			}
			m.permRequest = nil
			m.state = StateStreaming
			return m, nil
		case "esc":
			agent := m.orchestrator.getAgent(m.permRequest.Agent)
			if agent != nil {
				agent.RespondToPermission(-1) // Cancel
			}
			m.permRequest = nil
			m.state = StateIdle
			return m, nil
		}
		return m, nil
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "esc":
		if m.state == StateWaiting || m.state == StateStreaming {
			m.state = StateIdle
			return m, m.orchestrator.Cancel()
		}
		return m, nil

	case "enter":
		if m.state != StateIdle || m.input.Value() == "" {
			return m, nil
		}

		input := strings.TrimSpace(m.input.Value())
		m.input.Reset()

		// Check for session commands
		if cmd := m.handleCommand(input); cmd != nil {
			return m, cmd
		}

		m.state = StateWaiting
		m.err = nil

		// Update viewport with new user message
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()

		return m, tea.Batch(
			m.orchestrator.Send(input),
			m.spinner.Tick,
		)

	case "up", "pgup", "home":
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd

	case "down", "pgdown", "end":
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	// Pass to input
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) handleCommand(input string) tea.Cmd {
	switch strings.ToLower(input) {
	case ":exit", ":quit":
		return tea.Quit
	case ":status":
		status := m.orchestrator.Status()
		msg := fmt.Sprintf("Claude: %s, Gemini: %s",
			statusText(status["claude"]),
			statusText(status["gemini"]))
		m.orchestrator.AddSystemMessage(msg)
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		return nil
	case ":clear":
		m.orchestrator.ClearHistory()
		m.viewport.SetContent("")
		return nil
	case ":history":
		// Already showing history in viewport
		m.viewport.GotoTop()
		return nil
	default:
		return nil
	}
}

func statusText(connected bool) string {
	if connected {
		return "connected"
	}
	return "disconnected"
}

func (m Model) handleAgentStatus(msg AgentStatusMsg) (tea.Model, tea.Cmd) {
	m.initInProgress--

	switch msg.Agent {
	case "claude":
		m.claudeStatus = msg.Connected
		if msg.Err != nil {
			m.orchestrator.AddSystemMessage(fmt.Sprintf("Claude connection failed: %v", msg.Err))
		}
	case "gemini":
		m.geminiStatus = msg.Connected
		if msg.Err != nil {
			m.orchestrator.AddSystemMessage(fmt.Sprintf("Gemini connection failed: %v", msg.Err))
		}
	}

	// Update viewport with status messages
	m.viewport.SetContent(m.renderMessages())

	// Check if both agents have finished initializing
	if m.initInProgress == 0 {
		if !m.claudeStatus && !m.geminiStatus {
			m.err = fmt.Errorf("both agents failed to connect")
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m Model) handleAgentChunk(msg AgentChunkMsg) (tea.Model, tea.Cmd) {
	if msg.Content == "" {
		return m, nil
	}

	m.state = StateStreaming
	m.currentAgent = msg.Agent
	m.orchestrator.AppendChunk(msg.Agent, msg.Content)

	// Update viewport
	m.viewport.SetContent(m.renderMessages())
	m.viewport.GotoBottom()

	// Wait for next chunk
	agent := m.orchestrator.getAgent(msg.Agent)
	if agent != nil {
		return m, agent.WaitForChunk()
	}
	return m, nil
}

func (m Model) handleAgentComplete(msg AgentCompleteMsg) (tea.Model, tea.Cmd) {
	m.orchestrator.FinalizeMessage(msg.Agent, msg.Content)

	// Update viewport
	m.viewport.SetContent(m.renderMessages())
	m.viewport.GotoBottom()

	// Check if there are more agents to prompt
	if m.orchestrator.HasPending() {
		m.state = StateWaiting
		return m, tea.Batch(
			m.orchestrator.NextAgent(),
			m.spinner.Tick,
		)
	}

	m.state = StateIdle
	m.currentAgent = ""
	return m, nil
}

// View renders the UI
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	var b strings.Builder

	// Header with status
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// Viewport (conversation history)
	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	// Input area
	b.WriteString(m.renderInput())

	return b.String()
}

func (m Model) renderHeader() string {
	claudeIndicator := RenderStatus(m.claudeStatus)
	geminiIndicator := RenderStatus(m.geminiStatus)

	status := ""
	switch m.state {
	case StateWaiting:
		status = m.spinner.View() + " Waiting..."
	case StateStreaming:
		status = m.spinner.View() + " " + m.currentAgent + " is typing..."
	case StatePermission:
		status = "Permission required"
	}

	header := fmt.Sprintf("Colosseum | Claude %s | Gemini %s | %s",
		claudeIndicator, geminiIndicator, status)

	return HeaderStyle.Width(m.width).Render(header)
}

func (m Model) renderInput() string {
	// Show permission dialog if active
	if m.state == StatePermission && m.permRequest != nil {
		return m.renderPermissionDialog()
	}

	var status string
	switch m.state {
	case StateWaiting, StateStreaming:
		status = m.spinner.View() + " "
	}

	input := "> " + m.input.View()
	help := HelpStyle.Render("enter: send | esc: cancel | ctrl+c: quit | @claude/@gemini: direct")

	return InputStyle.Width(m.width).Render(status + input + "\n" + help)
}

func (m Model) renderPermissionDialog() string {
	if m.permRequest == nil {
		return ""
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Permission request from %s:\n", m.permRequest.Agent))
	b.WriteString(fmt.Sprintf("  %s\n\n", m.permRequest.Title))

	for i, opt := range m.permRequest.Options {
		cursor := "  "
		if i == m.permSelection {
			cursor = "> "
		}
		b.WriteString(fmt.Sprintf("%s%s (%s)\n", cursor, opt.Name, opt.Kind))
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("↑/↓: select | Enter: confirm | Esc: cancel"))

	return InputStyle.Width(m.width).Render(b.String())
}

func (m Model) renderMessages() string {
	history := m.orchestrator.History()
	if len(history) == 0 {
		return HelpStyle.Render("No messages yet. Type a question to get started!")
	}

	var b strings.Builder
	for i, msg := range history {
		// Role label
		b.WriteString(RenderRoleName(msg.Role))
		b.WriteString("\n")

		// Content with proper wrapping
		content := msg.Content
		if msg.Streaming {
			content += "▌" // Cursor for streaming
		}

		// Wrap text to viewport width
		wrapped := lipgloss.NewStyle().Width(m.viewport.Width - 2).Render(content)
		b.WriteString(wrapped)

		if i < len(history)-1 {
			b.WriteString("\n\n")
		}
	}

	return b.String()
}
