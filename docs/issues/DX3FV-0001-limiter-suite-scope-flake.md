# DX3FV-0001：Limiter contract suite readiness 竞态

- 影响：contract 测试基础设施
- 生产数据面：不受影响
- 数据库 Schema：不变（Schema 12）

## 调查

`TestBatchAssignRollbackWhenLimiterDispatchFailsContract` 在历史 suite 运行中
偶发得到成功响应，但独立运行通常通过。当前基线的目标测试连续 100 次、完整
contract suite 单次以及完整 suite 连续 5 次均通过，因此不能把历史失败伪装
成稳定复现。

源码调用链仍存在确定的时序窗口：Gorilla WebSocket 客户端完成 HTTP Upgrade
后，mock helper 立即返回；服务端要在 Upgrade 返回后才把 Node session 写入
`ws.Server.nodes`。测试紧接着发起 batch assign 时，可能把尚未注册的 mock Node
视为离线，因而没有执行预期失败的 AddLimiters dispatch。

## 修复

mock Node 启动读取循环后发送 WebSocket Ping，并等待服务端 Pong。服务端只有在
完成 Node session 注册并进入读取循环后才能处理 Ping，因此 helper 返回时命令
派发目标已经就绪。该同步不使用 sleep、扩大业务超时或放宽断言。

## 验证

- 目标测试高次数重复。
- 完整 contract suite 多次运行。
- Go race detector 覆盖 limiter contract 文件。

## 回滚

回滚此 Commit 即可；没有生产代码、数据库或部署状态需要恢复。
