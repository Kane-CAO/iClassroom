// 讲师端 mock 数据（迁移自 docs/prototypes/iClassroom.html 的内联脚本）。
// 页面组件不应再写死这些数据，统一从此处读取。
import type {
  AnnouncementVM,
  AssignmentVM,
  CourseVM,
  ProgressStatVM,
  SubmissionVM,
} from '../types'

/** 当前登录讲师（原型右上角头像）。 */
export const teacherProfile = {
  name: 'Evelyn Chen',
  initials: 'EC',
  role: 'Teacher',
}

export const courses: CourseVM[] = [
  {
    id: 'industrial',
    title: 'Industrial Design Studio',
    code: 'IDS-402',
    students: 24,
    assignments: 6,
    last: 'Today, 10:32',
    cover:
      'https://images.unsplash.com/photo-1518005020951-eccb494ad742?auto=format&fit=crop&w=900&q=80',
    summary: 'Prototype critique, material studies, and studio submissions.',
  },
  {
    id: 'visual',
    title: 'Visual Communication',
    code: 'VC-218',
    students: 31,
    assignments: 8,
    last: 'Yesterday, 16:45',
    cover:
      'https://images.unsplash.com/photo-1497366754035-f200968a6e72?auto=format&fit=crop&w=900&q=80',
    summary: 'Campaign systems, layouts, and design process files.',
  },
  {
    id: 'ux',
    title: 'UX Research Methods',
    code: 'UXR-310',
    students: 28,
    assignments: 5,
    last: 'Mon, 09:12',
    cover:
      'https://images.unsplash.com/photo-1556761175-b413da4baf72?auto=format&fit=crop&w=900&q=80',
    summary: 'Interview plans, synthesis boards, and written findings.',
  },
]

export const assignments: AssignmentVM[] = [
  { id: 'persona', title: 'Persona Research Pack', due: 'Jun 18, 2026 17:00', count: '19 / 24', status: 'Active', course: 'industrial' },
  { id: 'joint', title: 'Material Joint Prototype', due: 'Jun 21, 2026 14:00', count: '7 / 24', status: 'Active', course: 'industrial' },
  { id: 'ergo', title: 'Ergonomic Observation Notes', due: 'Jun 12, 2026 18:00', count: '24 / 24', status: 'Closed', course: 'industrial' },
  { id: 'memo', title: 'Research Memo Draft', due: 'Jun 24, 2026 12:00', count: '0 / 24', status: 'Draft', course: 'industrial' },
]

export const announcements: AnnouncementVM[] = [
  { id: 'a1', title: 'Studio critique room changed', body: 'Industrial Design Studio meets in B214 on Friday.' },
  { id: 'a2', title: 'Review PDFs available', body: 'Published feedback is now downloadable for Assignment 2.' },
  { id: 'a3', title: 'Office hours posted', body: 'Wednesday 14:00-16:00 for UX Research Methods.' },
]

/** Review Summary 进度条（原型固定百分比）。 */
export const reviewSummary: ProgressStatVM[] = [
  { label: 'Submitted', value: 19, total: 24, tone: 'brand' },
  { label: 'Reviewed', value: 12, total: 19, tone: 'emerald' },
  { label: 'Feedback Published', value: 8, total: 19, tone: 'sky' },
]

/** 迷你日历（原型 June 2026，标记的事件日期 / 今天）。 */
export const calendar = {
  monthLabel: 'June 2026',
  daysInMonth: 30,
  startOffset: 1, // 6/1 是周一 → 第一格留空 1 个
  today: 9,
  eventDays: [9, 12, 18, 21, 24],
}

export const submissions: SubmissionVM[] = [
  {
    id: 'alex',
    name: 'Alex Kim',
    initials: 'AK',
    status: 'reviewed',
    time: 'Jun 9, 2026 10:14',
    score: 8,
    feedback:
      'Strong synthesis and clear persona framing. Add one more direct quote to support the accessibility pain point.',
    response:
      'My persona, Sarah, is a 28-year-old UX researcher who struggles with information overload in data-heavy review sessions. The uploaded journey map shows where navigation friction and weak system feedback slow her down.',
    images: [
      { label: 'Persona board', src: 'https://images.unsplash.com/photo-1552664730-d307ca884978?auto=format&fit=crop&w=800&q=80' },
      { label: 'Journey map', src: 'https://images.unsplash.com/photo-1542744095-291d1f67b221?auto=format&fit=crop&w=800&q=80' },
    ],
    files: [
      { name: 'Interview-notes.pdf', size: '2.4 MB', type: 'PDF' },
      { name: 'Persona-board.fig', size: '8.1 MB', type: 'FIG' },
    ],
    history: ['Score 7 saved on Jun 8', 'Annotated PDF uploaded on Jun 9', 'Feedback published on Jun 9'],
  },
  {
    id: 'maya',
    name: 'Maya Rodriguez',
    initials: 'MR',
    status: 'submitted',
    time: 'Jun 9, 2026 10:41',
    score: 7,
    feedback: '',
    response:
      'I developed two personas: Marcus, a time-constrained product manager, and Jin, a student with low technical confidence. The contrast suggests two onboarding paths and contextual help for complex features.',
    images: [
      { label: 'Persona A', src: 'https://images.unsplash.com/photo-1519389950473-47ba0277781c?auto=format&fit=crop&w=800&q=80' },
      { label: 'Persona B', src: 'https://images.unsplash.com/photo-1557426272-fc759fdf7a8d?auto=format&fit=crop&w=800&q=80' },
      { label: 'Notes wall', src: 'https://images.unsplash.com/photo-1531482615713-2afd69097998?auto=format&fit=crop&w=800&q=80' },
    ],
    files: [
      { name: 'Persona-analysis.docx', size: '1.1 MB', type: 'DOCX' },
      { name: 'Interview-audio.zip', size: '46 MB', type: 'ZIP' },
    ],
    history: ['Submission received on Jun 9'],
  },
  {
    id: 'theo',
    name: 'Theo Lin',
    initials: 'TL',
    status: 'submitted',
    time: 'Jun 9, 2026 11:03',
    score: 9,
    feedback: '',
    response:
      'Three interviews point to Elena, a 41-year-old operations director who needs rapid decision support. Her main issues are notification fatigue, weak data trust, and unclear export states.',
    images: [
      { label: 'Interview synthesis', src: 'https://images.unsplash.com/photo-1522202176988-66273c2fd55f?auto=format&fit=crop&w=800&q=80' },
      { label: 'Persona card', src: 'https://images.unsplash.com/photo-1500530855697-b586d89ba3ee?auto=format&fit=crop&w=800&q=80' },
    ],
    files: [
      { name: 'Research-pack.pdf', size: '5.7 MB', type: 'PDF' },
      { name: 'Raw-transcripts.docx', size: '3.5 MB', type: 'DOCX' },
    ],
    history: ['Submission received on Jun 9'],
  },
  {
    id: 'sofia',
    name: 'Sofia Park',
    initials: 'SP',
    status: 'pending',
    time: 'Not submitted',
    score: 0,
    feedback: '',
    response: '',
    images: [],
    files: [],
    history: ['Reminder sent on Jun 8'],
  },
  {
    id: 'nora',
    name: 'Nora Patel',
    initials: 'NP',
    status: 'reviewed',
    time: 'Jun 8, 2026 16:28',
    score: 9,
    feedback:
      'Excellent interview evidence. The annotated PDF explains next steps for tightening your findings.',
    response:
      'I focused on remote design students and how they organize project references before critique. The persona highlights file naming, visual overload, and feedback timing.',
    images: [
      { label: 'Affinity map', src: 'https://images.unsplash.com/photo-1517245386807-bb43f82c33c4?auto=format&fit=crop&w=800&q=80' },
    ],
    files: [
      { name: 'Annotated-review.pdf', size: '1.8 MB', type: 'PDF' },
      { name: 'Source-research.zip', size: '22 MB', type: 'ZIP' },
    ],
    history: ['Score 9 saved on Jun 8', 'Review PDF uploaded on Jun 8', 'Feedback published on Jun 8'],
  },
]

/** 当前批改的作业标题（原型 Submission Review 顶部）。 */
export const reviewAssignmentTitle = 'Persona Research Pack'

/** 课堂计时器初始值（秒），原型 45:00。 */
export const classTimerSeconds = 45 * 60
