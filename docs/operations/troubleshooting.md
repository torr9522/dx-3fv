# 故障排查

## Panel 无法启动

检查 `systemctl status` 与 journal，随后核对环境文件权限、数据库路径、磁盘空间和
`MANIFEST`。不要输出完整环境文件。Migration 失败时保留现场，恢复升级前备份，
不要手工修改 `schema_meta`。

## Agent ONLINE 但业务不通

按 Desired State、Observed State、Listener、Lease/Owner、Agent journal、GOST
Runtime 顺序检查。确认入口/IX/Landing 的主机、端口和协议来自当前业务记录，而非
历史记录。TCP Connect PASS 不能证明 UDP，诊断 Echo timeout 也可能只是目标不是
Echo Server。

## UDP 首次或持续超时

记录时间线、业务 ID、端口、Agent/Runtime 启动时间、首包与重试结果。检查 UDP
Association、NAT session、Socket bind 和 Runtime readiness。若重启恢复，只能作为
恢复动作和定位证据，不能直接宣称根因已修复。当前 RC 的前海 UDP 陈旧会话风险见
Known Issues。

## 端口无法分配

检查 Node ID 是否正确、Pool/Resource 是否 enabled、协议能力、保留端口、Lease、
原版 Forward/Tunnel ownership 和实际 bind 冲突。公共 IP 相同不表示 NAT 目标相同。

## 回滚失败

保持 Panel 停止，核对目标 Release、匹配的数据库备份和安全副本。不要删除当前
数据库或旧 symlink；先用脚本自动恢复回滚前状态，再从 journal 定位失败步骤。
