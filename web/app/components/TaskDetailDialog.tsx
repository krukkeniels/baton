'use client'

import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { motion, AnimatePresence } from 'framer-motion'
import { formatDistanceToNow, format } from 'date-fns'
import {
  X,
  Calendar,
  Clock,
  User,
  Tag,
  Link as LinkIcon,
  FileText,
  History,
  Edit3,
  RefreshCw,
  ChevronDown,
  ChevronRight,
} from 'lucide-react'

import { Task, PRIORITY_CONFIG, STATE_CONFIG } from '../types'
import { apiClient } from '../lib/api'
import { UpdateTaskDialog } from './UpdateTaskDialog'

interface TaskDetailDialogProps {
  task: Task
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function TaskDetailDialog({ task, open, onOpenChange }: TaskDetailDialogProps) {
  const [isUpdateDialogOpen, setIsUpdateDialogOpen] = useState(false)
  const [activeTab, setActiveTab] = useState<'overview' | 'artifacts' | 'history'>('overview')
  const [expandedArtifacts, setExpandedArtifacts] = useState<Set<string>>(new Set())

  const { data: auditHistory, isLoading: historyLoading } = useQuery({
    queryKey: ['audit', task.id],
    queryFn: () => apiClient.getAuditHistory(task.id),
    enabled: open && activeTab === 'history',
  })

  const { data: taskDetail, isLoading: detailLoading } = useQuery({
    queryKey: ['task', task.id],
    queryFn: () => apiClient.getTask(task.id),
    enabled: open,
  })

  const priorityConfig = PRIORITY_CONFIG[task.priority as keyof typeof PRIORITY_CONFIG]
  const stateConfig = STATE_CONFIG[task.state]

  const toggleArtifact = (artifactName: string) => {
    const newExpanded = new Set(expandedArtifacts)
    if (newExpanded.has(artifactName)) {
      newExpanded.delete(artifactName)
    } else {
      newExpanded.add(artifactName)
    }
    setExpandedArtifacts(newExpanded)
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
        className="relative bg-card border border-border rounded-lg shadow-2xl w-full max-w-4xl max-h-[90vh] m-4"
        initial={{ opacity: 0, scale: 0.95, y: 20 }}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        exit={{ opacity: 0, scale: 0.95, y: 20 }}
        transition={{ duration: 0.2 }}
      >
        {/* Header */}
        <div className="flex items-start justify-between p-6 border-b border-border">
          <div className="flex-1 min-w-0">
            <div className="flex items-center space-x-3 mb-2">
              <h2 className="text-xl font-bold text-foreground">
                {task.title}
              </h2>
              <span className={`badge ${priorityConfig.color}`}>
                Priority {task.priority}
              </span>
              <span className="px-3 py-1 rounded-full text-xs font-medium bg-blue-500/20 text-blue-400 border border-blue-500/30">
                {stateConfig.icon} {stateConfig.label}
              </span>
            </div>
            <div className="flex items-center space-x-4 text-sm text-muted-foreground">
              <div className="flex items-center space-x-1">
                <User className="w-4 h-4" />
                <span>{task.owner}</span>
              </div>
              <div className="flex items-center space-x-1">
                <Calendar className="w-4 h-4" />
                <span>Created {format(new Date(task.created_at), 'MMM d, yyyy')}</span>
              </div>
              <div className="flex items-center space-x-1">
                <Clock className="w-4 h-4" />
                <span>Updated {formatDistanceToNow(new Date(task.updated_at), { addSuffix: true })}</span>
              </div>
            </div>
          </div>
          <div className="flex items-center space-x-2">
            <button
              onClick={() => setIsUpdateDialogOpen(true)}
              className="btn-tech-secondary"
            >
              <Edit3 className="w-4 h-4 mr-2" />
              Update
            </button>
            <button
              onClick={() => onOpenChange(false)}
              className="btn-tech-ghost p-2"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* Tabs */}
        <div className="border-b border-border">
          <nav className="flex space-x-8 px-6" aria-label="Tabs">
            {[
              { id: 'overview', label: 'Overview', icon: FileText },
              { id: 'artifacts', label: 'Artifacts', icon: FileText, count: taskDetail?.artifacts?.length },
              { id: 'history', label: 'History', icon: History },
            ].map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id as any)}
                className={`flex items-center space-x-2 py-4 px-1 border-b-2 font-medium text-sm transition-colors ${
                  activeTab === tab.id
                    ? 'border-primary text-primary'
                    : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
                }`}
              >
                <tab.icon className="w-4 h-4" />
                <span>{tab.label}</span>
                {tab.count !== undefined && tab.count > 0 && (
                  <span className="bg-muted text-muted-foreground px-2 py-0.5 rounded-full text-xs">
                    {tab.count}
                  </span>
                )}
              </button>
            ))}
          </nav>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6">
          <AnimatePresence mode="wait">
            {activeTab === 'overview' && (
              <motion.div
                key="overview"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -20 }}
                className="space-y-6"
              >
                {/* Description */}
                {task.description && (
                  <div>
                    <h3 className="text-lg font-semibold mb-3">Description</h3>
                    <div className="bg-muted/20 rounded-lg p-4">
                      <pre className="whitespace-pre-wrap text-sm text-foreground">
                        {task.description}
                      </pre>
                    </div>
                  </div>
                )}

                {/* Tags */}
                {task.tags.length > 0 && (
                  <div>
                    <h3 className="text-lg font-semibold mb-3 flex items-center">
                      <Tag className="w-5 h-5 mr-2" />
                      Tags
                    </h3>
                    <div className="flex flex-wrap gap-2">
                      {task.tags.map((tag) => (
                        <span key={tag} className="badge-tag">
                          {tag}
                        </span>
                      ))}
                    </div>
                  </div>
                )}

                {/* Dependencies */}
                {task.dependencies.length > 0 && (
                  <div>
                    <h3 className="text-lg font-semibold mb-3 flex items-center">
                      <LinkIcon className="w-5 h-5 mr-2" />
                      Dependencies
                    </h3>
                    <div className="space-y-2">
                      {task.dependencies.map((depId) => (
                        <div key={depId} className="bg-muted/20 rounded-lg p-3">
                          <div className="font-mono text-sm text-muted-foreground">
                            Task ID: {depId}
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </motion.div>
            )}

            {activeTab === 'artifacts' && (
              <motion.div
                key="artifacts"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -20 }}
                className="space-y-4"
              >
                {detailLoading ? (
                  <div className="flex items-center justify-center py-8">
                    <RefreshCw className="w-6 h-6 animate-spin text-muted-foreground" />
                  </div>
                ) : taskDetail?.artifacts && taskDetail.artifacts.length > 0 ? (
                  taskDetail.artifacts.map((artifact) => (
                    <div key={`${artifact.name}-${artifact.version}`} className="border border-border rounded-lg">
                      <button
                        onClick={() => toggleArtifact(artifact.name)}
                        className="w-full flex items-center justify-between p-4 text-left hover:bg-accent/20 transition-colors"
                      >
                        <div className="flex items-center space-x-3">
                          <FileText className="w-5 h-5 text-muted-foreground" />
                          <div>
                            <h4 className="font-semibold">{artifact.name}</h4>
                            <p className="text-sm text-muted-foreground">
                              Version {artifact.version} • {format(new Date(artifact.created_at), 'MMM d, yyyy HH:mm')}
                            </p>
                          </div>
                        </div>
                        {expandedArtifacts.has(artifact.name) ? (
                          <ChevronDown className="w-5 h-5 text-muted-foreground" />
                        ) : (
                          <ChevronRight className="w-5 h-5 text-muted-foreground" />
                        )}
                      </button>
                      <AnimatePresence>
                        {expandedArtifacts.has(artifact.name) && (
                          <motion.div
                            initial={{ height: 0, opacity: 0 }}
                            animate={{ height: 'auto', opacity: 1 }}
                            exit={{ height: 0, opacity: 0 }}
                            className="border-t border-border"
                          >
                            <div className="p-4 bg-muted/20">
                              <pre className="whitespace-pre-wrap text-sm text-foreground overflow-x-auto">
                                {artifact.content}
                              </pre>
                            </div>
                          </motion.div>
                        )}
                      </AnimatePresence>
                    </div>
                  ))
                ) : (
                  <div className="text-center py-8 text-muted-foreground">
                    <FileText className="w-12 h-12 mx-auto mb-4 opacity-50" />
                    <p>No artifacts found for this task.</p>
                  </div>
                )}
              </motion.div>
            )}

            {activeTab === 'history' && (
              <motion.div
                key="history"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -20 }}
                className="space-y-4"
              >
                {historyLoading ? (
                  <div className="flex items-center justify-center py-8">
                    <RefreshCw className="w-6 h-6 animate-spin text-muted-foreground" />
                  </div>
                ) : auditHistory && auditHistory.length > 0 ? (
                  <div className="space-y-4">
                    {auditHistory.map((entry, index) => (
                      <div key={entry.id} className="relative">
                        {index < auditHistory.length - 1 && (
                          <div className="absolute left-4 top-12 bottom-0 w-0.5 bg-border" />
                        )}
                        <div className="flex items-start space-x-4">
                          <div className="flex-shrink-0 w-8 h-8 bg-primary/20 rounded-full flex items-center justify-center">
                            <div className="w-2 h-2 bg-primary rounded-full" />
                          </div>
                          <div className="flex-1 min-w-0">
                            <div className="bg-muted/20 rounded-lg p-4">
                              <div className="flex items-center justify-between mb-2">
                                <div className="flex items-center space-x-2">
                                  <span className="font-semibold text-sm">
                                    {entry.prev_state} → {entry.next_state}
                                  </span>
                                  <span className="text-xs text-muted-foreground">
                                    by {entry.actor}
                                  </span>
                                </div>
                                <span className="text-xs text-muted-foreground">
                                  {format(new Date(entry.created_at), 'MMM d, yyyy HH:mm')}
                                </span>
                              </div>
                              {entry.reason && (
                                <p className="text-sm text-muted-foreground mb-2">
                                  {entry.reason}
                                </p>
                              )}
                              {entry.note && (
                                <p className="text-sm text-foreground">
                                  {entry.note}
                                </p>
                              )}
                              {entry.commands && entry.commands.length > 0 && (
                                <div className="mt-2">
                                  <p className="text-xs text-muted-foreground mb-1">Commands executed:</p>
                                  <div className="bg-background/50 rounded p-2 font-mono text-xs">
                                    {entry.commands.map((cmd, i) => (
                                      <div key={i}>{cmd}</div>
                                    ))}
                                  </div>
                                </div>
                              )}
                            </div>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="text-center py-8 text-muted-foreground">
                    <History className="w-12 h-12 mx-auto mb-4 opacity-50" />
                    <p>No history available for this task.</p>
                  </div>
                )}
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      </motion.div>

      {/* Update Task Dialog */}
      <UpdateTaskDialog
        task={task}
        open={isUpdateDialogOpen}
        onOpenChange={setIsUpdateDialogOpen}
      />
    </div>
  )
}