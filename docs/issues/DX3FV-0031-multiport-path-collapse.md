# DX3FV-0031: 多端口路径分组折叠

## 目标

多端口业务增长后，每个 PathProfile 可以独立收起业务列表。默认全部展开，路径名称、
状态、节点拓扑和业务数量在收起时仍然可见。

## 实现

- 每个路径 Header 增加固定尺寸的展开/收起图标按钮。
- 折叠只控制对应 `CardBody` 是否进入 React DOM。
- 状态按 PathProfile ID 存入浏览器 Local Storage。
- 每个路径独立切换；刷新页面后保持偏好。
- 新路径没有存储项，因此默认展开。

该状态不会调用 API，不修改 PathProfile、IX Route、Lease、Listener、Desired State 或
Reconciler。编辑、删除和新增业务入口仍保留在路径 Header。

## 验证

- 纯函数测试覆盖默认展开、两个路径独立收起和单独展开。
- TypeScript 与 Vite 生产构建。
- 1440px 与 390px Chromium 验证独立折叠、多组同时折叠、刷新持久化和恢复展开。
- 折叠前后数据库业务数、状态和端口保持一致。

浏览器证据保存到
`/root/dx-3fv-phase2-evidence/screenshots/dx3fv-v0.1.39-multiport-collapse/`。

## 回滚

切换到 `v0.1.38-feature-entry-role-labels`。无数据库变更；旧版本忽略浏览器中的
`dx3fv-multiport-collapsed-paths` 键。
