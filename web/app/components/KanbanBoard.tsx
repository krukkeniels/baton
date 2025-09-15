'use client'

import { useState, useEffect } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { DragDropContext, Droppable, Draggable, DropResult } from 'react-beautiful-dnd'
import { motion, AnimatePresence } from 'framer-motion'
import { Plus, RefreshCw, AlertCircle, Wifi, WifiOff } from 'lucide-react'

import { Task, TaskState, STATE_CONFIG } from '../types'
import { apiClient } from '../lib/api'
import { useWebSocket } from '../hooks/useWebSocket'
import { TaskCard } from './TaskCard'
import { CreateTaskDialog } from './CreateTaskDialog'

const COLUMN_ORDER: TaskState[] = [
  'ready_for_plan',
  'planning',
  'ready_for_implementation',
  'implementing',
  'ready_for_code_review',
  'reviewing',
  'ready_for_commit',
  'committing',
  'needs_fixes',
  'fixing',
  'DONE'
]

export function KanbanBoard() {
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false)
  const queryClient = useQueryClient()
  const { isConnected, lastMessage } = useWebSocket()

  const { data: tasks = [], isLoading, error, refetch } = useQuery({
    queryKey: ['tasks'],
    queryFn: () => apiClient.getTasks(),
    refetchInterval: 30000, // Refetch every 30 seconds as fallback
  })

  // Handle real-time updates via WebSocket
  useEffect(() => {
    if (lastMessage) {
      switch (lastMessage.type) {
        case 'task_created':
        case 'task_updated':
        case 'task_deleted':
          // Invalidate and refetch tasks
          queryClient.invalidateQueries({ queryKey: ['tasks'] })
          break
        case 'status_update':
          // Could update a status indicator here
          break
      }
    }
  }, [lastMessage, queryClient])

  // Group tasks by state
  const tasksByState = tasks.reduce((acc, task) => {
    if (!acc[task.state]) {
      acc[task.state] = []
    }
    acc[task.state].push(task)
    return acc
  }, {} as Record<TaskState, Task[]>)

  const handleDragEnd = async (result: DropResult) => {
    const { destination, source, draggableId } = result

    // No destination or same position
    if (!destination ||
        (destination.droppableId === source.droppableId &&
         destination.index === source.index)) {
      return
    }

    const newState = destination.droppableId as TaskState
    const taskId = draggableId

    try {
      await apiClient.updateTaskState(taskId, newState, 'Moved via kanban drag & drop')
      // The WebSocket will handle the real-time update
    } catch (error) {
      console.error('Failed to update task state:', error)
      // Could show a toast notification here
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="flex items-center space-x-2 text-muted-foreground">
          <RefreshCw className="w-6 h-6 animate-spin" />
          <span>Loading tasks...</span>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="flex items-center space-x-2 text-destructive">
          <AlertCircle className="w-6 h-6" />
          <span>Failed to load tasks. Please try again.</span>
          <button
            onClick={() => refetch()}
            className="btn-tech-secondary"
          >
            Retry
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b border-border">
        <div className="flex items-center space-x-4">
          <h1 className="text-2xl font-bold font-mono">Baton Tasks</h1>
          <div className="flex items-center space-x-2">
            {isConnected ? (
              <div className="flex items-center space-x-1 text-green-400">
                <Wifi className="w-4 h-4" />
                <span className="text-sm">Live</span>
              </div>
            ) : (
              <div className="flex items-center space-x-1 text-yellow-400">
                <WifiOff className="w-4 h-4" />
                <span className="text-sm">Reconnecting...</span>
              </div>
            )}
          </div>
        </div>
        <div className="flex items-center space-x-2">
          <button
            onClick={() => refetch()}
            className="btn-tech-ghost"
            disabled={isLoading}
          >
            <RefreshCw className={`w-4 h-4 mr-2 ${isLoading ? 'animate-spin' : ''}`} />
            Refresh
          </button>
          <button
            onClick={() => setIsCreateDialogOpen(true)}
            className="btn-tech-primary"
          >
            <Plus className="w-4 h-4 mr-2" />
            New Task
          </button>
        </div>
      </div>

      {/* Kanban Board */}
      <div className="flex-1 overflow-x-auto">
        <DragDropContext onDragEnd={handleDragEnd}>
          <div className="flex space-x-4 p-4 min-w-max">
            {COLUMN_ORDER.map((state) => {
              const stateTasks = tasksByState[state] || []
              const config = STATE_CONFIG[state]

              return (
                <div key={state} className="flex-shrink-0 w-80">
                  <div className="bg-card border border-border rounded-lg">
                    {/* Column Header */}
                    <div className="p-4 border-b border-border">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center space-x-2">
                          <span className="text-lg">{config.icon}</span>
                          <h3 className="font-semibold text-sm font-mono">
                            {config.label}
                          </h3>
                        </div>
                        <span className="bg-muted text-muted-foreground px-2 py-1 rounded-full text-xs">
                          {stateTasks.length}
                        </span>
                      </div>
                      <p className="text-xs text-muted-foreground mt-1">
                        {config.description}
                      </p>
                    </div>

                    {/* Column Content */}
                    <Droppable droppableId={state}>
                      {(provided, snapshot) => (
                        <div
                          ref={provided.innerRef}
                          {...provided.droppableProps}
                          className={`p-2 min-h-[200px] transition-colors duration-200 ${
                            snapshot.isDraggingOver ? 'bg-accent/20' : ''
                          }`}
                        >
                          <AnimatePresence>
                            {stateTasks.map((task, index) => (
                              <Draggable
                                key={task.id}
                                draggableId={task.id}
                                index={index}
                              >
                                {(provided, snapshot) => (
                                  <div
                                    ref={provided.innerRef}
                                    {...provided.draggableProps}
                                    {...provided.dragHandleProps}
                                    className={`mb-2 ${
                                      snapshot.isDragging ? 'opacity-75 rotate-2' : ''
                                    }`}
                                  >
                                    <motion.div
                                      initial={{ opacity: 0, y: 20 }}
                                      animate={{ opacity: 1, y: 0 }}
                                      exit={{ opacity: 0, y: -20 }}
                                      transition={{ duration: 0.2 }}
                                    >
                                      <TaskCard
                                        task={task}
                                        isDragging={snapshot.isDragging}
                                      />
                                    </motion.div>
                                  </div>
                                )}
                              </Draggable>
                            ))}
                          </AnimatePresence>
                          {provided.placeholder}
                        </div>
                      )}
                    </Droppable>
                  </div>
                </div>
              )
            })}
          </div>
        </DragDropContext>
      </div>

      {/* Create Task Dialog */}
      <CreateTaskDialog
        open={isCreateDialogOpen}
        onOpenChange={setIsCreateDialogOpen}
      />
    </div>
  )
}