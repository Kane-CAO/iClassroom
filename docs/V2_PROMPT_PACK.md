# iClassroom V2.0 Prompt Pack

本文档提供给 Claude / Codex / Cursor 使用。每次只使用一个 Prompt，不要一次性投喂完整 Prompt Pack。

---

## 通用 Prompt 模板

```text
你正在 iClassroom V2.0 项目中工作。

请先阅读：
1. README.md
2. Claude.md
3. docs/api.md
4. docs/database.md
5. docs/V2_ENGINEERING_PLAN.md
6. 当前任务相关代码

本次任务：
【填写任务名称】

目标：
- 【目标 1】
- 【目标 2】

允许修改：
- 【文件或目录】

禁止修改：
- 【文件或目录】

必须遵守：
- 不提交 .env 或密钥
- 不删除已有业务校验
- 不跨模块重构
- handler 不写复杂业务逻辑
- service 负责业务规则
- repository 负责数据库读写
- 前端 API 调用必须集中在 api/ 目录

完成后请输出：
1. 修改文件清单
2. 新增接口 / 页面 / 表
3. 测试命令和结果
4. 手动验收步骤
5. 风险和后续建议
```

---

## Prompt 0：V2.0 项目审计

```text
请审计当前项目与 V2.0 文档的差距。

重点检查：
- 是否仍依赖 teacherToken
- 是否缺少 admin / teacher 账号表
- 是否缺少登录接口
- 是否缺少统一附件模型
- 是否缺少文件上传
- 是否缺少管理者端页面
- 是否缺少 AITIC 视觉体系

只输出差距清单和建议分支，不要改代码。
```

---

## Prompt 1：数据库模型改造

```text
请实现 V2.0 数据库 migration。

目标：
- 新增 admin_users
- 新增 teacher_accounts
- 新增 auth_sessions 或说明 JWT 替代方案
- rooms 新增 teacher_id
- 新增 submission_attachments
- 保留旧 submission_images 和 teacher_token 兼容

允许修改：
- backend/migrations/
- docs/database.md
- backend/internal/domain/ 中相关实体

禁止修改：
- 前端
- 上传业务
- 导出业务

完成后运行 go test ./...。
```

---

## Prompt 2：账号与权限基础架构

```text
请实现管理者 / 讲师登录基础架构。

目标：
- 密码哈希
- token 生成和校验
- admin 鉴权中间件
- teacher 鉴权中间件
- /api/auth/admin/login
- /api/auth/teacher/login
- /api/auth/logout
- /api/auth/me

不要实现讲师账号管理页面。
不要改观众端接口。
```

---

## Prompt 3：管理者讲师账号管理

```text
请实现管理者端讲师账号管理 API。

目标：
- 创建讲师
- 查询讲师列表
- 停用 / 启用讲师
- 重置密码
- 删除讲师或软删除

权限：
- 只有 admin 可访问
- teacher 不可访问

完成后补 handler/service 测试。
```

---

## Prompt 4：讲师登录与房间归属迁移

```text
请把讲师房间接口迁移到登录态。

目标：
- 创建房间时写入 teacher_id
- 讲师只能查看自己的房间
- 讲师只能操作自己的任务、提交、看板、导出
- 旧 teacherToken 只保留兼容，不作为主鉴权

重点：
- 不要破坏观众通过 roomCode 入场
- 不要让 roomCode 成为老师权限
```

---

## Prompt 5：观众入场与匿名会话

```text
请实现或校正 V2.0 观众入场。

目标：
- 获取房间信息
- 昵称验重
- 小组满员校验
- 加入房间返回 clientToken
- resume 恢复会话
- ended 后不能 join/resume/submit

观众仍然无账号。
```

---

## Prompt 6：任务发布与截止控制

```text
请实现 V2.0 任务发布和状态控制。

目标：
- 讲师发布任务
- 支持全场 / 指定小组
- 暂停任务
- 关闭任务
- 截止时间后禁止提交
- 观众只能看到自己目标范围内任务
```

---

## Prompt 7：文字 / 图片 / 文件提交

```text
请实现统一提交和附件上传。

目标：
- multipart 提交 contentText、images[]、files[]
- 图片单图 <= 5MB，最多 3 张
- 文件不允许视频
- 文件类型和大小有白名单
- 保存 submission_attachments
- 上传失败不创建不完整提交

不要实现评分和导出。
```

---

## Prompt 8：讲师审阅与评分

```text
请实现讲师审阅和评分。

目标：
- 按任务查看提交
- 返回文字、图片、文件
- 评分 1-10
- 可写评语
- 小组分数按差值事务更新
- 学生可查看自己的结果
```

---

## Prompt 9：排行榜、大屏、精选答案

```text
请实现排行榜、大屏数据和精选答案。

目标：
- 小组排行榜
- 当前任务完成度
- 精选答案
- 匿名 / 显示小组
- 大屏接口只允许房间所属讲师访问
```

---

## Prompt 10：后台数据看板

```text
请实现讲师后台数据看板。

MVP 指标：
- 整体参与率
- 各小组得分对比
- 任务完成情况
- 提交时间分布

不要做跨活动历史趋势。
不要做个人成长曲线。
```

---

## Prompt 11：活动结束与导出

```text
请实现活动结束、归档和导出。

目标：
- 结束活动后观众不能 join/resume/submit
- 讲师仍可查看和导出
- 导出 zip 包含 submissions.xlsx、images/、files/
- Excel 包含房间、任务、观众、小组、提交、附件、评分、评语、时间
```

---

## Prompt 12：前端管理者端

```text
请实现前端管理者端。

目标：
- /admin/login
- /admin/teachers
- 创建讲师
- 停用 / 启用讲师
- 重置密码
- 删除讲师
- admin 路由保护

设计：
- 电脑端后台工具
- 信息密度清晰
```

---

## Prompt 13：前端讲师端账号化改造

```text
请把讲师端从 teacherToken 模式迁移到登录态。

目标：
- /teacher/login
- teacher session
- teacher 路由保护
- 创建房间使用 Authorization
- 讲师工作台只显示自己的活动
```

---

## Prompt 14：前端观众端上传与移动端体验

```text
请实现观众端 V2.0 答题体验。

目标：
- 手机优先
- 文字草稿自动保存
- 图片上传
- 文件上传
- 上传失败提示和重试
- 已提交 / 已评分状态
- 我的得分和排名
```

---

## Prompt 15：AITIC 视觉体系改造

```text
请按 AITIC 视觉要求调整前端。

目标：
- 阿里橙主色
- 深墨文字和导航
- 暖中性背景
- 管理者 / 讲师端偏效率工具
- 观众端手机优先
- 大屏端适合现场展示

不要改业务逻辑。
```

---

## Prompt 16：E2E 测试与验收

```text
请补充 V2.0 E2E 和手动验收清单。

必须覆盖：
- 管理者创建讲师
- 讲师登录
- 讲师建房
- 观众入场
- 讲师发任务
- 观众提交文字 / 图片 / 文件
- 讲师评分
- 大屏更新
- 后台数据更新
- 结束活动
- 导出
```
