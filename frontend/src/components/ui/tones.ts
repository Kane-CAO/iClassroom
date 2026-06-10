import type { BadgeTone } from '../../types'

// 徽章 / 药丸（pill）的色调 → Tailwind class 映射。
// 取自 docs/prototypes 中 badgeClass / submissionBadge 等函数的配色，集中管理避免散落。
export const badgeToneClass: Record<BadgeTone, string> = {
  brand:
    'bg-brand-50 text-brand-700 dark:bg-brand-500/10 dark:text-brand-100',
  emerald:
    'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300',
  amber:
    'bg-amber-50 text-amber-700 dark:bg-amber-500/10 dark:text-amber-300',
  slate:
    'bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300',
  sky: 'bg-sky-50 text-sky-700 dark:bg-sky-500/10 dark:text-sky-300',
  rose: 'bg-rose-50 text-rose-700 dark:bg-rose-500/10 dark:text-rose-300',
}

// 进度条填充色（Review Summary / 排行榜）。
export const barToneClass: Record<BadgeTone, string> = {
  brand: 'bg-brand-600',
  emerald: 'bg-emerald-500',
  amber: 'bg-amber-500',
  slate: 'bg-slate-400',
  sky: 'bg-sky-500',
  rose: 'bg-rose-500',
}
