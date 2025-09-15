export interface Task {
  id: string
  title: string
  description: string
  state: TaskState
  priority: number
  owner: string
  tags: string[]
  dependencies: string[]
  created_at: string
  updated_at: string
  artifacts?: Artifact[]
}

export type TaskState =
  | 'ready_for_plan'
  | 'planning'
  | 'ready_for_implementation'
  | 'implementing'
  | 'ready_for_code_review'
  | 'reviewing'
  | 'ready_for_commit'
  | 'committing'
  | 'needs_fixes'
  | 'fixing'
  | 'DONE'

export interface Artifact {
  id: string
  task_id: string
  name: string
  version: number
  content: string
  metadata: Record<string, any>
  created_at: string
}

export interface AuditEntry {
  id: string
  task_id: string
  task_title?: string
  prev_state: string
  next_state: string
  actor: string
  reason: string
  note: string
  commands?: string[]
  follow_ups?: string[]
  inputs_summary: string
  outputs_summary: string
  created_at: string
}

export interface Status {
  tasks_by_state: Record<TaskState, number>
  total_tasks: number
  recent_activity: AuditEntry[]
}

export interface WSMessage {
  type: 'task_created' | 'task_updated' | 'task_deleted' | 'status_update'
  timestamp: number
  data: any
}

export interface CreateTaskRequest {
  prompt: string
  owner?: string
}

export interface UpdateTaskRequest {
  task_id: string
  prompt: string
}

// State configuration for UI display
export const STATE_CONFIG: Record<TaskState, {
  label: string
  color: string
  description: string
  icon: string
}> = {
  ready_for_plan: {
    label: 'Ready for Planning',
    color: 'state-ready-plan',
    description: 'Task is ready to be planned by an architect',
    icon: '📋'
  },
  planning: {
    label: 'Planning',
    color: 'state-planning',
    description: 'Architect is creating implementation plan',
    icon: '🎯'
  },
  ready_for_implementation: {
    label: 'Ready for Implementation',
    color: 'state-ready-impl',
    description: 'Task has a plan and is ready for development',
    icon: '🚀'
  },
  implementing: {
    label: 'Implementing',
    color: 'state-implementing',
    description: 'Developer is implementing the task',
    icon: '⚡'
  },
  ready_for_code_review: {
    label: 'Ready for Review',
    color: 'state-ready-review',
    description: 'Implementation is complete and awaiting review',
    icon: '👀'
  },
  reviewing: {
    label: 'Reviewing',
    color: 'state-reviewing',
    description: 'Code reviewer is examining the implementation',
    icon: '🔍'
  },
  ready_for_commit: {
    label: 'Ready for Commit',
    color: 'state-ready-commit',
    description: 'Review passed, ready to commit changes',
    icon: '✅'
  },
  committing: {
    label: 'Committing',
    color: 'state-committing',
    description: 'Changes are being committed to repository',
    icon: '💾'
  },
  needs_fixes: {
    label: 'Needs Fixes',
    color: 'state-needs-fixes',
    description: 'Issues found that require developer attention',
    icon: '🔧'
  },
  fixing: {
    label: 'Fixing',
    color: 'state-fixing',
    description: 'Developer is addressing review feedback',
    icon: '🛠️'
  },
  DONE: {
    label: 'Done',
    color: 'state-done',
    description: 'Task is complete and committed',
    icon: '🎉'
  }
}

export const PRIORITY_CONFIG = {
  1: { label: 'Lowest', color: 'badge-priority-low' },
  2: { label: 'Very Low', color: 'badge-priority-low' },
  3: { label: 'Low', color: 'badge-priority-low' },
  4: { label: 'Below Normal', color: 'badge-priority-low' },
  5: { label: 'Normal', color: 'badge-priority-medium' },
  6: { label: 'Above Normal', color: 'badge-priority-medium' },
  7: { label: 'High', color: 'badge-priority-high' },
  8: { label: 'Very High', color: 'badge-priority-high' },
  9: { label: 'Highest', color: 'badge-priority-high' },
  10: { label: 'Critical', color: 'badge-priority-high' },
}