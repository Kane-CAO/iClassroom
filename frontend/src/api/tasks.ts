import { apiClient } from './client'
import type {
  CreateTaskRequest,
  FeatureSubmissionRequest,
  FeatureSubmissionResponse,
  GradeSubmissionRequest,
  GradeSubmissionResponse,
  Leaderboard,
  StudentResults,
  Submission,
  SubmitTaskRequest,
  Task,
  TaskStatusResponse,
} from '../types/api'

interface TeacherAuth {
  token?: string
  teacherToken?: string
}

interface StudentAuth {
  studentToken: string
}

export function createTask(roomCode: string, body: CreateTaskRequest, auth: TeacherAuth) {
  return apiClient.post<Task, CreateTaskRequest>(`/teacher/rooms/${roomCode}/tasks`, body, auth)
}

export function listTeacherTasks(roomCode: string, auth: TeacherAuth) {
  return apiClient.get<Task[]>(`/teacher/rooms/${roomCode}/tasks`, auth)
}

export function listStudentTasks(auth: StudentAuth) {
  return apiClient.get<Task[]>('/student/me/tasks', { studentToken: auth.studentToken })
}

export function pauseTask(taskId: number, auth: TeacherAuth) {
  return apiClient.patch<TaskStatusResponse>(`/teacher/tasks/${taskId}/pause`, undefined, auth)
}

export function closeTask(taskId: number, auth: TeacherAuth) {
  return apiClient.patch<TaskStatusResponse>(`/teacher/tasks/${taskId}/close`, undefined, auth)
}

export function submitTask(taskId: number, body: SubmitTaskRequest, auth: StudentAuth) {
  return apiClient.post<Submission, SubmitTaskRequest>(`/student/tasks/${taskId}/submit`, body, {
    studentToken: auth.studentToken,
  })
}

export function submitTaskForm(taskId: number, body: FormData, auth: StudentAuth) {
  return apiClient.post<Submission, FormData>(`/student/tasks/${taskId}/submit`, body, {
    studentToken: auth.studentToken,
  })
}

export function listTaskSubmissions(taskId: number, auth: TeacherAuth) {
  return apiClient.get<Submission[]>(`/teacher/tasks/${taskId}/submissions`, auth)
}

export function gradeSubmission(submissionId: number, body: GradeSubmissionRequest, auth: TeacherAuth) {
  return apiClient.post<GradeSubmissionResponse, GradeSubmissionRequest>(
    `/teacher/submissions/${submissionId}/grade`,
    body,
    auth,
  )
}

export function getStudentResults(auth: StudentAuth) {
  return apiClient.get<StudentResults>('/student/me/results', { studentToken: auth.studentToken })
}

export function getTeacherLeaderboard(roomCode: string, auth: TeacherAuth) {
  return apiClient.get<Leaderboard>(`/teacher/rooms/${roomCode}/leaderboard`, auth)
}

export function getStudentLeaderboard(auth: StudentAuth) {
  return apiClient.get<Leaderboard>('/student/me/leaderboard', { studentToken: auth.studentToken })
}

export function featureSubmission(submissionId: number, body: FeatureSubmissionRequest, auth: TeacherAuth) {
  return apiClient.post<FeatureSubmissionResponse, FeatureSubmissionRequest>(
    `/teacher/submissions/${submissionId}/feature`,
    body,
    auth,
  )
}
