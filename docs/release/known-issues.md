# Release Candidate 已知问题

## 前海 UDP 历史会话陈旧风险

历史持续回归曾两次出现前海 UDP 业务 timeout；同一 Landing 的 QIX UDP 正常，只重启
广州入口 Agent 后恢复。DX3FV-0058 已用确定性失败测试定位到 UDP conn 关闭/到期与
datagram 入队竞态，并完成有针对性的 session lifecycle 修复。canary 的 450/450 UDP、
70 秒空闲后首包和两轮删除重建均通过。

这些证据关闭已复现竞态，不等于无限期无故障证明。DX3FV-0060 已完成 3666 秒真实
持续流量：前海 UDP 为 1748/1749，单次超时在下一请求自行恢复，未复现持续到重启才
恢复的陈旧会话。QIX 一条 UDP lane 为 1743/1749，均为单包超时；这些实际丢包保留在
发布记录中，不能把 Release Candidate 描述为零丢包 SLA。

## 稳定性覆盖有限

已完成真实 TCP/UDP/Relay 业务、重启恢复和 61 分钟并行持续流量。持续测试没有节点
离线、进程重启、SQLite 错误或资源单调泄漏，但观察到 7 次 UDP 单包超时与 12 次外部
TLS 握手失败。极限高并发和所有部分失败组合仍由 DX3FV-0061/0062 验收；Release
Candidate 不应描述为无限负载、零丢包或长期无故障。

## 前端剩余依赖公告

DX3FV-0059 已把 Declarative Mode 从 React Router 6.30.6 迁移到固定的
`react-router@7.18.6`，Router 相关 audit findings 为 0。剩余公告位于其他运行时或
构建依赖，不能为了清零数字合并无关主版本升级；发布前继续按可达性、严重性和回归
成本逐项审计。

## 历史 Secret 候选项

完整 Git 历史包含三个已删除上游 GOST 配置中的 32 位十六进制 secret/password。
当前 Tree 和发布归档不包含这些值，但仓库无法证明其从未使用或已经失效。因禁止
重写上游历史，公开推送前必须由所有者确认失效或轮换并接受历史保留风险。

## 平台范围

原生发布主要在 Debian 12 amd64、systemd、Caddy、SQLite 上验证。其他发行版、
架构、PostgreSQL 大规模数据和多 Panel 副本需要单独验收。WebAuthn challenge 保存在
进程内存，多 Panel 副本需要共享会话设计。
