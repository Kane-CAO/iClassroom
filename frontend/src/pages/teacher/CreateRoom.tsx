import { useState } from 'react'
import { Send } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import TeacherHeader from '../../components/layout/TeacherHeader'
import Card from '../../components/ui/Card'
import { btnPrimary } from '../../components/ui/buttons'
import { useToast } from '../../components/ui/useToast'
import { createRoom } from '../../api/rooms'
import { setTeacherRoomSession } from '../../utils/session'
import { useTeacherSession } from '../../hooks/useTeacherSession'

const inputClass =
  'mt-2 w-full rounded-lg border border-line bg-white px-3 py-2.5 text-sm outline-none focus:border-brand-500 dark:border-slate-800 dark:bg-slate-950'

const DEFAULT_GROUP_COUNT = 3
const DEFAULT_GROUP_CAPACITY = 5

// /teacher/create-room
// 迁移自 docs/prototypes/iClassroom.html 的 #page-create（Create Assignment 表单）。
// 注意：按需求，本阶段不实现附件上传，文件输入仅作视觉占位（disabled）。
export default function CreateRoom() {
  const navigate = useNavigate()
  const { showToast, ToastView } = useToast()
  const { token, hasTeacherAccess } = useTeacherSession()

  const [title, setTitle] = useState('用户画像研究包')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [submissionTypes, setSubmissionTypes] = useState({
    written: true,
    image: true,
    file: true,
  })

  const onPublish = async () => {
    const roomTitle = title.trim()
    if (!roomTitle) {
      const message = '请输入课堂标题'
      setError(message)
      showToast(message)
      return
    }
    if (!hasTeacherAccess) {
      const message = '请先登录讲师账号'
      setError(message)
      showToast(message)
      navigate('/teacher/login')
      return
    }

    setLoading(true)
    setError(null)
    showToast('正在创建课堂...')

    try {
      const room = await createRoom(
        {
          title: roomTitle,
          groupCount: DEFAULT_GROUP_COUNT,
          groupCapacity: DEFAULT_GROUP_CAPACITY,
          allowChooseGroup: true,
        },
        { token },
      )

      setTeacherRoomSession({
        roomCode: room.roomCode,
        teacherToken: room.teacherToken,
      })

      showToast('课堂已创建')
      setTimeout(() => navigate(`/teacher/rooms/${room.roomCode}/dashboard`), 500)
    } catch (err) {
      const message = getErrorMessage(err)
      setError(message)
      showToast(message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-canvas text-ink dark:bg-slate-950 dark:text-slate-100">
      <TeacherHeader roomCode="新建" />

      <main className="px-8 py-7">
        <div className="grid grid-cols-1 gap-6 lg:grid-cols-[1fr_340px]">
          <Card padded className="!p-6">
            <form
              onSubmit={(e) => {
                e.preventDefault()
                onPublish()
              }}
            >
              <div className="mb-6 flex items-start justify-between">
                <div>
                  <p className="text-sm font-semibold text-brand-600 dark:text-brand-100">创建课堂</p>
                  <h1 className="mt-2 text-2xl font-semibold tracking-normal">新建课堂任务集</h1>
                </div>
                <button type="submit" disabled={loading} className={btnPrimary.replace('shadow-soft', '')}>
                  <Send className="h-4 w-4" />
                  {loading ? '正在创建...' : '创建并发布'}
                </button>
              </div>

              {error && (
                <div className="mb-5 rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm font-semibold text-rose-700 dark:border-rose-500/20 dark:bg-rose-500/10 dark:text-rose-300">
                  {error}
                </div>
              )}

              <div className="grid grid-cols-1 gap-5 sm:grid-cols-2">
                <label className="block sm:col-span-2">
                  <span className="text-sm font-semibold">课堂标题</span>
                  <input
                    className={inputClass}
                    value={title}
                    disabled={loading}
                    onChange={(e) => setTitle(e.target.value)}
                  />
                </label>
                <label className="block sm:col-span-2">
                  <span className="text-sm font-semibold">说明</span>
                  <input
                    className={inputClass}
                    defaultValue="提交你的访谈综合分析、用户画像板和最终研究备忘录。"
                  />
                </label>
                <label className="block sm:col-span-2">
                  <span className="text-sm font-semibold">提交要求</span>
                  <textarea
                    className={`${inputClass} h-32 leading-6`}
                    defaultValue="提交一份文字回答、至少两张图片材料和相关支撑文件。原始访谈笔记可作为 PDF 或 DOCX 附上。"
                  />
                </label>
                <label className="block">
                  <span className="text-sm font-semibold">截止日期</span>
                  <input type="date" className={inputClass} defaultValue="2026-06-18" />
                </label>
                <label className="block">
                  <span className="text-sm font-semibold">课堂时长</span>
                  <select className={inputClass} defaultValue="45 分钟">
                    <option>45 分钟</option>
                    <option>60 分钟</option>
                    <option>90 分钟</option>
                    <option>不限时</option>
                  </select>
                </label>
                <label className="block sm:col-span-2">
                  <span className="text-sm font-semibold">附件上传</span>
                  <div className="mt-2 flex items-center justify-center rounded-lg border border-dashed border-line bg-slate-50 px-3 py-6 text-sm text-muted dark:border-slate-800 dark:bg-slate-950">
                    附件上传暂未实现（本阶段仅做视觉参考）
                  </div>
                </label>
                <label className="block sm:col-span-2">
                  <span className="text-sm font-semibold">可选 PDF 附件</span>
                  <div className="mt-2 flex items-center justify-center rounded-lg border border-dashed border-line bg-slate-50 px-3 py-6 text-sm text-muted dark:border-slate-800 dark:bg-slate-950">
                    PDF 上传暂未实现
                  </div>
                </label>
              </div>
            </form>
          </Card>

          <aside className="space-y-5">
            <Card padded>
              <h2 className="text-sm font-semibold">学生提交设置</h2>
              <div className="mt-4 space-y-3">
                {(
                  [
                    ['written', '文字回答'],
                    ['image', '图片上传'],
                    ['file', '文件上传'],
                  ] as const
                ).map(([key, label]) => (
                  <label
                    key={key}
                    className="flex items-center justify-between rounded-lg border border-line px-3 py-2 dark:border-slate-800"
                  >
                    <span className="text-sm font-semibold">{label}</span>
                    <input
                      type="checkbox"
                      checked={submissionTypes[key]}
                      onChange={(e) => setSubmissionTypes((s) => ({ ...s, [key]: e.target.checked }))}
                    />
                  </label>
                ))}
              </div>
            </Card>

            <Card padded>
              <h2 className="text-sm font-semibold">学生上传预览</h2>
              <div className="mt-4 rounded-lg border border-line bg-slate-50 p-4 dark:border-slate-800 dark:bg-slate-950">
                <div className="mb-3 h-20 rounded-lg border border-dashed border-slate-300 bg-white dark:border-slate-700 dark:bg-slate-900" />
                <p className="text-xs leading-5 text-muted dark:text-slate-400">
                  学生提交前可附加图片、PDF、DOCX 和项目文件。本阶段部分上传能力仍是视觉占位。
                </p>
              </div>
            </Card>
          </aside>
        </div>
      </main>

      <ToastView />
    </div>
  )
}

function getErrorMessage(error: unknown) {
  if (error instanceof Error && error.message.trim()) {
    return error.message
  }
  return '创建课堂失败'
}
