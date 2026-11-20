# DX3FV-0044：前端运行时依赖安全更新

- 基线：`v0.1.47-fix-contract-node-readiness`
- 数据库 Schema：不变（Schema 12）
- 范围：浏览器运行时直接依赖

## 审计

2026-11-20 的 pnpm audit 报告包含 92 条公告。Critical `tar` 以及多数 glob、
minimatch、Babel、PostCSS、Tailwind 和 Workbox 路径只存在于构建工具，不进入
浏览器 JavaScript。不能把 audit 总数直接等同于可利用的生产风险。

浏览器运行时可达的旧锁定版本为 Axios 1.15.2 和 React Router DOM 6.30.3。
Axios 链包含已修复的请求构造、prototype pollution gadget、ReDoS 以及传递依赖
公告；Router 6.30.3 包含同源 `//` redirect 和 Link XSS 修复缺口。

## 修复

- Axios 更新到 1.21.1。
- follow-redirects 更新到 1.16.3。
- form-data 更新到 4.0.6。
- React Router DOM/React Router 更新到 6.30.6。
- `@remix-run/router` 更新到 1.23.4。

没有一键升级其他依赖。剩余两条 React Router 公告要求升级到 7.18，涉及本项目
不使用的 SSR error hydration 和用户可控导航目的地；主版本迁移会改变路由 API，
本轮记录为 deferred，后续单独迁移和回归。

## 验证

- TypeScript 与 Vite production build。
- 登录、导航、深链接、后退/前进和所有 Forward 子页面。
- Chromium 1440px 与 390px。
- 对比更新前后 pnpm audit 的直接运行时路径。

## 回滚

回滚此 Commit 并重新执行 `pnpm install --frozen-lockfile`。没有数据库、Agent 或
Runtime 状态需要恢复。
