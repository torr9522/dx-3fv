# DX3FV-0029: 落地 IP 与端口独立输入

## 问题

原隧道转发表单要求用户在多行文本框中手工拼接 `IP:Port`，新增的单端口页面也沿用了
单个 `host:port` 输入框。IPv6 还要求用户理解方括号存储格式，格式错误只能以整行报错。

## 设计

操作层统一使用结构化的 `{ host, port }`：

- 单端口转发固定显示一组“落地 IP / 域名”和“落地端口”。
- 隧道转发默认一组目标；需要负载均衡时通过“添加落地目标”显式增加。
- 多端口原本已使用独立字段，继续保持其既有 API 和生命周期。
- IPv4、IPv6、域名、端口范围和重复目标分别校验，错误定位到具体字段。

API、SQLite、Agent 和 GOST 仍使用 FLVX 的 `remoteAddr` 字符串契约。前端只在 API
边界序列化，IPv6 自动补方括号。编辑旧规则时，逗号分隔的单目标或多目标会解析回
结构化行，旧数据无需 Migration。

## 修改文件

- `vite-frontend/src/components/landing-target-fields.tsx`
- `vite-frontend/src/pages/forward/landing-target.ts`
- `vite-frontend/src/pages/forward/landing-target.test.ts`
- `vite-frontend/src/pages/forward.tsx`
- `vite-frontend/src/pages/single-forward.tsx`

## 自动化验证

- 旧 IPv4、域名和括号 IPv6 解析。
- 兼容 `remoteAddr` 序列化。
- 非法 Host、非法端口和重复目标拒绝。
- 无效数字地址不能被域名校验误接纳。
- TypeScript 与 Vite 生产构建。

真人浏览器证据保存到
`/root/dx-3fv-phase2-evidence/screenshots/dx3fv-v0.1.37-landing-target-fields/`。

## 回滚

切换到 `v0.1.36-fix-duplicate-diagnosis-dom`。本 Issue 没有数据库 Migration，已保存
数据仍是原版兼容的 `remoteAddr`，旧页面可以继续读取和编辑。
