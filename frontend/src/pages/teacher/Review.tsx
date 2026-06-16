import { useCallback, useEffect, useMemo, useState } from 'react'
import { ArrowLeft, ArrowRight, Clock, Pause, Play, Plus, RotateCcw, Search } from 'lucide-react'
import { useParams } from 'react-router-dom'
import TeacherHeader from '../../components/layout/TeacherHeader'
import SubmissionCard, { reviewStatusTone } from '../../components/SubmissionCard'
import ImagePreview from '../../components/ImagePreview'
import Badge from '../../components/ui/Badge'
import { btnPrimary, btnSecondary } from '../../components/ui/buttons'
import { useToast } from '../../components/ui/useToast'
import { useCountdown } from '../../hooks/useCountdown'
import { useRoomWebSocket } from '../../hooks/useRoomWebSocket'
import { teacherMocks } from '../../mocks'
import type { RoomWebSocketEventType } from '../../api/websocket'
import type { ReviewStatus, SubmissionVM } from '../../types'

type Filter = 'all' | ReviewStatus

const FILTERS: { key: Filter; label: string }[] = [
  { key: 'all', label: 'All' },
  { key: 'submitted', label: 'Submitted' },
  { key: 'pending', label: 'Pending' },
  { key: 'reviewed', label: 'Reviewed' },
]

const REVIEW_WS_EVENTS: readonly RoomWebSocketEventType[] = [
  'student_joined',
  'submission_created',
  'score_updated',
  'ranking_updated',
  'room_ended',
]

// /teacher/rooms/:roomCode/review
// 迁移自 docs/prototypes/iClassroom.html 的 #page-review（提交批改 + 课堂计时器）。
export default function Review() {
  const { roomCode = 'ABC123' } = useParams()
  const { showToast, ToastView } = useToast()
  const timer = useCountdown(teacherMocks.classTimerSeconds)

  // 本地可变副本：批改保存会改写状态 / 历史，因此从 mock 拷贝到 state。
  const [items, setItems] = useState<SubmissionVM[]>(() =>
    teacherMocks.submissions.map((s) => ({ ...s })),
  )
  const [selectedId, setSelectedId] = useState(items[0].id)
  const [filter, setFilter] = useState<Filter>('all')
  const [query, setQuery] = useState('')
  const [score, setScore] = useState(items[0].score || 1)
  const [feedback, setFeedback] = useState(items[0].feedback)
  const [refreshVersion, setRefreshVersion] = useState(0)

  const refreshReviewData = useCallback(() => {
    setRefreshVersion((version) => version + 1)
  }, [])

  useEffect(() => {
    // TODO: replace mock submissions with API refetch for the current review task.
  }, [refreshVersion])

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

  const current = items.find((s) => s.id === selectedId) ?? items[0]

  const visible = useMemo(
    () =>
      items.filter((s) => {
        const matchFilter = filter === 'all' || s.status === filter
        const matchQuery = !query || s.name.toLowerCase().includes(query.trim().toLowerCase())
        return matchFilter && matchQuery
      }),
    [items, filter, query],
  )

  const selectStudent = (id: string) => {
    const next = items.find((s) => s.id === id)
    if (!next) return
    setSelectedId(id)
    setScore(next.score || 1)
    setFeedback(next.feedback)
  }

  const moveStudent = (step: number) => {
    const index = items.findIndex((s) => s.id === selectedId)
    const nextIndex = Math.max(0, Math.min(items.length - 1, index + step))
    selectStudent(items[nextIndex].id)
  }

  const saveReview = (publish = false) => {
    setItems((prev) =>
      prev.map((s) => {
        if (s.id !== selectedId) return s
        const history = [`Score ${score} saved on Jun 9`, ...s.history]
        if (publish) history.unshift('Feedback published and PDF review prepared')
        return {
          ...s,
          score,
          feedback,
          status: s.status === 'pending' ? s.status : 'reviewed',
          history,
        }
      }),
    )
    showToast(publish ? 'Feedback published to student' : 'Review saved')
  }

  const isEmpty = current.status === 'pending'

  return (
    <div className="min-h-screen bg-canvas text-ink dark:bg-slate-950 dark:text-slate-100">
      <TeacherHeader roomCode={roomCode} active="assignments" />

      <main className="px-8 py-7">
        {/* 顶部：作业信息 + 课堂计时器 */}
        <div className="mb-5 rounded-lg border border-brand-100 bg-brand-50 p-4 shadow-soft dark:border-brand-500/20 dark:bg-brand-500/10">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div>
              <p className="text-sm font-semibold text-brand-700 dark:text-brand-100">
                {teacherMocks.reviewAssignmentTitle}
              </p>
              <h1 className="mt-1 text-2xl font-semibold tracking-normal">Submission Review</h1>
            </div>
            <div className="flex items-center gap-3">
              <div className="rounded-lg bg-white px-5 py-3 text-center shadow-sm dark:bg-slate-900">
                <p className="text-xs font-semibold uppercase text-muted dark:text-slate-400">Class Timer</p>
                <p className="mt-1 font-mono text-3xl font-bold tabular-nums text-brand-700 dark:text-brand-100">
                  {timer.mmss}
                </p>
              </div>
              <button className={btnPrimary.replace('px-4 py-2.5', 'px-3.5 py-2').replace('shadow-soft', '')} onClick={timer.toggle}>
                {timer.running ? <Pause className="h-4 w-4" /> : <Play className="h-4 w-4" />}
                {timer.running ? 'Pause' : 'Start'}
              </button>
              <button className={btnSecondary} onClick={timer.reset}>
                <RotateCcw className="h-4 w-4" />
                Reset
              </button>
              <button
                className={btnSecondary}
                onClick={() => {
                  timer.extend(5)
                  showToast('5 minutes added')
                }}
              >
                <Plus className="h-4 w-4" />5 min
              </button>
            </div>
          </div>
        </div>

        <div className="grid gap-5 lg:h-[calc(100vh-184px)] lg:grid-cols-[300px_1fr_360px]">
          {/* 左侧：学生列表 */}
          <aside className="flex flex-col overflow-hidden rounded-lg border border-line bg-white shadow-soft dark:border-slate-800 dark:bg-slate-900">
            <div className="border-b border-line p-4 dark:border-slate-800">
              <h2 className="text-sm font-semibold">Student Submissions</h2>
              <div className="relative mt-3">
                <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted" />
                <input
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  placeholder="Search student"
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
              {visible.length === 0 ? (
                <p className="p-3 text-sm text-muted dark:text-slate-400">No students match this filter.</p>
              ) : (
                visible.map((s) => (
                  <SubmissionCard
                    key={s.id}
                    submission={s}
                    active={s.id === selectedId}
                    onClick={() => selectStudent(s.id)}
                  />
                ))
              )}
            </div>
          </aside>

          {/* 中间：提交详情 */}
          <section className="overflow-y-auto rounded-lg border border-line bg-white shadow-soft dark:border-slate-800 dark:bg-slate-900">
            <div className="border-b border-line p-5 dark:border-slate-800">
              <div className="flex items-start justify-between">
                <div>
                  <h2 className="text-xl font-semibold tracking-normal">{current.name}</h2>
                  <p className="mt-1 text-sm text-muted dark:text-slate-400">Submission time: {current.time}</p>
                </div>
                <Badge tone={reviewStatusTone[current.status]} className="capitalize">
                  {current.status}
                </Badge>
              </div>
            </div>

            {isEmpty ? (
              <div className="flex h-[520px] items-center justify-center p-10 text-center">
                <div>
                  <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-lg bg-amber-50 text-amber-700 dark:bg-amber-500/10 dark:text-amber-300">
                    <Clock className="h-7 w-7" />
                  </div>
                  <h3 className="mt-4 text-lg font-semibold">Waiting for submission</h3>
                  <p className="mt-2 max-w-sm text-sm leading-6 text-muted dark:text-slate-400">
                    This student has not uploaded files or images yet.
                  </p>
                </div>
              </div>
            ) : (
              <div className="space-y-6 p-5">
                <section>
                  <div className="mb-3 flex items-center justify-between">
                    <h3 className="text-sm font-semibold">Uploaded Images</h3>
                    <span className="text-xs text-muted dark:text-slate-400">{current.images.length} images</span>
                  </div>
                  <ImagePreview images={current.images} showCaption emptyText="No images in this submission." />
                </section>

                <section>
                  <h3 className="mb-3 text-sm font-semibold">Uploaded Files</h3>
                  <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                    {current.files.map((file) => (
                      <button
                        key={file.name}
                        className="flex items-center gap-3 rounded-lg border border-line p-3 text-left hover:bg-slate-50 dark:border-slate-800 dark:hover:bg-slate-800"
                      >
                        <span className="flex h-10 w-10 items-center justify-center rounded-lg bg-brand-50 text-xs font-bold text-brand-700 dark:bg-brand-500/10 dark:text-brand-100">
                          {file.type}
                        </span>
                        <span className="min-w-0">
                          <span className="block truncate text-sm font-semibold">{file.name}</span>
                          <span className="text-xs text-muted dark:text-slate-400">{file.size}</span>
                        </span>
                      </button>
                    ))}
                  </div>
                </section>

                <section>
                  <h3 className="mb-3 text-sm font-semibold">Written Response</h3>
                  <div className="rounded-lg border border-line bg-slate-50 p-4 text-sm leading-7 text-slate-700 dark:border-slate-800 dark:bg-slate-950 dark:text-slate-300">
                    {current.response}
                  </div>
                </section>

                <section>
                  <h3 className="mb-3 text-sm font-semibold">PDF Preview</h3>
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

          {/* 右侧：教师反馈 */}
          <aside className="overflow-y-auto rounded-lg border border-line bg-white shadow-soft dark:border-slate-800 dark:bg-slate-900">
            <div className="border-b border-line p-5 dark:border-slate-800">
              <h2 className="text-lg font-semibold tracking-normal">Teacher Feedback</h2>
              <p className="mt-1 text-sm text-muted dark:text-slate-400">
                Score, comment, upload annotated PDF, and publish.
              </p>
            </div>
            <div className="space-y-5 p-5">
              <label className="block">
                <span className="flex items-center justify-between text-sm font-semibold">
                  Score
                  <span className="text-brand-700 dark:text-brand-100">{score} / 10</span>
                </span>
                <input
                  type="range"
                  min={1}
                  max={10}
                  value={score}
                  onChange={(e) => setScore(Number(e.target.value))}
                  className="mt-3 w-full"
                />
              </label>

              <label className="block">
                <span className="text-sm font-semibold">Written Feedback</span>
                <textarea
                  value={feedback}
                  onChange={(e) => setFeedback(e.target.value)}
                  placeholder="Write focused feedback for the student."
                  className="mt-2 h-40 w-full rounded-lg border border-line bg-white p-3 text-sm leading-6 outline-none focus:border-brand-500 dark:border-slate-800 dark:bg-slate-950"
                />
              </label>

              <div className="rounded-lg border border-dashed border-line bg-slate-50 px-3 py-5 text-center text-sm text-muted dark:border-slate-800 dark:bg-slate-950">
                批注 PDF 上传暂未实现
              </div>

              <div className="grid grid-cols-2 gap-2">
                <button className={btnSecondary} onClick={() => moveStudent(-1)}>
                  <ArrowLeft className="h-4 w-4" />
                  Previous
                </button>
                <button className={btnSecondary} onClick={() => moveStudent(1)}>
                  Next
                  <ArrowRight className="h-4 w-4" />
                </button>
              </div>

              <button
                className="w-full rounded-lg border border-line px-3 py-2.5 text-sm font-semibold hover:bg-slate-50 dark:border-slate-800 dark:hover:bg-slate-800"
                onClick={() => saveReview(false)}
              >
                Save Review
              </button>
              <button
                className="w-full rounded-lg bg-brand-600 px-3 py-2.5 text-sm font-semibold text-white hover:bg-brand-700"
                onClick={() => saveReview(true)}
              >
                Publish Feedback
              </button>

              <section className="pt-2">
                <h3 className="mb-3 text-sm font-semibold">Feedback History</h3>
                <div className="space-y-3">
                  {current.history.map((item, i) => (
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
