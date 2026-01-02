# Contract: CLI Interface

**Purpose**: Defines the command-line interface for Colosseum.

## Usage

```
colosseum - Multi-agent consultation for design and implementation decisions

Usage:
  colosseum [flags]

Flags:
  --claude string    Claude ACP command (default: "npx -y @zed-industries/claude-code-acp@latest")
  --gemini string    Gemini CLI path (default: "gemini")
  --yolo             Auto-approve all tool permission requests
  --timeout int      Response timeout in seconds (default: 120)
  --no-color         Disable colored output
  --debug            Enable debug logging
  -h, --help         Display help
```

## Flag Specifications

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--claude` | string | `npx -y @zed-industries/claude-code-acp@latest` | Command to spawn Claude Code in ACP mode |
| `--gemini` | string | `gemini` | Path to Gemini CLI binary |
| `--yolo` | bool | false | Auto-approve all permission requests without prompting |
| `--timeout` | int | 120 | Response timeout in seconds |
| `--no-color` | bool | false | Disable ANSI color codes in output |
| `--debug` | bool | false | Enable debug logging to stderr |

## In-Session Commands

| Command | Description |
|---------|-------------|
| `/status` | Display connection status for both agents |
| `/history` | Display message count and jump to top |
| `/clear` | Clear conversation history (viewport only, not agent context) |
| `/exit` | Terminate session and exit |

## Message Routing Directives

| Directive | Behavior |
|-----------|----------|
| `@claude <message>` | Only Claude responds |
| `@gemini <message>` | Only Gemini responds |
| `@both <message>` | Both agents respond (explicit form of default) |
| `<message>` | Both agents respond (Claude first, then Gemini) |

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `Ctrl+C` | Exit application |
| `Escape` | Cancel current request |
| `↑` / `↓` | Scroll conversation history |
| `PgUp` / `PgDn` | Scroll page up/down |
| `Home` / `End` | Jump to start/end of history |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Normal exit |
| 1 | Both agents failed to connect |
| 2 | Invalid command-line arguments |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `NO_COLOR` | If set, disables colors (same as `--no-color`) |
| `COLOSSEUM_DEBUG` | If set, enables debug logging |

## Examples

```bash
# Basic usage
colosseum

# Auto-approve all tool permissions
colosseum --yolo

# Use custom Gemini binary with longer timeout
colosseum --gemini /usr/local/bin/gemini --timeout 300

# Debug mode with no colors
colosseum --debug --no-color
```
