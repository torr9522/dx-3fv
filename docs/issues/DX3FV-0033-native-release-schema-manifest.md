# DX3FV-0033: Native Release Manifest Schema 版本失真

## 问题

原生发布脚本在 `MANIFEST` 中将 `SCHEMA=11` 写死。0027 已将后端
`currentSchemaVersion` 提升至 12，因此 v0.1.31 的发布元数据与二进制实际迁移行为
不一致。

## 影响

- 不影响 Panel 二进制执行迁移。
- 会误导升级前置检查、证据审计和人工回滚判断。
- 后续 schema 再次升级时容易重复遗漏。

## 根因

发布脚本与后端 schema 常量维护了两份版本来源，且没有一致性校验。

## 修复

`scripts/build-native-release.sh` 在构建开始时从后端
`const currentSchemaVersion = N` 读取唯一版本值。无法得到单一数字时立即中止构建，
不再生成具有不可信元数据的 Release。

## 验证

- 对当前源码解析结果必须为 `12`。
- 新 Release 的 `MANIFEST` 必须包含 `SCHEMA=12`。
- Release 内部 `SHA256SUMS` 和外层归档 SHA256 必须重新生成并通过校验。

## 修改文件

- `scripts/build-native-release.sh`

## 回滚

回滚该 Commit 会恢复写死的 schema 版本，不建议执行。代码回滚不要求更改数据库；
发布包回滚仍应使用部署前 SQLite 备份与上一 Release 目录。
