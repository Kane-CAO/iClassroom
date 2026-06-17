import { useCallback, useEffect, useMemo, useState } from 'react'
import { ArrowLeft, ArrowRight, Clock, Pause, Play, Plus, RotateCcw, Search, Star } from 'lucide-react'
import { useNavigate, useParams } from 'react-router-dom'
import TeacherHeader from '../../components/layout/TeacherHeader'
import SubmissionCard, { reviewStatusTone } from '../../components/SubmissionCard'
import ImagePreview from '../../components/ImagePreview'
import Badge from '../../components/ui/Badge'
import { btnPrimary, btnSecondary } from '../../components/ui/buttons'
import { useToast } from '../../components/ui/useToast'
import { useCountdown } from '../../hooks/useCountdown'
import { useRoomWebSocket } from '../../hooks/useRoomWebSocket'
import { useTeacherSession } from '../../hooks/useTeacherSession'
import { ApiRequestError } from '../../api/client'
import { featureSubmission, gradeSubmission, listTaskSubmissions, listTeacherTasks } from '../../api/tasks'
import type { RoomWebSocketEventType } from '../../api/websocket'
import type { ReviewStatus, SubmissionVM } from '../../types'
import type { Submission, Task } from '../../types/api'

type Filter = 'all' | ReviewStatus

const FILTERS: { key: Filter; label: string }[] = [
  { key: 'all', label: '全部' },
  { key: 'submitted', label: '已提交' },
  { key: 'pending', label: '待提交' },
  { key: 'reviewed', label: '已批改' },
]

const REVIEW_WS_EVENTS: readonly RoomWebSocketEventType[] = [
  'submission_created',
  'score_updated',
  'ranking_updated',
  'room_ended',
]

export default function Review() {
  const { roomCode = 'ABC123' } = useParams()
  const navigate = useNavigate()
  const { showToast, ToastView } = useToast()
  const timer = useCountdown(45 * 60)
  const { teacherToken, clear } = useTeacherSession()

  const [tasks, setTasks] = useState<Task[]>([])
  const [selectedTaskId, setSelectedTaskId] = useState<number | null>(null)
  const [items, setItems] = useState<Submission[]>([])
  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [filter, setFilter] = useState<Filter>('all')
  const [query, setQuery] = useState('')
  const [score, setScore] = useState(1)
  const [feedback, setFeedback] = useState('')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [featuring, setFeaturing] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [refreshVersion, setRefreshVersion] = useState(0)

  const refreshReviewData = useCallback(() => {
    setRefreshVersion((version) => version + 1)
  }, [])

  useEffect(() => {
    let cancelled = false

    async function loadReviewData() {
      setLoading(true)
      setError(null)

      if (!teacherToken) {
        setTasks([])
        setItems([])
        setError('老师会话缺失，请重新创建课堂。')
        setLoading(false)
        return
      }

      try {
        const auth = { teacherToken }
        const taskData = await listTeacherTasks(roomCode, auth)
        const nextTaskId = resolveTaskId(taskData, selectedTaskId)
        const submissions = nextTaskId ? await listTaskSubmissions(nextTaskId, auth) : []

        if (cancelled) {
          return
        }

        setTasks(taskData)
        setSelectedTaskId(nextTaskId)
        setItems(submissions)
        setSelectedId((prev) => resolveSubmissionId(submissions, prev))
      } catch (err) {
        if (cancelled) {
          return
        }

        if (isAuthError(err)) {
          clear()
          setError('老师凭证无效或已过期，请重新创建课堂。')
        } else {
          setError(getErrorMessage(err, '加载提交列表失败。'))
        }
        setTasks([])
        setItems([])
        setSelectedId(null)
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    loadReviewData()
    return () => {
      cancelled = true
    }
  }, [clear, refreshVersion, roomCode, selectedTaskId, teacherToken])

  useRoomWebSocket({
    roomCode,
    role: 'teacher',
    onEvent: (event) => {
      if (REVIEW_WS_EVENTS.includes(event.type)) {
        refreshReviewData()
      }
    },
    onReconnect: refreshReviewData,
  })

  const current = items.find((s) => s.submissionId === selectedId) ?? items[0] ?? null
  const currentTask = tasks.find((task) => task.taskId === selectedTaskId) ?? tasks[0] ?? null

  useEffect(() => {
    if (!current) {
      setScore(1)
      setFeedback('')
      return
    }
    setScore(current.score ?? 1)
    setFeedback(current.comment ?? '')
  }, [current?.submissionId, current?.score, current?.comment])

  const visible = useMemo(
    () =>
      items.filter((submission) => {
        const status = toReviewStatus(submission)
        const name = submission.nickname ?? `学生 ${submission.studentId}`
        const matchFilter = filter === 'all' || status === filter
        const matchQuery = !query || name.toLowerCase().includes(query.trim().toLowerCase())
        return matchFilter && matchQuery
      }),
    [items, filter, query],
  )

  const selectSubmission = (id: number) => {
    const next = items.find((s) => s.submissionId === id)
    if (!next) return
    setSelectedId(id)
    setScore(next.score ?? 1)
    setFeedback(next.comment ?? '')
  }

  const moveSubmission = (step: number) => {
    const index = items.findIndex((s) => s.submissionId === selectedId)
    const nextIndex = Math.max(0, Math.min(items.length - 1, index + step))
    const next = items[nextIndex]
    if (next) {
      selectSubmission(next.submissionId)
    }
  }

  const saveReview = async () => {
    if (!current || !teacherToken) {
      showToast('未选择提交')
      return
    }

    setSaving(true)
    setError(null)

    try {
      await gradeSubmission(current.submissionId, { score, comment: feedback.trim() }, { teacherToken })
      showToast('批改已保存')
      refreshReviewData()
    } catch (err) {
      const message = getErrorMessage(err, '保存批改失败。')
      setError(message)
      showToast(message)
    } finally {
      setSaving(false)
    }
  }

  const featureCurrent = async () => {
    if (!current || !teacherToken) {
      showToast('未选择提交')
      return
    }

    setFeaturing(true)
    setError(null)

    try {
      await featureSubmission(current.submissionId, { displayMode: 'showGroup' }, { teacherToken })
      showToast('精选答案已更新')
    } catch (err) {
      const message = getErrorMessage(err, '设置精选答案失败。')
      setError(message)
      showToast(message)
    } finally {
      setFeaturing(false)
    }
  }

  const currentStatus = current ? toReviewStatus(current) : 'pending'
  const currentVM = current ? toSubmissionVM(current) : null

  return (
    <div className="min-h-screen bg-canvas text-ink dark:bg-slate-950 dark:text-slate-100">
      <TeacherHeader roomCode={roomCode} active="assignments" />

      <main className="px-8 py-7">
        <div className="mb-5 rounded-lg border border-brand-100 bg-brand-50 p-4 shadow-soft dark:border-brand-500/20 dark:bg-brand-500/10">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div>
              <p className="text-sm font-semibold text-brand-700 dark:text-brand-100">
                {currentTask?.title ?? '未选择任务'}
              </p>
              <h1 className="mt-1 text-2xl font-semibold tracking-normal">提交批改</h1>
              <select
                value={selectedTaskId ?? ''}
                onChange={(event) => {
                  const value = Number(event.target.value)
                  setSelectedTaskId(Number.isFinite(value) ? value : null)
                  setSelectedId(null)
                }}
                className="mt-3 rounded-lg border border-brand-200 bg-white px-3 py-2 text-sm font-semibold outline-none dark:border-brand-500/20 dark:bg-slate-900"
              >
                {tasks.length === 0 && <option value="">暂无任务</option>}
                {tasks.map((task) => (
                  <option key={task.taskId} value={task.taskId}>
                    {task.title}
                  </option>
                ))}
              </select>
            </div>
            <div className="flex items-center gap-3">
              <div className="rounded-lg bg-white px-5 py-3 text-center shadow-sm dark:bg-slate-900">
                <p className="text-xs font-semibold uppercase text-muted dark:text-slate-400">课堂计时器</p>
                <p className="mt-1 font-mono text-3xl font-bold tabular-nums text-brand-700 dark:text-brand-100">
                  {timer.mmss}
                </p>
              </div>
              <button className={btnPrimary.replace('px-4 py-2.5', 'px-3.5 py-2').replace('shadow-soft', '')} onClick={timer.toggle}>
                {timer.running ? <Pause className="h-4 w-4" /> : <Play className="h-4 w-4" />}
                {timer.running ? '暂停' : '开始'}
              </button>
              <button className={btnSecondary} onClick={timer.reset}>
                <RotateCcw className="h-4 w-4" />
                重置
              </button>
              <button
                className={btnSecondary}
                onClick={() => {
                  timer.extend(5)
                  showToast('已增加 5 分钟')
                }}
              >
                <Plus className="h-4 w-4" />5 min
              </button>
            </div>
          </div>
        </div>

        {error && (
          <div className="mb-5 rounded-lg border border-rose-200 bg-rose-50 p-4 text-sm font-semibold text-rose-700 dark:border-rose-500/20 dark:bg-rose-500/10 dark:text-rose-300">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <span>{error}</span>
              {error.toLowerCase().includes('teacher token') || error.toLowerCase().includes('session') ? (
                <button className={btnSecondary} onClick={() => navigate('/teacher/create-room')}>
                  创建课堂
                </button>
              ) : null}
            </div>
          </div>
        )}

        <div className="grid gap-5 lg:h-[calc(100vh-184px)] lg:grid-cols-[300px_1fr_360px]">
          <aside className="flex flex-col overflow-hidden rounded-lg border border-line bg-white shadow-soft dark:border-slate-800 dark:bg-slate-900">
            <div className="border-b border-line p-4 dark:border-slate-800">
              <h2 className="text-sm font-semibold">学生提交</h2>
              <div className="relative mt-3">
                <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted" />
                <input
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  placeholder="搜索学生"
                  className="w-full rounded-lg border border-line bg-white py-2 pl-9 pr-3 text-sm outline-none focus:border-brand-500 dark:border-slate-800 dark:bg-slate-950"
                />
              </div>
              <div className="mt-3 flex flex-wrap gap-2">
                {FILTERS.map((f) => (
                  <button
                    key={f.key}
                    onClick={() => setFilter(f.key)}
                    className={`rounded-lg border px-2.5 py-1.5 text-xs font-semibold transition ${
                      filter === f.key
                        ? 'border-brand-200 bg-brand-50 text-brand-700 dark:border-brand-500/24 dark:bg-brand-500/14 dark:text-brand-100'
                        : 'border-line dark:border-slate-800'
                    }`}
                  >
                    {f.label}
                  </button>
                ))}
              </div>
            </div>
            <div className="flex-1 overflow-y-auto p-3">
              {loading ? (
                <p className="p-3 text-sm text-muted dark:text-slate-400">正在加载提交...</p>
              ) : visible.length === 0 ? (
                <p className="p-3 text-sm text-muted dark:text-slate-400">没有符合筛选条件的学生。</p>
              ) : (
                visible.map((submission) => {
                  const vm = toSubmissionVM(submission)
                  return (
                    <SubmissionCard
                      key={submission.submissionId}
                      submission={vm}
                      active={submission.submissionId === selectedId}
                      onClick={() => selectSubmission(submission.submissionId)}
                    />
                  )
                })
              )}
            </div>
          </aside>

          <section className="overflow-y-auto rounded-lg border border-line bg-white shadow-soft dark:border-slate-800 dark:bg-slate-900">
            <div className="border-b border-line p-5 dark:border-slate-800">
              <div className="flex items-start justify-between">
                <div>
                  <h2 className="text-xl font-semibold tracking-normal">{currentVM?.name ?? '未选择提交'}</h2>
                  <p className="mt-1 text-sm text-muted dark:text-slate-400">
                    提交时间：{currentVM?.time ?? '未提交'}
                  </p>
                </div>
                <Badge tone={reviewStatusTone[currentStatus]} className="capitalize">
                  {reviewStatusLabel(currentStatus)}
                </Badge>
              </div>
            </div>

            {!current || !currentVM ? (
              <div className="flex h-[520px] items-center justify-center p-10 text-center">
                <div>
                  <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-lg bg-amber-50 text-amber-700 dark:bg-amber-500/10 dark:text-amber-300">
                    <Clock className="h-7 w-7" />
                  </div>
                  <h3 className="mt-4 text-lg font-semibold">等待学生提交</h3>
                  <p className="mt-2 max-w-sm text-sm leading-6 text-muted dark:text-slate-400">
                    请选择已有提交的任务，或等待学生完成提交。
                  </p>
                </div>
              </div>
            ) : (
              <div className="space-y-6 p-5">
                <section>
                  <div className="mb-3 flex items-center justify-between">
                    <h3 className="text-sm font-semibold">已上传图片</h3>
                    <span className="text-xs text-muted dark:text-slate-400">{currentVM.images.length} 张图片</span>
                  </div>
                  <ImagePreview images={currentVM.images} showCaption emptyText="本次提交没有图片。" />
                </section>

                <section>
                  <h3 className="mb-3 text-sm font-semibold">已上传文件</h3>
                  <div className="rounded-lg border border-dashed border-line bg-slate-50 p-5 text-center text-sm text-muted dark:border-slate-800 dark:bg-slate-950">
                    文件附件预览暂未实现。
                  </div>
                </section>

                <section>
                  <h3 className="mb-3 text-sm font-semibold">文字回答</h3>
                  <div className="rounded-lg border border-line bg-slate-50 p-4 text-sm leading-7 text-slate-700 dark:border-slate-800 dark:bg-slate-950 dark:text-slate-300">
                    {current.contentText || '暂无文字回答。'}
                  </div>
                </section>

                <section>
                  <h3 className="mb-3 text-sm font-semibold">PDF 预览</h3>
                  <div className="rounded-lg border border-line bg-slate-50 p-5 dark:border-slate-800 dark:bg-slate-950">
                    <div className="mx-auto h-80 w-60 rounded-md border border-line bg-white p-5 shadow-sm dark:border-slate-800 dark:bg-slate-900">
                      <div className="mb-5 h-4 w-36 rounded bg-slate-200 dark:bg-slate-700" />
                      <div className="space-y-2">
                        <div className="h-2 rounded bg-slate-200 dark:bg-slate-700" />
                        <div className="h-2 rounded bg-slate-200 dark:bg-slate-700" />
                        <div className="h-2 w-4/5 rounded bg-slate-200 dark:bg-slate-700" />
                      </div>
                      <div className="mt-6 h-28 rounded border border-dashed border-slate-300 bg-slate-50 dark:border-slate-700 dark:bg-slate-950" />
                      <div className="mt-6 space-y-2">
                        <div className="h-2 rounded bg-slate-200 dark:bg-slate-700" />
                        <div className="h-2 w-2/3 rounded bg-slate-200 dark:bg-slate-700" />
                      </div>
                    </div>
                  </div>
                </section>
              </div>
            )}
          </section>

          <aside className="overflow-y-auto rounded-lg border border-line bg-white shadow-soft dark:border-slate-800 dark:bg-slate-900">
            <div className="border-b border-line p-5 dark:border-slate-800">
              <h2 className="text-lg font-semibold tracking-normal">老师反馈</h2>
              <p className="mt-1 text-sm text-muted dark:text-slate-400">
                给分、填写评语，并在后续版本上传批注 PDF。
              </p>
            </div>
            <div className="space-y-5 p-5">
              <label className="block">
                <span className="flex items-center justify-between text-sm font-semibold">
                  分数
                  <span className="text-brand-700 dark:text-brand-100">{score} / 10</span>
                </span>
                <input
                  type="range"
                  min={1}
                  max={10}
                  value={score}
                  disabled={!current || saving}
                  onChange={(e) => setScore(Number(e.target.value))}
                  className="mt-3 w-full"
                />
              </label>

              <label className="block">
                <span className="text-sm font-semibold">文字评语</span>
                <textarea
                  value={feedback}
                  disabled={!current || saving}
                  onChange={(e) => setFeedback(e.target.value)}
                  placeholder="给学生写一段有针对性的反馈。"
                  className="mt-2 h-40 w-full rounded-lg border border-line bg-white p-3 text-sm leading-6 outline-none focus:border-brand-500 disabled:opacity-70 dark:border-slate-800 dark:bg-slate-950"
                />
              </label>

              <div className="rounded-lg border border-dashed border-line bg-slate-50 px-3 py-5 text-center text-sm text-muted dark:border-slate-800 dark:bg-slate-950">
                批注 PDF 上传暂未实现
              </div>

              <div className="grid grid-cols-2 gap-2">
                <button className={btnSecondary} disabled={!current} onClick={() => moveSubmission(-1)}>
                  <ArrowLeft className="h-4 w-4" />
                  上一个
                </button>
                <button className={btnSecondary} disabled={!current} onClick={() => moveSubmission(1)}>
                  下一个
                  <ArrowRight className="h-4 w-4" />
                </button>
              </div>

              <button
                className="w-full rounded-lg border border-line px-3 py-2.5 text-sm font-semibold hover:bg-slate-50 disabled:opacity-60 dark:border-slate-800 dark:hover:bg-slate-800"
                disabled={!current || saving}
                onClick={saveReview}
              >
                {saving ? '保存中...' : '保存批改'}
              </button>
              <button
                className="inline-flex w-full items-center justify-center gap-2 rounded-lg bg-brand-600 px-3 py-2.5 text-sm font-semibold text-white hover:bg-brand-700 disabled:opacity-60"
                disabled={!current || featuring}
                onClick={featureCurrent}
              >
                <Star className="h-4 w-4" />
                {featuring ? '设置中...' : '设为精选答案'}
              </button>

              <section className="pt-2">
                <h3 className="mb-3 text-sm font-semibold">反馈历史</h3>
                <div className="space-y-3">
                  {(current ? buildHistory(current) : ['未选择提交']).map((item, i) => (
                    <div
                      key={`${item}-${i}`}
                      className="flex gap-3 rounded-lg border border-line p-3 dark:border-slate-800"
                    >
                      <span className="mt-1 h-2 w-2 shrink-0 rounded-full bg-brand-600" />
                      <p className="text-xs leading-5 text-slate-600 dark:text-slate-300">{item}</p>
                    </div>
                  ))}
                </div>
              </section>
            </div>
          </aside>
        </div>
      </main>

      <ToastView />
    </div>
  )
}

function resolveTaskId(tasks: Task[], selectedTaskId: number | null) {
  if (selectedTaskId && tasks.some((task) => task.taskId === selectedTaskId)) {
    return selectedTaskId
  }
  return tasks[0]?.taskId ?? null
}

function resolveSubmissionId(submissions: Submission[], selectedId: number | null) {
  if (selectedId && submissions.some((submission) => submission.submissionId === selectedId)) {
    return selectedId
  }
  return submissions[0]?.submissionId ?? null
}

function toReviewStatus(submission: Submission): ReviewStatus {
  return submission.status === 'graded' ? 'reviewed' : 'submitted'
}

function reviewStatusLabel(status: ReviewStatus) {
  switch (status) {
    case 'reviewed':
      return '已批改'
    case 'pending':
      return '待提交'
    case 'submitted':
      return '已提交'
  }
}

function toSubmissionVM(submission: Submission): SubmissionVM {
  const name = submission.nickname ?? `学生 ${submission.studentId}`
  return {
    id: String(submission.submissionId),
    name,
    initials: initials(name),
    status: toReviewStatus(submission),
    time: formatDateTime(submission.submittedAt),
    score: submission.score ?? 0,
    feedback: submission.comment ?? '',
    response: submission.contentText,
    images: submission.images.map((image) => ({
      label: image.fileName,
      src: image.fileUrl,
    })),
    files: [],
    history: buildHistory(submission),
  }
}

function buildHistory(submission: Submission) {
  const history = [`收到提交：${formatDateTime(submission.submittedAt)}`]
  if (submission.status === 'graded') {
    history.unshift(`已保存分数：${submission.score ?? '-'}`)
  }
  if (submission.gradedAt) {
    history.unshift(`完成批改：${formatDateTime(submission.gradedAt)}`)
  }
  return history
}

function initials(name: string) {
  return (
    name
      .split(/\s+/)
      .filter(Boolean)
      .slice(0, 2)
      .map((part) => part[0])
      .join('')
      .toUpperCase() || 'ST'
  )
}

function isAuthError(error: unknown) {
  return error instanceof ApiRequestError && (error.status === 401 || error.status === 403)
}

function getErrorMessage(error: unknown, fallback: string) {
  if (error instanceof Error && error.message.trim()) {
    return error.message
  }
  return fallback
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
