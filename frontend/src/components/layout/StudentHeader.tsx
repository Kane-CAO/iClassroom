import DarkToggle from '../ui/DarkToggle'

interface StudentHeaderProps {
  roomCode: string
  connected?: boolean
  /** 已加入后显示的学生信息（昵称 + 小组）。 */
  student?: { name: string; team: string } | null
}

function initials(name: string) {
  return (
    name
      .split(/\s+/)
      .filter(Boolean)
      .slice(0, 2)
      .map((part) => part[0])
      .join('')
      .toUpperCase() || 'ST'
  )
}

// 学生端顶部栏。迁移自 student.html 的 <header>。
export default function StudentHeader({ roomCode, connected = true, student }: StudentHeaderProps) {
  return (
    <header className="border-b border-line bg-white/95 backdrop-blur dark:border-slate-800 dark:bg-slate-950/95">
      <div className="mx-auto flex h-16 max-w-[1180px] items-center justify-between px-6 sm:px-8">
        <div className="flex items-center gap-3">
          <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-gradient-to-br from-brand-600 to-violetx-600 text-sm font-bold text-white">
            iC
          </div>
          <div>
            <p className="text-sm font-semibold leading-4">iClassroom</p>
            <p className="text-xs text-muted dark:text-slate-400">QR invitation · Room {roomCode}</p>
          </div>
        </div>

        <div className="flex items-center gap-3">
          {connected && (
            <span className="rounded-full bg-emerald-50 px-3 py-1.5 text-xs font-bold text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300">
              Connected
            </span>
          )}
          <DarkToggle />
          {student && (
            <div className="flex items-center gap-2 rounded-lg border border-line bg-white py-1.5 pl-2 pr-3 dark:border-slate-800 dark:bg-slate-900">
              <span className="flex h-8 w-8 items-center justify-center rounded-full bg-brand-600 text-xs font-bold text-white">
                {initials(student.name)}
              </span>
              <span>
                <span className="block text-sm font-semibold leading-4">{student.name}</span>
                <span className="block text-xs text-muted dark:text-slate-400">{student.team}</span>
              </span>
            </div>
          )}
        </div>
      </div>
    </header>
  )
}
