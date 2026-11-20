# Panel / Agent 协议

Agent 使用节点 Secret 认证 WebSocket，Panel 会话实现位于
`go-backend/internal/ws/server.go`，Agent 实现位于
`go-gost/x/socket/websocket_reporter.go`。命令 envelope 包含 `type`、`data` 和可选
`requestId`；响应返回同一 `requestId`、`success`、`message`、`data`。

常用变更命令包括 Add/Update/Delete/Pause/Resume Service、Chain、Limiter 和 Agent
upgrade/rollback；诊断命令包括 TCP、UDP payload echo 和 monitor check。变更命令串行，
只读命令有并发上限。新增命令必须保持旧 Agent 可识别的失败语义，并先做协议兼容审计。

流量上传是独立 HTTPS 路径，payload 用节点 Secret 派生的加密机制，带可选
`X-DX3FV-Flow-Report-ID` 实现持久幂等确认。
