# DX3FV-0041：中文显示名称与稳定内部身份

- 基线：`v0.1.44-feature-ipv4-only-default`
- 数据库 Schema：不变（Schema 12）
- Migration：不需要

## 审计结论

Node、Tunnel、Forward、IX PathProfile、IX Route 和 Group 均以数值 ID
作为主键。Agent 连接使用 Node ID/Secret，端口 Lease 使用 Owner Type/ID，
Forward Runtime Service 使用 `forward_id/user_id/user_tunnel_id`，多端口
Runtime 使用 `route_id/role/protocol`。这些内部身份均不拼接用户显示名称。

`name` 字段已经是可修改的 UTF-8 文本，并用于列表、搜索、诊断标题和错误
信息。SQLite、PostgreSQL、JSON API 和 React 均能原样处理中文，因此新增
第二套 `display_name` 字段会制造不必要的同步与迁移风险。

Federation 的部分导入资源仍使用约定前缀查找，这是独立的远端资源协议。
本 Issue 不批量改 Federation 合成名称，也不修改生产或原版 FLVX。

## 验证范围

- Node 从英文改为中文后 ID、Secret、ONLINE 绑定不变。
- Tunnel、Forward、PathProfile、IX Route 中文名称可完整往返。
- Runtime Service Name 继续只使用稳定数值 ID。
- 浏览器创建、搜索、编辑、诊断和删除中文测试资源。
- 测试站四节点显示名改为：广州2号、前海IX2、盲盒IX、QIX。
- 三条本地多端口路径改为可识别的中文显示名。

## 回滚

代码回滚到 `v0.1.44-feature-ipv4-only-default`。显示名称属于可逆配置，
必要时可通过 UI 改回原英文值；Node ID、Secret、Runtime、Lease 和端口均
不需要回滚。
