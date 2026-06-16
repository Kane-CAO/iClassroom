import { useCallback, useEffect, useState } from 'react'
import { ArrowLeft, ArrowRight, Check, ImagePlus, Send } from 'lucide-react'
import { useNavigate, useParams } from 'react-router-dom'
import StudentHeader from '../../components/layout/StudentHeader'
import Card from '../../components/ui/Card'
import Badge from '../../components/ui/Badge'
import { btnSecondary, btnGradient } from '../../components/ui/buttons'
import { useToast } from '../../components/ui/useToast'
import { useRoomWebSocket } from '../../hooks/useRoomWebSocket'
import { questionDone, useStudentSession } from '../../hooks/useStudentSession'
import { studentMocks } from '../../mocks'
import type { RoomWebSocketEventType } from '../../api/websocket'

const TASK_DETAIL_WS_EVENTS: readonly RoomWebSocketEventType[] = [
  'task_published',
  'task_paused',
  'task_closed',
  'score_updated',
  'ranking_updated',
  'room_ended',
]

// /student/tasks/:taskId
// 迁移自 docs/prototypes/student.html 的答题面板（题目导航 + 文字 / 图片提交）。
// 注意：按需求不实现图片上传，"Image Upload" 仅作视觉占位。
export default function TaskDetail() {
  const { taskId } = useParams()
  const navigate = useNavigate()
  const { showToast, ToastView } = useToast()
  const { identity, answers, submitted, setAnswer, submit } = useStudentSession()
  const room = studentMocks.studentRoom
  const roomCode = identity?.roomCode ?? room.roomCode
  const questions = studentMocks.questions
  const [refreshVersion, setRefreshVersion] = useState(0)

  const foundIndex = questions.findIndex((q) => q.id === taskId)
  const index = foundIndex >= 0 ? foundIndex : 0
  const question = questions[index]

  const [tab, setTab] = useState<'text' | 'image'>('text')
  const isLast = index === questions.length - 1
  const text = answers[index] ?? ''

  const goTo = (i: number) => navigate(`/student/tasks/${questions[i].id}`)

  const refreshTaskDetailData = useCallback(() => {
    setRefreshVersion((version) => version + 1)
  }, [])

  useEffect(() => {
    // TODO: replace mock task detail/results data with API refetch for the current student.
  }, [refreshVersion])

  const ws = useRoomWebSocket({
    roomCode,
    role: 'student',
    clientToken: identity?.clientToken,
    onEvent: (event) => {
      if (TASK_DETAIL_WS_EVENTS.includes(event.type)) {
        refreshTaskDetailData()
      }
    },
    onReconnect: refreshTaskDetailData,
  })

  const onSubmit = () => {
    submit()
    showToast('Assignment submitted')
    setTimeout(() => navigate('/student/classroom'), 500)
  }

  const tabClass = (active: boolean) =>
    `rounded-md border px-4 py-2 text-sm font-semibold transition ${
      active
        ? 'border-brand-500 bg-brand-500 text-white'
        : 'border-transparent text-slate-600 dark:text-slate-300'
    }`

  return (
    <div className="min-h-screen bg-soft text-ink dark:bg-slate-950 dark:text-slate-100">
      <StudentHeader roomCode={roomCode} connected={ws.isConnected} student={identity} />

      <main className="mx-auto max-w-[1180px] px-6 py-8 sm:px-8">
        <Card>
          <div className="border-b border-line p-5 dark:border-slate-800">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-lg font-semibold tracking-normal">Assignment Questions</h2>
                <p className="mt-1 text-sm text-muted dark:text-slate-400">
                  Drafts save automatically in this browser.
                </p>
              </div>
              {/* 题目导航圆点 */}
              <div className="flex items-center gap-2">
                {questions.map((q, i) => {
                  const done = questionDone(answers[i] ?? '')
                  const active = i === index
                  return (
                    <button
                      key={q.id}
                      onClick={() => goTo(i)}
                      title={q.title}
                      className={`flex h-9 w-9 items-center justify-center rounded-lg border text-sm font-bold transition ${
                        active
                          ? 'border-brand-500 bg-brand-500 text-white'
                          : 'border-line dark:border-slate-800'
                      }`}
                    >
                      {done ? <Check className="h-4 w-4" /> : i + 1}
                    </button>
                  )
                })}
              </div>
            </div>
          </div>

          <div className="p-5">
            <div className="mb-5 rounded-lg bg-slate-50 p-4 dark:bg-slate-950">
              <div className="mb-2 flex items-center justify-between">
                <span className="text-xs font-bold uppercase text-brand-700 dark:text-brand-100">
                  Question {index + 1} of {questions.length}
                </span>
                <Badge tone={questionDone(text) ? 'brand' : 'slate'}>
                  {questionDone(text) ? 'Draft saved' : 'Draft'}
                </Badge>
              </div>
              <h3 className="text-xl font-semibold tracking-normal">{question.title}</h3>
              <p className="mt-2 text-sm leading-6 text-muted dark:text-slate-400">{question.prompt}</p>
            </div>

            <div className="flex rounded-lg border border-line bg-slate-50 p-1 dark:border-slate-800 dark:bg-slate-950">
              <button className={tabClass(tab === 'text')} onClick={() => setTab('text')}>
                Text Submission
              </button>
              <button className={tabClass(tab === 'image')} onClick={() => setTab('image')}>
                Image Upload
              </button>
            </div>

            {tab === 'text' ? (
              <div className="mt-5">
                <label className="block">
                  <span className="mb-2 block text-sm font-semibold">Written Response</span>
                  <textarea
                    value={text}
                    disabled={submitted}
                    onChange={(e) => setAnswer(index, e.target.value)}
                    placeholder="Write your answer here..."
                    className="h-[300px] w-full rounded-lg border border-line bg-white p-4 text-sm leading-7 outline-none transition focus:border-brand-500 disabled:opacity-70 dark:border-slate-800 dark:bg-slate-950"
                  />
                </label>
              </div>
            ) : (
              <div className="mt-5">
                <div className="flex h-[220px] flex-col items-center justify-center rounded-lg border-2 border-dashed border-line bg-slate-50 text-center dark:border-slate-800 dark:bg-slate-950">
                  <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-white text-brand-600 shadow-sm dark:bg-slate-900 dark:text-brand-100">
                    <ImagePlus className="h-6 w-6" />
                  </div>
                  <p className="mt-4 text-sm font-semibold">图片上传暂未实现</p>
                  <p className="mt-1 text-xs text-muted dark:text-slate-400">
                    本阶段仅作视觉参考，最多 {studentMocks.maxImagesPerQuestion} 张 · PNG / JPG
                  </p>
                </div>
              </div>
            )}

            <div className="mt-5 flex items-center justify-between gap-3 rounded-lg bg-slate-50 p-4 dark:bg-slate-950">
              <button className={btnSecondary} disabled={index === 0} onClick={() => goTo(index - 1)}>
                <ArrowLeft className="h-4 w-4" />
                Previous
              </button>
              <p className="hidden text-sm text-muted sm:block dark:text-slate-400">
                {submitted ? 'Submitted. Waiting for teacher review.' : 'Complete all questions before submitting.'}
              </p>
              <div className="flex items-center gap-2">
                {!isLast && (
                  <button className={btnSecondary} onClick={() => goTo(index + 1)}>
                    Next
                    <ArrowRight className="h-4 w-4" />
                  </button>
                )}
                {isLast && !submitted && (
                  <button className={btnGradient} onClick={onSubmit}>
                    <Send className="h-4 w-4" />
                    Submit Assignment
                  </button>
                )}
              </div>
            </div>
          </div>
        </Card>
      </main>

      <ToastView />
    </div>
  )
}
