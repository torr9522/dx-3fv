package service

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-gost/x/config"
)

func TestGlobalTrafficManagerRemoveServicesDropsCachedEntries(t *testing.T) {
	m := &GlobalTrafficManager{serviceTraffic: map[string]*ServiceTraffic{
		"svc-a": {ServiceName: "svc-a"},
		"svc-b": {ServiceName: "svc-b"},
	}}

	m.RemoveServices("svc-a")

	if _, ok := m.serviceTraffic["svc-a"]; ok {
		t.Fatalf("expected svc-a traffic entry to be removed")
	}
	if _, ok := m.serviceTraffic["svc-b"]; !ok {
		t.Fatalf("expected svc-b traffic entry to remain")
	}
}

func TestGlobalTrafficManagerRestoresCountersAfterReportFailure(t *testing.T) {
	origReportDo := reportDo
	origReportURL := httpReportURL
	origAESCrypto := httpAESCrypto
	defer func() {
		reportDo = origReportDo
		httpReportURL = origReportURL
		httpAESCrypto = origAESCrypto
	}()

	httpReportURL = "http://panel.example.com/flow/upload?secret=abc"
	httpAESCrypto = nil
	reportDo = func(_ context.Context, _ *http.Request, _ time.Duration) (*http.Response, error) {
		return nil, errors.New("panel unavailable")
	}

	m := &GlobalTrafficManager{
		serviceTraffic: map[string]*ServiceTraffic{
			"pending-service": {ServiceName: "pending-service", UpBytes: 10, DownBytes: 20},
		},
		ctx: context.Background(),
	}

	m.collectAndReport()

	up, down := m.GetServiceTraffic("pending-service")
	if up != 10 || down != 20 {
		t.Fatalf("failed report lost traffic: up=%d down=%d", up, down)
	}
}

func TestGlobalTrafficManagerRecoversDurableOutboxAfterRestart(t *testing.T) {
	origReportDo := reportDo
	origReportURL := httpReportURL
	origAESCrypto := httpAESCrypto
	origPersistPath := config.PersistPath()
	defer func() {
		reportDo = origReportDo
		httpReportURL = origReportURL
		httpAESCrypto = origAESCrypto
		config.SetPersistPath(origPersistPath)
	}()

	dir := t.TempDir()
	config.SetPersistPath(filepath.Join(dir, "gost.json"))
	httpReportURL = "http://panel.example.com/flow/upload?secret=abc"
	httpAESCrypto = nil
	reportDo = func(_ context.Context, _ *http.Request, _ time.Duration) (*http.Response, error) {
		return nil, errors.New("panel unavailable")
	}

	first := &GlobalTrafficManager{
		serviceTraffic: map[string]*ServiceTraffic{
			"restart-service": {ServiceName: "restart-service", UpBytes: 31, DownBytes: 47},
		},
		ctx: context.Background(),
	}
	first.collectAndReport()

	pending, err := loadTrafficOutbox()
	if err != nil || pending == nil {
		t.Fatalf("durable outbox missing: pending=%#v err=%v", pending, err)
	}
	info, err := os.Stat(trafficOutboxPath())
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("outbox permission=%o want=600", info.Mode().Perm())
	}

	var gotID, gotBody string
	reportDo = func(_ context.Context, req *http.Request, _ time.Duration) (*http.Response, error) {
		gotID = req.Header.Get("X-DX3FV-Flow-Report-ID")
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatal(err)
		}
		gotBody = string(body)
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("ok"))}, nil
	}
	restarted := &GlobalTrafficManager{serviceTraffic: map[string]*ServiceTraffic{}, ctx: context.Background()}
	restarted.collectAndReport()

	if gotID != pending.ReportID || !strings.Contains(gotBody, "restart-service") {
		t.Fatalf("recovered report mismatch: id=%q body=%s", gotID, gotBody)
	}
	if _, err := os.Stat(trafficOutboxPath()); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("acknowledged outbox still exists: %v", err)
	}
}

func TestGlobalTrafficManagerQueuesNewTrafficBehindUncertainBatch(t *testing.T) {
	origReportDo := reportDo
	origReportURL := httpReportURL
	origAESCrypto := httpAESCrypto
	origPersistPath := config.PersistPath()
	defer func() {
		reportDo = origReportDo
		httpReportURL = origReportURL
		httpAESCrypto = origAESCrypto
		config.SetPersistPath(origPersistPath)
	}()

	config.SetPersistPath(filepath.Join(t.TempDir(), "gost.json"))
	httpReportURL = "http://panel.example.com/flow/upload?secret=abc"
	httpAESCrypto = nil
	reportDo = func(_ context.Context, _ *http.Request, _ time.Duration) (*http.Response, error) {
		return nil, errors.New("uncertain delivery")
	}

	traffic := &ServiceTraffic{ServiceName: "queued-service", UpBytes: 10, DownBytes: 20}
	m := &GlobalTrafficManager{serviceTraffic: map[string]*ServiceTraffic{"queued-service": traffic}, ctx: context.Background()}
	m.collectAndReport()
	traffic.mu.Lock()
	traffic.UpBytes += 30
	traffic.DownBytes += 40
	traffic.mu.Unlock()
	m.collectAndReport()

	queue, err := loadTrafficOutboxQueue()
	if err != nil {
		t.Fatal(err)
	}
	if len(queue.Batches) != 2 {
		t.Fatalf("queued batches=%d want=2", len(queue.Batches))
	}
	if queue.Batches[0].ReportID == queue.Batches[1].ReportID {
		t.Fatal("uncertain and new traffic reused one report ID")
	}
	if queue.Batches[0].Items[0].U != 10 || queue.Batches[1].Items[0].U != 30 {
		t.Fatalf("queued traffic was merged incorrectly: %+v", queue.Batches)
	}
}

func TestGlobalTrafficManagerRetainServicesDropsStaleEntries(t *testing.T) {
	m := &GlobalTrafficManager{serviceTraffic: map[string]*ServiceTraffic{
		"svc-a": {ServiceName: "svc-a"},
		"svc-b": {ServiceName: "svc-b"},
	}}

	m.RetainServices(map[string]struct{}{"svc-b": {}})

	if _, ok := m.serviceTraffic["svc-a"]; ok {
		t.Fatalf("expected stale svc-a traffic entry to be removed")
	}
	if _, ok := m.serviceTraffic["svc-b"]; !ok {
		t.Fatalf("expected active svc-b traffic entry to remain")
	}
}

func TestGlobalTrafficManagerAddTrafficIgnoresUnregisteredService(t *testing.T) {
	m := &GlobalTrafficManager{serviceTraffic: make(map[string]*ServiceTraffic)}

	m.AddTraffic("deleted-service", 10, 20)

	if _, ok := m.serviceTraffic["deleted-service"]; ok {
		t.Fatalf("expected unregistered service traffic to be ignored")
	}
}

func TestGlobalTrafficManagerCollectAndReportDropsStaleEntriesAfterReporting(t *testing.T) {
	origReportDo := reportDo
	origReportURL := httpReportURL
	origAESCrypto := httpAESCrypto
	defer func() {
		reportDo = origReportDo
		httpReportURL = origReportURL
		httpAESCrypto = origAESCrypto
	}()

	httpReportURL = "http://panel.example.com/flow/upload?secret=abc"
	httpAESCrypto = nil

	var requestBody string
	reportDo = func(_ context.Context, req *http.Request, _ time.Duration) (*http.Response, error) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		requestBody = string(body)
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("ok")),
		}, nil
	}

	m := &GlobalTrafficManager{
		serviceTraffic: map[string]*ServiceTraffic{
			"stale-service": {ServiceName: "stale-service", UpBytes: 10, DownBytes: 20},
		},
		ctx: context.Background(),
	}

	m.collectAndReport()

	if !strings.Contains(requestBody, "stale-service") {
		t.Fatalf("expected pending stale traffic to be reported first, body=%s", requestBody)
	}
	if _, ok := m.serviceTraffic["stale-service"]; ok {
		t.Fatalf("expected stale traffic entry to be removed after report collection")
	}
}
