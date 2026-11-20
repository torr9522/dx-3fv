# 模块依赖

```text
React/Vite
   | raw JWT Authorization
   v
Go Panel HTTP handlers ---> Repository ---> SQLite / PostgreSQL
   |                         ^
   | requestId WebSocket     | Receipt / Observation / Desired State
   v                         |
Agent WebSocket reporter ---> embedded GOST registries ---> TCP/UDP sockets
                                      |
                                      +-> durable traffic outbox
```

Panel 依赖 Agent 命令确认，但数据库写入和 Runtime 应用不是同一事实。IXRoute 使用
`DesiredState`、`ObservedState`、`IXRouteRuntime.Applied` 和 Lease 状态表达差异；
重试入口在 `handler/jobs.go`。

模块边界：Frontend 不生成 GOST 配置；Repository 不发 WebSocket 命令；Agent 不直接
写 Panel 数据库；GOST fork 不掌握用户、PathProfile 或 Lease。跨边界行为应通过 API、
持久化状态或带 `requestId` 的命令完成。
