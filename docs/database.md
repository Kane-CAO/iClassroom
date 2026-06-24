# iClassroom V2.0 Database Design

> 本文档描述 V2.0 目标数据库模型。当前迁移文件仍可能是 V1.x MVP 结构，后续应通过新增 migration 演进，不要直接覆盖生产历史迁移。

---

## 1. 实体关系

```text
admin_users (N)
teacher_accounts (N)
teacher_sessions (N)

teacher_accounts (1) ──< rooms (N)
rooms (1) ──< groups (N)
rooms (1) ──< students (N) >── groups (1)
rooms (1) ──< tasks (N)
tasks (1) ──< task_target_groups (N) >── groups (1)
tasks (1) ──< submissions (N) >── students (1)
submissions (1) ──< submission_attachments (N)
submissions (1) ──< featured_answers (0..1)
```

---

## 2. 账号表

### 2.1 `admin_users`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | BIGINT UNSIGNED PK | 管理者 ID |
| `username` | VARCHAR(64) UNIQUE | 登录名 |
| `password_hash` | VARCHAR(255) | 密码哈希 |
| `display_name` | VARCHAR(64) | 显示名 |
| `status` | ENUM(`active`,`disabled`) | 状态 |
| `created_at` / `updated_at` | DATETIME | 时间 |

### 2.2 `teacher_accounts`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | BIGINT UNSIGNED PK | 讲师 ID |
| `username` | VARCHAR(64) UNIQUE | 登录名 |
| `password_hash` | VARCHAR(255) | 密码哈希 |
| `display_name` | VARCHAR(64) | 显示名 |
| `status` | ENUM(`active`,`disabled`) | 停用后不可登录 |
| `created_by_admin_id` | BIGINT UNSIGNED FK | 创建者 |
| `last_login_at` | DATETIME NULL | 最近登录 |
| `created_at` / `updated_at` | DATETIME | 时间 |

### 2.3 `auth_sessions`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | BIGINT UNSIGNED PK | 会话 ID |
| `user_type` | ENUM(`admin`,`teacher`) | 用户类型 |
| `user_id` | BIGINT UNSIGNED | 用户 ID |
| `token_hash` | VARCHAR(255) UNIQUE | token 哈希 |
| `expires_at` | DATETIME | 过期时间 |
| `revoked_at` | DATETIME NULL | 登出 / 撤销时间 |
| `created_at` | DATETIME | 创建时间 |

> 如果技术组选择 JWT，可不落 `auth_sessions`，但仍要在实现文档中说明失效、停用账号和密码重置后的处理方式。

---

## 3. 活动核心表

### 3.1 `rooms`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | BIGINT UNSIGNED PK | 房间 ID |
| `teacher_id` | BIGINT UNSIGNED FK→teacher_accounts | 创建讲师 |
| `room_code` | VARCHAR(16) UNIQUE | 房间号 |
| `title` | VARCHAR(255) | 活动名称 |
| `status` | ENUM(`created`,`active`,`ended`) | 状态 |
| `group_count` | INT | 小组数 |
| `group_capacity` | INT | 每组容量 |
| `allow_choose_group` | TINYINT(1) | 是否允许自选小组 |
| `legacy_teacher_token` | VARCHAR(64) NULL | 旧版本迁移兼容，不作为 V2 主鉴权 |
| `created_at` / `updated_at` | DATETIME | 时间 |
| `ended_at` | DATETIME NULL | 结束时间 |

关键索引：

- `uk_rooms_room_code (room_code)`
- `idx_rooms_teacher_id (teacher_id)`
- `idx_rooms_status (status)`

### 3.2 `groups`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | BIGINT UNSIGNED PK | 小组 ID |
| `room_id` | BIGINT UNSIGNED FK→rooms | 房间 |
| `group_name` | VARCHAR(64) | 组名 |
| `capacity` | INT | 容量 |
| `score_total` | INT | 小组总分 |
| `created_at` / `updated_at` | DATETIME | 时间 |

### 3.3 `students`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | BIGINT UNSIGNED PK | 观众 ID |
| `room_id` | BIGINT UNSIGNED FK→rooms | 房间 |
| `group_id` | BIGINT UNSIGNED FK→groups | 小组 |
| `nickname` | VARCHAR(64) | 昵称 |
| `client_token` | VARCHAR(64) UNIQUE | 观众会话凭证 |
| `online_status` | ENUM(`online`,`offline`) | 在线状态 |
| `last_seen_at` | DATETIME NULL | 最近在线 |
| `created_at` / `updated_at` | DATETIME | 时间 |

约束：

- `UNIQUE(room_id, nickname)`

---

## 4. 任务与提交

### 4.1 `tasks`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | BIGINT UNSIGNED PK | 任务 ID |
| `room_id` | BIGINT UNSIGNED FK→rooms | 房间 |
| `title` | VARCHAR(255) | 标题 |
| `description` | TEXT NULL | 描述 |
| `deadline_at` | DATETIME | 截止时间 |
| `target_type` | ENUM(`all`,`groups`) | 发布范围 |
| `status` | ENUM(`published`,`paused`,`closed`) | 状态 |
| `created_at` / `updated_at` | DATETIME | 时间 |

### 4.2 `task_target_groups`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | BIGINT UNSIGNED PK | 主键 |
| `task_id` | BIGINT UNSIGNED FK→tasks | 任务 |
| `group_id` | BIGINT UNSIGNED FK→groups | 目标小组 |

约束：

- `UNIQUE(task_id, group_id)`

### 4.3 `submissions`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | BIGINT UNSIGNED PK | 提交 ID |
| `room_id` | BIGINT UNSIGNED FK→rooms | 冗余房间 |
| `task_id` | BIGINT UNSIGNED FK→tasks | 任务 |
| `student_id` | BIGINT UNSIGNED FK→students | 观众 |
| `group_id` | BIGINT UNSIGNED FK→groups | 提交时所在小组 |
| `content_text` | TEXT NULL | 文字答案 |
| `status` | ENUM(`submitted`,`graded`) | 状态 |
| `score` | INT NULL | 1-10 分 |
| `comment` | VARCHAR(1024) NULL | 评语 |
| `submitted_at` | DATETIME | 提交时间 |
| `graded_at` | DATETIME NULL | 评分时间 |
| `created_at` / `updated_at` | DATETIME | 时间 |

约束：

- `UNIQUE(task_id, student_id)`
- `CHECK(score IS NULL OR score BETWEEN 1 AND 10)`

### 4.4 `submission_attachments`

统一承载图片和文件，替代只支持图片的 `submission_images` 单一模型。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | BIGINT UNSIGNED PK | 附件 ID |
| `submission_id` | BIGINT UNSIGNED FK→submissions | 提交 |
| `kind` | ENUM(`image`,`file`) | 图片或文件 |
| `file_url` | VARCHAR(1024) | 可访问 URL |
| `file_path` | VARCHAR(1024) | 服务端路径 |
| `original_file_name` | VARCHAR(255) | 原始文件名，仅展示用 |
| `stored_file_name` | VARCHAR(255) | 后端存储文件名 |
| `file_size` | BIGINT UNSIGNED | 字节 |
| `mime_type` | VARCHAR(128) | MIME |
| `created_at` | DATETIME | 时间 |

索引：

- `idx_submission_attachments_submission_id`
- `idx_submission_attachments_kind`

---

## 5. 看板与导出

### 5.1 `featured_answers`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | BIGINT UNSIGNED PK | 精选 ID |
| `room_id` | BIGINT UNSIGNED FK→rooms | 房间 |
| `submission_id` | BIGINT UNSIGNED FK→submissions | 提交 |
| `display_mode` | ENUM(`anonymous`,`showGroup`) | 展示模式 |
| `created_at` | DATETIME | 时间 |

约束：

- `UNIQUE(submission_id)`

### 5.2 `export_jobs` 可选

如果导出耗时明显，可引入异步导出任务。MVP 可以同步导出，不强制建表。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | BIGINT UNSIGNED PK | 任务 ID |
| `room_id` | BIGINT UNSIGNED FK→rooms | 房间 |
| `status` | ENUM(`pending`,`running`,`done`,`failed`) | 状态 |
| `file_url` | VARCHAR(1024) NULL | 导出结果 |
| `created_at` / `updated_at` | DATETIME | 时间 |

---

## 6. 迁移策略

当前已有 V1.x 表：

- `rooms`
- `groups`
- `students`
- `tasks`
- `task_target_groups`
- `submissions`
- `submission_images`
- `featured_answers`

V2.0 推荐新增 migration：

1. 新增 `admin_users`、`teacher_accounts`、`auth_sessions`。
2. `rooms` 新增 `teacher_id`，旧 `teacher_token` 改为兼容字段或保留原字段。
3. 新增 `submission_attachments`。
4. 将现有 `submission_images` 数据迁移为 `submission_attachments.kind = image`。
5. 后续确认稳定后再考虑废弃 `submission_images`。

---

## 7. 索引重点

| 查询 | 索引 |
| --- | --- |
| 管理者 / 讲师登录 | `username` unique |
| token 校验 | `token_hash` unique |
| 讲师活动列表 | `rooms.teacher_id` |
| 房间号入场 | `rooms.room_code` |
| 房间内昵称验重 | `students(room_id, nickname)` |
| 学生 token 恢复 | `students.client_token` |
| 任务列表 | `tasks.room_id` |
| 提交列表 | `submissions.task_id` |
| 导出聚合 | `submissions.room_id`, `submissions.group_id` |
| 附件导出 | `submission_attachments.submission_id` |
