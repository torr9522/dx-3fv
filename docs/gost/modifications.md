# GOST 修改边界

DX-3FV/FLVX 在 GOST 上增加 Panel WebSocket 命令层、配置持久化、Service/Chain 原子替换、
暂停恢复、节点诊断、Agent 升级、全局流量统计和 durable outbox。核心文件包括：

- `go-gost/x/socket/websocket_reporter.go`
- `go-gost/x/socket/service.go`、`chain.go`、`limiter.go`
- `go-gost/x/service/global_traffic_manager.go`
- `go-gost/x/service/traffic_outbox.go`
- Relay/Tunnel 安全与恢复测试位于 `go-gost/x/chain/*relay*_test.go`

具体差异应使用 `docs/history/source-provenance.md` 记录的 upstream Commit 获取独立基线
后做 diff；不要批量同步官方 go-gost/x，因为会改变已验证的数据平面。
