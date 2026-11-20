# 运维指南

## 日常检查

- `systemctl is-active dx-3fv-panel.service` 与本机 `/flow/test`。
- Caddy HTTPS、证书和 4xx/5xx。
- Panel 中 Agent ONLINE、Observed State、Runtime Listener 与 Desired State。
- Lease/Owner、端口池排除项、Outbox/Receipt 和累计流量是否持续推进。
- CPU、内存、FD、磁盘与数据库备份是否正常。

## 业务操作

新建业务前确认节点能力和 Provider NAT。TCP-only 节点不得创建 UDP Lease/Listener。
手动端口与自动端口使用相同所有权规则。删除业务后，应先确认该业务自己的 Runtime
Service 消失，再确认 Lease 释放；共享 Path/Tunnel 上其他业务必须保持原端口。

诊断必须从 Panel 的真实业务入口执行。不要对无关目标扫描。`UNVERIFIED` 表示缺少
受控响应条件，不等于 PASS，也不等于真实业务 FAIL。

## 备份

SQLite 使用 `.backup` 而非运行中直接复制文件。备份数据库、`panel.env`、Agent
配置、版本 Manifest 和 Listener 清单，但应分开存储并限制权限。定期做恢复演练。

## 变更窗口

一次只升级一个组件或节点，记录前后版本、SHA256、业务证据与回滚路径。Provider
NAT、防火墙、SSH 映射不属于 DX-3FV 发布脚本的修改范围。
