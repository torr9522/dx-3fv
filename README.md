# DX-3FV

DX-3FV 是基于 FLVX `3.0.1-rc16` 冻结基线继续开发的原生流量转发面板。项目保留
FLVX 的 Forward、Tunnel、Node、用户、监控、限速、共享与 Agent/GOST 能力，并
增加面向受限 NAT/IX 资源的多端口业务、持久化端口租约、统一诊断和累计流量。

本项目的正式部署方式是 Linux 原生二进制、systemd 与 Caddy，不使用 Docker 或
Podman。Release Candidate 以稳定、可升级、可回滚为目标，不代表所有长期稳定性
风险已经关闭。

## 核心能力

- 单端口转发：前置入口直接转发到一个 IPv4 Landing。
- 隧道转发：复用真实 FLVX Tunnel/Relay 资源创建 Forward。
- 多端口转发：前置入口经 IX/NAT 节点到 Landing，端口由持久化 Lease/Owner 管理。
- Tunnel：支持原版 Type 1/2/3 资源模型；Type 2 Relay 可共享一个公网 Listener。
- TCP、UDP 与 TCP+UDP：能力以节点、Provider NAT 与业务类型的交集为准。
- 统一诊断：区分 PASS、FAIL、UNVERIFIED 与 NOT_PROBED，不把缺少 Echo 响应误报
  为业务故障。
- IPv4-only 默认：新业务默认只创建 IPv4 Listener、拨号和诊断目标；历史配置保留
  读取兼容。
- 中文显示名称、累计流量、重启恢复、删除隔离和原生发布生命周期。

## 架构

```text
Browser -> Caddy -> Panel API -> SQLite/PostgreSQL
                         |
                         +-> authenticated WebSocket -> Agent -> embedded GOST
```

Panel 保存 Desired State、资源所有权和流量数据；Agent 执行节点命令并管理本机
GOST Runtime；Reconciler 根据持久状态恢复或清理服务。完整说明见
[架构概览](docs/architecture/overview.md)。

未来维护者可从[源码技能树](docs/architecture/skill-tree.md)、
[模块依赖](docs/architecture/module-dependencies.md)、
[数据库 Schema](docs/database/schema.md)、[Agent 协议](docs/agent/protocol.md)和
[开发时间线](docs/history/development-timeline.md)恢复完整上下文。

## 快速开始

1. 下载同一版本的二进制发布包和源码包，并校验旁边的 `.sha256`。
2. 阅读[原生安装](docs/installation/native.md)，准备 Debian 12/兼容 Linux、systemd、
   Caddy、SQLite CLI、OpenSSL 与 CA bundle。
3. 解压发布包，替换 Caddy 示例域名；仅在需要通行证密钥时配置公开 HTTPS origin。
4. 以 root 执行：

   ```bash
   ./scripts/install-native.sh /absolute/path/to/dx-3fv-VERSION
   ```

5. 安装 Caddy 配置并通过 HTTPS 登录，立即修改初始管理员密码。
6. 在 Panel 创建节点并取得该节点专属 Secret 后，按安装文档部署 Agent。

升级前必须备份数据库和环境文件，并使用发布包自带脚本。参见
[升级](docs/upgrade/native.md)与[回滚](docs/rollback/native.md)。

## 配置

Panel 环境模板位于 `deploy/panel.env.example`：

- `SERVER_ADDR`：Panel 本地监听地址，默认 `127.0.0.1:17365`。
- `DB_TYPE`：`sqlite` 或 `postgres`。
- `DB_PATH`：SQLite 数据库路径。
- `JWT_SECRET`：安装时生成；不得复用模板占位符。
- `FLUX_VERSION`：由安装/升级脚本维护。
- `FLVX_WEBAUTHN_ORIGIN`：可选，必须是浏览器实际访问的完整 HTTPS origin。

Agent 配置包含 Panel 地址和节点 Secret，必须以 `0600` 保存，且不得写入 Git、日志
或问题报告。

## 构建

构建要求 Go `1.26.11`、Node `22.23.4`（或仓库允许的已修复版本）与 pnpm
`10.28.1`。只能从 clean worktree 上指向 HEAD 的 Annotated Tag 构建：

```bash
./scripts/build-native-release.sh v1.0.0-rc.1 /absolute/output/directory
```

脚本生成二进制发布包、匹配源码包、SHA256、Manifest、依赖清单、许可证清单及
中文部署文档。详见[构建说明](docs/development/build.md)。

## 安全与限制

- 不要把 Panel API 或 Agent WebSocket 暴露给不可信网络而不使用 HTTPS/TLS。
- IX/NAT 管理地址、Provider 映射和业务 Landing 是不同安全边界。
- UDP 成功必须有实际 payload 响应证据；ICMP 或 TCP Connect 不能证明 UDP 成功。
- 当前已知风险见[已知问题](docs/release/known-issues.md)。
- 配置、数据库、私钥、节点 Secret 和真实代理 URI 不属于发布物。

## 上游与许可证

DX-3FV 是 FLVX 的派生开发；FLVX 又源自 Apache-2.0 的 flux-panel。DX-3FV 的
修改及相应源代码按根目录 `LICENSE`（GPLv3）发布，同时保留上游 Apache NOTICE。
内嵌 GOST/go-gost fork 及前端组件还包含 MIT 等第三方软件。

分发二进制时必须同时履行 GPLv3 源码提供义务，并保留 `NOTICE`、
`THIRD_PARTY_NOTICES.md`、`LICENSES/` 与 SBOM。完整边界及未决事项见
[许可证合规说明](docs/release/license-compliance.md)。该说明是工程审计，不替代法律
意见。
