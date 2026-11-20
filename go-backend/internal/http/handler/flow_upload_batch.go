package handler

import (
	"errors"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-backend/internal/store/model"
	"go-backend/internal/store/repo"
)

type flowPolicyTarget struct {
	UserID       int64
	UserTunnelID int64
}

type flowUploadBatch struct {
	flowDeltas            []repo.FlowUploadCounterDelta
	quotaUsage            map[int64]int64
	policyTargets         []flowPolicyTarget
	forwardTraffic        map[int64]tunnelTrafficDelta
	orphanServices        map[string]struct{}
	peerShareForwardItems map[string]flowItem
	peerShareRuntimeItems map[int64]flowItem
	peerShareUsage        map[int64]int64
	ixRouteTraffic        map[int64]tunnelTrafficDelta
}

func (h *Handler) buildFlowUploadBatch(items []flowItem, metas map[int64]repo.FlowUploadForwardMeta) flowUploadBatch {
	batch := flowUploadBatch{
		quotaUsage:            make(map[int64]int64),
		forwardTraffic:        make(map[int64]tunnelTrafficDelta),
		orphanServices:        make(map[string]struct{}),
		peerShareForwardItems: make(map[string]flowItem),
		peerShareRuntimeItems: make(map[int64]flowItem),
		ixRouteTraffic:        make(map[int64]tunnelTrafficDelta),
	}
	policySeen := map[flowPolicyTarget]struct{}{}
	flowSeen := map[int64]int{}

	for _, item := range items {
		serviceName := strings.TrimSpace(item.N)
		if serviceName == "" || serviceName == "web_api" {
			continue
		}
		if routeID, ok := parseIXRouteEntryServiceID(serviceName); ok {
			if item.U >= 0 && item.D >= 0 {
				delta := batch.ixRouteTraffic[routeID]
				delta.bytesIn += item.D
				delta.bytesOut += item.U
				batch.ixRouteTraffic[routeID] = delta
			}
			continue
		}
		if runtimeID, ok := parsePeerShareRuntimeServiceID(serviceName); ok {
			merged := batch.peerShareRuntimeItems[runtimeID]
			merged.N = serviceName
			merged.U += item.U
			merged.D += item.D
			batch.peerShareRuntimeItems[runtimeID] = merged
			continue
		}
		forwardID, userID, userTunnelID, ok := parseFlowServiceIDs(serviceName)
		if !ok {
			continue
		}
		normalized := normalizeForwardRuntimeServiceName(serviceName)
		merged := batch.peerShareForwardItems[normalized]
		merged.N = normalized
		merged.U += item.U
		merged.D += item.D
		batch.peerShareForwardItems[normalized] = merged

		meta, exists := metas[forwardID]
		if !exists {
			batch.orphanServices[serviceName] = struct{}{}
			continue
		}
		// Local accounting uses database ownership, not foreign IDs embedded in
		// a service name (including stale user-tunnel IDs after reassignment).
		userID, userTunnelID = meta.UserID, meta.UserTunnelID

		raw := batch.forwardTraffic[forwardID]
		raw.bytesIn += item.D
		raw.bytesOut += item.U
		batch.forwardTraffic[forwardID] = raw

		scaledIn := int64(float64(item.D)*meta.TrafficRatio) * meta.TunnelFlow
		scaledOut := int64(float64(item.U)*meta.TrafficRatio) * meta.TunnelFlow
		if idx, ok := flowSeen[forwardID]; ok {
			batch.flowDeltas[idx].InFlow += scaledIn
			batch.flowDeltas[idx].OutFlow += scaledOut
		} else {
			flowSeen[forwardID] = len(batch.flowDeltas)
			batch.flowDeltas = append(batch.flowDeltas, repo.FlowUploadCounterDelta{
				ForwardID:    forwardID,
				UserID:       userID,
				UserTunnelID: userTunnelID,
				InFlow:       scaledIn,
				OutFlow:      scaledOut,
			})
		}
		batch.quotaUsage[userID] += scaledIn + scaledOut

		target := flowPolicyTarget{UserID: userID, UserTunnelID: userTunnelID}
		if _, seen := policySeen[target]; !seen {
			policySeen[target] = struct{}{}
			batch.policyTargets = append(batch.policyTargets, target)
		}

	}

	sort.Slice(batch.policyTargets, func(i, j int) bool {
		if batch.policyTargets[i].UserID == batch.policyTargets[j].UserID {
			return batch.policyTargets[i].UserTunnelID < batch.policyTargets[j].UserTunnelID
		}
		return batch.policyTargets[i].UserID < batch.policyTargets[j].UserID
	})

	return batch
}

func (h *Handler) applyFlowUploadBatch(nodeID int64, batch flowUploadBatch, now time.Time) {
	if h == nil || h.repo == nil {
		return
	}
	h.applyFlowDeltasWithFallback(nodeID, batch.flowDeltas)
	h.applyIXRouteFlowDeltas(nodeID, batch.ixRouteTraffic)
	quotaViews := h.applyQuotaUsageWithFallback(nodeID, batch.quotaUsage, now)
	h.applyFlowUploadPostCommit(nodeID, batch, quotaViews, true)
}

func (h *Handler) applyDurableFlowUploadBatch(nodeID int64, reportID string, batch flowUploadBatch, now time.Time) (bool, error) {
	if h == nil || h.repo == nil {
		return false, errors.New("flow repository unavailable")
	}
	applied, quotaViews, err := h.repo.ApplyDurableFlowReport(
		nodeID,
		reportID,
		batch.flowDeltas,
		ixRouteFlowDeltas(batch.ixRouteTraffic),
		batch.quotaUsage,
		batch.peerShareUsage,
		now,
	)
	if err != nil || !applied {
		return applied, err
	}
	h.applyFlowUploadPostCommit(nodeID, batch, quotaViews, false)
	return true, nil
}

func (h *Handler) applyFlowUploadPostCommit(nodeID int64, batch flowUploadBatch, quotaViews map[int64]*model.UserQuotaView, persistPeerShareUsage bool) {
	for userID, quota := range quotaViews {
		h.enforceUserQuotaIfNeeded(userID, quota)
	}
	for _, target := range batch.policyTargets {
		if target.UserID <= 0 || target.UserTunnelID <= 0 {
			continue
		}
		h.enforceFlowPolicies(target.UserID, target.UserTunnelID)
	}
	if len(batch.orphanServices) > 0 {
		serviceNames := make([]string, 0, len(batch.orphanServices))
		for serviceName := range batch.orphanServices {
			serviceNames = append(serviceNames, serviceName)
		}
		h.sendDeleteOrphanedForwardServices(nodeID, serviceNames)
	}
	for serviceName, item := range batch.peerShareForwardItems {
		forwardID, _, _, ok := parseFlowServiceIDs(serviceName)
		if ok {
			h.processPeerShareFlowFromForward(forwardID, nodeID, serviceName, item)
		}
	}
	for runtimeID, item := range batch.peerShareRuntimeItems {
		h.processPeerShareFlow(nodeID, runtimeID, item)
	}
	for shareID, delta := range batch.peerShareUsage {
		if persistPeerShareUsage {
			h.addPeerShareFlow(nodeID, shareID, delta)
			continue
		}
		share, err := h.repo.GetPeerShare(shareID)
		if err == nil && share != nil && share.NodeID == nodeID && isPeerShareFlowExceeded(share) {
			h.enforcePeerShareFlowLimit(shareID)
		}
	}
}

func parseIXRouteEntryServiceID(name string) (int64, bool) {
	name = strings.TrimSpace(name)
	const prefix = "dx3fv_ix_"
	if !strings.HasPrefix(name, prefix) {
		return 0, false
	}
	remainder := strings.TrimPrefix(name, prefix)
	var suffix string
	for _, candidate := range []string{"_entry_tcp", "_entry_udp"} {
		if strings.HasSuffix(remainder, candidate) {
			suffix = candidate
			break
		}
	}
	if suffix == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(strings.TrimSuffix(remainder, suffix), 10, 64)
	return id, err == nil && id > 0
}

func (h *Handler) applyIXRouteFlowDeltas(nodeID int64, traffic map[int64]tunnelTrafficDelta) {
	if h == nil || h.repo == nil || nodeID <= 0 || len(traffic) == 0 {
		return
	}
	if err := h.repo.AddIXRouteFlowBatch(nodeID, ixRouteFlowDeltas(traffic)); err != nil {
		log.Printf("IX route flow upload failed node_id=%d err=%v", nodeID, err)
	}
}

func ixRouteFlowDeltas(traffic map[int64]tunnelTrafficDelta) []repo.IXRouteFlowDelta {
	ids := make([]int64, 0, len(traffic))
	for id := range traffic {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	deltas := make([]repo.IXRouteFlowDelta, 0, len(ids))
	for _, id := range ids {
		delta := traffic[id]
		deltas = append(deltas, repo.IXRouteFlowDelta{RouteID: id, InFlow: delta.bytesIn, OutFlow: delta.bytesOut})
	}
	return deltas
}

func (h *Handler) applyFlowDeltasWithFallback(nodeID int64, deltas []repo.FlowUploadCounterDelta) {
	if h == nil || h.repo == nil || len(deltas) == 0 {
		return
	}
	if err := h.repo.ApplyFlowUploadDeltasBatch(deltas); err == nil {
		return
	} else {
		log.Printf("flow upload write failed op=flow.batch_apply node_id=%d err=%v", nodeID, err)
	}
	for _, delta := range deltas {
		if err := h.repo.AddFlow(delta.ForwardID, delta.UserID, delta.UserTunnelID, delta.InFlow, delta.OutFlow); err != nil {
			log.Printf("flow upload write failed op=flow.single_apply node_id=%d forward_id=%d user_id=%d user_tunnel_id=%d err=%v", nodeID, delta.ForwardID, delta.UserID, delta.UserTunnelID, err)
		}
	}
}

func (h *Handler) applyQuotaUsageWithFallback(nodeID int64, usages map[int64]int64, now time.Time) map[int64]*model.UserQuotaView {
	if h == nil || h.repo == nil || len(usages) == 0 {
		return map[int64]*model.UserQuotaView{}
	}
	quotaViews, err := h.repo.AddUserQuotaUsageBatch(usages, now)
	if err == nil {
		return quotaViews
	}
	log.Printf("flow upload write failed op=quota.batch_apply node_id=%d err=%v", nodeID, err)

	userIDs := make([]int64, 0, len(usages))
	for userID := range usages {
		if userID > 0 {
			userIDs = append(userIDs, userID)
		}
	}
	sort.Slice(userIDs, func(i, j int) bool { return userIDs[i] < userIDs[j] })

	quotaViews = make(map[int64]*model.UserQuotaView, len(userIDs))
	for _, userID := range userIDs {
		quota, singleErr := h.repo.AddUserQuotaUsage(userID, usages[userID], now)
		if singleErr != nil {
			log.Printf("flow upload write failed op=quota.single_apply node_id=%d user_id=%d err=%v", nodeID, userID, singleErr)
			continue
		}
		if quota != nil {
			quotaViews[userID] = quota
		}
	}
	return quotaViews
}
