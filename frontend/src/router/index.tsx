import { createBrowserRouter } from 'react-router-dom'
import Home from '../pages/Home'
import AdminLogin from '../pages/admin/AdminLogin'
import TeacherAccounts from '../pages/admin/TeacherAccounts'
import TeacherLogin from '../pages/teacher/TeacherLogin'
import CreateRoom from '../pages/teacher/CreateRoom'
import Dashboard from '../pages/teacher/Dashboard'
import CourseDetail from '../pages/teacher/CourseDetail'
import Review from '../pages/teacher/Review'
import Analytics from '../pages/teacher/Analytics'
import Display from '../pages/display/Display'
import StudentEntry from '../pages/student/StudentEntry'
import Classroom from '../pages/student/Classroom'
import TaskDetail from '../pages/student/TaskDetail'

// 路由表。页面均使用 src/mocks 的 mock 数据，暂不接入真实后端 / WebSocket。
export const router = createBrowserRouter([
  { path: '/', element: <Home /> },

  // 管理者端
  { path: '/admin/login', element: <AdminLogin /> },
  { path: '/admin/teachers', element: <TeacherAccounts /> },

  // 讲师端（电脑优先）
  { path: '/teacher/login', element: <TeacherLogin /> },
  { path: '/teacher/create-room', element: <CreateRoom /> },
  { path: '/teacher/rooms/:roomCode/dashboard', element: <Dashboard /> },
  { path: '/teacher/rooms/:roomCode/course', element: <CourseDetail /> },
  { path: '/teacher/rooms/:roomCode/review', element: <Review /> },
  { path: '/teacher/rooms/:roomCode/display', element: <Display /> },
  { path: '/teacher/rooms/:roomCode/analytics', element: <Analytics /> },

  // 学生端（手机 / 平板优先）
  { path: '/student', element: <StudentEntry /> },
  { path: '/student/classroom', element: <Classroom /> },
  { path: '/student/tasks/:taskId', element: <TaskDetail /> },
])
