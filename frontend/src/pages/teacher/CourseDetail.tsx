import { useCallback, useEffect, useState } from 'react'
import { ArrowRight, ClipboardCheck, Copy, FilePlus2, Lock, Pause, Pencil, QrCode, X } from 'lucide-react'
import { useNavigate, useParams } from 'react-router-dom'
import TeacherHeader from '../../components/layout/TeacherHeader'
import TeacherRoomActions from '../../components/teacher/TeacherRoomActions'
import Card from '../../components/ui/Card'
import TaskCard from '../../components/TaskCard'
import { btnPrimary, btnSecondary, btnGhostRow } from '../../components/ui/buttons'
import { barToneClass } from '../../components/ui/tones'
import { useToast } from '../../components/ui/useToast'
import { useRoomWebSocket } from '../../hooks/useRoomWebSocket'
import { useTeacherSession } from '../../hooks/useTeacherSession'
import { ApiRequestError } from '../../api/client'
import { getRoom, getRoomOverview } from '../../api/rooms'
import { closeTask, createTask, listTeacherTasks, pauseTask } from '../../api/tasks'
import type { RoomWebSocketEventType } from '../../api/websocket'
import type { BadgeTone } from '../../types'
import type { CreateTaskRequest, Room, RoomOverview, Task, TaskStatus, TaskTargetType } from '../../types/api'

const COURSE_WS_EVENTS: readonly RoomWebSocketEventType[] = [
  'student_joined',
  'task_published',
  'task_paused',
  'task_closed',
  'room_ended',
]

const statusTone: Record<TaskStatus, BadgeTone> = {
  published: 'emerald',
  paused: 'amber',
  closed: 'slate',
}

const inputClass =
  'mt-2 w-full rounded-lg border border-line bg-white px-3 py-2.5 text-sm outline-none focus:border-brand-500 disabled:opacity-70 dark:border-slate-800 dark:bg-slate-950'

interface TaskFormState {
  title: string
  description: string
  deadlineAt: string
  targetType: TaskTargetType
  targetGroupIds: number[]
}

const emptyTaskForm: TaskFormState = {
  title: '',
  description: '',
  deadlineAt: '',
  targetType: 'all',
  targetGroupIds: [],
}

// /teacher/rooms/:roomCode/course
// 迁移自 docs/prototypes/iClassroom.html 的 #page-course（课程详情 / 作业管理）。
export default function CourseDetail() {
  const { roomCode = 'ABC123' } = useParams()
  const navigate = useNavigate()
  const { showToast, ToastView } = useToast()
  const { teacherToken, clear } = useTeacherSession()
  const [room, setRoom] = useState<Room | null>(null)
  const [overview, setOverview] = useState<RoomOverview | null>(null)
  const [tasks, setTasks] = useState<Task[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [createOpen, setCreateOpen] = useState(false)
  const [taskForm, setTaskForm] = useState<TaskFormState>(emptyTaskForm)
  const [formError, setFormError] = useState<string | null>(null)
  const [savingTask, setSavingTask] = useState(false)
  const [updatingTaskId, setUpdatingTaskId] = useState<number | null>(null)
  const roomStatus = room?.status ?? overview?.status ?? 'created'
  const isRoomEnded = roomStatus === 'ended'

  const refreshCourseData = useCallback(() => {
    setLoading(true)
  }, [])

  const openCreateTask = () => {
    if (isRoomEnded) {
      showToast('课堂已结束，不能再创建新任务。')
      return
    }
    setTaskForm({
      ...emptyTaskForm,
      deadlineAt: defaultDeadlineInputValue(),
    })
    setFormError(null)
    setCreateOpen(true)
  }

  const submitTaskForm = async () => {
    const nextTitle = taskForm.title.trim()
    const nextDescription = taskForm.description.trim()

    if (!teacherToken) {
      setFormError('老师会话缺失，请重新创建课堂。')
      return
    }
    if (isRoomEnded) {
      setFormError('课堂已结束，不能再创建新任务。')
      return
    }
    if (!nextTitle) {
      setFormError('请输入任务标题。')
      return
    }
    if (!nextDescription) {
      setFormError('请输入任务说明。')
      return
    }
    if (!taskForm.deadlineAt) {
      setFormError('请选择截止时间。')
      return
    }
    if (taskForm.targetType === 'groups' && taskForm.targetGroupIds.length === 0) {
      setFormError('请至少选择一个目标小组。')
      return
    }

    const deadlineAt = toISOString(taskForm.deadlineAt)
    if (!deadlineAt) {
      setFormError('请选择有效的截止时间。')
      return
    }

    const body: CreateTaskRequest = {
      title: nextTitle,
      description: nextDescription,
      attachmentUrl: '',
      deadlineAt,
      targetType: taskForm.targetType,
      targetGroupIds: taskForm.targetType === 'groups' ? taskForm.targetGroupIds : [],
    }

    setSavingTask(true)
    setFormError(null)

    try {
      await createTask(roomCode, body, { teacherToken })
      showToast('任务已发布')
      setCreateOpen(false)
      refreshCourseData()
    } catch (err) {
      const message = getErrorMessage(err, '发布任务失败。')
      setFormError(message)
      showToast(message)
    } finally {
      setSavingTask(false)
    }
  }

  const updateTaskStatus = async (task: Task, action: 'pause' | 'close') => {
    if (!teacherToken) {
      setError('老师会话缺失，请重新创建课堂。')
      showToast('老师会话缺失')
      return
    }
    if (isRoomEnded) {
      showToast('课堂已结束，不能修改任务状态。')
      return
    }

    setUpdatingTaskId(task.taskId)
    setError(null)

    try {
      if (action === 'pause') {
        await pauseTask(task.taskId, { teacherToken })
        showToast('任务已暂停')
      } else {
        await closeTask(task.taskId, { teacherToken })
        showToast('任务已关闭')
      }
      refreshCourseData()
    } catch (err) {
      const message = getErrorMessage(err, action === 'pause' ? '暂停任务失败。' : '关闭任务失败。')
      setError(message)
      showToast(message)
    } finally {
      setUpdatingTaskId(null)
    }
  }

  const handleRoomActionError = useCallback((message: string) => {
    setError(message)
    showToast(message)
  }, [showToast])

  const handleRoomActionSuccess = useCallback((message: string) => {
    showToast(message)
  }, [showToast])

  useEffect(() => {
    let cancelled = false

    async function loadCourseData() {
      setError(null)

      if (!teacherToken) {
        setRoom(null)
        setOverview(null)
        setTasks([])
        setError('老师会话缺失，请重新创建课堂。')
        setLoading(false)
        return
      }

      try {
        const auth = { teacherToken }
        const [roomData, overviewData, taskData] = await Promise.all([
          getRoom(roomCode, auth),
          getRoomOverview(roomCode, auth),
          listTeacherTasks(roomCode, auth),
        ])

        if (cancelled) {
          return
        }

        setRoom(roomData)
        setOverview(overviewData)
        setTasks(taskData)
      } catch (err) {
        if (cancelled) {
          return
        }
        if (isAuthError(err)) {
          clear()
          setError('老师凭证无效或已过期，请重新创建课堂。')
        } else {
          setError(getErrorMessage(err, '加载课堂数据失败。'))
        }
        setRoom(null)
        setOverview(null)
        setTasks([])
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    if (loading) {
      loadCourseData()
    }

    return () => {
      cancelled = true
    }
  }, [clear, loading, roomCode, teacherToken])

  useRoomWebSocket({
    roomCode,
    role: 'teacher',
    onEvent: (event) => {
      if (COURSE_WS_EVENTS.includes(event.type)) {
        refreshCourseData()
      }
    },
    onReconnect: refreshCourseData,
  })

  const roomTitle = room?.title ?? overview?.title ?? '课堂'
  const studentCount = overview?.studentCount ?? 0
  const groupCount = overview?.groups.length ?? room?.groupCount ?? 0
  const taskCount = tasks.length

  return (
    <div className="min-h-screen bg-canvas text-ink dark:bg-slate-950 dark:text-slate-100">
      <TeacherHeader roomCode={roomCode} active="assignments" />

      <main className="px-8 py-7">
        <Card padded className="!p-6">
          <div className="flex items-start justify-between gap-4">
            <div>
              <p className="text-sm font-semibold text-brand-600 dark:text-brand-100">课程详情</p>
              <h1 className="mt-2 text-3xl font-semibold tracking-normal">{roomTitle}</h1>
              <div className="mt-3 flex flex-wrap items-center gap-4 text-sm text-muted dark:text-slate-400">
                <span>房间 {roomCode}</span>
                <span>{studentCount} 名学生</span>
                <span>{groupCount} 个小组</span>
                <span>{taskCount} 个任务</span>
                <span>{roomStatusLabel(roomStatus)}</span>
              </div>
            </div>
            <div className="flex flex-wrap justify-end gap-2">
              <button className={btnPrimary} disabled={isRoomEnded} onClick={openCreateTask}>
                <FilePlus2 className="h-4 w-4" />
                创建任务
              </button>
              <TeacherRoomActions
                roomCode={roomCode}
                teacherToken={teacherToken}
                roomEnded={isRoomEnded}
                showExport={false}
                onEnded={refreshCourseData}
                onError={handleRoomActionError}
                onSuccess={handleRoomActionSuccess}
              />
            </div>
          </div>
        </Card>

        {error && (
          <Card className="mt-6 border-rose-200 bg-rose-50 p-4 text-rose-700 dark:border-rose-500/20 dark:bg-rose-500/10 dark:text-rose-300">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <p className="text-sm font-semibold">{error}</p>
              <button className={btnPrimary.replace('px-4 py-2.5', 'px-3 py-2')} onClick={() => navigate('/teacher/create-room')}>
                <FilePlus2 className="h-4 w-4" />
                创建课堂
              </button>
            </div>
          </Card>
        )}

        <div className="mt-6 grid grid-cols-1 gap-6 lg:grid-cols-[1fr_320px]">
          <div>
            <div className="mb-4 flex items-center justify-between">
              <div>
                <h2 className="text-lg font-semibold">任务列表</h2>
                <p className="text-sm text-muted dark:text-slate-400">
                  发布、暂停、关闭任务，并查看学生提交与批改结果。
                </p>
              </div>
              <button
                className={btnSecondary}
                disabled={isRoomEnded}
                onClick={() => showToast(isRoomEnded ? '课堂已结束' : '草稿已复制')}
              >
                <Copy className="h-4 w-4" />
                批量复制
              </button>
            </div>

            <div className="space-y-4">
              {loading && (
                <Card className="p-5">
                  <p className="text-sm text-muted dark:text-slate-400">正在加载任务...</p>
                </Card>
              )}

              {!loading && !error && tasks.length === 0 && (
                <Card className="p-5">
                  <p className="text-sm font-semibold">暂无任务</p>
                  <p className="mt-1 text-sm text-muted dark:text-slate-400">
                    创建一个任务后，就可以开始收集学生提交。
                  </p>
                </Card>
              )}

              {!loading && !error && tasks.map((item) => (
                <TaskCard
                  key={item.taskId}
                  title={item.title}
                  status={{ label: taskStatusLabel(item.status), tone: statusTone[item.status] }}
                  due={formatDateTime(item.deadlineAt)}
                  submitted={`${item.submittedCount ?? 0} / ${item.targetStudentCount ?? 0}`}
                  actions={
                    <>
                      <button className={btnSecondary} disabled={isRoomEnded} onClick={() => navigate('/teacher/create-room')}>
                        <Pencil className="h-4 w-4" />
                        编辑
                      </button>
                      <button className={btnSecondary} disabled={isRoomEnded} onClick={() => showToast('任务已复制')}>
                        <Copy className="h-4 w-4" />
                        复制
                      </button>
                      <button
                        className={btnSecondary}
                        disabled={isRoomEnded || item.status !== 'published' || updatingTaskId === item.taskId}
                        onClick={() => updateTaskStatus(item, 'pause')}
                      >
                        <Pause className="h-4 w-4" />
                        暂停
                      </button>
                      <button
                        className={btnSecondary}
                        disabled={isRoomEnded || item.status === 'closed' || updatingTaskId === item.taskId}
                        onClick={() => updateTaskStatus(item, 'close')}
                      >
                        <Lock className="h-4 w-4" />
                        {updatingTaskId === item.taskId ? '保存中' : '关闭'}
                      </button>
                      <button
                        className={btnPrimary.replace('px-4 py-2.5', 'px-3 py-2')}
                        onClick={() => navigate(`/teacher/rooms/${roomCode}/review`)}
                      >
                        <ClipboardCheck className="h-4 w-4" />
                        批改
                      </button>
                    </>
                  }
                />
              ))}
            </div>
          </div>

          <aside className="space-y-5">
            <Card padded>
              <h3 className="text-sm font-semibold">批改概览</h3>
              <div className="mt-5 space-y-4">
                {buildReviewSummary(tasks, studentCount).map((stat) => {
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
              <h3 className="text-sm font-semibold">快捷操作</h3>
              <div className="mt-4 grid gap-2">
                <button className={btnGhostRow} disabled={isRoomEnded} onClick={openCreateTask}>
                  发布任务 <FilePlus2 className="h-4 w-4" />
                </button>
                <button className={btnGhostRow} onClick={() => navigate(`/teacher/rooms/${roomCode}/review`)}>
                  打开批改列表 <ArrowRight className="h-4 w-4" />
                </button>
                <button className={btnGhostRow} onClick={() => navigate(`/teacher/rooms/${roomCode}/display`)}>
                  显示加入二维码 <QrCode className="h-4 w-4" />
                </button>
                <TeacherRoomActions
                  roomCode={roomCode}
                  teacherToken={teacherToken}
                  roomEnded={isRoomEnded}
                  className="grid gap-2"
                  buttonClassName={btnGhostRow}
                  showEnd={false}
                  onEnded={refreshCourseData}
                  onError={handleRoomActionError}
                  onSuccess={handleRoomActionSuccess}
                />
              </div>
            </Card>
          </aside>
        </div>
      </main>

      {createOpen && (
        <div className="fixed inset-0 z-40 flex items-center justify-center bg-slate-950/40 px-4 py-6">
          <Card className="max-h-[92vh] w-full max-w-2xl overflow-y-auto bg-white p-6 dark:bg-slate-900">
            <div className="mb-5 flex items-start justify-between gap-4">
              <div>
                <p className="text-sm font-semibold text-brand-600 dark:text-brand-100">发布任务</p>
                <h2 className="mt-1 text-xl font-semibold tracking-normal">{roomTitle}</h2>
              </div>
              <button
                type="button"
                className="rounded-lg border border-line p-2 text-slate-500 hover:bg-slate-50 dark:border-slate-800 dark:hover:bg-slate-800"
                onClick={() => setCreateOpen(false)}
                aria-label="关闭发布任务弹窗"
              >
                <X className="h-4 w-4" />
              </button>
            </div>

            {formError && (
              <div className="mb-5 rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm font-semibold text-rose-700 dark:border-rose-500/20 dark:bg-rose-500/10 dark:text-rose-300">
                {formError}
              </div>
            )}

            <form
              className="grid gap-4"
              onSubmit={(event) => {
                event.preventDefault()
                submitTaskForm()
              }}
            >
              <label className="block">
                <span className="text-sm font-semibold">标题</span>
                <input
                  value={taskForm.title}
                  disabled={savingTask}
                  onChange={(event) => setTaskForm((prev) => ({ ...prev, title: event.target.value }))}
                  className={inputClass}
                />
              </label>

              <label className="block">
                <span className="text-sm font-semibold">说明</span>
                <textarea
                  value={taskForm.description}
                  disabled={savingTask}
                  onChange={(event) => setTaskForm((prev) => ({ ...prev, description: event.target.value }))}
                  className={`${inputClass} h-28 leading-6`}
                />
              </label>

              <label className="block">
                <span className="text-sm font-semibold">截止时间</span>
                <input
                  type="datetime-local"
                  value={taskForm.deadlineAt}
                  disabled={savingTask}
                  onChange={(event) => setTaskForm((prev) => ({ ...prev, deadlineAt: event.target.value }))}
                  className={inputClass}
                />
              </label>

              <div>
                <span className="text-sm font-semibold">发布范围</span>
                <div className="mt-2 grid grid-cols-2 gap-2">
                  {(['all', 'groups'] as const).map((targetType) => (
                    <label
                      key={targetType}
                      className="flex items-center gap-2 rounded-lg border border-line px-3 py-2 text-sm font-semibold dark:border-slate-800"
                    >
                      <input
                        type="radio"
                        name="targetType"
                        value={targetType}
                        checked={taskForm.targetType === targetType}
                        disabled={savingTask}
                        onChange={() =>
                          setTaskForm((prev) => ({
                            ...prev,
                            targetType,
                            targetGroupIds: targetType === 'all' ? [] : prev.targetGroupIds,
                          }))
                        }
                      />
                      {targetType === 'all' ? '全部学生' : '指定小组'}
                    </label>
                  ))}
                </div>
              </div>

              {taskForm.targetType === 'groups' && (
                <div>
                  <span className="text-sm font-semibold">目标小组</span>
                  <div className="mt-2 grid gap-2 sm:grid-cols-2">
                    {(overview?.groups ?? []).map((group) => (
                      <label
                        key={group.groupId}
                        className="flex items-center justify-between rounded-lg border border-line px-3 py-2 text-sm dark:border-slate-800"
                      >
                        <span>
                          <span className="block font-semibold">{group.groupName}</span>
                          <span className="text-xs text-muted dark:text-slate-400">
                            {group.currentCount} / {group.capacity} 人
                          </span>
                        </span>
                        <input
                          type="checkbox"
                          checked={taskForm.targetGroupIds.includes(group.groupId)}
                          disabled={savingTask}
                          onChange={() => setTaskForm((prev) => toggleTargetGroup(prev, group.groupId))}
                        />
                      </label>
                    ))}
                  </div>
                  {(overview?.groups.length ?? 0) === 0 && (
                    <p className="mt-2 text-sm text-muted dark:text-slate-400">暂无可选小组。</p>
                  )}
                </div>
              )}

              <div className="rounded-lg border border-dashed border-line bg-slate-50 px-3 py-4 text-sm text-muted dark:border-slate-800 dark:bg-slate-950 dark:text-slate-400">
                附件上传暂未实现，本次发布会以空附件地址提交。
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <button type="button" className={btnSecondary} disabled={savingTask} onClick={() => setCreateOpen(false)}>
                  取消
                </button>
                <button type="submit" className={btnPrimary} disabled={savingTask}>
                  <FilePlus2 className="h-4 w-4" />
                  {savingTask ? '正在发布...' : '发布'}
                </button>
              </div>
            </form>
          </Card>
        </div>
      )}

      <ToastView />
    </div>
  )
}

function getErrorMessage(error: unknown, fallback: string) {
  if (error instanceof Error && error.message.trim()) {
    return error.message
  }
  return fallback
}

function isAuthError(error: unknown) {
  return error instanceof ApiRequestError && (error.status === 401 || error.status === 403)
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

function taskStatusLabel(status: TaskStatus) {
  switch (status) {
    case 'published':
      return '进行中'
    case 'paused':
      return '已暂停'
    case 'closed':
      return '已关闭'
  }
}

function roomStatusLabel(status: string) {
  switch (status) {
    case 'active':
      return '进行中'
    case 'ended':
      return '已结束'
    case 'created':
      return '已创建'
    default:
      return status
  }
}

function buildReviewSummary(tasks: Task[], studentCount: number) {
  if (tasks.length === 0) {
    const safeTotal = Math.max(studentCount, 1)
    return [
      { label: '已提交', value: 0, total: safeTotal, tone: 'brand' as const },
      { label: '开放任务', value: 0, total: 1, tone: 'emerald' as const },
      { label: '已关闭任务', value: 0, total: 1, tone: 'sky' as const },
    ]
  }

  const targetTotal = tasks.reduce((sum, task) => sum + (task.targetStudentCount ?? 0), 0)
  const submittedTotal = tasks.reduce((sum, task) => sum + (task.submittedCount ?? 0), 0)
  const safeTotal = Math.max(targetTotal, 1)

  return [
    { label: '已提交', value: submittedTotal, total: safeTotal, tone: 'brand' as const },
    { label: '开放任务', value: tasks.filter((task) => task.status !== 'closed').length, total: tasks.length, tone: 'emerald' as const },
    { label: '已关闭任务', value: tasks.filter((task) => task.status === 'closed').length, total: tasks.length, tone: 'sky' as const },
  ]
}

function defaultDeadlineInputValue() {
  const date = new Date()
  date.setHours(date.getHours() + 1, 0, 0, 0)
  return toLocalDateTimeInputValue(date)
}

function toLocalDateTimeInputValue(date: Date) {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hour = String(date.getHours()).padStart(2, '0')
  const minute = String(date.getMinutes()).padStart(2, '0')
  return `${year}-${month}-${day}T${hour}:${minute}`
}

function toISOString(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return null
  }
  return date.toISOString()
}

function toggleTargetGroup(form: TaskFormState, groupId: number): TaskFormState {
  const selected = form.targetGroupIds.includes(groupId)
  return {
    ...form,
    targetGroupIds: selected
      ? form.targetGroupIds.filter((id) => id !== groupId)
      : [...form.targetGroupIds, groupId],
  }
}
