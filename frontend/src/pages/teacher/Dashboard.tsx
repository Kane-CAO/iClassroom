import { useCallback, useEffect, useState } from 'react'
import { ChevronLeft, ChevronRight, Plus } from 'lucide-react'
import { useNavigate, useParams } from 'react-router-dom'
import TeacherHeader from '../../components/layout/TeacherHeader'
import TeacherRoomActions from '../../components/teacher/TeacherRoomActions'
import Card from '../../components/ui/Card'
import { btnPrimary } from '../../components/ui/buttons'
import { useToast } from '../../components/ui/useToast'
import { useRoomWebSocket } from '../../hooks/useRoomWebSocket'
import { useTeacherSession } from '../../hooks/useTeacherSession'
import { teacherMocks } from '../../mocks'
import { ApiRequestError } from '../../api/client'
import { getRoom, getRoomOverview } from '../../api/rooms'
import { listTeacherTasks } from '../../api/tasks'
import type { RoomWebSocketEventType } from '../../api/websocket'
import type { Room, RoomOverview, Task } from '../../types/api'

const DASHBOARD_WS_EVENTS: readonly RoomWebSocketEventType[] = [
  'student_joined',
  'task_published',
  'submission_created',
  'score_updated',
  'ranking_updated',
  'room_ended',
]

// /teacher/rooms/:roomCode/dashboard
// 迁移自 docs/prototypes/iClassroom.html 的 #page-home（My Courses 概览）。
export default function Dashboard() {
  const { roomCode = '' } = useParams()
  const navigate = useNavigate()
  const { showToast, ToastView } = useToast()
  const { teacherToken, clear } = useTeacherSession()
  const { courses, announcements, calendar } = teacherMocks
  const fallbackCourse = courses[0]
  const [room, setRoom] = useState<Room | null>(null)
  const [overview, setOverview] = useState<RoomOverview | null>(null)
  const [tasks, setTasks] = useState<Task[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [refreshVersion, setRefreshVersion] = useState(0)

  const refreshDashboardData = useCallback(() => {
    setRefreshVersion((version) => version + 1)
  }, [])

  const handleRoomActionError = useCallback((message: string) => {
    setError(message)
    showToast(message)
  }, [showToast])

  const handleRoomActionSuccess = useCallback((message: string) => {
    showToast(message)
  }, [showToast])

  useEffect(() => {
    let cancelled = false

    async function loadDashboardData() {
      setLoading(true)
      setError(null)

      if (!teacherToken) {
        setRoom(null)
        setOverview(null)
        setTasks([])
        setError('老师会话缺失，请重新创建课堂。')
        setLoading(false)
        return
      }

      try {
        const auth = { teacherToken }
        const [roomData, overviewData, taskData] = await Promise.all([
          getRoom(roomCode, auth),
          getRoomOverview(roomCode, auth),
          listTeacherTasks(roomCode, auth),
        ])

        if (cancelled) {
          return
        }

        setRoom(roomData)
        setOverview(overviewData)
        setTasks(taskData)
      } catch (err) {
        if (cancelled) {
          return
        }
        if (isAuthError(err)) {
          clear()
          setError('老师凭证无效或已过期，请重新创建课堂。')
        } else {
          setError(getErrorMessage(err, '加载课堂概览失败。'))
        }
        setRoom(null)
        setOverview(null)
        setTasks([])
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    loadDashboardData()
    return () => {
      cancelled = true
    }
  }, [clear, refreshVersion, roomCode, teacherToken])

  useRoomWebSocket({
    roomCode,
    role: 'teacher',
    onEvent: (event) => {
      if (DASHBOARD_WS_EVENTS.includes(event.type)) {
        refreshDashboardData()
      }
    },
    onReconnect: refreshDashboardData,
  })

  const courseTitle = room?.title ?? overview?.title ?? fallbackCourse.title
  const studentCount = overview?.studentCount ?? 0
  const groupCount = overview?.groups.length ?? room?.groupCount ?? 0
  const taskCount = tasks.length
  const roomStatus = room?.status ?? overview?.status ?? 'created'
  const isRoomEnded = roomStatus === 'ended'

  return (
    <div className="min-h-screen bg-canvas text-ink dark:bg-slate-950 dark:text-slate-100">
      <TeacherHeader roomCode={roomCode} active="dashboard" />

      <main className="px-8 py-7">
        <div className="grid grid-cols-1 gap-6 lg:grid-cols-[1fr_340px]">
          <div>
            <div className="mb-5 flex items-end justify-between">
              <div>
                <h1 className="text-2xl font-semibold tracking-normal">我的课堂</h1>
                <p className="mt-1 text-sm text-muted dark:text-slate-400">
                  {loading ? '正在加载课堂工作区...' : `房间 ${roomCode} · ${roomStatusLabel(roomStatus)}`}
                </p>
              </div>
              <div className="flex flex-wrap justify-end gap-2">
                <button
                  className={btnPrimary}
                  disabled={isRoomEnded}
                  onClick={() => {
                    if (isRoomEnded) {
                      showToast('课堂已结束，不能再创建新任务。')
                      return
                    }
                    navigate(`/teacher/rooms/${roomCode}/course`)
                  }}
                >
                  <Plus className="h-4 w-4" />
                  创建任务
                </button>
                <TeacherRoomActions
                  roomCode={roomCode}
                  teacherToken={teacherToken}
                  roomEnded={isRoomEnded}
                  onEnded={refreshDashboardData}
                  onError={handleRoomActionError}
                  onSuccess={handleRoomActionSuccess}
                />
              </div>
            </div>

            {error && (
              <Card className="mb-5 border-rose-200 bg-rose-50 p-4 text-rose-700 dark:border-rose-500/20 dark:bg-rose-500/10 dark:text-rose-300">
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <p className="text-sm font-semibold">{error}</p>
                  <button className={btnPrimary.replace('px-4 py-2.5', 'px-3 py-2')} onClick={() => navigate('/teacher/create-room')}>
                    <Plus className="h-4 w-4" />
                    创建课堂
                  </button>
                </div>
              </Card>
            )}

            <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 xl:grid-cols-3">
              {loading && (
                <Card className="p-5">
                  <p className="text-sm text-muted dark:text-slate-400">正在加载课堂数据...</p>
                </Card>
              )}

              {!loading && !error && (
                <button
                  key={roomCode}
                  onClick={() => navigate(`/teacher/rooms/${roomCode}/course`)}
                  className="hover-zoom rounded-lg border border-line bg-white text-left shadow-soft transition hover:-translate-y-0.5 hover:border-brand-200 dark:border-slate-800 dark:bg-slate-900"
                >
                  <div className="h-36 overflow-hidden rounded-t-lg bg-slate-200">
                    <img className="h-full w-full object-cover" src={fallbackCourse.cover} alt={`${courseTitle} 封面`} />
                  </div>
                  <div className="p-5">
                    <h2 className="text-lg font-semibold tracking-normal">{courseTitle}</h2>
                    <p className="mt-2 min-h-[40px] text-sm leading-5 text-muted dark:text-slate-400">
                      已有 {studentCount} 名学生，分布在 {groupCount} 个小组。
                    </p>
                    <div className="mt-5 grid grid-cols-3 gap-3 text-sm">
                      <div>
                        <p className="font-semibold">{studentCount}</p>
                        <p className="text-xs text-muted dark:text-slate-400">学生</p>
                      </div>
                      <div>
                        <p className="font-semibold">{taskCount}</p>
                        <p className="text-xs text-muted dark:text-slate-400">任务</p>
                      </div>
                      <div>
                        <p className="font-semibold">{roomStatusLabel(roomStatus)}</p>
                        <p className="text-xs text-muted dark:text-slate-400">状态</p>
                      </div>
                    </div>
                  </div>
                </button>
              )}
            </div>
          </div>

          <aside className="space-y-5">
            <Card>
              <div className="border-b border-line px-5 py-4 dark:border-slate-800">
                <h2 className="text-sm font-semibold">公告</h2>
              </div>
              <div className="divide-y divide-line dark:divide-slate-800">
                {announcements.map((item) => (
                  <div key={item.id} className="px-5 py-4">
                    <p className="text-sm font-semibold">{item.title}</p>
                    <p className="mt-1 text-xs leading-5 text-muted dark:text-slate-400">{item.body}</p>
                  </div>
                ))}
              </div>
            </Card>

            <Card>
              <div className="border-b border-line px-5 py-4 dark:border-slate-800">
                <h2 className="text-sm font-semibold">近期任务</h2>
              </div>
              <div className="divide-y divide-line dark:divide-slate-800">
                {tasks.length === 0 && (
                  <div className="px-5 py-4 text-sm text-muted dark:text-slate-400">暂无任务。</div>
                )}
                {tasks.slice(0, 3).map((item) => (
                  <button
                    key={item.taskId}
                    onClick={() => navigate(`/teacher/rooms/${roomCode}/review`)}
                    className="flex w-full items-center gap-3 px-5 py-4 text-left hover:bg-slate-50 dark:hover:bg-slate-800"
                  >
                    <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-brand-50 text-sm font-bold text-brand-700 dark:bg-brand-500/10 dark:text-brand-100">
                      {formatDay(item.deadlineAt)}
                    </span>
                    <span className="min-w-0">
                      <span className="block truncate text-sm font-semibold">{item.title}</span>
                      <span className="block text-xs text-muted dark:text-slate-400">{formatDateTime(item.deadlineAt)}</span>
                    </span>
                  </button>
                ))}
              </div>
            </Card>

            <Card padded>
              <div className="mb-4 flex items-center justify-between">
                <h2 className="text-sm font-semibold">{calendar.monthLabel}</h2>
                <div className="flex gap-1 text-slate-400">
                  <ChevronLeft className="h-4 w-4" />
                  <ChevronRight className="h-4 w-4" />
                </div>
              </div>
              <div className="grid grid-cols-7 gap-1 text-center text-xs">
                {['S', 'M', 'T', 'W', 'T', 'F', 'S'].map((day, i) => (
                  <div key={`h-${i}`} className="py-1 font-semibold text-muted dark:text-slate-500">
                    {day}
                  </div>
                ))}
                {Array.from({ length: calendar.startOffset }).map((_, i) => (
                  <div key={`pad-${i}`} />
                ))}
                {Array.from({ length: calendar.daysInMonth }).map((_, i) => {
                  const day = i + 1
                  const isToday = day === calendar.today
                  const isEvent = calendar.eventDays.includes(day)
                  const cls = isToday
                    ? 'bg-brand-600 text-white font-bold'
                    : isEvent
                      ? 'bg-brand-50 text-brand-700 font-semibold dark:bg-brand-500/10 dark:text-brand-100'
                      : 'text-slate-600 dark:text-slate-300'
                  return (
                    <div key={day} className={`rounded-md py-2 ${cls}`}>
                      {day}
                    </div>
                  )
                })}
              </div>
            </Card>
          </aside>
        </div>
      </main>
      <ToastView />
    </div>
  )
}

function getErrorMessage(error: unknown, fallback: string) {
  if (error instanceof Error && error.message.trim()) {
    return error.message
  }
  return fallback
}

function isAuthError(error: unknown) {
  return error instanceof ApiRequestError && (error.status === 401 || error.status === 403)
}

function formatDateTime(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return date.toLocaleString('zh-CN', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
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

function formatDay(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return '--'
  }
  return String(date.getDate()).padStart(2, '0')
}
