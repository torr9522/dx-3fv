package repo

import (
	"database/sql"
	"testing"
	"time"

	"go-backend/internal/store/model"
)

func TestListForwardsExposesDirectEntryPublicIP(t *testing.T) {
	r, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open repo: %v", err)
	}
	defer r.Close()

	now := time.Now().UnixMilli()
	node := model.Node{
		Name:          "entry",
		Secret:        "test-secret",
		ServerIP:      "192.0.2.10",
		ServerIPV4:    sql.NullString{String: "198.51.100.10", Valid: true},
		Port:          "22",
		CreatedTime:   now,
		Status:        1,
		TCPListenAddr: "0.0.0.0",
		UDPListenAddr: "0.0.0.0",
	}
	if err := r.DB().Create(&node).Error; err != nil {
		t.Fatalf("create node: %v", err)
	}

	createDirectForward := func(name string, port int, inIP sql.NullString) int64 {
		t.Helper()
		forward := model.Forward{
			UserID:      1,
			UserName:    "admin",
			Name:        name,
			TunnelID:    0,
			RemoteAddr:  "203.0.113.10:443",
			Strategy:    "fifo",
			CreatedTime: now,
			UpdatedTime: now,
			Status:      1,
		}
		if err := r.DB().Create(&forward).Error; err != nil {
			t.Fatalf("create forward: %v", err)
		}
		if err := r.DB().Create(&model.ForwardPort{
			ForwardID: forward.ID,
			NodeID:    node.ID,
			Port:      port,
			InIP:      inIP,
		}).Error; err != nil {
			t.Fatalf("create forward port: %v", err)
		}
		return forward.ID
	}

	fallbackID := createDirectForward("node-address", 41001, sql.NullString{})
	overrideID := createDirectForward("business-address", 41002, sql.NullString{
		String: "203.0.113.25",
		Valid:  true,
	})

	items, err := r.ListForwards()
	if err != nil {
		t.Fatalf("ListForwards: %v", err)
	}

	entryIPByID := make(map[int64]interface{}, len(items))
	for _, item := range items {
		id, ok := item["id"].(int64)
		if ok {
			entryIPByID[id] = item["entryIp"]
		}
	}

	if got := entryIPByID[fallbackID]; got != "198.51.100.10" {
		t.Fatalf("node IPv4 fallback = %#v, want 198.51.100.10", got)
	}
	if got := entryIPByID[overrideID]; got != "203.0.113.25" {
		t.Fatalf("business entry override = %#v, want 203.0.113.25", got)
	}
}
