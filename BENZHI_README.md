# 多语言术语一致性校验服务

## 离线自检

```bash
GOTOOLCHAIN=local go run ./cmd/terminology --smoke-test
```

自检会在临时 SQLite 数据库中创建并发布术语库，导入文档、扫描禁用词并生成一致性建议。

## HTTP 服务

```bash
go run ./cmd/terminology --db terminology.db --addr :8080
curl http://127.0.0.1:8080/healthz
```

服务重启后继续使用同一个 SQLite 文件；启动阶段会把持久化的 running 检查任务恢复为 pending。

## 容器验证

```bash
bash build_benzhi_docker.sh terminology linux/amd64
docker run --rm terminology:latest bash -lc 'go run ./cmd/terminology --smoke-test'
```

镜像构建阶段执行 `go build ./...`，不会携带业务数据库或外部服务凭据。
