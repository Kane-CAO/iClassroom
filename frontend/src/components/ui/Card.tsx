import type { HTMLAttributes } from 'react'

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  /** 是否带内边距（部分卡片内部自带分区，不需要统一 padding）。 */
  padded?: boolean
}

// 通用卡片容器：白底 / 描边 / soft 阴影 + dark 变体。
// 对应原型里反复出现的 `rounded-lg border border-line bg-white shadow-soft ...`。
export default function Card({ padded = false, className = '', children, ...rest }: CardProps) {
  return (
    <div
      className={`rounded-lg border border-line bg-white shadow-soft dark:border-slate-800 dark:bg-slate-900 ${
        padded ? 'p-5' : ''
      } ${className}`}
      {...rest}
    >
      {children}
    </div>
  )
}
