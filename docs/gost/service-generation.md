# Service 生成

Panel 根据 Forward/Tunnel/IXRoute 构造 GOST `config.ServiceConfig` 和 `ChainConfig`，经
Agent WebSocket 发送。普通 TCP/UDP Forward 会生成各自 Listener；TCP+UDP 是两个协议
Service，不是 TLS。Tunnel/Relay 中业务 network 与外层 Dialer transport 分离。

Agent `socket.createServices` 解析配置后注册 Service；`updateServices` 对不变配置跳过
重启，对替换失败恢复旧 Runtime；pause/resume 保留可恢复配置。Service name 包含稳定
ID，是流量归属和清理依据，不能由用户显示名称生成。
