# 发布历史

功能 Tag 和问题映射以 `git tag --format='%(refname:short) %(contents:subject)'` 与
`docs/issues/` 为准。冻结候选：

- `v0.1.48-fix-frontend-runtime-dependencies`：业务冻结基线，Schema 12。
- `v1.0.0-rc.1`：构建成功但 SBOM 含构建机路径，淘汰。
- `v1.0.0-rc.2`：修复路径清洗但正式构建在 Go template 阶段失败，淘汰。
- `v1.0.0-rc.3`：修复 JSON stream 解析，通过原生构建、安装/升级/回滚和真实业务验收。

正式 `v1.0.0` 尚未创建。Phase 4 的 publication Tag 只表示公开历史/安装器候选，不得
冒充最终稳定版。
