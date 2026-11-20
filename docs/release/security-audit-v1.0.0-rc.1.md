# v1.0.0-rc.1 安全审计

审计日期：2026-11-20。扫描输出使用 100% redaction，证据文件不进入发布包。

## Secret 扫描

gitleaks 8.28.0 扫描当前工作树约 12 MB，得到 5 个候选项：文档中的示例 token 与
Authorization 环境变量、公开 Keygen account ID、导航 localStorage key 及其构建
副本。逐项检查后均不包含凭据。

RC1 当时的完整 Git 历史扫描覆盖 1,035 个 Commit、约 61 MB，得到 9 个候选项。后续
RC3 历史已增长到 1,275 个 Commit，Phase 4 在公开前重新扫描。除上述误报外，
三个 2025 年上游 Commit 中曾提交、随后删除的 GOST 示例配置值呈现为 32 位十六进制
secret/password，且没有明确占位符标记。它们不在当前 Tree、二进制发布包或 RC 源码
归档中；仅凭仓库无法证明它们从未用于真实环境或已经失效。

本项目不重写上游历史。公开推送完整 Git 历史前，发布者必须确认这些旧值已失效或
完成轮换，并在接受历史保留风险后授权。RC 的 Secret 结论因此为 PARTIAL，而不是
无条件 PASS。

## 前端依赖

`pnpm audit --prod` 报告 15 条公告：1 critical、8 high、6 moderate。

- 13 条位于 Tailwind/PostCSS 构建链：`tar`、`postcss`、`nanoid`、
  `source-map-js`。它们不会进入浏览器静态 bundle，且发布构建不处理不可信 tar/CSS/
  source map 输入；仍需在兼容版本可用时单独升级和回归。
- 2 条 moderate 位于 React Router 6.30.6，修复要求 Router 7.18。涉及 SSR error
  hydration 与用户可控导航目的地；DX-3FV 是 client-only SPA，当前路由目标来自固定
  应用导航。主版本迁移继续延期，不为清零数字破坏冻结 UI。

该判断是可达性分析，不代表公告不存在。新的可达证据或上游兼容修复出现时，应建立
独立安全 Issue。

## 其他边界

- 当前 Runtime/Release 模板没有真实域名、Panel IP、数据库、私钥或 Agent Secret。
- pnpm SBOM 和生产许可证清单移除了构建机绝对 `path/paths`。
- GPLv3 二进制发布配套同 Tagged Commit 源码归档；许可证未决项见
  `license-compliance.md`。
