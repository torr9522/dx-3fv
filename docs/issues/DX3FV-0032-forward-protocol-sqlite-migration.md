# DX3FV-0032: SQLite 缺失 Forward Protocol 列

## 问题

`v0.1.31-single-port-forward-navigation` 在既有 SQLite 数据库上启动后，
`schema_version` 已记录为 12，但 `forward.protocol` 未创建。真实浏览器访问
`/single-forward` 时，`POST /api/v1/forward/list` 返回：

```text
SQL logic error: no such column: forward.protocol (1)
```

节点请求本身成功，但页面的并发加载因 Forward 查询失败而整体进入错误分支，
导致新增表单没有可选的前置入口。

## 根因

SQLite 为保护历史表约束，`autoMigrateAll` 会跳过已存在的 `forward` 表，改由
`prepareSQLiteLegacyColumns` 对允许的列逐个执行 `Migrator.AddColumn`。0027 新增
`Forward.Protocol` 并将 schema 版本提升至 12，却没有把 `Protocol` 加入该兼容列表。
随后 `migrateSchema` 无条件写入版本 12，因此版本标记不能证明该列存在。

## 修复

- 将 `Protocol` 加入 SQLite Forward 兼容列列表。
- 该准备步骤在 schema 版本判断前执行，因此也能修复已错误标记为版本 12 的数据库。
- 历史 Forward 使用模型默认值 `tcp+udp`，保持原版双协议语义。

## 测试

`TestOpenBackfillsSQLiteLegacyForwardColumns` 现在建立一个缺少 `protocol`、但已经标记
为当前 schema 版本的旧库，验证重新打开后列存在，且历史记录回填为 `tcp+udp`。

部署验收需进一步确认：

- `/forward/list` 恢复成功。
- 单端口页面可选择 `gz2`。
- 原有 Tunnel Forward 仍按 `tcp+udp` 工作。
- SQLite `PRAGMA integrity_check` 为 `ok`。

## 修改文件

- `go-backend/internal/store/repo/repository.go`
- `go-backend/internal/store/repo/repository_migrate_test.go`

## 回滚

代码可回滚到上一 Release；新增列为向后兼容列，不应从数据库删除。回滚前恢复部署前
SQLite 备份，或保留该列供旧二进制忽略。
