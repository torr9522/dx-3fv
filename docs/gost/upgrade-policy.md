# GOST 升级策略

功能冻结期不升级嵌入 GOST。只有可复现发布阻断、安全或数据完整性问题才能独立升级，
且必须审计当前 fork 与目标上游差异、许可证、注册表、TCP/UDP、Relay auth、Tunnel、
traffic outbox、Agent 协议和所有真实业务。

升级必须单独 Issue/Commit/Annotated Tag，保留旧 Agent 二进制和配置，逐节点灰度并验证
回滚。不得用“官方更新”作为缺少 Runtime 证据的替代。
