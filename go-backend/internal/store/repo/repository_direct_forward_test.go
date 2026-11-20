package repo

import (
	"path/filepath"
	"testing"
	"time"

	"go-backend/internal/store/model"
)

func TestCreateDirectForwardPersistsNodeProtocolAndNoTunnel(t *testing.T) {
	r, err := Open(filepath.Join(t.TempDir(), "direct-forward.db"))
	if err != nil {
		t.Fatalf("open repository: %v", err)
	}
	defer r.Close()
	now := time.Now().UnixMilli()
	node := model.Node{
		Name:          "entry",
		Secret:        "secret",
		ServerIP:      "127.0.0.1",
		Port:          "20000-20100",
		TCPListenAddr: "127.0.0.1",
		UDPListenAddr: "127.0.0.1",
		Status:        1,
		CreatedTime:   now,
	}
	if err := r.db.Create(&node).Error; err != nil {
		t.Fatalf("create node: %v", err)
	}

	forwardID, err := r.CreateDirectForwardTx(
		1, "admin", "direct", "tcp", "127.0.0.1:443", "fifo", now, 1,
		node.ID, 20001, "", nil, 0, 0, nil, 0, 0, 0, false,
	)
	if err != nil {
		t.Fatalf("create direct forward: %v", err)
	}

	forward, err := r.GetForwardRecord(forwardID)
	if err != nil {
		t.Fatalf("get direct forward: %v", err)
	}
	if forward == nil || forward.TunnelID != 0 || forward.Protocol != "tcp" {
		t.Fatalf("unexpected direct forward: %#v", forward)
	}
	ports, err := r.ListForwardPorts(forwardID)
	if err != nil {
		t.Fatalf("list direct forward ports: %v", err)
	}
	if len(ports) != 1 || ports[0].NodeID != node.ID || ports[0].Port != 20001 {
		t.Fatalf("unexpected direct ports: %#v", ports)
	}
}
