import { useEffect, useState } from 'react'
import { LogIn } from 'lucide-react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import StudentHeader from '../../components/layout/StudentHeader'
import Card from '../../components/ui/Card'
import GroupTable from '../../components/GroupTable'
import { btnGradient } from '../../components/ui/buttons'
import { useToast } from '../../components/ui/useToast'
import { useStudentSession } from '../../hooks/useStudentSession'
import { getStudentRoom, joinStudentRoom, resumeStudentRoom } from '../../api/student'
import type { StudentRoomResponse } from '../../api/student'
import { getStudentSession } from '../../utils/session'

export default function StudentEntry() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { showToast, ToastView } = useToast()
  const { identity, join, clear } = useStudentSession()

  const roomCode = (searchParams.get('room') ?? '').trim().toUpperCase()
  const [roomLookupCode, setRoomLookupCode] = useState('')
  const [room, setRoom] = useState<StudentRoomResponse | null>(null)
  const [nickname, setNickname] = useState(identity?.name ?? '')
  const [groupId, setGroupId] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [joining, setJoining] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    if (!roomCode) {
      setRoom(null)
      setGroupId(null)
      setError(null)
      setLoading(false)
      return () => {
        cancelled = true
      }
    }

    async function loadRoom() {
      setLoading(true)
      setError(null)

      try {
        const data = await getStudentRoom(roomCode)
        if (cancelled) {
          return
        }

        setRoom(data)

        const firstAvailable = data.groups.find((g) => g.available)
        if (firstAvailable) {
          setGroupId(String(firstAvailable.groupId))
        } else {
          setGroupId(null)
        }

        const storedSession = getStudentSession()
        if (storedSession?.clientToken && matchesRoom(storedSession.roomCode, roomCode)) {
          try {
            const resumed = await resumeStudentRoom(roomCode, { studentToken: storedSession.clientToken })
            if (cancelled) {
              return
            }

            join({
              studentId: resumed.studentId,
              roomCode: resumed.roomCode,
              nickname: resumed.nickname,
              groupId: resumed.groupId,
              groupName: resumed.groupName,
              clientToken: storedSession.clientToken,
            })
            showToast('已恢复上次会话')
            navigate('/student/classroom')
          } catch {
            clear()
            if (!cancelled) {
              showToast('会话已过期，请重新加入。')
            }
          }
        }
      } catch (err) {
        const message = getErrorMessage(err, '加载课堂失败')
        setRoom(null)
        setError(message)
        showToast(message)
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    loadRoom()
    return () => {
      cancelled = true
    }
  }, [clear, join, navigate, roomCode, showToast])

  useEffect(() => {
    if (identity?.nickname) {
      setNickname(identity.nickname)
    }
  }, [identity?.nickname])

  const onLookupRoom = () => {
    const code = roomLookupCode.trim().toUpperCase()
    if (!code) {
      const message = '请输入房间码'
      setError(message)
      showToast(message)
      return
    }
    navigate(`/student?room=${encodeURIComponent(code)}`)
  }

  const onJoin = async () => {
    const name = nickname.trim()
    setError(null)

    if (!name) {
      const message = '请输入昵称'
      setError(message)
      showToast(message)
      return
    }

    if (!room || groupId === null) {
      const message = '请选择小组'
      setError(message)
      showToast(message)
      return
    }

    setJoining(true)

    try {
      const res = await joinStudentRoom(room.roomCode, {
        nickname: name,
        groupId: Number(groupId),
      })

      join(res)
      showToast('已加入课堂')
      setTimeout(() => navigate('/student/classroom'), 500)
    } catch (err) {
      const message = getErrorMessage(err, '加入课堂失败')
      setError(message)
      showToast(message)
    } finally {
      setJoining(false)
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-soft p-8 text-sm text-muted dark:bg-slate-950 dark:text-slate-400">
        正在加载课堂...
        <ToastView />
      </div>
    )
  }

  if (!roomCode) {
    return (
      <div className="min-h-screen bg-soft text-ink dark:bg-slate-950 dark:text-slate-100">
        <main className="mx-auto flex min-h-screen max-w-xl flex-col justify-center px-6 py-10 sm:px-8">
          <Card padded className="!p-6">
            <span className="inline-flex h-11 w-11 items-center justify-center rounded-lg bg-brand-600 text-sm font-bold text-white">
              iC
            </span>
            <p className="mt-5 text-sm font-semibold text-brand-600 dark:text-brand-100">学生端</p>
            <h1 className="mt-2 text-2xl font-semibold tracking-normal">加入课堂</h1>
            <p className="mt-3 text-sm leading-6 text-muted dark:text-slate-400">
              请输入老师提供的房间码。
            </p>

            {error && (
              <div className="mt-5 rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm font-semibold text-rose-700 dark:border-rose-500/20 dark:bg-rose-500/10 dark:text-rose-300">
                {error}
              </div>
            )}

            <form
              className="mt-5 grid gap-3 sm:grid-cols-[1fr_auto]"
              onSubmit={(event) => {
                event.preventDefault()
                onLookupRoom()
              }}
            >
              <label className="min-w-0">
                <span className="sr-only">房间码</span>
                <input
                  value={roomLookupCode}
                  onChange={(event) => setRoomLookupCode(event.target.value.toUpperCase())}
                  placeholder="例如 ABC123"
                  className="h-11 w-full rounded-lg border border-line bg-white px-3 text-sm font-semibold tracking-wide outline-none focus:border-brand-500 dark:border-slate-800 dark:bg-slate-950"
                />
              </label>
              <button type="submit" className={`${btnGradient} h-11 px-4 py-0`}>
                <LogIn className="h-4 w-4" />
                查找课堂
              </button>
            </form>
          </Card>
        </main>
        <ToastView />
      </div>
    )
  }

  if (!room) {
    return (
      <div className="min-h-screen bg-soft p-8 text-sm text-muted dark:bg-slate-950 dark:text-slate-400">
        <p>{error ?? '未找到课堂。'}</p>
        <ToastView />
      </div>
    )
  }

  const groups = room.groups.map((g) => ({
    id: String(g.groupId),
    name: g.groupName,
    filled: g.currentCount,
    capacity: g.capacity,
    full: !g.available,
  }))
  const hasGroups = room.groups.length > 0
  const hasAvailableGroup = room.groups.some((g) => g.available)

  const stats = [
    { value: room.groups.length, label: '小组', tone: 'bg-brand-50 text-brand-700 dark:bg-brand-500/10 dark:text-brand-100' },
    { value: roomStatusLabel(room.status), label: '状态', tone: 'bg-violet-50 text-violet-600 dark:bg-violet-500/10 dark:text-violet-300' },
    { value: room.roomCode, label: '房间码', tone: 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300' },
  ]

  return (
    <div className="min-h-screen bg-soft text-ink dark:bg-slate-950 dark:text-slate-100">
      <StudentHeader roomCode={room.roomCode} />

      <main className="mx-auto max-w-[1180px] px-6 py-8 sm:px-8">
        <section className="grid grid-cols-1 gap-8 lg:grid-cols-[1fr_420px]">
          <Card padded className="!p-8">
            <p className="text-sm font-semibold text-brand-600 dark:text-brand-100">加入课堂</p>
            <h1 className="mt-3 text-4xl font-semibold tracking-normal">{room.title}</h1>
            <p className="mt-4 max-w-2xl text-sm leading-7 text-muted dark:text-slate-400">
              输入你的课堂昵称，选择小组后即可进入当前课堂任务。
            </p>

            <div className="mt-8 grid grid-cols-3 gap-4">
              {stats.map((s) => (
                <div key={s.label} className={`rounded-lg p-4 ${s.tone}`}>
                  <p className="text-2xl font-bold">{s.value}</p>
                  <p className="mt-1 text-xs font-semibold">{s.label}</p>
                </div>
              ))}
            </div>
          </Card>

          <Card padded className="!p-6">
            <h2 className="text-lg font-semibold">加入房间</h2>
            <p className="mt-1 text-sm text-muted dark:text-slate-400">无需注册账号。</p>

            {error && (
              <div className="mt-5 rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm font-semibold text-rose-700 dark:border-rose-500/20 dark:bg-rose-500/10 dark:text-rose-300">
                {error}
              </div>
            )}

            <label className="mt-5 block">
              <span className="text-sm font-semibold">昵称</span>
              <input
                value={nickname}
                maxLength={24}
                disabled={joining}
                onChange={(e) => setNickname(e.target.value)}
                className="mt-2 w-full rounded-lg border border-line bg-white px-3 py-2.5 text-sm outline-none focus:border-brand-500 dark:border-slate-800 dark:bg-slate-950"
              />
            </label>

            <div className="mt-5">
              <p className="text-sm font-semibold">选择小组</p>
              <div className="mt-2">
                {hasGroups ? (
                  <GroupTable groups={groups} selectable value={groupId ?? undefined} onChange={setGroupId} />
                ) : (
                  <div className="rounded-lg border border-line bg-slate-50 px-3 py-3 text-sm text-muted dark:border-slate-800 dark:bg-slate-950 dark:text-slate-400">
                    暂无可选小组。
                  </div>
                )}
              </div>
              {hasGroups && !hasAvailableGroup && (
                <p className="mt-2 text-sm font-semibold text-rose-600 dark:text-rose-300">
                  所有小组都已满员。
                </p>
              )}
            </div>

            <button
              className={`${btnGradient} mt-6 w-full disabled:cursor-not-allowed disabled:opacity-60`}
              disabled={joining || !hasAvailableGroup}
              onClick={onJoin}
            >
              <LogIn className="h-4 w-4" />
              {joining ? '正在加入...' : '进入课堂'}
            </button>
          </Card>
        </section>
      </main>

      <ToastView />
    </div>
  )
}

function matchesRoom(storedRoomCode: string, currentRoomCode: string) {
  return !storedRoomCode || storedRoomCode === currentRoomCode
}

function getErrorMessage(error: unknown, fallback: string) {
  if (error instanceof Error && error.message.trim()) {
    return error.message
  }
  return fallback
}

function roomStatusLabel(status: string) {
  switch (status) {
    case 'active':
      return '进行中'
    case 'ended':
      return '已结束'
    case 'created':
      return '已创建'
    default:
      return status
  }
}
