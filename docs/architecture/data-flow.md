# 数据流

## 普通 Forward

浏览器提交规则到 `forwardCreate`；Handler 校验入口、端口、协议和 Landing，Repository
写 `forward`/`forward_port`，随后 Handler 构造 Service 配置并经 WebSocket 发到入口
Agent。Agent `handleAddService` 调用 `socket.createServices` 注册 TCP/UDP Listener。

## Tunnel Forward

Tunnel CRUD 保存 `tunnel`、`chain_tunnel` 和节点关系。Forward 引用真实 Tunnel ID；
Panel 先确保 Chain/Relay Service，再应用 Forward Service。Type 2 Relay 的逻辑业务可
承载 TCP/UDP，而 `tcp`、`tls`、`wss`、`mtcp`、`kcp` 等是外层传输选择。

## Multiport

```text
IXPathProfile
   -> ixRouteCreate
   -> AllocateIXPorts(entry lane + IX lane, per protocol)
   -> IXRoute + IXRouteRuntime desired rows
   -> Agent AddService acknowledgements
   -> Observation / ObservedState=active
```

编辑先写 `IXRouteMigration` journal，申请新 Lease 并应用新 Runtime，确认后释放旧 Lease；
失败由 `rollbackIXRouteEdit` 或后台 Reconciler 恢复。删除只释放该 Route 所有的 Service
和 Lease。

## 流量与诊断

GOST Service 计数 -> Agent durable outbox -> `/api/v1/flow/upload` -> Receipt 去重 ->
Forward/IXRoute 累计值和 metrics。诊断由 Panel 编排节点命令，结果保存测试方法、证据
级别和状态；UDP 只有 payload 响应才是 PASS。
