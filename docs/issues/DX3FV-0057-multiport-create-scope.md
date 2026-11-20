# DX3FV-0057：多端口新增业务路径作用域

- GitHub Issue：`torr9522/dx-3fv#1`
- 优先级：P0
- 基线：`v0.1.59-fix-publication-grpc-sources`
- 数据库 Schema：不变（Schema 12）
- Agent/GOST Runtime：不变

## 根因

分组标题内的“新增业务”已经预填当前 `pathProfileId`，但与编辑和全局创建复用的
Modal 无条件渲染全部 PathProfile 下拉框。后端只验证最终提交的 PathProfile 是否存在，
无法区分全局创建与分组作用域，也就无法拒绝 `pathProfileId` 与分组上下文不一致的请求。

页面同时缺少独立的全局“新增业务”入口。协议兼容性只在端口分配阶段判断，导致
TCP-only IX 仍可能在选择器中显示为 UDP/TCP+UDP 可用路径。

## 修复

- 页面右上角保留“新增路径”，并在其左侧增加全局“新增业务”。
- 分组内创建记录 `scopePathProfileId`，自动绑定当前 PathProfile，只显示不可编辑的路径
  名称，不渲染其他路径下拉框。
- 全局创建按 PathProfile、Node、Pool 状态及 NAT 资源的 TCP/UDP 映射能力过滤路径。
  TCP-only 路径只提供 TCP；协议切换会清除不再兼容的全局路径选择。
- 后端在创建 Route、Lease 或 Runtime 前验证可选 scope 与提交路径完全一致，并验证
  PathProfile、两个 Node、两个 Pool、Node/Pool 归属和协议映射能力。
- 管理员权限继续由现有 `requireIXPortAdmin` 强制；全局创建仍允许管理员选择任一有效
  路径。

## 测试

- Handler 契约：篡改 `scopePathProfileId`/`pathProfileId` 必须失败，且 Route、Lease、
  Runtime 均无残留。
- Repository：TCP-only NAT 资源拒绝 UDP；禁用 PathProfile 或 IX Pool 均拒绝创建。
- `(cd go-backend && GOTOOLCHAIN=go1.26.11 go test ./...)`。
- `(cd vite-frontend && pnpm exec eslint src/pages/multiport.tsx src/api/index.ts)`。
- `(cd vite-frontend && pnpm run build)`。
- 桌面 1440px 与手机 390px 的真实浏览器、全局/分组创建和跨组请求验收随该版本部署
  证据记录。

## 回滚

回滚本提交和对应原生发布包即可。没有 Migration，不修改已有 PathProfile、Route、Lease、
Owner、端口或 Runtime；旧前端会忽略新请求字段，旧数据库无需处理。
