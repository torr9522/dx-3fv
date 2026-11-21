# DX3FV-0064：发布 RC4 制品与固定版本安装器映射

- GitHub Issue：`torr9522/dx-3fv#8`
- 优先级：P1 发布准备
- RC Tag：`v1.0.0-rc.4`
- 安装器版本：`0.2.0`
- Schema：12

## 变更

将从不可变 RC4 Tag 构建并校验的原生包加入 `release/artifacts/`。安装器默认版本从 RC3
切换到固定 RC4，并内置精确文件名和 SHA256；RC3 映射继续保留，用于显式旧版本安装和
回滚验证。安装器仍不查询 GitHub latest，也不下载开发机本地文件。

`curl` 或 `wget` 只是 Debian 系统用于取得 `scripts/install.sh` 的下载组件。一键安装器
是下载后执行的完整脚本：它完成系统/架构/资源/端口检查、依赖安装、固定制品校验、
原生 Panel 部署、数据库初始化、随机或自定义管理员、Caddy、ACME 和凭据权限设置；
不把单独执行下载工具称为安装完成。

## 制品

- 二进制包：`dx-3fv-v1.0.0-rc.4.tar.gz`
- SHA256：`86493c9fac8500ef3e8d3504bc555ea14850e915f5f274012ed5f6f594406c17`
- `scripts/install.sh` SHA256：`80f0a2d542148e4f49b9e3c4aefb07a284760b86acc749902b184783ab09fe18`
- RC Commit：`c0ca9bd7eb2b9ee5c5b7ace5b1ac8f456039f227`
- RC Tree：`fc8866369adacab6c51576113d3705b98dcddbbd`

## 验证与回滚

`bash -n`、ShellCheck、installer contract、包内 SHA256SUMS、远程 GitHub 下载 SHA256 和
Debian 12 安装/升级/回滚需要全部通过。回滚安装器 publication commit 会把默认值恢复
为 RC3，但不删除 RC4 Tag 或制品历史；已安装系统使用原生 rollback 脚本和保留 Release。
