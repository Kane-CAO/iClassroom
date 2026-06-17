import { useState } from 'react'
import { ArrowRight, LogIn, Plus } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import Card from '../components/ui/Card'
import { btnPrimary, btnSecondary } from '../components/ui/buttons'
import { useToast } from '../components/ui/useToast'

// 首页：真实 MVP 入口。老师创建课堂，学生输入房间码加入。
export default function Home() {
  const navigate = useNavigate()
  const { showToast, ToastView } = useToast()
  const [roomCode, setRoomCode] = useState('')

  const joinRoom = () => {
    const code = roomCode.trim().toUpperCase()
    if (!code) {
      showToast('请输入房间码')
      return
    }
    navigate(`/student?room=${encodeURIComponent(code)}`)
  }

  return (
    <div className="min-h-screen bg-canvas text-ink dark:bg-slate-950 dark:text-slate-100">
      <main className="mx-auto flex min-h-screen max-w-[1180px] flex-col justify-center px-6 py-10 sm:px-8">
        <section className="mb-8">
          <span className="inline-flex h-11 w-11 items-center justify-center rounded-lg bg-brand-600 text-sm font-bold text-white">
            iC
          </span>
          <h1 className="mt-5 text-3xl font-semibold tracking-normal sm:text-4xl">
            iClassroom 轻量课堂互动系统
          </h1>
          <p className="mt-3 max-w-2xl text-sm leading-7 text-muted dark:text-slate-400">
            无需账号，老师创建课堂后分享房间码或链接，学生即可加入互动。
          </p>
        </section>

        <section className="grid grid-cols-1 gap-5 lg:grid-cols-2">
          <Card padded className="!p-6">
            <div className="flex h-full flex-col">
              <div>
                <p className="text-sm font-semibold text-brand-600 dark:text-brand-100">老师端</p>
                <h2 className="mt-2 text-xl font-semibold tracking-normal">创建并管理课堂</h2>
                <p className="mt-3 text-sm leading-6 text-muted dark:text-slate-400">
                  创建课堂、发布任务、批改提交、查看数据与大屏。
                </p>
              </div>
              <button
                className={`${btnPrimary} mt-6 w-full sm:w-fit`}
                onClick={() => navigate('/teacher/create-room')}
              >
                <Plus className="h-4 w-4" />
                创建课堂
              </button>
            </div>
          </Card>

          <Card padded className="!p-6">
            <div className="flex h-full flex-col">
              <div>
                <p className="text-sm font-semibold text-brand-600 dark:text-brand-100">学生端</p>
                <h2 className="mt-2 text-xl font-semibold tracking-normal">输入房间码加入</h2>
                <p className="mt-3 text-sm leading-6 text-muted dark:text-slate-400">
                  输入老师提供的房间码，加入课堂并提交答案。
                </p>
              </div>
              <form
                className="mt-6 flex flex-col gap-3 sm:flex-row"
                onSubmit={(event) => {
                  event.preventDefault()
                  joinRoom()
                }}
              >
                <label className="min-w-0 flex-1">
                  <span className="sr-only">房间码</span>
                  <input
                    value={roomCode}
                    onChange={(event) => setRoomCode(event.target.value.toUpperCase())}
                    placeholder="输入房间码"
                    className="h-11 w-full rounded-lg border border-line bg-white px-3 text-sm font-semibold tracking-wide outline-none focus:border-brand-500 dark:border-slate-800 dark:bg-slate-950"
                  />
                </label>
                <button type="submit" className={`${btnSecondary} h-11 shrink-0`}>
                  <LogIn className="h-4 w-4" />
                  加入课堂
                  <ArrowRight className="h-4 w-4" />
                </button>
              </form>
            </div>
          </Card>
        </section>
      </main>
      <ToastView />
    </div>
  )
}
