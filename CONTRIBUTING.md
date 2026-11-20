# 参与 DX-3FV 开发

DX-3FV 当前处于功能冻结期。提交前先阅读 `AGENTS.md`、
`docs/architecture/skill-tree.md` 和 `docs/development/workflow.md`。

每个独立问题必须有一个 `docs/issues/DX3FV-NNNN-*.md`，记录证据、根因、测试、
回滚方式、Commit 和 Annotated Tag。不要把无关修改合入同一提交，不要用 API/Mock
成功代替真实 Runtime 或浏览器验收。

最小验证集合：

```bash
(cd go-backend && go test ./...)
(cd go-gost && go test ./...)
(cd go-gost/x && go test ./...)
(cd vite-frontend && pnpm install --frozen-lockfile && pnpm run build)
```

涉及数据库、Agent 协议、GOST Service、端口所有权或发布脚本时，还必须执行对应契约
测试和原生安装/升级/回滚演练。提交不得包含数据库、环境文件、Agent Secret、私钥、
真实代理 URI 或服务器凭据。
