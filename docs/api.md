# iClassroom MVP API Contract

> 本文档为 iClassroom MVP 阶段的 API 契约（API Contract），用于前后端联调对齐。
> 本阶段只定义接口契约，不代表后端已实现。字段、错误码以本文档为准。
>
> - Base URL（HTTP）：`http://localhost:8080/api`（前端通过 `VITE_API_BASE_URL` 注入，禁止硬编码）
> - Base URL（WebSocket，后续接入）：`ws://localhost:8080/ws`
> - 学生端入口：`http://localhost:5173/student`
> - 所有时间字段统一使用 UTC ISO-8601 格式（例如 `2026-06-10T18:00:00Z`）。
> - 所有字段命名统一使用 **camelCase**，禁止混用 snake_case。

---

## 目录

1. [身份与权限设计](#一身份与权限设计)
2. [统一 API 响应格式](#二统一-api-响应格式)
3. [统一 Header 规范](#三统一-header-规范)
4. [MVP API 明细](#四mvp-api-明细)
5. [错误码汇总](#五错误码汇总)
6. [待确认问题](#六待确认问题)

---

## 一、身份与权限设计

MVP 阶段**不做学生账号注册**，也**不做老师账号登录系统**。

> **重要：不登录账号 ≠ 无权限控制。**
> 本项目采用**房间级临时身份凭证**进行权限控制。

### 1. 学生端身份

学生端无需登录。进入房间流程：

```text
扫码 / 链接进入
→ 输入 nickname
→ 选择 group
→ 后端创建 student
→ 后端返回 clientToken
→ 前端将 roomCode、studentId、clientToken、nickname、groupId 存入 localStorage
```

学生后续访问自己的任务、提交答案、查看结果时，必须携带：

```http
X-Student-Token: student_xxxxx
```

### 2. 老师端身份

老师端 MVP 不做账号登录。老师创建房间后，后端返回一个**房间级管理凭证** `teacherToken`，老师前端将其存入 localStorage。

老师后续管理房间、发布/暂停/关闭任务、查看提交、评分、查看大屏与后台数据、结束课堂、导出数据时，必须携带：

```http
X-Teacher-Token: teacher_xxxxx
```

### 3. roomCode 的作用

`roomCode` **只用于识别房间和学生入场**，**不能作为老师端管理权限**。

原因：学生扫码链接中也包含 `roomCode`。如果老师端只依赖 `roomCode`，学生知道 `roomCode` 后即可访问老师接口，这是不允许的。因此老师端权限必须依赖不可预测的 `teacherToken`。

### 4. MVP 权限结论

```text
学生端：不登录，使用 nickname + group + clientToken
老师端：不登录账号，使用 roomCode + teacherToken 管理房间
大屏端：由老师端打开，MVP 可复用 teacherToken，后续可扩展 displayToken
```

---

## 二、统一 API 响应格式

所有接口（导出二进制接口除外）均返回统一 JSON 结构。

### 成功响应

```json
{
  "success": true,
  "message": "success",
  "data": {}
}
```

### 错误响应

```json
{
  "success": false,
  "message": "nickname already exists",
  "errorCode": "NICKNAME_DUPLICATED"
}
```

字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `success` | boolean | 是否成功 |
| `message` | string | 可读信息，成功时为 `"success"` |
| `data` | object / array / null | 业务数据，错误时可省略或为 `null` |
| `errorCode` | string | 错误码，仅错误响应包含，见[错误码汇总](#五错误码汇总) |

> 字段命名统一 camelCase，例如：
> `roomCode` `teacherToken` `clientToken` `studentId` `groupId` `taskId` `submissionId` `deadlineAt` `createdAt` `submittedAt` `gradedAt`

---

## 三、统一 Header 规范

### 老师端接口 Header

```http
X-Teacher-Token: teacher_xxxxx
```

适用于所有老师管理接口：获取房间详情、获取 overview、创建任务、获取任务列表、暂停任务、关闭任务、查看提交、评分、设置精选答案、查看大屏数据、查看后台数据、结束课堂、导出数据。

### 学生端接口 Header

```http
X-Student-Token: student_xxxxx
```

适用于学生身份相关接口：resume 恢复会话、获取我的任务、提交任务、查看我的结果、查看小组排名。

### 图片上传 Header

```http
Content-Type: multipart/form-data
X-Student-Token: student_xxxxx
```

### Header 速查表

| 接口 | Method | Path | X-Teacher-Token | X-Student-Token |
| --- | --- | --- | :---: | :---: |
| 创建房间 | POST | `/api/teacher/rooms` | — | — |
| 房间详情 | GET | `/api/teacher/rooms/:roomCode` | ✅ | — |
| 房间 overview | GET | `/api/teacher/rooms/:roomCode/overview` | ✅ | — |
| 学生获取房间信息 | GET | `/api/student/rooms/:roomCode` | — | — |
| 学生加入房间 | POST | `/api/student/rooms/:roomCode/join` | — | — |
| 学生 resume | POST | `/api/student/rooms/:roomCode/resume` | — | ✅ |
| 发布任务 | POST | `/api/teacher/rooms/:roomCode/tasks` | ✅ | — |
| 老师任务列表 | GET | `/api/teacher/rooms/:roomCode/tasks` | ✅ | — |
| 学生任务列表 | GET | `/api/student/me/tasks` | — | ✅ |
| 暂停任务 | PATCH | `/api/teacher/tasks/:taskId/pause` | ✅ | — |
| 关闭任务 | PATCH | `/api/teacher/tasks/:taskId/close` | ✅ | — |
| 学生提交任务 | POST | `/api/student/tasks/:taskId/submit` | — | ✅ |
| 查看任务提交 | GET | `/api/teacher/tasks/:taskId/submissions` | ✅ | — |
| 评分 | POST | `/api/teacher/submissions/:submissionId/grade` | ✅ | — |
| 学生查看结果 | GET | `/api/student/me/results` | — | ✅ |
| 排行榜 | GET | `/api/student/rooms/:roomCode/ranking` | — | ✅ |
| 大屏看板 | GET | `/api/teacher/rooms/:roomCode/display` | ✅ | — |
| 设置精选答案 | POST | `/api/teacher/submissions/:submissionId/feature` | ✅ | — |
| 后台数据看板 | GET | `/api/teacher/rooms/:roomCode/analytics` | ✅ | — |
| 结束课堂 | POST | `/api/teacher/rooms/:roomCode/end` | ✅ | — |
| 导出数据 | GET | `/api/teacher/rooms/:roomCode/export` | ✅ | — |

---

## 四、MVP API 明细

---

### 1. 讲师创建房间

**功能**：老师创建课堂房间。MVP 不需要老师登录账号，创建成功后返回 `teacherToken`。

- **Method**：`POST`
- **Path**：`/api/teacher/rooms`
- **Header**：无（这是获取 `teacherToken` 的入口接口）

**Request**

```json
{
  "title": "Demo Class",
  "groupCount": 6,
  "groupCapacity": 10,
  "allowChooseGroup": true
}
```

**Response**

```json
{
  "success": true,
  "message": "success",
  "data": {
    "roomCode": "ABC123",
    "teacherToken": "teacher_xxxxx",
    "joinUrl": "http://localhost:5173/student?room=ABC123",
    "teacherDashboardUrl": "http://localhost:5173/teacher/rooms/ABC123/dashboard",
    "groups": [
      { "groupId": 1, "groupName": "第1组", "capacity": 10, "currentCount": 0 }
    ]
  }
}
```

**权限规则**

- 入口接口，无需 token。创建成功后返回的 `teacherToken` 用于后续所有老师管理接口的鉴权。

**业务规则**

- `roomCode` 必须唯一。
- `teacherToken` 必须不可预测。
- 创建房间时自动创建 groups。
- `groupCount` 默认 6，`groupCapacity` 默认 10。
- 创建房间与创建小组必须在**同一事务**中完成。
- `teacherToken` 只返回给创建者，前端本地保存。
- 后续老师端管理接口必须通过 `X-Teacher-Token` 校验。

**错误场景**

```text
INVALID_GROUP_COUNT
INVALID_GROUP_CAPACITY
ROOM_CREATE_FAILED
```

---

### 2. 讲师获取房间详情

**功能**：根据 `roomCode` 获取房间基础信息。

- **Method**：`GET`
- **Path**：`/api/teacher/rooms/:roomCode`
- **Header**：`X-Teacher-Token: teacher_xxxxx`

**Response**

```json
{
  "success": true,
  "message": "success",
  "data": {
    "roomCode": "ABC123",
    "title": "Demo Class",
    "status": "active",
    "groupCount": 6,
    "groupCapacity": 10,
    "allowChooseGroup": true,
    "createdAt": "2026-06-10T10:00:00Z"
  }
}
```

**权限规则**

- 必须携带 `X-Teacher-Token`，且该 token 必须属于本房间。

**业务规则**

- `status` 取值：`active` / `ended`。

**错误场景**

```text
ROOM_NOT_FOUND
INVALID_TEACHER_TOKEN
ROOM_ACCESS_DENIED
```

---

### 3. 讲师获取房间 Overview

**功能**：获取房间总览（学生数、小组得分、任务列表），用于老师主控台。

- **Method**：`GET`
- **Path**：`/api/teacher/rooms/:roomCode/overview`
- **Header**：`X-Teacher-Token: teacher_xxxxx`

**Response**

```json
{
  "success": true,
  "message": "success",
  "data": {
    "roomCode": "ABC123",
    "title": "Demo Class",
    "status": "active",
    "joinUrl": "http://localhost:5173/student?room=ABC123",
    "studentCount": 12,
    "groups": [
      { "groupId": 1, "groupName": "第1组", "capacity": 10, "currentCount": 3, "scoreTotal": 18 }
    ],
    "tasks": []
  }
}
```

**权限规则**

- 必须携带 `X-Teacher-Token`，且该 token 必须属于本房间。

**业务规则**

- `scoreTotal` 为小组累计总分。
- `tasks` 为房间任务概览数组（可为空）。

**错误场景**

```text
ROOM_NOT_FOUND
INVALID_TEACHER_TOKEN
ROOM_ACCESS_DENIED
```

---

### 4. 学生获取房间信息

**功能**：学生打开链接后，根据 `roomCode` 获取房间信息和可选小组。

- **Method**：`GET`
- **Path**：`/api/student/rooms/:roomCode`
- **Header**：无（学生入场前尚无 token）

**Response**

```json
{
  "success": true,
  "message": "success",
  "data": {
    "roomCode": "ABC123",
    "title": "Demo Class",
    "status": "active",
    "groups": [
      { "groupId": 1, "groupName": "第1组", "capacity": 10, "currentCount": 3, "available": true }
    ]
  }
}
```

**权限规则**

- 公开接口，不需要 token。

**业务规则**

- 房间不存在返回 `ROOM_NOT_FOUND`。
- 房间已结束返回 `ROOM_ENDED`，学生不能进入。
- 已满小组 `available = false`。

**错误场景**

```text
ROOM_NOT_FOUND
ROOM_ENDED
```

---

### 5. 学生加入房间

**功能**：学生输入 nickname、选择 group 后加入房间，后端创建 student 并返回 `clientToken`。

- **Method**：`POST`
- **Path**：`/api/student/rooms/:roomCode/join`
- **Header**：无（这是获取 `clientToken` 的入口接口）

**Request**

```json
{
  "nickname": "Tom",
  "groupId": 1
}
```

**Response**

```json
{
  "success": true,
  "message": "success",
  "data": {
    "studentId": 1,
    "clientToken": "student_xxxxx",
    "roomCode": "ABC123",
    "nickname": "Tom",
    "groupId": 1,
    "groupName": "第1组"
  }
}
```

**权限规则**

- 入口接口，无需 token。返回的 `clientToken` 用于后续学生身份接口鉴权。

**业务规则**

- 学生无需登录。
- 房间必须存在，且不能是 `ended`。
- `nickname` 在同一 room 内必须唯一。
- `group` 必须属于当前 room。
- 小组已满则不能加入。
- 后端生成不可预测 `clientToken`。
- 前端将 `clientToken` 保存到 localStorage。

**错误场景**

```text
ROOM_NOT_FOUND
ROOM_ENDED
NICKNAME_DUPLICATED
GROUP_NOT_FOUND
GROUP_FULL
INVALID_NICKNAME
```

---

### 6. 学生恢复会话 resume

**功能**：使用 `roomCode + clientToken` 恢复学生身份（页面刷新 / 重新进入）。

- **Method**：`POST`
- **Path**：`/api/student/rooms/:roomCode/resume`
- **Header**：`X-Student-Token: student_xxxxx`

**Response**

```json
{
  "success": true,
  "message": "success",
  "data": {
    "studentId": 1,
    "roomCode": "ABC123",
    "nickname": "Tom",
    "groupId": 1,
    "groupName": "第1组",
    "roomStatus": "active"
  }
}
```

**权限规则**

- 必须携带 `X-Student-Token`，且该 token 必须属于本房间。

**业务规则**

- 使用 `roomCode + clientToken` 恢复学生身份。
- room `ended` 后学生不能继续进入课堂操作，返回 `ROOM_ENDED`。
- token 无效返回 `INVALID_STUDENT_TOKEN`。

**错误场景**

```text
INVALID_STUDENT_TOKEN
ROOM_NOT_FOUND
ROOM_ENDED
```

---

### 7. 讲师发布任务

**功能**：老师在房间内发布任务。

- **Method**：`POST`
- **Path**：`/api/teacher/rooms/:roomCode/tasks`
- **Header**：`X-Teacher-Token: teacher_xxxxx`

**Request**

```json
{
  "title": "课堂测试一",
  "description": "请提交你的思路",
  "attachmentUrl": "",
  "deadlineAt": "2026-06-10T18:00:00Z",
  "targetType": "all",
  "targetGroupIds": []
}
```

**Response**

```json
{
  "success": true,
  "message": "success",
  "data": {
    "taskId": 1,
    "roomCode": "ABC123",
    "title": "课堂测试一",
    "description": "请提交你的思路",
    "deadlineAt": "2026-06-10T18:00:00Z",
    "targetType": "all",
    "targetGroupIds": [],
    "status": "published"
  }
}
```

**权限规则**

- 必须校验 `X-Teacher-Token`，且该 token 必须属于本房间。

**业务规则**

- 标题必填。
- `deadlineAt` 必须晚于当前时间。
- `targetType` 只能是 `all` 或 `groups`。
- `targetType = groups` 时，`targetGroupIds` 不能为空。
- 指定小组必须属于当前 room。
- 房间 `ended` 后不能发布任务。
- `status` 取值：`published` / `paused` / `closed`。

**错误场景**

```text
INVALID_TEACHER_TOKEN
ROOM_NOT_FOUND
ROOM_ENDED
INVALID_TASK_TITLE
INVALID_DEADLINE
INVALID_TARGET_TYPE
INVALID_TARGET_GROUP
```

---

### 8. 讲师获取任务列表

**功能**：获取房间内所有任务及其提交统计。

- **Method**：`GET`
- **Path**：`/api/teacher/rooms/:roomCode/tasks`
- **Header**：`X-Teacher-Token: teacher_xxxxx`

**Response**

```json
{
  "success": true,
  "message": "success",
  "data": [
    {
      "taskId": 1,
      "title": "课堂测试一",
      "description": "请提交你的思路",
      "deadlineAt": "2026-06-10T18:00:00Z",
      "targetType": "all",
      "status": "published",
      "submittedCount": 8,
      "targetStudentCount": 20
    }
  ]
}
```

**权限规则**

- 必须携带 `X-Teacher-Token`，且该 token 必须属于本房间。

**业务规则**

- `submittedCount` 为已提交学生数；`targetStudentCount` 为任务目标范围内学生总数。

**错误场景**

```text
INVALID_TEACHER_TOKEN
ROOM_NOT_FOUND
ROOM_ACCESS_DENIED
```

---

### 9. 学生获取自己的任务列表

**功能**：学生查看自己可见的任务及自身提交状态。

- **Method**：`GET`
- **Path**：`/api/student/me/tasks`
- **Header**：`X-Student-Token: student_xxxxx`

**Response**

```json
{
  "success": true,
  "message": "success",
  "data": [
    {
      "taskId": 1,
      "title": "课堂测试一",
      "description": "请提交你的思路",
      "deadlineAt": "2026-06-10T18:00:00Z",
      "status": "published",
      "mySubmissionStatus": "notSubmitted",
      "myScore": null
    }
  ]
}
```

**权限规则**

- 必须携带 `X-Student-Token`。学生身份由 token 解析得到（无需传 studentId）。

**业务规则**

- 学生只能看到全班任务（`targetType = all`）或自己小组范围内的任务（`targetType = groups` 且包含自己组）。
- `mySubmissionStatus` 取值：`notSubmitted` / `submitted` / `graded`。
- 任务已暂停、关闭或截止时，前端仍可展示，但不可提交。
- **默认策略**：房间 `ended` 后，本接口返回 `ROOM_ENDED`；学生端应提示"课堂已结束"，引导其改用[学生查看结果](#15-学生查看自己的结果)接口只读查看历史结果（详见[待确认问题 3](#六待确认问题)）。

**错误场景**

```text
INVALID_STUDENT_TOKEN
ROOM_ENDED
```

---

### 10. 暂停任务

**功能**：老师暂停任务，暂停期间学生不能提交。

- **Method**：`PATCH`
- **Path**：`/api/teacher/tasks/:taskId/pause`
- **Header**：`X-Teacher-Token: teacher_xxxxx`

**Response**

```json
{
  "success": true,
  "message": "success",
  "data": { "taskId": 1, "status": "paused" }
}
```

**权限规则**

- 只有老师能暂停。`X-Teacher-Token` 必须属于该任务所在房间。

**业务规则**

- 暂停后学生不能继续提交。
- 已提交内容不受影响。

**错误场景**

```text
INVALID_TEACHER_TOKEN
TASK_NOT_FOUND
ROOM_ACCESS_DENIED
```

---

### 11. 关闭任务

**功能**：老师关闭任务，关闭后学生不能提交。

- **Method**：`PATCH`
- **Path**：`/api/teacher/tasks/:taskId/close`
- **Header**：`X-Teacher-Token: teacher_xxxxx`

**Response**

```json
{
  "success": true,
  "message": "success",
  "data": { "taskId": 1, "status": "closed" }
}
```

**权限规则**

- 只有老师能关闭。`X-Teacher-Token` 必须属于该任务所在房间。

**业务规则**

- `closed` 后学生不能提交。
- 已提交内容不受影响。

**错误场景**

```text
INVALID_TEACHER_TOKEN
TASK_NOT_FOUND
ROOM_ACCESS_DENIED
```

---

### 12. 学生提交任务

**功能**：学生提交文字答案及图片（multipart 上传）。

- **Method**：`POST`
- **Path**：`/api/student/tasks/:taskId/submit`
- **Header**：
  ```http
  X-Student-Token: student_xxxxx
  Content-Type: multipart/form-data
  ```

**Form Data**

```text
contentText: "我的答案内容"
images[]: file1
images[]: file2
```

**Response**

```json
{
  "success": true,
  "message": "success",
  "data": {
    "submissionId": 1,
    "taskId": 1,
    "status": "submitted",
    "contentText": "我的答案内容",
    "images": [
      {
        "imageId": 1,
        "fileUrl": "http://localhost:8080/uploads/rooms/ABC123/tasks/1/students/1/image1.jpg",
        "fileName": "image1.jpg",
        "fileSize": 102400,
        "mimeType": "image/jpeg"
      }
    ],
    "submittedAt": "2026-06-10T17:30:00Z"
  }
}
```

**权限规则**

- 必须携带 `X-Student-Token`。
- 任务必须属于学生所在 room，且学生必须在任务目标范围内。

**业务规则**

- 任务必须存在。
- 任务 `paused` / `closed` / 超过 `deadline` 后不能提交。
- 一个学生对一个任务只能提交一次。
- 图片直接上传，不接受外链。
- 单题最多 **3 张**图片。
- 单图最大 **5MB**。
- 允许类型：`image/jpeg`、`image/png`、`image/webp`。
- 后端必须校验图片数量、大小、MIME。
- 文件名必须由后端重新生成，**不信任用户原始文件名**。
- 上传失败时前端应保留文字草稿。

**错误场景**

```text
INVALID_STUDENT_TOKEN
TASK_NOT_FOUND
TASK_NOT_IN_STUDENT_ROOM
TASK_NOT_TARGETED_TO_STUDENT
TASK_PAUSED
TASK_CLOSED
TASK_DEADLINE_PASSED
SUBMISSION_ALREADY_EXISTS
TOO_MANY_IMAGES
IMAGE_TOO_LARGE
INVALID_IMAGE_TYPE
UPLOAD_FAILED
```

---

### 13. 讲师查看任务提交

**功能**：老师查看某任务的全部提交（含图片、评分），便于分组展示与批改。

- **Method**：`GET`
- **Path**：`/api/teacher/tasks/:taskId/submissions`
- **Header**：`X-Teacher-Token: teacher_xxxxx`

**Response**

```json
{
  "success": true,
  "message": "success",
  "data": [
    {
      "submissionId": 1,
      "taskId": 1,
      "studentId": 1,
      "nickname": "Tom",
      "groupId": 1,
      "groupName": "第1组",
      "contentText": "我的答案内容",
      "images": [
        {
          "imageId": 1,
          "fileUrl": "http://localhost:8080/uploads/rooms/ABC123/tasks/1/students/1/image1.jpg",
          "fileName": "image1.jpg"
        }
      ],
      "status": "submitted",
      "score": null,
      "comment": "",
      "submittedAt": "2026-06-10T17:30:00Z",
      "gradedAt": null
    }
  ]
}
```

**权限规则**

- 只能由对应房间的 `teacherToken` 查看。

**业务规则**

- 返回结果包含 `groupId` / `groupName` 字段，便于前端按组分组展示。
- `fileUrl` 必须可直接用于前端预览。

**错误场景**

```text
INVALID_TEACHER_TOKEN
TASK_NOT_FOUND
ROOM_ACCESS_DENIED
```

---

### 14. 讲师评分

**功能**：老师对单个提交评分并写评语，自动累计小组总分。

- **Method**：`POST`
- **Path**：`/api/teacher/submissions/:submissionId/grade`
- **Header**：`X-Teacher-Token: teacher_xxxxx`

**Request**

```json
{
  "score": 8,
  "comment": "思路清晰"
}
```

**Response**

```json
{
  "success": true,
  "message": "success",
  "data": {
    "submissionId": 1,
    "score": 8,
    "comment": "思路清晰",
    "status": "graded",
    "gradedAt": "2026-06-10T18:00:00Z",
    "groupScoreTotal": 18
  }
}
```

**权限规则**

- 必须携带 `X-Teacher-Token`，且该 token 必须属于提交所在房间。

**业务规则**

- `score` 必须是整数，范围 **1–10**，**不允许 0 分**。
- 可填写 `comment`。
- 保存评分后 `submission.status = graded`。
- 小组总分自动累计。
- **重新评分必须使用差值更新**（`新总分 = 旧总分 - 旧分 + 新分`），不能重复累计。
- submission 更新与 group score 更新必须在**同一事务**中完成。

**错误场景**

```text
INVALID_TEACHER_TOKEN
SUBMISSION_NOT_FOUND
INVALID_SCORE
ROOM_ACCESS_DENIED
GRADE_FAILED
```

---

### 15. 学生查看自己的结果

**功能**：学生查看自己所有任务的提交状态、得分与评语。

- **Method**：`GET`
- **Path**：`/api/student/me/results`
- **Header**：`X-Student-Token: student_xxxxx`

**Response**

```json
{
  "success": true,
  "message": "success",
  "data": {
    "studentId": 1,
    "nickname": "Tom",
    "groupId": 1,
    "groupName": "第1组",
    "results": [
      {
        "taskId": 1,
        "taskTitle": "课堂测试一",
        "submissionStatus": "graded",
        "score": 8,
        "comment": "思路清晰",
        "submittedAt": "2026-06-10T17:30:00Z",
        "gradedAt": "2026-06-10T18:00:00Z"
      }
    ]
  }
}
```

**权限规则**

- 必须携带 `X-Student-Token`，学生身份由 token 解析。

**业务规则**

- `submissionStatus` 取值：`notSubmitted` / `submitted` / `graded`。
- 未提交或未评分时，`score` / `gradedAt` 为 `null`。

**错误场景**

```text
INVALID_STUDENT_TOKEN
STUDENT_NOT_FOUND
```

---

### 16. 获取排行榜

**功能**：按小组总分排序的排行榜。

- **Method**：`GET`
- **Path**：`/api/student/rooms/:roomCode/ranking`
- **Header**：`X-Student-Token: student_xxxxx`

**Response**

```json
{
  "success": true,
  "message": "success",
  "data": [
    { "rank": 1, "groupId": 1, "groupName": "第1组", "scoreTotal": 35, "studentCount": 6 }
  ]
}
```

**权限规则**

- 必须携带 `X-Student-Token`，学生只能查看自己所在 room 的 ranking。

**业务规则**

- 排行榜按 `scoreTotal` 降序。
- 大屏与学生端可使用同一数据源。

**错误场景**

```text
INVALID_STUDENT_TOKEN
ROOM_NOT_FOUND
ROOM_ACCESS_DENIED
```

---

### 17. 大屏看板数据

**功能**：老师端全屏展示视图所需数据（排行榜、当前任务完成度、精选答案）。

- **Method**：`GET`
- **Path**：`/api/teacher/rooms/:roomCode/display`
- **Header**：`X-Teacher-Token: teacher_xxxxx`

**Response**

```json
{
  "success": true,
  "message": "success",
  "data": {
    "roomCode": "ABC123",
    "title": "Demo Class",
    "ranking": [
      { "rank": 1, "groupId": 1, "groupName": "第1组", "scoreTotal": 35 }
    ],
    "currentTask": {
      "taskId": 1,
      "title": "课堂测试一",
      "deadlineAt": "2026-06-10T18:00:00Z",
      "submittedCount": 8,
      "targetStudentCount": 20,
      "completionRate": 0.4
    },
    "featuredAnswers": []
  }
}
```

**权限规则**

- 大屏是老师端内的全屏展示视图。MVP 复用 `teacherToken`（后续可扩展只读的 `displayToken`，见[待确认问题 2](#六待确认问题)）。

**业务规则**

- 数据源应与排行榜、提交完成度保持一致。
- `targetStudentCount` 为 0 时 `completionRate` 返回 0，不能除零报错。

**错误场景**

```text
INVALID_TEACHER_TOKEN
ROOM_NOT_FOUND
ROOM_ACCESS_DENIED
```

---

### 18. 设置精选答案

**功能**：老师将某提交设为精选答案，用于大屏展示。

- **Method**：`POST`
- **Path**：`/api/teacher/submissions/:submissionId/feature`
- **Header**：`X-Teacher-Token: teacher_xxxxx`

**Request**

```json
{
  "displayMode": "anonymous"
}
```

**Response**

```json
{
  "success": true,
  "message": "success",
  "data": {
    "featuredId": 1,
    "submissionId": 1,
    "displayMode": "anonymous"
  }
}
```

**权限规则**

- 必须携带 `X-Teacher-Token`，且精选答案必须属于当前老师管理的房间。

**业务规则**

- `displayMode` 支持 `anonymous` / `showGroup`。
- `anonymous` 模式下不返回学生昵称。

**错误场景**

```text
INVALID_TEACHER_TOKEN
SUBMISSION_NOT_FOUND
ROOM_ACCESS_DENIED
INVALID_DISPLAY_MODE
```

---

### 19. 后台数据看板

**功能**：讲师私下查看的统计数据（区别于大屏展示）。

- **Method**：`GET`
- **Path**：`/api/teacher/rooms/:roomCode/analytics`
- **Header**：`X-Teacher-Token: teacher_xxxxx`

**Response**

```json
{
  "success": true,
  "message": "success",
  "data": {
    "studentCount": 30,
    "onlineCount": 18,
    "submissionRate": 0.76,
    "groupScores": [
      { "groupId": 1, "groupName": "第1组", "scoreTotal": 35 }
    ],
    "taskCompletion": [
      {
        "taskId": 1,
        "taskTitle": "课堂测试一",
        "submittedCount": 8,
        "targetStudentCount": 20,
        "completionRate": 0.4
      }
    ],
    "submissionTimeline": [
      { "time": "2026-06-10T17:30:00Z", "count": 3 }
    ]
  }
}
```

**权限规则**

- 必须携带 `X-Teacher-Token`，且该 token 必须属于本房间。

**业务规则**

- 后台数据看板是讲师私下查看，不是大屏展示。
- 无数据时不能报错（返回空数组 / 0）。
- 除数为 0 时 `completionRate` / `submissionRate` 应返回 0。

**错误场景**

```text
INVALID_TEACHER_TOKEN
ROOM_NOT_FOUND
ROOM_ACCESS_DENIED
```

---

### 20. 结束课堂

**功能**：老师结束课堂，房间进入 `ended` 状态。

- **Method**：`POST`
- **Path**：`/api/teacher/rooms/:roomCode/end`
- **Header**：`X-Teacher-Token: teacher_xxxxx`

**Response**

```json
{
  "success": true,
  "message": "success",
  "data": {
    "roomCode": "ABC123",
    "status": "ended",
    "endedAt": "2026-06-10T19:00:00Z"
  }
}
```

**权限规则**

- 只有老师能结束课堂。`X-Teacher-Token` 必须属于本房间。

**业务规则**

- 结束后 `room.status = ended`。
- `ended` 后学生不能 `join`。
- `ended` 后学生不能 `resume` 进入课堂操作。
- `ended` 后学生不能 `submit`。
- 老师仍可查看数据和导出。
- 前端结束课堂必须有**二次确认**。

**错误场景**

```text
INVALID_TEACHER_TOKEN
ROOM_NOT_FOUND
ROOM_ALREADY_ENDED
```

---

### 21. 导出 Excel + 图片 Zip

**功能**：导出房间所有提交为 Excel，并打包图片原件为 Zip。

- **Method**：`GET`
- **Path**：`/api/teacher/rooms/:roomCode/export`
- **Header**：`X-Teacher-Token: teacher_xxxxx`

**Response（二进制流，非统一 JSON 格式）**

```http
Content-Type: application/zip
Content-Disposition: attachment; filename="export_room_ABC123.zip"
```

> 注意：成功时返回二进制 Zip 流，不遵循统一 JSON 响应格式。失败时仍返回统一 JSON 错误响应。

**Zip 结构**

```text
export_room_ABC123.zip
├── submissions.xlsx
└── images/
    ├── task_1/
    │   ├── group_1/
    │   │   └── student_tom_img1.jpg
```

**Excel 字段（至少包含）**

```text
roomCode
roomTitle
groupName
studentNickname
taskTitle
contentText
imageFileNames
score
comment
submittedAt
gradedAt
```

**权限规则**

- 只有老师能导出。`X-Teacher-Token` 必须属于本房间。

**业务规则**

- 导出必须包含 Excel。
- 有图片时必须包含图片原件。
- 没有图片时也能导出 Excel。
- 导出失败要返回明确错误。

**错误场景**

```text
INVALID_TEACHER_TOKEN
ROOM_NOT_FOUND
EXPORT_FAILED
IMAGE_FILE_MISSING
```

---

## 五、错误码汇总

| 错误码 | 含义 | HTTP（建议） |
| --- | --- | --- |
| `ROOM_NOT_FOUND` | 房间不存在 | 404 |
| `ROOM_ENDED` | 房间已结束（学生入场/操作时） | 409 |
| `ROOM_ALREADY_ENDED` | 房间已结束（老师重复结束时） | 409 |
| `ROOM_ACCESS_DENIED` | token 有效但无权访问该房间资源 | 403 |
| `ROOM_CREATE_FAILED` | 房间创建失败 | 500 |
| `INVALID_GROUP_COUNT` | groupCount 非法 | 400 |
| `INVALID_GROUP_CAPACITY` | groupCapacity 非法 | 400 |
| `INVALID_TEACHER_TOKEN` | 老师 token 无效或缺失 | 401 |
| `INVALID_STUDENT_TOKEN` | 学生 token 无效或缺失 | 401 |
| `INVALID_NICKNAME` | nickname 非法（空 / 超长等） | 400 |
| `NICKNAME_DUPLICATED` | 同房间内 nickname 重复 | 409 |
| `GROUP_NOT_FOUND` | 小组不存在或不属于该房间 | 404 |
| `GROUP_FULL` | 小组已满 | 409 |
| `STUDENT_NOT_FOUND` | 学生不存在 | 404 |
| `INVALID_TASK_TITLE` | 任务标题为空或非法 | 400 |
| `INVALID_DEADLINE` | deadlineAt 不晚于当前时间 | 400 |
| `INVALID_TARGET_TYPE` | targetType 非 all/groups | 400 |
| `INVALID_TARGET_GROUP` | 目标小组为空或不属于房间 | 400 |
| `TASK_NOT_FOUND` | 任务不存在 | 404 |
| `TASK_NOT_IN_STUDENT_ROOM` | 任务不属于学生所在房间 | 403 |
| `TASK_NOT_TARGETED_TO_STUDENT` | 学生不在任务目标范围内 | 403 |
| `TASK_PAUSED` | 任务已暂停，不能提交 | 409 |
| `TASK_CLOSED` | 任务已关闭，不能提交 | 409 |
| `TASK_DEADLINE_PASSED` | 已过截止时间，不能提交 | 409 |
| `SUBMISSION_ALREADY_EXISTS` | 学生已提交过该任务 | 409 |
| `SUBMISSION_NOT_FOUND` | 提交不存在 | 404 |
| `INVALID_SCORE` | 分数非整数或不在 1–10 | 400 |
| `GRADE_FAILED` | 评分写入失败 | 500 |
| `TOO_MANY_IMAGES` | 图片数量超过 3 张 | 400 |
| `IMAGE_TOO_LARGE` | 单图超过 5MB | 400 |
| `INVALID_IMAGE_TYPE` | 图片 MIME 不被允许 | 400 |
| `UPLOAD_FAILED` | 文件上传失败 | 500 |
| `INVALID_DISPLAY_MODE` | displayMode 非 anonymous/showGroup | 400 |
| `EXPORT_FAILED` | 导出失败 | 500 |
| `IMAGE_FILE_MISSING` | 导出时图片原件缺失 | 500 |

---

## 六、待确认问题

> 以下问题需团队确认，最终以团队决定为准。

1. **teacherToken 数据库中是否存 hash？**
   MVP 可先明文存储，但建议生产环境改为 hash 存储。

2. **大屏是否复用 teacherToken，还是单独生成 displayToken？**
   MVP 复用 `teacherToken`；后续可扩展只读的 `displayToken` 以降低凭证泄露风险。

3. **房间 ended 后学生是否还能查看自己的历史结果？**
   建议 MVP 中学生端提示"课堂已结束"，不再进入课堂操作；是否保留只读查看历史结果（`/api/student/me/results`）待定。本文档[接口 9](#9-学生获取自己的任务列表)默认采用"任务列表返回 `ROOM_ENDED`、引导改用结果接口只读"的策略。

4. **导出接口是否必须在课堂 ended 后才能使用？**
   建议 MVP 中课中、课后都可导出，最终以团队决定为准。

5. **图片本地存储还是对象存储？**
   MVP 可使用本地 `backend/uploads`，后续替换为 OSS / S3 / MinIO。

6. **teacherToken 如果丢失是否支持恢复？**
   MVP 不支持，后续讲师账号体系再解决。

7. **【实现差异提醒】统一响应格式字段名。**
   本文档约定为 `{ success, message, data, errorCode }`。
   但当前前端 `frontend/src/api/client.ts` 中的 `ApiResponse<T>` 定义为 `{ code, message, data }`（使用 `code` 而非 `success`）。
   联调前需统一二者：建议以本文档为准，后续调整前端 client 的类型定义（本阶段不修改代码）。

---

## Backend compatibility note for Prompt 9-11 contract fix

The standard contract routes are:

- `POST /api/teacher/submissions/:submissionId/grade`
- `GET /api/student/me/results`
- `GET /api/student/rooms/:roomCode/ranking`

Temporary compatibility routes are still supported for existing callers:

- `PATCH /api/teacher/submissions/:submissionId/grade`
- `GET /api/teacher/rooms/:roomCode/leaderboard`
- `GET /api/student/me/leaderboard`

These compatibility routes can be removed after all frontend and teammate branches migrate to the standard contract paths.
