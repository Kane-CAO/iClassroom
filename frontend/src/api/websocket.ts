const wsBaseURL = import.meta.env.VITE_WS_BASE_URL ?? ''

export type RoomWebSocketRole = 'teacher' | 'student' | 'display'

export type RoomWebSocketEventType =
  | 'student_joined'
  | 'task_published'
  | 'task_paused'
  | 'task_closed'
  | 'submission_created'
  | 'score_updated'
  | 'ranking_updated'
  | 'featured_answer_updated'
  | 'room_ended'

export interface RoomWebSocketEvent<TData extends Record<string, unknown> = Record<string, unknown>> {
  type: RoomWebSocketEventType
  roomCode: string
  data: TData
  occurredAt: string
}

export interface RoomWebSocketConfig {
  roomCode: string
  role: RoomWebSocketRole
  clientToken?: string
}

export function canConnectRoomWebSocket(config: RoomWebSocketConfig) {
  if (!wsBaseURL || !config.roomCode.trim()) {
    return false
  }
  if (config.role === 'student') {
    return Boolean(config.clientToken?.trim())
  }
  return true
}

export function buildRoomWebSocketURL(config: RoomWebSocketConfig) {
  if (!canConnectRoomWebSocket(config)) {
    return null
  }

  const url = new URL(wsBaseURL, getWebSocketOrigin())
  url.searchParams.set('room', config.roomCode.trim())
  url.searchParams.set('role', config.role)

  if (config.role === 'student' && config.clientToken) {
    url.searchParams.set('token', config.clientToken.trim())
  }

  return url.toString()
}

export function createRoomWebSocket(config: RoomWebSocketConfig) {
  const url = buildRoomWebSocketURL(config)
  if (!url) {
    return null
  }
  return new WebSocket(url)
}

export function parseRoomWebSocketEvent(message: string): RoomWebSocketEvent | null {
  try {
    const value = JSON.parse(message) as unknown
    if (!isRoomWebSocketEvent(value)) {
      return null
    }
    return value
  } catch {
    return null
  }
}

function getWebSocketOrigin() {
  if (typeof window === 'undefined') {
    return 'ws://localhost'
  }
  return window.location.origin.replace(/^http/, 'ws')
}

function isRoomWebSocketEvent(value: unknown): value is RoomWebSocketEvent {
  if (!isRecord(value)) {
    return false
  }

  return (
    isRoomWebSocketEventType(value.type) &&
    typeof value.roomCode === 'string' &&
    isRecord(value.data) &&
    typeof value.occurredAt === 'string'
  )
}

function isRoomWebSocketEventType(value: unknown): value is RoomWebSocketEventType {
  return (
    value === 'student_joined' ||
    value === 'task_published' ||
    value === 'task_paused' ||
    value === 'task_closed' ||
    value === 'submission_created' ||
    value === 'score_updated' ||
    value === 'ranking_updated' ||
    value === 'featured_answer_updated' ||
    value === 'room_ended'
  )
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}
