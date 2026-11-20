# DX3FV-0052：建立 sanitized publication history

- 类型：Release security
- 状态：完成
- 产品/Schema/Agent/GOST 影响：无

## 根因

原 1,275 Commit 历史包含归属和失效状态无法确认的上游 GOST 配置候选。直接公开会把
已删除 blob 永久复制到新仓库。上游工作流还会在 push/Tag 时自动发布 GHCR、npm 或
GitHub Release，与授权边界冲突。

## 处理

从 `v0.1.55-publication-knowledge-base` 导出完整已跟踪源码，建立独立 root Commit；
记录原 RC3/知识库 SHA 和来源，移除自动发布工作流，增加无写权限、无容器的 Native
CI。原开发仓库、Source Vault 和 FLVX 基线未改写。

## 验证

- publication Tree 与导出快照差异仅为本 Issue 记录、GitHub 模板/安全 CI 和属性文件。
- 当前 Tree 与完整 publication history gitleaks 扫描。
- 三个 Go module tests 与前端 production build。

## 回滚

删除未推送的 publication 仓库并从已记录 Tag 重新导出；不得回退或改写原开发仓库。
