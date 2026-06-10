// 学生端 mock 数据（迁移自 docs/prototypes/student.html、studentphone.html）。
// 三个学生原型（桌面 / 平板 / 手机）共享同一份数据，因此集中放在这里。
import type { FeedbackVM, QuestionVM, RankingVM, StudentGroupVM } from '../types'

/** 课堂 / 房间信息（原型 join 屏与课堂头部）。 */
export const studentRoom = {
  course: 'Industrial Design Studio',
  assignment: 'Persona Research Sprint',
  teacher: 'Evelyn Chen',
  roomCode: 'FX7K91',
  questionsCount: 3,
  minutesLeft: 42,
  scoreRange: '1-10',
}

/** 计时器初始剩余秒数（原型 42:00）。 */
export const remainingSeconds = 42 * 60

export const questions: QuestionVM[] = [
  {
    id: 'q1',
    title: 'Q1. Define the primary persona',
    prompt:
      'Describe one target learner or user. Include goals, pain points, context, and one direct observation from your research.',
  },
  {
    id: 'q2',
    title: 'Q2. Map one key journey',
    prompt:
      'Explain the user journey from first contact to task completion. Identify the moment where the experience breaks down.',
  },
  {
    id: 'q3',
    title: 'Q3. Propose one improvement',
    prompt:
      'Suggest a focused design improvement. Explain how it solves the pain point and what evidence would prove it works.',
  },
]

/** 可选小组（原型 Choose Group）。 */
export const studentGroups: StudentGroupVM[] = [
  { id: 'indigo', name: 'Team Indigo', filled: 4, capacity: 5, full: false },
  { id: 'violet', name: 'Team Violet', filled: 3, capacity: 5, full: false },
  { id: 'azure', name: 'Team Azure', filled: 5, capacity: 5, full: true },
]

/** 学生默认身份（原型默认昵称 / 小组）。 */
export const defaultStudentIdentity = {
  name: 'Alex Kim',
  team: 'Team Indigo',
}

/** 小组排行：批改前 / 批改后两种状态（原型 rankingsBeforeReview / AfterReview）。 */
export const rankingsBeforeReview: RankingVM[] = [
  { team: 'Team Indigo', score: 82 },
  { team: 'Team Violet', score: 78 },
  { team: 'Team Azure', score: 74 },
]
export const rankingsAfterReview: RankingVM[] = [
  { team: 'Team Indigo', score: 91 },
  { team: 'Team Violet', score: 83 },
  { team: 'Team Azure', score: 76 },
]

/** 评分后教师反馈（原型 Scored 状态固定文案）。 */
export const studentFeedback: FeedbackVM = {
  score: 9,
  max: 10,
  comment:
    'Great structure across all three answers. Your persona has clear goals and pain points. The journey map is strongest; next time, add one direct quote to support the proposed improvement.',
}

/** 我的进度统计（原型 My Progress 的 Completed / Pending 基数）。 */
export const myProgress = {
  baseCompleted: 5,
  reviewedCompleted: 8,
}

/** 单题最多上传图片数（原型 maxImages，仅用于展示限制文案）。 */
export const maxImagesPerQuestion = 3
