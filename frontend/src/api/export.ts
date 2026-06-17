import { apiClient } from './client'
import type { Export } from '../types/api'

interface TeacherAuth {
  teacherToken: string
}

export function exportRoom(roomCode: string, auth: TeacherAuth): Promise<Export> {
  return apiClient.download(`/teacher/rooms/${roomCode}/export`, { teacherToken: auth.teacherToken })
}
