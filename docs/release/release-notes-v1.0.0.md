# DX-3FV v1.0.0 - First Stable Release

## 项目介绍

DX-3FV 是基于 FLVX 发展而来的个人转发管理系统。它使用原生 Linux 二进制、systemd、
Caddy 和 SQLite 部署，由 Panel、Agent、内嵌 GOST Fork 与 Reconciler 共同维护转发
Desired/Observed State，不使用 Docker 或 Podman。

## 核心功能

- 普通单端口转发和原版 Tunnel 转发。
- IX 多端口编排、全局/分组新增业务及独立 PathProfile 管理。
- 自动/手动端口、IX PortPool、持久 Lease/Owner 与删除隔离。
- TCP、UDP、TCP+UDP 和 QIX 多 Forward 共享 Relay。
- 单线路累计流量、三段链路诊断与故障恢复。
- IPv4-only 默认、中文显示名称、折叠转发导航。
- Debian 12 amd64 一键原生安装、升级、回滚和备份恢复。

## 架构与数据一致性

浏览器经 Caddy 访问 Panel API；Panel 使用 SQLite 保存 Desired State、资源所有权、
Migration 和流量账本，通过认证 WebSocket 向 Agent 下发命令。Agent 管理内嵌 GOST
Runtime 和 Listener，Reconciler 在重连或重启后恢复 Desired/Observed State。数据库
版本为 Schema 12。

## 安装

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/torr9522/dx-3fv/main/scripts/install.sh)
```

安装器默认固定 `v1.0.0`，不会跟随未发布的 `main` 源码；也可显式使用
`--version v1.0.0`。安装器校验正式原生包的精确 SHA256。

## 验证摘要

- Debian 12 amd64 全新安装、RC3/RC4 升级与回滚、重启恢复均已验收。
- 真实四节点覆盖单端口、Tunnel、前海 TCP/UDP/混合、盲盒 TCP-only 与 QIX 共享
  Relay；SQLite integrity 为 `ok`。
- 完成 3666 秒持续流量和七场景最高 100 并发验证。
- 完成 Panel、SQLite、四 Agent、GOST Runtime、部分创建/删除、Lease、端口冲突、
  编辑回滚和共享 Relay 故障恢复矩阵。

## 已知限制

1. 公网 UDP 受前置服务器、IX、Provider NAT、Landing、公网路由、UDP Session、外部
   代理服务和 Panel/Agent/Runtime 状态共同影响，不承诺零丢包。
2. UDP session expiry 入队竞态已修复；RC4 的 61 分钟窗口仍记录少量可自行恢复的单包
   timeout。该数据如实保留，不足以支持无依据继续修改 Runtime。
3. 单次 UDP timeout 不能单独证明 DX-3FV 产品缺陷，也不能证明问题必然来自外部网络。
4. 高并发能力取决于服务器资源、线路质量和节点数量；100 并发是已验证级别，不是
   无限容量承诺。
5. 主要验收平台为 Debian 12 amd64、systemd、Caddy 和 SQLite；其他系统、架构、
   PostgreSQL 大数据量及多 Panel 副本需要独立测试。

## 上游与许可证

DX-3FV 派生自 FLVX，FLVX 源自 Apache-2.0 的 flux-panel；内嵌 GOST/go-gost fork
及前端组件包含 MIT 等第三方软件。DX-3FV 修改和相应源码按 GPLv3 发布，并保留上游
版权、Apache NOTICE、`THIRD_PARTY_NOTICES.md`、`LICENSES/` 与 SBOM。网络部署者应
阅读仓库许可证合规说明并履行适用的源码提供义务。
