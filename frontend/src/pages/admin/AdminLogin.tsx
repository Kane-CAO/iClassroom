import { useState } from 'react'
import { LogIn } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { loginAdmin } from '../../api/auth'
import Card from '../../components/ui/Card'
import { btnPrimary } from '../../components/ui/buttons'
import { useToast } from '../../components/ui/useToast'
import { setAdminSession } from '../../utils/session'

const inputClass =
  'mt-2 w-full rounded-lg border border-line bg-white px-3 py-2.5 text-sm outline-none focus:border-brand-500 dark:border-slate-800 dark:bg-slate-950'

export default function AdminLogin() {
  const navigate = useNavigate()
  const { showToast, ToastView } = useToast()
  const [username, setUsername] = useState('admin')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const submit = async () => {
    setLoading(true)
    setError('')
    try {
      const res = await loginAdmin({ username: username.trim(), password })
      setAdminSession({ token: res.token, user: res.user })
      showToast('管理员登录成功')
      navigate('/admin/teachers')
    } catch (err) {
      const message = err instanceof Error ? err.message : '登录失败'
      setError(message)
      showToast(message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-canvas px-6 py-10 text-ink dark:bg-slate-950 dark:text-slate-100">
      <main className="mx-auto max-w-md">
        <Card padded className="!p-6">
          <p className="text-sm font-semibold text-brand-600 dark:text-brand-100">管理者端</p>
          <h1 className="mt-2 text-2xl font-semibold tracking-normal">管理员登录</h1>
          {error && (
            <div className="mt-5 rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm font-semibold text-rose-700 dark:border-rose-500/20 dark:bg-rose-500/10 dark:text-rose-300">
              {error}
            </div>
          )}
          <form
            className="mt-5 space-y-4"
            onSubmit={(event) => {
              event.preventDefault()
              submit()
            }}
          >
            <label className="block">
              <span className="text-sm font-semibold">账号</span>
              <input className={inputClass} value={username} onChange={(event) => setUsername(event.target.value)} />
            </label>
            <label className="block">
              <span className="text-sm font-semibold">密码</span>
              <input
                className={inputClass}
                type="password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
              />
            </label>
            <button className={`${btnPrimary} w-full justify-center`} disabled={loading}>
              <LogIn className="h-4 w-4" />
              {loading ? '登录中...' : '登录'}
            </button>
          </form>
        </Card>
      </main>
      <ToastView />
    </div>
  )
}
