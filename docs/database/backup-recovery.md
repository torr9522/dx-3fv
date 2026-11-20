# 备份与恢复

SQLite 备份必须在停止 Panel 或使用一致性 SQLite backup 后完成，同时保存 `/etc/dx-3fv`
环境文件、当前 Release symlink、版本 Manifest 和 Agent Listener 清单。数据库与 env
含 Secret，不进入 Git 或公开工单。

恢复顺序：停止 Panel，校验备份与目标 Schema，保存当前失败现场，恢复数据库和 env，
切换匹配 Release，启动 Panel，核对登录、Schema、四节点 ONLINE、Listener、TCP/UDP、
Relay、Lease/Owner、Outbox/Receipt 和累计流量。详见 `docs/rollback/native.md`。
