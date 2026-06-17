import { Link } from 'react-router-dom'

const ROOM = 'ABC123'

const teacherLinks = [
  { to: '/teacher/rooms/ABC123/dashboard', label: '房间管理', note: '课堂概览' },
  { to: '/teacher/rooms/ABC123/course', label: '课程详情', note: '任务管理' },
  { to: '/teacher/create-room', label: '创建课堂', note: '新建房间' },
  { to: '/teacher/rooms/ABC123/review', label: '提交批改', note: '评分与反馈' },
  { to: '/teacher/rooms/ABC123/display', label: '大屏看板', note: '课堂展示' },
  { to: '/teacher/rooms/ABC123/analytics', label: '数据看板', note: '课堂数据' },
]

const studentLinks = [
  { to: '/student', label: '入场', note: '输入昵称并选组' },
  { to: '/student/classroom', label: '课堂首页', note: '任务与反馈' },
  { to: '/student/tasks/q1', label: '答题', note: '提交答案' },
]

// 首页：入口导航。标注每个页面对应的原型来源，便于人工检查。
export default function Home() {
  return (
    <div className="min-h-screen bg-canvas text-ink dark:bg-slate-950 dark:text-slate-100">
      <section className="mx-auto max-w-3xl px-6 py-12">
        <h1 className="text-3xl font-semibold tracking-normal">iClassroom</h1>
        <p className="mt-1 text-sm text-muted dark:text-slate-400">
          轻量级线上课堂互动平台 · 原型迁移演示（room: {ROOM}）
        </p>

        <h2 className="mt-8 text-sm font-semibold uppercase tracking-wide text-muted">讲师端</h2>
        <nav className="mt-3 grid gap-2">
          {teacherLinks.map((l) => (
            <Link
              key={l.to}
              to={l.to}
              className="flex items-center justify-between rounded-lg border border-line bg-white px-4 py-3 shadow-soft transition hover:border-brand-200 dark:border-slate-800 dark:bg-slate-900"
            >
              <span className="text-sm font-semibold">{l.label}</span>
              <span className="text-xs text-muted dark:text-slate-400">{l.note}</span>
            </Link>
          ))}
        </nav>

        <h2 className="mt-8 text-sm font-semibold uppercase tracking-wide text-muted">学生端</h2>
        <nav className="mt-3 grid gap-2">
          {studentLinks.map((l) => (
            <Link
              key={l.to}
              to={l.to}
              className="flex items-center justify-between rounded-lg border border-line bg-white px-4 py-3 shadow-soft transition hover:border-brand-200 dark:border-slate-800 dark:bg-slate-900"
            >
              <span className="text-sm font-semibold">{l.label}</span>
              <span className="text-xs text-muted dark:text-slate-400">{l.note}</span>
            </Link>
          ))}
        </nav>
      </section>
    </div>
  )
}
