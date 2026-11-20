# WebSocket 生命周期

`WebSocketReporter.run` 以带 jitter 的指数退避连接；`buildWebSocketCandidates` 优先安全
scheme，并确保日志用 `sanitizeWebSocketURL` 隐去 Secret。连接后指标、配置和命令响应
共用受锁写入路径。

Panel `ws.Server` 按 Node 管理 session、pending request 和 heartbeat。命令发出者必须等待
匹配 `requestId`，超时不能被解释为 Runtime 已应用。断线会令 Node offline，但持久
Desired State 保留，重连 hook 和 Reconciler 负责恢复。
