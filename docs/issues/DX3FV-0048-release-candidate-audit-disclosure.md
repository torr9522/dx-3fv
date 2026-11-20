# DX3FV-0048：Release Candidate 审计披露

- 类型：发布审计记录
- 基线：`v0.1.51-fix-relay-cleanup-contract-race`
- 产品与 Schema：不变

## 发现

最终 Secret 扫描覆盖当前 Tree 与 1,035 个历史 Commit。当前 Tree 只有文档示例、
公开标识和 localStorage key 误报；历史中存在三个已删除的上游 GOST 配置值，外形为
非占位的 32 位十六进制 secret/password。无法仅凭源码证明它们未使用或已失效。

前端生产范围 audit 报告 15 条公告，其中 13 条只在构建工具链，2 条 Router moderate
要求不兼容的 Router 7 主版本迁移。当前可达性证据不支持在功能冻结期盲目升级。

## 处理

- 不修改或隐藏上游 Git 历史。
- 当前发布模板、二进制包与 tag source snapshot 不包含历史候选值。
- 将历史候选值、前端公告范围与推送前置条件写入 Security Audit、Known Issues 和
  Release Notes。
- Secret 扫描最终状态标为 PARTIAL；公开 Git Push 继续为 NONE。

## 回滚

回滚只移除审计披露，不改变任何运行行为。由于会隐藏真实风险，不建议回滚。
