# iClassroom Frontend Test Checklist

Use this checklist for manual verification after frontend or API-contract changes. Unless noted otherwise, run the backend on `http://localhost:8080` and the frontend on `http://localhost:5173`.

## Environment

- [ ] `frontend/.env` or shell env sets `VITE_API_BASE_URL` to the backend origin, for example `http://localhost:8080`.
- [ ] `VITE_API_BASE_URL` does not include `/api`; the frontend API client adds `/api`.
- [ ] `VITE_WS_BASE_URL` points to the backend WebSocket endpoint, for example `ws://localhost:8080/ws`; when omitted in dev, the app falls back to same-origin `/ws`.
- [ ] `VITE_STUDENT_BASE_URL` points to the student entry page, for example `http://localhost:5173/student`.
- [ ] No page code hardcodes `localhost` or `127.0.0.1`; environment and API client files are the only allowed local defaults.
- [ ] No frontend page/component/API file uses `any`.
- [ ] No frontend code logs `teacherToken`, `clientToken`, or `studentToken` to the console.
- [ ] `npm run build` completes successfully.

## Teacher Flow

- [ ] Open `/teacher/create-room`.
- [ ] Create a room with a title, group count, group capacity, and group selection setting.
- [ ] Confirm success toast appears.
- [ ] Confirm browser storage contains the teacher session with `roomCode` and `teacherToken`.
- [ ] Confirm navigation goes to `/teacher/rooms/:roomCode/dashboard` using the real room code.
- [ ] Refresh the dashboard; it should still load the room via teacher session.
- [ ] Dashboard shows real room title, student count, group count, task count, and status.
- [ ] Open `/teacher/rooms/:roomCode/course`.
- [ ] Publish a text task with title, description, deadline, target type, and optional target groups.
- [ ] Confirm success toast appears and the task list refreshes.
- [ ] Pause a published task and confirm status refreshes.
- [ ] Close a task and confirm status refreshes.
- [ ] End the classroom from teacher actions; confirm browser confirmation appears before the API call.
- [ ] After ending, room status changes to `ended`.
- [ ] After ending, publishing and task status actions are disabled or show a clear ended-state message.
- [ ] Missing, expired, or wrong teacher token shows an error and offers a path back to room creation.

## Student Flow

- [ ] Open `/student?room=:roomCode` from the teacher-created room.
- [ ] Student entry loads real room title, room code, groups, capacities, and availability.
- [ ] Join with a nickname and available group.
- [ ] Confirm success toast appears and student session is stored with `studentId`, `roomCode`, `nickname`, `groupId`, `groupName`, and `clientToken`.
- [ ] Refresh or reopen `/student?room=:roomCode`; resume should succeed and navigate to `/student/classroom`.
- [ ] With an invalid or stale token, resume clears local student session and allows rejoin.
- [ ] Classroom loads real tasks, results, and ranking with `X-Student-Token`.
- [ ] Open a task detail page.
- [ ] Submit a text answer.
- [ ] Confirm duplicate submit is blocked and the backend submission status is shown.
- [ ] Paused, closed, expired, or ended tasks cannot be submitted.
- [ ] After teacher grading, classroom shows score and comment from `/api/student/me/results`.
- [ ] Ranking updates after scoring.
- [ ] Missing or expired student token shows an error with a rejoin path.

## Review Flow

- [ ] Open `/teacher/rooms/:roomCode/review`.
- [ ] Select the published task.
- [ ] Confirm student submissions appear after a student submits.
- [ ] Filter submissions by `All`, `Submitted`, `Pending`, and `Reviewed`.
- [ ] Search by student nickname.
- [ ] Save score and feedback.
- [ ] Confirm success toast appears and submissions refresh.
- [ ] Student results and ranking update after scoring.
- [ ] Feature the current answer.
- [ ] Confirm success toast appears.
- [ ] PDF annotation/upload remains a visible disabled placeholder only.

## Display Flow

- [ ] Open `/teacher/rooms/:roomCode/display`.
- [ ] Display loads without `X-Teacher-Token`.
- [ ] Display shows real room title, status, groups, current task, ranking, and featured answers.
- [ ] Empty states appear when there are no groups, rankings, current tasks, or featured answers.
- [ ] After teacher features an answer, the display refreshes and shows the featured content.
- [ ] After the room ends, display shows the ended state.
- [ ] Display layout remains usable at desktop and narrow mobile widths.

## WebSocket Flow

- [ ] Teacher pages reconnect and refetch current data after `student_joined`, `task_published`, `task_paused`, `task_closed`, `submission_created`, `score_updated`, `ranking_updated`, or `room_ended` events relevant to the page.
- [ ] Student classroom refetches after `task_published`, `task_paused`, `task_closed`, `score_updated`, `ranking_updated`, and `room_ended`.
- [ ] Student task detail refetches after `task_paused`, `task_closed`, and `room_ended`.
- [ ] Display refetches after `ranking_updated`, `featured_answer_updated`, `submission_created`, and `room_ended`.
- [ ] Reconnecting a WebSocket triggers the page's current-data refetch.
- [ ] Student WebSocket requires `clientToken`; teacher and display WebSockets do not expose tokens in logs.
- [ ] If `VITE_WS_BASE_URL` is omitted in local dev, the same-origin `/ws` fallback works through the Vite proxy.

## Export Flow

- [ ] Export action calls `GET /api/teacher/rooms/:roomCode/export` with `X-Teacher-Token`.
- [ ] Successful export downloads a `.zip` file.
- [ ] Zip contains task/submission data for the room.
- [ ] Export remains available after the classroom has ended.
- [ ] Export failures show an error toast/message.
- [ ] Missing or invalid teacher token prevents export and shows a session error.

## Mobile Smoke

- [ ] Student entry at a narrow viewport does not overlap text or controls.
- [ ] Student classroom at a narrow viewport can scroll through status, tasks, feedback, ranking, and progress.
- [ ] Student task detail at a narrow viewport can switch between text and image tabs and submit text.
- [ ] Teacher dashboard/course/review pages remain scrollable at narrow widths, even if dense admin controls wrap.
- [ ] Display page remains readable at narrow widths and shows ended/empty states without layout breakage.
