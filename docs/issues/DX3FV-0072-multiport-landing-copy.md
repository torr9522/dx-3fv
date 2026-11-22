# DX3FV-0072：统一多端口“落地”用户文案

- GitHub Issue：`torr9522/dx-3fv#16`
- 基线：`v1.0.1`
- Schema：不变（12）

## 根因与范围

多端口页面和诊断 API 仍有历史英文 `Landing` 用户文案，造成新建、校验、帮助说明与
诊断路径中英混杂。本修复只统一用户可见字符串为“落地”；`landingHost`、
`landingPort`、Go 模型字段、JSON key、类型名和 Runtime 配置不变。

## 验证与回滚

- 前端契约检查旧文案不存在且新建/校验/帮助文案完整。
- 后端诊断测试检查出口和端到端节点统一返回“落地”。
- Chromium 桌面/手机检查新建弹窗、诊断结果、Console、5xx 和水平溢出。
- 回滚本提交即可；无 Migration，不改变 Listener、Lease/Owner 或数据面。
