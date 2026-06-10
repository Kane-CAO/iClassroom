import { useState } from 'react'
import { Send } from 'lucide-react'
import { useNavigate, useParams } from 'react-router-dom'
import TeacherHeader from '../../components/layout/TeacherHeader'
import Card from '../../components/ui/Card'
import { btnPrimary } from '../../components/ui/buttons'
import { useToast } from '../../components/ui/useToast'

const inputClass =
  'mt-2 w-full rounded-lg border border-line bg-white px-3 py-2.5 text-sm outline-none focus:border-brand-500 dark:border-slate-800 dark:bg-slate-950'

// /teacher/create-room
// 迁移自 docs/prototypes/iClassroom.html 的 #page-create（Create Assignment 表单）。
// 注意：按需求，本阶段不实现附件上传，文件输入仅作视觉占位（disabled）。
export default function CreateRoom() {
  const { roomCode = 'ABC123' } = useParams()
  const navigate = useNavigate()
  const { showToast, ToastView } = useToast()

  const [submissionTypes, setSubmissionTypes] = useState({
    written: true,
    image: true,
    file: true,
  })

  const onPublish = () => {
    showToast('Assignment published')
    setTimeout(() => navigate(`/teacher/rooms/${roomCode}/course`), 600)
  }

  return (
    <div className="min-h-screen bg-canvas text-ink dark:bg-slate-950 dark:text-slate-100">
      <TeacherHeader roomCode={roomCode} active="assignments" />

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
                  <p className="text-sm font-semibold text-brand-600 dark:text-brand-100">Create Assignment</p>
                  <h1 className="mt-2 text-2xl font-semibold tracking-normal">New classroom collection</h1>
                </div>
                <button type="submit" className={btnPrimary.replace('shadow-soft', '')}>
                  <Send className="h-4 w-4" />
                  Publish
                </button>
              </div>

              <div className="grid grid-cols-1 gap-5 sm:grid-cols-2">
                <label className="block sm:col-span-2">
                  <span className="text-sm font-semibold">Assignment Title</span>
                  <input className={inputClass} defaultValue="Persona Research Pack" />
                </label>
                <label className="block sm:col-span-2">
                  <span className="text-sm font-semibold">Description</span>
                  <input
                    className={inputClass}
                    defaultValue="Submit your interview synthesis, persona board, and final research memo."
                  />
                </label>
                <label className="block sm:col-span-2">
                  <span className="text-sm font-semibold">Instructions</span>
                  <textarea
                    className={`${inputClass} h-32 leading-6`}
                    defaultValue="Upload a written response, at least two image artifacts, and supporting files. Include the source interview notes as a PDF or DOCX."
                  />
                </label>
                <label className="block">
                  <span className="text-sm font-semibold">Due Date</span>
                  <input type="date" className={inputClass} defaultValue="2026-06-18" />
                </label>
                <label className="block">
                  <span className="text-sm font-semibold">Time Limit</span>
                  <select className={inputClass} defaultValue="45 minutes">
                    <option>45 minutes</option>
                    <option>60 minutes</option>
                    <option>90 minutes</option>
                    <option>No timed session</option>
                  </select>
                </label>
                <label className="block sm:col-span-2">
                  <span className="text-sm font-semibold">Attachment Upload</span>
                  <div className="mt-2 flex items-center justify-center rounded-lg border border-dashed border-line bg-slate-50 px-3 py-6 text-sm text-muted dark:border-slate-800 dark:bg-slate-950">
                    附件上传暂未实现（本阶段仅做视觉参考）
                  </div>
                </label>
                <label className="block sm:col-span-2">
                  <span className="text-sm font-semibold">Optional PDF Attachment</span>
                  <div className="mt-2 flex items-center justify-center rounded-lg border border-dashed border-line bg-slate-50 px-3 py-6 text-sm text-muted dark:border-slate-800 dark:bg-slate-950">
                    PDF 上传暂未实现
                  </div>
                </label>
              </div>
            </form>
          </Card>

          <aside className="space-y-5">
            <Card padded>
              <h2 className="text-sm font-semibold">Student Submission Settings</h2>
              <div className="mt-4 space-y-3">
                {(
                  [
                    ['written', 'Written response'],
                    ['image', 'Image upload'],
                    ['file', 'File upload'],
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
              <h2 className="text-sm font-semibold">Student Upload Preview</h2>
              <div className="mt-4 rounded-lg border border-line bg-slate-50 p-4 dark:border-slate-800 dark:bg-slate-950">
                <div className="mb-3 h-20 rounded-lg border border-dashed border-slate-300 bg-white dark:border-slate-700 dark:bg-slate-900" />
                <p className="text-xs leading-5 text-muted dark:text-slate-400">
                  Students can attach images, PDFs, DOCX, and project files before submitting.
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
