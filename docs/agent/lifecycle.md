# Agent 生命周期

Agent 启动入口是 `go-gost/main.go`，`program.go` 完成初始配置解析和 GOST Runtime 准备
后才连接 Panel，避免首个命令早于配置加载。Agent 配置和 traffic outbox 必须为 `0600`。

升级通过校验和验证、stage、备份和 systemd/OpenRC restart helper 完成；不得仅凭 Panel
版本号推定 Agent 兼容。逐节点升级后验证 ONLINE、Listener、TCP/UDP、Relay、流量、重启
恢复和旧二进制回滚。
