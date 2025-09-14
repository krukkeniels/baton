# Getting Started with Baton

This guide will walk you through installing, configuring, and using Baton for the first time.

## Prerequisites

- **Go 1.21+** (for building from source)
- **Claude Code CLI** (for LLM integration)
- **Git** (optional, for version control)

## Installation

### Option 1: Build from Source

```bash
# Clone the repository
git clone <repository-url>
cd baton

# Build the binary
make build

# Install to your PATH (optional)
make install
```

### Option 2: Download Pre-built Binary

```bash
# Download for your platform
curl -L https://github.com/<user>/baton/releases/latest/download/baton-linux-amd64 -o baton
chmod +x baton
sudo mv baton /usr/local/bin/
```

## Quick Start

### 1. Initialize Workspace

```bash
# Create a new Baton workspace
baton init

# This creates:
# - baton.yaml (configuration file)
# - baton.db (SQLite database)
# - plan.md (sample plan file)
```

### 2. Configure Your Project

Edit the generated `plan.md` file with your project requirements:

```markdown
# My Project Plan

## Functional Requirements

**FR-1**: The system shall allow users to create tasks.
**FR-2**: The system shall allow users to view task lists.
**FR-3**: The system shall allow users to mark tasks complete.

## Non-Functional Requirements

**NFR-1**: The system shall respond within 200ms.
**NFR-2**: The system shall be available 99.9% of the time.
```

### 3. Ingest Your Plan

```bash
# Parse the plan file and extract requirements
baton ingest plan.md
```

### 4. Check Status

```bash
# View workspace status
baton status

# See what task would be selected next
baton tasks next

# List all tasks
baton tasks list
```

### 5. Run Your First Cycle

```bash
# Execute a dry-run first to see what would happen
baton start --dry-run

# Execute a real cycle (requires Claude Code CLI)
baton start
```

## Configuration

Baton uses a `baton.yaml` configuration file. Key settings:

```yaml
# Core settings
plan_file: "./plan.md"
workspace: "./"
database: "./baton.db"
mcp_port: 8080

# LLM integration
llm:
  primary: "claude"
  claude:
    command: "claude"
    headless_args: ["-p"]
    output_format: "stream-json"

# Task selection
selection:
  algorithm: "priority_dependency"
  dependency_strict: true
  prefer_leaf_tasks: true
```

## Basic Usage Patterns

### Working with Tasks

```bash
# List tasks by state
baton tasks list --state ready_for_implementation

# List high-priority tasks
baton tasks list --priority 8

# Update task state manually
baton tasks update --id task-123 --state implementing --note "Started work"

# View task selection reasoning
baton tasks next
```

### Plan Management

```bash
# Re-ingest plan after changes
baton ingest plan.md

# Check for validation issues
baton ingest --validate-only plan.md
```

### Monitoring Progress

```bash
# Overall workspace status
baton status

# JSON output for automation
baton status --json

# View task history
baton cycles show --task task-123
```

## Understanding the Cycle

Each `baton start` executes one cycle:

1. **Context Reset**: Clears conversational memory
2. **Task Selection**: Picks next task based on priority/dependencies
3. **Agent Execution**: Runs appropriate agent (architect, developer, reviewer)
4. **Completion Handshake**: Ensures state updates and handovers
5. **Audit**: Records complete cycle execution

## State Machine Flow

Tasks progress through these states:

```
ready_for_plan → planning → ready_for_implementation → implementing
→ ready_for_code_review → reviewing → ready_for_commit → committing → DONE
```

With error paths leading to `needs_fixes` → `fixing` → back to review.

## Agent Roles

- **Architect**: Handles `ready_for_plan` and `planning` states
- **Developer**: Handles `implementing` and `fixing` states
- **Reviewer**: Handles `reviewing` state

Each agent creates specific handover artifacts:
- **Implementation Plan** (architect → developer)
- **Change Summary** (developer → reviewer)
- **Review Findings** (reviewer → next action)

## Troubleshooting

### Common Issues

1. **"No selectable tasks"**
   - Check if tasks are blocked by dependencies
   - Run `baton tasks list` to see all tasks
   - Use `baton tasks next` to see selection reasoning

2. **"LLM client not available"**
   - Ensure Claude Code CLI is installed and in PATH
   - Test with `claude --version`
   - Check configuration in `baton.yaml`

3. **"State transition failed"**
   - Review required handover artifacts
   - Check state machine transitions
   - Use `--dry-run` to preview changes

### Debug Mode

```bash
# Enable verbose output
baton start --verbose

# Run dry-run to see what would happen
baton start --dry-run

# Check configuration
baton config validate
```

## Advanced Usage

### Custom Agents

Edit `baton.yaml` to customize agent behavior:

```yaml
agents:
  developer:
    permissions:
      can_execute_commands: true
      allowed_commands: ["git", "npm", "go", "python"]
```

### Integration with CI/CD

```bash
# Non-interactive usage
baton status --json | jq '.completion_rate'

# Automated cycle execution
baton start --timeout 10m
```

### Multi-Project Setup

```bash
# Different workspace per project
baton --workspace ./project-a start
baton --workspace ./project-b start

# Different configuration per project
baton --config ./project-a/baton.yaml start
```

## Best Practices

1. **Small, Focused Tasks**: Break work into single-state transitions
2. **Clear Requirements**: Use specific, testable requirement statements
3. **Regular Cycles**: Run cycles frequently for continuous progress
4. **Monitor Dependencies**: Keep task dependencies up to date
5. **Review Handovers**: Ensure artifacts contain sufficient detail

## Next Steps

- Read the [Configuration Guide](CONFIGURATION.md)
- Learn about [MCP API Integration](MCP_API.md)
- Explore [Advanced Workflows](ADVANCED.md)
- Join the community discussions

## Getting Help

- Check the documentation: `baton help <command>`
- Review status and logs: `baton status --verbose`
- File issues: [GitHub Issues](https://github.com/<user>/baton/issues)
- Join discussions: [GitHub Discussions](https://github.com/<user>/baton/discussions)