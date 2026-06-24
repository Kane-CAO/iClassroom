import { useEffect, useState } from 'react'
import { KeyRound, Plus, Power, Trash2 } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import {
  createTeacherAccount,
  deleteTeacherAccount,
  listTeacherAccounts,
  resetTeacherPassword,
  updateTeacherStatus,
} from '../../api/auth'
import Card from '../../components/ui/Card'
import Badge from '../../components/ui/Badge'
import { btnPrimary, btnSecondary } from '../../components/ui/buttons'
import { useToast } from '../../components/ui/useToast'
import { clearAdminSession, getAdminSession } from '../../utils/session'
import type { TeacherAccount } from '../../types/api'

const inputClass =
  'mt-2 w-full rounded-lg border border-line bg-white px-3 py-2.5 text-sm outline-none focus:border-brand-500 dark:border-slate-800 dark:bg-slate-950'

export default function TeacherAccounts() {
  const navigate = useNavigate()
  const { showToast, ToastView } = useToast()
  const session = getAdminSession()
  const [teachers, setTeachers] = useState<TeacherAccount[]>([])
  const [username, setUsername] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [password, setPassword] = useState('ChangeMe123')
  const [temporaryPassword, setTemporaryPassword] = useState('')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const token = session?.token ?? ''

  useEffect(() => {
    if (!token) {
      navigate('/admin/login')
      return
    }
    void refresh()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token])

  async function refresh() {
    setLoading(true)
    setError('')
    try {
      setTeachers(await listTeacherAccounts(token))
    } catch (err) {
      const message = err instanceof Error ? err.message : '加载讲师列表失败'
      setError(message)
      if (message.includes('session') || message.includes('登录') || message.includes('Unauthorized')) {
        clearAdminSession()
        navigate('/admin/login')
      }
    } finally {
      setLoading(false)
    }
  }

  async function createTeacher() {
    if (!username.trim() || !password.trim()) {
      setError('账号和初始密码必填')
      return
    }
    setSaving(true)
    setError('')
    try {
      await createTeacherAccount(
        {
          username: username.trim(),
          displayName: displayName.trim(),
          initialPassword: password,
        },
        token,
      )
      setUsername('')
      setDisplayName('')
      setPassword('ChangeMe123')
      showToast('讲师已创建')
      await refresh()
    } catch (err) {
      const message = err instanceof Error ? err.message : '创建讲师失败'
      setError(message)
      showToast(message)
    } finally {
      setSaving(false)
    }
  }

  async function toggleStatus(teacher: TeacherAccount) {
    const next = teacher.status === 'active' ? 'disabled' : 'active'
    try {
      await updateTeacherStatus(teacher.teacherId, { status: next }, token)
      showToast(next === 'active' ? '讲师已启用' : '讲师已停用')
      await refresh()
    } catch (err) {
      showToast(err instanceof Error ? err.message : '更新状态失败')
    }
  }

  async function resetPassword(teacher: TeacherAccount) {
    try {
      const res = await resetTeacherPassword(teacher.teacherId, {}, token)
      setTemporaryPassword(`${teacher.username}: ${res.temporaryPassword}`)
      showToast('密码已重置')
    } catch (err) {
      showToast(err instanceof Error ? err.message : '重置密码失败')
    }
  }

  async function removeTeacher(teacher: TeacherAccount) {
    if (!window.confirm(`确定删除讲师 ${teacher.username} 吗？`)) {
      return
    }
    try {
      await deleteTeacherAccount(teacher.teacherId, token)
      showToast('讲师已删除')
      await refresh()
    } catch (err) {
      showToast(err instanceof Error ? err.message : '删除讲师失败')
    }
  }

  return (
    <div className="min-h-screen bg-canvas px-8 py-7 text-ink dark:bg-slate-950 dark:text-slate-100">
      <header className="mb-6 flex flex-wrap items-end justify-between gap-4">
        <div>
          <p className="text-sm font-semibold text-brand-600 dark:text-brand-100">管理者端</p>
          <h1 className="mt-2 text-2xl font-semibold tracking-normal">讲师账号管理</h1>
        </div>
        <button
          className={btnSecondary}
          onClick={() => {
            clearAdminSession()
            navigate('/admin/login')
          }}
        >
          退出登录
        </button>
      </header>

      {error && (
        <Card className="mb-5 border-rose-200 bg-rose-50 p-4 text-sm font-semibold text-rose-700 dark:border-rose-500/20 dark:bg-rose-500/10 dark:text-rose-300">
          {error}
        </Card>
      )}
      {temporaryPassword && (
        <Card className="mb-5 border-amber-200 bg-amber-50 p-4 text-sm font-semibold text-amber-800 dark:border-amber-500/20 dark:bg-amber-500/10 dark:text-amber-300">
          临时密码：{temporaryPassword}
        </Card>
      )}

      <div className="grid gap-6 lg:grid-cols-[360px_1fr]">
        <Card padded className="!p-5">
          <h2 className="text-lg font-semibold tracking-normal">创建讲师</h2>
          <form
            className="mt-5 space-y-4"
            onSubmit={(event) => {
              event.preventDefault()
              createTeacher()
            }}
          >
            <label className="block">
              <span className="text-sm font-semibold">账号</span>
              <input className={inputClass} value={username} onChange={(event) => setUsername(event.target.value)} />
            </label>
            <label className="block">
              <span className="text-sm font-semibold">显示名</span>
              <input
                className={inputClass}
                value={displayName}
                onChange={(event) => setDisplayName(event.target.value)}
              />
            </label>
            <label className="block">
              <span className="text-sm font-semibold">初始密码</span>
              <input className={inputClass} value={password} onChange={(event) => setPassword(event.target.value)} />
            </label>
            <button className={`${btnPrimary} w-full justify-center`} disabled={saving}>
              <Plus className="h-4 w-4" />
              {saving ? '创建中...' : '创建讲师'}
            </button>
          </form>
        </Card>

        <Card>
          <div className="border-b border-line px-5 py-4 dark:border-slate-800">
            <h2 className="text-lg font-semibold tracking-normal">讲师列表</h2>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full min-w-[760px] text-left text-sm">
              <thead className="bg-slate-50 text-xs font-bold uppercase text-muted dark:bg-slate-950 dark:text-slate-400">
                <tr>
                  <th className="px-5 py-3">账号</th>
                  <th className="px-5 py-3">显示名</th>
                  <th className="px-5 py-3">状态</th>
                  <th className="px-5 py-3">最近登录</th>
                  <th className="px-5 py-3">操作</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-line dark:divide-slate-800">
                {loading && (
                  <tr>
                    <td className="px-5 py-5 text-muted dark:text-slate-400" colSpan={5}>
                      正在加载...
                    </td>
                  </tr>
                )}
                {!loading && teachers.length === 0 && (
                  <tr>
                    <td className="px-5 py-5 text-muted dark:text-slate-400" colSpan={5}>
                      暂无讲师账号。
                    </td>
                  </tr>
                )}
                {teachers.map((teacher) => (
                  <tr key={teacher.teacherId}>
                    <td className="px-5 py-4 font-semibold">{teacher.username}</td>
                    <td className="px-5 py-4">{teacher.displayName}</td>
                    <td className="px-5 py-4">
                      <Badge tone={teacher.status === 'active' ? 'emerald' : 'slate'}>
                        {teacher.status === 'active' ? '启用' : '停用'}
                      </Badge>
                    </td>
                    <td className="px-5 py-4 text-muted dark:text-slate-400">
                      {teacher.lastLoginAt ? formatTime(teacher.lastLoginAt) : '未登录'}
                    </td>
                    <td className="px-5 py-4">
                      <div className="flex flex-wrap gap-2">
                        <button className={btnSecondary} onClick={() => toggleStatus(teacher)}>
                          <Power className="h-4 w-4" />
                          {teacher.status === 'active' ? '停用' : '启用'}
                        </button>
                        <button className={btnSecondary} onClick={() => resetPassword(teacher)}>
                          <KeyRound className="h-4 w-4" />
                          重置
                        </button>
                        <button className={btnSecondary} onClick={() => removeTeacher(teacher)}>
                          <Trash2 className="h-4 w-4" />
                          删除
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      </div>
      <ToastView />
    </div>
  )
}

function formatTime(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return date.toLocaleString('zh-CN', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}
