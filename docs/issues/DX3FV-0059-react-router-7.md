# DX3FV-0059：React Router 7 主版本迁移

- GitHub Issue：`torr9522/dx-3fv#3`
- 优先级：P1
- 基线：`v0.1.61-fix-udp-session-expiry`
- 目标版本：`react-router@7.18.6`
- 数据库 Schema、Panel API、Agent/GOST Runtime：不变

## 官方依据

执行日 npm `latest` 中 `react-router-dom` 的最新 v7 稳定版为 `7.18.6`；该版本要求
Node 20+、React 18+ 和 React DOM 18+。项目使用 Node 26.13.0、React/React DOM
18.3.1，满足 peer dependency 与 engine 要求。

官方 v6→v7 指南：

`https://github.com/remix-run/react-router/blob/react-router%407.18.6/docs/upgrading/v6.md`

指南说明启用 v7 future behavior 后升级不应产生额外 breaking change，并明确 v7 已简化
包结构：Declarative Mode 可移除 `react-router-dom`，统一从 `react-router` 导入。官方
7.18.6 Changelog：

`https://github.com/remix-run/react-router/blob/react-router%407.18.6/CHANGELOG.md`

7.18.x 包含 URL validation、SSR/RSC 和 CSRF 相关加固。本项目不使用 SSR/RSC 或
Framework Mode，但客户端导航仍采用修复后的实现。

## 项目路由模式审计

项目使用 `BrowserRouter`、`Routes`、`Route`、`Navigate` 和 navigation/location hooks，
属于 Declarative Mode。没有 `createBrowserRouter`/`RouterProvider`、loader/action、
fetcher、SSR hydration、basename、splat route 或组件内部 `React.lazy`。

因此不引入 React Router Vite plugin、Framework Mode、SSR 或新的路由模型；Data Router
专用 future flags 与本项目无关。现有 URL、认证守卫、桌面/手机布局和
`AnimatePresence` 保持原样。

## 修改

- 将直接依赖从 `react-router-dom@6.30.6` 固定迁移到 `react-router@7.18.6`。
- 更新 pnpm lockfile，移除 v6 的 `@remix-run/router` 与 DOM wrapper 图谱。
- 13 个运行时源码文件统一从 `react-router` 导入；路由声明和导航调用不重构。
- 历史设计文档中的示例导入同步到当前包名。

## 验证

- TypeScript + Vite production build PASS。
- 路由相关 ESLint：0 errors（存在本次变更前的格式 warning）。
- 全部轻量 TypeScript 单元测试 PASS。
- `pnpm audit` 中 React Router advisories 从 2 条降为 0；总 findings 从 56 降至 54，
  剩余项不属于 Router 迁移。
- 真实部署后覆盖登录/退出、权限重定向、全部导航、深链接刷新、后退/前进、404、
  Modal、中文搜索，以及 1440px/390px，无 JS error、HTTP 5xx 或水平溢出。

## 回滚

回滚本提交并恢复上一发布归档即可；package.json/lockfile 与静态前端必须作为同一发布
单元回滚。没有 Migration 或数据变更，Panel、Agent 和 Runtime 不需要回滚。
