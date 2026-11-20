package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"go-backend/internal/security"
	"go-backend/internal/store/model"
)

func connectAdditionalIXTestNode(t *testing.T, agent *cleanupAgent, nodeID int64, secret string) {
	t.Helper()
	if err := agent.h.repo.DB().Create(&model.Node{ID: nodeID, Name: "ix", Secret: secret, ServerIP: "127.0.0.1", Port: "38000-38099", Status: 1, TCPListenAddr: "[::]", UDPListenAddr: "[::]"}).Error; err != nil {
		t.Fatal(err)
	}
	online := make(chan struct{})
	agent.h.wsServer.SetNodeOnlineHook(func(onlineID int64) {
		if onlineID == nodeID {
			select {
			case <-online:
			default:
				close(online)
			}
		}
	})
	done := make(chan struct{})
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		defer close(done)
		agent.h.wsServer.ServeHTTP(w, request)
	}))
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(ts.URL, "http")+"?type=1&secret="+secret, nil)
	if err != nil {
		ts.Close()
		t.Fatal(err)
	}
	crypto, err := security.NewAESCrypto(secret)
	if err != nil {
		t.Fatal(err)
	}
	readerDone := make(chan struct{})
	go func() {
		defer close(readerDone)
		for {
			_, payload, err := conn.ReadMessage()
			if err != nil {
				return
			}
			var envelope struct {
				Encrypted bool   `json:"encrypted"`
				Data      string `json:"data"`
			}
			if err := json.Unmarshal(payload, &envelope); err != nil {
				return
			}
			if envelope.Encrypted {
				payload, err = crypto.Decrypt(envelope.Data)
				if err != nil {
					return
				}
			}
			var command cleanupAgentCommand
			if err := json.Unmarshal(payload, &command); err != nil {
				return
			}
			command.NodeID = nodeID
			agent.mu.Lock()
			agent.commands = append(agent.commands, command)
			hook := agent.responseHook
			dataHook := agent.responseDataHook
			agent.mu.Unlock()
			success, message := true, ""
			if hook != nil {
				success, message = hook(command)
			}
			var data map[string]interface{}
			if dataHook != nil {
				data = dataHook(command)
			}
			if err := conn.WriteJSON(map[string]interface{}{"type": command.Type, "requestId": command.RequestID, "success": success, "message": message, "data": data}); err != nil {
				return
			}
		}
	}()
	t.Cleanup(func() {
		_ = conn.Close()
		<-readerDone
		ts.Close()
		<-done
	})
	select {
	case <-online:
	case <-time.After(3 * time.Second):
		t.Fatal("additional IX test node did not come online")
	}
}

func setupIXRouteControlPlane(t *testing.T) (*cleanupAgent, *model.IXPathProfile) {
	t.Helper()
	agent := newCleanupAgent(t)
	connectAdditionalIXTestNode(t, agent, 2, "ix-route-node-two")
	entryPool := &model.IXPortPool{NodeID: 1, Name: "gz2", Kind: "full", RangeStart: 1, RangeEnd: 65535, NextPort: 24000, Status: 1}
	if err := agent.h.repo.CreateIXPortPool(entryPool); err != nil {
		t.Fatal(err)
	}
	ixPool := &model.IXPortPool{NodeID: 2, Name: "ix2", Kind: "nat", RangeStart: 38000, RangeEnd: 38099, NextPort: 38001, Status: 1}
	if err := agent.h.repo.CreateIXPortPool(ixPool); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 2; i++ {
		resource := &model.IXPortResource{PoolID: ixPool.ID, NodeID: 2, PublicIP: "127.0.0.1", PublicPort: 38000 + i, InternalPort: 30000 + i, TCPMapped: true, UDPMapped: true, VerificationStatus: "verified"}
		if err := agent.h.repo.SaveIXPortResource(resource); err != nil {
			t.Fatal(err)
		}
	}
	profile := &model.IXPathProfile{Name: "gz2-ix2", EntryNodeID: 1, IXNodeID: 2, EntryPoolID: entryPool.ID, IXPoolID: ixPool.ID}
	if err := agent.h.repo.CreateIXPathProfile(profile); err != nil {
		t.Fatal(err)
	}
	return agent, profile
}

func createIXRouteThroughHandler(t *testing.T, agent *cleanupAgent, profileID int64, name string, entryPort, ixPort int) *model.IXRoute {
	t.Helper()
	recorder := httptest.NewRecorder()
	request := ixPortRequest(t, http.MethodPost, "/api/v1/ix-route/create", map[string]interface{}{
		"pathProfileId": profileID, "name": name, "entryPort": entryPort, "ixPort": ixPort,
		"landingHost": "127.0.0.1", "landingPort": 8080, "protocolMode": "tcp_udp",
	}, 0)
	agent.h.ixRouteCreate(recorder, request)
	if code := ixPortResponseCode(t, recorder); code != 0 {
		t.Fatalf("route create failed code=%d body=%s", code, recorder.Body.String())
	}
	var payload struct {
		Data model.IXRoute `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	return &payload.Data
}

func TestIXRouteScopedCreateRejectsCrossPathTamperingWithoutResidue(t *testing.T) {
	agent, profile := setupIXRouteControlPlane(t)
	scopeID := profile.ID + 1000
	recorder := httptest.NewRecorder()
	request := ixPortRequest(t, http.MethodPost, "/api/v1/ix-route/create", map[string]interface{}{
		"pathProfileId": profile.ID, "scopePathProfileId": scopeID,
		"name": "tampered", "entryPort": 0, "ixPort": 0,
		"landingHost": "127.0.0.1", "landingPort": 8080, "protocolMode": "tcp",
	}, 0)
	agent.h.ixRouteCreate(recorder, request)
	if code := ixPortResponseCode(t, recorder); code == 0 {
		t.Fatalf("cross-path scoped create unexpectedly succeeded: %s", recorder.Body.String())
	}
	routes, err := agent.h.repo.ListIXRoutes()
	if err != nil {
		t.Fatal(err)
	}
	if len(routes) != 0 {
		t.Fatalf("rejected scoped create left routes: %+v", routes)
	}
	leases, err := agent.h.repo.ListIXPortLeases("", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(leases) != 0 {
		t.Fatalf("rejected scoped create left leases: %+v", leases)
	}
	var runtimeCount int64
	if err := agent.h.repo.DB().Model(&model.IXRouteRuntime{}).Count(&runtimeCount).Error; err != nil {
		t.Fatal(err)
	}
	if runtimeCount != 0 {
		t.Fatalf("rejected scoped create left %d runtimes", runtimeCount)
	}
}

func mutateIXRouteThroughHandler(t *testing.T, handler func(http.ResponseWriter, *http.Request), path string, id int64) {
	t.Helper()
	recorder := httptest.NewRecorder()
	handler(recorder, ixPortRequest(t, http.MethodPost, path, map[string]interface{}{"id": id}, 0))
	if code := ixPortResponseCode(t, recorder); code != 0 {
		t.Fatalf("route mutation failed code=%d body=%s", code, recorder.Body.String())
	}
}

func TestIXRouteTCPUDPOrchestrationAndLifecycle(t *testing.T) {
	agent, profile := setupIXRouteControlPlane(t)
	first := createIXRouteThroughHandler(t, agent, profile.ID, "first", 24001, 38001)
	if first.ObservedState != "active" {
		t.Fatalf("created route state = %s, want active", first.ObservedState)
	}
	runtimes, err := agent.h.repo.ListIXRouteRuntimes(first.ID)
	if err != nil || len(runtimes) != 4 {
		t.Fatalf("TCP+UDP should own four runtime lanes: count=%d err=%v", len(runtimes), err)
	}
	updates := agent.commandsOfType("UpdateService")
	if len(updates) != 2 {
		t.Fatalf("create sent %d UpdateService commands, want IX then entry", len(updates))
	}
	if !strings.Contains(string(updates[0].Data), "_ix_tcp") || !strings.Contains(string(updates[1].Data), "_entry_tcp") {
		t.Fatalf("runtime apply order is not IX then entry: %s / %s", updates[0].Data, updates[1].Data)
	}

	mutateIXRouteThroughHandler(t, agent.h.ixRoutePause, "/api/v1/ix-route/pause", first.ID)
	leasing, err := agent.h.repo.ListIXPortLeases("", first.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, lease := range leasing {
		if lease.State != "active" {
			t.Fatalf("pause released lease %+v", lease)
		}
	}
	mutateIXRouteThroughHandler(t, agent.h.ixRouteResume, "/api/v1/ix-route/resume", first.ID)
	resumed, _ := agent.h.repo.GetIXRoute(first.ID)
	if resumed.ObservedState != "active" {
		t.Fatalf("resume state = %s", resumed.ObservedState)
	}
}

func TestIXRouteDeleteIsIsolatedAndReleasesAfterRuntimeAck(t *testing.T) {
	agent, profile := setupIXRouteControlPlane(t)
	first := createIXRouteThroughHandler(t, agent, profile.ID, "first", 24001, 38001)
	second := createIXRouteThroughHandler(t, agent, profile.ID, "second", 24002, 38002)
	mutateIXRouteThroughHandler(t, agent.h.ixRouteDelete, "/api/v1/ix-route/delete", first.ID)

	deleted, _ := agent.h.repo.GetIXRoute(first.ID)
	remaining, _ := agent.h.repo.GetIXRoute(second.ID)
	if deleted.ObservedState != "deleted" || remaining.ObservedState != "active" {
		t.Fatalf("delete isolation failed: deleted=%s remaining=%s", deleted.ObservedState, remaining.ObservedState)
	}
	firstLeases, _ := agent.h.repo.ListIXPortLeases("", first.ID)
	secondLeases, _ := agent.h.repo.ListIXPortLeases("", second.ID)
	for _, lease := range firstLeases {
		if lease.State != "released" {
			t.Fatalf("deleted route retained lease after listener ACK: %+v", lease)
		}
	}
	for _, lease := range secondLeases {
		if lease.State != "active" {
			t.Fatalf("sibling route lease changed: %+v", lease)
		}
	}
	deletes := agent.commandsOfType("DeleteService")
	for _, command := range deletes {
		if strings.Contains(string(command.Data), fmt.Sprintf("dx3fv_ix_%d", second.ID)) {
			t.Fatalf("deleting route %d touched sibling %d: %s", first.ID, second.ID, command.Data)
		}
	}
}

func TestIXRouteAutomaticPortRetriesRuntimeBindConflict(t *testing.T) {
	agent, profile := setupIXRouteControlPlane(t)
	failed := false
	agent.responseHook = func(command cleanupAgentCommand) (bool, string) {
		if command.Type == "UpdateService" && strings.Contains(string(command.Data), "_ix_tcp") && !failed {
			failed = true
			return false, "listen tcp: bind: address already in use"
		}
		return true, ""
	}
	route := createIXRouteThroughHandler(t, agent, profile.ID, "retry", 24001, 0)
	if !failed {
		t.Fatal("test did not inject a Runtime bind conflict")
	}
	if route.IXPublicPort != 38002 || route.ObservedState != "active" {
		t.Fatalf("automatic allocation did not move to the next verified mapping: %+v", route)
	}
	leasing, err := agent.h.repo.ListIXPortLeases(ixRouteOwnerIX, route.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(leasing) != 4 || leasing[0].State != "released" || leasing[1].State != "released" || leasing[2].State != "active" || leasing[3].State != "active" {
		t.Fatalf("retry lease generations are incorrect: %+v", leasing)
	}
}

func TestIXRouteManualPortConflictFailsWithoutSubstitution(t *testing.T) {
	agent, profile := setupIXRouteControlPlane(t)
	agent.responseHook = func(command cleanupAgentCommand) (bool, string) {
		if command.Type == "UpdateService" && strings.Contains(string(command.Data), "_ix_tcp") {
			return false, "listen tcp: bind: address already in use"
		}
		return true, ""
	}
	recorder := httptest.NewRecorder()
	agent.h.ixRouteCreate(recorder, ixPortRequest(t, http.MethodPost, "/api/v1/ix-route/create", map[string]interface{}{
		"pathProfileId": profile.ID, "name": "manual-conflict", "entryPort": 24001, "ixPort": 38001,
		"landingHost": "127.0.0.1", "landingPort": 8080, "protocolMode": "tcp_udp",
	}, 0))
	if code := ixPortResponseCode(t, recorder); code == 0 {
		t.Fatalf("manual conflict was silently accepted: %s", recorder.Body.String())
	}
	routes, err := agent.h.repo.ListIXRoutes()
	if err != nil {
		t.Fatal(err)
	}
	if len(routes) != 0 {
		t.Fatalf("failed manual route remained active: %+v", routes)
	}
	leasing, err := agent.h.repo.ListIXPortLeases("", 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, lease := range leasing {
		if lease.State != "released" {
			t.Fatalf("manual conflict leaked lease: %+v", lease)
		}
	}
}

func TestIXRouteTrafficCountsOnlyOwnedEntryLane(t *testing.T) {
	agent, profile := setupIXRouteControlPlane(t)
	route := createIXRouteThroughHandler(t, agent, profile.ID, "traffic", 12100, 38001)

	entryTCP := fmt.Sprintf("dx3fv_ix_%d_entry_tcp", route.ID)
	ixTCP := fmt.Sprintf("dx3fv_ix_%d_ix_tcp", route.ID)
	batch := agent.h.buildNodeFlowUploadBatch(profile.EntryNodeID, []flowItem{
		{N: entryTCP, U: 120, D: 80},
		{N: ixTCP, U: 900, D: 700},
	}, nil)
	agent.h.applyFlowUploadBatch(profile.EntryNodeID, batch, time.Now())

	stored, err := agent.h.repo.GetIXRoute(route.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.InFlow != 80 || stored.OutFlow != 120 {
		t.Fatalf("entry traffic mismatch: in=%d out=%d", stored.InFlow, stored.OutFlow)
	}

	spoofed := agent.h.buildNodeFlowUploadBatch(profile.IXNodeID, []flowItem{{N: entryTCP, U: 500, D: 400}}, nil)
	agent.h.applyFlowUploadBatch(profile.IXNodeID, spoofed, time.Now())
	stored, err = agent.h.repo.GetIXRoute(route.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.InFlow != 80 || stored.OutFlow != 120 {
		t.Fatalf("non-entry node changed route traffic: in=%d out=%d", stored.InFlow, stored.OutFlow)
	}
}

func TestIXRouteDiagnosisDoesNotMisreportUDPConnectAsPayloadSuccess(t *testing.T) {
	agent, profile := setupIXRouteControlPlane(t)
	route := createIXRouteThroughHandler(t, agent, profile.ID, "diagnosis", 24001, 38001)
	agent.responseDataHook = func(command cleanupAgentCommand) map[string]interface{} {
		if command.Type == "TcpPing" {
			return map[string]interface{}{"success": true, "averageTime": 12.5, "packetLoss": 0.0}
		}
		return nil
	}
	result, err := agent.h.runIXRouteDiagnosis(route.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	if result.OverallStatus != ixProbeUnverified || len(result.Legs) != 6 {
		t.Fatalf("unexpected diagnosis summary: %+v", result)
	}
	for _, leg := range result.Legs {
		if leg.Protocol == "udp" && leg.Status != ixProbeUnverified {
			t.Fatalf("UDP connect was promoted to payload proof: %+v", leg)
		}
	}
	if len(agent.commandsOfType("UdpPing")) != 0 {
		t.Fatal("diagnosis called legacy UdpPing even though it cannot prove payload")
	}
	stored, err := agent.h.repo.GetLatestIXRouteDiagnosis(route.ID)
	if err != nil || stored == nil || stored.Diagnosis.OverallStatus != ixProbeUnverified {
		t.Fatalf("diagnosis evidence was not persisted: %+v %v", stored, err)
	}
}

func TestIXRouteDiagnosisRecordsNonceEchoWithoutInventingUDPSegmentEvidence(t *testing.T) {
	agent, profile := setupIXRouteControlPlane(t)
	route := createIXRouteThroughHandler(t, agent, profile.ID, "echo-diagnosis", 24001, 38001)
	agent.responseDataHook = func(command cleanupAgentCommand) map[string]interface{} {
		switch command.Type {
		case "TcpPing":
			return map[string]interface{}{"success": true, "averageTime": 8.0, "packetLoss": 0.0}
		case "EchoProbe":
			return map[string]interface{}{"success": true, "latencyMs": 15.0, "bytes": 55.0}
		default:
			return nil
		}
	}
	result, err := agent.h.runIXRouteDiagnosis(route.ID, true)
	if err != nil {
		t.Fatal(err)
	}
	if result.OverallStatus != ixProbeUnverified || len(result.Legs) != 6 {
		t.Fatalf("unexpected nonce diagnosis result: %+v", result)
	}
	if got := len(agent.commandsOfType("EchoProbe")); got != 3 {
		t.Fatalf("sent %d payload probes, want TCP end-to-end plus two evidence-bearing UDP probes", got)
	}
	if leg := result.Legs[3]; leg.Protocol != "udp" || leg.Leg != "gz_to_ix" || leg.Status != ixProbeUnverified || leg.Method != "udp_segment_echo_unavailable" {
		t.Fatalf("UDP first segment was presented as independently measured: %+v", leg)
	}
}

func TestIXRouteDiagnosisDoesNotTreatMissingEchoAsBusinessFailure(t *testing.T) {
	agent, profile := setupIXRouteControlPlane(t)
	route := createIXRouteThroughHandler(t, agent, profile.ID, "non-echo-business", 24001, 38001)
	agent.responseDataHook = func(command cleanupAgentCommand) map[string]interface{} {
		switch command.Type {
		case "TcpPing":
			return map[string]interface{}{"success": true, "averageTime": 8.0, "packetLoss": 0.0}
		case "EchoProbe":
			return map[string]interface{}{"success": false, "errorMessage": "read: i/o timeout"}
		default:
			return nil
		}
	}
	result, err := agent.h.runIXRouteDiagnosis(route.ID, true)
	if err != nil {
		t.Fatal(err)
	}
	if result.OverallStatus != ixProbeUnverified {
		t.Fatalf("missing Echo response was reported as route failure: %+v", result)
	}
	for _, leg := range result.Legs {
		if leg.Method == "tcp_connect" && leg.Status != ixProbePass {
			t.Fatalf("successful TCP segment changed status: %+v", leg)
		}
		if strings.Contains(leg.Method, "nonce_echo") && leg.Status != ixProbeUnverified {
			t.Fatalf("missing Echo response was not UNVERIFIED: %+v", leg)
		}
	}
}

func TestIXRouteDiagnosisUsesOnlyPersistedBusinessTargets(t *testing.T) {
	agent, profile := setupIXRouteControlPlane(t)
	route := createIXRouteThroughHandler(t, agent, profile.ID, "target-bound-diagnosis", 24001, 38001)
	agent.responseDataHook = func(command cleanupAgentCommand) map[string]interface{} {
		if command.Type == "TcpPing" {
			return map[string]interface{}{"success": true, "averageTime": 9.5, "packetLoss": 0.0}
		}
		return nil
	}
	if _, err := agent.h.runIXRouteDiagnosis(route.ID, false); err != nil {
		t.Fatal(err)
	}

	commands := agent.commandsOfType("TcpPing")
	if len(commands) != 2 {
		t.Fatalf("sent %d TCP probes, want exactly the two configured business legs", len(commands))
	}
	type target struct {
		IP   string `json:"ip"`
		Port int    `json:"port"`
	}
	wants := []struct {
		nodeID int64
		ip     string
		port   int
	}{{1, "127.0.0.1", 38001}, {2, route.LandingHost, route.LandingPort}}
	for i, command := range commands {
		var got target
		if err := json.Unmarshal(command.Data, &got); err != nil {
			t.Fatal(err)
		}
		if command.NodeID != wants[i].nodeID || got.IP != wants[i].ip || got.Port != wants[i].port {
			t.Fatalf("probe %d escaped persisted route target: node=%d target=%s:%d want node=%d target=%s:%d", i, command.NodeID, got.IP, got.Port, wants[i].nodeID, wants[i].ip, wants[i].port)
		}
	}
}

func TestIXRouteDiagnosisFailsClosedOnMismatchedNATResource(t *testing.T) {
	agent, profile := setupIXRouteControlPlane(t)
	route := createIXRouteThroughHandler(t, agent, profile.ID, "mismatched-resource", 24001, 38001)
	if err := agent.h.repo.DB().Model(&model.IXPortResource{}).
		Where("node_id = ? AND public_port = ?", profile.IXNodeID, route.IXPublicPort).
		Update("public_port", 39001).Error; err != nil {
		t.Fatal(err)
	}
	before := len(agent.commandsOfType("TcpPing"))
	if _, err := agent.h.runIXRouteDiagnosis(route.ID, false); err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("mismatched resource was accepted: %v", err)
	}
	if got := len(agent.commandsOfType("TcpPing")); got != before {
		t.Fatalf("target validation dispatched a probe before failing: before=%d after=%d", before, got)
	}
}

func updateIXRouteThroughHandler(t *testing.T, agent *cleanupAgent, route *model.IXRoute, entryPort, ixPort int, protocolMode string) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	agent.h.ixRouteUpdate(recorder, ixPortRequest(t, http.MethodPost, "/api/v1/ix-route/update", map[string]interface{}{
		"id": route.ID, "pathProfileId": route.PathProfileID, "name": route.Name + " edited",
		"entryPort": entryPort, "ixPort": ixPort, "landingHost": "127.0.0.2",
		"landingPort": 9090, "protocolMode": protocolMode,
	}, 0))
	return recorder
}

func TestIXRouteEditMigratesListenersBeforeReleasingOldLeases(t *testing.T) {
	agent, profile := setupIXRouteControlPlane(t)
	route := createIXRouteThroughHandler(t, agent, profile.ID, "editable", 24001, 38001)
	recorder := updateIXRouteThroughHandler(t, agent, route, 24003, 38002, "tcp")
	if code := ixPortResponseCode(t, recorder); code != 0 {
		t.Fatalf("route edit failed code=%d body=%s", code, recorder.Body.String())
	}
	stored, err := agent.h.repo.GetIXRoute(route.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.EntryPort != 24003 || stored.IXPublicPort != 38002 || stored.ProtocolMode != "tcp" || stored.LandingPort != 9090 || stored.ObservedState != "active" {
		t.Fatalf("edited route mismatch: %+v", stored)
	}
	runtimes, err := agent.h.repo.ListIXRouteRuntimes(route.ID)
	if err != nil || len(runtimes) != 2 {
		t.Fatalf("edited TCP route should own two runtimes: %+v %v", runtimes, err)
	}
	leasing, err := agent.h.repo.ListIXPortLeases("", 0)
	if err != nil {
		t.Fatal(err)
	}
	active, released := 0, 0
	for _, lease := range leasing {
		if lease.OwnerID != route.ID {
			continue
		}
		if lease.State == "active" {
			active++
		} else if lease.State == "released" {
			released++
		}
	}
	if active != 2 || released != 4 {
		t.Fatalf("lease generations after edit active=%d released=%d all=%+v", active, released, leasing)
	}
	if migration, err := agent.h.repo.GetActiveIXRouteMigration(route.ID); err != nil || migration != nil {
		t.Fatalf("completed edit retained migration: %+v %v", migration, err)
	}
}

func TestIXRouteEditFailureRestoresOldRuntimeAndKeepsOldLeases(t *testing.T) {
	agent, profile := setupIXRouteControlPlane(t)
	route := createIXRouteThroughHandler(t, agent, profile.ID, "rollback", 24001, 38001)
	failed := false
	agent.responseHook = func(command cleanupAgentCommand) (bool, string) {
		if command.Type == "UpdateService" && strings.Contains(string(command.Data), "24002") && !failed {
			failed = true
			return false, "replacement entry rejected"
		}
		return true, ""
	}
	recorder := updateIXRouteThroughHandler(t, agent, route, 24002, 38002, "tcp_udp")
	if code := ixPortResponseCode(t, recorder); code == 0 || !failed {
		t.Fatalf("failed edit was accepted: body=%s failed=%v", recorder.Body.String(), failed)
	}
	stored, err := agent.h.repo.GetIXRoute(route.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.EntryPort != 24001 || stored.IXPublicPort != 38001 || stored.ProtocolMode != "tcp_udp" || stored.ObservedState != "active" {
		t.Fatalf("failed edit changed active route: %+v", stored)
	}
	leasing, err := agent.h.repo.ListIXPortLeases("", 0)
	if err != nil {
		t.Fatal(err)
	}
	oldActive, migrationActive := 0, 0
	for _, lease := range leasing {
		if lease.OwnerID == route.ID && lease.State == "active" {
			oldActive++
		}
		if strings.HasPrefix(lease.OwnerType, "ix_route_migration_") && lease.State == "active" {
			migrationActive++
		}
	}
	if oldActive != 4 || migrationActive != 0 {
		t.Fatalf("rollback lease ownership mismatch old=%d migration=%d all=%+v", oldActive, migrationActive, leasing)
	}
	if migration, err := agent.h.repo.GetActiveIXRouteMigration(route.ID); err != nil || migration != nil {
		t.Fatalf("successful rollback retained migration: %+v %v", migration, err)
	}
}

func TestIXRouteInterruptedEditConvergesByRollingBackAndRetainingOldPorts(t *testing.T) {
	agent, profile := setupIXRouteControlPlane(t)
	route := createIXRouteThroughHandler(t, agent, profile.ID, "recover-edit", 24001, 38001)
	stage := 0
	agent.responseHook = func(command cleanupAgentCommand) (bool, string) {
		if command.Type != "UpdateService" {
			return true, ""
		}
		if stage == 0 && strings.Contains(string(command.Data), "24002") {
			stage = 1
			return false, "replacement entry rejected"
		}
		if stage == 1 {
			stage = 2
			return false, "rollback transport interrupted"
		}
		return true, ""
	}
	recorder := updateIXRouteThroughHandler(t, agent, route, 24002, 38002, "tcp_udp")
	if code := ixPortResponseCode(t, recorder); code == 0 || stage != 2 {
		t.Fatalf("interrupted edit was not preserved for recovery: body=%s stage=%d", recorder.Body.String(), stage)
	}
	migration, err := agent.h.repo.GetActiveIXRouteMigration(route.ID)
	if err != nil || migration == nil || migration.State != "rollback_pending" {
		t.Fatalf("rollback journal missing: %+v %v", migration, err)
	}
	pending, err := agent.h.repo.ListPendingIXRouteIDs()
	if err != nil || len(pending) == 0 || pending[0] != route.ID {
		t.Fatalf("migration was not discoverable by reconciler: %+v %v", pending, err)
	}
	agent.responseHook = nil
	if err := agent.h.reconcileIXRoute(route.ID); err != nil {
		t.Fatalf("recovery reconciliation failed: %v", err)
	}
	stored, _ := agent.h.repo.GetIXRoute(route.ID)
	if stored.EntryPort != 24001 || stored.IXPublicPort != 38001 || stored.ObservedState != "active" {
		t.Fatalf("recovery did not preserve old route: %+v", stored)
	}
	if migration, err = agent.h.repo.GetActiveIXRouteMigration(route.ID); err != nil || migration != nil {
		t.Fatalf("recovery retained active journal: %+v %v", migration, err)
	}
}
