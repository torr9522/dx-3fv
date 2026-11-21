# DX3FV-0069：修复跨通道更新检查的降级误报

- GitHub Issue：`torr9522/dx-3fv#13`
- 基线：`v1.0.1-rc.1`
- 数据库 Schema：不变（Schema 12）

## 根因

手动检查接口以 `latest != current` 判断是否存在更新。运行版本为
`v1.0.1-rc.1`、stable 通道最新为 `v1.0.0` 时，接口会把较旧版本误报为更新。

## 修复

使用 Go 官方模块工具链的 `golang.org/x/mod/semver` 比较版本。只有目标 Release
严格高于当前版本时 `hasUpdate=true`；相等、较旧或无法可靠解析的版本均为 false。
Release 列表仍会返回，管理员仍可手动查看各通道版本。

## 测试与回滚

自动测试覆盖 patch 升级、RC 到同版本 stable、RC 高于旧 stable、相等版本、未知本地
版本与无效远端版本。回滚本提交即可；无 Migration，不改变 Agent、Runtime、Lease 或
转发业务。
