# Colosseum

A multi-agent consultation TUI for getting perspectives from both Claude and Gemini on technical decisions.

## Installation

```bash
cd example/colosseum
go build .
```

## Prerequisites

- **Go 1.21+**
- **Node.js** (for Claude Code via npx)
- **Gemini CLI** installed and authenticated
- **Claude Code** authenticated (`~/.claude/settings.json` exists)

## Usage

```bash
./colosseum [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--claude` | `npx -y @zed-industries/claude-code-acp@latest` | Claude ACP command |
| `--gemini` | `gemini` | Gemini CLI path |
| `--yolo` | `false` | Auto-approve all tool permission requests |
| `--timeout` | `120` | Response timeout in seconds |
| `--no-color` | `false` | Disable colored output |
| `--debug` | `false` | Enable debug logging to stderr |

### Environment Variables

| Variable | Description |
|----------|-------------|
| `NO_COLOR` | If set, disables colors |
| `COLOSSEUM_DEBUG` | If set, enables debug logging |

## Message Routing

| Syntax | Behavior |
|--------|----------|
| `@claude <message>` | Only Claude responds |
| `@gemini <message>` | Only Gemini responds |
| `@both <message>` | Both agents respond (Claude first) |
| `<message>` | Both agents respond (default) |

## Session Commands

| Command | Description |
|---------|-------------|
| `/help` | Show available commands |
| `/status` | Show connection status for both agents |
| `/clear` | Clear conversation viewport |
| `/history` | Show message count and jump to top |
| `/exit` | Exit the application |

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `Ctrl+C` | Quit application |
| `Esc` | Cancel current request |
| `↑/↓` | Scroll history |
| `PgUp/PgDn` | Page up/down |
| `Home/End` | Jump to start/end |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Normal exit |
| 1 | Both agents failed to connect |

## Examples

```bash
# Basic usage
./colosseum

# Auto-approve all tool permissions
./colosseum --yolo

# Use custom Gemini path with longer timeout
./colosseum --gemini /usr/local/bin/gemini --timeout 300

# Debug mode
./colosseum --debug 2>debug.log
```

## Troubleshooting

### Claude fails to connect

1. Check Node.js is installed: `node --version`
2. Check Claude Code auth: `ls ~/.claude/settings.json`
3. Test manually: `npx -y @zed-industries/claude-code-acp@latest`

### Gemini fails to connect

1. Check Gemini is installed: `which gemini`
2. Check auth: `gemini --version`

### Single agent mode

If one agent fails to connect, Colosseum continues with the available agent. A warning message is displayed. Use `:status` to check connection status.
