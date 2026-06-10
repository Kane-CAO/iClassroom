import type { RankingVM } from '../types'

interface RankingBoardProps {
  rankings: RankingVM[]
  /** 当前学生所属小组名，用于高亮“My Group”。 */
  myTeam?: string
  /** 是否显示分数进度条（手机端 / 大屏使用，桌面端原型不带条）。 */
  withBars?: boolean
}

// 小组排行榜。复用于：学生课堂排行、大屏排行。
// 迁移自 student.html / studentphone.html 的 renderRankings()。
export default function RankingBoard({ rankings, myTeam, withBars = false }: RankingBoardProps) {
  return (
    <div className="space-y-3">
      {rankings.map((row, index) => {
        const mine = row.team === myTeam
        const bg = mine ? 'bg-brand-50 dark:bg-brand-500/10' : 'bg-slate-50 dark:bg-slate-950'
        const text = mine ? 'text-brand-700 dark:text-brand-100' : ''
        return (
          <div key={row.team} className={`rounded-lg ${bg} px-3 py-3`}>
            <div className="flex items-center justify-between">
              <span className={`text-sm font-semibold ${text}`}>
                {index + 1}. {row.team}
                {mine ? ' · My Group' : ''}
              </span>
              <span className={`text-sm font-bold ${text}`}>{row.score}</span>
            </div>
            {withBars && (
              <div className="mt-2 h-2 rounded-full bg-white dark:bg-slate-800">
                <div
                  className="h-full rounded-full bg-gradient-to-r from-brand-600 to-violetx-600"
                  style={{ width: `${row.score}%` }}
                />
              </div>
            )}
          </div>
        )
      })}
    </div>
  )
}
