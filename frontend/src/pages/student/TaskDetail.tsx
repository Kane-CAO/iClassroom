import { useCallback, useEffect, useState } from 'react'
import { ArrowLeft, FileUp, ImagePlus, Send, X } from 'lucide-react'
import { useNavigate, useParams } from 'react-router-dom'
import StudentHeader from '../../components/layout/StudentHeader'
import Card from '../../components/ui/Card'
import Badge from '../../components/ui/Badge'
import { btnSecondary, btnGradient } from '../../components/ui/buttons'
import { useToast } from '../../components/ui/useToast'
import { useRoomWebSocket } from '../../hooks/useRoomWebSocket'
import { useStudentSession } from '../../hooks/useStudentSession'
import { ApiRequestError } from '../../api/client'
import { getStudentTask } from '../../api/student'
import { submitTaskForm } from '../../api/tasks'
import type { RoomWebSocketEventType } from '../../api/websocket'
import type { BadgeTone } from '../../types'
import type { Task } from '../../types/api'

const TASK_DETAIL_WS_EVENTS: readonly RoomWebSocketEventType[] = [
  'task_paused',
  'task_closed',
  'room_ended',
]

export default function TaskDetail() {
  const { taskId } = useParams()
  const navigate = useNavigate()
  const { showToast, ToastView } = useToast()
  const { identity, clear } = useStudentSession()
  const roomCode = identity?.roomCode ?? ''
  const studentToken = identity?.clientToken ?? ''
  const parsedTaskId = Number(taskId)

  const [task, setTask] = useState<Task | null>(null)
  const [answer, setAnswer] = useState('')
  const [tab, setTab] = useState<'answer' | 'attachments'>('answer')
  const [images, setImages] = useState<File[]>([])
  const [files, setFiles] = useState<File[]>([])
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [roomEnded, setRoomEnded] = useState(false)
  const [refreshVersion, setRefreshVersion] = useState(0)
  const draftKey = roomCode && parsedTaskId ? `student:draft:${roomCode}:${parsedTaskId}` : ''
  const submitted = task?.mySubmissionStatus === 'submitted' || task?.mySubmissionStatus === 'graded'

  const refreshTaskDetailData = useCallback(() => {
    setRefreshVersion((version) => version + 1)
  }, [])

  useEffect(() => {
    let cancelled = false

    async function loadTask() {
      setLoading(true)
      setError(null)

      if (!roomCode || !studentToken) {
        setTask(null)
        setError('学生会话缺失，请重新加入课堂。')
        setLoading(false)
        return
      }

      if (!Number.isInteger(parsedTaskId) || parsedTaskId <= 0) {
        setTask(null)
        setError('未找到任务。')
        setLoading(false)
        return
      }

      try {
        const data = await getStudentTask(parsedTaskId, { studentToken })
        if (cancelled) {
          return
        }
        setTask(data)
      } catch (err) {
        if (cancelled) {
          return
        }

        if (isAuthError(err)) {
          clear()
          setError('你的会话已过期，请重新加入课堂。')
        } else if (isRoomEndedError(err)) {
          setRoomEnded(true)
          setError('课堂已结束。')
        } else {
          setError(getErrorMessage(err, '加载任务失败。'))
        }
        setTask(null)
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    loadTask()
    return () => {
      cancelled = true
    }
  }, [clear, parsedTaskId, refreshVersion, roomCode, studentToken])

  useEffect(() => {
    if (!draftKey) {
      return
    }
    const saved = window.localStorage.getItem(draftKey)
    if (saved && !answer) {
      setAnswer(saved)
    }
  }, [answer, draftKey])

  useEffect(() => {
    if (!draftKey || submitted) {
      return
    }
    window.localStorage.setItem(draftKey, answer)
  }, [answer, draftKey, submitted])

  const ws = useRoomWebSocket({
    roomCode,
    role: 'student',
    clientToken: studentToken,
    onEvent: (event) => {
      if (TASK_DETAIL_WS_EVENTS.includes(event.type)) {
        if (event.type === 'room_ended') {
          setRoomEnded(true)
          showToast('课堂已结束')
        }
        refreshTaskDetailData()
      }
    },
    onReconnect: refreshTaskDetailData,
  })

  const blockedReason = task ? getSubmitBlockedReason(task, roomEnded) : '当前任务不可用。'
  const hasPayload = answer.trim().length > 0 || images.length > 0 || files.length > 0
  const canSubmit = Boolean(task && !blockedReason && hasPayload && !submitting)
  const currentStatus = task ? taskStatus(task, roomEnded) : { label: '加载中', tone: 'slate' as const }

  const onSubmit = async () => {
    if (!task || blockedReason) {
      if (blockedReason) {
        showToast(blockedReason)
      }
      return
    }

    if (!hasPayload) {
      setError('提交前请填写答案或选择附件。')
      showToast('请填写答案或选择附件')
      return
    }

    setSubmitting(true)
    setError(null)

    try {
      const body = new FormData()
      body.set('contentText', answer.trim())
      images.forEach((file) => body.append('images[]', file))
      files.forEach((file) => body.append('files[]', file))
      await submitTaskForm(task.taskId, body, { studentToken })
      if (draftKey) {
        window.localStorage.removeItem(draftKey)
      }
      showToast('任务已提交')
      refreshTaskDetailData()
      setTimeout(() => navigate('/student/classroom'), 500)
    } catch (err) {
      const message = getErrorMessage(err, '提交任务失败。')
      setError(message)
      showToast(message)
      refreshTaskDetailData()
    } finally {
      setSubmitting(false)
    }
  }

  const addImages = (fileList: FileList | null) => {
    if (!fileList) {
      return
    }
    const next = [...images, ...Array.from(fileList)].slice(0, 3)
    const tooLarge = next.some((file) => file.size > 5 * 1024 * 1024)
    if (tooLarge) {
      setError('图片单张不能超过 5MB。')
      showToast('图片单张不能超过 5MB')
      return
    }
    setImages(next)
  }

  const addFiles = (fileList: FileList | null) => {
    if (!fileList) {
      return
    }
    const next = [...files, ...Array.from(fileList)].slice(0, 5)
    const hasVideo = next.some((file) => file.type.startsWith('video/'))
    if (hasVideo) {
      setError('不支持上传视频文件。')
      showToast('不支持上传视频文件')
      return
    }
    const tooLarge = next.some((file) => file.size > 20 * 1024 * 1024)
    if (tooLarge) {
      setError('文件单个不能超过 20MB。')
      showToast('文件单个不能超过 20MB')
      return
    }
    setFiles(next)
  }

  const removeImage = (index: number) => {
    setImages((current) => current.filter((_, i) => i !== index))
  }

  const removeFile = (index: number) => {
    setFiles((current) => current.filter((_, i) => i !== index))
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
        {error && (
          <Card className="mb-6 border-rose-200 bg-rose-50 p-4 text-rose-700 dark:border-rose-500/20 dark:bg-rose-500/10 dark:text-rose-300">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <p className="text-sm font-semibold">{error}</p>
              {isSessionError(error) && (
                <button
                  className={btnSecondary}
                  onClick={() => navigate(roomCode ? `/student?room=${roomCode}` : '/student')}
                >
                  重新加入
                </button>
              )}
            </div>
          </Card>
        )}

        <Card>
          <div className="border-b border-line p-5 dark:border-slate-800">
            <div className="flex items-center justify-between gap-4">
              <div>
                <h2 className="text-lg font-semibold tracking-normal">任务详情</h2>
                <p className="mt-1 text-sm text-muted dark:text-slate-400">
                  {loading ? '正在加载任务...' : `房间 ${roomCode || '未知'}`}
                </p>
              </div>
              <Badge tone={currentStatus.tone}>{currentStatus.label}</Badge>
            </div>
          </div>

          <div className="p-5">
            {loading && <p className="text-sm text-muted dark:text-slate-400">正在加载任务详情...</p>}

            {!loading && task && (
              <>
                <div className="mb-5 rounded-lg bg-slate-50 p-4 dark:bg-slate-950">
                  <div className="mb-2 flex items-center justify-between gap-3">
                    <span className="text-xs font-bold uppercase text-brand-700 dark:text-brand-100">
                      截止时间 {formatDateTime(task.deadlineAt)}
                    </span>
                    <Badge tone={currentStatus.tone}>{currentStatus.label}</Badge>
                  </div>
                  <h3 className="text-xl font-semibold tracking-normal">{task.title}</h3>
                  <p className="mt-2 text-sm leading-6 text-muted dark:text-slate-400">{task.description}</p>
                  {blockedReason && (
                    <p className="mt-3 text-sm font-semibold text-amber-700 dark:text-amber-300">{blockedReason}</p>
                  )}
                </div>

                <div className="flex rounded-lg border border-line bg-slate-50 p-1 dark:border-slate-800 dark:bg-slate-950">
                  <button className={tabClass(tab === 'answer')} onClick={() => setTab('answer')}>
                    文字提交
                  </button>
                  <button className={tabClass(tab === 'attachments')} onClick={() => setTab('attachments')}>
                    附件上传
                  </button>
                </div>

                {tab === 'answer' ? (
                  <div className="mt-5">
                    <label className="block">
                      <span className="mb-2 block text-sm font-semibold">文字回答</span>
                      <textarea
                        value={answer}
                        disabled={Boolean(blockedReason) || submitting}
                        onChange={(event) => setAnswer(event.target.value)}
                        placeholder={submitted ? '该任务已提交。' : '在这里填写你的答案...'}
                        className="h-[300px] w-full rounded-lg border border-line bg-white p-4 text-sm leading-7 outline-none transition focus:border-brand-500 disabled:opacity-70 dark:border-slate-800 dark:bg-slate-950"
                      />
                    </label>
                    {!submitted && (
                      <p className="mt-2 text-xs font-semibold text-muted dark:text-slate-400">
                        草稿会自动保存在本机，提交成功后清除。
                      </p>
                    )}
                  </div>
                ) : (
                  <div className="mt-5">
                    <div className="grid gap-4 md:grid-cols-2">
                      <label className="flex min-h-[180px] cursor-pointer flex-col items-center justify-center rounded-lg border-2 border-dashed border-line bg-slate-50 p-5 text-center dark:border-slate-800 dark:bg-slate-950">
                        <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-white text-brand-600 shadow-sm dark:bg-slate-900 dark:text-brand-100">
                          <ImagePlus className="h-6 w-6" />
                        </div>
                        <p className="mt-4 text-sm font-semibold">选择图片</p>
                        <p className="mt-1 text-xs text-muted dark:text-slate-400">最多 3 张，单张不超过 5MB。</p>
                        <input
                          className="hidden"
                          type="file"
                          accept="image/png,image/jpeg,image/webp"
                          multiple
                          disabled={Boolean(blockedReason) || submitting}
                          onChange={(event) => addImages(event.target.files)}
                        />
                      </label>
                      <label className="flex min-h-[180px] cursor-pointer flex-col items-center justify-center rounded-lg border-2 border-dashed border-line bg-slate-50 p-5 text-center dark:border-slate-800 dark:bg-slate-950">
                        <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-white text-brand-600 shadow-sm dark:bg-slate-900 dark:text-brand-100">
                          <FileUp className="h-6 w-6" />
                        </div>
                        <p className="mt-4 text-sm font-semibold">选择文件</p>
                        <p className="mt-1 text-xs text-muted dark:text-slate-400">PDF、Office、压缩包或文本文件。</p>
                        <input
                          className="hidden"
                          type="file"
                          multiple
                          disabled={Boolean(blockedReason) || submitting}
                          onChange={(event) => addFiles(event.target.files)}
                        />
                      </label>
                    </div>

                    {(images.length > 0 || files.length > 0) && (
                      <div className="mt-4 space-y-2">
                        {images.map((file, index) => (
                          <AttachmentRow key={`${file.name}-${index}`} file={file} onRemove={() => removeImage(index)} />
                        ))}
                        {files.map((file, index) => (
                          <AttachmentRow key={`${file.name}-${index}`} file={file} onRemove={() => removeFile(index)} />
                        ))}
                      </div>
                    )}
                  </div>
                )}
              </>
            )}

            <div className="mt-5 flex items-center justify-between gap-3 rounded-lg bg-slate-50 p-4 dark:bg-slate-950">
              <button className={btnSecondary} onClick={() => navigate('/student/classroom')}>
                <ArrowLeft className="h-4 w-4" />
                返回
              </button>
              <p className="hidden text-sm text-muted sm:block dark:text-slate-400">
                {blockedReason ?? '确认答案和附件无误后提交。'}
              </p>
              <button className={btnGradient} disabled={!canSubmit} onClick={onSubmit}>
                <Send className="h-4 w-4" />
                {submitting ? '提交中...' : '提交任务'}
              </button>
            </div>
          </div>
        </Card>
      </main>

      <ToastView />
    </div>
  )
}

function AttachmentRow({ file, onRemove }: { file: File; onRemove: () => void }) {
  return (
    <div className="flex items-center justify-between gap-3 rounded-lg border border-line bg-white px-3 py-2 text-sm dark:border-slate-800 dark:bg-slate-900">
      <span className="min-w-0 truncate font-semibold">{file.name}</span>
      <span className="shrink-0 text-xs text-muted dark:text-slate-400">{formatFileSize(file.size)}</span>
      <button
        type="button"
        className="shrink-0 rounded-md border border-line p-1 text-slate-500 hover:bg-slate-50 dark:border-slate-800 dark:hover:bg-slate-800"
        onClick={onRemove}
        aria-label="移除附件"
      >
        <X className="h-4 w-4" />
      </button>
    </div>
  )
}

function getSubmitBlockedReason(task: Task, roomEnded: boolean) {
  if (roomEnded) {
    return '课堂已结束。'
  }
  if (task.mySubmissionStatus === 'submitted' || task.mySubmissionStatus === 'graded') {
    return task.mySubmissionStatus === 'graded' ? '该任务已评分。' : '该任务已提交。'
  }
  if (task.status === 'paused') {
    return '任务已暂停。'
  }
  if (task.status === 'closed') {
    return '任务已关闭。'
  }
  if (!isBeforeDeadline(task.deadlineAt)) {
    return '任务已超过截止时间。'
  }
  return null
}

function taskStatus(task: Task, roomEnded: boolean): { label: string; tone: BadgeTone } {
  if (roomEnded) {
    return { label: '已结束', tone: 'slate' }
  }
  if (task.mySubmissionStatus === 'graded') {
    return { label: '已评分', tone: 'emerald' }
  }
  if (task.mySubmissionStatus === 'submitted') {
    return { label: '已提交', tone: 'brand' }
  }
  if (task.status === 'paused') {
    return { label: '已暂停', tone: 'amber' }
  }
  if (task.status === 'closed') {
    return { label: '已关闭', tone: 'slate' }
  }
  if (!isBeforeDeadline(task.deadlineAt)) {
    return { label: '已过期', tone: 'slate' }
  }
  return { label: '可提交', tone: 'amber' }
}

function isAuthError(error: unknown) {
  return error instanceof ApiRequestError && (error.status === 401 || error.status === 403)
}

function isRoomEndedError(error: unknown) {
  return error instanceof ApiRequestError && error.errorCode === 'ROOM_ENDED'
}

function getErrorMessage(error: unknown, fallback: string) {
  if (error instanceof Error && error.message.trim()) {
    return error.message
  }
  return fallback
}

function isSessionError(message: string) {
  const normalized = message.toLowerCase()
  return normalized.includes('session') || message.includes('会话')
}

function isBeforeDeadline(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return false
  }
  return Date.now() < date.getTime()
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

function formatFileSize(size: number) {
  if (!Number.isFinite(size) || size <= 0) {
    return '0 KB'
  }
  if (size < 1024 * 1024) {
    return `${Math.ceil(size / 1024)} KB`
  }
  return `${(size / 1024 / 1024).toFixed(1)} MB`
}
