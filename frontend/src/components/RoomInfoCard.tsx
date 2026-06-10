import type { ReactNode } from 'react'
import type { BadgeTone } from '../types'
import Card from './ui/Card'
import Badge from './ui/Badge'

interface RoomInfoCardProps {
  /** 小标题 / 课程名（原型 brand 色的 eyebrow 文案）。 */
  eyebrow?: string
  /** 主标题（作业 / 任务名）。 */
  title: string
  /** 次要信息行，如 "Teacher: Evelyn Chen · Room FX7K91"。 */
  meta?: string
  status?: { label: string; tone: BadgeTone }
  /** 右侧插槽，常用于计时器。 */
  right?: ReactNode
  titleClassName?: string
}

// 房间 / 课堂信息卡。复用于：学生课堂头部、讲师批改头部、大屏。
// 迁移自 student.html 课堂头部与 iClassroom.html 课程详情头部。
export default function RoomInfoCard({
  eyebrow,
  title,
  meta,
  status,
  right,
  titleClassName = 'text-3xl',
}: RoomInfoCardProps) {
  return (
    <Card padded>
      <div className="flex items-start justify-between gap-6">
        <div className="min-w-0">
          {eyebrow && (
            <p className="text-sm font-semibold text-brand-600 dark:text-brand-100">{eyebrow}</p>
          )}
          <h1 className={`mt-2 font-semibold tracking-normal ${titleClassName}`}>{title}</h1>
          {meta && <p className="mt-2 text-sm leading-6 text-muted dark:text-slate-400">{meta}</p>}
        </div>
        {status && !right && <Badge tone={status.tone}>{status.label}</Badge>}
        {right}
      </div>
    </Card>
  )
}
