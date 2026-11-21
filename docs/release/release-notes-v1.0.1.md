# DX-3FV v1.0.1 - Stable Release

`v1.0.1` 是基于已验收 `v1.0.1-rc.2` 源码的维护稳定版，不改变转发数据面、Agent
协议或数据库 Schema。

## 维护内容

- 单端口转发显示业务所属前置 Node 的公网入口 IP 与入口端口，并支持分别复制。
- 多端口转发按 PathProfile 的 `entryNodeId` 显示前置入口公网 IP；IP、端口可分别复制。
- 不可用、回环、私网、link-local、CGNAT、IPv6 或 `.local` 入口地址显示“未配置”，
  不猜测公网地址。
- 隧道转发页面、Forward/Tunnel 数据模型、Lease、Owner 与 Runtime 保持不变。
- 产品更新源固定为 `torr9522/dx-3fv`，不会使用 FLVX 或其他上游 Release 判断更新。
- 自动 Release 检查、静态资源更新提示、自动下载与自动安装默认关闭；设置页保留显式
  手动检查。
- 修正 stable/RC 通道和语义版本比较，RC 不会把较旧稳定版误报为升级。

## 更新安全

原生安装器只安装明确支持的固定版本并校验精确 SHA256，不跟随未发布的 `main` 或
`latest`。原生升级脚本在替换版本前备份 SQLite 与环境配置，校验包内 SHA256SUMS，
执行健康检查，并在失败时恢复上一版本。

当前 GitHub Release 不提供旧 Docker GUI 所期待的独立 `gost-{ARCH}` 资产，因此不
宣称 Agent GUI 升级可用。Agent 维护使用完整原生包中的 `install-agent-native.sh`。

## 安装

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/torr9522/dx-3fv/main/scripts/install.sh)
```

安装器默认固定 `v1.0.1`；也可以显式传入 `--version v1.0.1`。

## 兼容性与限制

- 数据库版本仍为 Schema 12，既有 Listener、Route、Lease/Owner 不需要迁移。
- 主要验收平台为 Debian 12 amd64、systemd、Caddy 与 SQLite。
- 公网 UDP 仍受前置、IX、Provider NAT、Landing、公网路由和会话状态影响，不承诺
  零丢包。
- 高并发能力取决于服务器资源、线路质量和节点数量。

## 上游与许可证

DX-3FV 派生自 FLVX，并包含 GOST/go-gost fork 及其他第三方组件。发布包保留 GPLv3、
Apache-2.0、MIT 等适用许可证、NOTICE、`THIRD_PARTY_NOTICES.md`、`LICENSES/` 与 SBOM。
