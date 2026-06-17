import { useCallback, useState } from 'react'
import {
  clearStudentSession,
  getStudentSession,
  setStudentSession,
  type StudentIdentity,
  type StudentSession,
} from '../utils/session'

export function questionDone(text: string) {
  return text.trim().length > 20
}

export function useStudentSession() {
  const [identity, setIdentityState] = useState<StudentIdentity | null>(() => getStudentSession())

  const join = useCallback((next: StudentSession) => {
    setIdentityState(setStudentSession(next))
  }, [])

  const clear = useCallback(() => {
    clearStudentSession()
    setIdentityState(null)
  }, [])

  return {
    identity,
    answers: [],
    submitted: false,
    reviewed: false,
    doneCount: 0,
    join,
    clear,
    setAnswer: () => undefined,
    submit: () => undefined,
    review: () => undefined,
  }
}
