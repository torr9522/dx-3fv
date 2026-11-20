package handler

import (
	crand "crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-backend/internal/http/response"
	"go-backend/internal/store/model"
)

const (
	ixProbePass       = "PASS"
	ixProbeFail       = "FAIL"
	ixProbeUnverified = "UNVERIFIED"
)

type ixRouteDiagnosisRequest struct {
	ID          int64 `json:"id"`
	PayloadEcho bool  `json:"payloadEcho"`
}

type ixRouteProbeLeg struct {
	Protocol  string  `json:"protocol"`
	Leg       string  `json:"leg"`
	From      string  `json:"from"`
	To        string  `json:"to"`
	Method    string  `json:"method"`
	Status    string  `json:"status"`
	LatencyMS float64 `json:"latencyMs"`
	Loss      float64 `json:"loss"`
	Error     string  `json:"error,omitempty"`
}

type ixRouteDiagnosisResult struct {
	ID            int64             `json:"id"`
	RouteID       int64             `json:"routeId"`
	ProtocolMode  string            `json:"protocolMode"`
	PayloadEcho   bool              `json:"payloadEcho"`
	OverallStatus string            `json:"overallStatus"`
	StartedTime   int64             `json:"startedTime"`
	CompletedTime int64             `json:"completedTime"`
	Legs          []ixRouteProbeLeg `json:"legs"`
}

func numberFromCommand(data map[string]interface{}, key string) float64 {
	if data == nil {
		return 0
	}
	switch value := data[key].(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	case int:
		return float64(value)
	case int64:
		return float64(value)
	default:
		return 0
	}
}

func boolFromCommand(data map[string]interface{}, key string) bool {
	value, _ := data[key].(bool)
	return value
}

func stringFromCommand(data map[string]interface{}, key string) string {
	value, _ := data[key].(string)
	return strings.TrimSpace(value)
}

func unverifiedIXLeg(protocol, leg, from, to, method, reason string) ixRouteProbeLeg {
	return ixRouteProbeLeg{Protocol: protocol, Leg: leg, From: from, To: to, Method: method, Status: ixProbeUnverified, Error: reason}
}

func (h *Handler) tcpConnectIXLeg(nodeID int64, leg, from, to, host string, port int) ixRouteProbeLeg {
	item := ixRouteProbeLeg{Protocol: "tcp", Leg: leg, From: from, To: to, Method: "tcp_connect"}
	data, err := h.tcpPingViaNode(nodeID, host, port, diagnosisExecOptions{pingCount: 1, pingTimeoutMS: 5000, commandTimeout: 8 * time.Second})
	if err != nil {
		item.Status, item.Error = ixProbeFail, err.Error()
		return item
	}
	item.LatencyMS = numberFromCommand(data, "averageTime")
	item.Loss = numberFromCommand(data, "packetLoss")
	if boolFromCommand(data, "success") {
		item.Status = ixProbePass
	} else {
		item.Status = ixProbeFail
		item.Error = stringFromCommand(data, "errorMessage")
		if item.Error == "" {
			item.Error = "TCP connect failed"
		}
	}
	return item
}

func (h *Handler) echoIXLeg(nodeID int64, protocol, leg, from, to, host string, port int, nonce string) ixRouteProbeLeg {
	method := protocol + "_nonce_echo"
	item := ixRouteProbeLeg{Protocol: protocol, Leg: leg, From: from, To: to, Method: method}
	res, err := h.sendNodeCommandWithTimeout(nodeID, "EchoProbe", map[string]interface{}{
		"protocol": protocol, "ip": host, "port": port, "payload": nonce, "timeout": 5000,
	}, 8*time.Second, false, false)
	if err != nil {
		item.Status, item.Error = ixProbeFail, err.Error()
		return item
	}
	item.LatencyMS = numberFromCommand(res.Data, "latencyMs")
	if boolFromCommand(res.Data, "success") {
		item.Status = ixProbePass
	} else {
		// A raw nonce timeout against an arbitrary business service only proves
		// that the target did not behave as an echo server. It does not prove
		// that the forwarding path is down.
		item.Status = ixProbeUnverified
		item.Error = stringFromCommand(res.Data, "errorMessage")
		if item.Error == "" {
			item.Error = "nonce payload did not return"
		}
		item.Error = "受控 Echo 未返回 nonce，无法据此判定业务链路失败: " + item.Error
	}
	return item
}

func ixLocalProbeHost(node *model.Node, protocol string) string {
	if node == nil {
		return "127.0.0.1"
	}
	host := node.TCPListenAddr
	if protocol == "udp" {
		host = node.UDPListenAddr
	}
	host = strings.Trim(strings.TrimSpace(host), "[]")
	switch host {
	case "", "0.0.0.0", "*":
		return "127.0.0.1"
	case "::":
		return "::1"
	default:
		return host
	}
}

func overallIXProbeStatus(legs []ixRouteProbeLeg) string {
	overall := ixProbePass
	for _, leg := range legs {
		if leg.Status == ixProbeFail {
			return ixProbeFail
		}
		if leg.Status == ixProbeUnverified {
			overall = ixProbeUnverified
		}
	}
	return overall
}

func (h *Handler) activeIXRouteResource(routeID int64) (*model.IXPortResource, error) {
	leases, err := h.repo.ListIXPortLeases(ixRouteOwnerIX, routeID)
	if err != nil {
		return nil, err
	}
	for i := len(leases) - 1; i >= 0; i-- {
		if leases[i].State != "active" {
			continue
		}
		resource, err := h.repo.GetIXPortResource(leases[i].ResourceID)
		if err != nil {
			return nil, err
		}
		if resource != nil {
			return resource, nil
		}
	}
	return nil, errors.New("IX active NAT resource is unavailable")
}

func (h *Handler) runIXRouteDiagnosis(routeID int64, payloadEcho bool) (*ixRouteDiagnosisResult, error) {
	route, err := h.repo.GetIXRoute(routeID)
	if err != nil || route == nil {
		return nil, errors.New("多端口业务不存在")
	}
	if route.DesiredState != "active" || route.ObservedState != "active" {
		return nil, errors.New("只有已生效的多端口业务可以诊断")
	}
	profile, err := h.repo.GetIXPathProfile(route.PathProfileID)
	if err != nil || profile == nil {
		return nil, errors.New("前置入口到 IX 路径不存在")
	}
	entryNode, err := h.repo.GetNodeByID(profile.EntryNodeID)
	if err != nil || entryNode == nil {
		return nil, errors.New("前置入口节点不存在")
	}
	ixNode, err := h.repo.GetNodeByID(profile.IXNodeID)
	if err != nil || ixNode == nil {
		return nil, errors.New("IX 节点不存在")
	}
	resource, err := h.activeIXRouteResource(route.ID)
	if err != nil {
		return nil, err
	}
	if resource.NodeID != profile.IXNodeID || resource.PoolID != profile.IXPoolID ||
		resource.PublicPort != route.IXPublicPort || resource.InternalPort != route.IXInternalPort {
		return nil, errors.New("IX active NAT resource does not match the current business route")
	}
	landingHost := strings.TrimSpace(route.LandingHost)
	if landingHost == "" || route.LandingPort < 1 || route.LandingPort > 65535 {
		return nil, errors.New("Landing diagnosis target is invalid")
	}
	ixPublicHost := strings.TrimSpace(resource.PublicIP)
	if ixPublicHost == "" {
		ixPublicHost = strings.TrimSpace(ixNode.ServerIP)
	}
	if ixPublicHost == "" || route.IXPublicPort < 1 || route.IXPublicPort > 65535 {
		return nil, errors.New("IX diagnosis target is invalid")
	}
	started := time.Now().UnixMilli()
	result := &ixRouteDiagnosisResult{RouteID: route.ID, ProtocolMode: route.ProtocolMode, PayloadEcho: payloadEcho, StartedTime: started}
	protocols, _ := ixRouteProtocols(route.ProtocolMode)
	nonceBytes := make([]byte, 24)
	if _, err := crand.Read(nonceBytes); err != nil {
		return nil, err
	}
	nonce := "dx3fv-" + hex.EncodeToString(nonceBytes)
	for _, protocol := range protocols {
		if protocol == "tcp" {
			result.Legs = append(result.Legs,
				h.tcpConnectIXLeg(profile.EntryNodeID, "gz_to_ix", entryNode.Name, ixNode.Name, ixPublicHost, route.IXPublicPort),
				h.tcpConnectIXLeg(profile.IXNodeID, "ix_to_landing", ixNode.Name, "Landing", landingHost, route.LandingPort),
			)
			if payloadEcho {
				result.Legs = append(result.Legs, h.echoIXLeg(profile.EntryNodeID, "tcp", "end_to_end", entryNode.Name, "Landing", ixLocalProbeHost(entryNode, "tcp"), route.EntryPort, nonce))
			} else {
				result.Legs = append(result.Legs, unverifiedIXLeg("tcp", "end_to_end", entryNode.Name, "Landing", "controlled_payload_required", "未启用受控 Echo，TCP Connect 不能替代端到端 payload"))
			}
			continue
		}
		if !payloadEcho {
			result.Legs = append(result.Legs,
				unverifiedIXLeg("udp", "gz_to_ix", entryNode.Name, ixNode.Name, "udp_nonce_echo_required", "UDP connect 不证明 payload 可达"),
				unverifiedIXLeg("udp", "ix_to_landing", ixNode.Name, "Landing", "udp_nonce_echo_required", "UDP connect 不证明 payload 可达"),
				unverifiedIXLeg("udp", "end_to_end", entryNode.Name, "Landing", "udp_nonce_echo_required", "未启用受控 UDP Echo"),
			)
			continue
		}
		result.Legs = append(result.Legs,
			unverifiedIXLeg("udp", "gz_to_ix", entryNode.Name, ixNode.Name, "udp_segment_echo_unavailable", "IX 业务 Listener 会继续转发到 Landing，无独立 UDP Echo 响应时不能分段验证"),
			h.echoIXLeg(profile.IXNodeID, "udp", "ix_to_landing", ixNode.Name, "Landing", landingHost, route.LandingPort, nonce),
			h.echoIXLeg(profile.EntryNodeID, "udp", "end_to_end", entryNode.Name, "Landing", ixLocalProbeHost(entryNode, "udp"), route.EntryPort, nonce),
		)
	}
	result.OverallStatus = overallIXProbeStatus(result.Legs)
	result.CompletedTime = time.Now().UnixMilli()
	stored := &model.IXRouteDiagnosis{RouteID: route.ID, ProtocolMode: route.ProtocolMode, PayloadEcho: payloadEcho, OverallStatus: result.OverallStatus, StartedTime: result.StartedTime, CompletedTime: result.CompletedTime}
	if err := h.repo.SaveIXRouteDiagnosis(stored, result.Legs); err != nil {
		return nil, err
	}
	result.ID = stored.ID
	return result, nil
}

func (h *Handler) ixRouteDiagnose(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !requireIXPortAdmin(w, r) {
		return
	}
	var request ixRouteDiagnosisRequest
	if err := decodeJSON(r.Body, &request); err != nil || request.ID <= 0 {
		response.WriteJSON(w, response.Err(500, "业务不能为空"))
		return
	}
	result, err := h.runIXRouteDiagnosis(request.ID, request.PayloadEcho)
	if err != nil {
		response.WriteJSON(w, response.Err(-2, err.Error()))
		return
	}
	response.WriteJSON(w, response.OK(result))
}

func (h *Handler) ixRouteDiagnosisLatest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet || !requireIXPortAdmin(w, r) {
		return
	}
	id, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("id")), 10, 64)
	if err != nil || id <= 0 {
		response.WriteJSON(w, response.Err(500, "业务不能为空"))
		return
	}
	item, err := h.repo.GetLatestIXRouteDiagnosis(id)
	if err != nil {
		response.WriteJSON(w, response.Err(-2, err.Error()))
		return
	}
	response.WriteJSON(w, response.OK(item))
}
