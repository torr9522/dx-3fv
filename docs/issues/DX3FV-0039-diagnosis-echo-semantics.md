# DX3FV-0039：端到端诊断 Echo 误报

- 基线：`v0.1.42-disable-tunnel-load-balancing-ui`
- 数据库变更：无
- 影响：多端口链路诊断证据语义，不改变 Runtime 转发

## 复现与根因

2026-11-20 测试 Panel 的 `ix_route_diagnosis.id=21`显示：入口 TCP 与出口 TCP Connect 均 PASS，而 TCP/UDP nonce Echo 均 timeout。同路径真实 Shadowsocks TCP/UDP payload PASS。

`[::1]:37712` 不是 Landing。`37712` 是当前业务的 GZ2 入口端口；`[::1]` 由 `ixLocalProbeHost` 把节点通配监听 `[::]` 投影为 Agent 本机 loopback。GZ2 Agent 连接本地 GOST Listener 后，Runtime 仍正常经 IX 转发到 Landing。

`EchoProbe` 会向目标写入随机 nonce 并要求返回完全相同字节。现有 Landing 是 Shadowsocks 服务，不是 raw TCP/UDP Echo Server，因此 read timeout 只证明“没有 Echo 响应”，不证明转发失败。

另外，旧 `udp_nonce_echo_via_ix_route` 从 GZ2 向 IX 业务端口发送 nonce。IX Listener 会继续把该 payload 转发到 Landing，所以它测量的是完整路径，不能标记为独立 GZ→IX UDP 证据。

## 修复

- Agent 命令派发失败仍为 FAIL，因为诊断未能执行。
- 真实 TCP Connect 失败仍为 FAIL。
- EchoProbe 完成调用但 nonce 未返回时改为 UNVERIFIED，明确说明 Echo 能力缺失不等于业务失败。
- UDP GZ→IX 在没有 IX 分段 Echo 响应器时固定为 UNVERIFIED，不再用完整业务路径冒充分段测量。
- 成功返回精确 nonce 时仍为 PASS。

## 测试

- Handler 覆盖 TCP Connect PASS + Echo timeout 时整体 UNVERIFIED。
- Handler 覆盖受控 Echo 成功及 UDP 分段证据边界。
- 真实浏览器在前海/QIX/盲盒业务执行诊断，并与 Shadowsocks payload 结果对照。

## 回滚

切回 `v0.1.42-disable-tunnel-load-balancing-ui`。本 Issue 无 Migration、无 Runtime 配置变更。
