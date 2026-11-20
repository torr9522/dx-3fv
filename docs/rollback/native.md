# 原生回滚

## 原则

回滚包含程序和数据两部分。仅切换 symlink 不能证明回滚完成；若新版本执行了旧版
不兼容的 Migration，必须恢复与目标版本匹配的升级前数据库。

## 执行

默认回滚到 `/var/lib/dx-3fv/previous-release`：

```bash
sudo /opt/dx-3fv/current/scripts/rollback-native.sh
```

显式指定目标和数据库备份：

```bash
sudo /opt/dx-3fv/current/scripts/rollback-native.sh \
  /opt/dx-3fv/releases/dx-3fv-PREVIOUS \
  /var/lib/dx-3fv/backups/TIMESTAMP/dx-3fv.db
```

脚本先保存当前环境和可选数据库安全副本，再切换版本并健康检查。失败时恢复回滚前
状态。不要删除安全副本或旧 Release，直到真实业务复验完成。

## 验收

确认 Panel 登录、Schema、节点 ONLINE、原 Listener、单端口、Tunnel、多端口、
共享 Relay、TCP/UDP、累计流量和诊断均与回滚前记录一致。回滚后若计划再次升级，
先解释首次失败原因并重新备份。

Agent 回滚必须逐节点切回保存的 binary/config/systemd 组合，重启后检查 Runtime
Listener 与业务；不要只看版本字符串。
