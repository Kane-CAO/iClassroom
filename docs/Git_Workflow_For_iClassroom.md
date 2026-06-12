# iClassroom Git 分支与 PR 工作流说明

> 适用对象：iClassroom 项目全体成员
> 目的：让每个人知道什么时候拉取最新代码、什么时候新建分支、什么时候提交 PR，以及如果暂时不开发，如何同步查看最新进度。

------

## 1. 为什么要规范 Git 流程

我们现在是多人协作开发，后端和前端都会不断更新。如果每个人都直接在旧分支上长期开发，容易出现几个问题：

```text
1. PR 变得很大，review 时看不清楚每次到底做了什么
2. 之前已经完成的内容和新内容混在一起，难以回溯
3. main / dev 更新后，本地分支容易落后，后续合并容易冲突
4. 不同成员修改同一类文件时，容易互相覆盖
5. 出问题时，不容易定位是哪一个阶段引入的问题
```

所以推荐：

```text
一个阶段 / 一个功能 / 一个 Prompt 对应一个独立分支和一个 PR
```

这样 PR 更清晰，也方便回溯每个阶段的代码变化。

------

## 2. 推荐流程：每个阶段新建一个分支

适合对象：

```text
对 Git 比较熟悉的成员
负责模块较多的成员
需要连续完成多个 Prompt 的成员
```

例如：

```text
Prompt 5：feature/backend-step0-foundation
Prompt 6：feature/backend-step1-database
Prompt 7：feature/backend-room-student
Prompt 9：feature/backend-task
Prompt 10：feature/backend-submission
```

### 2.1 开始新阶段前

每次开始一个新任务前，先回到主分支并拉取最新代码。

团队以 `main` 为基准：

```bash
git checkout main
git pull origin main
```

然后新建自己的功能分支：

```bash
git checkout -b feature/your-name-task-name
```

例如：

```bash
git checkout -b feature/lusitao-backend-step1
```

### 2.2 开发完成后

先查看改动：

```bash
git status
```

提交代码：

```bash
git add .
git commit -m "feat(backend): add database migrations and domain models"
```

推送到远程：

```bash
git push origin feature/your-name-task-name
```

然后在 GitHub 上创建 PR：

```text
base: main 或 dev
compare: feature/your-name-task-name
```

### 2.3 PR 合并后

等负责人 review 并 merge 后，本地不要继续直接在旧分支上做新任务。

应该重新回到 `main` 或 `dev`，拉取最新代码，再开新分支：

```bash
git checkout main
git pull origin main
git checkout -b feature/your-name-next-task
```

如果团队使用 `dev`：

```bash
git checkout dev
git pull origin dev
git checkout -b feature/your-name-next-task
```

### 2.4 推荐原因

这种方式的好处是：

```text
1. 每个 PR 只对应一个清晰阶段
2. review 更容易
3. 出问题可以快速定位是哪次 PR 引入的
4. main / dev 始终是最新基准
5. 每个人的工作边界更清楚
6. 后续冲突更少
```

------

## 3. 简化流程：一直使用自己的同一个分支

适合对象：

```text
Git 不太熟悉的成员
只负责一个小模块的成员
当前阶段主要做少量修改的成员
```

例如一直使用：

```text
feature/hans-backend
feature/xuan-backend
feature/upload-export
```

这种方式也可以，但需要注意：**每次开始开发前，都要把 main / dev 的最新代码同步到自己的分支。**

### 3.1 进入自己的分支

```bash
git checkout feature/your-branch-name
```

例如：

```bash
git checkout feature/hans-backend
```

### 3.2 拉取远程最新信息

```bash
git fetch origin
```

### 3.3 把 main 的最新代码同步进自己的分支

如果团队以 `main` 为基准：

```bash
git pull --rebase origin main
```

如果团队以 `dev` 为基准：

```bash
git pull --rebase origin dev
```

### 3.4 开始开发并提交

```bash
git status
git add .
git commit -m "feat(backend): add task publish API"
git push origin feature/your-branch-name
```

如果提示分支已经分叉，可以先执行：

```bash
git pull --rebase origin feature/your-branch-name
git push origin feature/your-branch-name
```

### 3.5 这种方式的注意点

一直使用同一个分支的优点：

```text
1. 操作简单
2. 不需要频繁创建新分支
3. 对 Git 不熟悉的成员更容易执行
```

缺点是：

```text
1. PR 可能越来越大
2. 容易把多个阶段混在一起
3. review 时不如每阶段新分支清楚
4. 如果分支长期不合并，后期冲突可能更多
```

所以如果使用同一个分支，建议每完成一个明确功能就尽快提交 PR，不要等所有东西都做完再提交。

------

## 4. 如果暂时不开发，只想同步最新 main / dev 观察进度

有些成员当前阶段可能暂时没有任务，但想拉取最新代码看项目进展。

### 4.1 只看最新 main

如果只是想看主分支最新状态：

```bash
git checkout main
git pull origin main
```

然后可以查看项目文件、运行项目、阅读代码。

### 4.2 只看最新 dev

如果团队主要在 `dev` 上集成：

```bash
git checkout dev
git pull origin dev
```

### 4.3 自己的分支也想同步最新 main

如果你有自己的分支，但暂时不开发，只是想让自己的分支跟上最新进度，可以这样：

```bash
git checkout feature/your-branch-name
git fetch origin
git pull --rebase origin main
```

如果团队使用 `dev`：

```bash
git checkout feature/your-branch-name
git fetch origin
git pull --rebase origin dev
```

然后推送同步后的分支：

```bash
git push origin feature/your-branch-name
```

如果 Git 提示需要强制推送，先不要直接操作，可以在群里问一下。一般自己独立使用的分支可以用：

```bash
git push --force-with-lease origin feature/your-branch-name
```

但多人共用的分支不要随便 force push。

------

## 5. 每次开发前的安全检查

开始写代码前建议执行：

```bash
git status
```

确认没有未提交改动。

如果有未提交改动，不要直接切分支。先判断：

```text
1. 这些改动是否需要提交
2. 是否只是临时文件
3. 是否需要 stash 暂存
```

常用命令：

```bash
git status
git branch
git log --oneline --decorate --graph --all -10
```

------

## 6. 每次提交前的检查

提交前建议执行：

```bash
git status
```

后端成员执行：

```bash
cd backend
go test ./...
```

前端成员执行：

```bash
cd frontend
npm run build
```

确认没有问题后再提交：

```bash
git add .
git commit -m "feat(scope): short description"
```

------

## 7. Commit Message 建议

格式建议：

```text
type(scope): description
```

常见类型：

```text
feat: 新功能
fix: 修复问题
docs: 文档修改
test: 测试相关
refactor: 重构
chore: 工程配置或杂项
```

示例：

```bash
git commit -m "feat(backend): add health check endpoint"
git commit -m "feat(backend): add room creation API"
git commit -m "feat(frontend): add student join page"
git commit -m "docs: update API contract"
git commit -m "fix(backend): validate duplicate nickname"
```

------

## 8. PR 描述必须写清楚

每个 PR 至少包含：

```text
1. 本次实现了什么
2. 修改了哪些文件或模块
3. 新增了哪些接口或页面
4. 如何运行
5. 如何测试
6. 是否有未完成 TODO
7. 是否影响其他成员
```

示例：

~~~md
## 本次实现

完成 Backend Step 0：后端基础服务搭建。

## 修改内容

- 添加 Gin server 启动入口
- 添加 config 读取 .env
- 添加 MySQL 连接池
- 添加统一 response
- 添加 CORS middleware
- 添加 GET /health

## 测试方式

```bash
cd backend
go test ./...
go run ./cmd/server
curl http://localhost:8080/health
~~~

## 注意事项

- 本次没有实现业务接口
- .env 不应提交
- 后续 Prompt 6 会继续补数据库 migration 和 domain model

```
---

## 9. 推荐团队执行方式

### 对比较熟悉 Git 的成员

推荐：

```text
每个阶段新建一个分支
每个阶段一个 PR
PR merge 后，从最新 main / dev 重新开下一个分支
```

流程：

```bash
git checkout main
git pull origin main
git checkout -b feature/your-name-task
# 开发
git add .
git commit -m "feat(scope): description"
git push origin feature/your-name-task
# GitHub 创建 PR
```

### 对不太熟悉 Git 的成员

可以先使用：

```text
一直使用自己的固定分支
每次开发前同步 main / dev
每完成一小块功能就提交 PR
```

流程：

```bash
git checkout feature/your-branch-name
git fetch origin
git pull --rebase origin main
# 开发
git add .
git commit -m "feat(scope): description"
git push origin feature/your-branch-name
```

------

## 10. 最重要的规则

```text
1. 不要直接 push 到 main
2. 不要提交 .env、密码、token、密钥
3. 开发前先同步 main / dev
4. 每次 PR 尽量只做一个明确功能
5. 遇到 conflict 不要乱删，先发到群里确认
6. 不熟悉 Git 的成员可以继续使用固定分支，但要经常同步 main / dev
7. 熟悉 Git 的成员建议每个阶段新开分支，保持 PR 清晰和可回溯
```