# DX-3FV v1.0.0-rc.1 Release Notes

## 定位

这是功能冻结后的首个正式 Release Candidate。它基于 v0.1.48 已验收业务功能，
只修复发布模板可移植性并补齐独立文档、源码归档、SBOM 与许可证材料，没有新增
转发功能或改变 Schema 12。

## 冻结功能

原版 Forward/Tunnel、单端口/隧道/多端口转发、IX PortPool/PathProfile、Lease/
Owner、TCP/UDP、共享 Relay、累计流量、统一诊断、IPv4-only 默认、中文显示名称、
折叠导航、白色浅色背景和 Agent/Runtime 生命周期均进入冻结。

## 发布物

- 原生 Panel 与 Agent 二进制、前端静态文件。
- systemd、Caddy、配置模板、安装/升级/回滚/健康检查脚本。
- Schema/Migration 说明、Manifest、逐文件和归档 SHA256。
- Go/Frontend SBOM、生产依赖许可证清单。
- GPLv3 对应源码归档、NOTICE、第三方许可证与中文文档。

## 升级

必须先备份数据库和环境文件，再使用发布包 `upgrade-native.sh`。Schema 仍为 12。
Agent 可以保持兼容的 v0.1.46 构建运行，但正式统一升级必须逐节点验证 Listener、
TCP/UDP、Relay 与回滚。

## 已知风险

前海 UDP 陈旧会话根因未关闭；小时级持续流量、极限高并发和完整部分失败矩阵未全部
完成。React Router 7 主版本迁移延期。Git 历史中三个已删除的上游 GOST 配置值需
在公开推送前确认失效或完成轮换。详情见 `known-issues.md` 与
`security-audit-v1.0.0-rc.1.md`，这些状态不得写成无条件 PASS。
