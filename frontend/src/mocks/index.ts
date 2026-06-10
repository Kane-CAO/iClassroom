// 本地 mock 数据汇总入口。当前阶段不接真实后端，页面统一从此处读取演示数据。
// 原型迁移后的视图 mock 数据按端拆分到独立文件，从此处统一再导出。
import type { Group, Room, Task } from '../types'

export * as teacherMocks from './teacher'
export * as studentMocks from './student'

export const mockRoom: Room = {
  id: 1,
  roomCode: 'ABC123',
  title: 'Demo Class',
  groupCount: 3,
  groupCapacity: 2,
  allowChooseGroup: true,
  status: 'active',
}

export const mockGroups: Group[] = [
  { id: 1, roomId: 1, name: '第1组', capacity: 2, scoreTotal: 0 },
  { id: 2, roomId: 1, name: '第2组', capacity: 2, scoreTotal: 0 },
  { id: 3, roomId: 1, name: '第3组', capacity: 2, scoreTotal: 0 },
]

export const mockTasks: Task[] = [
  {
    id: 1,
    roomId: 1,
    title: '课堂测试一',
    description: '请提交你对 AI Classroom 的理解',
    deadlineAt: '2026-06-08T18:00:00Z',
    status: 'published',
    targetType: 'all',
  },
]
