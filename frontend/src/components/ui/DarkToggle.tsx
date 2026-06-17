import { useEffect, useState } from 'react'
import { Moon, Sun } from 'lucide-react'
import { getItem, setItem } from '../../utils/storage'

// 暗色模式切换：迁移自原型的 toggleDark()，状态持久化到 localStorage。
// 通过切换 <html> 上的 .dark class 生效（Tailwind darkMode: 'class'）。
const STORAGE_KEY = 'theme:dark'

function applyDark(dark: boolean) {
  document.documentElement.classList.toggle('dark', dark)
}

export default function DarkToggle({ compact = false }: { compact?: boolean }) {
  const [dark, setDark] = useState(false)

  useEffect(() => {
    const saved = getItem<boolean>(STORAGE_KEY) ?? false
    setDark(saved)
    applyDark(saved)
  }, [])

  const toggle = () => {
    const next = !dark
    setDark(next)
    applyDark(next)
    setItem(STORAGE_KEY, next)
  }

  const Icon = dark ? Sun : Moon
  const label = dark ? '浅色' : '深色'

  return (
    <button
      onClick={toggle}
      aria-label="切换深色模式"
      className="flex h-9 items-center gap-2 rounded-lg border border-line bg-white px-3 text-sm font-semibold text-slate-600 transition hover:bg-slate-50 dark:border-slate-800 dark:bg-slate-900 dark:text-slate-300 dark:hover:bg-slate-800"
    >
      <Icon className="h-4 w-4" />
      {!compact && <span>{label}</span>}
    </button>
  )
}
