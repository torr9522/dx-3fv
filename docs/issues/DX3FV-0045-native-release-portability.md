# DX3FV-0045：原生发布模板可移植性

- 优先级：P0（发布阻断）
- 基线：`v0.1.48-fix-frontend-runtime-dependencies`
- 数据库 Schema：不变（Schema 12）
- 范围：原生发布配置模板，不涉及业务 Runtime

## 根因

`release/native/panel.env.example` 与 `release/native/dx-3fv.caddy` 包含测试站的
真实域名。新用户直接使用发布包时，Caddy 会申请错误域名的证书，WebAuthn 也会
绑定到测试站 origin。该问题不会改变现有部署，但使发布包不能独立、安全地用于
其他环境。

## 修复

- Caddy 使用保留示例域名 `panel.example.com`，并要求安装前替换。
- `FLVX_WEBAUTHN_ORIGIN` 默认留空。通行证密钥为显式启用功能，部署者必须填写
  浏览器实际访问的 HTTPS origin。
- 契约测试拒绝测试域名和测试 Panel IP 重新进入运行时发布模板。

没有修改 Panel、Agent、GOST、数据库、转发协议或现有测试站配置。

## 验证

- Go contract test 验证 WebAuthn 默认关闭及 Caddy 示例域名。
- 最终发布包再次扫描真实测试域名和测试 Panel IP。
- 独立安装演练必须使用部署目标自己的配置，不直接使用模板占位符。

## 回滚

回滚本 Commit 即可。Schema、数据库和 Runtime 状态均不需要恢复。回滚会重新引入
发布包绑定测试环境的问题，因此只适用于历史比对，不建议用于新部署。
