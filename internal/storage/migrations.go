package storage

const CreateTablesSQL = `
-- Tasks table
CREATE TABLE IF NOT EXISTS tasks (
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
CREATE TABLE IF NOT EXISTS requirements (
    id TEXT PRIMARY KEY,
    key TEXT UNIQUE NOT NULL, -- e.g., "FR-P1"
    title TEXT NOT NULL,
    text TEXT NOT NULL,
    type TEXT NOT NULL, -- functional|nonfunctional|constraint|risk
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Task-Requirement links
CREATE TABLE IF NOT EXISTS task_requirements (
    task_id TEXT NOT NULL,
    requirement_id TEXT NOT NULL,
    PRIMARY KEY (task_id, requirement_id),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (requirement_id) REFERENCES requirements(id) ON DELETE CASCADE
);

-- Artifacts table
CREATE TABLE IF NOT EXISTS artifacts (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    name TEXT NOT NULL, -- implementation_plan, change_summary, etc.
    version INTEGER NOT NULL DEFAULT 1,
    content TEXT NOT NULL,
    meta TEXT, -- JSON metadata
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    UNIQUE(task_id, name, version)
);

-- Agents table
CREATE TABLE IF NOT EXISTS agents (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    role TEXT NOT NULL,
    description TEXT,
    routing_policy TEXT, -- JSON configuration
    permissions TEXT, -- JSON permissions
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Audit logs table
CREATE TABLE IF NOT EXISTS audit_logs (
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
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_tasks_state ON tasks(state);
CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority);
CREATE INDEX IF NOT EXISTS idx_tasks_updated_at ON tasks(updated_at);
CREATE INDEX IF NOT EXISTS idx_requirements_key ON requirements(key);
CREATE INDEX IF NOT EXISTS idx_requirements_type ON requirements(type);
CREATE INDEX IF NOT EXISTS idx_artifacts_task_id ON artifacts(task_id);
CREATE INDEX IF NOT EXISTS idx_artifacts_name ON artifacts(name);
CREATE INDEX IF NOT EXISTS idx_audit_logs_task_id ON audit_logs(task_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_cycle_id ON audit_logs(cycle_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

-- Triggers to update updated_at timestamps
CREATE TRIGGER IF NOT EXISTS update_tasks_updated_at
    AFTER UPDATE ON tasks
    FOR EACH ROW
    BEGIN
        UPDATE tasks SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_requirements_updated_at
    AFTER UPDATE ON requirements
    FOR EACH ROW
    BEGIN
        UPDATE requirements SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;
`