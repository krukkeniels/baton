'use client'

import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { motion, AnimatePresence } from 'framer-motion'
import { X, Sparkles, Loader2, User, CheckCircle, AlertCircle } from 'lucide-react'

import { apiClient } from '../lib/api'

interface CreateTaskDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function CreateTaskDialog({ open, onOpenChange }: CreateTaskDialogProps) {
  const [prompt, setPrompt] = useState('')
  const [owner, setOwner] = useState('')
  const queryClient = useQueryClient()

  const createTaskMutation = useMutation({
    mutationFn: (data: { prompt: string; owner?: string }) =>
      apiClient.createTask(data),
    onSuccess: () => {
      // Invalidate tasks query to refetch data
      queryClient.invalidateQueries({ queryKey: ['tasks'] })

      // Reset form and close dialog
      setPrompt('')
      setOwner('')
      onOpenChange(false)
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!prompt.trim()) return

    createTaskMutation.mutate({
      prompt: prompt.trim(),
      owner: owner.trim() || undefined,
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
            <div className="w-8 h-8 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
              <Sparkles className="w-4 h-4 text-white" />
            </div>
            <div>
              <h2 className="text-xl font-bold text-foreground">Create New Task</h2>
              <p className="text-sm text-muted-foreground">
                Describe what you want to accomplish and let AI create a structured task
              </p>
            </div>
          </div>
          <button
            onClick={() => onOpenChange(false)}
            className="btn-tech-ghost p-2"
            disabled={createTaskMutation.isPending}
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Content */}
        <form onSubmit={handleSubmit} className="p-6 space-y-6">
          {/* Owner (optional) */}
          <div>
            <label htmlFor="owner" className="block text-sm font-medium text-foreground mb-2">
              <User className="w-4 h-4 inline mr-2" />
              Owner (optional)
            </label>
            <input
              id="owner"
              type="text"
              value={owner}
              onChange={(e) => setOwner(e.target.value)}
              placeholder="Who will work on this task?"
              className="input-tech w-full"
              disabled={createTaskMutation.isPending}
            />
            <p className="text-xs text-muted-foreground mt-1">
              Leave empty to assign to system
            </p>
          </div>

          {/* Prompt */}
          <div>
            <label htmlFor="prompt" className="block text-sm font-medium text-foreground mb-2">
              <Sparkles className="w-4 h-4 inline mr-2" />
              Task Description
            </label>
            <textarea
              id="prompt"
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
              placeholder="Describe what needs to be done. Be as detailed as possible. For example:

• Create a user registration form with email validation
• Fix the authentication bug in the login system
• Implement dark mode toggle for the dashboard
• Add unit tests for the payment processing module"
              rows={8}
              className="input-tech w-full resize-none"
              disabled={createTaskMutation.isPending}
              required
            />
            <p className="text-xs text-muted-foreground mt-1">
              The more details you provide, the better the AI can structure your task
            </p>
          </div>

          {/* AI Processing Examples */}
          <div className="bg-muted/20 rounded-lg p-4 space-y-3">
            <h3 className="text-sm font-semibold text-foreground flex items-center">
              <Sparkles className="w-4 h-4 mr-2 text-blue-400" />
              What AI will do:
            </h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-xs">
              <div className="flex items-start space-x-2">
                <CheckCircle className="w-3 h-3 text-green-400 mt-0.5 flex-shrink-0" />
                <span className="text-muted-foreground">
                  Generate clear title and description
                </span>
              </div>
              <div className="flex items-start space-x-2">
                <CheckCircle className="w-3 h-3 text-green-400 mt-0.5 flex-shrink-0" />
                <span className="text-muted-foreground">
                  Set appropriate priority level
                </span>
              </div>
              <div className="flex items-start space-x-2">
                <CheckCircle className="w-3 h-3 text-green-400 mt-0.5 flex-shrink-0" />
                <span className="text-muted-foreground">
                  Extract relevant tags and keywords
                </span>
              </div>
              <div className="flex items-start space-x-2">
                <CheckCircle className="w-3 h-3 text-green-400 mt-0.5 flex-shrink-0" />
                <span className="text-muted-foreground">
                  Create acceptance criteria
                </span>
              </div>
            </div>
          </div>

          {/* Error Display */}
          <AnimatePresence>
            {createTaskMutation.isError && (
              <motion.div
                initial={{ opacity: 0, y: -10 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -10 }}
                className="flex items-start space-x-2 p-3 bg-destructive/10 border border-destructive/20 rounded-lg"
              >
                <AlertCircle className="w-4 h-4 text-destructive mt-0.5 flex-shrink-0" />
                <div className="text-sm">
                  <p className="font-medium text-destructive">Failed to create task</p>
                  <p className="text-destructive/80 mt-1">
                    {createTaskMutation.error instanceof Error
                      ? createTaskMutation.error.message
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
              disabled={createTaskMutation.isPending}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="btn-tech-primary"
              disabled={createTaskMutation.isPending || !prompt.trim()}
            >
              {createTaskMutation.isPending ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  Creating Task...
                </>
              ) : (
                <>
                  <Sparkles className="w-4 h-4 mr-2" />
                  Create Task
                </>
              )}
            </button>
          </div>
        </form>
      </motion.div>
    </div>
  )
}