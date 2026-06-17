import { useCallback, useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { getDisplayState } from '../../api/display'
import type { RoomWebSocketEventType } from '../../api/websocket'
import QRCodePanel from '../../components/QRCodePanel'
import GroupTable from '../../components/GroupTable'
import RankingBoard from '../../components/RankingBoard'
import TaskCard from '../../components/TaskCard'
import Card from '../../components/ui/Card'
import Badge from '../../components/ui/Badge'
import { useToast } from '../../components/ui/useToast'
import { useRoomWebSocket } from '../../hooks/useRoomWebSocket'
import type { DisplayFeaturedAnswer, DisplayState, DisplayTask } from '../../types/api'
import type { BadgeTone, RankingVM, StudentGroupVM } from '../../types'

const DISPLAY_WS_EVENTS: readonly RoomWebSocketEventType[] = [
  'ranking_updated',
  'featured_answer_updated',
  'submission_created',
  'room_ended',
]

export default function Display() {
  const { roomCode } = useParams()
  const code = roomCode ?? ''
  const { showToast, ToastView } = useToast()
  const [display, setDisplay] = useState<DisplayState | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [roomEnded, setRoomEnded] = useState(false)
  const [refreshVersion, setRefreshVersion] = useState(0)

  const refreshDisplayData = useCallback(() => {
    setRefreshVersion((version) => version + 1)
  }, [])

  useEffect(() => {
    let active = true

    async function loadDisplay() {
      if (!code) {
        setLoading(false)
        setError('缺少房间码')
        return
      }

      setLoading(true)
      setError('')
      try {
        const data = await getDisplayState(code)
        if (!active) {
          return
        }
        setDisplay(data)
        setRoomEnded(data.status === 'ended')
      } catch (err: unknown) {
        if (!active) {
          return
        }
        const message = err instanceof Error ? err.message : '加载大屏数据失败'
        setDisplay(null)
        setError(message)
        showToast(message)
      } finally {
        if (active) {
          setLoading(false)
        }
      }
    }

    void loadDisplay()

    return () => {
      active = false
    }
  }, [code, refreshVersion, showToast])

  useRoomWebSocket({
    roomCode: code,
    role: 'display',
    enabled: Boolean(code),
    onEvent: (event) => {
      if (DISPLAY_WS_EVENTS.includes(event.type)) {
        if (event.type === 'room_ended') {
          setRoomEnded(true)
          showToast('课堂已结束')
        }
        refreshDisplayData()
      }
    },
    onReconnect: refreshDisplayData,
  })

  const base = import.meta.env.VITE_STUDENT_BASE_URL ?? `${window.location.origin}/student`
  const joinUrl = `${base}?room=${code}`
  const groups = mapGroups(display)
  const rankings = mapRankings(display)
  const currentTask = display?.currentTask ?? null
  const featuredAnswers = display?.featuredAnswers ?? []
  const completionLabel = currentTask ? `${Math.round(currentTask.completionRate * 100)}%` : roomEnded ? '已结束' : '--'
  const roomTitle = display?.title ?? '课堂大屏'

  if (loading && !display) {
    return (
      <div className="min-h-screen bg-canvas px-6 py-6 text-ink dark:bg-slate-950 dark:text-slate-100 sm:px-10 sm:py-8">
        <Card padded className="mx-auto mt-20 max-w-xl text-center">
          <p className="text-sm font-semibold text-brand-600 dark:text-brand-100">正在加载大屏</p>
          <p className="mt-2 text-sm text-muted dark:text-slate-400">正在获取实时课堂数据...</p>
        </Card>
        <ToastView />
      </div>
    )
  }

  if (error && !display) {
    return (
      <div className="min-h-screen bg-canvas px-6 py-6 text-ink dark:bg-slate-950 dark:text-slate-100 sm:px-10 sm:py-8">
        <Card padded className="mx-auto mt-20 max-w-xl text-center">
          <Badge tone="rose">大屏不可用</Badge>
          <h1 className="mt-4 text-2xl font-semibold tracking-normal">{code || '无房间码'}</h1>
          <p className="mt-2 text-sm text-muted dark:text-slate-400">{error}</p>
        </Card>
        <ToastView />
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-canvas px-6 py-6 text-ink dark:bg-slate-950 dark:text-slate-100 sm:px-10 sm:py-8">
      {/* 顶部标题栏 */}
      <header className="mb-6 flex flex-wrap items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <span className="flex h-11 w-11 items-center justify-center rounded-xl bg-gradient-to-br from-brand-600 to-violetx-600 text-base font-bold text-white">
            iC
          </span>
          <div>
            <p className="text-sm font-semibold text-brand-600 dark:text-brand-100">实时课堂</p>
            <h1 className="text-3xl font-semibold tracking-normal sm:text-4xl">{roomTitle}</h1>
          </div>
        </div>
        <div className="flex items-center gap-4">
          <Badge tone={roomEnded ? 'slate' : 'emerald'}>
            {roomEnded ? '已结束' : '进行中'} · 房间 {code}
          </Badge>
          <div className="rounded-lg bg-brand-50 px-6 py-3 text-center dark:bg-brand-500/10">
            <p className="text-xs font-bold uppercase text-brand-700 dark:text-brand-100">完成率</p>
            <p className="font-mono text-4xl font-bold tabular-nums text-brand-700 dark:text-brand-100">
              {completionLabel}
            </p>
          </div>
        </div>
      </header>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-[380px_1fr]">
        {/* 左列：加入二维码 + 小组 */}
        <div className="space-y-6">
          <QRCodePanel
            roomCode={code}
            joinUrl={joinUrl}
            subtitle={`${groups.length} 个小组 · ${rankings.length} 个排行项`}
          />
          <Card padded>
            <h2 className="text-sm font-semibold">小组</h2>
            <div className="mt-4">
              {groups.length > 0 ? (
                <GroupTable groups={groups} />
              ) : (
                <p className="rounded-lg border border-dashed border-line px-3 py-4 text-sm text-muted dark:border-slate-800 dark:text-slate-400">
                  暂无小组。
                </p>
              )}
            </div>
          </Card>
        </div>

        {/* 右列：排行榜 + 任务概览 */}
        <div className="space-y-6">
          <Card padded>
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold">小组排行榜</h2>
              <span className="text-sm text-muted dark:text-slate-400">
                {loading ? '刷新中...' : `${rankings.length} 个小组`}
              </span>
            </div>
            <div className="mt-5 text-base">
              {rankings.length > 0 ? (
                <RankingBoard rankings={rankings} withBars />
              ) : (
                <p className="rounded-lg border border-dashed border-line px-3 py-4 text-sm text-muted dark:border-slate-800 dark:text-slate-400">
                  暂无排行数据。
                </p>
              )}
            </div>
          </Card>

          <Card padded>
            <h2 className="text-lg font-semibold">当前任务</h2>
            <div className="mt-4 space-y-2">
              {currentTask ? (
                <TaskCard
                  compact
                  title={`${currentTask.title} · 已提交 ${currentTask.submittedCount} / ${currentTask.targetStudentCount}`}
                  status={taskStatus(currentTask, roomEnded)}
                />
              ) : (
                <p className="rounded-lg border border-dashed border-line px-3 py-4 text-sm text-muted dark:border-slate-800 dark:text-slate-400">
                  {roomEnded ? '课堂已结束。' : '暂无进行中的任务。'}
                </p>
              )}
            </div>
          </Card>

          <Card padded>
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold">精选答案</h2>
              <span className="text-sm text-muted dark:text-slate-400">已展示 {featuredAnswers.length} 条</span>
            </div>
            <div className="mt-4 space-y-3">
              {featuredAnswers.length > 0 ? (
                featuredAnswers.map((answer) => (
                  <FeaturedAnswerRow key={answer.featuredId} answer={answer} />
                ))
              ) : (
                <p className="rounded-lg border border-dashed border-line px-3 py-4 text-sm text-muted dark:border-slate-800 dark:text-slate-400">
                  暂无精选答案。
                </p>
              )}
            </div>
          </Card>
        </div>
      </div>
      <ToastView />
    </div>
  )
}

function mapGroups(display: DisplayState | null): StudentGroupVM[] {
  if (!display) {
    return []
  }
  return display.groups.map((group) => ({
    id: String(group.groupId),
    name: group.groupName,
    filled: group.currentCount,
    capacity: group.capacity,
    full: group.currentCount >= group.capacity,
  }))
}

function mapRankings(display: DisplayState | null): RankingVM[] {
  if (!display) {
    return []
  }
  return display.ranking.map((entry) => ({
    team: entry.groupName,
    score: entry.scoreTotal,
  }))
}

function taskStatus(task: DisplayTask, roomEnded: boolean): { label: string; tone: BadgeTone } {
  if (roomEnded) {
    return { label: '已结束', tone: 'slate' }
  }
  if (task.targetStudentCount > 0 && task.submittedCount >= task.targetStudentCount) {
    return { label: '已完成', tone: 'emerald' }
  }
  return { label: '进行中', tone: 'amber' }
}

function FeaturedAnswerRow({ answer }: { answer: DisplayFeaturedAnswer }) {
  return (
    <div className="rounded-lg border border-line bg-slate-50 px-4 py-3 dark:border-slate-800 dark:bg-slate-950">
      <div className="mb-2 flex flex-wrap items-center justify-between gap-2">
        <Badge tone={answer.displayMode === 'showGroup' ? 'sky' : 'slate'}>
          {answer.displayMode === 'showGroup' ? answer.groupName ?? '小组答案' : '匿名'}
        </Badge>
        <span className="text-xs font-semibold text-muted dark:text-slate-400">
          {answer.score == null ? '待评分' : `${answer.score} 分`}
        </span>
      </div>
      <p className="line-clamp-3 text-sm leading-6 text-ink dark:text-slate-100">{answer.contentText}</p>
      <p className="mt-2 text-xs text-muted dark:text-slate-500">{formatTime(answer.submittedAt)}</p>
    </div>
  )
}

function formatTime(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return new Intl.DateTimeFormat('zh-CN', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}
