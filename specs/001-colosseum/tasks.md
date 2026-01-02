# Tasks: Colosseum Multi-Agent Consultation TUI

**Input**: Design documents from `/specs/001-colosseum/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Manual testing scenarios per DESIGN.md and quickstart.md (no automated tests requested)

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Project location**: `example/colosseum/` (per plan.md project structure)
- **Files**: Flat structure per DESIGN.md specification

---

## Phase 1: Setup

**Purpose**: Project initialization and basic structure

- [x] T001 Create project directory at `example/colosseum/`
- [x] T002 Initialize Go module in `example/colosseum/go.mod`
- [x] T003 Add dependencies: bubbletea, bubbles, lipgloss in `example/colosseum/go.mod`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**CRITICAL**: No user story work can begin until this phase is complete

- [x] T004 [P] Create centralized styles in `example/colosseum/styles.go` per Constitution Principle V
- [x] T005 [P] Create key bindings in `example/colosseum/keys.go` per contracts/cli.md
- [x] T006 [P] Define message types (AgentChunkMsg, AgentCompleteMsg, AgentStatusMsg, PermissionRequestMsg) in `example/colosseum/messages.go`
- [x] T007 Implement Agent struct and interface in `example/colosseum/agent.go` per contracts/agent.go.md
- [x] T008 Implement ACP Client interface methods (SessionUpdate, RequestPermission) in `example/colosseum/agent.go` per research.md Section 4
- [x] T009 Implement agent subprocess spawning with stdin/stdout pipes in `example/colosseum/agent.go`
- [x] T010 Implement Connect(), Prompt(), Cancel(), Close() methods in `example/colosseum/agent.go`
- [x] T011 Implement Orchestrator struct with target parsing in `example/colosseum/orchestrator.go` per contracts/orchestrator.go.md
- [x] T012 Implement context formatting (BuildContext) in `example/colosseum/orchestrator.go`
- [x] T013 Implement sequential agent prompting (Send, NextAgent) in `example/colosseum/orchestrator.go`

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Consult Both Agents on a Design Decision (Priority: P1) MVP

**Goal**: User types a question and both Claude and Gemini respond sequentially with shared context

**Independent Test**: Start Colosseum, type a question, verify both agents respond with relevant perspectives

### Implementation for User Story 1

- [x] T014 [US1] Create Model struct with viewport, input, spinner, state in `example/colosseum/model.go` per data-model.md
- [x] T015 [US1] Implement Init() with agent connection commands in `example/colosseum/model.go`
- [x] T016 [US1] Implement View() with header (status indicators), viewport, and input in `example/colosseum/model.go`
- [x] T017 [US1] Implement Update() for WindowSizeMsg (responsive layout) in `example/colosseum/model.go`
- [x] T018 [US1] Implement Update() for KeyEnter (send message to orchestrator) in `example/colosseum/model.go`
- [x] T019 [US1] Implement Update() for AgentStatusMsg (connection updates) in `example/colosseum/model.go`
- [x] T020 [US1] Implement Update() for AgentChunkMsg (streaming append) in `example/colosseum/model.go`
- [x] T021 [US1] Implement Update() for AgentCompleteMsg (finalize, prompt next) in `example/colosseum/model.go`
- [x] T022 [US1] Implement Update() for spinner.TickMsg (animate during streaming) in `example/colosseum/model.go`
- [x] T023 [US1] Implement message rendering with role-based styling in `example/colosseum/model.go`
- [x] T024 [US1] Create main() with CLI flag parsing in `example/colosseum/main.go` per contracts/cli.md
- [x] T025 [US1] Implement Bubble Tea program initialization in `example/colosseum/main.go`
- [x] T026 [US1] Implement keyboard scrolling (arrows, PgUp/PgDn, Home/End) in `example/colosseum/model.go`

**Checkpoint**: User Story 1 complete - both agents respond to questions with streaming

---

## Phase 4: User Story 2 - Direct a Question to a Specific Agent (Priority: P2)

**Goal**: User can use @claude or @gemini directive to address only that agent

**Independent Test**: Use @claude prefix, verify only Claude responds; use @gemini prefix, verify only Gemini responds

### Implementation for User Story 2

- [x] T027 [US2] Implement ParseTarget() function in `example/colosseum/orchestrator.go` per research.md Section 5
- [x] T028 [US2] Update Send() to use parsed target in `example/colosseum/orchestrator.go`
- [x] T029 [US2] Add @both explicit directive support in `example/colosseum/orchestrator.go`

**Checkpoint**: User Story 2 complete - @claude and @gemini directives work

---

## Phase 5: User Story 3 - Graceful Degradation with Single Agent (Priority: P3)

**Goal**: Colosseum continues operating with one agent when the other fails to connect

**Independent Test**: Misconfigure one agent, verify warning shown and available agent works

### Implementation for User Story 3

- [x] T030 [US3] Implement connection failure handling in `example/colosseum/agent.go`
- [x] T031 [US3] Add warning message display for agent failures in `example/colosseum/model.go`
- [x] T032 [US3] Update orchestrator to skip unavailable agents in `example/colosseum/orchestrator.go`
- [x] T033 [US3] Handle exit when both agents fail (exit code 1) in `example/colosseum/main.go`

**Checkpoint**: User Story 3 complete - single agent mode works

---

## Phase 6: User Story 4 - Session Commands (Priority: P4)

**Goal**: User can use :status, :clear, and :history commands

**Independent Test**: Type each command, verify expected behavior

### Implementation for User Story 4

- [x] T034 [US4] Implement :status command (show connection status) in `example/colosseum/model.go`
- [x] T035 [US4] Implement :clear command (clear viewport) in `example/colosseum/model.go`
- [x] T036 [US4] Implement :history command (display full history) in `example/colosseum/model.go`
- [x] T037 [US4] Implement :exit command (graceful shutdown) in `example/colosseum/model.go`
- [x] T038 [US4] Add command parsing helper for colon-prefixed inputs in `example/colosseum/model.go`

**Checkpoint**: User Story 4 complete - session commands work

---

## Phase 7: User Story 5 - Cancel In-Progress Request (Priority: P5)

**Goal**: User can press Escape to cancel current request

**Independent Test**: Send a message, press Escape while waiting, verify request is cancelled

### Implementation for User Story 5

- [x] T039 [US5] Implement Update() for KeyEscape (cancel request) in `example/colosseum/model.go`
- [x] T040 [US5] Implement Cancel() coordination in `example/colosseum/orchestrator.go`
- [x] T041 [US5] Add cancellation context propagation in `example/colosseum/agent.go`

**Checkpoint**: User Story 5 complete - Escape cancels in-progress requests

---

## Phase 8: Permission Handling (Cross-Cutting)

**Purpose**: Agent permission requests require user approval (or auto-approve with --yolo)

- [x] T042 Implement PermissionRequest struct and state in `example/colosseum/model.go`
- [x] T043 Implement permission dialog rendering in `example/colosseum/model.go`
- [x] T044 Implement permission selection handling in `example/colosseum/model.go`
- [x] T045 Implement auto-approve logic for --yolo flag in `example/colosseum/agent.go`

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [x] T046 [P] Implement --debug flag logging to stderr in `example/colosseum/main.go`
- [x] T047 [P] Implement --no-color flag handling in `example/colosseum/styles.go`
- [x] T048 [P] Implement --timeout flag for agent response timeout in `example/colosseum/agent.go`
- [x] T049 [P] Implement Ctrl+C graceful exit in `example/colosseum/model.go`
- [x] T050 [P] Create README.md with usage documentation in `example/colosseum/README.md`
- [x] T051 Validate all quickstart.md test scenarios work

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-7)**: All depend on Foundational phase completion
  - User stories can proceed sequentially in priority order (P1 → P2 → P3 → P4 → P5)
- **Permission Handling (Phase 8)**: Can start after Foundational, but typically after US1
- **Polish (Phase 9)**: Depends on user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - Core functionality
- **User Story 2 (P2)**: Builds on US1 but independently testable
- **User Story 3 (P3)**: Builds on US1 but independently testable
- **User Story 4 (P4)**: Builds on US1 but independently testable
- **User Story 5 (P5)**: Builds on US1 but independently testable

### Within Each User Story

- Model before view rendering
- Update handlers before full integration
- Story complete before moving to next priority

### Parallel Opportunities

- **Phase 2 (Foundational)**: T004, T005, T006 can run in parallel (different files)
- **Phase 9 (Polish)**: T046, T047, T048, T049, T050 can run in parallel (different files/features)

---

## Parallel Example: Foundational Phase

```bash
# Launch all foundational file creation together:
Task: "Create centralized styles in example/colosseum/styles.go"
Task: "Create key bindings in example/colosseum/keys.go"
Task: "Define message types in example/colosseum/messages.go"

# Then sequentially:
Task: "Implement Agent struct and interface in example/colosseum/agent.go"
# (depends on messages.go for AgentChunkMsg etc.)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test with both agents responding to questions
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Both agents respond → MVP!
3. Add User Story 2 → @claude/@gemini directives work
4. Add User Story 3 → Graceful degradation when one agent fails
5. Add User Story 4 → Session commands (:status, :clear, :history, :exit)
6. Add User Story 5 → Escape cancellation
7. Add Phase 8 → Permission handling
8. Add Phase 9 → Polish (debug, no-color, timeout, docs)

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Reference: quickstart.md for test scenarios, research.md for patterns
