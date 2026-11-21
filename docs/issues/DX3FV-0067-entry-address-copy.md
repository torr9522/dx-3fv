# DX3FV-0067：单端口与多端口入口地址独立复制

- 优先级：P1
- GitHub Issue：`torr9522/dx-3fv#11`
- 基线：`v1.0.0`
- 数据库 Schema：不变（Schema 12）
- Agent/GOST Runtime：不变

## 根因

单端口卡片只渲染 Node 名称和入口端口，多端口业务只渲染 `entryPort`。两页虽然已经
加载前置 Node 数据，但没有把业务所属入口 Node 的登记公网地址呈现给用户。单端口 API
也没有把 `forward_port.in_ip` 的业务级入口覆盖单独返回，前端无法在不解析旧复合字段的
情况下区分入口 IP 与端口。

## 地址规则

- 单端口优先使用业务 `forward_port.in_ip`，其次使用所属 Node 的 `serverIpV4`，最后兼容
  `serverIp`。
- 多端口由 PathProfile 的 `entryNodeId` 唯一确定前置 Node，使用该 Node 的登记 IPv4。
- PathProfile 当前强制前置 Pool 为 `full`；NAT Pool 不能作为前置 Pool。IX NAT 的
  `publicIp/publicPort` 只描述 IX 中间映射，不得作为前置入口展示。
- `0.0.0.0`、loopback、RFC1918、link-local、CGNAT、IPv6 和 `.local` 地址不作为公网
  入口展示；缺失或无效时显示“未配置”。

## 修复

- 单端口列表 API 新增只读 `entryIp`，不修改 Forward、Port 或 Node 数据。
- 单端口和多端口卡片分别显示并复制入口 IP、入口端口。
- 复制控件支持鼠标、触摸和键盘；优先 Clipboard API，兼容旧浏览器复制命令，并明确
  提示成功或失败。
- 隧道转发页面和 Tunnel/Forward 数据模型保持原样。

## 测试与回滚

- Repository 覆盖业务入口覆盖、Node IPv4 回退和端口不变。
- 前端 TypeScript、ESLint、生产构建通过。
- 桌面 1440×1000、手机 390×844 读取真实剪贴板验证；刷新和业务编辑后复核。
- 回滚本提交和对应原生发布包即可；没有 Migration，数据库与 Runtime 无需回滚。
