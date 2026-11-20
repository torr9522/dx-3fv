# DX3FV-0053：公开仓库 Secret 扫描零误报基线

- 类型：Release security
- 状态：完成
- 产品/Schema/Agent/GOST 影响：无

## 根因

sanitized root history 已隔离全部上游危险候选，但默认 gitleaks 仍把三处截断 JWT 文档
示例和一个 localStorage 常量名判断为 Secret。未经分类的固定误报会降低后续扫描信号。

## 修复

JWT 示例改为 `<jwt_token>`/`${FLVX_TOKEN}`。`.gitleaks.toml` 只允许精确文本
`FORWARD_NAV_STORAGE_KEY`，不按文件或通用 key 模式放行。

## 测试

- `gitleaks git --redact=100 .`：0 findings。
- Markdown 本地链接检查通过。
- 不改变生成 bundle 的字符串值或运行行为。

## 回滚

回退本提交即可；无数据库、部署或 Runtime 回滚。
