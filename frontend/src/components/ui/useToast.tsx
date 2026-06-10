import { useCallback, useEffect, useRef, useState } from 'react'

// 轻量 toast：迁移自原型的 showToast()。每个页面自带一份，无需全局 Provider。
// 用法：const { showToast, ToastView } = useToast(); ... <ToastView />
export function useToast() {
  const [message, setMessage] = useState<string | null>(null)
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null)

  const showToast = useCallback((text: string) => {
    setMessage(text)
    if (timer.current) clearTimeout(timer.current)
    timer.current = setTimeout(() => setMessage(null), 2200)
  }, [])

  useEffect(() => () => {
    if (timer.current) clearTimeout(timer.current)
  }, [])

  const ToastView = useCallback(
    () =>
      message ? (
        <div className="fixed bottom-6 right-6 z-50 rounded-lg bg-slate-950 px-4 py-3 text-sm font-semibold text-white shadow-soft dark:bg-white dark:text-slate-950">
          {message}
        </div>
      ) : null,
    [message],
  )

  return { showToast, ToastView }
}
