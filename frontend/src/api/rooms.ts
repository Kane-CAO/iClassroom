import { apiClient } from './client'
import type { CreateRoomRequest, CreateRoomResponse, Room, RoomOverview } from '../types/api'

interface TeacherAuth {
  token?: string
  teacherToken?: string
}

export function createRoom(body: CreateRoomRequest, auth?: TeacherAuth) {
  return apiClient.post<CreateRoomResponse, CreateRoomRequest>('/teacher/rooms', body, auth)
}

export function getRoom(roomCode: string, auth: TeacherAuth) {
  return apiClient.get<Room>(`/teacher/rooms/${roomCode}`, auth)
}

export function getRoomOverview(roomCode: string, auth: TeacherAuth) {
  return apiClient.get<RoomOverview>(`/teacher/rooms/${roomCode}/overview`, auth)
}

export function endRoom(roomCode: string, auth: TeacherAuth) {
  return apiClient.post<Pick<Room, 'roomCode' | 'status' | 'endedAt'>>(
    `/teacher/rooms/${roomCode}/end`,
    undefined,
    auth,
  )
}
