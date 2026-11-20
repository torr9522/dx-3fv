# DX3FV-0047：Relay cleanup contract 证据窗口竞态

- 优先级：P0（RC 测试阻断）
- 基线：`v0.1.50-release-candidate-assets`
- 产品代码：不变
- Schema：不变（Schema 12）

## 复现

完整 backend suite 曾使
`TestRedesiredRelayKeepsOldPortReservedUntilReplacementAcknowledged` 失败，报告重新
声明的 Relay 收到 `DeleteService`。隔离测试和 race detector 通过；相关测试在并发
节点上线 reconciliation 下可重复触发。

## 根因

测试在发起第二次 Tunnel update **之前**清空命令记录，然后释放 recorder mutex。
节点上线 reconciliation 可以在 update 取得 Tunnel lifecycle lock 之前处理仍然处于
“middle 已移除”的旧 Desired State，并合法尝试删除旧 listener。测试随后成功重新
加入 middle，却把更新前已记录的命令当成更新后的延迟误删。

命令 payload 证据确认目标是 tunnel 1；源码调用链确认 cleanup 与 update 使用同一
Tunnel lifecycle lock，不存在持久化新 Desired State 之后绕过锁的删除。

## 修复

第二次 update 返回、即 replacement Desired State 已提交之后才清空命令证据。随后
继续验证：旧端口在 replacement ACK 前被保留，ACK 后可复用，且之后不能再删除
重新声明的 Relay。失败信息保留完整 DeleteService payload，便于未来定位。

没有增加 sleep、扩大 timeout、删除断言或修改任何 Runtime 行为。

## 验证与回滚

- 相关三测试在 `GOMAXPROCS=1/2` 下高次数运行。
- contract suite 连续运行与 backend 全套测试。
- `-race` 单项验证。

回滚只恢复旧测试证据窗口；不影响部署和数据，但会重新引入 suite-scope 假失败。
