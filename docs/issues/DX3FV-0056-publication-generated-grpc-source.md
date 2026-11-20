# DX3FV-0056：公开快照缺少生成的 gRPC 源码

- 类型：Release build blocker
- 状态：完成
- 产品/Schema/Agent/GOST Runtime 语义影响：无

## 复现

GitHub Native CI 的冷构建在 `go test ./...` 失败，`dialer/grpc` 与 `listener/grpc`
找不到 `GostTunel` 生成类型。本地复用 Go build cache 时未暴露。

## 根因

原开发仓库已经跟踪 `gost_grpc.pb.go` 和 `gost_grpcx.go`，但新 publication root Commit
重新 `git add` 时，原 `.gitignore` 的广泛 `gost_*` 规则把它们排除。sanitized snapshot
因此不是完整可冷构建源码。

## 修复

从冻结知识库 Tag 恢复两个文件，SHA256 与原开发树一致；恢复模块 `.gitignore` 并增加
仅针对这两个路径的 negation。没有重新生成或修改 protobuf 内容。

## 测试

- 使用全新空 `GOCACHE` 运行 `(cd go-gost && go test ./...)`。
- 使用同一空 cache 运行 `(cd go-gost/x && go test ./...)`。
- GitHub Native CI 三个 job 全部通过。

## 回滚

回退本提交会重新导致 clean clone Agent 构建失败；无运行时或数据库回滚。
