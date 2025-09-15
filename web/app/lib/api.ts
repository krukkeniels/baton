import { Task, TaskState, Status, AuditEntry, CreateTaskRequest, UpdateTaskRequest } from '../types'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3001/api'

class ApiClient {
  private baseUrl: string

  constructor(baseUrl: string = API_BASE_URL) {
    this.baseUrl = baseUrl
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`

    const config: RequestInit = {
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      ...options,
    }

    const response = await fetch(url, config)

    if (!response.ok) {
      throw new Error(`API request failed: ${response.status} ${response.statusText}`)
    }

    return response.json()
  }

  // Task operations
  async getTasks(filters?: { state?: TaskState; priority?: number }): Promise<Task[]> {
    const params = new URLSearchParams()

    if (filters?.state) {
      params.append('state', filters.state)
    }
    if (filters?.priority) {
      params.append('priority', filters.priority.toString())
    }

    const query = params.toString()
    const endpoint = query ? `/tasks?${query}` : '/tasks'

    return this.request<Task[]>(endpoint)
  }

  async getTask(id: string): Promise<Task> {
    return this.request<Task>(`/tasks/${id}`)
  }

  async updateTaskState(
    id: string,
    state: TaskState,
    note?: string
  ): Promise<Task> {
    return this.request<Task>(`/tasks/${id}`, {
      method: 'PUT',
      body: JSON.stringify({ state, note }),
    })
  }

  async createTask(request: CreateTaskRequest): Promise<Task> {
    return this.request<Task>('/tasks/create', {
      method: 'POST',
      body: JSON.stringify(request),
    })
  }

  async updateTask(request: UpdateTaskRequest): Promise<Task> {
    return this.request<Task>('/tasks/update', {
      method: 'POST',
      body: JSON.stringify(request),
    })
  }

  // Status and monitoring
  async getStatus(): Promise<Status> {
    return this.request<Status>('/status')
  }

  async getAuditHistory(taskId: string): Promise<AuditEntry[]> {
    return this.request<AuditEntry[]>(`/audit/${taskId}`)
  }
}

export const apiClient = new ApiClient()