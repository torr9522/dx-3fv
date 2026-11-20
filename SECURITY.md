# 安全策略

请不要在公开 Issue 中提交数据库、日志原文、节点 Secret、JWT、SSH 凭据、私钥或真实
代理 URI。安全报告应只包含可复现的最小脱敏信息，并由仓库维护者在私密渠道接收。

支持范围以最新 Release Candidate 和最新稳定 Tag 为准。修复必须单独建 Issue、测试、
Annotated Tag，并在 `docs/security/` 记录影响和升级建议。

Panel 默认应仅监听 `127.0.0.1`，由 HTTPS 反向代理对外提供服务。Agent Secret 与
Panel `JWT_SECRET` 必须使用 CSPRNG 生成并以 `0600` 保存。发布前必须扫描当前 Tree、
完整发布历史、源码包和二进制包；详见
`docs/security/release-security.md` 与
`docs/release/publication-history-security.md`。
