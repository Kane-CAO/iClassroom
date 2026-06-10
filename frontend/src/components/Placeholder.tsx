import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'

interface PlaceholderProps {
  title: string
  description?: string
  children?: ReactNode
}

// 统一的占位页面。各业务页面在实现前先复用它展示路由是否生效。
export default function Placeholder({ title, description, children }: PlaceholderProps) {
  return (
    <section style={{ maxWidth: 720, margin: '0 auto', padding: 24 }}>
      <h1 style={{ marginBottom: 8 }}>{title}</h1>
      {description && <p style={{ color: '#6b7280', marginTop: 0 }}>{description}</p>}
      <p style={{ color: '#9ca3af' }}>占位页面 · 业务逻辑待实现</p>
      {children}
      <p style={{ marginTop: 24 }}>
        <Link to="/">← 返回首页</Link>
      </p>
    </section>
  )
}
