# GOST 修改边界

DX-3FV/FLVX 在 GOST 上增加 Panel WebSocket 命令层、配置持久化、Service/Chain 原子替换、
暂停恢复、节点诊断、Agent 升级、全局流量统计和 durable outbox。核心文件包括：

- `go-gost/x/socket/websocket_reporter.go`
- `go-gost/x/socket/service.go`、`chain.go`、`limiter.go`
- `go-gost/x/service/global_traffic_manager.go`
- `go-gost/x/service/traffic_outbox.go`
- Relay/Tunnel 安全与恢复测试位于 `go-gost/x/chain/*relay*_test.go`

## UDP session lifecycle

DX3FV-0058 修复 `internal/net/udp` 中 session 关闭、idle expiry 和 datagram 入队之间的
竞态：conn 生命周期操作串行化，pool 以 compare-and-swap 发布 replacement，并以
compare-and-delete 防止旧 session 清理误删新值。修复不改变 TTL 或通用 timeout。

对应回归测试位于 `go-gost/x/internal/net/udp/listener_lifecycle_test.go`。上游对照与真实
数据面证据见 `docs/issues/DX3FV-0058-udp-session-expiry-race.md`。同步新上游版本时不得
丢失这些生命周期不变量。

具体差异应使用 `docs/history/source-provenance.md` 记录的 upstream Commit 获取独立基线
后做 diff；不要批量同步官方 go-gost/x，因为会改变已验证的数据平面。
