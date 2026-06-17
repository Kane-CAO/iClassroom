# iClassroom
前端和后端，所有的内容都可以同步到这里，大家在个人本地修改完代码记得上传
新人从这里下入本地

## Docker 本地一键运行

这套 Docker 配置用于内部演示和团队本地统一运行环境，会同时启动：

- `frontend`：React + Vite 前端，访问地址 `http://localhost:3000`
- `backend`：Go + Gin 后端，健康检查 `http://localhost:8080/health`
- `mysql`：MySQL 8.4，首次启动会自动执行 `backend/migrations/000001_init_schema.up.sql`

### 启动

在项目根目录执行：

```bash
docker compose up --build
```

启动完成后打开：

```text
http://localhost:3000
```

### 后台运行

```bash
docker compose up --build -d
```

查看日志：

```bash
docker compose logs -f
```

停止服务：

```bash
docker compose down
```

### 数据库

Docker 内部后端连接 MySQL 服务名 `mysql:3306`。

如果需要从宿主机连接数据库，使用：

```text
Host: 127.0.0.1
Port: 3307
Database: iclassroom
User: iclassroom
Password: iclassroom_password
```

数据会保存在 Docker volume `mysql_data` 中，普通 `docker compose down` 不会删除数据。

如果需要清空本地演示数据并重新建表：

```bash
docker compose down -v
docker compose up --build
```

### 上传目录

后端上传文件会保存在 Docker volume `backend_uploads` 中。

### 注意

这套配置是本地演示版，解决的是“一键启动”和“团队环境一致”。真实二维码扫码、公网访问、HTTPS、云数据库和 OSS 文件存储，仍需要后续部署到云服务器并配置正式域名。
