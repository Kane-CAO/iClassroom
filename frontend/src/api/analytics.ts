import { apiClient } from './client'
import type { Analytics } from '../types/api'

interface TeacherAuth {
  teacherToken: string
}

export function getAnalytics(roomCode: string, auth: TeacherAuth) {
  return apiClient.get<Analytics>(`/teacher/rooms/${roomCode}/analytics`, { teacherToken: auth.teacherToken })
}
