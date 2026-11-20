# DX3FV-0034: Forward 协议缩减后残留旧 Listener

## 问题与证据

通过 v0.1.33 的真实单端口页面创建 `TCP+UDP` 规则后，将协议编辑为 `TCP`。
数据库和页面均显示 TCP，但 GZ2 上 `ss` 仍显示同一端口的 UDP Listener。
Agent 日志证明编辑仅收到包含 TCP Service 的 `UpdateService`，没有收到旧 UDP Service
的删除命令。

## 根因

Agent 的 `UpdateService` 只更新 payload 中具名的 Service，不会推断或删除同一 Forward
下已不再出现的协议 Service。Panel 在协议集合缩小时直接按新集合下发，因此旧协议实例
继续运行。

## 修复

Forward 编辑检测协议集合变化。对于新旧拓扑共同拥有的入口节点，先使用既有 Service
family 清理逻辑删除 `base`、`base_tcp`、`base_udp` 全部候选，再下发新协议集合。
节点迁移的原有清理路径保持不变。

## 自动化测试

`TestForwardUpdateProtocolRemovesOldServiceFamily` 使用真实 WebSocket 命令传输测试桩：

- 建立 `tcp+udp` 直连 Forward。
- 经 HTTP handler 编辑为 `tcp`。
- 验证先发送完整 family `DeleteService`。
- 验证替换 payload 只包含 TCP Service。

## 真人与 Runtime 验收

修复部署后需再次通过浏览器从 TCP+UDP 编辑为 TCP，并在 GZ2 验证：

- TCP Listener 保留。
- UDP Listener 数量为 0。
- Shadowsocks TCP payload 仍成功。
- 其他单端口、Tunnel Forward 和多端口业务不受影响。

## 修改文件

- `go-backend/internal/http/handler/mutations.go`
- `go-backend/internal/http/handler/forward_protocol_update_test.go`

## 回滚

切回上一 Release 并恢复升级前数据库备份。回滚代码会重新暴露协议缩减残留问题；若已
发生，需通过正常删除/重建该 Forward 让 Agent 清理完整 Service family。
