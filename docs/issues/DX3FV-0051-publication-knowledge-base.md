# DX3FV-0051：公开发布知识库与历史隔离审计

- 类型：Release documentation
- 状态：完成
- 产品影响：无
- Schema/Agent/GOST 影响：无

## 依据

RC3 当前 Tree 不包含真实凭据，但完整原始历史包含直接上游曾删除的 GOST 配置值。
候选位于 `go-gost/config.json`、`go-gost/gost.json`、`go-gost/ss/config.json` 及其
Android 静态构建副本。相关提交属于上游作者；本项目不能证明其从未使用或已撤销。

因此公开仓库不得直接推送原始 1,275 Commit 历史。发布采用独立 snapshot 历史，
保留完整 RC3 源码、来源、原始 SHA/Tag 对照、时间线和根因记录，不保留危险 blob。
原开发仓库、Source Vault 和 FLVX 冻结基线均不改写。

## 交付

补齐贡献、安全、变更记录，以及架构技能树、数据模型、Agent/GOST、开发流程、来源
血缘、发布历史和公开历史安全说明。所有关键调用链引用实际源码文件和函数。

## 验证

- 275 个 Markdown 文件的本地相对链接检查通过。
- `git diff --check` 通过。
- 当前 Tree gitleaks 仍为 5 个既有误报；本 Issue 未新增候选。
- 原始历史脱敏重扫检查 1,040 个有 diff 的 Commit，仍为相同 9 个候选。
- 后续 publication snapshot 全历史 gitleaks 扫描。

## 版本

- Commit：`docs: build publication knowledge base`。
- Annotated Tag：`v0.1.55-publication-knowledge-base`。

## 回滚

仅删除本 Issue 增加的文档并恢复索引；不需要数据库或 Runtime 回滚。
