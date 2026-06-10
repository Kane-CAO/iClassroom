# iClassroom_README.md

轻量级线上课堂互动平台。  
核心定位：**无需注册 · 扫码即入 · 一次性课堂 · 讲师端保留数据并可导出**。

本项目的 MVP 主线为：

```text
讲师建房 → 学生扫码/链接入场 → 讲师发布任务 → 学生提交文字/图片 → 讲师审阅评分 → 大屏/排名更新 → 课后归档导出
```

---

## 1. 项目背景

现有线上教学工具通常偏重，需要安装 App、注册登录、组织架构或账号体系支持。  
iClassroom 面向临时课堂、工作坊、培训、分组竞赛等轻量互动场景，提供一个纯网页、低门槛、可快速启动的课堂互动工具。

### 1.1 核心价值

- 学生无需注册账号，扫码或打开链接即可进入课堂。
- 讲师使用电脑端创建房间、发布任务、审阅提交、手动评分。
- 学生端支持手机、平板、电脑访问，手机体验优先。
- 课堂结束后关闭学生入口，学生端不再可进入。
- 讲师端保留课堂数据，支持导出 Excel 和图片原件 Zip。
- 大屏看板用于课堂投屏展示排行榜、完成度、精选答案。
- 后台数据看板用于讲师私下查看参与率、提交率、任务完成情况等。

---

## 2. 角色说明

| 角色 | 设备 | 主要能力 |
|---|---|---|
| 讲师 | 电脑网页端 | 创建房间、发任务、审阅提交、手动打分、查看数据、投屏展示、导出 |
| 学生 | 手机/平板/电脑网页端 | 扫码/链接入场、填写昵称、选择小组、提交文字和图片、查看得分和排名 |
| 大屏 | 讲师端内的全屏视图 | 展示排行榜、任务完成度、精选答案 |

---

## 3. MVP 功能范围

### 3.1 P0 必做功能

| 模块 | 功能 | 说明 |
|---|---|---|
| 房间 | 创建房间 | 讲师创建课堂，系统生成房间号、学生入口链接和二维码 |
| 入场 | 扫码/链接入场 | 学生无需注册，通过 roomCode 进入 |
| 入场 | 昵称房间内验重 | 同一个房间内昵称不能重复 |
| 入场 | 小组自选 | 学生选择小组；已满小组不可选 |
| 任务 | 发布任务 | 讲师发布任务，支持标题、描述、附件链接、截止时间 |
| 任务 | 指定范围 | 可推送全班或指定小组 |
| 提交 | 文字提交 | 学生提交文字答案 |
| 提交 | 图片上传 | 单图不超过 5MB，单题最多 3 张 |
| 提交 | 草稿保存 | 学生端本地保存草稿，避免误触丢失 |
| 控制 | 暂停/截止提交 | 任务暂停或截止后禁止提交 |
| 评分 | 手动打分 | 讲师按小组审阅，整数 1–10 分，最低 1 分 |
| 评分 | 评语 | 讲师可填写评语 |
| 排名 | 小组积分 | 分数自动累计到小组总分 |
| 大屏 | 排行榜 | 展示实时小组总榜 |
| 大屏 | 完成度 | 展示任务提交完成度和倒计时 |
| 大屏 | 精选答案 | 讲师选择优秀答案投屏，可匿名或显示小组 |
| 数据 | 后台数据看板 | 参与率、提交率、各组得分、任务完成情况、提交时间分布 |
| 会话 | 断线重连 | 学生端凭本地身份恢复会话 |
| 归档 | 结束课堂 | 关闭学生端入口 |
| 导出 | Excel + 图片 Zip | 讲师导出评分、提交、图片原件 |

### 3.2 非 MVP 功能

以下功能暂不进入第一版主线：

- 选择题题型。
- 抢答和额外加分项。
- 单轮排名。
- 课后自动报告。
- 词云、选项分布、复杂互动统计。
- 举手提问。
- 小组协作区。
- 课堂模板。
- 讲师账号体系和跨课堂历史趋势。

---

## 4. 推荐技术栈

> 最终技术栈以团队技术负责人决定为准。以下为 MVP 推荐实现方式。

### 4.1 前端

- 框架：React + TypeScript + Vite  
  或 Vue 3 + TypeScript + Vite
- 路由：React Router / Vue Router
- 请求：Axios / Fetch 封装
- 状态管理：Zustand / Pinia / Context
- 样式：Tailwind CSS / CSS Modules
- 图表：ECharts / Recharts / 简单表格优先
- 二维码：qrcode.react / qrcode.vue
- WebSocket：原生 WebSocket 封装

### 4.2 后端

- 语言：Go
- Web 框架：Gin
- ORM：GORM
- 数据库：MySQL
- 实时通信：WebSocket
- 文件上传：multipart/form-data
- 图片存储：
  - MVP 本地：`uploads/`
  - 后续生产：OSS / S3 / MinIO
- 导出：
  - Excel：excelize
  - Zip：archive/zip
- 配置：`.env`

### 4.3 部署

MVP 可以先使用：

- 前端：Vercel / Netlify / Nginx 静态部署
- 后端：云服务器 + Docker / Docker Compose
- 数据库：云 MySQL / 自建 MySQL
- 图片：本地挂载目录 / OSS / S3 / MinIO
- 域名：固定主域名，后续配置 HTTPS

---

## 5. 项目目录结构

推荐 Monorepo：

```text
iclassroom/
├── README.md
├── CLAUDE.md
├── docs/
│   ├── api.md
│   ├── database.md
│   ├── websocket-events.md
│   ├── deployment.md
│   └── test-checklist.md
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── domain/
│   │   ├── repository/
│   │   ├── service/
│   │   ├── handler/
│   │   ├── middleware/
│   │   ├── websocket/
│   │   ├── storage/
│   │   ├── export/
│   │   └── response/
│   ├── migrations/
│   ├── uploads/
│   ├── go.mod
│   ├── go.sum
│   └── .env.example
├── frontend/
│   ├── src/
│   │   ├── api/
│   │   ├── router/
│   │   ├── pages/
│   │   │   ├── teacher/
│   │   │   ├── student/
│   │   │   └── display/
│   │   ├── components/
│   │   ├── stores/
│   │   ├── hooks/
│   │   ├── utils/
│   │   └── types/
│   ├── public/
│   ├── package.json
│   └── .env.example
└── docker-compose.yml
```

---

## 6. 数据库模型

### 6.1 rooms

```text
id
room_code
title
group_count
group_capacity
allow_choose_group
status: created / active / ended
created_at
updated_at
ended_at
```

### 6.2 groups

```text
id
room_id
name
capacity
score_total
created_at
updated_at
```

### 6.3 students

```text
id
room_id
group_id
nickname
client_token
online_status
last_seen_at
created_at
updated_at
```

约束：

```text
UNIQUE(room_id, nickname)
```

### 6.4 tasks

```text
id
room_id
title
description
attachment_url
deadline_at
status: published / paused / closed
target_type: all / groups
created_at
updated_at
```

### 6.5 task_target_groups

```text
id
task_id
group_id
```

### 6.6 submissions

```text
id
room_id
task_id
student_id
group_id
content_text
status: submitted / graded
score
comment
submitted_at
graded_at
created_at
updated_at
```

约束：

```text
UNIQUE(task_id, student_id)
```

### 6.7 submission_images

```text
id
submission_id
file_url
file_path
file_name
file_size
mime_type
created_at
```

### 6.8 featured_answers

```text
id
room_id
task_id
submission_id
display_mode: anonymous / show_group
created_at
```

---

## 7. API 设计

### 7.1 讲师端 API

```http
POST /api/teacher/rooms
GET  /api/teacher/rooms/:roomCode
GET  /api/teacher/rooms/:roomCode/overview
GET  /api/teacher/rooms/:roomCode/students
GET  /api/teacher/rooms/:roomCode/groups

POST /api/teacher/rooms/:roomCode/tasks
GET  /api/teacher/rooms/:roomCode/tasks
PATCH /api/teacher/tasks/:taskId/pause
PATCH /api/teacher/tasks/:taskId/close

GET  /api/teacher/tasks/:taskId/submissions
POST /api/teacher/submissions/:submissionId/grade
POST /api/teacher/submissions/:submissionId/feature

GET  /api/teacher/rooms/:roomCode/display
GET  /api/teacher/rooms/:roomCode/analytics

POST /api/teacher/rooms/:roomCode/end
GET  /api/teacher/rooms/:roomCode/export
```

### 7.2 学生端 API

```http
GET  /api/student/rooms/:roomCode
POST /api/student/rooms/:roomCode/join
POST /api/student/rooms/:roomCode/resume

GET  /api/student/me/tasks
GET  /api/student/tasks/:taskId
POST /api/student/tasks/:taskId/submit

GET  /api/student/me/results
GET  /api/student/rooms/:roomCode/ranking
```

### 7.3 WebSocket

```http
GET /ws?room=ABC123&role=teacher
GET /ws?room=ABC123&role=student&token=xxx
GET /ws?room=ABC123&role=display
```

---

## 8. WebSocket 事件

### 8.1 student_joined

```json
{
  "type": "student_joined",
  "data": {
    "studentId": 1,
    "nickname": "Tom",
    "groupId": 1,
    "groupName": "第1组"
  }
}
```

### 8.2 task_published

```json
{
  "type": "task_published",
  "data": {
    "taskId": 1,
    "title": "课堂测试一",
    "deadlineAt": "2026-06-08T18:00:00"
  }
}
```

### 8.3 submission_created

```json
{
  "type": "submission_created",
  "data": {
    "taskId": 1,
    "studentId": 1,
    "groupId": 1,
    "submittedAt": "2026-06-08T17:30:00"
  }
}
```

### 8.4 ranking_updated

```json
{
  "type": "ranking_updated",
  "data": [
    {
      "groupId": 1,
      "groupName": "第1组",
      "scoreTotal": 35,
      "rank": 1
    }
  ]
}
```

### 8.5 room_ended

```json
{
  "type": "room_ended",
  "data": {
    "message": "课堂已结束"
  }
}
```

---

## 9. 本地开发环境

### 9.1 环境要求

```text
Go >= 1.22
Node.js >= 20
MySQL >= 8.0
Git
```

### 9.2 后端环境变量

复制：

```bash
cd backend
cp .env.example .env
```

示例：

```env
APP_ENV=development
APP_PORT=8080

DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=iclassroom

UPLOAD_DIR=./uploads
PUBLIC_BASE_URL=http://localhost:8080

CORS_ALLOWED_ORIGINS=http://localhost:5173
```

### 9.3 前端环境变量

复制：

```bash
cd frontend
cp .env.example .env
```

示例：

```env
VITE_API_BASE_URL=http://localhost:8080/api
VITE_WS_BASE_URL=ws://localhost:8080/ws
VITE_STUDENT_BASE_URL=http://localhost:5173/student
```

---

## 10. 本地运行步骤

> 该部分在项目实际代码完成后，需要根据最终技术栈更新。

### 10.1 启动数据库

```bash
mysql -u root -p
CREATE DATABASE iclassroom CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

或使用 Docker Compose：

```bash
docker compose up -d mysql
```

### 10.2 执行数据库迁移

待补充：

```bash
cd backend
# TODO: 根据最终迁移工具补充
```

### 10.3 启动后端

待补充：

```bash
cd backend
go mod tidy
go run ./cmd/server
```

健康检查：

```bash
curl http://localhost:8080/health
```

### 10.4 启动前端

待补充：

```bash
cd frontend
npm install
npm run dev
```

访问：

```text
http://localhost:5173
```

---

## 11. MVP 手动测试流程

### 11.1 房间与入场

- 讲师创建房间。
- 页面生成 roomCode 和学生入口链接。
- 学生打开 `/student?room=ABC123`。
- 学生填写昵称并选择小组。
- 重复昵称应提示更换。
- 已满小组应不可选择。
- 刷新页面后学生身份应恢复。

### 11.2 任务发布

- 讲师创建任务。
- 学生端能看到新任务。
- 指定小组任务只有目标小组学生可见。
- 任务截止时间显示正确。

### 11.3 答题提交

- 学生填写文字答案。
- 学生上传 1–3 张图片。
- 单图超过 5MB 应被拦截。
- 超过 3 张图片应被拦截。
- 提交成功后状态变为已提交。
- 重复提交应被拒绝。
- 任务暂停或截止后不可提交。
- 图片上传失败时文字草稿不丢失。

### 11.4 评分排名

- 讲师按小组查看提交。
- 讲师打 1–10 整数分。
- 0 分、负数、小数、超过 10 分应被拒绝。
- 保存评分后学生能看到分数与评语。
- 小组总分自动累计。
- 排行榜排序正确。

### 11.5 大屏与实时同步

- 任务发布后学生端自动出现任务。
- 学生提交后讲师端完成度变化。
- 评分后大屏排行榜更新。
- 精选答案能在大屏展示。
- 结束课堂后学生端自动提示课堂已结束。

### 11.6 归档导出

- 讲师结束课堂。
- 学生再次访问提示课堂已结束。
- 讲师导出 Zip。
- Zip 内含 Excel 和图片原件。
- Excel 字段完整，包含学生、小组、任务、提交、图片、分数、评语、时间。

---

## 12. 上线部署步骤

> 该部分预留给最终上线时补充。

### 12.1 生产环境变量

待补充：

```env
APP_ENV=production
APP_PORT=8080

DB_HOST=
DB_PORT=
DB_USER=
DB_PASSWORD=
DB_NAME=

UPLOAD_DIR=
PUBLIC_BASE_URL=https://api.iclassroom.app
CORS_ALLOWED_ORIGINS=https://iclassroom.app
```

### 12.2 后端部署

待补充：

```bash
# Docker build
# Docker run
# 或 Docker Compose
```

### 12.3 前端部署

待补充：

```bash
npm run build
# 部署 dist/
```

### 12.4 域名与 HTTPS

待补充：

```text
前端域名：
后端 API 域名：
HTTPS 证书：
反向代理：
CORS：
WebSocket 代理：
```

### 12.5 生产验收

- HTTPS 正常。
- 前端可以访问后端 API。
- WebSocket 可以连接。
- 图片可以上传和访问。
- 导出 Zip 可以下载。
- 房间结束后学生入口关闭。
- 数据库备份策略确认。
- 图片 30 天清理策略确认。

---

## 13. Git 协作规范

### 13.1 分支命名

```text
main
develop
feature/room-flow
feature/task-flow
feature/submission-upload
feature/grading-ranking
feature/websocket-display
feature/export
fix/xxx
```

### 13.2 Commit 规范

```text
feat: add room creation API
feat: add student join flow
feat: add task publish page
fix: validate duplicate nickname in room
refactor: split room service and repository
test: add room join service tests
docs: update API documentation
```

### 13.3 PR 检查项

- 代码能启动。
- 没有明显 lint 错误。
- 后端接口已用 curl/Postman 测过。
- 前端页面主流程能跑。
- 新增数据库字段有迁移说明。
- README 或 docs 已更新。
- 没有提交 `.env`、密码、token、密钥。

---

## 14. 当前开发状态

> 每完成一个模块后更新。

| 模块 | 状态 | 负责人 | 备注 |
|---|---|---|---|
| 项目初始化 | TODO |  |  |
| 房间创建 | TODO |  |  |
| 学生入场 | TODO |  |  |
| 任务发布 | TODO |  |  |
| 学生提交 | TODO |  |  |
| 图片上传 | TODO |  |  |
| 评分排名 | TODO |  |  |
| WebSocket | TODO |  |  |
| 大屏看板 | TODO |  |  |
| 后台数据看板 | TODO |  |  |
| 归档导出 | TODO |  |  |
| 部署上线 | TODO |  |  |
```

---

## 15. 重要边界

- MVP 不做学生账号注册。
- MVP 不做讲师账号体系，除非技术负责人另行确定。
- 学生身份使用本地缓存和后端临时 token。
- 同一房间内昵称必须唯一。
- 一个学生对一个任务只允许提交一次。
- 图片必须直接上传，不使用外链。
- 单图 ≤ 5MB，单题 ≤ 3 张。
- 评分为整数 1–10，最低 1 分，不打 0 分。
- 课堂结束后学生端不可再次进入。
- 讲师端保留数据，支持 30 天内导出。
