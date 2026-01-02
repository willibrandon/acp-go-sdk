package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	defaultClaudeCmd = "npx -y @zed-industries/claude-code-acp@latest"
	defaultGeminiCmd = "gemini --experimental-acp"
	defaultTimeout   = 120 // seconds
)

func main() {
	// CLI flags
	claudeCmd := flag.String("claude", defaultClaudeCmd, "Claude ACP command")
	geminiCmd := flag.String("gemini", defaultGeminiCmd, "Gemini CLI path")
	yolo := flag.Bool("yolo", false, "Auto-approve all tool permission requests")
	timeoutSec := flag.Int("timeout", defaultTimeout, "Response timeout in seconds")
	noColorFlag := flag.Bool("no-color", false, "Disable colored output")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	// Check environment variables
	if os.Getenv("NO_COLOR") != "" {
		*noColorFlag = true
	}
	if os.Getenv("COLOSSEUM_DEBUG") != "" {
		*debug = true
	}

	// Apply no-color setting
	SetNoColor(*noColorFlag)

	// Set up debug logging
	if *debug {
		fmt.Fprintln(os.Stderr, "[debug] Debug mode enabled")
		fmt.Fprintf(os.Stderr, "[debug] Claude command: %s\n", *claudeCmd)
		fmt.Fprintf(os.Stderr, "[debug] Gemini command: %s\n", *geminiCmd)
		fmt.Fprintf(os.Stderr, "[debug] Timeout: %d seconds\n", *timeoutSec)
		fmt.Fprintf(os.Stderr, "[debug] Yolo mode: %v\n", *yolo)
	}

	timeout := time.Duration(*timeoutSec) * time.Second

	// Create context for the application
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create permission channel for agent-to-TUI communication
	permChan := make(chan PermissionRequestMsg, 1)

	// Create agents
	claude := NewAgent("claude", *claudeCmd, *yolo, timeout, permChan)
	gemini := NewAgent("gemini", *geminiCmd, *yolo, timeout, permChan)

	// Create orchestrator
	orchestrator := NewOrchestrator(ctx, claude, gemini)

	// Create model
	model := NewModel(orchestrator)

	// Run Bubble Tea program
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Start permission message relay in background
	go func() {
		for perm := range permChan {
			p.Send(perm)
		}
	}()

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Clean up
	if err := orchestrator.Close(); err != nil && *debug {
		fmt.Fprintf(os.Stderr, "[debug] Cleanup error: %v\n", err)
	}
}
