import { apiClient } from './client'
import type { CreateRoomRequest, CreateRoomResponse, Room, RoomOverview } from '../types/api'

interface TeacherAuth {
  teacherToken: string
}

export function createRoom(body: CreateRoomRequest) {
  return apiClient.post<CreateRoomResponse, CreateRoomRequest>('/teacher/rooms', body)
}

export function getRoom(roomCode: string, auth: TeacherAuth) {
  return apiClient.get<Room>(`/teacher/rooms/${roomCode}`, { teacherToken: auth.teacherToken })
}

export function getRoomOverview(roomCode: string, auth: TeacherAuth) {
  return apiClient.get<RoomOverview>(`/teacher/rooms/${roomCode}/overview`, { teacherToken: auth.teacherToken })
}

export function endRoom(roomCode: string, auth: TeacherAuth) {
  return apiClient.post<Pick<Room, 'roomCode' | 'status' | 'endedAt'>>(
    `/teacher/rooms/${roomCode}/end`,
    undefined,
    { teacherToken: auth.teacherToken },
  )
}
