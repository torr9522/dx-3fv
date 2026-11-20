# DX3FV-0054：公开历史文档示例精确 Allowlist

- 类型：Release security verification
- 状态：完成
- 产品/Schema/Agent/GOST 影响：无

## 根因

DX3FV-0053 使当前 Tree 达到 0 findings，但 sanitized root Commit 仍保存三处随后修正的
截断 JWT 示例，导致完整 publication history 扫描返回非零。

## 修复

增加只匹配字面量截断示例 `eyJhbGciOiJIUzI1NiIs...` 的 allowlist。它不匹配完整 JWT、
任意 token、整份文件或路径，不能隐藏未来真实凭据。

## 测试

- `gitleaks git --config .gitleaks.toml --redact=100 .`：0 findings。
- 当前 Tree 扫描：0 findings。

## 回滚

回退本提交会恢复三个已分类历史误报，不影响 Runtime。
