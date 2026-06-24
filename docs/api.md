# iClassroom V2.0 API Contract

> 本文档是 V2.0 目标 API 契约，用于前后端联调和 Prompt 开发对齐。
> 当前代码可能仍包含 V1.x 的 `X-Teacher-Token` 模式；V2.0 新开发以本文档为目标。

---

## 1. 通用约定

- Base URL：`http://localhost:8080`
- 前端通过 `VITE_API_BASE_URL` 注入，禁止在页面中硬编码。
- 所有业务接口统一带 `/api` 前缀。
- 所有时间字段使用 UTC ISO-8601。
- API 字段统一使用 `camelCase`。
- 数据库字段统一使用 `snake_case`。

---

## 2. 统一响应

成功：

```json
{
  "success": true,
  "message": "success",
  "data": {}
}
```

失败：

```json
{
  "success": false,
  "message": "teacher account disabled",
  "errorCode": "TEACHER_DISABLED"
}
```

导出接口返回二进制文件，不包统一 JSON。

---

## 3. 身份与 Header

### 3.1 管理者接口

```http
Authorization: Bearer admin_session_token
```

### 3.2 讲师接口

```http
Authorization: Bearer teacher_session_token
```

### 3.3 观众接口

```http
X-Student-Token: student_xxxxx
```

观众入场前接口不需要 token。

### 3.4 迁移说明

旧版 `X-Teacher-Token` 仅作为兼容或迁移参考。V2.0 新接口不得依赖房间级 `teacherToken` 作为主要鉴权方式。

---

## 4. API 总览

| 模块 | Method | Path | 权限 |
| --- | --- | --- | --- |
| 管理者登录 | POST | `/api/auth/admin/login` | 公开 |
| 讲师登录 | POST | `/api/auth/teacher/login` | 公开 |
| 登出 | POST | `/api/auth/logout` | 管理者 / 讲师 |
| 当前用户 | GET | `/api/auth/me` | 管理者 / 讲师 |
| 创建讲师 | POST | `/api/admin/teachers` | 管理者 |
| 讲师列表 | GET | `/api/admin/teachers` | 管理者 |
| 更新讲师状态 | PATCH | `/api/admin/teachers/:teacherId/status` | 管理者 |
| 重置讲师密码 | POST | `/api/admin/teachers/:teacherId/reset-password` | 管理者 |
| 删除讲师 | DELETE | `/api/admin/teachers/:teacherId` | 管理者 |
| 创建房间 | POST | `/api/teacher/rooms` | 讲师 |
| 讲师房间列表 | GET | `/api/teacher/rooms` | 讲师 |
| 房间详情 | GET | `/api/teacher/rooms/:roomCode` | 讲师本人 |
| 房间 overview | GET | `/api/teacher/rooms/:roomCode/overview` | 讲师本人 |
| 结束活动 | POST | `/api/teacher/rooms/:roomCode/end` | 讲师本人 |
| 观众获取房间 | GET | `/api/student/rooms/:roomCode` | 公开 |
| 观众加入 | POST | `/api/student/rooms/:roomCode/join` | 公开 |
| 观众恢复会话 | POST | `/api/student/rooms/:roomCode/resume` | 观众 |
| 发布任务 | POST | `/api/teacher/rooms/:roomCode/tasks` | 讲师本人 |
| 讲师任务列表 | GET | `/api/teacher/rooms/:roomCode/tasks` | 讲师本人 |
| 暂停任务 | PATCH | `/api/teacher/tasks/:taskId/pause` | 讲师本人 |
| 关闭任务 | PATCH | `/api/teacher/tasks/:taskId/close` | 讲师本人 |
| 观众任务列表 | GET | `/api/student/me/tasks` | 观众 |
| 观众任务详情 | GET | `/api/student/tasks/:taskId` | 观众 |
| 提交任务 | POST | `/api/student/tasks/:taskId/submit` | 观众 |
| 查看提交 | GET | `/api/teacher/tasks/:taskId/submissions` | 讲师本人 |
| 评分 | POST | `/api/teacher/submissions/:submissionId/grade` | 讲师本人 |
| 观众结果 | GET | `/api/student/me/results` | 观众 |
| 小组排名 | GET | `/api/student/rooms/:roomCode/ranking` | 观众 |
| 大屏数据 | GET | `/api/teacher/rooms/:roomCode/display` | 讲师本人 |
| 精选答案 | POST | `/api/teacher/submissions/:submissionId/feature` | 讲师本人 |
| 后台数据 | GET | `/api/teacher/rooms/:roomCode/analytics` | 讲师本人 |
| 导出 | GET | `/api/teacher/rooms/:roomCode/export` | 讲师本人 |

---

## 5. 账号接口

### 5.1 管理者登录

```http
POST /api/auth/admin/login
```

Request：

```json
{
  "username": "admin",
  "password": "password"
}
```

Response：

```json
{
  "success": true,
  "message": "success",
  "data": {
    "token": "admin_session_token",
    "user": {
      "userId": 1,
      "role": "admin",
      "username": "admin"
    }
  }
}
```

### 5.2 讲师登录

```http
POST /api/auth/teacher/login
```

Request：

```json
{
  "username": "teacher01",
  "password": "initialPassword"
}
```

Response：

```json
{
  "success": true,
  "message": "success",
  "data": {
    "token": "teacher_session_token",
    "user": {
      "userId": 12,
      "role": "teacher",
      "username": "teacher01",
      "displayName": "李老师"
    }
  }
}
```

### 5.3 创建讲师

```http
POST /api/admin/teachers
Authorization: Bearer admin_session_token
```

Request：

```json
{
  "username": "teacher01",
  "displayName": "李老师",
  "initialPassword": "ChangeMe123"
}
```

Response：

```json
{
  "success": true,
  "message": "success",
  "data": {
    "teacherId": 12,
    "username": "teacher01",
    "displayName": "李老师",
    "status": "active",
    "createdAt": "2026-06-24T08:00:00Z"
  }
}
```

---

## 6. 房间接口

### 6.1 创建房间

```http
POST /api/teacher/rooms
Authorization: Bearer teacher_session_token
```

Request：

```json
{
  "title": "AITIC 展厅互动活动",
  "groupCount": 6,
  "groupCapacity": 10,
  "allowChooseGroup": true
}
```

Response：

```json
{
  "success": true,
  "message": "success",
  "data": {
    "roomCode": "ABC123",
    "joinUrl": "https://example.com/room?id=ABC123",
    "displayUrl": "https://example.com/display/ABC123",
    "groups": [
      { "groupId": 1, "groupName": "第1组", "capacity": 10, "currentCount": 0 }
    ]
  }
}
```

规则：

- 房间必须关联当前登录讲师。
- `roomCode` 全局唯一。
- 创建房间和创建默认小组必须在同一事务内完成。

---

## 7. 观众入场接口

### 7.1 获取房间信息

```http
GET /api/student/rooms/:roomCode
```

返回活动名称、状态、可选小组、每组当前人数。

### 7.2 加入房间

```http
POST /api/student/rooms/:roomCode/join
```

Request：

```json
{
  "nickname": "Tom",
  "groupId": 1
}
```

Response：

```json
{
  "success": true,
  "message": "success",
  "data": {
    "studentId": 100,
    "clientToken": "student_xxxxx",
    "roomCode": "ABC123",
    "nickname": "Tom",
    "groupId": 1,
    "groupName": "第1组"
  }
}
```

规则：

- 房间 ended 后不能加入。
- 同一房间昵称唯一。
- 小组满员不能加入。

---

## 8. 任务与提交

### 8.1 发布任务

```http
POST /api/teacher/rooms/:roomCode/tasks
Authorization: Bearer teacher_session_token
```

Request：

```json
{
  "title": "方案亮点提交",
  "description": "请提交本组方案的核心亮点。",
  "deadlineAt": "2026-06-24T10:00:00Z",
  "targetType": "all",
  "targetGroupIds": []
}
```

### 8.2 提交任务

```http
POST /api/student/tasks/:taskId/submit
Content-Type: multipart/form-data
X-Student-Token: student_xxxxx
```

FormData：

```text
contentText: 文字答案
images[]: image1.jpg
images[]: image2.png
files[]: report.pdf
files[]: material.docx
```

规则：

- 一名观众对一个任务只能提交一次。
- 支持纯文字、文字 + 图片、文字 + 文件、文字 + 图片 + 文件。
- 图片单图 <= 5MB，单题最多 3 张。
- 文件不含视频；首版建议限制 PDF、Office 文档、图片、压缩包，具体大小由技术确认。
- 任务 paused、closed、超过截止时间或房间 ended 后不能提交。

---

## 9. 评分、看板、导出

### 9.1 评分

```http
POST /api/teacher/submissions/:submissionId/grade
Authorization: Bearer teacher_session_token
```

Request：

```json
{
  "score": 8,
  "comment": "观点清楚，材料完整。"
}
```

规则：

- 分数为整数 1-10。
- 评分后按差值更新小组总分。

### 9.2 大屏数据

```http
GET /api/teacher/rooms/:roomCode/display
Authorization: Bearer teacher_session_token
```

返回排行榜、当前任务完成度、精选答案。

### 9.3 后台数据

```http
GET /api/teacher/rooms/:roomCode/analytics
Authorization: Bearer teacher_session_token
```

返回整体参与率、各小组得分对比、任务完成情况、提交时间分布。

### 9.4 导出

```http
GET /api/teacher/rooms/:roomCode/export
Authorization: Bearer teacher_session_token
```

返回：

```text
export_room_ABC123.zip
├── submissions.xlsx
├── images/
└── files/
```

---

## 10. 错误码

| errorCode | HTTP | 说明 |
| --- | ---: | --- |
| `INVALID_REQUEST` | 400 | 请求格式错误 |
| `INVALID_CREDENTIALS` | 401 | 账号或密码错误 |
| `UNAUTHORIZED` | 401 | 未登录或会话失效 |
| `FORBIDDEN` | 403 | 无权访问 |
| `TEACHER_DISABLED` | 403 | 讲师账号已停用 |
| `ROOM_NOT_FOUND` | 404 | 房间不存在 |
| `ROOM_ENDED` | 409 | 活动已结束 |
| `NICKNAME_DUPLICATED` | 409 | 昵称重复 |
| `GROUP_FULL` | 409 | 小组已满 |
| `TASK_NOT_FOUND` | 404 | 任务不存在 |
| `TASK_CLOSED` | 409 | 任务已关闭 |
| `TASK_PAUSED` | 409 | 任务已暂停 |
| `TASK_DEADLINE_PASSED` | 409 | 任务已截止 |
| `SUBMISSION_ALREADY_EXISTS` | 409 | 已提交过该任务 |
| `INVALID_SCORE` | 400 | 分数不合法 |
| `TOO_MANY_IMAGES` | 400 | 图片数量超限 |
| `IMAGE_TOO_LARGE` | 400 | 图片过大 |
| `INVALID_IMAGE_TYPE` | 400 | 图片格式不支持 |
| `FILE_TOO_LARGE` | 400 | 文件过大 |
| `INVALID_FILE_TYPE` | 400 | 文件类型不支持 |
| `UPLOAD_FAILED` | 500 | 上传失败 |
| `EXPORT_FAILED` | 500 | 导出失败 |

---

## 11. 待技术确认

- 管理者初始账号如何初始化。
- 登录 token 使用 JWT、服务端 session 还是混合方案。
- 文件大小和文件类型白名单。
- 生产文件存储使用 OSS、S3 还是 MinIO。
- WebSocket 鉴权是否复用登录 token。
- 活动数据和附件 30 天清理任务的执行方式。
