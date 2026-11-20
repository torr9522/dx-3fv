# DX3FV-0049：Go module SBOM 可移植性

- 类型：Release Blocker / 信息泄漏
- 基线：`v1.0.0-rc.1`
- 影响：发布元数据；Panel、Agent、Schema 与业务行为不变

## 根因

原生构建脚本把 `go list -m -json all` 的输出原样写入两个 Go module SBOM。
该命令除模块身份、版本和校验和外，还输出本机构建目录、module cache 与 `go.mod`
绝对路径。`v1.0.0-rc.1` 因此包含 `/root/dx-3fv-dev` 和
`/root/go/pkg/mod`，不能作为独立、干净的正式候选发布。

## 修复

- 以 Go template 输出逐行 JSON module 对象，不进行文本替换。
- Node 清洗器逐行 `JSON.parse`，递归删除 `Dir`、`GoMod`、`GoModSum`，包括
  `Replace` 下的同名字段。
- 保留模块 `Path`、`Version`、`Sum` 以及 replacement module identity。
- 发布契约使用嵌套 replacement 样本真实执行清洗器，并拒绝路径字段与 `/root/`。

`v1.0.0-rc.1` 保持原样，后续候选使用新 Tag，不移动或覆盖旧 Tag。

## 验证

- Go 发布契约测试。
- Backend、Agent、内嵌 go-gost/x 全套测试与前端生产构建。
- 最终二进制包内两个 Go SBOM 的结构、字段和绝对路径扫描。

## 回滚

回滚本提交只会恢复旧的发布元数据生成方式，不改变运行系统；但会重新引入构建机
绝对路径，因此不能用于公开发布。
