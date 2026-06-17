import { useCallback, useEffect, useState } from 'react'
import { CheckCircle2, Clock, Download, RefreshCw } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import StudentHeader from '../../components/layout/StudentHeader'
import RoomInfoCard from '../../components/RoomInfoCard'
import Card from '../../components/ui/Card'
import Badge from '../../components/ui/Badge'
import TaskCard from '../../components/TaskCard'
import RankingBoard from '../../components/RankingBoard'
import { useToast } from '../../components/ui/useToast'
import { useRoomWebSocket } from '../../hooks/useRoomWebSocket'
import { useStudentSession } from '../../hooks/useStudentSession'
import { ApiRequestError } from '../../api/client'
import { getStudentRanking, getStudentResults, listStudentTasks } from '../../api/student'
import type { RoomWebSocketEventType } from '../../api/websocket'
import type { BadgeTone, RankingVM } from '../../types'
import type { LeaderboardEntry, StudentResults, Task } from '../../types/api'

const CLASSROOM_WS_EVENTS: readonly RoomWebSocketEventType[] = [
  'task_published',
  'task_paused',
  'task_closed',
  'score_updated',
  'ranking_updated',
  'room_ended',
]

// /student/classroom
// 迁移自 docs/prototypes/student.html 的 #classroomScreen 概览部分
// （课堂头部 + 状态 + 任务进度 + 教师反馈 + 排行 + 我的进度）。
export default function Classroom() {
  const navigate = useNavigate()
  const { showToast, ToastView } = useToast()
  const { identity, clear } = useStudentSession()
  const roomCode = identity?.roomCode ?? ''
  const clientToken = identity?.clientToken ?? ''
  const studentToken = clientToken
  const [tasks, setTasks] = useState<Task[]>([])
  const [results, setResults] = useState<StudentResults | null>(null)
  const [ranking, setRanking] = useState<LeaderboardEntry[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [roomEnded, setRoomEnded] = useState(false)
  const [refreshVersion, setRefreshVersion] = useState(0)
  const team = identity?.groupName ?? ''

  const refreshClassroomData = useCallback(() => {
    setRefreshVersion((version) => version + 1)
  }, [])

  useEffect(() => {
    let cancelled = false

    async function loadClassroomData() {
      setLoading(true)
      setError(null)

      if (!roomCode || !studentToken) {
        setTasks([])
        setResults(null)
        setRanking([])
        setError('学生会话缺失，请重新加入课堂。')
        setLoading(false)
        return
      }

      try {
        const auth = { studentToken }
        const [taskData, resultData, rankingData] = await Promise.all([
          listStudentTasks(auth),
          getStudentResults(auth),
          getStudentRanking(roomCode, auth),
        ])

        if (cancelled) {
          return
        }

        setTasks(taskData)
        setResults(resultData)
        setRanking(rankingData)
      } catch (err) {
        if (cancelled) {
          return
        }

        if (isAuthError(err)) {
          clear()
          setError('你的会话已过期，请重新加入课堂。')
        } else if (isRoomEndedError(err)) {
          setRoomEnded(true)
          setError('课堂已结束。')
        } else {
          setError(getErrorMessage(err, '加载课堂数据失败。'))
        }
        setTasks([])
        setResults(null)
        setRanking([])
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    loadClassroomData()
    return () => {
      cancelled = true
    }
  }, [clear, refreshVersion, roomCode, studentToken])

  const ws = useRoomWebSocket({
    roomCode,
    role: 'student',
    clientToken,
    onEvent: (event) => {
      if (CLASSROOM_WS_EVENTS.includes(event.type)) {
        if (event.type === 'room_ended') {
          setRoomEnded(true)
          showToast('课堂已结束')
        }
        refreshClassroomData()
      }
    },
    onReconnect: refreshClassroomData,
  })

  const resultItems = results?.results ?? []
  const submittedCount = resultItems.filter((item) => item.submissionStatus !== 'notSubmitted').length
  const gradedResults = resultItems.filter((item) => item.submissionStatus === 'graded' || item.score !== null)
  const reviewed = gradedResults.length > 0
  const submitted = submittedCount > 0
  const latestFeedback = gradedResults[0] ?? null

  const status: { label: string; tone: BadgeTone } = roomEnded
    ? { label: '已结束', tone: 'slate' }
    : reviewed
    ? { label: '已评分', tone: 'emerald' }
    : submitted
      ? { label: '已提交', tone: 'brand' }
      : { label: '进行中', tone: 'amber' }

  const rankings = toRankingVM(ranking)
  const completed = submittedCount
  const pending = Math.max(tasks.length - submittedCount, 0)

  const currentTitle = tasks[0]?.title ?? '课堂任务'
  const earliestDeadline = findEarliestDeadline(tasks)

  return (
    <div className="min-h-screen bg-soft text-ink dark:bg-slate-950 dark:text-slate-100">
      <StudentHeader roomCode={roomCode} connected={ws.isConnected} student={identity} />

      <main className="mx-auto max-w-[1180px] px-6 py-8 sm:px-8">
        <section className="mb-6 grid grid-cols-1 gap-6 lg:grid-cols-[1fr_340px]">
          <RoomInfoCard
            eyebrow={identity?.groupName ?? '学生课堂'}
            title={currentTitle}
            meta={`学生：${identity?.nickname ?? '未知'} · 房间 ${roomCode || '未知'}`}
            right={
              <div className="rounded-lg bg-brand-50 px-5 py-4 text-center dark:bg-brand-500/10">
                <p className="text-xs font-bold uppercase text-brand-700 dark:text-brand-100">截止时间</p>
                <p className="mt-1 text-sm font-bold tabular-nums text-brand-700 dark:text-brand-100">
                  {earliestDeadline ? formatDateTime(earliestDeadline) : '暂无任务'}
                </p>
              </div>
            }
          />
          <Card padded>
            <div className="flex items-center justify-between">
              <h2 className="text-sm font-semibold">当前状态</h2>
              <Badge tone={status.tone}>{status.label}</Badge>
            </div>
            <p className="mt-4 text-sm leading-6 text-muted dark:text-slate-400">
              {roomEnded
                ? '课堂已结束，你仍可以查看已有结果。'
                : submitted
                ? '你的答案已提交，正在等待老师评分和评语。'
                : '打开任务作答，确认无误后提交。'}
            </p>
          </Card>
        </section>

        {error && (
          <Card className="mb-6 border-rose-200 bg-rose-50 p-4 text-rose-700 dark:border-rose-500/20 dark:bg-rose-500/10 dark:text-rose-300">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <p className="text-sm font-semibold">{error}</p>
              {isSessionError(error) && (
                <button
                  className="rounded-lg bg-white px-3 py-2 text-sm font-semibold text-rose-700 shadow-sm hover:bg-rose-50 dark:bg-slate-900 dark:text-rose-300 dark:hover:bg-slate-800"
                  onClick={() => navigate(roomCode ? `/student?room=${roomCode}` : '/student')}
                >
                  重新加入
                </button>
              )}
            </div>
          </Card>
        )}

        {loading && (
          <Card className="mb-6 p-5">
            <p className="text-sm text-muted dark:text-slate-400">正在加载课堂数据...</p>
          </Card>
        )}

        {submitted && (
          <Card className="mb-6 border-emerald-200 bg-emerald-50 p-4 text-emerald-800 dark:border-emerald-500/20 dark:bg-emerald-500/10 dark:text-emerald-300">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div className="flex items-center gap-3">
                <CheckCircle2 className="h-5 w-5" />
                <div>
                  <p className="font-semibold">提交成功</p>
                  <p className="text-sm opacity-80">
                    {reviewed
                      ? '老师反馈已完成，可以查看分数和评语。'
                      : '正在等待老师批改和评分。'}
                  </p>
                </div>
              </div>
              {!reviewed && (
                <button
                  className="inline-flex items-center gap-2 rounded-lg bg-white px-3 py-2 text-sm font-semibold text-emerald-800 shadow-sm hover:bg-emerald-50 dark:bg-slate-900 dark:text-emerald-300 dark:hover:bg-slate-800"
                  onClick={() => {
                    refreshClassroomData()
                    showToast('正在检查老师反馈')
                  }}
                >
                  <RefreshCw className="h-4 w-4" />
                  查看老师反馈
                </button>
              )}
            </div>
          </Card>
        )}

        <div className="grid grid-cols-1 gap-6 lg:grid-cols-[1fr_340px]">
          <Card>
            <div className="border-b border-line p-5 dark:border-slate-800">
              <h2 className="text-lg font-semibold tracking-normal">任务问题</h2>
              <p className="mt-1 text-sm text-muted dark:text-slate-400">
                打开任务后填写答案，提交前请确认内容完整。
              </p>
            </div>
            <div className="space-y-2 p-5">
              {!loading && tasks.length === 0 && (
                <p className="text-sm text-muted dark:text-slate-400">暂无可用任务。</p>
              )}
              {tasks.map((task) => (
                <TaskCard
                  key={task.taskId}
                  compact
                  title={task.title}
                  status={taskStatus(task)}
                  onClick={() => navigate(`/student/tasks/${task.taskId}`)}
                />
              ))}
            </div>
          </Card>

          <aside className="space-y-6">
            <Card padded>
              <div className="flex items-center justify-between">
                <h2 className="text-sm font-semibold">老师反馈</h2>
                <Badge tone={reviewed ? 'emerald' : submitted ? 'amber' : 'slate'}>
                  {reviewed ? '已评分' : submitted ? '待批改' : '未提交'}
                </Badge>
              </div>
              <div className="mt-4 rounded-lg border border-line bg-slate-50 p-4 dark:border-slate-800 dark:bg-slate-950">
                {!submitted && (
                  <p className="text-sm leading-6 text-muted dark:text-slate-400">
                    请先提交答案。老师批改后，反馈会显示在这里。
                  </p>
                )}
                {submitted && !reviewed && (
                  <div className="flex items-start gap-3">
                    <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-amber-50 text-amber-700 dark:bg-amber-500/10 dark:text-amber-300">
                      <Clock className="h-5 w-5" />
                    </div>
                    <div>
                      <p className="text-sm font-semibold">等待老师批改</p>
                      <p className="mt-1 text-sm leading-6 text-muted dark:text-slate-400">
                        你的提交已送达。老师评分后，这里会显示分数和评语。
                      </p>
                    </div>
                  </div>
                )}
                {reviewed && (
                  <>
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-semibold">分数</span>
                      <Badge tone="brand" className="text-sm">
                        {latestFeedback?.score ?? 0} / 10
                      </Badge>
                    </div>
                    <p className="mt-4 text-sm leading-6 text-slate-700 dark:text-slate-300">
                      {latestFeedback?.comment || '老师评分已完成。'}
                    </p>
                    <button
                      className="mt-4 inline-flex w-full items-center justify-center gap-2 rounded-lg border border-line bg-white px-3 py-2.5 text-sm font-semibold text-slate-700 hover:bg-slate-50 dark:border-slate-800 dark:bg-slate-900 dark:text-slate-200 dark:hover:bg-slate-800"
                      onClick={() => showToast('导出 PDF 反馈暂未实现')}
                    >
                      <Download className="h-4 w-4" />
                      下载 PDF 反馈
                    </button>
                  </>
                )}
              </div>
            </Card>

            <Card padded>
              <h2 className="text-sm font-semibold">小组排行榜</h2>
              <div className="mt-4">
                {rankings.length > 0 ? (
                  <RankingBoard rankings={rankings} myTeam={team} />
                ) : (
                  <p className="text-sm text-muted dark:text-slate-400">暂无排行数据。</p>
                )}
              </div>
            </Card>

            <Card padded>
              <h2 className="text-sm font-semibold">我的进度</h2>
              <div className="mt-4 grid grid-cols-2 gap-3">
                <div className="rounded-lg bg-emerald-50 p-4 dark:bg-emerald-500/10">
                  <p className="text-2xl font-bold text-emerald-700 dark:text-emerald-300">{completed}</p>
                  <p className="mt-1 text-xs font-semibold text-emerald-700 dark:text-emerald-300">已完成</p>
                </div>
                <div className="rounded-lg bg-amber-50 p-4 dark:bg-amber-500/10">
                  <p className="text-2xl font-bold text-amber-700 dark:text-amber-300">{pending}</p>
                  <p className="mt-1 text-xs font-semibold text-amber-700 dark:text-amber-300">待完成</p>
                </div>
              </div>
            </Card>
          </aside>
        </div>
      </main>

      <ToastView />
    </div>
  )
}

function isAuthError(error: unknown) {
  return error instanceof ApiRequestError && (error.status === 401 || error.status === 403)
}

function isRoomEndedError(error: unknown) {
  return error instanceof ApiRequestError && error.errorCode === 'ROOM_ENDED'
}

function getErrorMessage(error: unknown, fallback: string) {
  if (error instanceof Error && error.message.trim()) {
    return error.message
  }
  return fallback
}

function isSessionError(message: string) {
  const normalized = message.toLowerCase()
  return normalized.includes('session') || message.includes('会话')
}

function toRankingVM(entries: LeaderboardEntry[]): RankingVM[] {
  return entries.map((entry) => ({
    team: entry.groupName,
    score: entry.scoreTotal,
  }))
}

function taskStatus(task: Task): { label: string; tone: BadgeTone } {
  if (task.mySubmissionStatus === 'graded') {
    return { label: '已评分', tone: 'emerald' }
  }
  if (task.mySubmissionStatus === 'submitted') {
    return { label: '已提交', tone: 'brand' }
  }
  if (task.status === 'closed') {
    return { label: '已关闭', tone: 'slate' }
  }
  if (task.status === 'paused') {
    return { label: '已暂停', tone: 'amber' }
  }
  return { label: '待提交', tone: 'slate' }
}

function findEarliestDeadline(tasks: Task[]) {
  const dates = tasks
    .map((task) => task.deadlineAt)
    .filter((value) => !Number.isNaN(new Date(value).getTime()))
    .sort((a, b) => new Date(a).getTime() - new Date(b).getTime())
  return dates[0] ?? null
}

function formatDateTime(value: string) {
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
