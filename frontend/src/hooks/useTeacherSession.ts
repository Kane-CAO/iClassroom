import { useCallback, useEffect, useState } from 'react'
import {
  clearTeacherRoomSession,
  getTeacherAuthToken,
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
    token: session?.token ?? getTeacherAuthToken(),
    teacherToken: session?.teacherToken ?? getTeacherToken(),
    hasTeacherAccess: Boolean(session?.token || getTeacherAuthToken() || session?.teacherToken || getTeacherToken()),
    save,
    clear,
  }
}
