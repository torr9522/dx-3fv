# DX3FV-0030: GZ 用户可见名称调整为前置入口

## 问题

`GZ` 是用户对具体广州节点的简称，不是业务角色。多端口操作层把入口角色写死为
`GZ`，会误导未来使用日本、美国或其他亚太优化节点的路径。

## 修复

多端口页面的路径说明、空状态、列表、路径表单、业务表单、校验提示和后端业务错误
统一使用“前置入口”。

以下技术身份保持不变：

- 真实节点名 `gz2`。
- Node ID、PathProfile、Pool、Lease 和 Runtime。
- API 字段 `entryNodeId`、`entryPort`。
- 已持久化诊断枚举 `gz_to_ix`。

因此本 Issue 没有数据库 Migration，也不会改变 Agent 命令或现有业务。

## 验证

- 前端生产构建。
- IX Route Handler 测试。
- 源码扫描确认用户可见 JSX 和后端错误中不再使用 `GZ` 角色名。
- 1440px 与 390px Chromium 检查新增/编辑路径、新增/编辑业务、列表和诊断。

浏览器证据保存到
`/root/dx-3fv-phase2-evidence/screenshots/dx3fv-v0.1.38-entry-labels/`。

## 回滚

切换到 `v0.1.37-feature-landing-target-fields`。无数据库变更，真实节点名和所有资源身份
从未修改。
