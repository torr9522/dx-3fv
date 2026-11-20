# DX3FV-0050：Go SBOM 连续 JSON 流兼容

- 类型：Release Blocker / 构建失败
- 基线：`v1.0.0-rc.2`
- 影响：正式发布构建；产品、Schema 与 Runtime 不变

## 根因

`DX3FV-0049` 为得到逐行 JSON，错误假设 `go list` 的 template namespace 提供
`json` 函数。Go 1.26.11 实际执行 `go list -m -f '{{json .}}' all` 时返回
`template: main:1: function "json" not defined`。清洗器的样本契约通过，但没有执行真实
`go list` 管道，因此 `v1.0.0-rc.2` 无法完成正式构建。

## 修复

- 恢复受支持的 `go list -m -json all` 连续 JSON 对象流。
- 清洗器按 JSON 字符串、转义和对象嵌套状态分隔完整对象，再对每个对象执行
  `JSON.parse`；不替换或拼接对象内容。
- 保留 `DX3FV-0049` 的递归字段清洗和 module identity 契约。
- 契约测试除多行嵌套样本外，直接执行真实 Go 1.26.11 module 列表与清洗管道。

`v1.0.0-rc.2` 和既有 Tag 保持原样；修复后使用新的 Release Candidate Tag。

## 回滚

回滚只影响发布构建元数据，不影响已部署服务；回滚会重新导致正式构建失败，因此不能
用于发布。
