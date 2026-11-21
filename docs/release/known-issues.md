# 已知问题与运行边界

## 前海 UDP 历史会话陈旧风险

历史持续回归曾两次出现前海 UDP 业务 timeout；同一 Landing 的 QIX UDP 正常，只重启
广州入口 Agent 后恢复。DX3FV-0058 已用确定性失败测试定位到 UDP conn 关闭/到期与
datagram 入队竞态，并完成有针对性的 session lifecycle 修复。canary 的 450/450 UDP、
70 秒空闲后首包和两轮删除重建均通过。

这些证据关闭已复现竞态，不等于无限期无故障证明。DX3FV-0060 已完成 3666 秒真实
持续流量：前海 UDP 为 1748/1749，单次超时在下一请求自行恢复，未复现持续到重启才
恢复的陈旧会话。QIX 一条 UDP lane 为 1743/1749，均为单包超时；这些实际丢包保留在
发布记录中，稳定版不提供公网 UDP 零丢包 SLA。

完整链路还受前置服务器、IX、Provider NAT、Landing、公网路由、UDP Session、外部
代理服务以及 Panel/Agent/Runtime 状态影响。单次 UDP timeout 不能单独证明 DX-3FV
存在产品缺陷，也不能反过来断言所有丢包均来自外部网络。现有证据不足以支持继续无
依据修改 Runtime；若后续取得可稳定复现的软件根因，应建立独立维护 Issue。

## 稳定性覆盖有限

已完成真实 TCP/UDP/Relay 业务、重启恢复和 61 分钟并行持续流量。持续测试没有节点
离线、进程重启、SQLite 错误或资源单调泄漏，但观察到 7 次 UDP 单包超时与 12 次外部
TLS 握手失败。DX3FV-0061 已完成七场景、两轮、最高 100 并发验证，并修复并发创建时
full Pool 穷举造成的全 Panel 阻塞；100 并发仍有 4 次外部 TLS 握手失败，容量上限未做
破坏性探测。DX3FV-0062 已完成 Panel/SQLite/四 Agent/Runtime/部分创建删除/Lease/
编辑回滚/共享 Relay/失败升级的可恢复故障矩阵。稳定版仍不应描述为无限负载、零丢包
或长期无故障。实际容量取决于服务器资源、线路质量和节点数量。

## 前端剩余依赖公告

DX3FV-0059 已把 Declarative Mode 从 React Router 6.30.6 迁移到固定的
`react-router@7.18.6`，Router 相关 audit findings 为 0。剩余公告位于其他运行时或
构建依赖，不能为了清零数字合并无关主版本升级；发布前继续按可达性、严重性和回归
成本逐项审计。

## 原始历史 Secret 隔离

原开发仓库完整历史包含三个已删除上游 GOST 配置中的 32 位十六进制
secret/password。仓库无法证明其从未使用或已经失效，因此该 ancestry 从未推送到
公开仓库。公开仓库使用有血缘对照的 sanitized snapshot history；当前 Tree、完整
publication history 和发布归档均不包含这些候选值。

## 平台范围

原生发布主要在 Debian 12 amd64、systemd、Caddy、SQLite 上验证。其他发行版、
架构、PostgreSQL 大规模数据和多 Panel 副本需要单独验收。WebAuthn challenge 保存在
进程内存，多 Panel 副本需要共享会话设计。
