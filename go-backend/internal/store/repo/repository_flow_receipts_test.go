package repo

import (
	"testing"
	"time"

	"go-backend/internal/store/model"
)

func TestApplyDurableFlowReportIsIdempotentAcrossCoreCounters(t *testing.T) {
	r := openIXPortTestRepo(t)
	now := time.Now()
	seedFlowBatchRows(t, r, now.UnixMilli())
	if err := r.DB().Create(&model.Node{ID: 50, Name: "gz2", Secret: "durable-flow-node", ServerIP: "127.0.0.1", Port: "1-65535", Status: 1, TCPListenAddr: "[::]", UDPListenAddr: "[::]"}).Error; err != nil {
		t.Fatal(err)
	}
	route := &model.IXRoute{PathProfileID: 1, Name: "route", LandingHost: "203.0.113.10", LandingPort: 443, ProtocolMode: "tcp", DesiredState: "active", ObservedState: "active", CreatedTime: now.UnixMilli(), UpdatedTime: now.UnixMilli()}
	if err := r.DB().Create(route).Error; err != nil {
		t.Fatal(err)
	}
	if err := r.DB().Create(&model.IXRouteRuntime{RouteID: route.ID, NodeID: 50, Role: "entry", Protocol: "tcp", ServiceName: "dx3fv_ix_1_entry_tcp", Config: "{}", DesiredState: "active", ObservedState: "active", Applied: true, Generation: 1, CreatedTime: now.UnixMilli(), UpdatedTime: now.UnixMilli()}).Error; err != nil {
		t.Fatal(err)
	}

	reportID := "0123456789abcdef0123456789abcdef"
	apply := func() bool {
		applied, _, err := r.ApplyDurableFlowReport(
			50,
			reportID,
			[]FlowUploadCounterDelta{{ForwardID: 20, UserID: 2, UserTunnelID: 10, InFlow: 10, OutFlow: 20}},
			[]IXRouteFlowDelta{{RouteID: route.ID, InFlow: 30, OutFlow: 40}},
			map[int64]int64{2: 30},
			nil,
			now,
		)
		if err != nil {
			t.Fatal(err)
		}
		return applied
	}
	if !apply() {
		t.Fatal("first durable report was not applied")
	}
	if apply() {
		t.Fatal("duplicate durable report was applied twice")
	}

	checks := []struct {
		query string
		want  int64
	}{
		{"SELECT in_flow FROM forward WHERE id = 20", 10},
		{"SELECT out_flow FROM user WHERE id = 2", 20},
		{"SELECT in_flow FROM user_tunnel WHERE id = 10", 10},
		{"SELECT in_flow FROM ix_route WHERE id = ?", 30},
		{"SELECT out_flow FROM ix_route WHERE id = ?", 40},
		{"SELECT daily_used_bytes FROM user_quota WHERE user_id = 2", 30},
		{"SELECT COUNT(*) FROM flow_upload_receipt WHERE node_id = 50", 1},
	}
	for _, check := range checks {
		args := []interface{}{}
		if check.query == "SELECT in_flow FROM ix_route WHERE id = ?" || check.query == "SELECT out_flow FROM ix_route WHERE id = ?" {
			args = append(args, route.ID)
		}
		if got := mustFlowBatchCount(t, r, check.query, args...); got != check.want {
			t.Fatalf("query %q got=%d want=%d", check.query, got, check.want)
		}
	}
	if err := r.PurgeFlowUploadReceiptsBefore(now.Add(time.Millisecond).UnixMilli()); err != nil {
		t.Fatal(err)
	}
	if got := mustFlowBatchCount(t, r, "SELECT COUNT(*) FROM flow_upload_receipt WHERE node_id = 50"); got != 0 {
		t.Fatalf("expired receipt was not purged: %d", got)
	}
}
