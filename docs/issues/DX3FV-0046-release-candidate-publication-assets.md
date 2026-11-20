# DX3FV-0046：Release Candidate 独立发布材料

- 优先级：P0（发布阻断）
- 基线：`v0.1.49-fix-release-portability`
- 数据库 Schema：不变（Schema 12）
- 产品行为：不变

## 根因

冻结基线的根 README 仍以 FLVX Docker 部署为主，而 DX-3FV 正式支持的是原生部署。
二进制发布包缺少中文安装、升级、回滚、运维、安全、Known Issues 与许可证合规
说明；也没有匹配源码包和完整第三方声明。前端 dependency tree 还记录构建机绝对
路径。新用户不能只依靠发布物完成可审计部署。

## 修复

- 以中文为主重写 DX-3FV README，补齐要求的 `docs/` 信息架构。
- 发布包加入 README、Migration 说明、部署文档、NOTICE、第三方许可证与生产依赖
  许可证清单。
- 对 pnpm dependency/license JSON 结构化移除 `path`/`paths`，保留版本和来源。
- 每个二进制发布包同时生成同 Tagged Commit 的 GPLv3 对应源码归档及 SHA256。
- 增加发布契约测试，防止文档、源码包和脱敏步骤被后续构建遗漏。

没有升级依赖、修改数据模型、改变 Agent 协议或触碰转发 Runtime。

## 验证

- Go release contract tests 与三个 Go module 全套测试。
- `pnpm install --frozen-lockfile`、TypeScript/Vite production build。
- 从最终 Annotated Tag 构建，校验内部/外部 SHA256、Manifest、Schema、文档、SBOM
  与源码归档。
- 对 Git 历史、工作树和最终归档执行敏感信息扫描。
- 在隔离目录完成全新安装，并在测试 Panel 完成升级、回滚、再次升级与浏览器验收。

## 回滚

回滚本 Commit 会删除新增发布材料和源码归档，不影响 Schema、Panel 数据或 Runtime。
产品可继续运行，但发布物不满足 RC 的独立部署与合规交付要求。
