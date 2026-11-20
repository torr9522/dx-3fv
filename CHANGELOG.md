# Changelog

本文件只概括公开里程碑，逐 Issue 证据位于 `docs/issues/`，完整时间线位于
`docs/history/development-timeline.md`。

## v1.0.0-rc.3

- 固定 Schema 12、Panel/Agent/GOST Runtime 和三种转发业务模型。
- 修复 Go module SBOM JSON 流解析并清除构建机绝对路径。
- RC1 因 SBOM 路径泄漏淘汰；RC2 因正式构建失败淘汰。

## v0.1.43 - v0.1.48

- 修正诊断 Echo 证据语义，默认 IPv4-only。
- 验证中文显示名称，增加转发一级导航折叠。
- 修复前端生产依赖并完成冻结前回归。

## v0.1.29 - v0.1.42

- 新增单端口、隧道转发、多端口三级导航与统一诊断。
- 拆分 Landing host/port，增加路径折叠，关闭负载均衡 UI 但保留底层兼容。

## v0.1.1 - v0.1.28

- 建立 IX PortPool、Resource、Lease/Owner、PathProfile、IXRoute、Reconciler、
  累计流量 Outbox/Receipt 与分段诊断。
