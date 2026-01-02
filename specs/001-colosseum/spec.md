# Feature Specification: Colosseum Multi-Agent Consultation TUI

**Feature Branch**: `001-colosseum`
**Created**: 2026-01-01
**Status**: Draft
**Input**: User description: "Multi-agent consultation TUI for design and implementation decisions"

## Overview

Colosseum provides a unified terminal interface for consulting multiple AI agents (Claude Code and Gemini) simultaneously during software design and implementation decisions. Both agents participate in the same conversation, see each other's responses, and can build on or challenge each other's reasoning.

### Problem Statement

During software design and implementation, developers face decisions that benefit from multiple perspectives:

- Architectural trade-offs with no clear "right" answer
- Implementation approaches with different cost/complexity profiles
- Debugging complex issues where fresh eyes help
- API design where consistency and ergonomics matter
- Technology selection with long-term implications

A single AI assistant provides one perspective. Different models have different strengths, training data, and reasoning patterns. Consulting multiple models manually is tedious and loses conversational context.

### Solution

Colosseum addresses this by providing a single terminal interface where:
- Both Claude and Gemini agents are connected via ACP (Agent Communication Protocol)
- Users can address both agents simultaneously or target specific agents
- Agents see each other's responses and can reference, build upon, or respectfully disagree
- Conversation context is maintained throughout the session

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Consult Both Agents on a Design Decision (Priority: P1)

A developer opens Colosseum and types a question about a design decision (e.g., database choice). Both Claude and Gemini respond sequentially, each providing their perspective. The developer can read both viewpoints and make an informed decision.

**Why this priority**: This is the core value proposition of Colosseum - getting multiple AI perspectives on a single question with shared context.

**Independent Test**: Can be fully tested by starting Colosseum, typing a question, and verifying both agents respond with relevant perspectives. Delivers immediate value by reducing manual effort of consulting multiple AIs separately.

**Acceptance Scenarios**:

1. **Given** Colosseum is running with both agents connected, **When** user types a question and presses Enter, **Then** Claude responds first, followed by Gemini, and both responses are visible in the conversation viewport
2. **Given** user has asked a question and received responses, **When** user scrolls the viewport, **Then** the full conversation history is visible and navigable
3. **Given** both agents have responded, **When** Gemini's response references Claude's points, **Then** the context sharing is working correctly

---

### User Story 2 - Direct a Question to a Specific Agent (Priority: P2)

A developer wants to follow up with a specific agent about their previous response. They use the @claude or @gemini directive to address only that agent while maintaining the full conversation context.

**Why this priority**: Enables focused follow-up without redundant responses, essential for efficient conversation flow.

**Independent Test**: Can be tested by using @claude or @gemini prefix and verifying only the targeted agent responds while the other remains silent.

**Acceptance Scenarios**:

1. **Given** a conversation is in progress, **When** user types "@claude What about using libSQL?", **Then** only Claude responds to this message
2. **Given** a conversation is in progress, **When** user types "@gemini Can you elaborate?", **Then** only Gemini responds to this message
3. **Given** user directed a message to one agent, **When** user sends the next message without a directive, **Then** both agents respond (default behavior restored)

---

### User Story 3 - Graceful Degradation with Single Agent (Priority: P3)

A developer starts Colosseum but Gemini CLI is not installed or authenticated. Colosseum warns the user and continues operating with only Claude available.

**Why this priority**: Ensures the tool remains useful even when one agent is unavailable, improving reliability.

**Independent Test**: Can be tested by intentionally misconfiguring one agent and verifying Colosseum still functions with the available agent.

**Acceptance Scenarios**:

1. **Given** Claude is available but Gemini fails to connect, **When** Colosseum starts, **Then** user sees a warning message and can continue with Claude only
2. **Given** Gemini is available but Claude fails to connect, **When** Colosseum starts, **Then** user sees a warning message and can continue with Gemini only
3. **Given** only one agent is connected, **When** user sends a message, **Then** only the connected agent responds

---

### User Story 4 - Session Commands (Priority: P4)

A developer wants to check connection status, clear the conversation history, or view the full history. They use built-in commands (:status, :clear, :history) to manage the session.

**Why this priority**: Provides session management capabilities for longer conversations.

**Independent Test**: Can be tested by typing each command and verifying the expected behavior.

**Acceptance Scenarios**:

1. **Given** a conversation is in progress, **When** user types ":status", **Then** connection status for both agents is displayed
2. **Given** a conversation has messages, **When** user types ":clear", **Then** conversation history is cleared and viewport is empty
3. **Given** a long conversation exists, **When** user types ":history", **Then** the full conversation history is displayed

---

### User Story 5 - Cancel In-Progress Request (Priority: P5)

A developer sends a complex question but realizes they need to rephrase it. They press Escape to cancel the current request before the agents finish responding.

**Why this priority**: Provides user control over long-running operations.

**Independent Test**: Can be tested by sending a message and pressing Escape while waiting for response.

**Acceptance Scenarios**:

1. **Given** an agent is currently responding, **When** user presses Escape, **Then** the request is cancelled and input returns to idle state
2. **Given** request was cancelled, **When** user types a new message, **Then** the new message is sent normally

---

### Edge Cases

- What happens when both agents fail to connect at startup? Application exits with an error message explaining the issue.
- How does the system handle a 120-second timeout on agent responses? Display a timeout message and return to idle state, allowing user to retry.
- What happens when an agent errors mid-response? Display the partial response with an error indicator, continue with the other agent if available.
- How does the system handle very long agent responses? Viewport scrolls automatically and user can navigate using keyboard.
- What happens when user input contains special characters or is very long? Input is passed through to agents as-is; agents handle interpretation.
- What happens when user denies a permission request? Agent receives denial, continues conversation without that tool action.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST spawn and manage Claude Code subprocess via ACP connection
- **FR-002**: System MUST spawn and manage Gemini CLI subprocess via ACP connection
- **FR-003**: System MUST display a scrollable viewport showing conversation history with role-based styling
- **FR-004**: System MUST provide a text input field for user messages
- **FR-005**: System MUST parse @claude, @gemini, and @both directives to route messages
- **FR-006**: System MUST stream agent responses in real-time with visual indicator during streaming
- **FR-007**: System MUST maintain conversation history visible to both agents for context
- **FR-008**: System MUST format context for each agent showing the full conversation with participant labels
- **FR-009**: System MUST display connection status indicators for both agents in the header
- **FR-010**: System MUST support keyboard navigation (scroll, page up/down, home/end)
- **FR-011**: System MUST support request cancellation via Escape key
- **FR-012**: System MUST support graceful exit via Ctrl+C or :exit command
- **FR-013**: System MUST handle agent connection failures gracefully, continuing with available agent(s)
- **FR-014**: System MUST timeout agent responses after a configurable period (default 120 seconds)
- **FR-015**: System MUST support :status, :clear, and :history commands
- **FR-016**: System MUST use distinct visual styling for each participant (User, Claude, Gemini)
- **FR-017**: System MUST respond to terminal resize events and adjust layout accordingly
- **FR-018**: System MUST prompt user to approve or deny agent permission requests (e.g., file read, file write, command execution)
- **FR-019**: When --yolo flag is set, system MUST auto-approve all agent permission requests without prompting
- **FR-020**: When --debug flag is set, system MUST output debug logging information

### Key Entities

- **Message**: Represents a single conversation turn with role (user/claude/gemini/system), content, timestamp, and streaming state
- **Agent**: Represents a connected AI agent with name, connection state, session ID, and response buffer
- **Conversation**: The ordered sequence of messages shared between user and agents within a session

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can start Colosseum and have both agents connected within 10 seconds on typical hardware
- **SC-002**: Users can send a message and see both agent responses begin streaming within 5 seconds
- **SC-003**: Conversation viewport maintains smooth scroll performance with 100+ messages
- **SC-004**: Users successfully complete a 10-turn multi-agent consultation without errors or lost context
- **SC-005**: When one agent is unavailable, users can still complete consultations with the available agent
- **SC-006**: Users can navigate conversation history using keyboard (scroll, page up/down) without mouse
- **SC-007**: Agent responses correctly reference previous conversation context in 90% of follow-up questions

## Scope

### In Scope

- Two agents: Claude Code (via ACP) and Gemini CLI (via ACP)
- Terminal-based user interface using Bubble Tea framework
- Sequential agent responses (Claude first, then Gemini)
- Session-based conversation history (not persisted)
- CLI flags for configuration (--claude, --gemini, --yolo, --timeout, --no-color, --debug)
- Graceful degradation when one agent unavailable

### Out of Scope (Future Considerations)

- Additional agents (GPT, local models via Ollama)
- Parallel response streaming (both agents responding simultaneously)
- Conversation export/import (save/load sessions)
- Voting/ranking mode for responses
- Web UI version
- Configuration persistence (remember preferences)
- Persistent conversation history across sessions

## Clarifications

### Session 2026-01-01

- Q: What does --yolo flag do? → A: Auto-approves all agent permission requests without prompting the user
- Q: How are agent permission requests handled? → A: User is prompted to approve/deny unless --yolo is set
- Q: What does --debug flag do? → A: Outputs debug logging information

## Assumptions

- Users have Node.js installed (required for claude-code-acp via npx)
- Users have Gemini CLI installed and authenticated
- Users have Claude Code authenticated (~/.claude/settings.json exists)
- Terminal supports 256 colors for styling
- Terminal provides accurate window size information
- ACP protocol is stable and both agents implement it consistently

## Dependencies

- Local reference libraries available: Bubble Tea, Bubbles, Lip Gloss, Huh (Charm ecosystem)
- ACP Go SDK for agent communication
- External runtime: claude-code-acp (via npx) and gemini CLI
