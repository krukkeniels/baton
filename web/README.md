# Baton Web UI

Modern, responsive web interface for Baton CLI Orchestrator with real-time task management and LLM-powered interactions.

## Features

### 🎯 Modern Kanban Board
- **Visual Task Flow**: Tasks move through states from planning to completion
- **Real-time Updates**: WebSocket integration for live task updates
- **Drag & Drop**: Intuitive task state transitions
- **State Indicators**: Color-coded columns with progress indicators

### 🤖 LLM-Powered Task Management
- **Smart Task Creation**: Natural language prompts converted to structured tasks
- **Intelligent Updates**: Describe changes in plain English
- **Auto-categorization**: AI extracts tags, priorities, and metadata
- **Validation**: Smart state transition validation

### 📊 Rich Task Details
- **Complete History**: Full audit trail with timestamps
- **Artifact Viewer**: Expandable implementation plans and documentation
- **Dependencies**: Visual dependency tracking
- **Metadata**: Priority, owner, tags, and timeline information

### 🎨 Technical Design
- **Dark Theme**: Sleek, developer-friendly interface
- **Responsive**: Works on desktop, tablet, and mobile
- **Fast**: Optimized for performance with intelligent caching
- **Accessible**: Keyboard navigation and screen reader support

## Technology Stack

- **Frontend**: Next.js 14, React 18, TypeScript
- **Styling**: Tailwind CSS with custom design system
- **State Management**: TanStack Query + Zustand
- **Animations**: Framer Motion
- **Real-time**: WebSocket integration
- **Build**: Static export for embedding in Go binary

## Development

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Type checking
npm run type-check

# Lint code
npm run lint
```

## Architecture

### Component Structure
```
app/
├── components/
│   ├── KanbanBoard.tsx      # Main board view
│   ├── TaskCard.tsx         # Individual task cards
│   ├── TaskDetailDialog.tsx # Task detail modal
│   ├── CreateTaskDialog.tsx # LLM-powered task creation
│   └── UpdateTaskDialog.tsx # LLM-powered task updates
├── hooks/
│   └── useWebSocket.ts      # Real-time connection
├── lib/
│   └── api.ts              # Backend API client
├── types/
│   └── index.ts            # TypeScript definitions
└── providers/
    └── react-query-provider.tsx
```

### State Management
- **Server State**: TanStack Query for caching and synchronization
- **WebSocket**: Real-time updates from backend
- **Local State**: React hooks for UI interactions

### API Integration
- **REST API**: CRUD operations for tasks and metadata
- **WebSocket**: Real-time notifications
- **LLM Integration**: Natural language task operations

## Configuration

Environment variables:
- `NEXT_PUBLIC_API_URL`: Backend API URL (default: http://localhost:3001/api)
- `NEXT_PUBLIC_WS_URL`: WebSocket URL (default: ws://localhost:3001/api/ws)

## Usage

### Starting the Web UI
```bash
# Build and start backend + frontend
make dev-web

# Or separately:
baton web --port 3001
# Then in another terminal:
cd web && npm run dev
```

### Task Management Flow
1. **Create Task**: Click "New Task" and describe what you want to accomplish
2. **AI Processing**: LLM analyzes your prompt and creates structured task
3. **Kanban Board**: Task appears in appropriate column based on state
4. **Progress**: Drag tasks between columns or use detail view for updates
5. **History**: View complete audit trail and artifacts

### LLM Prompt Examples

**Creating Tasks:**
- "Create a user registration form with email validation"
- "Fix the authentication bug causing 500 errors on login"
- "Add unit tests for the payment processing module"

**Updating Tasks:**
- "Change priority to high - this is blocking other work"
- "Move to implementing state since planning is complete"
- "Add React and TypeScript tags to this task"

## Integration with Baton CLI

The web UI integrates seamlessly with the existing Baton architecture:

- **Shared Database**: Uses the same SQLite database as CLI
- **MCP Protocol**: Extends existing MCP server with web endpoints
- **State Machine**: Respects the same state transition rules
- **Audit Trail**: All changes are logged in the same audit system
- **Agents**: Web updates trigger the same agent workflows

## Performance

- **Bundle Size**: ~200KB gzipped (excluding dependencies)
- **Load Time**: <1s initial load with caching
- **Real-time**: <100ms WebSocket message latency
- **Responsive**: 60fps animations and transitions

## Browser Support

- Chrome/Edge 90+
- Firefox 88+
- Safari 14+
- Mobile browsers with modern JavaScript support