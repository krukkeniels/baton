# Baton v1.0.0 - Stable Release

## 🎉 First Stable Release

We're excited to announce the first stable release of Baton - an intelligent task orchestrator that bridges human planning with AI execution through Claude Code CLI.

## ✨ Major Features

### Core Orchestration Engine
- **Intelligent Task Selection**: Multiple algorithms for task prioritization (FIFO, priority-based, dependency-aware)
- **State Machine Management**: Robust task lifecycle with 10 distinct states and automatic transitions
- **Dependency Resolution**: Automatic handling of task dependencies and blockers
- **Audit Trail**: Complete history of all cycles, decisions, and outcomes

### Web UI & Visualization
- **Modern React Frontend**: Interactive Kanban board for visual task management
- **Real-time Updates**: WebSocket support for live task status changes
- **Drag-and-Drop**: Intuitive task state transitions
- **Task Details**: Rich task information with artifacts and cycle history

### Developer Experience
- **Interactive Wizard**: Guided setup for new projects
- **Claude Code Integration**: Seamless integration with Claude Code CLI via MCP servers
- **REST API**: Full programmatic access to all Baton functionality
- **Flexible Configuration**: YAML-based configuration with environment overrides

### Safety & Reliability
- **Validation Gates**: Pre-execution checks and post-execution verification
- **Dry Run Mode**: Preview changes before execution
- **Error Recovery**: Automatic state recovery and retry mechanisms
- **Security Features**: Workspace restrictions and command whitelisting

## 📦 What's Included

- Complete CLI with commands for task management, cycle execution, and monitoring
- Web server with REST API and WebSocket support
- React-based web UI with responsive design
- SQLite database for persistent storage
- Comprehensive documentation and examples

## 🚀 Getting Started

### Installation
```bash
# Download the latest release
wget https://github.com/krukkeniels/baton/releases/download/v1.0.0/baton-linux-amd64
chmod +x baton-linux-amd64
sudo mv baton-linux-amd64 /usr/local/bin/baton

# Or build from source
git clone https://github.com/krukkeniels/baton.git
cd baton
make build
```

### Quick Start
```bash
# Initialize a new project
baton init

# Start the web UI
baton web

# Open http://localhost:3000 in your browser
```

## 📊 Requirements

- Go 1.21+ (for building from source)
- Node.js 18+ (for web UI development)
- Claude Code CLI (for AI execution)
- SQLite3

## 🔄 State Transitions

Baton manages tasks through a comprehensive state machine:

- `pending` → `ready_for_plan` → `planning` → `implementing`
- `implementing` → `ready_for_code_review` → `reviewing`
- `reviewing` → `ready_to_commit` or `needs_fixes`
- `ready_to_commit` → `committing` → `committed`

## 🛠 Technical Stack

- **Backend**: Go with Cobra CLI framework
- **Frontend**: Next.js, React, TypeScript, Tailwind CSS
- **Database**: SQLite with GORM
- **Communication**: REST API + WebSockets
- **AI Integration**: Claude Code CLI via MCP protocol

## 📈 Performance

- Handles 1000+ concurrent tasks
- Sub-second task selection
- Real-time UI updates via WebSocket
- Efficient SQLite queries with indexing

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Claude Code CLI team for the excellent AI execution platform
- The Go community for amazing tools and libraries
- All early adopters and testers who provided valuable feedback

## 📮 Support

- GitHub Issues: https://github.com/krukkeniels/baton/issues
- Documentation: https://github.com/krukkeniels/baton/wiki

---

**Full Changelog**: https://github.com/krukkeniels/baton/commits/v1.0.0