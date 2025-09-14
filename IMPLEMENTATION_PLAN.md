# Baton Implementation Plan

**Target**: CLI Orchestrator for LLM-Driven Task Execution (Cycle-Based)

## Technology Stack

- **Language**: Go 1.21+
- **Storage**: SQLite (embedded)
- **CLI Framework**: Cobra + Viper
- **Config Format**: YAML
- **MCP Protocol**: Custom JSON-RPC 2.0 implementation
- **Primary LLM Integration**: Claude Code (headless mode)
- **Distribution**: Single binary

## Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Baton CLI     │───▶│  MCP Server     │◀───│  Claude Code    │
│  (Main Loop)    │    │  (Port 8080)    │    │   (Headless)    │
│  - Task Sel.    │    │  - Task Ops     │    │  - Agent Logic  │
│  - State Mgmt   │    │  - Artifacts    │    │  - Execution    │
│  - Handshake    │    │  - Requirements │    └─────────────────┘
└─────────────────┘    └─────────────────┘
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

## Project Structure

```
baton/
├── cmd/
│   ├── root.go              # Root command and global flags
│   ├── init.go              # Initialize workspace
│   ├── start.go             # Main cycle execution
│   ├── status.go            # Status reporting
│   ├── ingest.go            # Plan file ingestion
│   └── tasks.go             # Task management commands
├── internal/
│   ├── config/
│   │   ├── config.go        # Configuration loading and validation
│   │   └── defaults.go      # Default configuration values
│   ├── storage/
│   │   ├── sqlite.go        # SQLite database operations
│   │   ├── migrations.go    # Database schema migrations
│   │   ├── models.go        # Data models (Task, Requirement, etc.)
│   │   └── queries.go       # SQL query implementations
│   ├── statemachine/
│   │   ├── states.go        # State definitions and transitions
│   │   ├── validation.go    # State transition validation
│   │   └── selection.go     # Task selection algorithm
│   ├── mcp/
│   │   ├── server.go        # MCP server implementation
│   │   ├── handlers.go      # MCP method handlers
│   │   └── protocol.go      # JSON-RPC 2.0 protocol
│   ├── llm/
│   │   ├── client.go        # LLM client interface
│   │   ├── claude.go        # Claude Code integration
│   │   └── fallback.go      # Fallback LLM support
│   ├── cycle/
│   │   ├── engine.go        # Main cycle execution engine
│   │   ├── handshake.go     # Completion handshake logic
│   │   └── handover.go      # Handover artifact management
│   ├── plan/
│   │   ├── parser.go        # Markdown plan file parsing
│   │   └── requirements.go  # Requirements extraction
│   └── audit/
│       ├── logger.go        # Audit trail logging
│       └── models.go        # Audit data structures
├── pkg/
│   └── version/
│       └── version.go       # Version information
├── configs/
│   ├── default.yaml         # Default configuration template
│   └── example.yaml         # Example configuration
├── docs/
│   ├── GETTING_STARTED.md   # Quick start guide
│   ├── CONFIGURATION.md     # Configuration reference
│   └── MCP_API.md          # MCP API documentation
├── test/
│   ├── fixtures/           # Test data and fixtures
│   ├── integration/        # Integration tests
│   └── e2e/               # End-to-end tests
├── scripts/
│   ├── build.sh           # Build script
│   └── test.sh            # Test script
├── go.mod
├── go.sum
├── Makefile
├── README.md
├── PRD.md
└── IMPLEMENTATION_PLAN.md
```

## Database Schema

### Core Tables

```sql
-- Tasks table
CREATE TABLE tasks (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    state TEXT NOT NULL DEFAULT 'ready_for_plan',
    priority INTEGER NOT NULL DEFAULT 5,
    owner TEXT,
    tags TEXT, -- JSON array
    dependencies TEXT, -- JSON array of task IDs
    blocked_by TEXT, -- JSON array of task IDs
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Requirements table
CREATE TABLE requirements (
    id TEXT PRIMARY KEY,
    key TEXT UNIQUE NOT NULL, -- e.g., "FR-P1"
    title TEXT NOT NULL,
    text TEXT NOT NULL,
    type TEXT NOT NULL, -- functional|nonfunctional|constraint|risk
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Task-Requirement links
CREATE TABLE task_requirements (
    task_id TEXT NOT NULL,
    requirement_id TEXT NOT NULL,
    PRIMARY KEY (task_id, requirement_id),
    FOREIGN KEY (task_id) REFERENCES tasks(id),
    FOREIGN KEY (requirement_id) REFERENCES requirements(id)
);

-- Artifacts table
CREATE TABLE artifacts (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    name TEXT NOT NULL, -- implementation_plan, change_summary, etc.
    version INTEGER NOT NULL DEFAULT 1,
    content TEXT NOT NULL,
    meta TEXT, -- JSON metadata
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id),
    UNIQUE(task_id, name, version)
);

-- Agents table
CREATE TABLE agents (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    role TEXT NOT NULL,
    description TEXT,
    routing_policy TEXT, -- JSON configuration
    permissions TEXT, -- JSON permissions
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Audit logs table
CREATE TABLE audit_logs (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    cycle_id TEXT NOT NULL,
    prev_state TEXT,
    next_state TEXT,
    actor TEXT, -- agent ID
    selection_reason TEXT,
    inputs_summary TEXT, -- JSON: plan hash, requirements, artifacts used
    outputs_summary TEXT, -- JSON: handovers created/updated
    commands TEXT, -- JSON array of external commands executed
    result TEXT,
    note TEXT,
    follow_ups TEXT, -- JSON array of follow-up interactions
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id)
);
```

## Configuration System

### Default Configuration (`configs/default.yaml`)

```yaml
# Core settings
plan_file: "./plan.md"
workspace: "./"
database: "./baton.db"
mcp_port: 8080

# LLM CLI settings
llm:
  primary: "claude"
  fallback: null
  timeout_seconds: 300
  max_retries: 1

  # Claude Code configuration
  claude:
    command: "claude"
    headless_args: ["-p"]
    output_format: "stream-json"
    mcp_connect: true

  # Future: OpenAI CLI support
  openai:
    command: "openai"
    headless_args: ["--non-interactive"]

# Agent configuration
agents:
  architect:
    name: "System Architect"
    role: "Plans and designs system architecture"
    allowed_states: ["ready_for_plan", "planning"]
    routing_policy:
      llm_preference: "claude"
      prompt_template: "architect.md"
    permissions:
      can_read_plan: true
      can_update_artifacts: true
      can_transition_to: ["planning", "ready_for_implementation"]

  developer:
    name: "Developer"
    role: "Implements code and fixes issues"
    allowed_states: ["ready_for_implementation", "implementing", "fixing"]
    routing_policy:
      llm_preference: "claude"
      prompt_template: "developer.md"
    permissions:
      can_read_plan: true
      can_execute_commands: true
      can_update_artifacts: true
      can_transition_to: ["implementing", "ready_for_code_review", "needs_fixes"]

  reviewer:
    name: "Code Reviewer"
    role: "Reviews code and provides feedback"
    allowed_states: ["reviewing"]
    routing_policy:
      llm_preference: "claude"
      prompt_template: "reviewer.md"
    permissions:
      can_read_artifacts: true
      can_update_artifacts: true
      can_transition_to: ["ready_for_commit", "needs_fixes"]

# Task selection policy
selection:
  algorithm: "priority_dependency"
  priority_weight: 1.0
  dependency_strict: true
  prefer_leaf_tasks: true
  tie_breaker: "oldest_updated"

# Completion handshake settings
completion:
  max_retries: 2
  retry_delay_seconds: 5
  timeout_seconds: 600
  require_explicit_state_update: true
  follow_up_template: "Are you finished? The state is not updated. Please either update the task state or provide a structured outcome with reason and next state."

# Security and safety settings
security:
  allowed_commands:
    - "git"
    - "npm"
    - "go"
    - "python"
    - "pytest"
    - "cargo"
    - "make"
  workspace_restriction: true
  secret_patterns:
    - "sk-"
    - "pk-"
    - "token"
    - "password"
    - "secret"
  redact_in_logs: true

# Logging configuration
logging:
  level: "info"
  format: "json"
  file: "baton.log"
  audit_retention_days: 90

# Development settings
development:
  dry_run_default: false
  debug_mcp: false
  cycle_timebox_seconds: 3600 # 1 hour max per cycle
```

## Implementation Phases

### Phase 1: Core Foundation (Week 1-2)

**Deliverables:**
- [ ] Go project structure and basic CLI commands
- [ ] SQLite database with schema and migrations
- [ ] Configuration loading and validation
- [ ] Basic Task CRUD operations

**Key Files:**
```go
// internal/storage/sqlite.go
type Store struct {
    db *sql.DB
}

func (s *Store) CreateTask(task *Task) error
func (s *Store) GetTask(id string) (*Task, error)
func (s *Store) UpdateTaskState(id, state, note string) error
func (s *Store) ListTasks(filters TaskFilters) ([]*Task, error)

// internal/config/config.go
type Config struct {
    PlanFile  string `yaml:"plan_file"`
    Database  string `yaml:"database"`
    LLM       LLMConfig `yaml:"llm"`
    Agents    map[string]Agent `yaml:"agents"`
    // ... other fields
}

// cmd/init.go - Initialize workspace
func runInit(cmd *cobra.Command, args []string) error {
    // Create default config
    // Initialize database
    // Create sample plan file
}
```

**Testing:**
- Unit tests for database operations
- Config loading validation tests
- CLI command parsing tests

### Phase 2: State Machine & Task Selection (Week 2-3)

**Deliverables:**
- [ ] State machine with transition validation
- [ ] Task selection algorithm implementation
- [ ] Dependency resolution logic
- [ ] Basic audit logging

**Key Files:**
```go
// internal/statemachine/states.go
var ValidTransitions = map[State][]State{
    ReadyForPlan:    {Planning},
    Planning:        {ReadyForImplementation, NeedsFixes},
    // ... complete transition map
}

func ValidateTransition(from, to State) error

// internal/statemachine/selection.go
type TaskSelector struct {
    store  *storage.Store
    config *config.SelectionConfig
}

func (ts *TaskSelector) SelectNext() (*Task, string, error) {
    // Implement priority + dependency + leaf preference
    // Return task, selection reason, error
}

// internal/audit/logger.go
type AuditLogger struct {
    store *storage.Store
}

func (al *AuditLogger) LogCycle(entry *CycleAudit) error
```

**Testing:**
- State transition validation tests
- Task selection algorithm tests
- Dependency resolution edge cases

### Phase 3: MCP Server Implementation (Week 3-4)

**Deliverables:**
- [ ] JSON-RPC 2.0 MCP server
- [ ] Task management endpoints
- [ ] Requirement and artifact endpoints
- [ ] Permission enforcement

**Key Files:**
```go
// internal/mcp/server.go
type Server struct {
    store  *storage.Store
    config *config.Config
    port   int
}

func (s *Server) Start() error
func (s *Server) Stop() error

// internal/mcp/handlers.go
type TaskHandler struct {
    store *storage.Store
}

func (h *TaskHandler) GetNext(params json.RawMessage) (*Task, error)
func (h *TaskHandler) UpdateState(params json.RawMessage) error
func (h *TaskHandler) AppendNote(params json.RawMessage) error
func (h *TaskHandler) List(params json.RawMessage) ([]*Task, error)

type ArtifactHandler struct {
    store *storage.Store
}

func (h *ArtifactHandler) Upsert(params json.RawMessage) error
func (h *ArtifactHandler) Get(params json.RawMessage) (*Artifact, error)
func (h *ArtifactHandler) List(params json.RawMessage) ([]*Artifact, error)

// MCP API endpoints
// tasks.get_next -> TaskHandler.GetNext
// tasks.get -> TaskHandler.Get
// tasks.update_state -> TaskHandler.UpdateState
// tasks.append_note -> TaskHandler.AppendNote
// tasks.list -> TaskHandler.List
// artifacts.upsert -> ArtifactHandler.Upsert
// artifacts.get -> ArtifactHandler.Get
// requirements.list -> RequirementHandler.List
// plan.read -> PlanHandler.Read
```

**Testing:**
- MCP protocol compliance tests
- Endpoint integration tests
- Permission enforcement tests

### Phase 4: LLM Integration & Claude Code (Week 4-5)

**Deliverables:**
- [ ] Claude Code headless mode integration
- [ ] JSON output parsing and validation
- [ ] Completion handshake implementation
- [ ] Error handling and retry logic

**Key Files:**
```go
// internal/llm/claude.go
type ClaudeClient struct {
    config    *config.LLMConfig
    mcpPort   int
}

func (c *ClaudeClient) Execute(prompt, agentID string) (*LLMResponse, error) {
    // Build command: claude -p "prompt" --output-format stream-json
    // Add MCP server connection if configured
    // Parse JSON output
    // Handle streaming responses
}

type LLMResponse struct {
    Success    bool             `json:"success"`
    Content    string           `json:"content"`
    Cost       float64          `json:"total_cost_usd"`
    Duration   int              `json:"duration_ms"`
    SessionID  string           `json:"session_id"`
    Metadata   json.RawMessage  `json:"metadata"`
}

// internal/cycle/handshake.go
type CompletionHandshake struct {
    maxRetries int
    timeout    time.Duration
    store      *storage.Store
}

func (ch *CompletionHandshake) Enforce(taskID string, response *LLMResponse) error {
    // Check if state was updated via MCP
    // If not, send follow-up prompt
    // Retry with bounded attempts
    // Set needs_fixes if unresolved
}
```

**Testing:**
- Claude Code integration tests (with mocked responses)
- JSON parsing edge case tests
- Completion handshake scenarios
- Retry logic validation

### Phase 5: Cycle Engine & Plan Integration (Week 5-6)

**Deliverables:**
- [ ] Complete cycle execution engine
- [ ] Context reset and rehydration logic
- [ ] Plan file parsing and requirement extraction
- [ ] Handover artifact management
- [ ] End-to-end cycle execution

**Key Files:**
```go
// internal/cycle/engine.go
type CycleEngine struct {
    store     *storage.Store
    config    *config.Config
    mcpServer *mcp.Server
    llmClient llm.Client
    selector  *statemachine.TaskSelector
    auditor   *audit.Logger
}

func (ce *CycleEngine) ExecuteCycle(dryRun bool) (*CycleResult, error) {
    // 1. Context reset (clear conversational memory)
    // 2. Rehydrate context from stored sources
    // 3. Select next task
    // 4. Start MCP server
    // 5. Execute agent logic via LLM
    // 6. Enforce completion handshake
    // 7. Create/update handover artifacts
    // 8. Record audit entry
    // 9. Stop MCP server
    // 10. Return cycle result
}

// internal/plan/parser.go
type PlanParser struct{}

func (pp *PlanParser) Parse(filepath string) (*Plan, []*Requirement, error) {
    // Parse markdown file
    // Extract requirements sections (FR-*, NFR-*, etc.)
    // Return structured plan and requirements
}

// internal/cycle/handover.go
type HandoverManager struct {
    store *storage.Store
}

func (hm *HandoverManager) ValidateRequired(taskID string, fromState, toState State) error
func (hm *HandoverManager) CreateArtifact(taskID, name, content string) error

// Required handover validation by state transition
var RequiredHandovers = map[string]map[string]string{
    "planning->ready_for_implementation": "implementation_plan",
    "implementing->ready_for_code_review": "change_summary",
    // ... complete mapping
}
```

**Testing:**
- Full cycle execution tests
- Context isolation validation
- Plan parsing accuracy tests
- Handover artifact requirement tests
- Integration tests with real plan files

## MCP API Specification

### Namespace: `baton`

#### Task Operations

**`baton.tasks.get_next`**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "baton.tasks.get_next",
  "params": {},
  "id": 1
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "task": {
      "id": "task-123",
      "title": "Implement authentication",
      "state": "ready_for_implementation",
      "priority": 8,
      "artifacts": [...],
      "requirements": [...]
    },
    "selection_reason": "Highest priority (8) with satisfied dependencies"
  },
  "id": 1
}
```

**`baton.tasks.update_state`**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "baton.tasks.update_state",
  "params": {
    "task_id": "task-123",
    "state": "implementing",
    "note": "Started implementation phase"
  },
  "id": 2
}
```

#### Artifact Operations

**`baton.artifacts.upsert`**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "baton.artifacts.upsert",
  "params": {
    "task_id": "task-123",
    "name": "implementation_plan",
    "content": "# Implementation Plan\n...",
    "meta": {"format": "markdown"}
  },
  "id": 3
}
```

#### Requirements Operations

**`baton.requirements.list`**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "baton.requirements.list",
  "params": {
    "type": "functional",
    "linked_task": "task-123"
  },
  "id": 4
}
```

#### Plan Operations

**`baton.plan.read`**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "baton.plan.read",
  "params": {},
  "id": 5
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "content": "# Project Plan\n...",
    "hash": "abc123...",
    "modified_at": "2024-03-15T10:30:00Z"
  },
  "id": 5
}
```

## Agent Prompt Templates

### Architect Agent (`templates/architect.md`)
```markdown
# System Architect Role

You are the system architect for this project. Your role is to analyze requirements and create comprehensive implementation plans.

## Current Context
- **Task**: {{.Task.Title}}
- **Description**: {{.Task.Description}}
- **Requirements**: {{range .Requirements}}
  - {{.Key}}: {{.Title}}
{{end}}

## Your Responsibilities
1. Create detailed implementation plans
2. Define acceptance criteria
3. Identify dependencies and constraints
4. Plan testing approach
5. Assess risks and mitigations

## Expected Output
You must create an "implementation_plan" artifact with:
- Goal & scope
- Acceptance criteria
- Step-by-step implementation outline
- Test plan
- File/module touch list
- Risks & mitigations

## Important Rules
- Use the MCP tools to update task state and artifacts
- State must be updated to "ready_for_implementation" when done
- Implementation plan must be complete before proceeding
```

### Developer Agent (`templates/developer.md`)
```markdown
# Developer Role

You are a software developer implementing planned features and fixes.

## Current Context
- **Task**: {{.Task.Title}}
- **State**: {{.Task.State}}
- **Implementation Plan**: {{.Artifacts.implementation_plan}}

## Your Responsibilities
1. Implement code according to the plan
2. Write and run tests
3. Create change summary documentation
4. Ensure quality standards are met

## Expected Output
You must create a "change_summary" artifact with:
- What changed & why
- Diff summary (human-readable)
- Commands executed & results
- Known limitations/TODOs

## Important Rules
- Follow the implementation plan exactly
- Run tests and linting before completion
- Use MCP tools to update task state and artifacts
- State must be updated to "ready_for_code_review" when done
```

## Testing Strategy

### Unit Tests
```go
// internal/storage/sqlite_test.go
func TestCreateTask(t *testing.T)
func TestUpdateTaskState(t *testing.T)
func TestTaskSelection(t *testing.T)

// internal/statemachine/states_test.go
func TestValidateTransition(t *testing.T)
func TestInvalidTransitions(t *testing.T)

// internal/mcp/handlers_test.go
func TestTaskGetNext(t *testing.T)
func TestArtifactUpsert(t *testing.T)
```

### Integration Tests
```go
// test/integration/cycle_test.go
func TestFullCycleExecution(t *testing.T) {
    // Setup test database and config
    // Create test task and requirements
    // Execute full cycle with mocked LLM
    // Verify state transitions and artifacts
}

func TestMCPServerIntegration(t *testing.T) {
    // Start MCP server
    // Make JSON-RPC calls
    // Verify responses and side effects
}
```

### End-to-End Tests
```go
// test/e2e/baton_test.go
func TestPlanToImplementation(t *testing.T) {
    // Real plan file ingestion
    // Real task creation
    // Mocked Claude Code responses
    // Full cycle execution
    // Artifact verification
}
```

## Build & Distribution

### Makefile
```makefile
.PHONY: build test clean install

BINARY_NAME=baton
VERSION=$(shell git describe --tags --always --dirty)

build:
	go build -ldflags="-X pkg/version.Version=$(VERSION)" -o $(BINARY_NAME) ./cmd

test:
	go test -v ./...

integration-test:
	go test -v -tags=integration ./test/integration

e2e-test:
	go test -v -tags=e2e ./test/e2e

clean:
	go clean
	rm -f $(BINARY_NAME)

install: build
	cp $(BINARY_NAME) $(GOPATH)/bin/

release:
	goreleaser release --rm-dist

docker-build:
	docker build -t baton:$(VERSION) .
```

### Cross-Platform Builds
```bash
#!/bin/bash
# scripts/build.sh

PLATFORMS="darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64"
VERSION=$(git describe --tags --always --dirty)

for PLATFORM in $PLATFORMS; do
  GOOS=${PLATFORM%/*}
  GOARCH=${PLATFORM#*/}
  OUTPUT_NAME="baton-${VERSION}-${GOOS}-${GOARCH}"

  if [ $GOOS = "windows" ]; then
    OUTPUT_NAME+='.exe'
  fi

  echo "Building $OUTPUT_NAME..."
  env GOOS=$GOOS GOARCH=$GOARCH go build \
    -ldflags="-X pkg/version.Version=$VERSION" \
    -o dist/$OUTPUT_NAME ./cmd
done
```

## Security Considerations

1. **Command Execution**: Whitelist allowed commands in config
2. **File System Access**: Restrict to workspace directory
3. **Secret Redaction**: Pattern-based secret detection and redaction
4. **MCP Server**: Bind to localhost only, implement authentication if needed
5. **Database**: Use parameterized queries to prevent SQL injection
6. **Logging**: Ensure no secrets are logged in audit trails

## Performance Targets

- **Cycle Execution**: < 30 seconds overhead (excluding LLM time)
- **Task Selection**: < 1 second for up to 1000 tasks
- **Plan Ingestion**: < 5 seconds for 100KB plan files
- **Database Operations**: < 100ms for typical queries
- **MCP Server Response**: < 50ms for data retrieval operations

## Deployment & Distribution

### Single Binary
- Go cross-compilation for all major platforms
- Embed default configuration and SQL schema
- No external dependencies beyond Claude Code CLI

### Installation Methods
```bash
# Direct download
curl -L https://github.com/user/baton/releases/latest/download/baton-linux-amd64 -o baton

# Package managers (future)
brew install baton
apt install baton
```

### Docker Support (Optional)
```dockerfile
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY baton /usr/local/bin/
ENTRYPOINT ["baton"]
```

## Development Milestones

### Week 1-2: Foundation
- [ ] Project structure and build system
- [ ] Database schema and basic operations
- [ ] Configuration system
- [ ] CLI commands structure

### Week 3-4: Core Logic
- [ ] State machine implementation
- [ ] Task selection algorithm
- [ ] MCP server and API endpoints
- [ ] Basic audit logging

### Week 5-6: Integration
- [ ] Claude Code integration
- [ ] Completion handshake logic
- [ ] Plan parsing and requirements
- [ ] Full cycle execution engine

### Week 7: Testing & Polish
- [ ] Comprehensive test coverage
- [ ] Error handling and edge cases
- [ ] Documentation and examples
- [ ] Performance optimization

### Week 8: Release Preparation
- [ ] Cross-platform builds
- [ ] Release automation
- [ ] Getting started documentation
- [ ] Beta testing with real workflows

## Success Criteria

✅ **MVP Complete**: All acceptance criteria from PRD implemented and tested

✅ **Performance**: Meets or exceeds performance targets

✅ **Reliability**: Handles errors gracefully, maintains audit integrity

✅ **Usability**: Clear CLI help, actionable error messages

✅ **Integration**: Seamless Claude Code headless mode integration

✅ **Extensibility**: Configuration-driven customization working

This implementation plan provides a comprehensive roadmap for building Baton with the recommended Go stack, focusing on reliability, configurability, and seamless Claude Code integration.