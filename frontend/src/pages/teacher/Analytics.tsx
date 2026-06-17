import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { ApiRequestError } from '../../api/client'
import { getAnalytics } from '../../api/analytics'
import { getRoom } from '../../api/rooms'
import TeacherHeader from '../../components/layout/TeacherHeader'
import TeacherRoomActions from '../../components/teacher/TeacherRoomActions'
import Card from '../../components/ui/Card'
import { btnPrimary } from '../../components/ui/buttons'
import { barToneClass } from '../../components/ui/tones'
import { useToast } from '../../components/ui/useToast'
import { useTeacherSession } from '../../hooks/useTeacherSession'
import type { Analytics as AnalyticsData, Room } from '../../types/api'
import type { BadgeTone } from '../../types'

export default function Analytics() {
  const { roomCode = 'ABC123' } = useParams()
  const navigate = useNavigate()
  const { showToast, ToastView } = useToast()
  const { teacherToken, clear } = useTeacherSession()
  const [analytics, setAnalytics] = useState<AnalyticsData | null>(null)
  const [room, setRoom] = useState<Room | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [refreshVersion, setRefreshVersion] = useState(0)

  useEffect(() => {
    let cancelled = false

    async function loadAnalytics() {
      setLoading(true)
      setError(null)

      if (!teacherToken) {
        setRoom(null)
        setAnalytics(null)
        setError('老师会话缺失，请重新创建课堂。')
        setLoading(false)
        return
      }

      try {
        const auth = { teacherToken }
        const [roomData, analyticsData] = await Promise.all([
          getRoom(roomCode, auth),
          getAnalytics(roomCode, auth),
        ])
        if (cancelled) {
          return
        }
        setRoom(roomData)
        setAnalytics(analyticsData)
      } catch (err: unknown) {
        if (cancelled) {
          return
        }
        if (isAuthError(err)) {
          clear()
          setError('老师凭证无效，或无权访问该课堂。')
        } else {
          setError(getErrorMessage(err, '加载数据看板失败。'))
        }
        setRoom(null)
        setAnalytics(null)
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    void loadAnalytics()

    return () => {
      cancelled = true
    }
  }, [clear, refreshVersion, roomCode, teacherToken])

  const refreshRoomData = () => {
    setRefreshVersion((version) => version + 1)
  }

  const handleRoomActionError = (message: string) => {
    setError(message)
    showToast(message)
  }

  const handleRoomActionSuccess = (message: string) => {
    showToast(message)
  }

  const studentCount = analytics?.studentCount ?? 0
  const onlineCount = analytics?.onlineCount ?? 0
  const submissionRate = analytics?.submissionRate ?? 0
  const hasGroupScores = Boolean(analytics?.groupScores.length)
  const hasTaskCompletion = Boolean(analytics?.taskCompletion.length)
  const hasTimeline = Boolean(analytics?.submissionTimeline.length)
  const roomEnded = room?.status === 'ended'

  return (
    <div className="min-h-screen bg-canvas text-ink dark:bg-slate-950 dark:text-slate-100">
      <TeacherHeader roomCode={roomCode} active="courses" />

      <main className="px-8 py-7">
        <div className="flex flex-wrap items-end justify-between gap-4">
          <div>
            <h1 className="text-2xl font-semibold tracking-normal">数据看板</h1>
            <p className="mt-1 text-sm text-muted dark:text-slate-400">
              房间 {roomCode} · 实时课堂数据
            </p>
          </div>
          <div className="flex flex-wrap items-center justify-end gap-2">
            {loading && <span className="text-sm font-semibold text-muted dark:text-slate-400">加载中...</span>}
            <TeacherRoomActions
              roomCode={roomCode}
              teacherToken={teacherToken}
              roomEnded={roomEnded}
              onEnded={refreshRoomData}
              onError={handleRoomActionError}
              onSuccess={handleRoomActionSuccess}
            />
          </div>
        </div>

        {error && (
          <Card className="mt-6 border-rose-200 bg-rose-50 p-4 text-rose-700 dark:border-rose-500/20 dark:bg-rose-500/10 dark:text-rose-300">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <p className="text-sm font-semibold">{error}</p>
              <button className={btnPrimary.replace('px-4 py-2.5', 'px-3 py-2')} onClick={() => navigate('/teacher/create-room')}>
                创建课堂
              </button>
            </div>
          </Card>
        )}

        {loading && !analytics && !error && (
          <Card padded className="mt-6">
            <p className="text-sm text-muted dark:text-slate-400">正在加载数据看板...</p>
          </Card>
        )}

        {!loading && !error && analytics && (
          <>
            <div className="mt-6 grid grid-cols-1 gap-5 sm:grid-cols-3">
              <MetricCard label="学生数" value={studentCount} caption="已加入课堂" />
              <MetricCard label="在线数" value={onlineCount} caption="当前活跃" />
              <MetricCard label="提交率" value={`${Math.round(submissionRate * 100)}%`} caption="所有已发布任务" />
            </div>

            <div className="mt-6 grid grid-cols-1 gap-6 lg:grid-cols-[1fr_340px]">
              <div className="space-y-6">
                <Card padded>
                  <h2 className="text-sm font-semibold">任务完成情况</h2>
                  <div className="mt-5 space-y-4">
                    {hasTaskCompletion ? (
                      analytics.taskCompletion.map((task, index) => (
                        <ProgressRow
                          key={task.taskId}
                          label={task.taskTitle}
                          value={task.submittedCount}
                          total={task.targetStudentCount}
                          pct={task.completionRate}
                          tone={progressTone(index)}
                        />
                      ))
                    ) : (
                      <EmptyState text="暂无任务完成数据。" />
                    )}
                  </div>
                </Card>

                <Card padded>
                  <h2 className="text-sm font-semibold">小组得分</h2>
                  <div className="mt-5 space-y-4">
                    {hasGroupScores ? (
                      analytics.groupScores.map((group, index) => (
                        <ScoreRow
                          key={group.groupId}
                          label={group.groupName}
                          score={group.scoreTotal}
                          max={maxScore(analytics.groupScores)}
                          tone={progressTone(index)}
                        />
                      ))
                    ) : (
                      <EmptyState text="暂无小组得分。" />
                    )}
                  </div>
                </Card>
              </div>

              <Card padded>
                <h2 className="text-sm font-semibold">提交时间线</h2>
                <div className="mt-5">
                  {hasTimeline ? (
                    <Timeline points={analytics.submissionTimeline} />
                  ) : (
                    <EmptyState text="暂无提交记录。" />
                  )}
                </div>
              </Card>
            </div>
          </>
        )}

        {!loading && !error && analytics && !hasGroupScores && !hasTaskCompletion && !hasTimeline && (
          <Card padded className="mt-6">
            <p className="text-sm text-muted dark:text-slate-400">
              学生加入并提交任务后，这里会自动显示课堂数据。
            </p>
          </Card>
        )}
      </main>
      <ToastView />
    </div>
  )
}

function MetricCard({ label, value, caption }: { label: string; value: number | string; caption: string }) {
  return (
    <Card padded>
      <p className="text-xs font-bold uppercase text-muted dark:text-slate-400">{label}</p>
      <p className="mt-3 text-3xl font-semibold tracking-normal">{value}</p>
      <p className="mt-1 text-sm text-muted dark:text-slate-400">{caption}</p>
    </Card>
  )
}

function ProgressRow({
  label,
  value,
  total,
  pct,
  tone,
}: {
  label: string
  value: number
  total: number
  pct: number
  tone: BadgeTone
}) {
  const safePct = clampPercent(pct)
  return (
    <div>
      <div className="mb-2 flex justify-between gap-4 text-xs font-semibold">
        <span className="truncate">{label}</span>
        <span className="shrink-0">
          {value} / {total} · {Math.round(safePct * 100)}%
        </span>
      </div>
      <div className="h-2 rounded-full bg-slate-100 dark:bg-slate-800">
        <div className={`h-2 rounded-full ${barToneClass[tone]}`} style={{ width: `${safePct * 100}%` }} />
      </div>
    </div>
  )
}

function ScoreRow({
  label,
  score,
  max,
  tone,
}: {
  label: string
  score: number
  max: number
  tone: BadgeTone
}) {
  const pct = max > 0 ? score / max : 0
  return (
    <div>
      <div className="mb-2 flex justify-between gap-4 text-xs font-semibold">
        <span className="truncate">{label}</span>
        <span className="shrink-0">{score} 分</span>
      </div>
      <div className="h-2 rounded-full bg-slate-100 dark:bg-slate-800">
        <div className={`h-2 rounded-full ${barToneClass[tone]}`} style={{ width: `${clampPercent(pct) * 100}%` }} />
      </div>
    </div>
  )
}

function Timeline({ points }: { points: AnalyticsData['submissionTimeline'] }) {
  const peak = Math.max(...points.map((point) => point.count), 1)
  return (
    <div className="flex h-64 items-stretch gap-2 border-b border-line pb-3 dark:border-slate-800">
      {points.map((point) => {
        const height = Math.max(8, (point.count / peak) * 100)
        return (
          <div key={point.time} className="flex min-w-0 flex-1 flex-col items-center gap-2">
            <div className="flex min-h-0 w-full flex-1 items-end justify-center">
              <div
                className="w-full max-w-8 rounded-t-md bg-brand-600"
                style={{ height: `${height}%` }}
                title={`${formatTimelineLabel(point.time)} · ${point.count}`}
              />
            </div>
            <span className="w-full truncate text-center text-[10px] font-semibold text-muted dark:text-slate-500">
              {formatTimelineLabel(point.time)}
            </span>
          </div>
        )
      })}
    </div>
  )
}

function EmptyState({ text }: { text: string }) {
  return (
    <p className="rounded-lg border border-dashed border-line px-3 py-4 text-sm text-muted dark:border-slate-800 dark:text-slate-400">
      {text}
    </p>
  )
}

function maxScore(scores: AnalyticsData['groupScores']) {
  return Math.max(...scores.map((group) => group.scoreTotal), 0)
}

function progressTone(index: number): BadgeTone {
  const tones: BadgeTone[] = ['brand', 'emerald', 'sky', 'amber', 'slate']
  return tones[index % tones.length]
}

function clampPercent(value: number) {
  if (!Number.isFinite(value)) {
    return 0
  }
  return Math.min(Math.max(value, 0), 1)
}

function formatTimelineLabel(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return new Intl.DateTimeFormat('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

function isAuthError(err: unknown) {
  return err instanceof ApiRequestError && (err.status === 401 || err.status === 403)
}

function getErrorMessage(err: unknown, fallback: string) {
  if (err instanceof Error && err.message.trim()) {
    return err.message
  }
  return fallback
}
