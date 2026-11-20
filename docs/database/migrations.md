# Migration

`repo.Open`/`OpenPostgres` 先运行 `autoMigrateAll` 创建兼容表和索引，再运行
`migrateSchema`。Schema 版本存于 `schema_version`，当前为 12。PostgreSQL 还执行
序列、列类型等数据库特定修复。

Migration 必须可从真实旧库运行、保持现有 ID/所有权，并有 SQLite 与相关 PostgreSQL
契约测试。升级脚本在切换 Release 前备份 SQLite；向后不兼容时回滚必须同时提供匹配
数据库备份，不能只切 symlink。
