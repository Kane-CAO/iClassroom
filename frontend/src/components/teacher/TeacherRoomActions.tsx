import { useState } from 'react'
import { Download, Lock } from 'lucide-react'
import { exportRoom } from '../../api/export'
import { ApiRequestError } from '../../api/client'
import { endRoom } from '../../api/rooms'
import { btnSecondary } from '../ui/buttons'

interface TeacherRoomActionsProps {
  roomCode: string
  teacherToken: string
  roomEnded: boolean
  className?: string
  buttonClassName?: string
  showEnd?: boolean
  showExport?: boolean
  onEnded: () => void
  onError: (message: string) => void
  onSuccess: (message: string) => void
}

export default function TeacherRoomActions({
  roomCode,
  teacherToken,
  roomEnded,
  className = '',
  buttonClassName = btnSecondary,
  showEnd = true,
  showExport = true,
  onEnded,
  onError,
  onSuccess,
}: TeacherRoomActionsProps) {
  const [ending, setEnding] = useState(false)
  const [exporting, setExporting] = useState(false)

  const handleEndRoom = async () => {
    if (roomEnded) {
      onSuccess('课堂已经结束')
      return
    }
    if (!teacherToken) {
      onError('老师会话缺失，请重新创建课堂。')
      return
    }
    const confirmed = window.confirm(`确定结束课堂 ${roomCode} 吗？结束后学生将不能继续提交答案。`)
    if (!confirmed) {
      return
    }

    setEnding(true)
    try {
      await endRoom(roomCode, { teacherToken })
      onSuccess('课堂已结束')
      onEnded()
    } catch (err: unknown) {
      if (isAlreadyEndedError(err)) {
        onSuccess('课堂已经结束')
        onEnded()
        return
      }
      onError(getErrorMessage(err, '结束课堂失败。'))
    } finally {
      setEnding(false)
    }
  }

  const handleExportRoom = async () => {
    if (!teacherToken) {
      onError('老师会话缺失，请重新创建课堂。')
      return
    }

    setExporting(true)
    try {
      const result = await exportRoom(roomCode, { teacherToken })
      downloadBlob(result.blob, result.fileName || `${roomCode}-export.zip`)
      onSuccess('导出文件已下载')
    } catch (err: unknown) {
      onError(getErrorMessage(err, '导出课堂数据失败。'))
    } finally {
      setExporting(false)
    }
  }

  const containerClassName = className || 'flex flex-wrap gap-2'

  return (
    <div className={containerClassName}>
      {showEnd && (
        <button className={buttonClassName} disabled={ending || roomEnded} onClick={handleEndRoom}>
          <span>{ending ? '正在结束...' : roomEnded ? '课堂已结束' : '结束课堂'}</span>
          <Lock className="h-4 w-4" />
        </button>
      )}
      {showExport && (
        <button className={buttonClassName} disabled={exporting} onClick={handleExportRoom}>
          <span>{exporting ? '正在导出...' : '导出数据'}</span>
          <Download className="h-4 w-4" />
        </button>
      )}
    </div>
  )
}

function downloadBlob(blob: Blob, fileName: string) {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = fileName
  document.body.appendChild(link)
  link.click()
  link.remove()
  URL.revokeObjectURL(url)
}

function isAlreadyEndedError(err: unknown) {
  return err instanceof ApiRequestError && err.errorCode === 'ROOM_ALREADY_ENDED'
}

function getErrorMessage(err: unknown, fallback: string) {
  if (err instanceof Error && err.message.trim()) {
    return err.message
  }
  return fallback
}
