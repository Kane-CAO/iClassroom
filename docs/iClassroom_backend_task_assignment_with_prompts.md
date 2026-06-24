# iClassroom V2.0 后端任务分工与 Prompt 使用说明

> 本文档替代旧版 V1.x 后端 Prompt 分工。V2.0 新增管理者、讲师账号体系和文件附件模型，旧的 `teacherToken` 主线不再作为目标架构。

---

## 1. 使用原则

```text
一个成员只做自己的模块
一个 Prompt 只覆盖一个阶段
一个阶段完成后单独 PR
不要让 AI 一次性生成完整 V2.0 后端
```

所有后端成员必须先阅读：

1. `README.md`
2. `Claude.md`
3. `docs/api.md`
4. `docs/database.md`
5. `docs/V2_ENGINEERING_PLAN.md`
6. `docs/V2_PROMPT_PACK.md`

---

## 2. 后端成员分工

| 成员 | 角色 | 负责模块 | 主要 Prompt |
| --- | --- | --- | --- |
| Backend A | 账号权限 Owner | 数据库迁移、管理者 / 讲师账号、登录鉴权、房间归属迁移 | Prompt 0-5 |
| Backend B | 活动业务 Owner | 观众入场、任务发布、截止 / 暂停 / 关闭、提交基础校验 | Prompt 6-8 |
| Backend C | 上传与评分 Owner | 图片 / 文件上传、附件模型、审阅评分、小组分数事务 | Prompt 8-9 |
| Backend D | 数据与导出 Owner | 排行榜、大屏、后台数据、精选答案、归档导出 | Prompt 10-12 |

如果只有 3 名后端：

- A：账号权限 + 房间迁移。
- B：观众入场 + 任务 + 提交上传。
- C：评分 + 看板 + 导出。

---

## 3. Backend A：账号权限 Owner

### 负责范围

- 新增账号相关 migration。
- 新增 `admin_users`、`teacher_accounts`、`auth_sessions` 或等价模型。
- 管理者登录。
- 讲师登录。
- 管理者创建、停用、删除、重置讲师账号。
- 讲师鉴权中间件。
- 管理者鉴权中间件。
- 房间关联 `teacher_id`。
- 将老师接口从旧 `teacherToken` 模式迁移到登录态模式。

### 主要接口

```http
POST /api/auth/admin/login
POST /api/auth/teacher/login
POST /api/auth/logout
GET  /api/auth/me

POST   /api/admin/teachers
GET    /api/admin/teachers
PATCH  /api/admin/teachers/:teacherId/status
POST   /api/admin/teachers/:teacherId/reset-password
DELETE /api/admin/teachers/:teacherId

POST /api/teacher/rooms
GET  /api/teacher/rooms
```

### 禁止

- 不要实现上传、评分、导出。
- 不要删除旧表字段，先兼容迁移。
- 不要让讲师访问其他讲师房间。

---

## 4. Backend B：活动业务 Owner

### 负责范围

- 观众获取房间信息。
- 观众加入房间。
- 昵称验重。
- 小组满员校验。
- 观众恢复会话。
- 任务发布。
- 任务目标小组。
- 暂停 / 关闭 / 截止校验。

### 主要接口

```http
GET  /api/student/rooms/:roomCode
POST /api/student/rooms/:roomCode/join
POST /api/student/rooms/:roomCode/resume

POST  /api/teacher/rooms/:roomCode/tasks
GET   /api/teacher/rooms/:roomCode/tasks
PATCH /api/teacher/tasks/:taskId/pause
PATCH /api/teacher/tasks/:taskId/close

GET /api/student/me/tasks
GET /api/student/tasks/:taskId
```

### 禁止

- 不要绕过 A 的讲师登录鉴权。
- 不要实现通用文件存储。
- 不要修改导出逻辑。

---

## 5. Backend C：上传与评分 Owner

### 负责范围

- 统一附件模型 `submission_attachments`。
- 图片上传限制：单图 <= 5MB，单题最多 3 张。
- 文件上传限制：不含视频，大小和类型按 `docs/api.md`。
- 提交接口 multipart。
- 讲师查看提交。
- 手动评分。
- 小组总分事务更新。

### 主要接口

```http
POST /api/student/tasks/:taskId/submit
GET  /api/teacher/tasks/:taskId/submissions
POST /api/teacher/submissions/:submissionId/grade
GET  /api/student/me/results
```

### 禁止

- 不要改账号体系。
- 不要把文件原始名直接作为存储路径。
- 不要允许视频上传。
- 不要允许 0 分。

---

## 6. Backend D：数据看板与导出 Owner

### 负责范围

- 学生 / 大屏排行榜。
- 任务完成度。
- 精选答案。
- 后台数据看板。
- 活动结束。
- Excel + 图片 / 文件原件 Zip 导出。
- 30 天留存和清理策略预留。

### 主要接口

```http
GET  /api/student/rooms/:roomCode/ranking
GET  /api/teacher/rooms/:roomCode/display
POST /api/teacher/submissions/:submissionId/feature
GET  /api/teacher/rooms/:roomCode/analytics
POST /api/teacher/rooms/:roomCode/end
GET  /api/teacher/rooms/:roomCode/export
```

### 禁止

- 不要重写提交和评分核心逻辑。
- 不要让已结束活动允许观众继续提交。
- 不要导出服务端绝对路径。

---

## 7. PR 验收标准

每个 PR 必须包含：

- 修改文件清单。
- 新增接口说明。
- 数据库变更说明。
- 测试命令和结果。
- 手动验收步骤。
- 风险和后续 TODO。

后端默认测试：

```bash
go test ./...
```

涉及数据库 migration 的 PR 还必须说明：

- up 迁移。
- down 迁移。
- 旧数据兼容方式。
- 本地验证方式。

---

## 8. 旧文档迁移提醒

旧版文档中的以下结论已经废弃：

- 老师端不做账号登录。
- 老师端使用 `roomCode + teacherToken` 管理房间。
- 讲师账号体系作为二期。
- 只实现图片上传，不实现通用文件上传。

V2.0 以后以后端 `docs/api.md`、`docs/database.md` 和本文件为准。
