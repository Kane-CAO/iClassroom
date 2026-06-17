import { useCallback, useEffect, useState } from 'react'
import {
  clearTeacherRoomSession,
  getTeacherSession,
  getTeacherToken,
  setTeacherRoomSession,
  type TeacherSession,
} from '../utils/session'

export function useTeacherSession() {
  const [session, setSession] = useState<TeacherSession | null>(null)

  useEffect(() => {
    setSession(getTeacherSession())
  }, [])

  const save = useCallback((next: TeacherSession) => {
    setSession(setTeacherRoomSession(next))
  }, [])

  const clear = useCallback(() => {
    clearTeacherRoomSession()
    setSession(null)
  }, [])

  return {
    session,
    roomCode: session?.roomCode ?? '',
    teacherToken: session?.teacherToken ?? getTeacherToken(),
    save,
    clear,
  }
}
