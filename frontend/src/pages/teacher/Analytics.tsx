import { useParams } from 'react-router-dom'
import TeacherHeader from '../../components/layout/TeacherHeader'
import Card from '../../components/ui/Card'
import { barToneClass } from '../../components/ui/tones'
import { teacherMocks } from '../../mocks'

// /teacher/rooms/:roomCode/analytics
// 注意：docs/prototypes 中没有独立的“数据看板”页面。此页复用讲师端 mock
// （Review Summary + 课程统计）以同样的视觉风格组合而成，待后续有原型再细化。
export default function Analytics() {
  const { roomCode = 'ABC123' } = useParams()
  const { reviewSummary, courses } = teacherMocks

  return (
    <div className="min-h-screen bg-canvas text-ink dark:bg-slate-950 dark:text-slate-100">
      <TeacherHeader roomCode={roomCode} active="courses" />

      <main className="px-8 py-7">
        <h1 className="text-2xl font-semibold tracking-normal">Analytics</h1>
        <p className="mt-1 text-sm text-muted dark:text-slate-400">
          复用提交 / 批改进度与课程概览数据（无对应原型，组合展示）。
        </p>

        <div className="mt-6 grid grid-cols-1 gap-6 lg:grid-cols-[1fr_340px]">
          <Card padded>
            <h2 className="text-sm font-semibold">Review Progress</h2>
            <div className="mt-5 space-y-4">
              {reviewSummary.map((stat) => {
                const pct = Math.round((stat.value / stat.total) * 100)
                return (
                  <div key={stat.label}>
                    <div className="mb-2 flex justify-between text-xs font-semibold">
                      <span>{stat.label}</span>
                      <span>
                        {stat.value} / {stat.total} · {pct}%
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
            <h2 className="text-sm font-semibold">Courses Overview</h2>
            <div className="mt-4 space-y-3">
              {courses.map((course) => (
                <div
                  key={course.id}
                  className="flex items-center justify-between rounded-lg bg-slate-50 px-3 py-3 dark:bg-slate-950"
                >
                  <span className="text-sm font-semibold">{course.title}</span>
                  <span className="text-xs text-muted dark:text-slate-400">
                    {course.students} students · {course.assignments} assignments
                  </span>
                </div>
              ))}
            </div>
          </Card>
        </div>
      </main>
    </div>
  )
}
