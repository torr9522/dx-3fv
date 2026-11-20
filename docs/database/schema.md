# 数据库 Schema 12

模型定义以 `go-backend/internal/store/model/model.go` 为准，当前版本常量为
`internal/store/repo/repository.go:currentSchemaVersion`。

主要域：

- 身份与权限：`user`、`user_group`、权限关联、passkey。
- 节点与 Tunnel：`node`、`node_ssh_config`、`tunnel`、`chain_tunnel` 及节点关联。
- 规则：`forward`、`forward_port`、流量和统计表。
- IX：`ix_port_pool`、`ix_port_resource`、`ix_port_lease`、
  `ix_port_observation`、`ix_path_profile`、`ix_route`、`ix_route_runtime`、
  `ix_route_diagnosis`、`ix_route_migration`。
- 可靠性：runtime cleanup、pending release、`flow_upload_receipt`、metrics。

名称是 UTF-8 显示数据，Runtime identity 使用稳定数值 ID/Service name。不能通过改名
改变 Node ID、Forward ID、Tunnel ID、Lease Owner ID 或 Agent Secret。
