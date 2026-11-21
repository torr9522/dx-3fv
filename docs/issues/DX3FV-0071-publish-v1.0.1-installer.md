# DX3FV-0071：发布 v1.0.1 制品与安装器映射

- GitHub Issue：`torr9522/dx-3fv#15`
- 类型：publication-only
- 正式 Tag：`v1.0.1`
- 安装器版本：`1.0.1`

## 变更

将 `v1.0.1` Annotated Tag 构建的正式原生包加入 `release/artifacts/`，安装器默认值切换
为固定 `v1.0.1`，并内置精确文件名和 SHA256。`v1.0.0` 及既有 RC 映射继续保留，
用于显式安装或回滚；安装器不解析 GitHub `latest`，也不安装未发布的 `main` 源码。

正式制品身份：

- Commit：`5d24f883b33e840c163d079e3ab41c2e4723d677`
- Tree：`70b393c601d132f70b6f694430e6442a526d67f2`
- 原生包 SHA256：`207bba20d65b16614789aec0860f6b43d35089af829796852c44c743e3ad3f01`
- 源码包 SHA256：`be3b08bad2b3de79d0236307b05e6a1d62f92c817df29788a302e1f601be92c7`
- Schema：12

## 验证与回滚

发布前验证 installer contract、Native CI、远程 Raw 下载、包内 SHA256SUMS、GitHub
Release Assets 和 Gitleaks。运行环境升级前保留 RC2 数据库和原生版本备份；失败时使用
原生回滚脚本恢复，不移动正式或 RC Tag。
