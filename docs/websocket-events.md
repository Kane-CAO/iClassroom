# WebSocket 联调说明

本文档用于 Prompt 13 WebSocket 实时同步的前后端联调和人工检查。

## 1. 后端结构

后端 WebSocket 入口：

```text
GET /ws?room=ABC123&role=teacher
GET /ws?room=ABC123&role=student&token=xxx
GET /ws?room=ABC123&role=display
```

主要文件：

```text
backend/internal/websocket/event.go       # Event / EventType / NewEvent
backend/internal/websocket/client.go      # Client、读写 pump、ping/pong、断开清理
backend/internal/websocket/room_hub.go    # 单 roomCode 连接池和广播
backend/internal/websocket/hub_manager.go # 每个 roomCode 一个 RoomHub
backend/internal/handler/ws.go            # GET /ws 鉴权后 upgrade
backend/internal/service/ws_auth.go       # room/role/student token 鉴权
backend/internal/service/broadcast.go     # service 层广播接口和 emit
backend/cmd/server/main.go                # 创建 HubManager 并注入 service
```

结构关系：

```text
HTTP GET /ws
  -> handler.WSHandler
  -> service.WSAuthService.Authorize
  -> websocket.HubManager.Serve
  -> websocket.RoomHub
  -> websocket.Client read/write pumps
```

广播路径：

```text
业务 service 完成数据库操作
  -> service.emit(...)
  -> websocket.HubManager.Broadcast(roomCode, event)
  -> RoomHub.Broadcast(event)
  -> Client.enqueue(payload)
```

关键约束：

- 每个 `roomCode` 一个连接池。
- `teacher`、`student`、`display` 可以订阅同一个 room。
- `student` 连接必须校验 `token`，并确认 token 属于当前 room。
- `teacher` 和 `display` 当前只要求 room 存在。
- 断开连接后，`Client.cleanup` 会从 `HubManager` 中清理连接；空 room hub 会被删除。
- `HubManager` 和 `RoomHub` 内部 map 均使用 mutex 保护。
- 广播失败只记录日志或丢弃慢连接，不影响主业务返回。

## 2. 前端封装位置

基础封装：

```text
frontend/src/api/websocket.ts
```

提供：

- `RoomWebSocketRole`
- `RoomWebSocketEventType`
- `RoomWebSocketEvent`
- `buildRoomWebSocketURL`
- `createRoomWebSocket`
- `parseRoomWebSocketEvent`

React hook：

```text
frontend/src/hooks/useRoomWebSocket.ts
```

支持：

- `onEvent`
- `onReconnect`
- 自动重连
- 组件卸载时关闭连接
- 避免重复连接
- `student` 缺少 `clientToken` 时不发起连接

当前页面接入：

```text
frontend/src/pages/teacher/Dashboard.tsx
frontend/src/pages/teacher/Review.tsx
frontend/src/pages/student/Classroom.tsx
frontend/src/pages/student/TaskDetail.tsx
frontend/src/pages/display/Display.tsx
```

注意：当前部分页面仍使用 mock 数据。页面已在收到事件和重连后调用本地 `refresh...Data` 占位函数，并保留 TODO；真实 API 接入后，应在这些 TODO 位置重新拉取当前页面数据。

## 3. 事件格式

统一 JSON envelope：

```json
{
  "type": "task_published",
  "roomCode": "ABC123",
  "data": {},
  "occurredAt": "2026-06-16T10:00:00Z"
}
```

字段说明：

```text
type       事件类型，见下表
roomCode   房间码
data       最小必要 payload；前端收到后以重新拉取页面数据为主
occurredAt 后端生成的 UTC 时间
```

事件列表：

```text
student_joined
task_published
task_paused
task_closed
submission_created
score_updated
ranking_updated
featured_answer_updated
room_ended
```

当前后端 payload：

```json
{
  "type": "student_joined",
  "data": {
    "studentId": 1,
    "nickname": "Tom",
    "groupId": 1
  }
}
```

```json
{
  "type": "task_published",
  "data": {
    "taskId": 1,
    "status": "published",
    "targetType": "all"
  }
}
```

```json
{
  "type": "task_paused",
  "data": {
    "taskId": 1,
    "status": "paused"
  }
}
```

```json
{
  "type": "task_closed",
  "data": {
    "taskId": 1,
    "status": "closed"
  }
}
```

```json
{
  "type": "submission_created",
  "data": {
    "submissionId": 1,
    "taskId": 1,
    "studentId": 1,
    "groupId": 1
  }
}
```

```json
{
  "type": "score_updated",
  "data": {
    "submissionId": 1,
    "score": 8,
    "groupId": 1
  }
}
```

```json
{
  "type": "ranking_updated",
  "data": {
    "groupId": 1,
    "groupScoreTotal": 18
  }
}
```

```json
{
  "type": "featured_answer_updated",
  "data": {
    "featuredId": 1,
    "submissionId": 1,
    "displayMode": "anonymous"
  }
}
```

```json
{
  "type": "room_ended",
  "data": {
    "status": "ended"
  }
}
```

## 4. 多浏览器测试步骤

### 4.1 启动服务

后端：

```bash
cd backend
go run ./cmd/server
```

前端：

```bash
cd frontend
npm run dev
```

确认环境变量：

```text
frontend/.env
VITE_API_BASE_URL=http://localhost:8080
VITE_WS_BASE_URL=ws://localhost:8080/ws
VITE_STUDENT_BASE_URL=http://localhost:5173/student
```

### 4.2 准备三个窗口

1. 打开 teacher 窗口。
2. 打开 student 窗口，建议使用无痕窗口或另一个浏览器，避免 localStorage 混用。
3. 打开 display 窗口。
4. 打开三个窗口的 DevTools，切到 Network -> WS，观察 `/ws` 连接和 Messages。

建议 URL 形态：

```text
teacher: http://localhost:5173/teacher/rooms/ABC123/dashboard
review:  http://localhost:5173/teacher/rooms/ABC123/review
student: http://localhost:5173/student?room=ABC123
display: http://localhost:5173/teacher/rooms/ABC123/display
```

实际 roomCode 以创建房间接口返回值为准。

### 4.3 联调流程

1. 创建房间。
   - 通过老师端创建房间，或调用 `POST /api/teacher/rooms`。
   - 记录 `roomCode` 和 `teacherToken`。

2. 打开 teacher、student、display 三个浏览器窗口。
   - teacher 使用该 `roomCode` 的页面。
   - student 通过 `/student?room={roomCode}` 加入。
   - display 打开 `/teacher/rooms/{roomCode}/display`。

3. 学生加入。
   - student 选择小组并加入房间。
   - teacher 窗口的 WS Messages 应收到 `student_joined`。
   - teacher 页面收到事件后应触发当前页面数据刷新逻辑。

4. 老师发布任务。
   - teacher 调用发布任务接口或页面操作。
   - student 窗口的 WS Messages 应收到 `task_published`。
   - student 页面收到事件后应重新拉取任务/课堂数据。

5. 学生提交。
   - student 提交当前任务。
   - teacher 窗口的 WS Messages 应收到 `submission_created`。
   - display 窗口也应收到 `submission_created`。
   - teacher/review/display 页面收到事件后应重新拉取当前页面数据。

6. 老师评分。
   - teacher 对 submission 评分。
   - student 窗口应收到 `score_updated` 和/或 `ranking_updated`。
   - display 窗口应收到 `score_updated` / `ranking_updated`。
   - 大屏收到后应重新拉取排行榜/display 数据。

7. 老师精选答案。
   - teacher 设置精选答案。
   - display 窗口应收到 `featured_answer_updated`。
   - 大屏收到后应重新拉取 display 数据。

8. 老师结束课堂。
   - teacher 调用结束课堂接口或页面操作。
   - student 和 display 窗口应收到 `room_ended`。
   - student 页面应重新拉取课堂状态；后续真实 API 接入后，应提示课堂已结束或阻止继续提交。

9. 刷新或断网重连。
   - 刷新 student/display 页面，确认只建立一个 `/ws` 连接。
   - 临时断网或停止后端，再恢复后端。
   - 前端 hook 应自动重连。
   - 重连成功后 `onReconnect` 应触发，并重新拉取当前页面数据。

## 5. 人工检查点

后端：

- [ ] `/ws` 挂在根路径，不是 `/api/ws`。
- [ ] `room` 不存在时不能 upgrade。
- [ ] 非法 `role` 不能 upgrade。
- [ ] `student` 缺少 token 时不能 upgrade。
- [ ] `student` token 不属于当前 room 时不能 upgrade。
- [ ] 同一个 room 内 teacher、student、display 都能同时连接。
- [ ] 每个 roomCode 只有一个连接池；不同 roomCode 互不广播。
- [ ] 断开浏览器窗口后连接池数量会下降，空 room hub 会被清理。
- [ ] map 并发读写有锁保护。
- [ ] 业务 service 是数据库操作成功后再广播。
- [ ] 广播失败不影响 HTTP 主业务响应。

前端：

- [ ] `VITE_WS_BASE_URL` 未配置时不发起错误连接。
- [ ] student 没有 `clientToken` 时不发起 WebSocket。
- [ ] 组件卸载时关闭连接。
- [ ] 页面刷新后不会产生多个重复连接。
- [ ] 收到事件后不直接盲目插入重复数据，而是触发当前页面数据刷新。
- [ ] 重连后调用 `onReconnect`，重新拉取当前页面数据。
- [ ] teacher dashboard/review 监听：
  `student_joined`、`submission_created`、`score_updated`、`ranking_updated`、`room_ended`。
- [ ] student classroom/task detail 监听：
  `task_published`、`task_paused`、`task_closed`、`score_updated`、`ranking_updated`、`room_ended`。
- [ ] display 监听：
  `ranking_updated`、`featured_answer_updated`、`submission_created`、`room_ended`。

联调完成标准：

- [ ] `go test ./...` 通过。
- [ ] `npm run build` 通过。
- [ ] 多浏览器流程中所有目标事件都能在 WS Messages 中看到。
- [ ] 刷新或断线重连后，当前页面数据刷新逻辑被触发。
