# DX3FV-0036: 转发页面标题与导航不一致

## 问题

Phase 2.5 导航已将原规则页面命名为“隧道转发”、原多端口页面命名为“多端口转发”，
但页面内容区仍分别没有标题和显示旧标题“多端口”。用户从菜单进入后无法由 H1 再次
确认当前业务类型。

## 修复

- 隧道转发页增加 H1“隧道转发”及其 Tunnel Forward 说明。
- 多端口页 H1 从“多端口”改为“多端口转发”。

路由、菜单、API、数据库、Agent 和 Runtime 均未修改。

## 验证与回滚

前端生产构建以及 1440px/390px Chromium 分别检查三个转发页 H1 与菜单名称一致。
证据位于 `/root/dx-3fv-phase2-evidence/screenshots/dx3fv-v0.1.40-forward-page-titles/`。

回滚到 `v0.1.39-feature-multiport-path-collapse` 即恢复旧标题；无数据变更。
