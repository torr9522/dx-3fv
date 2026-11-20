package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"
	"time"

	"go-backend/internal/auth"
	"go-backend/internal/http/middleware"
	"go-backend/internal/store/model"
)

func TestForwardUpdateProtocolRemovesOldServiceFamily(t *testing.T) {
	agent := newCleanupAgent(t)
	now := time.Now().UnixMilli()
	forward := model.Forward{
		ID: 4, UserID: 1, UserName: "admin_user", Name: "direct",
		TunnelID: 0, Protocol: "tcp+udp", RemoteAddr: "127.0.0.1:9000",
		Strategy: "fifo", CreatedTime: now, UpdatedTime: now, Status: 1,
	}
	if err := agent.h.repo.DB().Create(&forward).Error; err != nil {
		t.Fatal(err)
	}
	if err := agent.h.repo.DB().Create(&model.ForwardPort{ForwardID: forward.ID, NodeID: 1, Port: 30000}).Error; err != nil {
		t.Fatal(err)
	}

	body, err := json.Marshal(map[string]interface{}{
		"id": forward.ID, "tunnelId": 0, "entryNodeId": 1,
		"inPort": 30000, "remoteAddr": "127.0.0.1:9000", "protocol": "tcp",
	})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/forward/update", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), middleware.ClaimsContextKey, auth.Claims{Sub: "1", RoleID: 0}))
	res := httptest.NewRecorder()
	agent.h.forwardUpdate(res, req)

	var response struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, res.Body.String())
	}
	if response.Code != 0 {
		t.Fatalf("update failed: code=%d msg=%s", response.Code, response.Msg)
	}

	deletes := agent.commandsOfType("DeleteService")
	if len(deletes) != 1 {
		t.Fatalf("expected one family cleanup, got %d", len(deletes))
	}
	var deletePayload struct {
		Services []string `json:"services"`
	}
	if err := json.Unmarshal(deletes[0].Data, &deletePayload); err != nil {
		t.Fatal(err)
	}
	sort.Strings(deletePayload.Services)
	wantDelete := []string{"4_1_0", "4_1_0_tcp", "4_1_0_udp"}
	if !reflect.DeepEqual(deletePayload.Services, wantDelete) {
		t.Fatalf("cleanup services = %v, want %v", deletePayload.Services, wantDelete)
	}

	updates := agent.commandsOfType("UpdateService")
	if len(updates) != 1 {
		t.Fatalf("expected one replacement update, got %d", len(updates))
	}
	var services []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(updates[0].Data, &services); err != nil {
		t.Fatal(err)
	}
	if len(services) != 1 || services[0].Name != "4_1_0_tcp" {
		t.Fatalf("replacement services = %+v, want only TCP", services)
	}
}
