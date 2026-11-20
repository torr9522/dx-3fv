# 数据库 Migration

DX-3FV Panel 在打开数据库时按顺序执行内嵌 Migration，并将当前版本记录在
`schema_meta`。本发布对应的 Schema 版本同时写入发布根目录 `MANIFEST`。

升级脚本会在停止 Panel 后使用 SQLite `.backup` 创建一致性备份，再启动新版本。
不要绕过升级脚本直接替换二进制。回滚到不兼容的旧 Schema 时，必须把匹配的升级
前数据库备份作为 `rollback-native.sh` 第二个参数。

本目录没有独立 SQL 文件；权威 Migration 实现在 Panel 源码
`go-backend/internal/store/repo/repository.go`，并编译进 Panel 二进制。
