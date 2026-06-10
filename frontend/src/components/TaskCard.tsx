import type { ReactNode } from 'react'
import { CalendarClock, Inbox } from 'lucide-react'
import type { BadgeTone } from '../types'
import Badge from './ui/Badge'

interface TaskCardProps {
  title: string
  status?: { label: string; tone: BadgeTone }
  /** 截止时间文案（讲师作业卡片）。 */
  due?: string
  /** 提交数文案，如 "19 / 24"（讲师作业卡片）。 */
  submitted?: string
  /** 右侧操作按钮组（讲师作业卡片）。 */
  actions?: ReactNode
  /** 紧凑模式：学生端任务列表行（仅标题 + 状态药丸）。 */
  compact?: boolean
  active?: boolean
  onClick?: () => void
}

// 任务 / 作业卡片。复用于：讲师作业管理列表、学生任务进度列表、大屏任务概览。
// 迁移自 iClassroom.html 的 assignment-list 与 student.html 的 taskList。
export default function TaskCard({
  title,
  status,
  due,
  submitted,
  actions,
  compact = false,
  active = false,
  onClick,
}: TaskCardProps) {
  if (compact) {
    return (
      <button
        onClick={onClick}
        className={`w-full rounded-lg border p-3 text-left transition hover:bg-slate-50 dark:hover:bg-slate-800 ${
          active ? 'border-brand-300' : 'border-line dark:border-slate-800'
        }`}
      >
        <div className="flex items-center justify-between gap-3">
          <span className="truncate text-sm font-semibold">{title}</span>
          {status && (
            <Badge tone={status.tone} className="shrink-0">
              {status.label}
            </Badge>
          )}
        </div>
      </button>
    )
  }

  return (
    <article className="rounded-lg border border-line bg-white p-5 shadow-soft dark:border-slate-800 dark:bg-slate-900">
      <div className="flex items-start justify-between gap-5">
        <div className="min-w-0">
          <div className="flex items-center gap-3">
            <h3 className="truncate text-lg font-semibold tracking-normal">{title}</h3>
            {status && (
              <Badge tone={status.tone} ring>
                {status.label}
              </Badge>
            )}
          </div>
          <div className="mt-3 flex flex-wrap items-center gap-5 text-sm text-muted dark:text-slate-400">
            {due && (
              <span className="inline-flex items-center gap-1.5">
                <CalendarClock className="h-4 w-4" />
                {due}
              </span>
            )}
            {submitted && (
              <span className="inline-flex items-center gap-1.5">
                <Inbox className="h-4 w-4" />
                {submitted} submitted
              </span>
            )}
          </div>
        </div>
        {actions && <div className="flex flex-wrap items-center gap-2">{actions}</div>}
      </div>
    </article>
  )
}
