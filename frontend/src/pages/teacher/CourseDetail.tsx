import { ArrowRight, ClipboardCheck, Copy, Download, FilePlus2, Lock, Pencil, QrCode } from 'lucide-react'
import { useNavigate, useParams } from 'react-router-dom'
import TeacherHeader from '../../components/layout/TeacherHeader'
import Card from '../../components/ui/Card'
import TaskCard from '../../components/TaskCard'
import { btnPrimary, btnSecondary, btnGhostRow } from '../../components/ui/buttons'
import { barToneClass } from '../../components/ui/tones'
import { useToast } from '../../components/ui/useToast'
import { teacherMocks } from '../../mocks'
import type { AssignmentStatus, BadgeTone } from '../../types'

const statusTone: Record<AssignmentStatus, BadgeTone> = {
  Active: 'emerald',
  Closed: 'slate',
  Draft: 'amber',
}

// /teacher/rooms/:roomCode/course
// 迁移自 docs/prototypes/iClassroom.html 的 #page-course（课程详情 / 作业管理）。
export default function CourseDetail() {
  const { roomCode = 'ABC123' } = useParams()
  const navigate = useNavigate()
  const { showToast, ToastView } = useToast()
  const { courses, assignments, reviewSummary } = teacherMocks
  const course = courses[0]

  return (
    <div className="min-h-screen bg-canvas text-ink dark:bg-slate-950 dark:text-slate-100">
      <TeacherHeader roomCode={roomCode} active="assignments" />

      <main className="px-8 py-7">
        <Card padded className="!p-6">
          <div className="flex items-start justify-between gap-4">
            <div>
              <p className="text-sm font-semibold text-brand-600 dark:text-brand-100">Course Detail</p>
              <h1 className="mt-2 text-3xl font-semibold tracking-normal">{course.title}</h1>
              <div className="mt-3 flex flex-wrap items-center gap-4 text-sm text-muted dark:text-slate-400">
                <span>{course.code}</span>
                <span>{course.students} students</span>
                <span>{course.assignments} assignments</span>
              </div>
            </div>
            <button className={btnPrimary} onClick={() => navigate('/teacher/create-room')}>
              <FilePlus2 className="h-4 w-4" />
              Create Assignment
            </button>
          </div>
        </Card>

        <div className="mt-6 grid grid-cols-1 gap-6 lg:grid-cols-[1fr_320px]">
          <div>
            <div className="mb-4 flex items-center justify-between">
              <div>
                <h2 className="text-lg font-semibold">Assignments</h2>
                <p className="text-sm text-muted dark:text-slate-400">
                  Create, edit, duplicate, close, and review submissions.
                </p>
              </div>
              <button className={btnSecondary} onClick={() => showToast('Draft duplicated')}>
                <Copy className="h-4 w-4" />
                Bulk Duplicate
              </button>
            </div>

            <div className="space-y-4">
              {assignments.map((item) => (
                <TaskCard
                  key={item.id}
                  title={item.title}
                  status={{ label: item.status, tone: statusTone[item.status] }}
                  due={item.due}
                  submitted={item.count}
                  actions={
                    <>
                      <button className={btnSecondary} onClick={() => navigate('/teacher/create-room')}>
                        <Pencil className="h-4 w-4" />
                        Edit
                      </button>
                      <button className={btnSecondary} onClick={() => showToast('Assignment duplicated')}>
                        <Copy className="h-4 w-4" />
                        Duplicate
                      </button>
                      <button className={btnSecondary} onClick={() => showToast('Assignment closed')}>
                        <Lock className="h-4 w-4" />
                        Close
                      </button>
                      <button
                        className={btnPrimary.replace('px-4 py-2.5', 'px-3 py-2')}
                        onClick={() => navigate(`/teacher/rooms/${roomCode}/review`)}
                      >
                        <ClipboardCheck className="h-4 w-4" />
                        Review
                      </button>
                    </>
                  }
                />
              ))}
            </div>
          </div>

          <aside className="space-y-5">
            <Card padded>
              <h3 className="text-sm font-semibold">Review Summary</h3>
              <div className="mt-5 space-y-4">
                {reviewSummary.map((stat) => {
                  const pct = Math.round((stat.value / stat.total) * 100)
                  return (
                    <div key={stat.label}>
                      <div className="mb-2 flex justify-between text-xs font-semibold">
                        <span>{stat.label}</span>
                        <span>
                          {stat.value} / {stat.total}
                        </span>
                      </div>
                      <div className="h-2 rounded-full bg-slate-100 dark:bg-slate-800">
                        <div className={`h-2 rounded-full ${barToneClass[stat.tone]}`} style={{ width: `${pct}%` }} />
                      </div>
                    </div>
                  )
                })}
              </div>
            </Card>

            <Card padded>
              <h3 className="text-sm font-semibold">Quick Actions</h3>
              <div className="mt-4 grid gap-2">
                <button className={btnGhostRow} onClick={() => navigate(`/teacher/rooms/${roomCode}/review`)}>
                  Open Review Queue <ArrowRight className="h-4 w-4" />
                </button>
                <button className={btnGhostRow} onClick={() => navigate(`/teacher/rooms/${roomCode}/display`)}>
                  Show Join QR <QrCode className="h-4 w-4" />
                </button>
                <button className={btnGhostRow} onClick={() => showToast('导出功能暂未实现')}>
                  Export Submissions <Download className="h-4 w-4" />
                </button>
              </div>
            </Card>
          </aside>
        </div>
      </main>

      <ToastView />
    </div>
  )
}
