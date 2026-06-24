import { apiClient } from './client'
import type { DisplayState } from '../types/api'

interface TeacherAuth {
  token?: string
  teacherToken?: string
}

export function getDisplayState(roomCode: string, auth?: TeacherAuth) {
  return apiClient.get<DisplayState>(`/teacher/rooms/${roomCode}/display`, auth)
}
