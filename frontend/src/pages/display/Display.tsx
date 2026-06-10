import { useParams } from 'react-router-dom'
import QRCodePanel from '../../components/QRCodePanel'
import GroupTable from '../../components/GroupTable'
import RankingBoard from '../../components/RankingBoard'
import TaskCard from '../../components/TaskCard'
import Card from '../../components/ui/Card'
import Badge from '../../components/ui/Badge'
import { useCountdown } from '../../hooks/useCountdown'
import { studentMocks } from '../../mocks'

// /teacher/rooms/:roomCode/display
// 说明：docs/prototypes 中没有独立的大屏 HTML。本页以相同视觉风格、复用
// RoomInfoCard / QRCodePanel / GroupTable / RankingBoard / TaskCard 等组件，
// 基于同一份 mock 数据组合出一个投屏看板。
export default function Display() {
  const { roomCode } = useParams()
  const room = studentMocks.studentRoom
  const code = roomCode ?? room.roomCode
  const timer = useCountdown(studentMocks.remainingSeconds, true)

  const base = import.meta.env.VITE_STUDENT_BASE_URL ?? `${window.location.origin}/student`
  const joinUrl = `${base}?room=${code}`

  return (
    <div className="min-h-screen bg-canvas px-6 py-6 text-ink dark:bg-slate-950 dark:text-slate-100 sm:px-10 sm:py-8">
      {/* 顶部标题栏 */}
      <header className="mb-6 flex flex-wrap items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <span className="flex h-11 w-11 items-center justify-center rounded-xl bg-gradient-to-br from-brand-600 to-violetx-600 text-base font-bold text-white">
            iC
          </span>
          <div>
            <p className="text-sm font-semibold text-brand-600 dark:text-brand-100">{room.course}</p>
            <h1 className="text-3xl font-semibold tracking-normal sm:text-4xl">{room.assignment}</h1>
          </div>
        </div>
        <div className="flex items-center gap-4">
          <Badge tone="emerald">Live · Room {code}</Badge>
          <div className="rounded-lg bg-brand-50 px-6 py-3 text-center dark:bg-brand-500/10">
            <p className="text-xs font-bold uppercase text-brand-700 dark:text-brand-100">Remaining</p>
            <p className="font-mono text-4xl font-bold tabular-nums text-brand-700 dark:text-brand-100">
              {timer.mmss}
            </p>
          </div>
        </div>
      </header>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-[380px_1fr]">
        {/* 左列：加入二维码 + 小组 */}
        <div className="space-y-6">
          <QRCodePanel roomCode={code} joinUrl={joinUrl} subtitle={`${room.teacher} · ${room.scoreRange} score`} />
          <Card padded>
            <h2 className="text-sm font-semibold">Groups</h2>
            <div className="mt-4">
              <GroupTable groups={studentMocks.studentGroups} />
            </div>
          </Card>
        </div>

        {/* 右列：排行榜 + 任务概览 */}
        <div className="space-y-6">
          <Card padded>
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold">Group Rankings</h2>
              <span className="text-sm text-muted dark:text-slate-400">{room.questionsCount} questions</span>
            </div>
            <div className="mt-5 text-base">
              <RankingBoard rankings={studentMocks.rankingsBeforeReview} withBars />
            </div>
          </Card>

          <Card padded>
            <h2 className="text-lg font-semibold">Assignment Tasks</h2>
            <div className="mt-4 space-y-2">
              {studentMocks.questions.map((q) => (
                <TaskCard
                  key={q.id}
                  compact
                  title={q.title}
                  status={{ label: 'In Progress', tone: 'amber' }}
                />
              ))}
            </div>
          </Card>
        </div>
      </div>
    </div>
  )
}
