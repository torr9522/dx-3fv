# DX-3FV v1.0.0-rc.2 Release Notes

## 定位

`v1.0.0-rc.2` 是功能冻结后的第二个 Release Candidate。冻结业务基线仍为
`v0.1.48-fix-frontend-runtime-dependencies`（Schema 12）；本候选只修复正式发布
材料与测试证据，不新增转发功能，不改变 Panel/Agent 协议、数据库或 Runtime 行为。

`v1.0.0-rc.1` 的最终归档审计发现两个 Go module SBOM 包含构建机绝对路径，因此
该 Tag 和归档原样保留作为被审计淘汰的候选，不用于部署或分发。

## 相对 rc.1 的变化

- Go backend/Agent module SBOM 改为结构化清洗后的 JSON 数组。
- 递归移除 `Dir`、`GoMod`、`GoModSum`，包括 replacement module 中的路径字段。
- 保留 module path、version、sum 和 replacement module identity。
- 发布契约实际执行清洗器，并拒绝路径字段和 `/root/`。

## 发布边界

发布包仍包含原生 Panel/Agent、前端、systemd、配置模板、安装/升级/回滚/健康检查、
Schema/Migration 说明、Manifest、逐文件与归档 SHA256、SBOM、第三方许可证、NOTICE、
同 Tag 源码归档和中文文档。禁止使用容器部署。

## 已知风险

本候选不关闭以下历史风险：前海 UDP 陈旧会话根因未定位；小时级持续流量、极限高并发
和完整部分失败矩阵覆盖有限；React Router 7 主版本迁移延期；完整 Git 历史中三个已
删除的上游 GOST 配置值需在公开推送前确认失效或轮换。

详情见 `known-issues.md`、`security-audit-v1.0.0-rc.1.md` 和
`license-compliance.md`。Secret 扫描最终结论继续为 PARTIAL，Git Push 为 NONE。
