# iClassroom 后端开发任务分工与 Prompt 使用说明

> 适用对象：iClassroom 后端 4 位成员  
> 目的：明确每个人负责什么、应该使用 Prompt Pack 中哪几个 Prompt、哪些文件可以改、怎么测试、怎么提 PR。  
> 当前标准文档：`README.md / README01.md`、`CLAUDE.md / Claude.md`、`docs/api.md`、`iClassroom Claude Prompt Pack`。

---

## 0.Prompt Pack 怎么用

大家手里已经有一份 Prompt Pack，但**不是每个人都从 Prompt 0 跑到 Prompt 16**。

正确方式是：

```text
每个人只使用自己负责模块对应的 Prompt。
不要一次性把所有 Prompt 都发给 Claude / Codex / Cursor。
不要让 AI 生成完整后端。
不要改自己模块以外的核心文件。
```

Prompt Pack 的作用是：

```text
README：告诉大家项目做什么
CLAUDE：告诉大家代码怎么写
docs/api.md：告诉大家接口怎么定
Prompt Pack：告诉 AI 每一步具体做什么
```

---

## 1. 当前统一结论

MVP 主流程：

```text
老师创建房间
→ 生成 roomCode + teacherToken
→ 学生扫码/链接进入
→ 学生填写 nickname + 选择 group
→ 后端生成 clientToken
→ 老师发布 task
→ 学生提交 submission
→ 老师评分 grade
→ 小组 ranking 更新
→ 大屏/后台看板展示
→ 老师结束课堂
→ 导出 Excel + 图片 Zip
```

身份设计：

```text
学生端：不登录，使用 nickname + group + clientToken
老师端：不做账号登录，使用 roomCode + teacherToken 管理房间
roomCode：只用于识别房间和学生入场，不能作为老师端权限
```

统一接口标准：

```text
接口以 docs/api.md 为准
字段统一 camelCase
老师端接口使用 X-Teacher-Token
学生端接口使用 X-Student-Token
返回格式统一 success / message / data / errorCode
```

---

## 2. Prompt Pack 使用分配总览

| 成员 | 模块 | 主要使用 Prompt | 暂时不要做的 Prompt |
|---|---|---|---|
| A：后端集成 Owner | 基础架构、数据库、房间、学生入场、总集成 | Prompt 5、Prompt 6、Prompt 7 | Prompt 9 以后先不要做 |
| B：核心业务 Owner | 任务发布、文字提交、评分、排行榜 | Prompt 9、Prompt 10、Prompt 11 | Prompt 12 图片上传先不要做 |
| C：数据看板 Owner | 大屏看板、后台数据看板、测试文档 | Prompt 14、Prompt 16 中 test-checklist 部分 | 不要改任务/提交/评分核心逻辑 |
| D：文件导出 Owner | 图片上传、图片元数据、Excel + Zip 导出 | Prompt 12、Prompt 15 | 不要重写 B 的文字提交逻辑 |
| 全员 | 项目理解和规则 | Prompt 0 可作为阅读审计参考；Prompt 4 已完成，参考 docs/api.md | 不要重复生成 docs/api.md |

---

## 3. 当前开发顺序

现在已经完成/正在完成：

```text
Prompt 0：项目审计
Prompt 1：整理目录
Prompt 2：前端工程壳子
Prompt 3：静态 HTML 迁移
Prompt 4：docs/api.md API Contract
```

接下来后端按这个顺序：

```text
1. A 先做 Prompt 5：后端基础服务
2. A 再做 Prompt 6：数据库迁移和实体
3. A 做 Prompt 7：房间创建和学生入场
4. B 基于 A 的地基做 Prompt 9、10、11
5. D 基于 B 的提交逻辑做 Prompt 12、15
6. C 基于已有数据做 Prompt 14 和测试文档
7. 最后再考虑 WebSocket，也就是 Prompt 13
```

注意：

```text
Prompt 13 WebSocket 不要太早做。
Prompt 12 图片上传要等文字提交先稳定。
Prompt 15 导出要等提交、图片、评分数据都有后再做。
```

---

## 4. A：后端集成 Owner

### 4.1 A 使用哪些 Prompt

A 主要使用：

```text
Prompt 5：后端 Step 0，基础服务
Prompt 6：后端 Step 1，数据库迁移和实体
Prompt 7：后端 Step 2 + Step 3，创建房间和学生入场
```

A 暂时不要做：

```text
Prompt 9：任务发布
Prompt 10：文字提交
Prompt 11：评分排行榜
Prompt 12：图片上传
Prompt 13：WebSocket
Prompt 14：大屏/后台看板
Prompt 15：导出
```

### 4.2 A 的负责范围

```text
后端基础结构
数据库 migration
domain model
创建房间
自动创建小组
学生获取房间信息
学生加入房间
学生 resume 恢复会话
整体合并 review
```

### 4.3 A 负责接口

```http
GET  /health

POST /api/teacher/rooms
GET  /api/teacher/rooms/:roomCode
GET  /api/teacher/rooms/:roomCode/overview

GET  /api/student/rooms/:roomCode
POST /api/student/rooms/:roomCode/join
POST /api/student/rooms/:roomCode/resume
```

### 4.4 A 主要文件范围

```text
backend/cmd/server/main.go

backend/internal/config/
backend/internal/response/
backend/internal/middleware/
backend/internal/router/

backend/internal/domain/
backend/migrations/

backend/internal/repository/room_repository.go
backend/internal/repository/group_repository.go
backend/internal/repository/student_repository.go

backend/internal/service/room_service.go
backend/internal/service/student_service.go

backend/internal/handler/room_handler.go
backend/internal/handler/student_handler.go

docs/database.md
docs/api.md 中 room / student 部分
```

### 4.5 A 给 AI 的使用方式

A 不要一次性把 Prompt 5、6、7 全部丢给 AI。

推荐顺序：

```text
先发 Prompt 5
检查能否 go run、/health、go test ./...
通过后 commit

再发 Prompt 6
检查 migration、domain、docs/database.md
通过后 commit

最后发 Prompt 7
检查房间创建、学生加入、resume
通过后 commit
```

### 4.6 A 的额外补充要求

给 AI 执行 Prompt 5/6/7 时，额外加上：

```text
必须以 docs/api.md 为接口标准。
teacherToken / clientToken 的命名和 Header 必须和 docs/api.md 一致。
不要改变统一响应格式。
如果发现 README / CLAUDE / docs/api.md 有冲突，先指出冲突，不要自行决定。
```

---

## 5. B：核心业务 Owner

### 5.1 B 使用哪些 Prompt

B 主要使用：

```text
Prompt 9：任务发布和任务列表
Prompt 10：文字提交和讲师审阅
Prompt 11：评分和排行榜
```

B 暂时不要做：

```text
Prompt 12：图片上传
Prompt 13：WebSocket
Prompt 14：大屏/后台数据看板
Prompt 15：导出
```

### 5.2 B 的负责范围

```text
任务发布
任务列表
暂停任务
关闭任务
学生文字提交
老师查看提交
老师评分
学生查看结果
小组排行榜
```

### 5.3 B 负责接口

```http
POST  /api/teacher/rooms/:roomCode/tasks
GET   /api/teacher/rooms/:roomCode/tasks
PATCH /api/teacher/tasks/:taskId/pause
PATCH /api/teacher/tasks/:taskId/close

GET   /api/student/me/tasks
POST  /api/student/tasks/:taskId/submit

GET   /api/teacher/tasks/:taskId/submissions
POST  /api/teacher/submissions/:submissionId/grade

GET   /api/student/me/results
GET   /api/student/rooms/:roomCode/ranking
```

### 5.4 B 主要文件范围

```text
backend/internal/repository/task_repository.go
backend/internal/repository/submission_repository.go
backend/internal/repository/ranking_repository.go

backend/internal/service/task_service.go
backend/internal/service/submission_service.go
backend/internal/service/grading_service.go
backend/internal/service/ranking_service.go

backend/internal/handler/task_handler.go
backend/internal/handler/submission_handler.go
backend/internal/handler/grading_handler.go

docs/api.md 中 task / submission / grade / ranking 部分
docs/test-checklist.md 中对应测试部分
```

### 5.5 B 给 AI 的使用方式

B 要等 A 的 Prompt 5/6/7 完成并合并后再开始。

推荐顺序：

```text
先发 Prompt 9
完成任务发布、任务列表、暂停、关闭
测试通过后 commit

再发 Prompt 10
完成文字提交和老师查看提交
测试通过后 commit

最后发 Prompt 11
完成评分、学生结果、排行榜
测试通过后 commit
```

### 5.6 B 的额外补充要求

给 AI 执行 Prompt 9/10/11 时，额外加上：

```text
必须基于现有 A 已完成的 room/student/domain/migration 结构继续开发。
不要重建项目结构。
不要重写公共 response、config、router 总入口。
不要实现图片上传，Prompt 10 只做文字提交。
所有接口字段和错误码以 docs/api.md 为准。
老师接口必须校验 X-Teacher-Token。
学生接口必须校验 X-Student-Token。
评分时必须用事务，改分必须按差值更新 group score，不能重复累计。
```

---

## 6. C：数据看板 Owner

### 6.1 C 使用哪些 Prompt

C 主要使用：

```text
Prompt 14：大屏看板和后台数据看板
Prompt 16：最终文档和部署占位更新中的 test-checklist 部分
```

C 暂时不要做：

```text
Prompt 9：任务发布
Prompt 10：提交
Prompt 11：评分
Prompt 12：图片上传
Prompt 15：导出
```

### 6.2 C 的负责范围

```text
大屏看板数据
后台数据看板 analytics
精选答案基础接口
docs/test-checklist.md
```

### 6.3 C 负责接口

```http
GET  /api/teacher/rooms/:roomCode/display
POST /api/teacher/submissions/:submissionId/feature
GET  /api/teacher/rooms/:roomCode/analytics
```

如果后续需要取消精选答案，可再补：

```http
DELETE /api/teacher/featured-answers/:featuredId
```

### 6.4 C 主要文件范围

```text
backend/internal/repository/display_repository.go
backend/internal/repository/analytics_repository.go
backend/internal/repository/featured_answer_repository.go

backend/internal/service/display_service.go
backend/internal/service/analytics_service.go
backend/internal/service/featured_answer_service.go

backend/internal/handler/display_handler.go
backend/internal/handler/analytics_handler.go
backend/internal/handler/featured_answer_handler.go

docs/api.md 中 display / analytics / feature 部分
docs/test-checklist.md
```

### 6.5 C 给 AI 的使用方式

C 建议等 B 完成任务、提交、评分之后再做 Prompt 14，因为 display 和 analytics 需要已有数据。

推荐顺序：

```text
先阅读 docs/api.md 中 display / analytics / feature 部分
等 A、B 的基础数据链路完成后，执行 Prompt 14
完成只读接口后，补充 docs/test-checklist.md
```

### 6.6 C 的额外补充要求

给 AI 执行 Prompt 14 时，额外加上：

```text
这些接口主要是 read-only 查询。
不要修改 room/task/submission/grade 的核心写入逻辑。
不要重写评分和排行榜逻辑。
老师接口必须校验 X-Teacher-Token。
无数据时返回 0 或空数组，不能报错。
completionRate / submissionRate 除数为 0 时返回 0。
anonymous 模式不要返回学生昵称。
```

---

## 7. D：文件导出 Owner

### 7.1 D 使用哪些 Prompt

D 主要使用：

```text
Prompt 12：图片上传
Prompt 15：结束课堂和导出
```

D 暂时不要做：

```text
Prompt 9：任务发布
Prompt 10：文字提交的核心逻辑
Prompt 11：评分排行榜
Prompt 13：WebSocket
Prompt 14：大屏/后台看板
```

### 7.2 D 的负责范围

```text
图片上传
图片文件保存
图片元数据入库
老师查看提交时返回图片 URL
Excel + 图片 Zip 导出
```

### 7.3 D 负责接口

```http
POST /api/student/tasks/:taskId/submit
GET  /api/teacher/rooms/:roomCode/export
```

注意：

```text
POST /api/student/tasks/:taskId/submit 原本由 B 先实现文字提交。
D 的任务是在 B 的基础上增加 images[] 支持，不要重写整个提交逻辑。
```

### 7.4 D 主要文件范围

```text
backend/internal/storage/
backend/internal/export/

backend/internal/repository/submission_image_repository.go

backend/internal/service/upload_service.go
backend/internal/service/export_service.go

backend/internal/handler/export_handler.go

docs/api.md 中 submit 图片上传 / export 部分
docs/test-checklist.md 中图片和导出测试部分
```

### 7.5 D 给 AI 的使用方式

D 必须等 B 的文字提交逻辑完成后再做 Prompt 12。

推荐顺序：

```text
等 B 的 Prompt 10 完成文字提交
再发 Prompt 12，给 submit 增加图片上传
图片上传稳定后，再做 Prompt 15 中 export 部分
结束课堂接口如果已由 A/B 处理，D 只负责 export
```

### 7.6 D 的额外补充要求

给 AI 执行 Prompt 12/15 时，额外加上：

```text
不要重写已有文字提交逻辑，只在其基础上增加 images[] 支持。
不要改变 submit 接口的字段和响应格式。
图片上传必须后端校验数量、大小、MIME。
文件名必须由后端重新生成。
必须防止目录穿越。
导出 Zip 必须包含 submissions.xlsx 和 images/。
没有图片时也能导出 Excel。
老师导出接口必须校验 X-Teacher-Token。
```

---

## 8. Prompt 13：WebSocket 谁来做

Prompt 13 暂时不单独做。

```text
等 HTTP 核心闭环稳定后，由 A + B 一起做 WebSocket。
C 可以协助测试大屏实时更新。
D 不需要参与 WebSocket 核心实现。
```

做 WebSocket 前必须已经跑通：

```text
创建房间
学生入场
发布任务
学生提交
老师评分
排行榜
```

否则 WebSocket 做早了只会增加联调复杂度。

---

## 9. 共享文件规则

### 9.1 强共享文件：不能随便改

以下文件只有 A / 技术负责人 review 后才能改：

```text
README.md
README01.md
CLAUDE.md
Claude.md
docs/api.md
docs/database.md

docker-compose.yml
backend/go.mod
backend/.env.example

backend/cmd/server/main.go
backend/internal/config/
backend/internal/response/
backend/internal/middleware/
backend/internal/router/
backend/internal/domain/
backend/migrations/
```

原因：

```text
这些是项目规则、接口合同、数据库结构、公共响应格式、启动入口和环境配置。
```

### 9.2 模块文件：各自负责

```text
B：task / submission / grading / ranking
C：display / analytics / featured answer
D：storage / upload / export
```

如果需要改别人模块，必须先在群里说明原因。

---

## 10. Git 分支规范

每个人从最新主分支拉自己的 feature 分支：

```bash
git checkout main
git pull origin main
git checkout -b feature/backend-room-student
```

建议分支名：

```text
A：feature/backend-room-student
B：feature/backend-task-submission-grading
C：feature/backend-display-analytics
D：feature/backend-upload-export
```

不要直接 push 到 main。

每个 PR 尽量小，不要一次提交几十个无关文件。

---

## 11. 每个 PR 必须包含

每个人提 PR 时必须写清楚：

```text
1. 本次实现了什么
2. 修改了哪些文件
3. 实现了哪些接口
4. 如何运行
5. curl 测试命令
6. go test ./... 是否通过
7. 是否更新 docs/api.md / docs/database.md / docs/test-checklist.md
8. 人工检查点
```

PR 模板：

```md
## 本次实现

- 

## 修改文件

- 

## 接口

- 

## 运行方式

```bash
cd backend
go test ./...
go run ./cmd/server
```

## curl 测试

```bash
curl ...
```

## 文档更新

- [ ] docs/api.md
- [ ] docs/database.md
- [ ] docs/test-checklist.md

## 人工检查点

- [ ] 遵守 handler/service/repository 分层
- [ ] 没有改无关文件
- [ ] 没有提交 .env / 密钥
- [ ] 错误处理清楚
- [ ] token 校验正确
- [ ] go test ./... 通过
```

---

## 12. Code Review 检查清单

Review 时重点检查：

```text
是否符合 handler/service/repository 分层
handler 是否没有写复杂业务
service 是否有完整业务校验
repository 是否只做数据库读写
是否需要事务
错误返回是否统一
是否使用 docs/api.md 的字段
是否校验 X-Teacher-Token / X-Student-Token
是否有重复提交、重复评分、越权访问风险
是否没有提交 .env 或密钥
```

关键业务检查：

```text
roomCode 不能作为老师权限
老师接口必须校验 teacherToken
学生接口必须校验 clientToken
房间 ended 后学生不能 join/resume/submit
nickname 同房间唯一
小组满员不能加入
任务截止后不能提交
一个学生一个任务只能提交一次
评分只能 1–10
改分不能重复累计
图片最多 3 张
单图最大 5MB
导出包含 Excel + 图片 Zip
```

---

## 13. 当前节奏

### 第 1 步：A 先完成地基

```text
Prompt 5
Prompt 6
Prompt 7
```

其他成员此时可以先：

```text
B：阅读 Prompt 9 / 10 / 11，准备业务逻辑
C：阅读 Prompt 14，准备 display / analytics 查询口径
D：阅读 Prompt 12 / 15，准备文件处理方案
```

### 第 2 步：B 开始核心业务

```text
Prompt 9
Prompt 10
Prompt 11
```

### 第 3 步：C 和 D 并行

```text
C：Prompt 14 + docs/test-checklist.md
D：Prompt 12 + Prompt 15
```

### 第 4 步：A + B 再考虑 WebSocket

```text
Prompt 13
```

---

## 14. 当前最重要的提醒

现在不要追求一次性做完整系统。

优先目标：

```text
先把 HTTP 核心闭环跑通：
创建房间 → 学生入场 → 发布任务 → 学生提交 → 老师评分 → 排行榜
```

后面再补：

```text
图片上传
WebSocket
大屏
后台数据看板
导出
```

所有人不管用 Claude、Codex、Cursor 还是自己写，都必须遵守：

```text
README
CLAUDE
docs/api.md
本文档
```
