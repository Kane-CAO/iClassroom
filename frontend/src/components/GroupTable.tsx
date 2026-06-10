import type { StudentGroupVM } from '../types'

interface GroupTableProps {
  groups: StudentGroupVM[]
  /** 是否可选（学生 join 屏可选；大屏 / 只读场景为 false）。 */
  selectable?: boolean
  /** 选中的小组 id（受控）。 */
  value?: string
  onChange?: (groupId: string) => void
}

// 小组列表 / 选择。复用于：学生 join 屏的 Choose Group、大屏小组概览。
// 迁移自 student.html / studentphone.html 的 Choose Group 区块。
export default function GroupTable({ groups, selectable = false, value, onChange }: GroupTableProps) {
  return (
    <div className="grid gap-2">
      {groups.map((group) => {
        const selected = selectable && value === group.id
        const disabled = group.full
        const base =
          'flex items-center justify-between rounded-lg border px-3 py-3 transition'
        const tone = disabled
          ? 'cursor-not-allowed border-line bg-slate-50 opacity-60 dark:border-slate-800 dark:bg-slate-950'
          : selected
            ? 'cursor-pointer border-brand-200 bg-brand-50 dark:border-brand-500/20 dark:bg-brand-500/10'
            : 'cursor-pointer border-line hover:bg-slate-50 dark:border-slate-800 dark:hover:bg-slate-800'

        const content = (
          <>
            <span>
              <span className="block text-sm font-semibold">{group.name}</span>
              <span className="block text-xs text-muted dark:text-slate-400">
                {disabled ? 'Full' : `${group.filled} / ${group.capacity} students`}
              </span>
            </span>
            {selectable && (
              <input
                type="radio"
                name="group"
                value={group.id}
                checked={selected}
                disabled={disabled}
                onChange={() => onChange?.(group.id)}
              />
            )}
          </>
        )

        if (!selectable) {
          return (
            <div key={group.id} className={`${base} ${tone}`}>
              {content}
            </div>
          )
        }

        return (
          <label key={group.id} className={`${base} ${tone}`}>
            {content}
          </label>
        )
      })}
    </div>
  )
}
