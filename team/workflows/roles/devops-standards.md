# DevOps 运维输出规范

> **角色**: DevOps
> **加载文件**: `{TEAM_PATH}/workflows/roles/devops-standards.md`
> **前置**: 先读 `{TEAM_PATH}/workflows/shared.md` 了解通用原则

---

## 0. 文档路径强制检查（最高优先级！）

**⚠️ 创建任何文档时，必须先检查路径！**

**内部文档 → `{docs_internal}`**
- 部署文档、运维文档、监控报告、技术报告

**对外文档 → `{docs_external}`**
- API 文档、用户手册

**必须输出**：
```
📂 文档路径检查:
   - 文档类型: [内部/对外]
   - 存放位置: [{docs_internal}/{docs_external}]
   - 路径: xxx/xxx.md
```

---

## 1. 触发条件

- 用户说：**"部署 xxx"**、**"运维 xxx"**、**"Docker xxx"**、**"K8s xxx"**、**"监控 xxx"**
- 例如：部署到生产环境、K8s 集群配置、Docker 镜像构建

---

## 2. 部署流程规范

### 2.1 部署前检查

| 检查项 | 说明 | 必须通过 |
|--------|------|---------|
| 代码冻结 | 所有 PR 已合并 | ✅ |
| 测试通过 | CI Pipeline 全部 Job 通过 | ✅ |
| 文档更新 | CHANGELOG、部署文档已更新 | ✅ |
| 配置检查 | 环境变量、配置文件已确认 | ✅ |
| 回滚方案 | 回滚脚本和方案已准备 | ✅ |

### 2.2 Docker 镜像构建

```dockerfile
# Multi-stage build
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-w -s' -o service ./cmd/

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/service .
EXPOSE 8080
ENTRYPOINT ["./service"]
```

### 2.3 K8s 部署配置

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {service-name}
  namespace: {namespace}
spec:
  replicas: 3
  selector:
    matchLabels:
      app: {service-name}
  template:
    metadata:
      labels:
        app: {service-name}
    spec:
      containers:
        - name: service
          image: {registry}/{service-name}:{version}
          ports:
            - containerPort: 8080
          resources:
            requests:
              memory: "256Mi"
              cpu: "250m"
            limits:
              memory: "512Mi"
              cpu: "500m"
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 10
            periodSeconds: 30
          readinessProbe:
            httpGet:
              path: /ready
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 10
```

---

## 3. CI/CD 流水线规范

### 3.1 Pipeline 阶段

| 阶段 | Job | 说明 |
|------|-----|------|
| Build | compile | 代码编译 |
| Test | unit-test | 单元测试 |
| Test | integration-test | 集成测试 |
| Security | scan | 安全扫描 |
| Image | docker-build | Docker 镜像构建 |
| Deploy | deploy-dev | 部署到 Dev |
| Deploy | deploy-staging | 部署到 Staging |
| Deploy | deploy-prod | 部署到 Prod（需审批） |

### 3.2 质量门禁

| 检查项 | 标准 | 阻塞发布 |
|--------|------|---------|
| 单元测试覆盖率 | ≥ 80% | ❌ 失败 |
| 安全扫描 | 无高危漏洞 | ❌ 失败 |
| Docker 镜像构建 | 成功 | ❌ 失败 |
| 部署到 Dev | 烟雾测试通过 | ❌ 失败 |

---

## 4. 监控与告警规范

### 4.1 监控指标

| 指标类型 | 指标名称 | 告警阈值 |
|---------|---------|---------|
| 可用性 | uptime | < 99.9% |
| 延迟 | P95 response time | > 500ms |
| 延迟 | P99 response time | > 1s |
| 错误率 | error rate | > 1% |
| 资源 | CPU usage | > 80% |
| 资源 | Memory usage | > 85% |
| 资源 | Disk usage | > 90% |

### 4.2 告警级别

| 级别 | 说明 | 处理时效 |
|------|------|---------|
| P0 | 服务不可用 | 立即处理 |
| P1 | 核心功能受损 | 15 分钟内 |
| P2 | 非核心功能异常 | 1 小时内 |
| P3 | 监控告警 | 下个工作日 |

---

## 5. 回滚规范

### 5.1 回滚触发条件

- 部署后功能测试失败
- 监控系统检测到异常
- 用户反馈严重问题

### 5.2 回滚流程

```
1. 确认回滚范围
2. 执行回滚命令
3. 验证服务恢复
4. 通知相关方
5. 记录事件
6. 分析根因
```

### 5.3 回滚检查清单

- [ ] 服务健康检查通过
- [ ] 核心功能验证通过
- [ ] 日志无异常
- [ ] 监控指标正常

---

## 6. 产出物清单

| 成果类型 | 文件路径 |
|---------|---------|
| 部署配置 | `deploy/docker/` |
| K8s 配置 | `deploy/k8s/` |
| 部署报告 | `{docs_internal}/release/v{version}/DEPLOY.md` |
| 回滚记录 | `{docs_internal}/release/v{version}/ROLLBACK.md` |
| 监控配置 | `deploy/monitoring/` |

---

## 7. 版本历史

| 版本 | 日期 | 说明 |
|------|------|------|
| v1.0 | 2026-04-17 | 初始版本 |
