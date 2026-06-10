import { useCallback, useEffect, useState } from 'react'
import { getItem, setItem } from '../utils/storage'
import { questions } from '../mocks/student'

// 学生会话状态（身份 + 草稿 + 提交/批改状态），持久化到 localStorage。
// 迁移自 student.html / studentphone.html 的 loadDraft / saveCurrentAnswer / submit 逻辑。
// 注意：按需求不实现图片上传，这里只保存文字草稿。
export interface StudentIdentity {
  name: string
  team: string
}

const KEY = {
  identity: 'student:identity',
  drafts: 'student:drafts',
  submitted: 'student:submitted',
  reviewed: 'student:reviewed',
}

export function questionDone(text: string) {
  return text.trim().length > 20
}

export function useStudentSession() {
  const [identity, setIdentityState] = useState<StudentIdentity | null>(null)
  const [answers, setAnswers] = useState<string[]>(() => questions.map(() => ''))
  const [submitted, setSubmittedState] = useState(false)
  const [reviewed, setReviewedState] = useState(false)

  // 初始化：从 localStorage 还原。
  useEffect(() => {
    setIdentityState(getItem<StudentIdentity>(KEY.identity))
    const drafts = getItem<string[]>(KEY.drafts)
    if (drafts) setAnswers(questions.map((_, i) => drafts[i] ?? ''))
    setSubmittedState(getItem<boolean>(KEY.submitted) ?? false)
    setReviewedState(getItem<boolean>(KEY.reviewed) ?? false)
  }, [])

  const join = useCallback((next: StudentIdentity) => {
    setIdentityState(next)
    setItem(KEY.identity, next)
  }, [])

  const setAnswer = useCallback((index: number, text: string) => {
    setAnswers((prev) => {
      const next = [...prev]
      next[index] = text
      setItem(KEY.drafts, next)
      return next
    })
  }, [])

  const submit = useCallback(() => {
    setSubmittedState(true)
    setItem(KEY.submitted, true)
  }, [])

  const review = useCallback(() => {
    setReviewedState(true)
    setItem(KEY.reviewed, true)
  }, [])

  const doneCount = answers.filter(questionDone).length

  return {
    identity,
    answers,
    submitted,
    reviewed,
    doneCount,
    join,
    setAnswer,
    submit,
    review,
  }
}
