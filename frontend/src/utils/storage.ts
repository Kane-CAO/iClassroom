// localStorage 读写封装（对齐 CLAUDE.md：统一前缀、按需隔离）。
const PREFIX = 'iclassroom:'

export function setItem<T>(key: string, value: T): void {
  localStorage.setItem(PREFIX + key, JSON.stringify(value))
}

export function getItem<T>(key: string): T | null {
  const raw = localStorage.getItem(PREFIX + key)
  if (raw === null) return null
  try {
    return JSON.parse(raw) as T
  } catch {
    return null
  }
}

export function removeItem(key: string): void {
  localStorage.removeItem(PREFIX + key)
}
