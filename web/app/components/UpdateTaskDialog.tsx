'use client'

import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { motion, AnimatePresence } from 'framer-motion'
import { X, Sparkles, Loader2, Edit3, CheckCircle, AlertCircle } from 'lucide-react'

import { Task } from '../types'
import { apiClient } from '../lib/api'

interface UpdateTaskDialogProps {
  task: Task
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function UpdateTaskDialog({ task, open, onOpenChange }: UpdateTaskDialogProps) {
  const [prompt, setPrompt] = useState('')
  const queryClient = useQueryClient()

  const updateTaskMutation = useMutation({
    mutationFn: (data: { task_id: string; prompt: string }) =>
      apiClient.updateTask(data),
    onSuccess: () => {
      // Invalidate tasks query to refetch data
      queryClient.invalidateQueries({ queryKey: ['tasks'] })
      queryClient.invalidateQueries({ queryKey: ['task', task.id] })

      // Reset form and close dialog
      setPrompt('')
      onOpenChange(false)
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!prompt.trim()) return

    updateTaskMutation.mutate({
      task_id: task.id,
      prompt: prompt.trim(),
    })
  }

  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <motion.div
        className="absolute inset-0 bg-black/80"
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        onClick={() => onOpenChange(false)}
      />

      {/* Dialog */}
      <motion.div
        className="relative bg-card border border-border rounded-lg shadow-2xl w-full max-w-2xl m-4"
        initial={{ opacity: 0, scale: 0.95, y: 20 }}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        exit={{ opacity: 0, scale: 0.95, y: 20 }}
        transition={{ duration: 0.2 }}
      >
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-border">
          <div className="flex items-center space-x-3">
            <div className="w-8 h-8 bg-gradient-to-br from-orange-500 to-red-600 rounded-lg flex items-center justify-center">
              <Edit3 className="w-4 h-4 text-white" />
            </div>
            <div>
              <h2 className="text-xl font-bold text-foreground">Update Task</h2>
              <p className="text-sm text-muted-foreground">
                Describe what changes you want to make
              </p>
            </div>
          </div>
          <button
            onClick={() => onOpenChange(false)}
            className="btn-tech-ghost p-2"
            disabled={updateTaskMutation.isPending}
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Content */}
        <form onSubmit={handleSubmit} className="p-6 space-y-6">
          {/* Current Task Info */}
          <div className="bg-muted/20 rounded-lg p-4 space-y-3">
            <h3 className="text-sm font-semibold text-foreground">Current Task:</h3>
            <div className="space-y-2">
              <div>
                <p className="text-sm font-medium text-foreground">{task.title}</p>
                <p className="text-xs text-muted-foreground">
                  State: {task.state} • Priority: {task.priority} • Owner: {task.owner}
                </p>
              </div>
              {task.description && (
                <div className="bg-background/50 rounded p-3">
                  <p className="text-xs text-muted-foreground whitespace-pre-wrap">
                    {task.description.slice(0, 200)}{task.description.length > 200 ? '...' : ''}
                  </p>
                </div>
              )}
            </div>
          </div>

          {/* Update Prompt */}
          <div>
            <label htmlFor="update-prompt" className="block text-sm font-medium text-foreground mb-2">
              <Sparkles className="w-4 h-4 inline mr-2" />
              What would you like to change?
            </label>
            <textarea
              id="update-prompt"
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
              placeholder="Describe the changes you want to make. For example:

• Change priority to high because this is blocking other tasks
• Move to implementing state since planning is complete
• Update description to include additional requirements
• Change owner to john.doe@company.com
• Add 'frontend' and 'react' tags to this task
• Mark as needs fixes due to failing tests"
              rows={8}
              className="input-tech w-full resize-none"
              disabled={updateTaskMutation.isPending}
              required
            />
            <p className="text-xs text-muted-foreground mt-1">
              AI will analyze your request and update the appropriate fields
            </p>
          </div>

          {/* AI Processing Info */}
          <div className="bg-muted/20 rounded-lg p-4 space-y-3">
            <h3 className="text-sm font-semibold text-foreground flex items-center">
              <Sparkles className="w-4 h-4 mr-2 text-orange-400" />
              What AI can update:
            </h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-xs">
              <div className="flex items-start space-x-2">
                <CheckCircle className="w-3 h-3 text-green-400 mt-0.5 flex-shrink-0" />
                <span className="text-muted-foreground">
                  Title and description
                </span>
              </div>
              <div className="flex items-start space-x-2">
                <CheckCircle className="w-3 h-3 text-green-400 mt-0.5 flex-shrink-0" />
                <span className="text-muted-foreground">
                  Priority and state
                </span>
              </div>
              <div className="flex items-start space-x-2">
                <CheckCircle className="w-3 h-3 text-green-400 mt-0.5 flex-shrink-0" />
                <span className="text-muted-foreground">
                  Tags and owner assignment
                </span>
              </div>
              <div className="flex items-start space-x-2">
                <CheckCircle className="w-3 h-3 text-green-400 mt-0.5 flex-shrink-0" />
                <span className="text-muted-foreground">
                  Dependencies (if mentioned)
                </span>
              </div>
            </div>
            <div className="mt-3 p-3 bg-yellow-500/10 border border-yellow-500/20 rounded">
              <p className="text-xs text-yellow-200">
                <strong>Note:</strong> State changes must follow valid transitions. AI will validate and suggest corrections if needed.
              </p>
            </div>
          </div>

          {/* Error Display */}
          <AnimatePresence>
            {updateTaskMutation.isError && (
              <motion.div
                initial={{ opacity: 0, y: -10 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -10 }}
                className="flex items-start space-x-2 p-3 bg-destructive/10 border border-destructive/20 rounded-lg"
              >
                <AlertCircle className="w-4 h-4 text-destructive mt-0.5 flex-shrink-0" />
                <div className="text-sm">
                  <p className="font-medium text-destructive">Failed to update task</p>
                  <p className="text-destructive/80 mt-1">
                    {updateTaskMutation.error instanceof Error
                      ? updateTaskMutation.error.message
                      : 'An unexpected error occurred. Please try again.'}
                  </p>
                </div>
              </motion.div>
            )}
          </AnimatePresence>

          {/* Actions */}
          <div className="flex items-center justify-end space-x-3 pt-4">
            <button
              type="button"
              onClick={() => onOpenChange(false)}
              className="btn-tech-ghost"
              disabled={updateTaskMutation.isPending}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="btn-tech-primary"
              disabled={updateTaskMutation.isPending || !prompt.trim()}
            >
              {updateTaskMutation.isPending ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  Updating Task...
                </>
              ) : (
                <>
                  <Sparkles className="w-4 h-4 mr-2" />
                  Update Task
                </>
              )}
            </button>
          </div>
        </form>
      </motion.div>
    </div>
  )
}