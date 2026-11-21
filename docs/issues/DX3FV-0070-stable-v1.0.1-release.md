# DX3FV-0070：发布 v1.0.1 Stable

- GitHub Issue：`torr9522/dx-3fv#14`
- 验收基线：`v1.0.1-rc.2`
- 基线 Commit：`2faf5a1d599196543122fb1fdd2f05c54ac7af81`
- 基线 Tree：`706fe21825349f40f54df026a634831eda99ba63`
- 数据库 Schema：不变（Schema 12）

## 范围

用户已完成 `v1.0.1-rc.2` 真人验收。本 Issue 只补充稳定版发布说明、构建正式
`v1.0.1` Manifest 与制品，并将固定版本安装器发布到 `main`；不修改产品逻辑、UI、
Agent 协议、Runtime、Lease/Owner 或数据库 Schema。

正式制品必须从指向最终 Release Commit 的 Annotated Tag 构建，不得重命名 RC2
压缩包。由于安装器内置正式制品的精确 SHA256，制品构建完成后使用独立 publication
Commit 添加制品和安装器映射，避免制品包含自身哈希的循环依赖。

## 验证与回滚

- 完整 Native CI、Secret scan、Manifest、SBOM、License/NOTICE 和 Asset SHA256。
- 从 GitHub 重新下载全部 Release Assets 并逐项比对。
- 升级前备份 SQLite、env、Manifest、Listener、Route、Lease/Owner 和 Agent 状态。
- 升级失败或健康检查失败时使用原生回滚脚本恢复 `v1.0.1-rc.2`。
- 不移动 `v1.0.0`、`v1.0.1-rc.1` 或 `v1.0.1-rc.2` Tag。
