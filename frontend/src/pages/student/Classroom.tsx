import { CheckCircle2, Clock, Download, RefreshCw } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import StudentHeader from '../../components/layout/StudentHeader'
import RoomInfoCard from '../../components/RoomInfoCard'
import Card from '../../components/ui/Card'
import Badge from '../../components/ui/Badge'
import TaskCard from '../../components/TaskCard'
import RankingBoard from '../../components/RankingBoard'
import { useToast } from '../../components/ui/useToast'
import { useCountdown } from '../../hooks/useCountdown'
import { questionDone, useStudentSession } from '../../hooks/useStudentSession'
import { studentMocks } from '../../mocks'
import type { BadgeTone } from '../../types'

// /student/classroom
// 迁移自 docs/prototypes/student.html 的 #classroomScreen 概览部分
// （课堂头部 + 状态 + 任务进度 + 教师反馈 + 排行 + 我的进度）。
export default function Classroom() {
  const navigate = useNavigate()
  const { showToast, ToastView } = useToast()
  const timer = useCountdown(studentMocks.remainingSeconds, true)
  const { identity, answers, submitted, reviewed, doneCount, review } = useStudentSession()
  const room = studentMocks.studentRoom
  const team = identity?.team ?? studentMocks.defaultStudentIdentity.team

  const status: { label: string; tone: BadgeTone } = reviewed
    ? { label: 'Scored', tone: 'emerald' }
    : submitted
      ? { label: 'Submitted', tone: 'brand' }
      : { label: 'In Progress', tone: 'amber' }

  const taskState = (text: string): { label: string; tone: BadgeTone } => {
    if (reviewed) return { label: 'Scored', tone: 'emerald' }
    if (submitted) return { label: 'Submitted', tone: 'brand' }
    if (questionDone(text)) return { label: 'Draft saved', tone: 'sky' }
    return { label: 'Pending', tone: 'slate' }
  }

  const rankings = reviewed ? studentMocks.rankingsAfterReview : studentMocks.rankingsBeforeReview
  const completed = reviewed
    ? studentMocks.myProgress.reviewedCompleted
    : studentMocks.myProgress.baseCompleted + (submitted ? studentMocks.questions.length : doneCount)
  const pending = submitted ? 0 : studentMocks.questions.length - doneCount

  return (
    <div className="min-h-screen bg-soft text-ink dark:bg-slate-950 dark:text-slate-100">
      <StudentHeader roomCode={room.roomCode} student={identity} />

      <main className="mx-auto max-w-[1180px] px-6 py-8 sm:px-8">
        <section className="mb-6 grid grid-cols-1 gap-6 lg:grid-cols-[1fr_340px]">
          <RoomInfoCard
            eyebrow={room.course}
            title={room.assignment}
            meta={`Teacher: ${room.teacher} · Room ${room.roomCode}`}
            right={
              <div className="rounded-lg bg-brand-50 px-5 py-4 text-center dark:bg-brand-500/10">
                <p className="text-xs font-bold uppercase text-brand-700 dark:text-brand-100">Remaining Time</p>
                <p className="mt-1 font-mono text-3xl font-bold tabular-nums text-brand-700 dark:text-brand-100">
                  {timer.mmss}
                </p>
              </div>
            }
          />
          <Card padded>
            <div className="flex items-center justify-between">
              <h2 className="text-sm font-semibold">Current Status</h2>
              <Badge tone={status.tone}>{status.label}</Badge>
            </div>
            <p className="mt-4 text-sm leading-6 text-muted dark:text-slate-400">
              {submitted
                ? 'Your answers are locked. Waiting for teacher score and comments.'
                : 'Answer each question, upload supporting images, then submit once.'}
            </p>
          </Card>
        </section>

        {submitted && (
          <Card className="mb-6 border-emerald-200 bg-emerald-50 p-4 text-emerald-800 dark:border-emerald-500/20 dark:bg-emerald-500/10 dark:text-emerald-300">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div className="flex items-center gap-3">
                <CheckCircle2 className="h-5 w-5" />
                <div>
                  <p className="font-semibold">Submitted Successfully</p>
                  <p className="text-sm opacity-80">
                    {reviewed
                      ? 'Teacher feedback is ready. Score, comments, and PDF review are available.'
                      : 'Waiting for your teacher to review and score your work.'}
                  </p>
                </div>
              </div>
              {!reviewed && (
                <button
                  className="inline-flex items-center gap-2 rounded-lg bg-white px-3 py-2 text-sm font-semibold text-emerald-800 shadow-sm hover:bg-emerald-50 dark:bg-slate-900 dark:text-emerald-300 dark:hover:bg-slate-800"
                  onClick={() => {
                    review()
                    showToast('Teacher feedback is ready')
                  }}
                >
                  <RefreshCw className="h-4 w-4" />
                  Check Teacher Feedback
                </button>
              )}
            </div>
          </Card>
        )}

        <div className="grid grid-cols-1 gap-6 lg:grid-cols-[1fr_340px]">
          <Card>
            <div className="border-b border-line p-5 dark:border-slate-800">
              <h2 className="text-lg font-semibold tracking-normal">Assignment Questions</h2>
              <p className="mt-1 text-sm text-muted dark:text-slate-400">
                Drafts save automatically in this browser. Open a question to answer it.
              </p>
            </div>
            <div className="space-y-2 p-5">
              {studentMocks.questions.map((q, i) => (
                <TaskCard
                  key={q.id}
                  compact
                  title={q.title.replace(/^Q\d+\.\s*/, '')}
                  status={taskState(answers[i] ?? '')}
                  onClick={() => navigate(`/student/tasks/${q.id}`)}
                />
              ))}
            </div>
          </Card>

          <aside className="space-y-6">
            <Card padded>
              <div className="flex items-center justify-between">
                <h2 className="text-sm font-semibold">Teacher Feedback</h2>
                <Badge tone={reviewed ? 'emerald' : submitted ? 'amber' : 'slate'}>
                  {reviewed ? 'Scored' : submitted ? 'Waiting' : 'Not submitted'}
                </Badge>
              </div>
              <div className="mt-4 rounded-lg border border-line bg-slate-50 p-4 dark:border-slate-800 dark:bg-slate-950">
                {!submitted && (
                  <p className="text-sm leading-6 text-muted dark:text-slate-400">
                    Submit your answers first. Teacher feedback will appear here after manual review.
                  </p>
                )}
                {submitted && !reviewed && (
                  <div className="flex items-start gap-3">
                    <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-amber-50 text-amber-700 dark:bg-amber-500/10 dark:text-amber-300">
                      <Clock className="h-5 w-5" />
                    </div>
                    <div>
                      <p className="text-sm font-semibold">Teacher review pending</p>
                      <p className="mt-1 text-sm leading-6 text-muted dark:text-slate-400">
                        Your submission has been sent. Refresh this card after your teacher scores it.
                      </p>
                    </div>
                  </div>
                )}
                {reviewed && (
                  <>
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-semibold">Score</span>
                      <Badge tone="brand" className="text-sm">
                        {studentMocks.studentFeedback.score} / {studentMocks.studentFeedback.max}
                      </Badge>
                    </div>
                    <p className="mt-4 text-sm leading-6 text-slate-700 dark:text-slate-300">
                      {studentMocks.studentFeedback.comment}
                    </p>
                    <button
                      className="mt-4 inline-flex w-full items-center justify-center gap-2 rounded-lg border border-line bg-white px-3 py-2.5 text-sm font-semibold text-slate-700 hover:bg-slate-50 dark:border-slate-800 dark:bg-slate-900 dark:text-slate-200 dark:hover:bg-slate-800"
                      onClick={() => showToast('导出 PDF 反馈暂未实现')}
                    >
                      <Download className="h-4 w-4" />
                      Download PDF Feedback
                    </button>
                  </>
                )}
              </div>
            </Card>

            <Card padded>
              <h2 className="text-sm font-semibold">Group Rankings</h2>
              <div className="mt-4">
                <RankingBoard rankings={rankings} myTeam={team} />
              </div>
            </Card>

            <Card padded>
              <h2 className="text-sm font-semibold">My Progress</h2>
              <div className="mt-4 grid grid-cols-2 gap-3">
                <div className="rounded-lg bg-emerald-50 p-4 dark:bg-emerald-500/10">
                  <p className="text-2xl font-bold text-emerald-700 dark:text-emerald-300">{completed}</p>
                  <p className="mt-1 text-xs font-semibold text-emerald-700 dark:text-emerald-300">Completed</p>
                </div>
                <div className="rounded-lg bg-amber-50 p-4 dark:bg-amber-500/10">
                  <p className="text-2xl font-bold text-amber-700 dark:text-amber-300">{pending}</p>
                  <p className="mt-1 text-xs font-semibold text-amber-700 dark:text-amber-300">Pending</p>
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
