import { useState } from 'react'
import { LogIn } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import StudentHeader from '../../components/layout/StudentHeader'
import Card from '../../components/ui/Card'
import GroupTable from '../../components/GroupTable'
import { btnGradient } from '../../components/ui/buttons'
import { useToast } from '../../components/ui/useToast'
import { useStudentSession } from '../../hooks/useStudentSession'
import { studentMocks } from '../../mocks'

// /student
// 迁移自 docs/prototypes/student.html 的 #joinScreen（扫码加入 / 选择小组）。
export default function StudentEntry() {
  const navigate = useNavigate()
  const { showToast, ToastView } = useToast()
  const { identity, join } = useStudentSession()
  const room = studentMocks.studentRoom
  const groups = studentMocks.studentGroups

  const [nickname, setNickname] = useState(identity?.name ?? studentMocks.defaultStudentIdentity.name)
  const [groupId, setGroupId] = useState(groups[0].id)

  const onJoin = () => {
    const name = nickname.trim()
    if (!name) {
      showToast('Please enter a nickname')
      return
    }
    const team = groups.find((g) => g.id === groupId)?.name ?? groups[0].name
    join({ name, team })
    showToast('Joined classroom')
    setTimeout(() => navigate('/student/classroom'), 500)
  }

  const stats = [
    { value: room.questionsCount, label: 'Questions', tone: 'bg-brand-50 text-brand-700 dark:bg-brand-500/10 dark:text-brand-100' },
    { value: room.minutesLeft, label: 'Minutes left', tone: 'bg-violetx-50 text-violetx-600 dark:bg-violet-500/10 dark:text-violet-300' },
    { value: room.scoreRange, label: 'Manual score', tone: 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300' },
  ]

  return (
    <div className="min-h-screen bg-soft text-ink dark:bg-slate-950 dark:text-slate-100">
      <StudentHeader roomCode={room.roomCode} />

      <main className="mx-auto max-w-[1180px] px-6 py-8 sm:px-8">
        <section className="grid grid-cols-1 gap-8 lg:grid-cols-[1fr_420px]">
          <Card padded className="!p-8">
            <p className="text-sm font-semibold text-brand-600 dark:text-brand-100">Scan-to-join classroom</p>
            <h1 className="mt-3 text-4xl font-semibold tracking-normal">{room.course}</h1>
            <p className="mt-4 max-w-2xl text-sm leading-7 text-muted dark:text-slate-400">
              You joined through a QR invitation. Enter a display name, choose your group, and start the current
              assignment. Your draft is saved in this browser during the live room.
            </p>
            <div className="mt-8 grid grid-cols-3 gap-4">
              {stats.map((s) => (
                <div key={s.label} className={`rounded-lg p-4 ${s.tone}`}>
                  <p className="text-2xl font-bold">{s.value}</p>
                  <p className="mt-1 text-xs font-semibold">{s.label}</p>
                </div>
              ))}
            </div>
          </Card>

          <Card padded className="!p-6">
            <h2 className="text-lg font-semibold">Join Room</h2>
            <p className="mt-1 text-sm text-muted dark:text-slate-400">No account required.</p>
            <label className="mt-5 block">
              <span className="text-sm font-semibold">Nickname</span>
              <input
                value={nickname}
                maxLength={24}
                onChange={(e) => setNickname(e.target.value)}
                className="mt-2 w-full rounded-lg border border-line bg-white px-3 py-2.5 text-sm outline-none focus:border-brand-500 dark:border-slate-800 dark:bg-slate-950"
              />
            </label>
            <div className="mt-5">
              <p className="text-sm font-semibold">Choose Group</p>
              <div className="mt-2">
                <GroupTable groups={groups} selectable value={groupId} onChange={setGroupId} />
              </div>
            </div>
            <button className={`${btnGradient} mt-6 w-full`} onClick={onJoin}>
              <LogIn className="h-4 w-4" />
              Enter Classroom
            </button>
          </Card>
        </section>
      </main>

      <ToastView />
    </div>
  )
}
