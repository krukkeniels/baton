'use client'

import { KanbanBoard } from './components/KanbanBoard'

export default function Home() {
  return (
    <div className="h-screen w-full overflow-hidden bg-background">
      <KanbanBoard />
    </div>
  )
}