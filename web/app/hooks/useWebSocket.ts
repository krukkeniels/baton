'use client'

import { useEffect, useRef, useState, useCallback } from 'react'
import { WSMessage } from '../types'

const WS_URL = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:3001/api/ws'

export function useWebSocket() {
  const [isConnected, setIsConnected] = useState(false)
  const [lastMessage, setLastMessage] = useState<WSMessage | null>(null)
  const ws = useRef<WebSocket | null>(null)
  const reconnectTimer = useRef<NodeJS.Timeout | null>(null)

  const connect = useCallback(() => {
    try {
      ws.current = new WebSocket(WS_URL)

      ws.current.onopen = () => {
        console.log('WebSocket connected')
        setIsConnected(true)

        // Clear any reconnect timer
        if (reconnectTimer.current) {
          clearTimeout(reconnectTimer.current)
          reconnectTimer.current = null
        }
      }

      ws.current.onmessage = (event) => {
        try {
          const message: WSMessage = JSON.parse(event.data)
          setLastMessage(message)
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error)
        }
      }

      ws.current.onclose = () => {
        console.log('WebSocket disconnected')
        setIsConnected(false)

        // Attempt to reconnect after 3 seconds
        reconnectTimer.current = setTimeout(() => {
          console.log('Attempting to reconnect...')
          connect()
        }, 3000)
      }

      ws.current.onerror = (error) => {
        console.error('WebSocket error:', error)
        setIsConnected(false)
      }
    } catch (error) {
      console.error('Failed to connect WebSocket:', error)
      // Retry connection after 5 seconds
      reconnectTimer.current = setTimeout(connect, 5000)
    }
  }, [])

  useEffect(() => {
    connect()

    return () => {
      // Clear reconnect timer
      if (reconnectTimer.current) {
        clearTimeout(reconnectTimer.current)
      }

      // Close WebSocket connection
      if (ws.current) {
        ws.current.close()
      }
    }
  }, [connect])

  return {
    isConnected,
    lastMessage,
  }
}