import { Bell } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import DarkToggle from '../ui/DarkToggle'
import { teacherMocks } from '../../mocks'

interface TeacherHeaderProps {
  roomCode: string
  /** 当前高亮的导航 tab。 */
  active?: 'courses' | 'assignments'
}

// 讲师端顶部导航。迁移自 iClassroom.html 的 <header>。
export default function TeacherHeader({ roomCode, active = 'courses' }: TeacherHeaderProps) {
  const navigate = useNavigate()
  const profile = teacherMocks.teacherProfile

  const tabClass = (tab: 'courses' | 'assignments') =>
    `rounded-md px-5 py-2 text-sm font-semibold transition ${
      active === tab
        ? 'bg-brand-50 text-brand-700 shadow-[inset_0_0_0_1px_rgba(37,99,235,.16)] dark:bg-brand-500/14 dark:text-brand-100'
        : 'text-slate-600 dark:text-slate-300'
    }`

  return (
    <header className="sticky top-0 z-40 border-b border-line bg-white/95 backdrop-blur dark:border-slate-800 dark:bg-slate-950/90">
      <div className="flex h-16 items-center px-8">
        <button
          className="flex items-center gap-3"
          onClick={() => navigate(`/teacher/rooms/${roomCode}/dashboard`)}
        >
          <span className="flex h-9 w-9 items-center justify-center rounded-lg bg-brand-600 text-sm font-bold text-white">
            iC
          </span>
          <span className="text-lg font-semibold tracking-normal">iClassroom</span>
        </button>

        <nav className="mx-auto flex items-center gap-2 rounded-lg border border-line bg-slate-50 p-1 dark:border-slate-800 dark:bg-slate-900">
          <button
            className={tabClass('courses')}
            onClick={() => navigate(`/teacher/rooms/${roomCode}/dashboard`)}
          >
            Courses
          </button>
          <button
            className={tabClass('assignments')}
            onClick={() => navigate(`/teacher/rooms/${roomCode}/course`)}
          >
            Assignments
          </button>
        </nav>

        <div className="flex items-center gap-3">
          <button
            className="relative flex h-9 w-9 items-center justify-center rounded-lg border border-line bg-white text-slate-600 transition hover:bg-slate-50 dark:border-slate-800 dark:bg-slate-900 dark:text-slate-300 dark:hover:bg-slate-800"
            aria-label="Notifications"
          >
            <Bell className="h-4 w-4" />
            <span className="absolute right-2 top-2 h-2 w-2 rounded-full bg-rose-500 ring-2 ring-white dark:ring-slate-900" />
          </button>
          <DarkToggle />
          <button className="flex items-center gap-3 rounded-lg border border-line bg-white py-1.5 pl-2 pr-3 transition hover:bg-slate-50 dark:border-slate-800 dark:bg-slate-900 dark:hover:bg-slate-800">
            <span className="flex h-8 w-8 items-center justify-center rounded-full bg-slate-900 text-xs font-bold text-white dark:bg-brand-500">
              {profile.initials}
            </span>
            <span className="text-left">
              <span className="block text-sm font-semibold leading-4">{profile.name}</span>
              <span className="block text-xs text-muted dark:text-slate-400">{profile.role}</span>
            </span>
          </button>
        </div>
      </div>
    </header>
  )
}
