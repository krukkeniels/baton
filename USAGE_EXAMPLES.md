# Baton Usage Examples

This document provides real-world examples of using Baton for different project types and workflows.

## Example 1: Web Application Development

### Initial Setup

```bash
# Initialize project
mkdir webapp-project && cd webapp-project
baton init

# Edit plan.md with requirements
```

**plan.md**:
```markdown
# Web Application Project

## Functional Requirements

**FR-1**: The system shall provide user registration functionality.
**FR-2**: The system shall allow users to create and manage tasks.
**FR-3**: The system shall provide a REST API for task operations.
**FR-4**: The system shall include a web-based user interface.

## Non-Functional Requirements

**NFR-1**: The system shall handle 1000 concurrent users.
**NFR-2**: The system shall respond to API calls within 200ms.

## Technical Constraints

**TC-1**: The backend must use Node.js and Express.
**TC-2**: The frontend must use React.
**TC-3**: Data must be stored in PostgreSQL.
```

### Workflow Execution

```bash
# Ingest requirements
baton ingest plan.md

# Check initial status
baton status
# Output: 0 tasks initially

# Create first task manually or import from plan
# Note: In a real implementation, tasks would be derived from requirements

# Check what would be selected
baton tasks next

# Execute planning cycle
baton start --dry-run  # Preview first
baton start           # Execute actual cycle

# Monitor progress
baton status
baton tasks list --state implementing
```

## Example 2: API Development Workflow

### Plan File
```markdown
# REST API Project

## Functional Requirements

**FR-API-1**: Implement user authentication endpoints.
**FR-API-2**: Create task CRUD operations API.
**FR-API-3**: Add API documentation with OpenAPI spec.
**FR-API-4**: Implement input validation middleware.

## Acceptance Criteria

**AC-1**: All endpoints return appropriate HTTP status codes.
**AC-2**: API responses follow consistent JSON schema.
**AC-3**: Authentication uses JWT tokens.
```

### Task Management

```bash
# After ingestion, manually create tasks
baton tasks update --id auth-task --state ready_for_plan --note "Starting auth implementation"

# Check dependencies
baton tasks list --state blocked

# Execute cycles systematically
while baton tasks next; do
    echo "Executing cycle..."
    baton start
    sleep 5  # Brief pause between cycles
done
```

## Example 3: Bug Fix Workflow

### Emergency Fix Process

```bash
# Check current state
baton status

# Create urgent fix task
baton tasks update --id bug-001 --state needs_fixes --note "Critical security vulnerability"

# Prioritize the fix
baton tasks update --id bug-001 --priority 10

# Execute fix cycle
baton start
```

### Post-Fix Verification

```bash
# Verify fix completion
baton tasks list --id bug-001

# Check audit trail
baton cycles show --task bug-001

# Ensure proper handover
baton artifacts list --task bug-001
```

## Example 4: Feature Development Pipeline

### Multi-State Feature Development

```bash
# Start with planning
baton tasks update --id feature-xyz --state ready_for_plan

# Execute planning cycle
baton start
# Expected: Creates implementation_plan artifact

# Move to implementation
baton start
# Expected: Transitions to implementing, then ready_for_code_review

# Code review cycle
baton start
# Expected: Creates review_findings artifact

# Based on review results, either commit or fix
baton start
# Expected: Either commits or transitions to needs_fixes
```

## Example 5: Configuration Examples

### Development Environment

**baton.yaml**:
```yaml
plan_file: "./development-plan.md"
workspace: "./dev-workspace"
database: "./dev-baton.db"

llm:
  primary: "claude"
  timeout_seconds: 600  # Longer timeout for complex tasks

selection:
  algorithm: "priority_dependency"
  prefer_leaf_tasks: false  # Work on blocking tasks first

development:
  dry_run_default: true    # Always preview changes
  debug_mcp: true          # Enable MCP debugging
  cycle_timebox_seconds: 1800  # 30-minute max per cycle

security:
  allowed_commands: ["git", "npm", "node", "pytest", "docker"]
```

### Production Environment

**production-baton.yaml**:
```yaml
plan_file: "./production-plan.md"
workspace: "/opt/project"
database: "/opt/project/baton.db"

llm:
  primary: "claude"
  timeout_seconds: 300

completion:
  max_retries: 3
  timeout_seconds: 900

security:
  workspace_restriction: true
  redact_in_logs: true
  allowed_commands: ["git", "make", "docker"]

logging:
  level: "info"
  format: "json"
  file: "/var/log/baton.log"
```

## Example 6: Integration with CI/CD

### GitHub Actions Integration

**.github/workflows/baton.yml**:
```yaml
name: Baton Automation
on:
  schedule:
    - cron: '0 */6 * * *'  # Every 6 hours
  workflow_dispatch:

jobs:
  run-cycle:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install Baton
        run: |
          wget https://github.com/user/baton/releases/latest/download/baton-linux-amd64
          chmod +x baton-linux-amd64
          sudo mv baton-linux-amd64 /usr/local/bin/baton

      - name: Install Claude Code CLI
        run: |
          # Install Claude Code CLI
          curl -o- https://claude.ai/install.sh | bash

      - name: Check Status
        run: |
          baton status --json > status.json
          cat status.json

      - name: Execute Cycle
        run: |
          baton start || echo "Cycle execution completed"

      - name: Upload Artifacts
        uses: actions/upload-artifact@v3
        with:
          name: baton-artifacts
          path: baton.log
```

## Example 7: Team Collaboration

### Multi-Developer Workflow

```bash
# Developer A: Work on backend
baton tasks list --owner backend-team
baton start

# Developer B: Work on frontend (parallel)
baton --workspace ./frontend-workspace start

# Team Lead: Monitor overall progress
baton status --json | jq '.completion_rate'

# Daily standup: Review recent activity
baton cycles show --since yesterday
```

### Code Review Integration

```bash
# After implementation, create PR
git checkout -b feature/task-123
baton start  # Implementation cycle
git add . && git commit -m "Implement feature as per baton cycle"

# Reviewer uses baton for systematic review
baton tasks update --id task-123 --state reviewing
baton start  # Review cycle creates findings

# Based on review, either approve or request changes
# Baton automatically transitions state based on review findings
```

## Example 8: Debugging and Troubleshooting

### Debug Mode Usage

```bash
# Enable verbose logging
baton start --verbose

# Check detailed status
baton status --json | jq '.blocked_tasks'

# Examine specific task issues
baton tasks next  # Shows selection reasoning

# Check MCP server connectivity
baton start --dry-run  # Preview without execution
```

### Recovery from Errors

```bash
# If cycle fails mid-execution
baton status
# Shows tasks in intermediate states

# Manual state correction if needed
baton tasks update --id stuck-task --state needs_fixes --note "Recovered from error"

# Re-run with dry-run first
baton start --dry-run
baton start
```

## Example 9: Metrics and Monitoring

### Progress Tracking

```bash
# Generate weekly report
echo "# Weekly Progress Report" > report.md
echo "Date: $(date)" >> report.md
baton status --json | jq -r '
  "## Summary",
  "- Total Tasks: \(.total_tasks)",
  "- Completion Rate: \(.completion_rate)%",
  "- Ready Tasks: \(.ready_tasks | length)",
  "- Blocked Tasks: \(.blocked_tasks | length)"
' >> report.md

# Track velocity (tasks completed per day)
baton cycles show --since "1 week ago" --format json | jq '.[] | select(.result == "success")' | wc -l
```

### Performance Analysis

```bash
# Analyze cycle execution times
baton cycles show --format json | jq '.[] | .duration_seconds' | awk '{sum+=$1} END {print "Average cycle time:", sum/NR, "seconds"}'

# Find bottleneck states
baton tasks list --json | jq -r '.[] | .state' | sort | uniq -c | sort -nr
```

## Best Practices from Examples

1. **Start Small**: Begin with simple, well-defined tasks
2. **Use Dry Runs**: Always preview changes before execution
3. **Monitor Dependencies**: Keep task relationships current
4. **Regular Status Checks**: Use `baton status` frequently
5. **Systematic Planning**: Let the architect agent create detailed plans
6. **Quality Gates**: Use the review cycle for all significant changes
7. **Audit Trail**: Leverage cycle history for retrospectives
8. **Configuration Management**: Use different configs for different environments
9. **Automation**: Integrate with CI/CD for continuous development
10. **Team Coordination**: Use workspace separation for parallel development

These examples demonstrate Baton's flexibility and power in managing complex development workflows while maintaining systematic progress and quality assurance.