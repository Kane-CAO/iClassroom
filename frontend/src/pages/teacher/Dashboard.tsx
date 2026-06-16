import { useCallback, useEffect, useState } from 'react'
import { ChevronLeft, ChevronRight, Plus } from 'lucide-react'
import { useNavigate, useParams } from 'react-router-dom'
import TeacherHeader from '../../components/layout/TeacherHeader'
import Card from '../../components/ui/Card'
import { btnPrimary } from '../../components/ui/buttons'
import { useRoomWebSocket } from '../../hooks/useRoomWebSocket'
import { teacherMocks } from '../../mocks'
import type { RoomWebSocketEventType } from '../../api/websocket'

const DASHBOARD_WS_EVENTS: readonly RoomWebSocketEventType[] = [
  'student_joined',
  'submission_created',
  'score_updated',
  'ranking_updated',
  'room_ended',
]

// /teacher/rooms/:roomCode/dashboard
// 迁移自 docs/prototypes/iClassroom.html 的 #page-home（My Courses 概览）。
export default function Dashboard() {
  const { roomCode = 'ABC123' } = useParams()
  const navigate = useNavigate()
  const { courses, announcements, assignments, calendar } = teacherMocks
  const [refreshVersion, setRefreshVersion] = useState(0)

  const refreshDashboardData = useCallback(() => {
    setRefreshVersion((version) => version + 1)
  }, [])

  useEffect(() => {
    // TODO: replace mock data with room overview/tasks API refetch when backend integration lands.
  }, [refreshVersion])

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

  return (
    <div className="min-h-screen bg-canvas text-ink dark:bg-slate-950 dark:text-slate-100">
      <TeacherHeader roomCode={roomCode} active="courses" />

      <main className="px-8 py-7">
        <div className="grid grid-cols-1 gap-6 lg:grid-cols-[1fr_340px]">
          <div>
            <div className="mb-5 flex items-end justify-between">
              <div>
                <h1 className="text-2xl font-semibold tracking-normal">My Courses</h1>
                <p className="mt-1 text-sm text-muted dark:text-slate-400">
                  Spring 2026 classroom workspace
                </p>
              </div>
              <button className={btnPrimary} onClick={() => navigate('/teacher/create-room')}>
                <Plus className="h-4 w-4" />
                Create Assignment
              </button>
            </div>

            <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 xl:grid-cols-3">
              {courses.map((course) => (
                <button
                  key={course.id}
                  onClick={() => navigate(`/teacher/rooms/${roomCode}/course`)}
                  className="hover-zoom rounded-lg border border-line bg-white text-left shadow-soft transition hover:-translate-y-0.5 hover:border-brand-200 dark:border-slate-800 dark:bg-slate-900"
                >
                  <div className="h-36 overflow-hidden rounded-t-lg bg-slate-200">
                    <img className="h-full w-full object-cover" src={course.cover} alt={`${course.title} cover`} />
                  </div>
                  <div className="p-5">
                    <h2 className="text-lg font-semibold tracking-normal">{course.title}</h2>
                    <p className="mt-2 min-h-[40px] text-sm leading-5 text-muted dark:text-slate-400">
                      {course.summary}
                    </p>
                    <div className="mt-5 grid grid-cols-3 gap-3 text-sm">
                      <div>
                        <p className="font-semibold">{course.students}</p>
                        <p className="text-xs text-muted dark:text-slate-400">Students</p>
                      </div>
                      <div>
                        <p className="font-semibold">{course.assignments}</p>
                        <p className="text-xs text-muted dark:text-slate-400">Assignments</p>
                      </div>
                      <div>
                        <p className="font-semibold">{course.last}</p>
                        <p className="text-xs text-muted dark:text-slate-400">Last Activity</p>
                      </div>
                    </div>
                  </div>
                </button>
              ))}
            </div>
          </div>

          <aside className="space-y-5">
            <Card>
              <div className="border-b border-line px-5 py-4 dark:border-slate-800">
                <h2 className="text-sm font-semibold">Announcements</h2>
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
                <h2 className="text-sm font-semibold">Upcoming Assignments</h2>
              </div>
              <div className="divide-y divide-line dark:divide-slate-800">
                {assignments.slice(0, 3).map((item) => (
                  <button
                    key={item.id}
                    onClick={() => navigate(`/teacher/rooms/${roomCode}/review`)}
                    className="flex w-full items-center gap-3 px-5 py-4 text-left hover:bg-slate-50 dark:hover:bg-slate-800"
                  >
                    <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-brand-50 text-sm font-bold text-brand-700 dark:bg-brand-500/10 dark:text-brand-100">
                      {item.due.replace(/^\w+\s/, '').slice(0, 2)}
                    </span>
                    <span className="min-w-0">
                      <span className="block truncate text-sm font-semibold">{item.title}</span>
                      <span className="block text-xs text-muted dark:text-slate-400">{item.due}</span>
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
    </div>
  )
}
