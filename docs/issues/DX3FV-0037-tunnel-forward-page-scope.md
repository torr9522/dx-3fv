# DX3FV-0037: 单端口规则泄漏到隧道转发页面

## 问题与根因

单端口页面以 `tunnel_id=0` 识别 direct Forward，但原隧道转发页面继续展示
`forward/list` 返回的全部记录，因此 direct Forward 被放入“未分类”分组。Runtime 和
数据所有权没有混合，问题位于页面数据映射边界。

## 修复

隧道转发页面只映射 `tunnelId>0` 的记录。单端口页面仍只映射 `tunnelId=0`；API、表、
流量、诊断、删除和 Runtime 生命周期均不改变。

## 验证与回滚

- 纯函数测试覆盖正 Tunnel ID、0、null 和 undefined。
- 前端生产构建。
- 桌面/手机确认单端口规则只在单端口页，Tunnel 规则只在隧道转发页。

证据位于 `/root/dx-3fv-phase2-evidence/screenshots/dx3fv-v0.1.41-forward-page-scope/`。
回滚到 `v0.1.40-fix-forward-page-titles` 恢复旧展示；无数据变更。
