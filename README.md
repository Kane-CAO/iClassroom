# iClassroom

iClassroom V2.0 是面向展厅 / 体验中心运营场景的轻量网页互动平台。

核心定位：

```text
管理者管账号
讲师用电脑运营活动
观众扫码即入
活动结束归档导出
```

V2.0 主流程：

```text
管理者创建讲师账号
→ 讲师登录
→ 讲师创建活动房间
→ 观众扫码 / 链接入场
→ 观众填写昵称并选择小组
→ 讲师发布任务
→ 观众提交文字 / 图片 / 文件
→ 讲师审阅评分
→ 大屏和后台数据更新
→ 活动结束
→ 导出 Excel + 原件 Zip
```

---

## 1. 角色

| 角色 | 账号 | 主要设备 | 核心能力 |
| --- | --- | --- | --- |
| 管理者 | 有 | 电脑 | 管理讲师账号、重置密码、停用账号、查看 P1 运营概览 |
| 讲师 | 有 | 电脑 | 登录、建房、发任务、审阅评分、看板、导出 |
| 观众 | 无 | 手机 / 平板 / 电脑 | 扫码入场、答题、看得分与排名 |

---

## 2. V2.0 P0 范围

- 管理者：讲师账号管理。
- 讲师：账号登录。
- 讲师：创建活动房间，生成房间号、二维码、链接。
- 观众：扫码 / 链接 + 房间号入场，无需注册。
- 观众：昵称房间内验重，小组自选，小组满员不可选。
- 任务：发布任务，支持截止时间和目标小组。
- 提交：文字 + 图片 + 文件直接上传。
- 控制：暂停提交，截止后关闭提交。
- 评分：讲师按小组审阅，整数 1-10 分，自动累计小组总分。
- 大屏：排行榜、完成度、精选答案。
- 后台数据：参与率、得分对比、任务完成情况、提交时间分布。
- 会话：观众断线重连。
- 归档：活动结束后关闭观众入口。
- 导出：Excel + 图片 / 文件原件 Zip。

---

## 3. 暂缓范围

以下功能不进入 V2.0 第一阶段：

- 选择题题型。
- 抢答加分。
- 单轮榜。
- 词云 / 选项分布。
- 课后自动报告。
- 管理者运营概览的复杂分析。
- 举手提问。
- 小组协作区。
- 活动模板。

明确不做：

- AI 功能。
- 视频通话、直播、录制、视频上传。
- 投屏、屏幕共享、远程屏幕控制。

---

## 4. 工程化文档

团队继续开发前，先按以下文档对齐：

- `Claude.md`：AI / Claude / Codex 开发规则。
- `docs/V2_ENGINEERING_PLAN.md`：V2.0 工程实施计划。
- `docs/V2_PROMPT_PACK.md`：专业 Prompt Pack。
- `docs/V2_TASK_ASSIGNMENT.md`：前端、后端、设计分工。
- `docs/api.md`：V2.0 目标 API 契约。
- `docs/database.md`：V2.0 目标数据库模型。

---

## 5. Docker 本地一键运行

当前 Docker 配置用于内部演示和团队本地统一运行环境，会启动：

- `frontend`：React + Vite 前端，访问地址 `http://localhost:3000`
- `backend`：Go + Gin 后端，健康检查 `http://localhost:8080/health`
- `mysql`：MySQL 8.4

启动：

```bash
docker compose up --build
```

后台运行：

```bash
docker compose up --build -d
```

查看日志：

```bash
docker compose logs -f
```

停止：

```bash
docker compose down
```

清空本地演示数据并重新建表：

```bash
docker compose down -v
docker compose up --build
```

---

## 6. 本地数据库

Docker 内部后端连接 MySQL 服务名 `mysql:3306`。

宿主机连接：

```text
Host: 127.0.0.1
Port: 3307
Database: iclassroom
User: iclassroom
Password: iclassroom_password
```

---

## 7. 注意

当前代码仍包含 V1.x MVP 实现痕迹，例如房间级 `teacherToken`。V2.0 新开发应以账号化权限体系为目标，旧逻辑只作为迁移兼容参考。

真实二维码扫码、公网访问、HTTPS、云数据库、OSS / S3 / MinIO 文件存储，仍需要后续部署到云服务器并配置正式域名。
