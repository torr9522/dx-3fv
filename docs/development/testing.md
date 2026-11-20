# 测试策略

自动化基线见 `CONTRIBUTING.md`。数据库变更运行 migration/contract tests；Agent 命令
运行 `go-gost/x/socket` 与 lifecycle tests；Relay 运行 chain transport/auth/recovery；
IXRoute 运行端口分配、迁移、回滚、删除隔离和诊断测试。

真实验收分为两类：纯净 Debian 12 验证安装/重启/登录；原测试拓扑验证单端口、Tunnel、
前海 TCP/UDP、盲盒 TCP-only、QIX 共享 Relay、流量和诊断。浏览器需覆盖 1440px 与
390px，执行真实点击、输入、保存、刷新和删除。
