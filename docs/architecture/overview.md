# 架构概览

## 组件

- **Panel**：Go HTTP API、认证、SQLite/PostgreSQL 存储、Desired State、端口租约、
  流量汇总、诊断编排与 Reconciler。
- **Frontend**：React/Vite 单页应用，通过原始 JWT `Authorization` 头调用 API。
- **Agent**：通过认证 WebSocket 接收目标节点命令，管理节点本机 GOST Runtime，
  上报状态、流量与诊断结果。
- **GOST fork**：位于 `go-gost/` 与 `go-gost/x/`，提供 Listener、Handler、Dialer、
  Relay 及 TCP/UDP 数据平面。

## 三种转发方式

1. 单端口转发使用原版 Forward 数据模型，入口节点直接连接 Landing，不创建 IX
   PathProfile、IX Lease 或 synthetic Tunnel。
2. 隧道转发使用真实 Tunnel。Tunnel Type 决定 Entry/Middle/Exit 拓扑；业务协议与
   TCP/TLS/WSS/MTCP/KCP 等外层传输包装是不同概念。
3. 多端口转发使用 IXPathProfile 和 IXRoute。入口与 IX 端口分别由 Node ID、内部
   Listener、协议维度上的持久化 Lease/Owner 保护。

## 一致性与生命周期

数据库成功不等于 Runtime 成功。系统分别记录 Desired State、Observed State、
Runtime service、Lease/Owner、Outbox 与 Receipt。创建/编辑/删除在部分失败时保留
恢复信息；只有确认所拥有 Listener 消失后才释放端口。共享 PathProfile 或 Relay 的
业务删除必须只清理自己的 Service。

## 诊断语义

- `PASS`：指定方法获得可验证成功证据。
- `FAIL`：测试已执行并获得明确失败证据。
- `UNVERIFIED`：缺少受控 Echo/Payload 响应，不能判定业务失败。
- `NOT_PROBED`：测试未执行。

分段 RTT 不能相加冒充端到端 RTT。UDP 必须由真实 payload 响应验证。

## 数据与版本

Panel 启动时顺序执行内嵌 Migration。发布 `MANIFEST` 记录 Commit、Tree、Schema、
Go、Node 与 pnpm 版本。Agent 与 Panel 使用同一发布包构建，但升级前仍须核对协议
兼容和现有 Listener。
