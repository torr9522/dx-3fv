# DX3FV-0066：发布 v1.0.0 制品与安装器映射

- GitHub Issue：`torr9522/dx-3fv#10`
- 类型：publication-only
- 正式 Tag：`v1.0.0`
- 安装器版本：`1.0.0`

## 变更

将正式源码 Annotated Tag 构建的原生包加入 `release/artifacts/`，安装器默认值切换为
固定 `v1.0.0`，并内置精确文件名和 SHA256。RC3/RC4 映射继续保留，用于显式回滚；
安装器不解析 GitHub latest，也不自动安装未发布的 `main` 代码。

正式制品身份：

- Commit：`56b357723e8844cd25711c4f9bd0c33e9f66493e`
- Tree：`170f884a3d1f94bd927809114ca9ce03d6dcd15c`
- 原生包 SHA256：`98e70f224329badf7315b415304ba08de24017f7304ecf2383a99d3921b41936`
- 源码包 SHA256：`b8cbaf6ce4d44f685cb8984d9a6903c56a55048bb8ff37f21aecb5073e0c84c6`
- Schema：12

## 验证与回滚

发布前必须通过 `bash -n`、ShellCheck、installer contract、Native CI、远程 Raw 下载、
包内 SHA256SUMS、Gitleaks、两站升级和重启恢复。回滚使用保留的 RC4 Release 和数据库
备份；不得移动正式 Tag、force push 或删除旧制品。
