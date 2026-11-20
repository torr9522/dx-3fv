# 开发工作流

1. 从最新冻结/发布分支建独立 Issue 分支。
2. 先复现并记录真实源码、日志、Runtime 与版本证据。
3. 只修改一个问题，补充自动化和真实操作测试。
4. 更新 `docs/issues/`，提交一个 focused Commit，创建唯一 Annotated Tag。
5. 构建原生发布物，记录 SHA256、Schema 和回滚。
6. 涉及公开仓库时只 push `torr9522/dx-3fv`，禁止 force push 和覆盖 Tag。

发布问题不能通过服务器手工补丁关闭；必须回到源码修复、提交、发布，再从远程仓库
重装验证。
