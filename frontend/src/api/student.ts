import { apiClient } from './client'
import type { JoinRoomRequest, LeaderboardEntry, Room, StudentResults, StudentSession, Task } from '../types/api'

interface StudentAuth {
  studentToken: string
}

export type StudentRoomResponse = Room & {
  groups: NonNullable<Room['groups']>
}

export function getStudentRoom(roomCode: string) {
  return apiClient.get<StudentRoomResponse>(`/student/rooms/${roomCode}`)
}

export function joinStudentRoom(roomCode: string, body: JoinRoomRequest) {
  return apiClient.post<StudentSession, JoinRoomRequest>(`/student/rooms/${roomCode}/join`, body)
}

export function resumeStudentRoom(roomCode: string, auth: StudentAuth) {
  return apiClient.post<Omit<StudentSession, 'clientToken'>>(`/student/rooms/${roomCode}/resume`, undefined, {
    studentToken: auth.studentToken,
  })
}

export function listStudentTasks(auth: StudentAuth) {
  return apiClient.get<Task[]>('/student/me/tasks', { studentToken: auth.studentToken })
}

export function getStudentTask(taskId: number, auth: StudentAuth) {
  return apiClient.get<Task>(`/student/tasks/${taskId}`, { studentToken: auth.studentToken })
}

export function getStudentResults(auth: StudentAuth) {
  return apiClient.get<StudentResults>('/student/me/results', { studentToken: auth.studentToken })
}

export function getStudentRanking(roomCode: string, auth: StudentAuth) {
  return apiClient.get<LeaderboardEntry[]>(`/student/rooms/${roomCode}/ranking`, { studentToken: auth.studentToken })
}
