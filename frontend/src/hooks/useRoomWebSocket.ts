import { useEffect, useMemo, useRef, useState } from 'react'
import {
  RoomWebSocketEvent,
  RoomWebSocketRole,
  buildRoomWebSocketURL,
  parseRoomWebSocketEvent,
} from '../api/websocket'

export type RoomWebSocketStatus = 'idle' | 'connecting' | 'open' | 'reconnecting' | 'closed' | 'error'

export interface UseRoomWebSocketOptions {
  roomCode: string
  role: RoomWebSocketRole
  clientToken?: string
  enabled?: boolean
  reconnectDelayMs?: number
  maxReconnectDelayMs?: number
  onEvent?: (event: RoomWebSocketEvent) => void
  onReconnect?: () => void
}

export function useRoomWebSocket({
  roomCode,
  role,
  clientToken,
  enabled = true,
  reconnectDelayMs = 1000,
  maxReconnectDelayMs = 10000,
  onEvent,
  onReconnect,
}: UseRoomWebSocketOptions) {
  const [status, setStatus] = useState<RoomWebSocketStatus>('idle')
  const [lastEvent, setLastEvent] = useState<RoomWebSocketEvent | null>(null)

  const socketRef = useRef<WebSocket | null>(null)
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const reconnectAttemptRef = useRef(0)
  const shouldReconnectRef = useRef(false)
  const onEventRef = useRef<typeof onEvent>(onEvent)
  const onReconnectRef = useRef<typeof onReconnect>(onReconnect)

  useEffect(() => {
    onEventRef.current = onEvent
  }, [onEvent])

  useEffect(() => {
    onReconnectRef.current = onReconnect
  }, [onReconnect])

  const url = useMemo(() => {
    if (!enabled) {
      return null
    }
    return buildRoomWebSocketURL({ roomCode, role, clientToken })
  }, [clientToken, enabled, role, roomCode])

  useEffect(() => {
    if (!url) {
      setStatus('idle')
      return undefined
    }

    shouldReconnectRef.current = true

    const clearReconnectTimer = () => {
      if (reconnectTimerRef.current) {
        clearTimeout(reconnectTimerRef.current)
        reconnectTimerRef.current = null
      }
    }

    const closeCurrentSocket = () => {
      if (socketRef.current) {
        socketRef.current.onopen = null
        socketRef.current.onmessage = null
        socketRef.current.onerror = null
        socketRef.current.onclose = null
        socketRef.current.close()
        socketRef.current = null
      }
    }

    const connect = (isReconnect: boolean) => {
      clearReconnectTimer()
      closeCurrentSocket()
      setStatus(isReconnect ? 'reconnecting' : 'connecting')

      const socket = new WebSocket(url)
      socketRef.current = socket

      socket.onopen = () => {
        reconnectAttemptRef.current = 0
        setStatus('open')
        if (isReconnect) {
          onReconnectRef.current?.()
        }
      }

      socket.onmessage = (message: MessageEvent<string>) => {
        const event = parseRoomWebSocketEvent(message.data)
        if (!event) {
          return
        }
        setLastEvent(event)
        onEventRef.current?.(event)
      }

      socket.onerror = () => {
        setStatus('error')
      }

      socket.onclose = () => {
        if (socketRef.current === socket) {
          socketRef.current = null
        }

        if (!shouldReconnectRef.current) {
          setStatus('closed')
          return
        }

        reconnectAttemptRef.current += 1
        const delay = Math.min(
          reconnectDelayMs * 2 ** Math.max(reconnectAttemptRef.current - 1, 0),
          maxReconnectDelayMs,
        )
        reconnectTimerRef.current = setTimeout(() => connect(true), delay)
      }
    }

    connect(false)

    return () => {
      shouldReconnectRef.current = false
      clearReconnectTimer()
      closeCurrentSocket()
      setStatus('closed')
    }
  }, [maxReconnectDelayMs, reconnectDelayMs, url])

  return {
    status,
    lastEvent,
    isConnected: status === 'open',
  }
}
