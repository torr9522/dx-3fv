# DX-3FV v1.0.0-rc.3 Release Notes

## 定位

`v1.0.0-rc.3` 是功能冻结后的第三个 Release Candidate。冻结业务基线仍为
`v0.1.48-fix-frontend-runtime-dependencies`（Schema 12）；本候选仅修复正式发布
材料与构建验证，不改变产品功能、Panel/Agent 协议、数据库或 Runtime 行为。

## 候选淘汰记录

- `v1.0.0-rc.1`：归档可构建，但 Go module SBOM 泄漏构建机绝对路径。
- `v1.0.0-rc.2`：路径清洗契约通过，但正式构建发现错误使用不存在的 Go template
  `json` 函数，构建在 SBOM 阶段失败。

两个 Tag 均保持不可变，不用于部署或分发。

## 修复

- 使用官方支持的 `go list -m -json all` 输出。
- 清洗器结构化解析连续 JSON 对象流，再递归移除构建路径字段。
- 保留 module path、version、sum 和 replacement module identity。
- 契约测试直接执行 Go 1.26.11 的真实 module 列表与清洗管道。

## 发布边界

发布包包含原生 Panel/Agent、前端、systemd、配置模板、安装/升级/回滚/健康检查、
Schema/Migration 说明、Manifest、逐文件与归档 SHA256、SBOM、第三方许可证、NOTICE、
同 Tag 源码归档和中文文档。禁止使用容器部署。

## 已知风险

前海 UDP 陈旧会话根因未定位；小时级持续流量、极限高并发和完整部分失败矩阵覆盖
有限；React Router 7 主版本迁移延期；完整 Git 历史中三个已删除的上游 GOST 配置值
需在公开推送前确认失效或轮换。详情见 `known-issues.md`、
`security-audit-v1.0.0-rc.1.md` 和 `license-compliance.md`。

Secret 扫描最终结论继续为 PARTIAL，生产访问与 Git Push 均为 NONE。
