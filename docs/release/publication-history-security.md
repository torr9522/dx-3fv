# 公开历史 Secret 隔离

## 审计结果

RC3 当前 Tree、源码包和二进制包没有确认的 Secret。原开发仓库在 RC3 HEAD 的
`git rev-list --count` 为 1,275；gitleaks 检查其中 1,040 个有 diff 的 Commit，得到 9 个
候选，其中当前文档示例、Authorization 示例和前端 storage key 为误报。其余候选来自
2025 年直接上游提交中的 GOST 配置文件及 Android 打包副本，形态为 32 位十六进制
secret/password。

候选 Commit 为 `9fa968d07cd6c7bc55b479a81c201f5e57d051be`、
`47696e928b14650e009def17dcb13320ec932a8c` 和
`b1647ed309af68a05d70f128ac316f4d5f7c694e`；原值不在本文、证据摘要或公开包中。它们
仍属于上游历史，本项目无法确认凭据归属、使用状态或撤销状态。

## 决策

不公开原始历史。新仓库从审核后的完整源码 snapshot 建立新 root Commit，并记录原始
RC3 HEAD/Tree、原始 Commit 数、Tag/Issue 对照、上游来源和淘汰 RC。原开发仓库、Source
Vault 与基线保持原样。

公开前和每次 push 前扫描 publication 工作树与全历史；任何确认 Secret 都是阻断项。
该方案保留可维护知识和发布血缘，但明确不保留原始逐 Commit Git 对象。
