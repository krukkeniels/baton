'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { formatDistanceToNow } from 'date-fns'
import {
  Calendar,
  Clock,
  Tag,
  User,
  ChevronRight,
  Link as LinkIcon,
} from 'lucide-react'

import { Task, PRIORITY_CONFIG, STATE_CONFIG } from '../types'
import { TaskDetailDialog } from './TaskDetailDialog'

interface TaskCardProps {
  task: Task
  isDragging?: boolean
}

export function TaskCard({ task, isDragging }: TaskCardProps) {
  const [isDetailDialogOpen, setIsDetailDialogOpen] = useState(false)

  const priorityConfig = PRIORITY_CONFIG[task.priority as keyof typeof PRIORITY_CONFIG]
  const stateConfig = STATE_CONFIG[task.state]

  const handleCardClick = (e: React.MouseEvent) => {
    // Don't open dialog if clicking on interactive elements
    if ((e.target as HTMLElement).closest('button')) {
      return
    }
    setIsDetailDialogOpen(true)
  }

  return (
    <>
      <motion.div
        className={`task-card cursor-pointer transition-all duration-200 ${
          isDragging ? 'shadow-lg scale-105' : 'hover:shadow-md hover:-translate-y-0.5'
        }`}
        data-state={task.state}
        onClick={handleCardClick}
        whileHover={{ scale: isDragging ? 1.05 : 1.02 }}
        whileTap={{ scale: 0.98 }}
      >
        {/* Header */}
        <div className="flex items-start justify-between mb-3">
          <div className="flex-1 min-w-0">
            <h4 className="font-semibold text-sm text-foreground truncate pr-2">
              {task.title}
            </h4>
            {task.description && (
              <p className="text-xs text-muted-foreground mt-1 line-clamp-2">
                {task.description}
              </p>
            )}
          </div>
          <div className="flex items-center space-x-1">
            {task.dependencies.length > 0 && (
              <div className="text-xs text-muted-foreground bg-muted rounded px-1.5 py-0.5">
                <LinkIcon className="w-3 h-3" />
              </div>
            )}
            <ChevronRight className="w-4 h-4 text-muted-foreground opacity-50" />
          </div>
        </div>

        {/* Tags */}
        {task.tags.length > 0 && (
          <div className="flex flex-wrap gap-1 mb-3">
            {task.tags.slice(0, 3).map((tag) => (
              <span key={tag} className="badge-tag text-xs">
                {tag}
              </span>
            ))}
            {task.tags.length > 3 && (
              <span className="badge-tag text-xs">
                +{task.tags.length - 3}
              </span>
            )}
          </div>
        )}

        {/* Footer */}
        <div className="flex items-center justify-between text-xs text-muted-foreground">
          <div className="flex items-center space-x-3">
            {/* Priority */}
            <div className="flex items-center space-x-1">
              <span className={`badge ${priorityConfig.color}`}>
                P{task.priority}
              </span>
            </div>

            {/* Owner */}
            {task.owner && (
              <div className="flex items-center space-x-1">
                <User className="w-3 h-3" />
                <span className="truncate max-w-[60px]">
                  {task.owner}
                </span>
              </div>
            )}
          </div>

          {/* Updated timestamp */}
          <div className="flex items-center space-x-1 text-xs">
            <Clock className="w-3 h-3" />
            <span>
              {formatDistanceToNow(new Date(task.updated_at), { addSuffix: true })}
            </span>
          </div>
        </div>

        {/* Progress indicator */}
        <div className="mt-3 pt-2 border-t border-border/50">
          <div className="flex items-center justify-between text-xs">
            <span className="text-muted-foreground">
              {stateConfig.label}
            </span>
            <div className="flex items-center space-x-1">
              <span className="text-muted-foreground">
                {stateConfig.icon}
              </span>
            </div>
          </div>
        </div>
      </motion.div>

      {/* Task Detail Dialog */}
      <TaskDetailDialog
        task={task}
        open={isDetailDialogOpen}
        onOpenChange={setIsDetailDialogOpen}
      />
    </>
  )
}