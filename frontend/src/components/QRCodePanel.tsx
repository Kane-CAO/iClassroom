import { QrCode } from 'lucide-react'
import Card from './ui/Card'

interface QRCodePanelProps {
  roomCode: string
  /** 学生加入链接（由调用方拼接，通常来自 VITE_STUDENT_BASE_URL）。 */
  joinUrl: string
  title?: string
  subtitle?: string
}

const GRID = 21

// 根据字符串生成确定性的伪二维码点阵（纯展示，不是可扫描的真实二维码）。
// 当前阶段不引入二维码库，仅作视觉占位；真实二维码待联调阶段接入。
function patternFor(seed: string): boolean[] {
  const cells: boolean[] = []
  let hash = 2166136261
  for (let i = 0; i < seed.length; i++) {
    hash ^= seed.charCodeAt(i)
    hash = Math.imul(hash, 16777619)
  }
  for (let i = 0; i < GRID * GRID; i++) {
    hash ^= hash << 13
    hash ^= hash >>> 17
    hash ^= hash << 5
    cells.push((hash & 1) === 1)
  }
  return cells
}

// 二维码 / 加入面板。复用于：大屏投屏、讲师“Show Join QR”。
// 注意：本阶段不实现真实二维码生成，渲染的是确定性点阵占位图。
export default function QRCodePanel({
  roomCode,
  joinUrl,
  title = '扫码加入',
  subtitle = '无需账号',
}: QRCodePanelProps) {
  const cells = patternFor(`${roomCode}|${joinUrl}`)

  return (
    <Card padded className="text-center">
      <div className="flex items-center justify-center gap-2 text-sm font-semibold text-muted dark:text-slate-400">
        <QrCode className="h-4 w-4" />
        {title}
      </div>

      <div className="mx-auto mt-4 w-fit rounded-xl border border-line bg-white p-3 dark:border-slate-800 dark:bg-slate-950">
        <div
          className="grid gap-px"
          style={{ gridTemplateColumns: `repeat(${GRID}, 1fr)`, width: 168, height: 168 }}
          aria-label={`房间 ${roomCode} 的加入二维码占位图`}
        >
          {cells.map((on, i) => (
            <span
              key={i}
              className={on ? 'rounded-[1px] bg-slate-900 dark:bg-slate-100' : ''}
            />
          ))}
        </div>
      </div>

      <p className="mt-4 text-xs font-semibold uppercase tracking-wide text-muted dark:text-slate-400">
        房间码
      </p>
      <p className="mt-1 font-mono text-2xl font-bold tracking-widest text-brand-700 dark:text-brand-100">
        {roomCode}
      </p>
      <p className="mt-2 break-all text-xs text-muted dark:text-slate-500">{joinUrl}</p>
      <p className="mt-3 text-xs text-muted dark:text-slate-400">{subtitle}</p>
    </Card>
  )
}
