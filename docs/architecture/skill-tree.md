# DX-3FV 技能树与源码导航

## Panel 启动与 API

`go-backend/cmd/paneld/main.go:main` 读取 `internal/config.FromEnv`，调用
`internal/app.New`。`app.New` 通过 `store/repo.Open` 或 `OpenPostgres` 打开数据库，
构造 `http/handler.New`、`http.NewRouter` 和 `http.Server`；`App.Run` 在监听前启动
`Handler.StartBackgroundJobs`。

API 入口集中在 `internal/http/handler/handler.go:Handler.Register`。认证边界在
`internal/http/router.go:NewRouter` 与 `internal/http/middleware/`。业务 Handler 不应
直接散落 SQL；持久化集中在 `internal/store/repo/`，表模型在
`internal/store/model/model.go`。

```text
paneld.main
  -> config.FromEnv
  -> app.New
     -> repo.Open/OpenPostgres -> autoMigrateAll -> migrateSchema
     -> handler.New -> ws.NewServer / metrics / health
     -> http.NewRouter -> Handler.Register
  -> Handler.StartBackgroundJobs
  -> http.Server.ListenAndServe
```

## Forward 与 Tunnel

- CRUD：`handler/mutations.go` 的 `forwardCreate`、`forwardUpdate`、`tunnelCreate`、
  `tunnelUpdate`。
- Runtime 配置：同目录 `control_plane.go` 的 `buildForwardServiceConfigs` 及 Tunnel
  Service/Chain 构造函数。
- 生命周期互斥：`handler/tunnel_lifecycle.go:lockTunnelLifecycle`。
- 清理恢复：`tunnel_runtime_cleanup.go`、`jobs.go:runFederationCleanupRetryLoop`。
- 单端口规则仍使用 `model.Forward`，`TunnelID=0`；隧道转发使用同一 Forward 模型和
  真实 Tunnel，不创建 synthetic Tunnel。

## Multiport / IX

- API：`handler/handler.go` 中 `/api/v1/ix-port/*`、`/ix-path/*`、`/ix-route/*`。
- 创建与编辑：`handler/ix_routes.go:ixPathCreate`、`ixRouteCreate` 及迁移 journal。
- 端口分配：`store/repo/repository_ix_ports.go:AllocateIXPorts`。
- 恢复：`handler/ix_routes.go:reconcileIXRoute` 与 `jobs.go:retryPendingIXRouteOperations`。
- 诊断：`handler/ix_route_diagnosis.go:ixRouteDiagnose`。
- 模型：`IXPortPool`、`IXPortResource`、`IXPortLease`、`IXPortObservation`、
  `IXPathProfile`、`IXRoute`、`IXRouteRuntime`、`IXRouteMigration`。

## Agent 与 GOST

Panel 的 `internal/ws/server.go:Server` 用节点 Secret 建立会话并以 `requestId` 关联命令
响应。Agent 的 `go-gost/x/socket/websocket_reporter.go:WebSocketReporter` 连接 Panel，
`dispatchCommand` 将变更命令放入串行队列，`routeCommand` 调用 Service/Chain/Limiter
处理器。`socket/service.go` 通过 GOST registry 原子增加、替换、暂停或删除 Service。

流量由 `go-gost/x/service/global_traffic_manager.go` 汇总；
`traffic_outbox.go` 先以 `0600` 原子落盘，再发送带 Report ID 的批次。Panel
`handler.go:flowUpload` 与 `repository_flow_receipts.go` 幂等接收，避免重启重传双计。

## Frontend 与原生发布

`vite-frontend/src/App.tsx` 定义路由；`layouts/admin.tsx` 定义权限和导航；
`pages/single-forward`、`pages/forward`、`pages/multiport.tsx` 分别对应三种转发。
API 客户端位于 `src/api/`。

`scripts/build-native-release.sh` 是唯一正式原生构建入口，校验 clean worktree 和
Annotated Tag，生成 Panel/Agent、静态前端、systemd、Caddy 示例、安装/升级/回滚、
Manifest、SHA256、SBOM、许可证与同 Commit 源码包。
