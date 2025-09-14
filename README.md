# Baton - CLI Orchestrator for LLM-Driven Task Execution

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

Baton is a CLI orchestrator that advances work one task state at a time through cycle-based execution. Each cycle advances exactly one task by one valid state transition, with context cleared between cycles and formal handover artifacts to bridge cycles.

## Features

- **🔄 Cycle-Based Execution**: One task, one transition, one cycle
- **🧠 State Machine**: Deterministic state transitions with validation
- **📝 Plan Integration**: Extract requirements from markdown plan files
- **🔗 MCP Protocol**: JSON-RPC 2.0 server for LLM integration
- **🤖 Claude Code Integration**: Seamless headless mode integration
- **📊 Task Selection**: Priority and dependency-based algorithms
- **📋 Audit Trail**: Complete cycle execution logging
- **🛑 Handover Artifacts**: Structured knowledge transfer between cycles

## Quick Start

### Installation

**🚀 One-line install (Linux/macOS):**
```bash
curl -fsSL https://raw.githubusercontent.com/krukkeniels/baton/main/install.sh | bash
```

**🪟 Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/krukkeniels/baton/main/install.ps1 | iex
```

**📦 Package Managers:**
```bash
# Homebrew (macOS/Linux)
brew tap krukkeniels/tap
brew install baton

# Go install (if you have Go)
go install github.com/krukkeniels/baton@latest
```

**📋 Manual Download:**
Download pre-built binaries from [GitHub Releases](https://github.com/krukkeniels/baton/releases)

### Initialize Workspace

```bash
# Create new workspace
baton init

# This creates:
# - baton.yaml (configuration)
# - baton.db (SQLite database)
# - plan.md (sample plan file)
```

### Basic Usage

```bash
# Ingest plan file and extract requirements
baton ingest plan.md

# Check workspace status
baton status

# See what task would be selected next
baton tasks next

# Execute one cycle
baton start

# Execute dry run cycle
baton start --dry-run

# List all tasks
baton tasks list

# Update task state manually
baton tasks update --id task-123 --state implementing --note "Starting work"
```

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Baton CLI     │───▶│  MCP Server     │◀───│  Claude Code    │
│  (Main Loop)    │    │  (Port 8080)    │    │   (Headless)    │
│  - Task Sel.    │    │  - Task Ops     │    │  - Agent Logic  │
│  - State Mgmt   │    │  - Artifacts    │    │  - Execution    │
│  - Handshake    │    │  - Requirements │    │                │
└─────────┬───────┘    └─────────┬───────┘    └─────────────────┘
         │                       │
         ▼                       ▼
┌─────────────────┐    ┌─────────────────┐
│   SQLite DB     │    │  Plan File      │
│ - tasks         │    │  (Markdown)     │
│ - requirements  │    │  - Vision       │
│ - artifacts     │    │  - Requirements │
│ - audit_logs    │    │  - Roadmap      │
└─────────────────┘    └─────────────────┘
```

## State Machine

Tasks flow through a deterministic state machine:

```
ready_for_plan
→ planning
→ ready_for_implementation
→ implementing
→ ready_for_code_review
→ reviewing
→ (ready_for_commit | needs_fixes)
→ (committing | fixing)
→ (DONE | ready_for_code_review)
```

## Cycle Execution

Each cycle follows this sequence:

1. **Context Reset**: Clear conversational memory
2. **Rehydrate**: Load context from stored sources
3. **Select**: Choose next task based on priority/dependencies
4. **Execute**: Run LLM agent via Claude Code headless mode
5. **Handshake**: Enforce completion and state updates
6. **Audit**: Record full cycle execution
7. **Stop**: End cycle, prepare for next

## MCP API

Baton exposes a JSON-RPC 2.0 MCP server for LLM integration:

### Task Operations
- `baton.tasks.get_next` - Get next task with selection reasoning
- `baton.tasks.get` - Get specific task by ID
- `baton.tasks.update_state` - Update task state
- `baton.tasks.list` - List tasks with filters

### Artifact Operations
- `baton.artifacts.upsert` - Create/update task artifacts
- `baton.artifacts.get` - Get specific artifact
- `baton.artifacts.list` - List task artifacts

### Plan & Requirements
- `baton.plan.read` - Read plan file contents
- `baton.requirements.list` - List requirements with filters

## Configuration

Baton uses YAML configuration with support for:

- **LLM Integration**: Claude Code, OpenAI CLI support
- **Agent Policies**: Role-based permissions and routing
- **Task Selection**: Priority algorithms and tie-breakers
- **Completion Handshake**: Retry logic and validation
- **Security**: Command allowlists and secret redaction

```yaml
plan_file: "./plan.md"
workspace: "./"
database: "./baton.db"
mcp_port: 8080

llm:
  primary: "claude"
  claude:
    command: "claude"
    headless_args: ["-p"]
    output_format: "stream-json"

selection:
  algorithm: "priority_dependency"
  dependency_strict: true
  prefer_leaf_tasks: true
```

## Development

```bash
# Install dependencies
go mod download

# Format and test
make dev

# Run tests
make test
make integration-test

# Build release binaries
make release
```

## Requirements

- **Go 1.21+** for building
- **Claude Code CLI** for LLM integration
- **SQLite** (embedded, no separate install)

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

For major changes, please open an issue first to discuss the proposed changes.

## Roadmap

- [ ] **Multi-LLM Support**: OpenAI, Anthropic API integration
- [ ] **Web UI**: Browser-based task management interface
- [ ] **Team Collaboration**: Multi-user workspaces
- [ ] **CI/CD Integration**: GitHub Actions, GitLab CI support
- [ ] **Plugin System**: Custom agents and selection algorithms
- [ ] **Metrics & Analytics**: Task completion analytics

---

**Built with ♥️ for systematic software development**