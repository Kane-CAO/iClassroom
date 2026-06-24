# iClassroom V2.0 Team Task Assignment

本文档用于团队继续开发时分配前端、后端、设计任务。

---

## 1. 推荐团队结构

团队配置：

- 产品 1 人
- 设计 2 人
- 前端 2 人
- 后端 3-4 人
- 技术负责人 1 人，可由后端或全栈兼任

---

## 2. Prompt Pack 分配总览

团队成员不要完整运行 `docs/V2_PROMPT_PACK.md` 中的所有 Prompt。每个人只使用自己负责模块对应的 Prompt。

| 成员 | 负责板块 | 使用 Prompt | 说明 |
| --- | --- | --- | --- |
| 产品负责人 / 技术负责人 | 项目审计、范围冻结、最终验收 | Prompt 0、Prompt 16 | Prompt 0 只做差距分析；Prompt 16 做最终 E2E 和验收清单 |
| Backend A | 账号权限、数据库、房间归属 | Prompt 1、2、3、4 | 先做，给后续所有讲师接口打地基 |
| Backend B | 观众入场、活动房间、任务发布 | Prompt 5、6 | 等 Backend A 的登录鉴权稳定后接入 |
| Backend C | 提交、附件上传、审阅评分 | Prompt 7、8 | 重点负责 `submission_attachments` 和评分事务 |
| Backend D | 排行榜、大屏、后台数据、归档导出 | Prompt 9、10、11 | 等提交和评分数据稳定后开发 |
| Frontend A | 管理者端、讲师登录、讲师端账号化 | Prompt 12、13 | 与 Backend A 并行联调 |
| Frontend B | 观众端、上传、大屏、数据看板联调 | Prompt 14，配合 Prompt 9、10、11 | Prompt 9-11 由后端实现，Frontend B 负责页面接入 |
| Design A | 管理者端、讲师端、后台数据看板视觉 | Prompt 15 | 负责后台效率型界面视觉统一 |
| Design B | 观众端、大屏、移动端视觉 | Prompt 15 | 负责手机端和大屏端视觉统一 |

执行顺序建议：

```text
Prompt 0
→ Prompt 1, 2, 3, 4
→ Prompt 5, 6
→ Prompt 7, 8
→ Prompt 9, 10, 11
→ Prompt 12, 13, 14
→ Prompt 15
→ Prompt 16
```

---

## 3. 产品负责人

负责：

- 冻结 P0 / P1 / P2 范围。
- 维护验收清单。
- 确认文件类型、文件大小、账号初始化方式。
- 组织每阶段评审。
- 避免 P1/P2 功能提前进入 P0。

输出：

- 页面清单。
- 验收用例。
- 每周优先级。

对应 Prompt：

- Prompt 0：V2.0 项目审计。
- Prompt 16：E2E 测试与验收。

---

## 4. 设计分工

### Design A：管理者端 + 讲师端

负责：

- 管理者登录。
- 讲师账号管理。
- 讲师登录。
- 讲师工作台。
- 创建活动。
- 创建任务。
- 审阅评分。
- 后台数据看板。
- 导出和归档状态。

设计原则：

- 电脑端优先。
- 信息密度清晰。
- 更像运营后台，不像营销落地页。

对应 Prompt：

- Prompt 15：AITIC 视觉体系改造。
- 需要参考 Prompt 12、Prompt 13 的页面范围，但不负责实现代码。

### Design B：观众端 + 大屏端

负责：

- 观众扫码入场。
- 手机答题页。
- 图片 / 文件上传。
- 我的得分。
- 小组排名。
- 大屏排行榜。
- 精选答案展示。
- AITIC 视觉规范。

设计原则：

- 手机优先。
- 现场使用路径短。
- 大屏适合远距离观看。
- 阿里橙 + 深墨 + 暖中性。

对应 Prompt：

- Prompt 15：AITIC 视觉体系改造。
- 需要参考 Prompt 14 的观众端范围，以及 Prompt 9、Prompt 10 的大屏 / 数据看板范围，但不负责实现代码。

---

## 5. 前端分工

### Frontend A：管理者端 + 讲师账号化

负责：

- `/admin/login`
- `/admin/teachers`
- `/teacher/login`
- 管理者 / 讲师 session。
- 路由保护。
- API client 鉴权改造。
- 讲师工作台只展示自己的活动。

主要文件：

```text
frontend/src/api/
frontend/src/router/
frontend/src/pages/admin/
frontend/src/pages/teacher/
frontend/src/hooks/
frontend/src/utils/session.ts
frontend/src/types/
```

对应 Prompt：

- Prompt 12：前端管理者端。
- Prompt 13：前端讲师端账号化改造。

不要使用：

- Prompt 14：观众端上传，除非需要做全局 API / session 配合。
- Prompt 15：视觉统一由设计给规范后再执行。

### Frontend B：观众端 + 上传 + 看板

负责：

- 观众入场。
- 小组选择。
- 任务列表。
- 答题页。
- 文字草稿。
- 图片 / 文件上传。
- 我的得分。
- 大屏看板。
- 后台数据看板。
- 导出和结束活动入口。

主要文件：

```text
frontend/src/pages/student/
frontend/src/pages/display/
frontend/src/pages/teacher/Review.tsx
frontend/src/pages/teacher/Analytics.tsx
frontend/src/components/
frontend/src/api/tasks.ts
frontend/src/api/analytics.ts
frontend/src/api/export.ts
```

对应 Prompt：

- Prompt 14：前端观众端上传与移动端体验。

需要联调但不主导的 Prompt：

- Prompt 9：排行榜、大屏、精选答案。
- Prompt 10：后台数据看板。
- Prompt 11：活动结束与导出。

这些 Prompt 的后端逻辑由 Backend D 负责，Frontend B 负责页面接入和交互联调。

---

## 6. 后端分工

### Backend A：账号权限

负责：

- 账号表。
- 登录接口。
- 鉴权中间件。
- 管理者讲师账号管理。
- 房间归属讲师。

对应 Prompt：

- Prompt 1：数据库模型改造。
- Prompt 2：账号与权限基础架构。
- Prompt 3：管理者讲师账号管理。
- Prompt 4：讲师登录与房间归属迁移。

不要使用：

- Prompt 7-11，避免账号模块和业务展示模块混在一个 PR。

### Backend B：活动和任务

负责：

- 房间创建。
- 观众入场。
- 小组容量。
- 任务发布。
- 暂停 / 关闭 / 截止。

对应 Prompt：

- Prompt 5：观众入场与匿名会话。
- Prompt 6：任务发布与截止控制。

前置依赖：

- Backend A 完成讲师登录鉴权和房间归属。

### Backend C：提交和评分

负责：

- 统一附件模型。
- 图片上传。
- 文件上传。
- 提交。
- 审阅。
- 评分。
- 小组分数更新。

对应 Prompt：

- Prompt 7：文字 / 图片 / 文件提交。
- Prompt 8：讲师审阅与评分。

前置依赖：

- Backend B 完成任务发布、目标范围、截止 / 暂停 / 关闭校验。

### Backend D：看板和导出

负责：

- 排行榜。
- 大屏。
- 后台数据。
- 精选答案。
- 活动结束。
- Excel + Zip 导出。

对应 Prompt：

- Prompt 9：排行榜、大屏、精选答案。
- Prompt 10：后台数据看板。
- Prompt 11：活动结束与导出。

前置依赖：

- Backend C 完成提交、附件、评分、小组分数更新。

---

## 7. 推荐排期

| 周期 | 重点 | 主要负责人 |
| --- | --- | --- |
|  1  | 账号权限、登录、管理者端基础 | Backend A, Frontend A, Design A |
|  2  | 房间、观众入场、任务发布 | Backend B, Frontend A/B |
|  3  | 上传、提交、评分 | Backend C, Frontend B, Design B |
|  4  | 看板、导出、视觉走查、联调 | Backend D, Frontend B, Design A/B |

---

## 8. 每日协作规则

每天同步：

- 昨天完成什么。
- 今天做什么。
- 是否阻塞。
- 是否影响接口 / 数据库 / 设计。

每个 PR 合并前：

- 对齐 `docs/api.md`。
- 对齐 `docs/database.md`。
- 跑测试。
- 给出手动验收步骤。

---

## 9. 最小验收链路

任何阶段不能破坏这条链路：

```text
管理者创建讲师
→ 讲师登录
→ 讲师创建房间
→ 观众加入
→ 讲师发布任务
→ 观众提交
→ 讲师评分
→ 查看排名
→ 结束活动
→ 导出
```
