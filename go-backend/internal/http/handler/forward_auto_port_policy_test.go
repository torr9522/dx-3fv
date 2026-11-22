package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-backend/internal/store/model"
	"go-backend/internal/store/repo"
)

type forwardAutoPortFixture struct {
	agent  *cleanupAgent
	tunnel bool
}

func newForwardAutoPortFixture(t *testing.T, tunnel bool) forwardAutoPortFixture {
	t.Helper()
	agent := newCleanupAgent(t, func(r *repo.Repository) {
		if err := r.DB().Model(&model.Node{}).Where("id = ?", 1).Update("port", "8000,9999,10000-10030").Error; err != nil {
			t.Fatal(err)
		}
		if !tunnel {
			return
		}
		now := time.Now().UnixMilli()
		if err := r.DB().Create(&model.Tunnel{ID: 1, Name: "entry-policy", Type: 1, Protocol: "tls", Flow: 1, TrafficRatio: 1, Status: 1, CreatedTime: now, UpdatedTime: now}).Error; err != nil {
			t.Fatal(err)
		}
		if err := r.DB().Create(&model.ChainTunnel{TunnelID: 1, ChainType: "1", NodeID: 1, Protocol: sql.NullString{String: "tcp", Valid: true}, Inx: sql.NullInt64{Int64: 1, Valid: true}}).Error; err != nil {
			t.Fatal(err)
		}
	})
	return forwardAutoPortFixture{agent: agent, tunnel: tunnel}
}

func (fixture forwardAutoPortFixture) create(t *testing.T, name string, port int) *model.Forward {
	t.Helper()
	payload := map[string]interface{}{
		"name": name, "remoteAddr": "127.0.0.1:8080", "inPort": port, "protocol": "tcp",
	}
	if fixture.tunnel {
		payload["tunnelId"] = 1
	} else {
		payload["tunnelId"] = 0
		payload["entryNodeId"] = 1
	}
	recorder := ixPortRequestRecorder(t, fixture.agent.h.forwardCreate, "/api/v1/forward/create", payload)
	if code := ixPortResponseCode(t, recorder); code != 0 {
		t.Fatalf("create %s failed code=%d body=%s", name, code, recorder.Body.String())
	}
	var forward model.Forward
	if err := fixture.agent.h.repo.DB().Where("name = ?", name).Last(&forward).Error; err != nil {
		t.Fatal(err)
	}
	return &forward
}

func ixPortRequestRecorder(t *testing.T, handler func(http.ResponseWriter, *http.Request), path string, payload map[string]interface{}) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	handler(recorder, ixPortRequest(t, http.MethodPost, path, payload, 0))
	return recorder
}

func (fixture forwardAutoPortFixture) port(t *testing.T, forwardID int64) int {
	t.Helper()
	var item model.ForwardPort
	if err := fixture.agent.h.repo.DB().Where("forward_id = ?", forwardID).First(&item).Error; err != nil {
		t.Fatal(err)
	}
	return item.Port
}

func (fixture forwardAutoPortFixture) updateClear(t *testing.T, forward *model.Forward) {
	t.Helper()
	payload := map[string]interface{}{
		"id": forward.ID, "name": forward.Name, "remoteAddr": forward.RemoteAddr,
		"inPort": 0, "protocol": "tcp", "tunnelId": forward.TunnelID,
	}
	if !fixture.tunnel {
		payload["entryNodeId"] = 1
	}
	recorder := ixPortRequestRecorder(t, fixture.agent.h.forwardUpdate, "/api/v1/forward/update", payload)
	if code := ixPortResponseCode(t, recorder); code != 0 {
		t.Fatalf("clear entry port failed code=%d body=%s", code, recorder.Body.String())
	}
}

func TestForwardTypesUseUnifiedAutomaticAndManualPortPolicy(t *testing.T) {
	for _, tunnel := range []bool{false, true} {
		name := "single"
		if tunnel {
			name = "tunnel"
		}
		t.Run(name, func(t *testing.T) {
			fixture := newForwardAutoPortFixture(t, tunnel)
			if got := fixture.port(t, fixture.create(t, name+"-manual-8000", 8000).ID); got != 8000 {
				t.Fatalf("manual 8000 became %d", got)
			}
			if got := fixture.port(t, fixture.create(t, name+"-manual-9999", 9999).ID); got != 9999 {
				t.Fatalf("manual 9999 became %d", got)
			}

			conflictPayload := map[string]interface{}{
				"name": name + "-occupied", "remoteAddr": "127.0.0.1:8080", "inPort": 8000, "protocol": "tcp",
			}
			if tunnel {
				conflictPayload["tunnelId"] = 1
			} else {
				conflictPayload["tunnelId"] = 0
				conflictPayload["entryNodeId"] = 1
			}
			conflict := ixPortRequestRecorder(t, fixture.agent.h.forwardCreate, "/api/v1/forward/create", conflictPayload)
			if code := ixPortResponseCode(t, conflict); code == 0 {
				t.Fatalf("occupied manual port was accepted: %s", conflict.Body.String())
			}

			automatic := fixture.create(t, name+"-automatic", 0)
			oldPort := fixture.port(t, automatic.ID)
			if oldPort < repo.AutoEntryPortMin {
				t.Fatalf("automatic port %d is below policy minimum", oldPort)
			}
			fixture.updateClear(t, automatic)
			newPort := fixture.port(t, automatic.ID)
			if newPort < repo.AutoEntryPortMin || newPort == oldPort {
				t.Fatalf("edit-clear port before=%d after=%d", oldPort, newPort)
			}
		})
	}
}

func commandServicePort(data json.RawMessage) (int, error) {
	var services []struct {
		Addr string `json:"addr"`
	}
	if err := json.Unmarshal(data, &services); err != nil {
		return 0, err
	}
	if len(services) == 0 {
		return 0, fmt.Errorf("service payload is empty")
	}
	_, portText, err := net.SplitHostPort(services[0].Addr)
	if err != nil {
		return 0, err
	}
	var port int
	if _, err := fmt.Sscanf(portText, "%d", &port); err != nil {
		return 0, err
	}
	return port, nil
}

func TestForwardTypesReallocateAutomaticPortAfterRuntimeConflict(t *testing.T) {
	for _, tunnel := range []bool{false, true} {
		name := "single"
		if tunnel {
			name = "tunnel"
		}
		t.Run(name, func(t *testing.T) {
			fixture := newForwardAutoPortFixture(t, tunnel)
			failedPort := 0
			fixture.agent.responseHook = func(command cleanupAgentCommand) (bool, string) {
				if command.Type != "UpdateService" && command.Type != "AddService" {
					return true, ""
				}
				port, err := commandServicePort(command.Data)
				if err != nil {
					return false, err.Error()
				}
				if failedPort == 0 {
					failedPort = port
				}
				if port == failedPort {
					return false, "listen tcp: bind: address already in use"
				}
				return true, ""
			}
			forward := fixture.create(t, name+"-runtime-retry", 0)
			port := fixture.port(t, forward.ID)
			if failedPort < repo.AutoEntryPortMin || port < repo.AutoEntryPortMin || port == failedPort {
				t.Fatalf("Runtime retry failed: conflict=%d final=%d", failedPort, port)
			}
			updates := fixture.agent.commandsOfType("UpdateService")
			if len(updates) < 2 {
				t.Fatalf("expected a retry, got %d UpdateService commands", len(updates))
			}
		})
	}
}
