# Runtime 生命周期

创建资源的成功条件是数据库状态和节点 Runtime 均可恢复，不是单次 HTTP 200。Agent
变更命令由 `WebSocketReporter.mutationQueue` 串行执行；`socket/service.go` 在替换失败时
恢复旧 Service，避免先删后建留下空窗。

Panel 启动后 `StartBackgroundJobs` 运行流量、健康、Tunnel 质量、nftables、联邦与
IXRoute 恢复任务。Agent 上线 hook 会重建所需 Runtime。IXRoute 的状态机至少包含
desired/observed/runtime/lease 四层；迁移 journal 防止编辑中断后新旧端口同时失去所有权。

删除顺序是请求停止 Runtime、记录确认/待清理状态、确认 Listener 不再归该 Owner 后
释放 Lease。共享 Relay 或 PathProfile 只删除目标业务自己的 Service。失败记录保留给
`retryPendingTunnelRuntimeCleanup`、`retryPendingPeerShareOperations` 和
`retryPendingIXRouteOperations`，不得通过直接删数据库行掩盖。
