import { useCallback, useEffect, useRef, useState } from 'react'

// 倒计时 hook。迁移自原型的 timer 逻辑（updateTimer / toggleTimer / resetTimer / extendTimer）。
// 讲师批改页手动 start/pause；学生课堂页可在挂载时自动开始。
export function useCountdown(initialSeconds: number, autoStart = false) {
  const [seconds, setSeconds] = useState(initialSeconds)
  const [running, setRunning] = useState(autoStart)
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const clear = useCallback(() => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current)
      intervalRef.current = null
    }
  }, [])

  useEffect(() => {
    if (!running) {
      clear()
      return
    }
    intervalRef.current = setInterval(() => {
      setSeconds((prev) => {
        if (prev <= 1) {
          clear()
          setRunning(false)
          return 0
        }
        return prev - 1
      })
    }, 1000)
    return clear
  }, [running, clear])

  const start = useCallback(() => setRunning(true), [])
  const pause = useCallback(() => setRunning(false), [])
  const toggle = useCallback(() => setRunning((r) => !r), [])
  const reset = useCallback(() => {
    setRunning(false)
    setSeconds(initialSeconds)
  }, [initialSeconds])
  const extend = useCallback((minutes: number) => setSeconds((s) => s + minutes * 60), [])

  const mmss = `${String(Math.floor(seconds / 60)).padStart(2, '0')}:${String(seconds % 60).padStart(2, '0')}`

  return { seconds, running, mmss, start, pause, toggle, reset, extend }
}
