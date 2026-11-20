package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-backend/internal/http/response"
	"go-backend/internal/store/model"
	"go-backend/internal/store/repo"
)

const ixRouteOwnerEntry = "ix_route_entry"
const ixRouteOwnerIX = "ix_route_ix"

type ixRouteCreateRequest struct {
	ID            int64  `json:"id"`
	PathProfileID int64  `json:"pathProfileId"`
	Name          string `json:"name"`
	EntryPort     int    `json:"entryPort"`
	IXPort        int    `json:"ixPort"`
	LandingHost   string `json:"landingHost"`
	LandingPort   int    `json:"landingPort"`
	ProtocolMode  string `json:"protocolMode"`
}

const ixRouteMigrationOwnerEntry = "ix_route_migration_entry"
const ixRouteMigrationOwnerIX = "ix_route_migration_ix"

type ixRouteIDRequest struct {
	ID int64 `json:"id"`
}

func (h *Handler) ixPathList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet || !requireIXPortAdmin(w, r) {
		return
	}
	items, err := h.repo.ListIXPathProfiles()
	if err != nil {
		response.WriteJSON(w, response.Err(-2, err.Error()))
		return
	}
	response.WriteJSON(w, response.OK(items))
}

func (h *Handler) ixPathCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !requireIXPortAdmin(w, r) {
		return
	}
	var item model.IXPathProfile
	if err := decodeJSON(r.Body, &item); err != nil {
		response.WriteJSON(w, response.Err(500, "请求参数错误"))
		return
	}
	if err := h.repo.CreateIXPathProfile(&item); err != nil {
		response.WriteJSON(w, response.Err(-2, err.Error()))
		return
	}
	response.WriteJSON(w, response.OK(item))
}

func (h *Handler) ixPathUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !requireIXPortAdmin(w, r) {
		return
	}
	var item model.IXPathProfile
	if err := decodeJSON(r.Body, &item); err != nil {
		response.WriteJSON(w, response.Err(500, "请求参数错误"))
		return
	}
	if err := h.repo.UpdateIXPathProfile(&item); err != nil {
		response.WriteJSON(w, response.Err(-2, err.Error()))
		return
	}
	response.WriteJSON(w, response.OK(nil))
}

func (h *Handler) ixPathDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !requireIXPortAdmin(w, r) {
		return
	}
	var request ixRouteIDRequest
	if err := decodeJSON(r.Body, &request); err != nil || request.ID <= 0 {
		response.WriteJSON(w, response.Err(500, "路径不能为空"))
		return
	}
	if err := h.repo.DeleteIXPathProfile(request.ID); err != nil {
		response.WriteJSON(w, response.Err(-2, err.Error()))
		return
	}
	response.WriteJSON(w, response.OK(nil))
}

func (h *Handler) ixRouteList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet || !requireIXPortAdmin(w, r) {
		return
	}
	items, err := h.repo.ListIXRoutes()
	if err != nil {
		response.WriteJSON(w, response.Err(-2, err.Error()))
		return
	}
	response.WriteJSON(w, response.OK(items))
}

func ixRouteProtocols(mode string) ([]string, error) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "tcp":
		return []string{"tcp"}, nil
	case "udp":
		return []string{"udp"}, nil
	case "tcp_udp":
		return []string{"tcp", "udp"}, nil
	default:
		return nil, errors.New("协议模式必须是 TCP、UDP 或 TCP+UDP")
	}
}

func (h *Handler) ixRouteCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !requireIXPortAdmin(w, r) {
		return
	}
	var request ixRouteCreateRequest
	if err := decodeJSON(r.Body, &request); err != nil {
		response.WriteJSON(w, response.Err(500, "请求参数错误"))
		return
	}
	protocols, err := ixRouteProtocols(request.ProtocolMode)
	if err != nil {
		response.WriteJSON(w, response.Err(500, err.Error()))
		return
	}
	request.LandingHost = strings.Trim(strings.TrimSpace(request.LandingHost), "[]")
	if request.LandingHost == "" || strings.ContainsAny(request.LandingHost, "/?#") {
		response.WriteJSON(w, response.Err(500, "Landing IP 或域名无效"))
		return
	}
	if err := validateIPv4OnlyHost(r.Context(), request.LandingHost); err != nil {
		response.WriteJSON(w, response.Err(500, "Landing 目标无效: "+err.Error()))
		return
	}
	route := &model.IXRoute{PathProfileID: request.PathProfileID, Name: request.Name, LandingHost: request.LandingHost, LandingPort: request.LandingPort, ProtocolMode: request.ProtocolMode}
	if err := h.repo.CreateIXRoute(route); err != nil {
		response.WriteJSON(w, response.Err(-2, err.Error()))
		return
	}
	profile, err := h.repo.GetIXPathProfile(route.PathProfileID)
	if err != nil || profile == nil {
		_ = h.repo.SetIXRouteState(route.ID, "deleted", "error", "path profile not found")
		response.WriteJSON(w, response.Err(-2, "路径不存在"))
		return
	}
	entryAllocation, err := h.repo.AllocateIXPorts(repo.IXPortAllocationRequest{PoolID: profile.EntryPoolID, RequestedPort: request.EntryPort, Protocols: protocols, OwnerType: ixRouteOwnerEntry, OwnerID: route.ID, Manual: request.EntryPort > 0})
	if err != nil {
		_ = h.repo.SetIXRouteState(route.ID, "deleted", "error", err.Error())
		response.WriteJSON(w, response.Err(-2, "前置入口端口分配失败: "+err.Error()))
		return
	}
	ixAllocation, err := h.repo.AllocateIXPorts(repo.IXPortAllocationRequest{PoolID: profile.IXPoolID, RequestedPort: request.IXPort, Protocols: protocols, OwnerType: ixRouteOwnerIX, OwnerID: route.ID, Manual: request.IXPort > 0})
	if err != nil {
		_ = h.repo.RequestIXPortLeaseRelease(ixRouteOwnerEntry, route.ID)
		_ = h.repo.CompleteIXPortLeaseRelease(ixRouteOwnerEntry, route.ID)
		_ = h.repo.SetIXRouteState(route.ID, "deleted", "error", err.Error())
		response.WriteJSON(w, response.Err(-2, "IX 中间端口分配失败: "+err.Error()))
		return
	}
	if err := h.repo.UpdateIXRouteAllocation(route.ID, entryAllocation.Resource.InternalPort, ixAllocation.Resource.PublicPort, ixAllocation.Resource.InternalPort); err != nil {
		h.releaseIXRouteLeasesWithoutRuntime(route.ID)
		_ = h.repo.SetIXRouteState(route.ID, "deleted", "error", err.Error())
		response.WriteJSON(w, response.Err(-2, err.Error()))
		return
	}
	route.EntryPort = entryAllocation.Resource.InternalPort
	route.IXPublicPort = ixAllocation.Resource.PublicPort
	route.IXInternalPort = ixAllocation.Resource.InternalPort
	items, err := h.buildIXRouteRuntimes(route, profile, entryAllocation.Resource, ixAllocation.Resource, protocols)
	if err != nil {
		h.releaseIXRouteLeasesWithoutRuntime(route.ID)
		_ = h.repo.SetIXRouteState(route.ID, "deleted", "error", err.Error())
		response.WriteJSON(w, response.Err(-2, err.Error()))
		return
	}
	if err := h.repo.SaveIXRouteRuntimes(items); err != nil {
		h.releaseIXRouteLeasesWithoutRuntime(route.ID)
		_ = h.repo.SetIXRouteState(route.ID, "deleted", "error", err.Error())
		response.WriteJSON(w, response.Err(-2, err.Error()))
		return
	}
	for attempt := 0; attempt < 8; attempt++ {
		reconcileErr := h.reconcileIXRoute(route.ID)
		if reconcileErr == nil || shouldDeferTunnelRuntimeApplyError(reconcileErr) {
			break
		}
		var laneErr *ixRouteReconcileError
		if !errors.As(reconcileErr, &laneErr) || !isBindAddressInUseError(reconcileErr) {
			response.WriteJSON(w, response.Err(-2, "业务已保存但 Runtime 部署失败: "+reconcileErr.Error()))
			return
		}
		manual := (laneErr.role == "entry" && request.EntryPort > 0) || (laneErr.role == "ix" && request.IXPort > 0)
		if manual {
			h.cleanupFailedIXRoute(route.ID, reconcileErr)
			response.WriteJSON(w, response.Err(-2, "手动指定端口已被 Runtime 占用: "+reconcileErr.Error()))
			return
		}
		ownerType := ixRouteOwnerEntry
		poolID := profile.EntryPoolID
		conflictPort := entryAllocation.Resource.InternalPort
		conflictNode := profile.EntryNodeID
		if laneErr.role == "ix" {
			ownerType = ixRouteOwnerIX
			poolID = profile.IXPoolID
			conflictPort = ixAllocation.Resource.InternalPort
			conflictNode = profile.IXNodeID
		}
		for _, protocol := range protocols {
			_ = h.repo.UpsertIXPortObservation(&model.IXPortObservation{NodeID: conflictNode, InternalPort: conflictPort, Protocol: protocol, State: "conflict", Error: reconcileErr.Error()})
		}
		if err := h.repo.RequestIXPortLeaseRelease(ownerType, route.ID); err != nil {
			h.cleanupFailedIXRoute(route.ID, err)
			response.WriteJSON(w, response.Err(-2, err.Error()))
			return
		}
		if err := h.repo.CompleteIXPortLeaseRelease(ownerType, route.ID); err != nil {
			h.cleanupFailedIXRoute(route.ID, err)
			response.WriteJSON(w, response.Err(-2, err.Error()))
			return
		}
		replacement, err := h.repo.AllocateIXPorts(repo.IXPortAllocationRequest{PoolID: poolID, Protocols: protocols, OwnerType: ownerType, OwnerID: route.ID})
		if err != nil {
			h.cleanupFailedIXRoute(route.ID, err)
			response.WriteJSON(w, response.Err(-2, "自动端口重试失败: "+err.Error()))
			return
		}
		if laneErr.role == "ix" {
			ixAllocation = replacement
			route.IXPublicPort = replacement.Resource.PublicPort
			route.IXInternalPort = replacement.Resource.InternalPort
		} else {
			entryAllocation = replacement
			route.EntryPort = replacement.Resource.InternalPort
		}
		if err := h.repo.UpdateIXRouteAllocation(route.ID, route.EntryPort, route.IXPublicPort, route.IXInternalPort); err != nil {
			h.cleanupFailedIXRoute(route.ID, err)
			response.WriteJSON(w, response.Err(-2, err.Error()))
			return
		}
		items, err = h.buildIXRouteRuntimes(route, profile, entryAllocation.Resource, ixAllocation.Resource, protocols)
		if err != nil {
			h.cleanupFailedIXRoute(route.ID, err)
			response.WriteJSON(w, response.Err(-2, err.Error()))
			return
		}
		if err := h.repo.SaveIXRouteRuntimes(items); err != nil {
			h.cleanupFailedIXRoute(route.ID, err)
			response.WriteJSON(w, response.Err(-2, err.Error()))
			return
		}
		if attempt == 7 {
			exhausted := errors.New("自动端口在 8 次 Runtime 冲突后仍无法分配")
			h.cleanupFailedIXRoute(route.ID, exhausted)
			response.WriteJSON(w, response.Err(-2, exhausted.Error()))
			return
		}
	}
	stored, _ := h.repo.GetIXRoute(route.ID)
	response.WriteJSON(w, response.OK(stored))
}

func activeProtocolsFromLeases(leases []model.IXPortLease) map[string]model.IXPortLease {
	out := make(map[string]model.IXPortLease)
	for _, lease := range leases {
		if lease.State == "active" {
			out[lease.Protocol] = lease
		}
	}
	return out
}

func (h *Handler) prepareIXRouteEditLane(ownerType, migrationOwner string, routeID, migrationID, poolID int64, currentPublicPort, requestedPort int, protocols []string) (model.IXPortResource, []int64, error) {
	leases, err := h.repo.ListIXPortLeases(ownerType, routeID)
	if err != nil {
		return model.IXPortResource{}, nil, err
	}
	active := activeProtocolsFromLeases(leases)
	if len(active) == 0 {
		return model.IXPortResource{}, nil, errors.New("活动端口 Lease 缺失")
	}
	var currentResource *model.IXPortResource
	for _, lease := range active {
		currentResource, err = h.repo.GetIXPortResource(lease.ResourceID)
		if err != nil {
			return model.IXPortResource{}, nil, err
		}
		if currentResource != nil {
			break
		}
	}
	if currentResource == nil {
		return model.IXPortResource{}, nil, errors.New("活动端口资源缺失")
	}
	samePort := requestedPort > 0 && requestedPort == currentPublicPort
	releaseIDs := make([]int64, 0, len(active))
	if samePort {
		desired := make(map[string]bool, len(protocols))
		missing := make([]string, 0, len(protocols))
		for _, protocol := range protocols {
			desired[protocol] = true
			if _, ok := active[protocol]; !ok {
				missing = append(missing, protocol)
			}
		}
		for protocol, lease := range active {
			if !desired[protocol] {
				releaseIDs = append(releaseIDs, lease.ID)
			}
		}
		if len(missing) > 0 {
			if _, err := h.repo.AllocateIXPorts(repo.IXPortAllocationRequest{PoolID: poolID, RequestedPort: requestedPort, Protocols: missing, OwnerType: migrationOwner, OwnerID: migrationID, Manual: true}); err != nil {
				return model.IXPortResource{}, nil, err
			}
		}
		return *currentResource, releaseIDs, nil
	}
	allocation, err := h.repo.AllocateIXPorts(repo.IXPortAllocationRequest{PoolID: poolID, RequestedPort: requestedPort, Protocols: protocols, OwnerType: migrationOwner, OwnerID: migrationID, Manual: requestedPort > 0})
	if err != nil {
		return model.IXPortResource{}, nil, err
	}
	for _, lease := range active {
		releaseIDs = append(releaseIDs, lease.ID)
	}
	return allocation.Resource, releaseIDs, nil
}

func serviceNamesByNode(items []model.IXRouteRuntime) map[int64]map[string]bool {
	out := make(map[int64]map[string]bool)
	for _, item := range items {
		if out[item.NodeID] == nil {
			out[item.NodeID] = make(map[string]bool)
		}
		out[item.NodeID][item.ServiceName] = true
	}
	return out
}

func (h *Handler) applyIXRouteRuntimeSet(desired, previous []model.IXRouteRuntime) error {
	items := append([]model.IXRouteRuntime(nil), desired...)
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Role == items[j].Role {
			return items[i].Protocol < items[j].Protocol
		}
		return items[i].Role == "ix"
	})
	for start := 0; start < len(items); {
		end := start + 1
		for end < len(items) && items[end].Role == items[start].Role && items[end].NodeID == items[start].NodeID {
			end++
		}
		configs := make([]map[string]interface{}, 0, end-start)
		for _, item := range items[start:end] {
			var config map[string]interface{}
			if err := json.Unmarshal([]byte(item.Config), &config); err != nil {
				return err
			}
			configs = append(configs, config)
		}
		if _, err := h.sendNodeCommand(items[start].NodeID, "UpdateService", configs, false, false); err != nil {
			return &ixRouteReconcileError{role: items[start].Role, err: err}
		}
		start = end
	}
	desiredNames := serviceNamesByNode(desired)
	staleByNode := make(map[int64][]string)
	for _, item := range previous {
		if !desiredNames[item.NodeID][item.ServiceName] {
			staleByNode[item.NodeID] = append(staleByNode[item.NodeID], item.ServiceName)
		}
	}
	for nodeID, names := range staleByNode {
		if _, err := h.sendNodeCommand(nodeID, "DeleteService", map[string]interface{}{"services": names}, false, true); err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) rollbackIXRouteEdit(migration *model.IXRouteMigration, cause error) error {
	oldRoute, _, oldRuntimes, newRuntimes, _, decodeErr := repo.DecodeIXRouteMigrationPlan(migration)
	if decodeErr == nil {
		if err := h.applyIXRouteRuntimeSet(oldRuntimes, newRuntimes); err != nil {
			_ = h.repo.SetIXRouteMigrationState(migration.ID, "rollback_pending", errors.Join(cause, err).Error())
			_ = h.repo.SetIXRouteState(migration.RouteID, "", "degraded", "编辑回滚待恢复: "+err.Error())
			return err
		}
	}
	message := ""
	if cause != nil {
		message = cause.Error()
	}
	if err := h.repo.RollbackIXRouteMigration(migration, message); err != nil {
		return err
	}
	if oldRoute != nil {
		_ = h.repo.SetIXRouteState(oldRoute.ID, "", "active", "编辑失败，旧配置已恢复: "+message)
	}
	return nil
}

func (h *Handler) ixRouteUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !requireIXPortAdmin(w, r) {
		return
	}
	var request ixRouteCreateRequest
	if err := decodeJSON(r.Body, &request); err != nil || request.ID <= 0 {
		response.WriteJSON(w, response.Err(500, "请求参数错误"))
		return
	}
	protocols, err := ixRouteProtocols(request.ProtocolMode)
	if err != nil || strings.TrimSpace(request.Name) == "" || strings.TrimSpace(request.LandingHost) == "" || request.LandingPort <= 0 || request.LandingPort > 65535 || request.EntryPort < 0 || request.EntryPort > 65535 || request.IXPort < 0 || request.IXPort > 65535 {
		if err == nil {
			err = errors.New("业务名称、Landing 和端口参数无效")
		}
		response.WriteJSON(w, response.Err(500, err.Error()))
		return
	}
	request.LandingHost = strings.Trim(strings.TrimSpace(request.LandingHost), "[]")
	current, err := h.repo.GetIXRoute(request.ID)
	if err != nil || current == nil {
		response.WriteJSON(w, response.Err(-2, "多端口业务不存在"))
		return
	}
	if request.LandingHost != strings.TrimSpace(current.LandingHost) {
		if err := validateIPv4OnlyHost(r.Context(), request.LandingHost); err != nil {
			response.WriteJSON(w, response.Err(500, "Landing 目标无效: "+err.Error()))
			return
		}
	}
	if current.DesiredState != "active" || current.ObservedState != "active" {
		response.WriteJSON(w, response.Err(-2, "只有已生效的业务可以编辑"))
		return
	}
	if request.PathProfileID != current.PathProfileID {
		response.WriteJSON(w, response.Err(-2, "活动业务不能在编辑时更换前置入口到 IX 路径"))
		return
	}
	migration, err := h.repo.BeginIXRouteMigration(current.ID)
	if err != nil {
		response.WriteJSON(w, response.Err(-2, err.Error()))
		return
	}
	failPreparing := func(cause error) {
		_ = h.repo.RollbackIXRouteMigration(migration, cause.Error())
		response.WriteJSON(w, response.Err(-2, cause.Error()))
	}
	profile, err := h.repo.GetIXPathProfile(current.PathProfileID)
	if err != nil || profile == nil {
		failPreparing(errors.New("前置入口到 IX 路径不存在"))
		return
	}
	oldRuntimes, err := h.repo.ListIXRouteRuntimes(current.ID)
	if err != nil || len(oldRuntimes) == 0 {
		failPreparing(errors.New("旧 Runtime 状态缺失"))
		return
	}
	entryResource, entryRelease, err := h.prepareIXRouteEditLane(ixRouteOwnerEntry, ixRouteMigrationOwnerEntry, current.ID, migration.ID, profile.EntryPoolID, current.EntryPort, request.EntryPort, protocols)
	if err != nil {
		failPreparing(errors.New("前置入口端口迁移失败: " + err.Error()))
		return
	}
	ixResource, ixRelease, err := h.prepareIXRouteEditLane(ixRouteOwnerIX, ixRouteMigrationOwnerIX, current.ID, migration.ID, profile.IXPoolID, current.IXPublicPort, request.IXPort, protocols)
	if err != nil {
		failPreparing(errors.New("IX 中间端口迁移失败: " + err.Error()))
		return
	}
	candidate := *current
	candidate.Name = strings.TrimSpace(request.Name)
	candidate.EntryPort = entryResource.InternalPort
	candidate.IXPublicPort = ixResource.PublicPort
	candidate.IXInternalPort = ixResource.InternalPort
	candidate.LandingHost = strings.TrimSpace(request.LandingHost)
	candidate.LandingPort = request.LandingPort
	candidate.ProtocolMode = strings.ToLower(strings.TrimSpace(request.ProtocolMode))
	newRuntimes, err := h.buildIXRouteRuntimes(&candidate, profile, entryResource, ixResource, protocols)
	if err != nil {
		failPreparing(err)
		return
	}
	var generation int64 = 1
	for _, item := range oldRuntimes {
		if item.Generation >= generation {
			generation = item.Generation + 1
		}
	}
	for i := range newRuntimes {
		newRuntimes[i].Generation = generation
	}
	releaseIDs := append(entryRelease, ixRelease...)
	if err := h.repo.SaveIXRouteMigrationPlan(migration, current, &candidate, oldRuntimes, newRuntimes, releaseIDs); err != nil {
		failPreparing(err)
		return
	}
	migration, err = h.repo.GetActiveIXRouteMigration(current.ID)
	if err != nil || migration == nil {
		if err == nil {
			err = errors.New("多端口业务迁移日志丢失")
		}
		failPreparing(err)
		return
	}
	_ = h.repo.SetIXRouteMigrationState(migration.ID, "applying", "")
	if err := h.applyIXRouteRuntimeSet(newRuntimes, oldRuntimes); err != nil {
		_ = h.repo.SetIXRouteMigrationState(migration.ID, "rollback_pending", err.Error())
		rollbackErr := h.rollbackIXRouteEdit(migration, err)
		if rollbackErr != nil {
			response.WriteJSON(w, response.Err(-2, "编辑失败且旧 Runtime 等待恢复: "+errors.Join(err, rollbackErr).Error()))
			return
		}
		response.WriteJSON(w, response.Err(-2, "编辑失败，旧配置已恢复: "+err.Error()))
		return
	}
	_ = h.repo.SetIXRouteMigrationState(migration.ID, "finalizing", "")
	if err := h.repo.CommitIXRouteMigration(migration); err != nil {
		_ = h.repo.SetIXRouteMigrationState(migration.ID, "rollback_pending", err.Error())
		rollbackErr := h.rollbackIXRouteEdit(migration, err)
		response.WriteJSON(w, response.Err(-2, "编辑提交失败: "+errors.Join(err, rollbackErr).Error()))
		return
	}
	stored, _ := h.repo.GetIXRoute(current.ID)
	response.WriteJSON(w, response.OK(stored))
}

func (h *Handler) cleanupFailedIXRoute(routeID int64, cause error) {
	message := ""
	if cause != nil {
		message = cause.Error()
	}
	_ = h.repo.SetIXRouteState(routeID, "deleted", "cleanup_pending", message)
	_ = h.repo.SetIXRouteRuntimesDesired(routeID, "deleted")
	for _, ownerType := range []string{ixRouteOwnerEntry, ixRouteOwnerIX} {
		_ = h.repo.RequestIXPortLeaseRelease(ownerType, routeID)
	}
	_ = h.reconcileIXRoute(routeID)
}

func (h *Handler) releaseIXRouteLeasesWithoutRuntime(routeID int64) {
	for _, ownerType := range []string{ixRouteOwnerEntry, ixRouteOwnerIX} {
		_ = h.repo.RequestIXPortLeaseRelease(ownerType, routeID)
		_ = h.repo.CompleteIXPortLeaseRelease(ownerType, routeID)
	}
}

func routeEndpoint(host string, port int) string {
	return processServerAddress(net.JoinHostPort(strings.Trim(strings.TrimSpace(host), "[]"), strconv.Itoa(port)))
}

func routeListenerAddress(node *model.Node, resource model.IXPortResource, protocol string) string {
	host := strings.Trim(strings.TrimSpace(resource.InternalIP), "[]")
	if host == "" && node != nil {
		host = strings.Trim(strings.TrimSpace(node.TCPListenAddr), "[]")
		if protocol == "udp" {
			host = strings.Trim(strings.TrimSpace(node.UDPListenAddr), "[]")
		}
	}
	if host == "" {
		host = "0.0.0.0"
	}
	return routeEndpoint(host, resource.InternalPort)
}

func ixRouteServiceConfig(name, protocol, address, target string) map[string]interface{} {
	service := map[string]interface{}{
		"name":     name,
		"addr":     address,
		"handler":  map[string]interface{}{"type": protocol},
		"listener": map[string]interface{}{"type": protocol},
		"forwarder": map[string]interface{}{
			"nodes":    buildForwarderNodes([]string{target}),
			"selector": map[string]interface{}{"strategy": "fifo", "maxFails": 1, "failTimeout": "30s"},
		},
	}
	if protocol == "udp" {
		service["listener"].(map[string]interface{})["metadata"] = map[string]interface{}{"keepAlive": true, "ttl": "30s"}
	}
	return service
}

func (h *Handler) buildIXRouteRuntimes(route *model.IXRoute, profile *model.IXPathProfile, entryResource, ixResource model.IXPortResource, protocols []string) ([]model.IXRouteRuntime, error) {
	entryNode, err := h.repo.GetNodeByID(profile.EntryNodeID)
	if err != nil || entryNode == nil {
		return nil, errors.New("前置入口节点不存在")
	}
	ixNode, err := h.repo.GetNodeByID(profile.IXNodeID)
	if err != nil || ixNode == nil {
		return nil, errors.New("IX 节点不存在")
	}
	ixTargetHost := strings.TrimSpace(ixResource.PublicIP)
	if ixTargetHost == "" {
		ixTargetHost = strings.TrimSpace(ixNode.ServerIP)
	}
	if ixTargetHost == "" {
		return nil, errors.New("IX 公网地址不能为空")
	}
	now := time.Now().UnixMilli()
	items := make([]model.IXRouteRuntime, 0, len(protocols)*2)
	for _, protocol := range protocols {
		for _, lane := range []struct {
			role    string
			nodeID  int64
			name    string
			address string
			target  string
		}{
			{role: "ix", nodeID: profile.IXNodeID, name: fmt.Sprintf("dx3fv_ix_%d_ix_%s", route.ID, protocol), address: routeListenerAddress(ixNode, ixResource, protocol), target: routeEndpoint(route.LandingHost, route.LandingPort)},
			{role: "entry", nodeID: profile.EntryNodeID, name: fmt.Sprintf("dx3fv_ix_%d_entry_%s", route.ID, protocol), address: routeListenerAddress(entryNode, entryResource, protocol), target: routeEndpoint(ixTargetHost, ixResource.PublicPort)},
		} {
			encoded, err := json.Marshal(ixRouteServiceConfig(lane.name, protocol, lane.address, lane.target))
			if err != nil {
				return nil, err
			}
			items = append(items, model.IXRouteRuntime{RouteID: route.ID, NodeID: lane.nodeID, Role: lane.role, Protocol: protocol, ServiceName: lane.name, Config: string(encoded), DesiredState: "active", ObservedState: "pending", Generation: 1, CreatedTime: now, UpdatedTime: now})
		}
	}
	return items, nil
}

type ixRouteReconcileError struct {
	role string
	err  error
}

func (e *ixRouteReconcileError) Error() string { return e.role + ": " + e.err.Error() }
func (e *ixRouteReconcileError) Unwrap() error { return e.err }

func (h *Handler) reconcileIXRoute(routeID int64) error {
	migration, migrationErr := h.repo.GetActiveIXRouteMigration(routeID)
	if migrationErr != nil {
		return migrationErr
	}
	if migration != nil {
		if migration.State != "rollback_pending" && time.Now().UnixMilli()-migration.UpdatedTime < 60_000 {
			return errors.New("运行时待同步: 多端口业务编辑正在进行")
		}
		cause := errors.New("恢复中断的多端口业务编辑")
		if err := h.rollbackIXRouteEdit(migration, cause); err != nil {
			return errors.New("运行时待同步: 多端口业务编辑回滚失败: " + err.Error())
		}
	}
	route, err := h.repo.GetIXRoute(routeID)
	if err != nil || route == nil {
		return errors.New("多端口业务不存在")
	}
	items, err := h.repo.ListIXRouteRuntimes(routeID)
	if err != nil {
		return err
	}
	if len(items) == 0 && route.DesiredState != "deleted" {
		return errors.New("多端口 Runtime desired state 缺失")
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Role == items[j].Role {
			return items[i].Protocol < items[j].Protocol
		}
		return items[i].Role == "ix"
	})
	for start := 0; start < len(items); {
		end := start + 1
		for end < len(items) && items[end].Role == items[start].Role && items[end].NodeID == items[start].NodeID {
			end++
		}
		group := items[start:end]
		var commandErr error
		if group[0].DesiredState == "active" {
			configs := make([]map[string]interface{}, 0, len(group))
			for _, item := range group {
				var config map[string]interface{}
				if err := json.Unmarshal([]byte(item.Config), &config); err != nil {
					commandErr = err
					break
				}
				configs = append(configs, config)
			}
			if commandErr == nil {
				_, commandErr = h.sendNodeCommand(group[0].NodeID, "UpdateService", configs, false, false)
			}
		} else {
			names := make([]string, 0, len(group))
			for _, item := range group {
				names = append(names, item.ServiceName)
			}
			_, commandErr = h.sendNodeCommand(group[0].NodeID, "DeleteService", map[string]interface{}{"services": names}, false, true)
		}
		if commandErr != nil {
			for _, item := range group {
				_ = h.repo.MarkIXRouteRuntimeObserved(item.ID, "error", commandErr.Error(), false)
			}
			state := "degraded"
			if shouldDeferTunnelRuntimeApplyError(commandErr) {
				state = "pending"
			}
			if route.DesiredState == "deleted" {
				state = "cleanup_pending"
			}
			_ = h.repo.SetIXRouteState(routeID, "", state, commandErr.Error())
			return &ixRouteReconcileError{role: group[0].Role, err: commandErr}
		}
		observed := "active"
		applied := true
		if group[0].DesiredState == "paused" {
			observed, applied = "paused", false
		}
		if group[0].DesiredState == "deleted" {
			observed, applied = "absent", false
		}
		for _, item := range group {
			_ = h.repo.MarkIXRouteRuntimeObserved(item.ID, observed, "", applied)
			_ = h.repo.UpsertIXPortObservation(&model.IXPortObservation{NodeID: item.NodeID, InternalPort: routePortForRole(route, item.Role), Protocol: item.Protocol, RuntimeName: item.ServiceName, State: observed})
		}
		start = end
	}
	switch route.DesiredState {
	case "paused":
		return h.repo.SetIXRouteState(routeID, "", "paused", "")
	case "deleted":
		for _, ownerType := range []string{ixRouteOwnerEntry, ixRouteOwnerIX} {
			if err := h.repo.CompleteIXPortLeaseRelease(ownerType, routeID); err != nil {
				_ = h.repo.SetIXRouteState(routeID, "", "cleanup_pending", err.Error())
				return err
			}
		}
		return h.repo.SetIXRouteState(routeID, "", "deleted", "")
	default:
		return h.repo.SetIXRouteState(routeID, "", "active", "")
	}
}

func routePortForRole(route *model.IXRoute, role string) int {
	if role == "ix" {
		return route.IXInternalPort
	}
	return route.EntryPort
}

func (h *Handler) reconcileIXRoutesOnNode(nodeID int64) error {
	ids, err := h.repo.ListIXRouteIDsByNode(nodeID)
	if err != nil {
		return err
	}
	var combined error
	for _, id := range ids {
		combined = errors.Join(combined, h.reconcileIXRoute(id))
	}
	return combined
}

func (h *Handler) setIXRouteDesired(id int64, desired string) error {
	route, err := h.repo.GetIXRoute(id)
	if err != nil || route == nil {
		return errors.New("多端口业务不存在")
	}
	if route.ObservedState == "deleted" {
		return errors.New("多端口业务已删除")
	}
	if err := h.repo.SetIXRouteState(id, desired, "pending", ""); err != nil {
		return err
	}
	if err := h.repo.SetIXRouteRuntimesDesired(id, desired); err != nil {
		return err
	}
	if desired == "deleted" {
		for _, ownerType := range []string{ixRouteOwnerEntry, ixRouteOwnerIX} {
			if err := h.repo.RequestIXPortLeaseRelease(ownerType, id); err != nil {
				return err
			}
		}
	}
	return h.reconcileIXRoute(id)
}

func (h *Handler) ixRouteMutation(w http.ResponseWriter, r *http.Request, desired string) {
	if r.Method != http.MethodPost || !requireIXPortAdmin(w, r) {
		return
	}
	var request ixRouteIDRequest
	if err := decodeJSON(r.Body, &request); err != nil || request.ID <= 0 {
		response.WriteJSON(w, response.Err(500, "业务不能为空"))
		return
	}
	err := h.setIXRouteDesired(request.ID, desired)
	if err != nil && !shouldDeferTunnelRuntimeApplyError(err) {
		response.WriteJSON(w, response.Err(-2, err.Error()))
		return
	}
	stored, _ := h.repo.GetIXRoute(request.ID)
	response.WriteJSON(w, response.OK(stored))
}

func (h *Handler) ixRoutePause(w http.ResponseWriter, r *http.Request) {
	h.ixRouteMutation(w, r, "paused")
}
func (h *Handler) ixRouteResume(w http.ResponseWriter, r *http.Request) {
	h.ixRouteMutation(w, r, "active")
}
func (h *Handler) ixRouteDelete(w http.ResponseWriter, r *http.Request) {
	h.ixRouteMutation(w, r, "deleted")
}

func (h *Handler) ixRouteReconcile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !requireIXPortAdmin(w, r) {
		return
	}
	var request ixRouteIDRequest
	if err := decodeJSON(r.Body, &request); err != nil || request.ID <= 0 {
		response.WriteJSON(w, response.Err(500, "业务不能为空"))
		return
	}
	if err := h.reconcileIXRoute(request.ID); err != nil {
		response.WriteJSON(w, response.Err(-2, err.Error()))
		return
	}
	stored, _ := h.repo.GetIXRoute(request.ID)
	response.WriteJSON(w, response.OK(stored))
}
