package repo

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"go-backend/internal/store/model"
)

func openIXPortTestRepo(t *testing.T) *Repository {
	t.Helper()
	r, err := Open(filepath.Join(t.TempDir(), "ix-ports.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = r.Close() })
	return r
}

func createIXTestPool(t *testing.T, r *Repository, nodeID int64, name, kind string, start, end int) *model.IXPortPool {
	t.Helper()
	if node, err := r.GetNodeByID(nodeID); err != nil {
		t.Fatal(err)
	} else if node == nil {
		now := int64(1)
		if err := r.DB().Create(&model.Node{ID: nodeID, Name: fmt.Sprintf("node-%d", nodeID), Secret: fmt.Sprintf("secret-%d", nodeID), ServerIP: "127.0.0.1", Port: "1-65535", CreatedTime: now, Status: 1, TCPListenAddr: "[::]", UDPListenAddr: "[::]"}).Error; err != nil {
			t.Fatal(err)
		}
	}
	pool := &model.IXPortPool{NodeID: nodeID, Name: name, Kind: kind, RangeStart: start, RangeEnd: end, NextPort: start, Status: 1}
	if err := r.CreateIXPortPool(pool); err != nil {
		t.Fatal(err)
	}
	return pool
}

func TestIXFullPoolExcludesSSHAndAllocatesTCPUDPAtomically(t *testing.T) {
	r := openIXPortTestRepo(t)
	pool := createIXTestPool(t, r, 10, "gz2", IXPoolKindFull, 20, 24)
	if pool.RangeStart != 1 || pool.RangeEnd != 65535 {
		t.Fatalf("full pool was narrowed to %d-%d", pool.RangeStart, pool.RangeEnd)
	}

	if _, err := r.AllocateIXPorts(IXPortAllocationRequest{PoolID: pool.ID, RequestedPort: 22, Protocols: []string{"tcp"}, OwnerType: "route", OwnerID: 1, Manual: true}); err == nil {
		t.Fatal("port 22 must be permanently excluded")
	}
	allocation, err := r.AllocateIXPorts(IXPortAllocationRequest{PoolID: pool.ID, RequestedPort: 23, Protocols: []string{"udp", "tcp"}, OwnerType: "route", OwnerID: 1, Manual: true})
	if err != nil {
		t.Fatal(err)
	}
	if allocation.Resource.InternalPort != 23 || len(allocation.Leases) != 2 {
		t.Fatalf("unexpected atomic allocation: %+v", allocation)
	}
	if _, err := r.AllocateIXPorts(IXPortAllocationRequest{PoolID: pool.ID, RequestedPort: 23, Protocols: []string{"udp"}, OwnerType: "route", OwnerID: 2, Manual: true}); err == nil {
		t.Fatal("an occupied protocol must reject the complete request")
	}
	leasing, err := r.ListIXPortLeases("route", 2)
	if err != nil || len(leasing) != 0 {
		t.Fatalf("failed atomic allocation leaked leases: %+v %v", leasing, err)
	}
}

func TestIXFullPoolKeepsTCPAndUDPOwnershipIndependent(t *testing.T) {
	r := openIXPortTestRepo(t)
	pool := createIXTestPool(t, r, 11, "gz2", IXPoolKindFull, 44000, 44010)
	if _, err := r.AllocateIXPorts(IXPortAllocationRequest{PoolID: pool.ID, RequestedPort: 44000, Protocols: []string{"tcp"}, OwnerType: "route", OwnerID: 1, Manual: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := r.AllocateIXPorts(IXPortAllocationRequest{PoolID: pool.ID, RequestedPort: 44000, Protocols: []string{"udp"}, OwnerType: "route", OwnerID: 2, Manual: true}); err != nil {
		t.Fatalf("UDP must remain independently allocatable: %v", err)
	}
	if _, err := r.AllocateIXPorts(IXPortAllocationRequest{PoolID: pool.ID, RequestedPort: 44000, Protocols: []string{"tcp", "udp"}, OwnerType: "route", OwnerID: 3, Manual: true}); err == nil {
		t.Fatal("TCP+UDP allocation must fail atomically when either protocol is owned")
	}
	leasing, err := r.ListIXPortLeases("route", 3)
	if err != nil || len(leasing) != 0 {
		t.Fatalf("failed dual-protocol request leaked leases: %+v %v", leasing, err)
	}
}

func TestIXAllocationIsScopedByNodeIdentityNotPublicIP(t *testing.T) {
	r := openIXPortTestRepo(t)
	first := createIXTestPool(t, r, 20, "gz2", IXPoolKindFull, 38000, 38010)
	second := createIXTestPool(t, r, 21, "frontsea-ix2", IXPoolKindNAT, 38000, 38099)
	resource := &model.IXPortResource{PoolID: second.ID, NodeID: 21, PublicIP: "198.51.100.10", PublicPort: 38001, InternalPort: 38001, TCPMapped: true, UDPMapped: true, VerificationStatus: "verified"}
	if err := r.SaveIXPortResource(resource); err != nil {
		t.Fatal(err)
	}
	if _, err := r.AllocateIXPorts(IXPortAllocationRequest{PoolID: first.ID, RequestedPort: 38001, Protocols: []string{"tcp"}, OwnerType: "route", OwnerID: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := r.AllocateIXPorts(IXPortAllocationRequest{PoolID: second.ID, RequestedPort: 38001, Protocols: []string{"tcp"}, OwnerType: "route", OwnerID: 2}); err != nil {
		t.Fatalf("same numeric port on another Node identity must remain independent: %v", err)
	}
}

func TestIXNATPoolRequiresVerifiedProtocolMapping(t *testing.T) {
	r := openIXPortTestRepo(t)
	pool := createIXTestPool(t, r, 30, "nat", IXPoolKindNAT, 38000, 38099)
	for i, resource := range []*model.IXPortResource{
		{PoolID: pool.ID, NodeID: 30, PublicPort: 38001, InternalPort: 30001, TCPMapped: true, UDPMapped: true, VerificationStatus: "unverified"},
		{PoolID: pool.ID, NodeID: 30, PublicPort: 38002, InternalPort: 30002, TCPMapped: true, UDPMapped: false, VerificationStatus: "verified"},
		{PoolID: pool.ID, NodeID: 30, PublicPort: 38003, InternalPort: 30003, TCPMapped: true, UDPMapped: true, VerificationStatus: "verified"},
	} {
		resource.PublicIP = "198.51.100.10"
		resource.InternalIP = "10.0.0.2"
		if err := r.SaveIXPortResource(resource); err != nil {
			t.Fatalf("resource %d: %v", i, err)
		}
	}
	allocation, err := r.AllocateIXPorts(IXPortAllocationRequest{PoolID: pool.ID, Protocols: []string{"tcp", "udp"}, OwnerType: "route", OwnerID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if allocation.Resource.PublicPort != 38003 || allocation.Resource.InternalPort != 30003 {
		t.Fatalf("allocated unverified or protocol-incomplete NAT resource: %+v", allocation.Resource)
	}
}

func TestIXConcurrentAllocationDoesNotDuplicateProtocolLease(t *testing.T) {
	r := openIXPortTestRepo(t)
	pool := createIXTestPool(t, r, 40, "full", IXPoolKindFull, 41000, 41020)
	const workers = 12
	ports := make(chan int, workers)
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(owner int64) {
			defer wg.Done()
			allocation, err := r.AllocateIXPorts(IXPortAllocationRequest{PoolID: pool.ID, Protocols: []string{"tcp", "udp"}, OwnerType: "route", OwnerID: owner})
			if err != nil {
				errs <- err
				return
			}
			ports <- allocation.Resource.InternalPort
		}(int64(i + 1))
	}
	wg.Wait()
	close(ports)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	seen := map[int]bool{}
	for port := range ports {
		if seen[port] {
			t.Fatalf("duplicate concurrent port %d", port)
		}
		seen[port] = true
	}
	if len(seen) != workers {
		t.Fatalf("allocated %d ports, want %d", len(seen), workers)
	}
}

func TestIXLeaseReleaseRequiresExplicitCompletion(t *testing.T) {
	r := openIXPortTestRepo(t)
	pool := createIXTestPool(t, r, 50, "full", IXPoolKindFull, 42000, 42001)
	request := IXPortAllocationRequest{PoolID: pool.ID, RequestedPort: 42000, Protocols: []string{"tcp"}, OwnerType: "route", OwnerID: 1}
	if _, err := r.AllocateIXPorts(request); err != nil {
		t.Fatal(err)
	}
	if err := r.RequestIXPortLeaseRelease("route", 1); err != nil {
		t.Fatal(err)
	}
	request.OwnerID = 2
	if _, err := r.AllocateIXPorts(request); err == nil {
		t.Fatal("pending release must retain ownership")
	}
	if err := r.CompleteIXPortLeaseRelease("route", 1); err != nil {
		t.Fatal(err)
	}
	if _, err := r.AllocateIXPorts(request); err != nil {
		t.Fatalf("completed release should permit reuse: %v", err)
	}
	leasing, err := r.ListIXPortLeases("", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(leasing) != 2 {
		t.Fatalf("lease history was not retained: %s", fmt.Sprint(leasing))
	}
}

func TestIXAllocationHonorsDeclaredRuntimeObservation(t *testing.T) {
	r := openIXPortTestRepo(t)
	pool := createIXTestPool(t, r, 60, "full", IXPoolKindFull, 43000, 43001)
	if err := r.UpsertIXPortObservation(&model.IXPortObservation{NodeID: 60, InternalPort: 43000, Protocol: "tcp", RuntimeName: "external", State: "listening"}); err != nil {
		t.Fatal(err)
	}
	allocation, err := r.AllocateIXPorts(IXPortAllocationRequest{PoolID: pool.ID, Protocols: []string{"tcp"}, OwnerType: "route", OwnerID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if allocation.Resource.InternalPort != 43001 {
		t.Fatalf("allocated observed listener port: %+v", allocation.Resource)
	}
}
