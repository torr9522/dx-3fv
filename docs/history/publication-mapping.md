# 公开历史映射

本仓库使用经过 Secret 隔离的 publication history，不包含原开发仓库的 Git objects。

| 项目 | 原始值 |
| --- | --- |
| FLVX 冻结 Commit | `f357b30500ef99a6cf45d18f0c8a1e1881826b57` |
| FLVX 冻结 Tree | `025c47c85613f4eba341c1e1c5fbd1a1676ca43a` |
| RC3 Commit | `e043e130efe878fe79350879af3bd517114bbe8c` |
| RC3 Tree | `865547d5f279f47c9e761c7c454bcc02f678271b` |
| 知识库快照 Commit | `7b77e0ec5b185861f480690e7bc3444e3caee617` |
| 原可达 Commit 数 | 1,275 |

产品代码以 RC3 为冻结点；知识库快照只增加文档。原始 Tag、对应 Issue 和主要变更在
`docs/issues/`、`docs/history/release-history.md` 与 `CHANGELOG.md` 中保留。RC1/RC2
淘汰原因也保留，但原 Tag object 不复制到本历史，避免把危险 blob 带入公开仓库。

公开 root Commit 的 SHA 会在本文件下一次文档提交中记录。任何比较或溯源必须使用
上表原始 SHA 与源码内容，不能把 publication SHA 当成原始开发 Commit。

上游自动发布、GHCR/Docker 和 npm 工作流没有在 publication snapshot 中启用，因为
它们会在 push/Tag 时产生本轮未授权的外部发布。公开仓库只保留 read-only Native CI。
