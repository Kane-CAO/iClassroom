import { getItem, removeItem, setItem } from './storage'

const KEY = {
  studentSession: 'student:identity',
  teacherSession: 'teacher:session',
}

export interface StudentSession {
  studentId: number
  roomCode: string
  nickname: string
  groupId: number
  groupName: string
  clientToken: string
}

export interface StudentIdentity extends StudentSession {
  name: string
  team: string
}

export interface TeacherSession {
  roomCode: string
  teacherToken: string
}

interface LegacyStudentIdentity {
  name: string
  team: string
  roomCode?: string
  clientToken?: string
}

export function getStudentSession(): StudentIdentity | null {
  const value = getItem<unknown>(KEY.studentSession)
  return normalizeStudentSession(value)
}

export function setStudentSession(session: StudentSession): StudentIdentity {
  const next = toStudentIdentity(session)
  setItem(KEY.studentSession, toStoredStudentSession(next))
  return next
}

export function clearStudentSession() {
  removeItem(KEY.studentSession)
}

export function getStudentToken() {
  return getStudentSession()?.clientToken ?? ''
}

export function getTeacherSession(): TeacherSession | null {
  const value = getItem<unknown>(KEY.teacherSession)
  return normalizeTeacherSession(value)
}

export function getTeacherToken() {
  return getTeacherSession()?.teacherToken ?? ''
}

export function setTeacherRoomSession(session: TeacherSession): TeacherSession {
  setItem(KEY.teacherSession, session)
  return session
}

export function clearTeacherRoomSession() {
  removeItem(KEY.teacherSession)
}

function normalizeStudentSession(value: unknown): StudentIdentity | null {
  if (isStoredStudentSession(value)) {
    return toStudentIdentity(value)
  }

  if (isLegacyStudentIdentity(value)) {
    return {
      studentId: 0,
      roomCode: value.roomCode ?? '',
      nickname: value.name,
      groupId: 0,
      groupName: value.team,
      clientToken: value.clientToken ?? '',
      name: value.name,
      team: value.team,
    }
  }

  return null
}

function normalizeTeacherSession(value: unknown): TeacherSession | null {
  if (!isRecord(value)) {
    return null
  }

  if (isNonEmptyString(value.roomCode) && isNonEmptyString(value.teacherToken)) {
    return {
      roomCode: value.roomCode,
      teacherToken: value.teacherToken,
    }
  }

  return null
}

function toStudentIdentity(session: StudentSession): StudentIdentity {
  return {
    ...session,
    name: session.nickname,
    team: session.groupName,
  }
}

function toStoredStudentSession(identity: StudentIdentity): StudentSession {
  return {
    studentId: identity.studentId,
    roomCode: identity.roomCode,
    nickname: identity.nickname,
    groupId: identity.groupId,
    groupName: identity.groupName,
    clientToken: identity.clientToken,
  }
}

function isStoredStudentSession(value: unknown): value is StudentSession {
  if (!isRecord(value)) {
    return false
  }

  return (
    typeof value.studentId === 'number' &&
    isNonEmptyString(value.roomCode) &&
    isNonEmptyString(value.nickname) &&
    typeof value.groupId === 'number' &&
    isNonEmptyString(value.groupName) &&
    isNonEmptyString(value.clientToken)
  )
}

function isLegacyStudentIdentity(value: unknown): value is LegacyStudentIdentity {
  if (!isRecord(value)) {
    return false
  }

  return (
    isNonEmptyString(value.name) &&
    isNonEmptyString(value.team) &&
    (value.roomCode === undefined || typeof value.roomCode === 'string') &&
    (value.clientToken === undefined || typeof value.clientToken === 'string')
  )
}

function isNonEmptyString(value: unknown): value is string {
  return typeof value === 'string' && value.trim().length > 0
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}
