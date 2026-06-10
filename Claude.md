# iClassroom_AITIC_Claude.md

本文件用于指导 Claude API / Claude Code 按照大厂工程化标准逐步开发 iClassroom MVP。  
Claude 在开发时必须严格遵守本文档的开发顺序、分层结构、测试要求和人工检查点。

---

## 0. 项目目标

开发一个轻量级线上课堂互动平台 iClassroom。

MVP 主流程：

```text
讲师建房 → 学生扫码/链接入场 → 学生填写昵称并选组 → 讲师发布任务 → 学生提交文字/图片 → 讲师审阅评分 → 小组排名更新 → 大屏展示 → 课后结束课堂 → 导出 Excel + 图片 Zip
```

核心原则：

```text
先跑通核心闭环，再做实时同步；
先实现 HTTP 接口，再补 WebSocket；
先实现业务正确性，再优化 UI 和工程细节；
每个阶段都必须有人为检查和测试。
```

---

## 1. Claude 开发行为准则

Claude 在写代码前必须先做以下动作：

1. 阅读 `README.md`。
2. 阅读本文件 `CLAUDE.md`。
3. 检查当前项目目录结构。
4. 检查已有代码，不要重复创建已有模块。
5. 每次只开发一个阶段，不要跨阶段一次性生成大量功能。
6. 每个阶段完成后必须输出：
   - 修改了哪些文件。
   - 新增了哪些接口或页面。
   - 如何本地运行。
   - 如何手动测试。
   - 哪些部分需要人类检查。
7. 不允许跳过测试。
8. 不允许提交 `.env`、密钥、token、数据库密码。
9. 不允许为了“跑通”而删除已有业务校验。
10. 不允许把所有逻辑塞进 handler 或前端页面，必须分层。

---

## 2. 推荐技术架构

### 2.1 前端

推荐结构：

```text
frontend/src/
├── api/
├── router/
├── pages/
│   ├── teacher/
│   ├── student/
│   └── display/
├── components/
├── stores/
├── hooks/
├── utils/
└── types/
```

前端必须做到：

- API 请求集中封装在 `api/`。
- 页面组件不直接写复杂业务请求。
- 类型定义放在 `types/`。
- localStorage 读写封装在 `utils/storage.ts`。
- WebSocket 连接封装在 `hooks/` 或 `utils/`。
- 学生端移动端优先。
- 讲师端电脑端优先。
- 大屏端全屏展示优先。

### 2.2 后端

推荐结构：

```text
backend/
├── cmd/server/main.go
├── internal/
│   ├── config/
│   ├── domain/
│   ├── repository/
│   ├── service/
│   ├── handler/
│   ├── middleware/
│   ├── websocket/
│   ├── storage/
│   ├── export/
│   └── response/
├── migrations/
├── uploads/
└── .env.example
```

后端必须分层：

```text
handler：只负责 HTTP 参数绑定、调用 service、返回 response
service：负责业务规则、事务、校验、状态流转
repository：负责数据库读写
domain/model：负责实体定义
middleware：负责跨域、身份校验、日志等
storage：负责图片存储
export：负责 Excel 和 Zip 导出
websocket：负责实时连接和事件广播
```

禁止：

- handler 直接写复杂 SQL。
- handler 直接操作多个表完成业务。
- 前端页面里硬编码大量接口 URL。
- 把 WebSocket 逻辑混进普通业务 service。
- 忽略错误处理。

---

## 3. 统一业务规则

### 3.1 房间规则

- 房间由讲师创建。
- 系统生成唯一 roomCode。
- 房间状态：
  - `created`
  - `active`
  - `ended`
- MVP 中创建后可以直接视为 active。
- 房间 ended 后，学生不能加入、不能恢复、不能提交。
- 讲师端仍然可以查看和导出数据。

### 3.2 学生规则

- 学生无需注册。
- 学生进入房间时填写昵称并选择小组。
- 同一房间内昵称唯一。
- 小组满员后不可加入。
- 后端生成 `clientToken`。
- 前端将 `roomCode`、`studentId`、`clientToken`、`nickname`、`groupId` 存入 localStorage。
- 断线或刷新后使用 `clientToken` 恢复身份。

### 3.3 任务规则

- 讲师发布任务。
- 任务包含：
  - 标题
  - 描述
  - 可选附件链接
  - 截止时间
  - 发布范围：全班或指定小组
- 任务状态：
  - `published`
  - `paused`
  - `closed`
- 任务 paused、closed 或超过 deadline 后，学生不能提交。
- P1 选择题不做。

### 3.4 提交规则

- 学生对一个任务只能提交一次。
- 提交支持文字和图片。
- 图片直接上传。
- 单图 ≤ 5MB。
- 单题 ≤ 3 张。
- 图片上传失败时前端应保留文字草稿。
- 任务截止后不能提交。
- 已提交内容不受任务截止影响。

### 3.5 评分规则

- 讲师手动评分。
- 分数必须是整数 1–10。
- 最低 1 分，不允许 0 分。
- 可填写评语。
- 保存评分后，submission 状态变为 `graded`。
- 小组总分自动累计。
- 排行榜按小组总分降序排列。

### 3.6 大屏规则

- 大屏是讲师端的全屏展示视图。
- 展示内容：
  - 小组排行榜
  - 当前任务完成度
  - 倒计时
  - 精选答案
- 精选答案可匿名或显示小组。

### 3.7 导出规则

- 课程结束后关闭学生入口。
- 导出包含 Excel 和图片原件 Zip。
- Excel 至少包含：
  - 房间号
  - 课堂名称
  - 小组
  - 学生昵称
  - 任务标题
  - 提交文字
  - 图片文件名
  - 分数
  - 评语
  - 提交时间
  - 评分时间
- Zip 结构建议：

```text
export_room_ABC123.zip
├── submissions.xlsx
└── images/
    ├── task_1/
    │   ├── group_1/
    │   │   └── student_tom_img1.jpg
```

---

## 4. 后端开发阶段

Claude 必须按照以下顺序开发后端。

---

### Backend Step 0：项目初始化

目标：后端服务能启动，连接数据库，提供健康检查。

任务：

1. 创建 Go module。
2. 创建目录结构。
3. 创建 config 读取 `.env`。
4. 创建数据库连接。
5. 创建统一 response。
6. 创建 Gin router。
7. 创建 `/health` 接口。
8. 创建 `.env.example`。
9. 添加 CORS middleware。

输出接口：

```http
GET /health
```

完成标准：

```text
go run ./cmd/server 能启动
curl http://localhost:8080/health 返回 ok
数据库连接正常
```

人为检查：

- 检查 `.env` 是否没有被提交。
- 检查 config 是否有默认值和错误提示。
- 检查 response 格式是否统一。
- 检查 CORS 是否只用于开发环境或可配置。

测试：

```bash
go test ./...
curl http://localhost:8080/health
```

---

### Backend Step 1：数据库迁移与基础实体

目标：创建核心表结构。

任务：

1. 创建 migration 文件。
2. 创建 domain model：
   - Room
   - Group
   - Student
   - Task
   - TaskTargetGroup
   - Submission
   - SubmissionImage
   - FeaturedAnswer
3. 添加数据库索引和唯一约束。
4. 编写基础迁移说明。

必须包含约束：

```text
rooms.room_code unique
students unique(room_id, nickname)
submissions unique(task_id, student_id)
```

人为检查：

- 检查字段是否覆盖 MVP。
- 检查 enum/status 是否清晰。
- 检查索引是否满足常见查询：
  - room_code
  - room_id
  - task_id
  - group_id
  - student_id
- 检查时间字段是否统一。

测试：

```bash
go test ./...
# 手动检查数据库表是否创建成功
```

---

### Backend Step 2：创建房间与小组

目标：讲师可以创建房间，系统生成 roomCode 和小组。

接口：

```http
POST /api/teacher/rooms
GET  /api/teacher/rooms/:roomCode
GET  /api/teacher/rooms/:roomCode/overview
```

请求示例：

```json
{
  "title": "AI Workshop",
  "groupCount": 6,
  "groupCapacity": 10,
  "allowChooseGroup": true
}
```

业务规则：

- groupCount 默认 6。
- groupCapacity 默认 10。
- roomCode 必须唯一。
- 创建房间时自动创建 groups。
- 创建成功后返回 joinUrl。

后端分层：

```text
RoomHandler
RoomService
RoomRepository
GroupRepository
```

人为检查：

- 检查 roomCode 是否可能重复。
- 检查创建房间和创建小组是否在同一事务。
- 检查非法 groupCount/groupCapacity 是否被拦截。
- 检查返回字段是否方便前端使用。

测试：

```bash
go test ./...
curl -X POST http://localhost:8080/api/teacher/rooms \
  -H "Content-Type: application/json" \
  -d '{"title":"Demo Class","groupCount":6,"groupCapacity":10,"allowChooseGroup":true}'
```

---

### Backend Step 3：学生入场与断线恢复

目标：学生可以通过 roomCode 加入房间，昵称验重，小组容量校验，本地 token 恢复身份。

接口：

```http
GET  /api/student/rooms/:roomCode
POST /api/student/rooms/:roomCode/join
POST /api/student/rooms/:roomCode/resume
```

join 请求：

```json
{
  "nickname": "Tom",
  "groupId": 1
}
```

业务规则：

- 房间不存在返回 404。
- 房间 ended 返回业务错误。
- nickname 在同一 room 内不可重复。
- group 必须属于当前 room。
- group 已满不能加入。
- 创建 student 时生成 clientToken。
- resume 根据 roomCode + clientToken 恢复身份。

人为检查：

- 检查重复昵称是否真的基于 room 维度。
- 检查并发加入同一小组是否可能突破容量。
- 检查 clientToken 是否不可预测。
- 检查 ended 房间是否不可加入。

测试：

```bash
go test ./...
# 1. 正常加入
# 2. 重复昵称
# 3. 小组满员
# 4. 房间结束后加入
# 5. resume 恢复身份
```

---

### Backend Step 4：任务发布与任务列表

目标：讲师发布任务，学生看到自己范围内的任务。

接口：

```http
POST /api/teacher/rooms/:roomCode/tasks
GET  /api/teacher/rooms/:roomCode/tasks
GET  /api/student/me/tasks
PATCH /api/teacher/tasks/:taskId/pause
PATCH /api/teacher/tasks/:taskId/close
```

创建任务请求：

```json
{
  "title": "课堂测试一",
  "description": "请提交你的思路",
  "attachmentUrl": "",
  "deadlineAt": "2026-06-08T18:00:00Z",
  "targetType": "all",
  "targetGroupIds": []
}
```

业务规则：

- 标题必填。
- deadlineAt 必须晚于当前时间。
- targetType 为 `all` 或 `groups`。
- targetType 为 groups 时，targetGroupIds 不能为空。
- 指定小组必须属于当前 room。
- paused/closed 任务在学生端展示但不可提交。

人为检查：

- 检查 deadline 时区处理。
- 检查指定小组权限。
- 检查学生只能看到自己小组范围内任务。
- 检查任务状态流转是否清晰。

测试：

```bash
go test ./...
# 创建全班任务
# 创建指定小组任务
# 学生任务列表过滤
# pause / close 状态变化
```

---

### Backend Step 5：文字提交

目标：学生可以提交文字答案，讲师可以查看提交。

接口：

```http
POST /api/student/tasks/:taskId/submit
GET  /api/teacher/tasks/:taskId/submissions
```

MVP 第一版可以先只支持文字，图片在 Step 6 增加。

业务规则：

- 校验学生身份 token。
- 校验任务是否存在。
- 校验任务是否属于学生房间。
- 校验学生是否在任务目标范围。
- 校验任务未 paused、未 closed、未超过 deadline。
- 同一个学生同一任务只能提交一次。
- 提交成功后状态为 submitted。

人为检查：

- 检查重复提交是否被数据库唯一约束兜底。
- 检查截止时间是否后端强校验，不能只依赖前端。
- 检查教师查询提交是否按小组返回。
- 检查错误码是否清晰。

测试：

```bash
go test ./...
# 正常提交
# 重复提交
# 截止后提交
# 暂停后提交
# 非目标小组提交
# 讲师查看提交
```

---

### Backend Step 6：图片上传

目标：学生提交时支持图片，后端存储图片元数据，讲师端可访问图片。

接口：

```http
POST /api/student/tasks/:taskId/submit
```

请求类型：

```text
multipart/form-data
```

字段：

```text
contentText
images[]
```

业务规则：

- 图片数量 ≤ 3。
- 单图 ≤ 5MB。
- 只允许常见图片 MIME：
  - image/jpeg
  - image/png
  - image/webp
- 文件名必须重新生成，不能直接信任用户文件名。
- 保存路径建议：
  `uploads/rooms/{roomCode}/tasks/{taskId}/students/{studentId}/`
- 数据库保存 fileUrl、filePath、fileName、fileSize、mimeType。
- 如果数据库写入失败，应避免留下脏文件，或记录待清理。

人为检查：

- 检查文件大小限制是否来自后端。
- 检查 MIME 校验是否存在。
- 检查文件路径是否避免目录穿越。
- 检查图片 URL 是否能被前端访问。
- 检查上传失败时错误信息是否明确。

测试：

```bash
go test ./...
# 1 张图片提交
# 3 张图片提交
# 4 张图片被拒绝
# 超过 5MB 被拒绝
# 非图片文件被拒绝
# 讲师端返回图片 URL
```

---

### Backend Step 7：评分与排行榜

目标：讲师评分后，学生结果和小组排行榜更新。

接口：

```http
POST /api/teacher/submissions/:submissionId/grade
GET  /api/student/me/results
GET  /api/student/rooms/:roomCode/ranking
```

评分请求：

```json
{
  "score": 8,
  "comment": "思路清晰"
}
```

业务规则：

- score 必须是整数。
- score 范围 1–10。
- 保存评分时更新 submission。
- 小组总分可以实时聚合，也可以冗余保存到 groups.score_total。
- 如果允许重新评分，必须正确处理差值：
  - oldScore = 6
  - newScore = 8
  - group.score_total += 2
- 排行榜按 score_total 降序。

人为检查：

- 检查重复评分是否会重复累加。
- 检查改分是否使用差值。
- 检查事务是否覆盖 submission 更新和 group score 更新。
- 检查学生只能看自己的结果。
- 检查 ranking 排序是否稳定。

测试：

```bash
go test ./...
# 打分 1
# 打分 10
# 打分 0 被拒绝
# 打分 11 被拒绝
# 小数被拒绝
# 重新评分不重复累加
# 排行榜排序正确
```

---

### Backend Step 8：WebSocket 实时同步

目标：讲师端、学生端、大屏端实时同步任务、提交、评分、排行榜和课堂结束状态。

接口：

```http
GET /ws?room=ABC123&role=teacher
GET /ws?room=ABC123&role=student&token=xxx
GET /ws?room=ABC123&role=display
```

事件：

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

架构要求：

```text
websocket.HubManager
websocket.RoomHub
websocket.Client
websocket.Event
```

业务要求：

- 每个 roomCode 维护一个连接池。
- teacher、student、display 可以订阅同一房间。
- 学生连接时校验 token。
- 断开连接后更新 online 状态。
- 业务 service 完成数据库操作后再广播事件。
- 广播失败不能影响主业务提交。

人为检查：

- 检查连接关闭是否清理。
- 检查并发读写 map 是否加锁。
- 检查广播事件结构是否统一。
- 检查 WebSocket 不要绕过 HTTP 业务校验。
- 检查断线重连是否能重新获取最新数据。

测试：

```bash
go test ./...
# 开多个浏览器窗口手测
# 讲师发任务 → 学生端自动出现
# 学生提交 → 讲师端完成度变化
# 讲师评分 → 大屏排行榜变化
# 结束课堂 → 学生端自动提示
```

---

### Backend Step 9：大屏和精选答案

目标：讲师可选择优秀答案，大屏展示排行榜、完成度、精选答案。

接口：

```http
GET  /api/teacher/rooms/:roomCode/display
POST /api/teacher/submissions/:submissionId/feature
DELETE /api/teacher/featured-answers/:featuredId
```

业务规则：

- 精选答案必须属于当前房间。
- 可选择匿名或显示小组。
- 大屏数据源与排行榜一致。
- 完成度按当前任务或所有任务统计，MVP 可优先当前任务。

人为检查：

- 检查精选答案是否可能跨房间。
- 检查图片 URL 是否能在大屏渲染。
- 检查匿名模式是否不返回学生昵称。
- 检查大屏接口是否足够前端直接渲染。

测试：

```bash
go test ./...
# 设置精选答案
# 匿名展示
# 显示小组展示
# 排行榜数据正确
# 完成度正确
```

---

### Backend Step 10：后台数据看板

目标：讲师查看课堂数据分析。

接口：

```http
GET /api/teacher/rooms/:roomCode/analytics
```

返回数据：

```json
{
  "studentCount": 30,
  "onlineCount": 18,
  "submissionRate": 0.76,
  "groupScores": [],
  "taskCompletion": [],
  "submissionTimeline": []
}
```

MVP 统计：

- 整体参与率。
- 在线/应到人数。
- 提交率。
- 各小组得分对比。
- 任务完成情况。
- 提交时间分布。

人为检查：

- 检查统计口径是否和 PRD 一致。
- 检查除数为 0 时不会报错。
- 检查任务完成率是否按目标学生数计算。
- 检查接口性能，避免 N+1 查询。

测试：

```bash
go test ./...
# 无学生时统计
# 有学生无提交
# 有提交无评分
# 有评分后小组得分
```

---

### Backend Step 11：结束课堂与导出

目标：讲师结束课堂，学生入口关闭，讲师导出 Excel + 图片 Zip。

接口：

```http
POST /api/teacher/rooms/:roomCode/end
GET  /api/teacher/rooms/:roomCode/export
```

业务规则：

- end 后 room.status = ended。
- end 后学生 join/resume/submit 全部拒绝。
- end 后讲师仍可查看数据。
- export 生成 zip。
- zip 包含 submissions.xlsx 和 images 目录。
- 图片使用原件。
- 导出失败要返回明确错误。

人为检查：

- 检查 ended 状态是否被所有学生接口拦截。
- 检查 Excel 字段是否完整。
- 检查 Zip 结构是否清晰。
- 检查不存在图片时也能导出 Excel。
- 检查大文件导出是否可能超时，MVP 可先同步导出。

测试：

```bash
go test ./...
# 结束课堂
# 学生再次访问
# 导出 Excel
# 导出图片 Zip
# 无图片导出
# 有图片导出
```

---

## 5. 前端开发阶段

Claude 必须按照以下顺序开发前端。

---

### Frontend Step 0：项目初始化

目标：前端项目可运行，路由和 API 封装完成。

任务：

1. 创建 Vite 项目。
2. 配置 TypeScript。
3. 配置路由：
   - `/`
   - `/teacher/create-room`
   - `/teacher/rooms/:roomCode/dashboard`
   - `/teacher/rooms/:roomCode/display`
   - `/teacher/rooms/:roomCode/analytics`
   - `/student`
   - `/student/classroom`
   - `/student/tasks/:taskId`
4. 创建 API client。
5. 创建统一错误提示。
6. 创建基础 Layout。

人为检查：

- 检查移动端页面是否有 viewport。
- 检查 API base URL 是否使用环境变量。
- 检查路由命名是否清晰。
- 检查不要硬编码 localhost 到业务代码。

测试：

```bash
npm run build
npm run dev
```

---

### Frontend Step 1：讲师创建房间页

页面：

```text
/teacher/create-room
```

功能：

- 输入课堂名称。
- 设置组数。
- 设置每组人数。
- 设置是否允许自选小组。
- 点击创建房间。
- 展示 roomCode。
- 展示学生入口链接。
- 展示二维码。
- 点击进入房间管理页。

人为检查：

- 检查表单校验。
- 检查创建成功后的数据展示。
- 检查 joinUrl 是否正确。
- 检查二维码扫描后是否进入 student 页。

测试：

```text
创建房间 → 获得 roomCode → 打开学生链接
```

---

### Frontend Step 2：学生入场页

页面：

```text
/student?room=ABC123
```

功能：

- 读取 URL room 参数。
- 无 room 时允许手动输入房间号。
- 获取房间信息。
- 展示小组列表和剩余名额。
- 填写昵称。
- 选择小组。
- 加入成功后写入 localStorage。
- 如果 localStorage 有 token，尝试 resume。

人为检查：

- 检查重复昵称错误提示。
- 检查已满小组置灰。
- 检查 localStorage 字段是否完整。
- 检查刷新后是否能恢复身份。
- 检查 ended 房间提示。

测试：

```text
正常加入
重复昵称
小组满员
刷新恢复
无 roomCode 手动输入
```

---

### Frontend Step 3：学生课堂首页

页面：

```text
/student/classroom
```

功能：

- 展示课堂名称。
- 展示当前昵称和小组。
- 展示任务列表。
- 区分待完成、已提交、已评分。
- 展示我的得分。
- 展示小组排名。
- 提供进入任务详情按钮。

人为检查：

- 检查无任务状态。
- 检查已提交状态。
- 检查已评分状态。
- 检查移动端布局。
- 检查没有 token 时跳回入场页。

测试：

```text
学生加入后看到课堂首页
讲师发布任务后刷新可见
```

---

### Frontend Step 4：讲师房间管理页

页面：

```text
/teacher/rooms/:roomCode/dashboard
```

功能：

- 展示房间号、二维码、学生入口链接。
- 展示小组人数分布。
- 展示学生列表。
- 展示任务列表。
- 创建任务弹窗。
- 进入提交审阅页。
- 进入大屏。
- 进入后台数据看板。
- 结束课堂。
- 导出。

人为检查：

- 检查页面信息是否足够讲师操作。
- 检查任务创建表单。
- 检查学生列表是否按小组展示。
- 检查结束课堂是否有确认弹窗。

测试：

```text
创建任务
查看任务列表
查看学生加入变化
```

---

### Frontend Step 5：学生答题页

页面：

```text
/student/tasks/:taskId
```

功能：

- 展示任务标题、描述、截止时间。
- 文字输入。
- 图片选择。
- 图片预览。
- 图片数量限制 3 张。
- 图片大小限制 5MB。
- 草稿保存到 localStorage。
- 提交答案。
- 提交成功后回到课堂首页或显示成功状态。

人为检查：

- 检查图片限制前端是否存在。
- 检查后端错误是否展示。
- 检查提交后不能重复提交。
- 检查草稿 key 是否按 roomCode + taskId + studentId 区分。
- 检查截止后按钮 disabled。

测试：

```text
文字提交
图片提交
超过 3 张
超过 5MB
刷新草稿恢复
```

---

### Frontend Step 6：讲师审阅评分页

页面可以在 dashboard 内，也可以单独：

```text
/teacher/tasks/:taskId/submissions
```

功能：

- 按小组查看提交。
- 展示学生昵称。
- 展示提交文字。
- 展示图片缩略图。
- 点击查看大图。
- 输入分数。
- 输入评语。
- 保存评分。
- 设置精选答案。

人为检查：

- 检查分数只能 1–10。
- 检查保存后 UI 状态更新。
- 检查图片预览。
- 检查按小组筛选。
- 检查空提交状态。

测试：

```text
查看提交
打分
改分
评语
设置精选答案
```

---

### Frontend Step 7：WebSocket 接入

功能：

- 封装 WebSocket hook/helper。
- 学生端监听任务、评分、排名、结束课堂。
- 讲师端监听学生加入、提交、排名变化。
- 大屏端监听排名、完成度、精选答案。
- 断线后尝试重连。
- 重连后重新拉取当前页面数据。

人为检查：

- 检查重连不会创建多个连接。
- 检查组件卸载时关闭连接。
- 检查 token 失效时提示重新进入。
- 检查收到事件后不要盲目重复插入数据，必要时重新拉接口。

测试：

```text
多个浏览器窗口联调
发任务实时出现
提交实时更新
评分实时更新
结束课堂实时提示
```

---

### Frontend Step 8：大屏看板

页面：

```text
/teacher/rooms/:roomCode/display
```

功能：

- 全屏排行榜。
- 当前任务完成度。
- 倒计时。
- 精选答案展示。
- 支持匿名/显示小组。
- 适合投屏展示。

人为检查：

- 检查全屏布局。
- 检查排行榜排序。
- 检查没有精选答案时的空状态。
- 检查图片展示不变形。
- 检查字体大小适合投屏。

测试：

```text
评分后排行榜更新
精选答案后大屏展示
倒计时显示正常
```

---

### Frontend Step 9：后台数据看板

页面：

```text
/teacher/rooms/:roomCode/analytics
```

功能：

- 参与人数。
- 在线人数。
- 提交率。
- 各组得分对比。
- 任务完成情况。
- 提交时间分布。

人为检查：

- 检查图表/表格是否清晰。
- 检查无数据状态。
- 检查数据口径说明。
- 检查不要和大屏页面混淆。

测试：

```text
无学生
有学生无提交
有提交
有评分
```

---

### Frontend Step 10：结束课堂与导出

功能：

- 结束课堂按钮。
- 二次确认弹窗。
- 结束后页面状态变化。
- 导出按钮。
- 下载 Zip 文件。
- 学生端访问 ended 房间时提示课堂已结束。

人为检查：

- 检查结束课堂是危险操作，有二次确认。
- 检查导出失败提示。
- 检查下载文件名。
- 检查 ended 后学生端不可操作。

测试：

```text
结束课堂
学生端刷新
导出 Zip
检查 Zip 内容
```

---

## 6. 联调顺序

前后端联调必须按以下顺序：

```text
1. health check
2. 创建房间
3. 学生获取房间信息
4. 学生加入房间
5. 学生 resume
6. 讲师发布任务
7. 学生获取任务列表
8. 学生提交文字
9. 学生提交图片
10. 讲师查看提交
11. 讲师评分
12. 学生查看结果
13. 排行榜
14. WebSocket 实时同步
15. 大屏看板
16. 后台数据看板
17. 结束课堂
18. 导出 Zip
```

---

## 7. 每个阶段 Claude 的输出模板

每完成一个阶段，Claude 必须按这个格式回复：

```text
## 本阶段完成内容

### 修改文件
- ...

### 新增接口/页面
- ...

### 运行方式
```bash
...
```

### 自动测试

```bash
...
```

### 手动测试步骤

1. ...
2. ...
3. ...

### 人为检查点

- [ ] ...
- [ ] ...
- [ ] ...

### 下一步建议

下一步进入 XXX 阶段。

```
---

## 8. 人工 Code Review 总清单

每次 Claude 生成代码后，人类必须检查：

### 8.1 后端

- [ ] 是否符合 handler/service/repository 分层。
- [ ] handler 是否没有直接写复杂业务。
- [ ] service 是否有完整业务校验。
- [ ] repository 是否只做数据读写。
- [ ] 数据库操作是否需要事务。
- [ ] 错误返回是否清晰。
- [ ] 是否有重复提交、重复评分、越权访问风险。
- [ ] 文件上传是否限制大小和类型。
- [ ] 是否有路径穿越风险。
- [ ] 是否有 N+1 查询问题。
- [ ] 是否没有提交密钥。

### 8.2 前端

- [ ] API 是否集中封装。
- [ ] 页面是否没有硬编码后端地址。
- [ ] 错误提示是否清晰。
- [ ] loading 状态是否存在。
- [ ] 空状态是否存在。
- [ ] 学生端移动端是否可用。
- [ ] localStorage 是否按 room/task/student 维度隔离。
- [ ] WebSocket 是否正确关闭和重连。
- [ ] 图片上传前是否有限制和预览。
- [ ] 危险操作是否有二次确认。

### 8.3 业务

- [ ] 房间结束后学生不能加入。
- [ ] 房间结束后学生不能提交。
- [ ] 昵称房间内唯一。
- [ ] 小组满员不能加入。
- [ ] 截止后不能提交。
- [ ] 一个学生一个任务只能提交一次。
- [ ] 分数只能 1–10。
- [ ] 改分不会重复累计。
- [ ] 排行榜正确。
- [ ] 导出包含 Excel 和图片。

---

## 9. 测试数据建议

Claude 可以为本地测试准备以下数据：

```text
房间：Demo Class
组数：3
每组人数：2

学生：
Tom → 第1组
Jerry → 第1组
Alice → 第2组
Bob → 第2组
Cindy → 第3组

任务：
任务1：请提交你对 AI Classroom 的理解
任务2：请上传一张课堂作品图片
```

测试场景：

```text
Tom 提交文字 + 图片，得 8 分
Jerry 提交文字，得 6 分
Alice 提交文字 + 2 张图片，得 9 分
Bob 不提交
Cindy 提交后任务截止
```

预期：

```text
第1组总分 14
第2组总分 9
第3组根据 Cindy 得分计算
未提交学生不计分
排行榜按总分排序
```

---

## 10. 禁止事项

Claude 不得做以下事情：

- 不得直接跳到完整大而全版本。
- 不得一口气生成大量不可运行代码。
- 不得忽略数据库迁移。
- 不得跳过错误处理。
- 不得跳过手动测试说明。
- 不得把讲师账号体系加入 MVP，除非用户明确要求。
- 不得加入 AI 自动评分，PRD 明确无 AI 预评。
- 不得加入选择题，选择题为 P1。
- 不得把图片处理为外链输入，必须直接上传。
- 不得允许 0 分。
- 不得允许任务截止后提交。
- 不得允许课堂结束后学生继续访问。
- 不得把 `.env` 提交进 Git。
- 不得把生产域名、数据库密码、API key 写死在代码里。

---

## 11. 推荐开发命令

### 后端

```bash
cd backend
go mod tidy
go test ./...
go run ./cmd/server
```

### 前端

```bash
cd frontend
npm install
npm run dev
npm run build
```

### Git

```bash
git status
git add .
git commit -m "feat: add room creation flow"
```

---

## 12. 最小可演示版本定义

Claude 必须优先完成以下最小版本：

```text
1. 讲师创建房间
2. 学生加入房间
3. 讲师发布文字任务
4. 学生提交文字答案
5. 讲师查看提交并评分
6. 学生查看得分
7. 小组排行榜正确
```

只有这个版本稳定后，才能继续做：

```text
8. 图片上传
9. WebSocket
10. 大屏看板
11. 后台数据看板
12. 导出 Zip
```

---

## 13. 最终上线前检查

上线前必须确认：

- [ ] 前端生产 API 地址正确。
- [ ] 后端生产 CORS 正确。
- [ ] WebSocket 生产地址正确。
- [ ] HTTPS 可用。
- [ ] 图片上传目录或对象存储可写。
- [ ] 导出目录可写。
- [ ] MySQL 连接池配置合理。
- [ ] 数据库有备份策略。
- [ ] 日志不会输出敏感信息。
- [ ] 学生端在手机微信/浏览器中可打开。
- [ ] 二维码链接正确。
- [ ] 课堂结束后学生入口关闭。
- [ ] 导出 Excel + 图片 Zip 可下载。