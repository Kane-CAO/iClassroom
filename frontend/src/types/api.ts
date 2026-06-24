export type RoomStatus = 'created' | 'active' | 'ended'
export type TaskStatus = 'published' | 'paused' | 'closed'
export type TaskTargetType = 'all' | 'groups'
export type SubmissionStatus = 'submitted' | 'graded'
export type StudentSubmissionStatus = SubmissionStatus | 'notSubmitted'
export type DisplayMode = 'anonymous' | 'showGroup'
export type UserRole = 'admin' | 'teacher'
export type AccountStatus = 'active' | 'disabled'

export interface AuthUser {
  userId: number
  role: UserRole
  username: string
  displayName?: string
}

export interface AuthLoginResponse {
  token: string
  user: AuthUser
}

export interface TeacherAccount {
  teacherId: number
  username: string
  displayName: string
  status: AccountStatus
  createdAt: string
  updatedAt?: string
  lastLoginAt?: string
}

export interface Group {
  groupId: number
  groupName: string
  capacity: number
  currentCount: number
  available?: boolean
  scoreTotal?: number
}

export interface Room {
  roomCode: string
  title: string
  status: RoomStatus
  groupCount?: number
  groupCapacity?: number
  allowChooseGroup?: boolean
  createdAt?: string
  endedAt?: string | null
  groups?: Group[]
}

export interface StudentSession {
  studentId: number
  roomCode: string
  nickname: string
  groupId: number
  groupName: string
  clientToken: string
  roomStatus?: RoomStatus
}

export interface Task {
  taskId: number
  roomCode?: string
  title: string
  description: string
  attachmentUrl?: string
  deadlineAt: string
  targetType?: TaskTargetType
  targetGroupIds?: number[]
  status: TaskStatus
  submittedCount?: number
  targetStudentCount?: number
  createdAt?: string
  mySubmissionStatus?: StudentSubmissionStatus
  myScore?: number | null
}

export interface SubmissionImage {
  imageId: number
  fileUrl: string
  fileName: string
  fileSize: number
  mimeType: string
}

export interface SubmissionFile {
  attachmentId: number
  fileUrl: string
  fileName: string
  storedFileName?: string
  fileSize: number
  mimeType: string
  originalFileName?: string
}

export interface Submission {
  submissionId: number
  taskId: number
  studentId: number
  groupId: number
  contentText: string
  images: SubmissionImage[]
  files?: SubmissionFile[]
  status: SubmissionStatus
  score: number | null
  comment: string
  submittedAt: string
  gradedAt: string | null
  nickname?: string
  groupName?: string
}

export interface LeaderboardEntry {
  rank: number
  groupId: number
  groupName: string
  scoreTotal: number
  currentCount?: number
  studentCount?: number
  isMyGroup?: boolean
}

export interface Leaderboard {
  roomCode: string
  leaderboard: LeaderboardEntry[]
}

export interface DisplayTask {
  taskId: number
  title: string
  deadlineAt: string
  submittedCount: number
  targetStudentCount: number
  completionRate: number
}

export interface DisplayFeaturedAnswer {
  featuredId: number
  submissionId: number
  taskId: number
  displayMode: DisplayMode
  contentText: string
  score: number | null
  submittedAt: string
  groupId?: number
  groupName?: string
}

export interface DisplayState {
  roomCode: string
  title: string
  status: RoomStatus
  groups: Group[]
  ranking: LeaderboardEntry[]
  currentTask: DisplayTask | null
  featuredAnswers: DisplayFeaturedAnswer[]
}

export interface AnalyticsGroupScore {
  groupId: number
  groupName: string
  scoreTotal: number
}

export interface AnalyticsTaskCompletion {
  taskId: number
  taskTitle: string
  submittedCount: number
  targetStudentCount: number
  completionRate: number
}

export interface AnalyticsSubmissionTimelinePoint {
  time: string
  count: number
}

export interface Analytics {
  studentCount: number
  onlineCount: number
  submissionRate: number
  groupScores: AnalyticsGroupScore[]
  taskCompletion: AnalyticsTaskCompletion[]
  submissionTimeline: AnalyticsSubmissionTimelinePoint[]
}

export interface Export {
  blob: Blob
  fileName: string
  contentType: string
}

export interface CreateRoomRequest {
  title: string
  groupCount: number
  groupCapacity: number
  allowChooseGroup: boolean
}

export interface CreateRoomResponse {
  roomCode: string
  teacherToken: string
  joinUrl: string
  teacherDashboardUrl: string
  groups: Group[]
}

export interface RoomOverview {
  roomCode: string
  title: string
  status: RoomStatus
  joinUrl: string
  studentCount: number
  groups: Group[]
  tasks: Task[]
}

export interface JoinRoomRequest {
  nickname: string
  groupId: number
}

export interface CreateTaskRequest {
  title: string
  description: string
  attachmentUrl?: string
  deadlineAt: string
  targetType: TaskTargetType
  targetGroupIds: number[]
}

export interface TaskStatusResponse {
  taskId: number
  status: TaskStatus
}

export interface SubmitTaskRequest {
  contentText: string
}

export interface GradeSubmissionRequest {
  score: number
  comment: string
}

export interface GradeSubmissionResponse {
  submissionId: number
  score: number
  comment: string
  status: SubmissionStatus
  gradedAt: string
  groupScoreTotal: number
}

export interface StudentResult {
  taskId: number
  taskTitle: string
  submissionStatus: StudentSubmissionStatus
  score: number | null
  comment: string
  submittedAt: string | null
  gradedAt: string | null
}

export interface StudentResults {
  studentId: number
  nickname: string
  groupId: number
  groupName: string
  results: StudentResult[]
}

export interface FeatureSubmissionRequest {
  displayMode: DisplayMode
}

export interface FeatureSubmissionResponse {
  featuredId: number
  submissionId: number
  displayMode: DisplayMode
}
