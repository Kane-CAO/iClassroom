import type { BadgeTone } from '../../types'
import { badgeToneClass } from './tones'

interface BadgeProps {
  tone?: BadgeTone
  /** 是否带 ring 描边（讲师作业卡片的状态药丸使用）。 */
  ring?: boolean
  className?: string
  children: React.ReactNode
}

// 状态药丸（pill）。集中复用原型里的圆角 + 配色。
export default function Badge({ tone = 'slate', ring = false, className = '', children }: BadgeProps) {
  return (
    <span
      className={`inline-flex items-center rounded-full px-2.5 py-1 text-xs font-bold ${
        badgeToneClass[tone]
      } ${ring ? 'ring-1 ring-inset ring-slate-200/70 dark:ring-slate-700/70' : ''} ${className}`}
    >
      {children}
    </span>
  )
}
