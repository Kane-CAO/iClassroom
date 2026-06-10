import type { ReviewStatus, SubmissionVM } from '../types'
import Badge from './ui/Badge'
import type { BadgeTone } from '../types'

interface SubmissionCardProps {
  submission: SubmissionVM
  active?: boolean
  onClick?: () => void
}

// 批改状态 → 徽章色调。
export const reviewStatusTone: Record<ReviewStatus, BadgeTone> = {
  reviewed: 'emerald',
  pending: 'amber',
  submitted: 'brand',
}

// 学生提交卡片（列表行）。复用于：讲师批改页左侧学生列表。
// 迁移自 iClassroom.html 的 renderStudentList()。
export default function SubmissionCard({ submission, active = false, onClick }: SubmissionCardProps) {
  return (
    <button
      onClick={onClick}
      className={`mb-2 w-full rounded-lg border p-3 text-left transition hover:bg-slate-50 dark:hover:bg-slate-800 ${
        active
          ? 'border-brand-500 bg-brand-50 dark:border-brand-400 dark:bg-brand-500/16'
          : 'border-line dark:border-slate-800'
      }`}
    >
      <div className="flex items-center gap-3">
        <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-slate-900 text-xs font-bold text-white dark:bg-brand-500">
          {submission.initials}
        </span>
        <span className="min-w-0 flex-1">
          <span className="block truncate text-sm font-semibold">{submission.name}</span>
          <span className="block text-xs text-muted dark:text-slate-400">{submission.time}</span>
        </span>
        <Badge tone={reviewStatusTone[submission.status]} className="capitalize">
          {submission.status}
        </Badge>
      </div>
    </button>
  )
}
