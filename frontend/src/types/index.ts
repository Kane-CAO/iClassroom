// 领域类型定义（对齐 README01.md 第 6 节数据库模型）。
// 仅前端视图所需字段，后续随接口联调补充。

export type RoomStatus = 'created' | 'active' | 'ended'
export type TaskStatus = 'published' | 'paused' | 'closed'
export type TaskTargetType = 'all' | 'groups'
export type SubmissionStatus = 'submitted' | 'graded'

export interface Room {
  id: number
  roomCode: string
  title: string
  groupCount: number
  groupCapacity: number
  allowChooseGroup: boolean
  status: RoomStatus
}

export interface Group {
  id: number
  roomId: number
  name: string
  capacity: number
  scoreTotal: number
}

export interface Task {
  id: number
  roomId: number
  title: string
  description: string
  attachmentUrl?: string
  deadlineAt: string
  status: TaskStatus
  targetType: TaskTargetType
}

export interface Submission {
  id: number
  taskId: number
  studentId: number
  groupId: number
  contentText: string
  status: SubmissionStatus
  score?: number
  comment?: string
  submittedAt: string
}

// ---------------------------------------------------------------------------
// 视图模型（View Model）：用于承载 docs/prototypes 原型中更丰富的展示字段。
// 这些类型只服务于当前 mock 阶段的页面渲染，待后端联调时再与领域类型对齐。
// ---------------------------------------------------------------------------

/** 通用徽章色调，对应原型里的状态药丸（pill）配色。 */
export type BadgeTone = 'brand' | 'emerald' | 'amber' | 'slate' | 'sky' | 'rose'

/** 讲师端 - 课程卡片（原型 iClassroom.html 的 courses）。 */
export interface CourseVM {
  id: string
  title: string
  code: string
  students: number
  assignments: number
  last: string
  cover: string
  summary: string
}

/** 讲师端 - 作业/任务（原型 assignments）。 */
export type AssignmentStatus = 'Active' | 'Closed' | 'Draft'
export interface AssignmentVM {
  id: string
  title: string
  due: string
  count: string
  status: AssignmentStatus
  course: string
}

/** 公告（原型 Announcements）。 */
export interface AnnouncementVM {
  id: string
  title: string
  body: string
}

/** 复习/批改进度条（原型 Review Summary）。 */
export interface ProgressStatVM {
  label: string
  value: number
  total: number
  tone: BadgeTone
}

/** 提交附件文件（原型 submissions[].files）。 */
export interface SubmissionFileVM {
  name: string
  size: string
  type: string
}

/** 提交图片（原型 submissions[].images）。 */
export interface SubmissionImageVM {
  label: string
  src: string
}

/** 讲师端 - 学生提交（原型 submissions）。 */
export type ReviewStatus = 'submitted' | 'pending' | 'reviewed'
export interface SubmissionVM {
  id: string
  name: string
  initials: string
  status: ReviewStatus
  time: string
  score: number
  feedback: string
  response: string
  images: SubmissionImageVM[]
  files: SubmissionFileVM[]
  history: string[]
}

/** 学生端 - 题目（原型 questions）。 */
export interface QuestionVM {
  id: string
  title: string
  prompt: string
}

/** 学生端 - 可选小组（原型 join 屏的 Choose Group）。 */
export interface StudentGroupVM {
  id: string
  name: string
  filled: number
  capacity: number
  full: boolean
}

/** 小组排行（原型 Group Rankings）。 */
export interface RankingVM {
  team: string
  score: number
}

/** 学生端 - 教师反馈（原型 Teacher Feedback - Scored 状态）。 */
export interface FeedbackVM {
  score: number
  max: number
  comment: string
}
