import { apiClient } from './client'
import type { DisplayState } from '../types/api'

export function getDisplayState(roomCode: string) {
  return apiClient.get<DisplayState>(`/teacher/rooms/${roomCode}/display`)
}
