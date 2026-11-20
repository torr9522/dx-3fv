# 原生升级

## 升级前

1. 校验发布包与源码包 SHA256、`MANIFEST`、Annotated Tag 和 Release Notes。
2. 记录当前 `/opt/dx-3fv/current`、Panel/Agent 版本、Schema、服务状态和 Listener。
3. 单独备份 `/etc/dx-3fv/panel.env`；不要输出其内容。
4. 使用 SQLite `.backup` 或 PostgreSQL 官方一致性备份工具创建可恢复备份。
5. 确认磁盘空间和上一版本 release 目录仍存在。

## 执行

```bash
tar -xzf dx-3fv-VERSION.tar.gz
cd dx-3fv-VERSION
sha256sum -c SHA256SUMS
sudo ./scripts/upgrade-native.sh "$PWD"
```

脚本创建带 UTC 时间戳的备份目录，停止 Panel，复制新 Release，更新 symlink 与
`FLUX_VERSION`，启动并检查健康。失败时会恢复旧 symlink、升级前数据库和环境文件。

Migration 在新 Panel 打开数据库时自动执行。不要手工跳过 Schema 版本，也不要用
旧二进制打开已升级数据库后继续写入。

## Agent

Agent 必须逐节点升级：先保存 binary、配置、systemd、版本、SHA256 和 Listener
清单，再在低风险节点验证 ONLINE、Runtime、业务与重启恢复。安装脚本当前用于新的
版本目录；已有节点升级应按运维变更窗口替换 symlink，并保留上一目录以便回滚。

## 升级后

检查 HTTPS、登录、Schema、四类状态：Panel、Agent、Runtime、数据库。真实验证
单端口、Tunnel Forward、多端口、共享 Relay、累计流量及节点允许的 TCP/UDP。
最后核对原有业务端口未改变，且没有遗留 Listener。
