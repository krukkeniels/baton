import type { Metadata } from 'next'
import { Inter, JetBrains_Mono } from 'next/font/google'
import './globals.css'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryProvider } from './providers/react-query-provider'

const inter = Inter({
  subsets: ['latin'],
  variable: '--font-sans'
})

const jetbrainsMono = JetBrains_Mono({
  subsets: ['latin'],
  variable: '--font-mono'
})

export const metadata: Metadata = {
  title: 'Baton - CLI Orchestrator',
  description: 'Modern web interface for Baton CLI Orchestrator - LLM-driven task execution with cycle-based state machine progression',
  keywords: ['baton', 'cli', 'orchestrator', 'llm', 'task management', 'kanban'],
  authors: [{ name: 'Baton Team' }],
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en" className="dark">
      <body className={`${inter.variable} ${jetbrainsMono.variable} font-sans`}>
        <ReactQueryProvider>
          {children}
        </ReactQueryProvider>
      </body>
    </html>
  )
}