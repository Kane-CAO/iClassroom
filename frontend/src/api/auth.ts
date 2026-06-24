import { apiClient } from './client'
import type { AccountStatus, AuthLoginResponse, AuthUser, TeacherAccount } from '../types/api'

export interface LoginRequest {
  username: string
  password: string
}

export function loginAdmin(body: LoginRequest) {
  return apiClient.post<AuthLoginResponse, LoginRequest>('/auth/admin/login', body)
}

export function loginTeacher(body: LoginRequest) {
  return apiClient.post<AuthLoginResponse, LoginRequest>('/auth/teacher/login', body)
}

export function logout(token: string) {
  return apiClient.post<Record<string, never>>('/auth/logout', undefined, { token })
}

export function getMe(token: string) {
  return apiClient.get<AuthUser>('/auth/me', { token })
}

export interface CreateTeacherRequest {
  username: string
  displayName: string
  initialPassword: string
}

export interface UpdateTeacherStatusRequest {
  status: AccountStatus
}

export interface ResetTeacherPasswordRequest {
  newPassword?: string
}

export interface ResetTeacherPasswordResponse extends TeacherAccount {
  temporaryPassword: string
}

export function createTeacherAccount(body: CreateTeacherRequest, token: string) {
  return apiClient.post<TeacherAccount, CreateTeacherRequest>('/admin/teachers', body, { token })
}

export function listTeacherAccounts(token: string) {
  return apiClient.get<TeacherAccount[]>('/admin/teachers', { token })
}

export function updateTeacherStatus(teacherId: number, body: UpdateTeacherStatusRequest, token: string) {
  return apiClient.patch<TeacherAccount, UpdateTeacherStatusRequest>(`/admin/teachers/${teacherId}/status`, body, {
    token,
  })
}

export function resetTeacherPassword(teacherId: number, body: ResetTeacherPasswordRequest, token: string) {
  return apiClient.post<ResetTeacherPasswordResponse, ResetTeacherPasswordRequest>(
    `/admin/teachers/${teacherId}/reset-password`,
    body,
    { token },
  )
}

export function deleteTeacherAccount(teacherId: number, token: string) {
  return apiClient.request<{ teacherId: number }>(`/admin/teachers/${teacherId}`, { method: 'DELETE', token })
}
