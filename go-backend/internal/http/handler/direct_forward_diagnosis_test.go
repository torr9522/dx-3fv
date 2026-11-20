package handler

import (
	"context"
	"testing"
	"time"

	"go-backend/internal/store/model"
)

func TestDirectForwardDiagnosisUsesTCPAndLeavesUDPUnverified(t *testing.T) {
	agent := newCleanupAgent(t)
	agent.responseDataHook = func(command cleanupAgentCommand) map[string]interface{} {
		if command.Type == "TcpPing" {
			return map[string]interface{}{"success": true, "averageTime": 8.5, "packetLoss": 0.0}
		}
		return nil
	}
	now := time.Now().UnixMilli()
	forward := model.Forward{
		ID: 8, UserID: 1, UserName: "admin_user", Name: "direct-diagnosis",
		TunnelID: 0, Protocol: "tcp+udp", RemoteAddr: "127.0.0.1:9000",
		Strategy: "fifo", CreatedTime: now, UpdatedTime: now, Status: 1,
	}
	if err := agent.h.repo.DB().Create(&forward).Error; err != nil {
		t.Fatal(err)
	}
	if err := agent.h.repo.DB().Create(&model.ForwardPort{ForwardID: forward.ID, NodeID: 1, Port: 30000}).Error; err != nil {
		t.Fatal(err)
	}

	payload, err := agent.h.diagnoseForwardRuntime(context.Background(), &forwardRecord{
		ID: forward.ID, Name: forward.Name, TunnelID: 0,
		Protocol: forward.Protocol, RemoteAddr: forward.RemoteAddr,
	})
	if err != nil {
		t.Fatal(err)
	}
	results, ok := payload["results"].([]map[string]interface{})
	if !ok || len(results) != 2 {
		t.Fatalf("results = %#v", payload["results"])
	}
	if results[0]["status"] != "PASS" || results[0]["protocol"] != "tcp" || results[0]["method"] != "tcp_connect" {
		t.Fatalf("TCP evidence = %#v", results[0])
	}
	if results[1]["status"] != "UNVERIFIED" || results[1]["protocol"] != "udp" || results[1]["method"] != "not_probed" {
		t.Fatalf("UDP evidence = %#v", results[1])
	}
	if len(agent.commandsOfType("TcpPing")) != 1 {
		t.Fatalf("expected one TCP probe, got %d", len(agent.commandsOfType("TcpPing")))
	}
	if len(agent.commandsOfType("UdpPing")) != 0 {
		t.Fatal("direct diagnosis must not treat UDP dial as payload evidence")
	}
}
