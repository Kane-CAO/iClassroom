# iClassroom MVP 数据库设计

> 本文档对应 **Backend Step 1**：核心表结构与约束。
> 字段命名：数据库使用 `snake_case`，API 对外使用 `camelCase`（见 [`docs/api.md`](./api.md)）。
> 时间统一使用 **UTC**，类型为 `DATETIME`。
>
> - 迁移文件：`backend/migrations/000001_init_schema.{up,down}.sql`
> - 引擎 / 字符集：`InnoDB` / `utf8mb4` / `utf8mb4_unicode_ci`
> - **要求 MySQL 8.0.16+**（`CHECK` 约束自该版本起强制生效）

---

## 一、实体关系总览

```text
rooms (1) ──< groups (N)
rooms (1) ──< students (N) >── (1) groups
rooms (1) ──< tasks (N)
tasks (1) ──< task_target_groups (N) >── (1) groups      # 仅 target_type = 'groups'
tasks (1) ──< submissions (N) >── (1) students
submissions (1) ──< submission_images (N)
submissions (1) ──< featured_answers (1)                 # 一个提交最多被精选一次
```

所有子表对父表使用 `ON DELETE CASCADE`，保证删除房间时数据整体清理（MVP 不主动删除）。

---

## 二、表结构

### 1. `rooms` 房间

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | BIGINT UNSIGNED PK | 主键 |
| `room_code` | VARCHAR(16) | 房间码，**唯一** |
| `title` | VARCHAR(255) | 课堂名称 |
| `status` | ENUM(`created`,`active`,`ended`) | 房间状态，默认 `created` |
| `group_count` | INT | 小组数，默认 6 |
| `group_capacity` | INT | 每组容量，默认 10 |
| `allow_choose_group` | TINYINT(1) | 是否允许自选小组，默认 1 |
| `teacher_token` | VARCHAR(64) | 房间级管理凭证，**唯一**（用于鉴权，不可预测） |
| `created_at` / `updated_at` | DATETIME | 创建 / 更新时间 |
| `ended_at` | DATETIME NULL | 结束时间，未结束为 NULL |

**约束 / 索引**：`uk_rooms_room_code (room_code)`、`uk_rooms_teacher_token (teacher_token)`

### 2. `groups` 小组

> `groups` 是 MySQL 保留字，SQL 中需用反引号 `` `groups` ``。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | BIGINT UNSIGNED PK | 主键 |
| `room_id` | BIGINT UNSIGNED FK→rooms | 所属房间 |
| `group_name` | VARCHAR(64) | 组名（如「第1组」） |
| `capacity` | INT | 容量，默认 10 |
| `score_total` | INT | 小组累计总分，默认 0（评分时按差值事务更新） |
| `created_at` / `updated_at` | DATETIME | 时间 |

**约束 / 索引**：`idx_groups_room_id`、FK `fk_groups_room`

> 当前成员数（API 的 `currentCount`）不落库，由 `COUNT(students)` 实时计算，避免计数漂移。

### 3. `students` 学生

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | BIGINT UNSIGNED PK | 主键 |
| `room_id` | BIGINT UNSIGNED FK→rooms | 所属房间 |
| `group_id` | BIGINT UNSIGNED FK→groups | 所属小组 |
| `nickname` | VARCHAR(64) | 昵称，**同房间内唯一** |
| `client_token` | VARCHAR(64) | 学生会话凭证，**唯一**（断线恢复用，不可预测） |
| `created_at` / `updated_at` | DATETIME | 时间 |

**约束 / 索引**：
- `uk_students_room_nickname (room_id, nickname)` — 房间内昵称唯一
- `uk_students_client_token (client_token)`
- `idx_students_room_id`、`idx_students_group_id`

### 4. `tasks` 任务

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | BIGINT UNSIGNED PK | 主键 |
| `room_id` | BIGINT UNSIGNED FK→rooms | 所属房间 |
| `title` | VARCHAR(255) | 标题（必填） |
| `description` | TEXT NULL | 描述 |
| `attachment_url` | VARCHAR(1024) NULL | 可选附件链接 |
| `deadline_at` | DATETIME | 截止时间（UTC） |
| `target_type` | ENUM(`all`,`groups`) | 发布范围，默认 `all` |
| `status` | ENUM(`published`,`paused`,`closed`) | 任务状态，默认 `published` |
| `created_at` / `updated_at` | DATETIME | 时间 |

**约束 / 索引**：`idx_tasks_room_id`、`idx_tasks_room_status (room_id, status)`、FK `fk_tasks_room`

### 5. `task_target_groups` 任务目标小组

> 仅当 `tasks.target_type = 'groups'` 时使用，每个目标小组一行。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | BIGINT UNSIGNED PK | 主键 |
| `task_id` | BIGINT UNSIGNED FK→tasks | 任务 |
| `group_id` | BIGINT UNSIGNED FK→groups | 目标小组 |

**约束 / 索引**：`uk_ttg_task_group (task_id, group_id)`、`idx_ttg_task_id`、`idx_ttg_group_id`

### 6. `submissions` 提交

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | BIGINT UNSIGNED PK | 主键 |
| `task_id` | BIGINT UNSIGNED FK→tasks | 所属任务 |
| `student_id` | BIGINT UNSIGNED FK→students | 提交学生 |
| `room_id` | BIGINT UNSIGNED FK→rooms | 冗余：所属房间（便于查询/导出） |
| `group_id` | BIGINT UNSIGNED FK→groups | 冗余：提交时所在小组（便于分组展示/聚合） |
| `content_text` | TEXT NULL | 文字内容 |
| `status` | ENUM(`submitted`,`graded`) | 提交状态，默认 `submitted` |
| `score` | INT NULL | 分数，**允许为空；有值时必须 1–10** |
| `comment` | VARCHAR(1024) NULL | 评语（`comment` 加反引号） |
| `submitted_at` | DATETIME | 提交时间 |
| `graded_at` | DATETIME NULL | 评分时间，未评为 NULL |
| `created_at` / `updated_at` | DATETIME | 时间 |

**约束 / 索引**：
- `uk_submissions_task_student (task_id, student_id)` — 一人一任务只能提交一次
- `chk_submissions_score CHECK (score IS NULL OR score BETWEEN 1 AND 10)` — 禁止 0 分及越界
- `idx_submissions_task_id`、`idx_submissions_student_id`、`idx_submissions_room_id`、`idx_submissions_group_id`

### 7. `submission_images` 提交图片

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | BIGINT UNSIGNED PK | 主键 |
| `submission_id` | BIGINT UNSIGNED FK→submissions | 所属提交 |
| `file_url` | VARCHAR(1024) | 可访问 URL（前端预览） |
| `file_path` | VARCHAR(1024) | 服务端存储路径（不对外暴露） |
| `file_name` | VARCHAR(255) | 文件名（后端重新生成，不信任原始名） |
| `file_size` | BIGINT UNSIGNED | 文件大小（字节） |
| `mime_type` | VARCHAR(64) | MIME 类型（jpeg/png/webp） |
| `created_at` | DATETIME | 时间 |

**约束 / 索引**：`idx_submission_images_submission_id`、FK `fk_submission_images_submission`

### 8. `featured_answers` 精选答案

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | BIGINT UNSIGNED PK | 主键（API 的 `featuredId`） |
| `room_id` | BIGINT UNSIGNED FK→rooms | 所属房间 |
| `submission_id` | BIGINT UNSIGNED FK→submissions | 精选的提交 |
| `display_mode` | ENUM(`anonymous`,`showGroup`) | 展示模式，默认 `anonymous` |
| `created_at` | DATETIME | 时间 |

**约束 / 索引**：`uk_featured_submission (submission_id)` — 一个提交最多精选一次、`idx_featured_room_id`、FK

---

## 三、关键约束汇总

| 约束 | 表 | 作用 |
| --- | --- | --- |
| `room_code` 唯一 | rooms | 房间码全局唯一 |
| `unique(room_id, nickname)` | students | 同房间昵称唯一 |
| `unique(task_id, student_id)` | submissions | 一人一任务一提交（重复提交兜底） |
| `CHECK score 1–10` | submissions | 评分合法性（禁止 0 分） |
| `unique(submission_id)` | featured_answers | 同一提交不重复精选 |

## 四、关键索引说明

| 常见查询 | 命中索引 |
| --- | --- |
| 按 `room_code` 找房间 | `uk_rooms_room_code` |
| 按 `teacher_token` / `client_token` 鉴权 | `uk_rooms_teacher_token` / `uk_students_client_token` |
| 房间下的小组 / 学生 / 任务 | `idx_groups_room_id` / `idx_students_room_id` / `idx_tasks_room_id` |
| 房间内某状态任务列表 | `idx_tasks_room_status (room_id, status)` |
| 某任务的全部提交 | `idx_submissions_task_id` |
| 某学生的全部提交 | `idx_submissions_student_id` |
| 按房间 / 小组聚合提交（导出、排行榜） | `idx_submissions_room_id` / `idx_submissions_group_id` |
| 某提交的图片 | `idx_submission_images_submission_id` |

---

## 五、如何执行迁移

### 方式 A：MySQL 客户端（最简单）

```bash
# 先确保数据库存在（库名与 .env 的 DB_NAME 一致，默认 iclassroom）
mysql -h 127.0.0.1 -P 3306 -u root -p -e "CREATE DATABASE IF NOT EXISTS iclassroom CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 应用迁移
mysql -h 127.0.0.1 -P 3306 -u root -p iclassroom < backend/migrations/000001_init_schema.up.sql

# 回滚
mysql -h 127.0.0.1 -P 3306 -u root -p iclassroom < backend/migrations/000001_init_schema.down.sql
```

### 方式 B：golang-migrate CLI（文件名已遵循其命名规范）

```bash
# 安装：brew install golang-migrate
migrate -path backend/migrations \
  -database "mysql://root:PASSWORD@tcp(127.0.0.1:3306)/iclassroom" up

# 回滚最近一次
migrate -path backend/migrations \
  -database "mysql://root:PASSWORD@tcp(127.0.0.1:3306)/iclassroom" down 1
```

### 验证

```bash
mysql -u root -p iclassroom -e "SHOW TABLES;"
# 应看到 8 张表：rooms, groups, students, tasks,
#                task_target_groups, submissions, submission_images, featured_answers

mysql -u root -p iclassroom -e "SHOW CREATE TABLE submissions\G"   # 确认 CHECK / UNIQUE 已生效
```

> **UTC 提示**：为使 `DEFAULT CURRENT_TIMESTAMP` 写入 UTC，启动 MySQL 时设置
> `--default-time-zone='+00:00'`，或在会话中 `SET time_zone = '+00:00';`。
> 业务时间字段（`deadline_at` / `submitted_at` / `graded_at`）由应用以 UTC 写入。
