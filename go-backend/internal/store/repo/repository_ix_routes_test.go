package repo

import (
	"testing"

	"go-backend/internal/store/model"
)

func TestIXPathProfileIsStaticAndNodeScoped(t *testing.T) {
	r := openIXPortTestRepo(t)
	entry := createIXTestPool(t, r, 101, "gz2", IXPoolKindFull, 1, 65535)
	ix := createIXTestPool(t, r, 102, "ix2", IXPoolKindNAT, 38000, 38099)
	profile := &model.IXPathProfile{Name: "gz2-ix2", EntryNodeID: 101, IXNodeID: 102, EntryPoolID: entry.ID, IXPoolID: ix.ID}
	if err := r.CreateIXPathProfile(profile); err != nil {
		t.Fatal(err)
	}
	leases, err := r.ListIXPortLeases("", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(leases) != 0 {
		t.Fatalf("static PathProfile allocated ports: %+v", leases)
	}
	var runtimeCount int64
	if err := r.DB().Model(&model.IXRouteRuntime{}).Count(&runtimeCount).Error; err != nil || runtimeCount != 0 {
		t.Fatalf("static PathProfile created runtime state: count=%d err=%v", runtimeCount, err)
	}
}

func TestIXPathProfileRejectsPoolFromAnotherNodeIdentity(t *testing.T) {
	r := openIXPortTestRepo(t)
	entry := createIXTestPool(t, r, 111, "entry", IXPoolKindFull, 1, 65535)
	wrong := createIXTestPool(t, r, 112, "wrong", IXPoolKindNAT, 38000, 38099)
	createIXTestPool(t, r, 113, "ix", IXPoolKindNAT, 38000, 38099)
	profile := &model.IXPathProfile{Name: "bad", EntryNodeID: 111, IXNodeID: 113, EntryPoolID: entry.ID, IXPoolID: wrong.ID}
	if err := r.CreateIXPathProfile(profile); err == nil {
		t.Fatal("profile accepted an IX pool owned by another Node identity")
	}
}

func TestIXPathProfileDeleteProtectsActiveBusinesses(t *testing.T) {
	r := openIXPortTestRepo(t)
	entry := createIXTestPool(t, r, 121, "entry", IXPoolKindFull, 1, 65535)
	ix := createIXTestPool(t, r, 122, "ix", IXPoolKindNAT, 38000, 38099)
	profile := &model.IXPathProfile{Name: "path", EntryNodeID: 121, IXNodeID: 122, EntryPoolID: entry.ID, IXPoolID: ix.ID}
	if err := r.CreateIXPathProfile(profile); err != nil {
		t.Fatal(err)
	}
	route := &model.IXRoute{PathProfileID: profile.ID, Name: "landing", LandingHost: "203.0.113.10", LandingPort: 443, ProtocolMode: "tcp"}
	if err := r.CreateIXRoute(route); err != nil {
		t.Fatal(err)
	}
	if err := r.DeleteIXPathProfile(profile.ID); err == nil {
		t.Fatal("active route did not protect its PathProfile")
	}
}

func TestIXPathProfileIdentityUpdateProtectsActiveBusinesses(t *testing.T) {
	r := openIXPortTestRepo(t)
	entry := createIXTestPool(t, r, 131, "entry", IXPoolKindFull, 1, 65535)
	entryReplacement := createIXTestPool(t, r, 131, "entry-replacement", IXPoolKindFull, 1, 65535)
	ix := createIXTestPool(t, r, 132, "ix", IXPoolKindNAT, 38000, 38099)
	profile := &model.IXPathProfile{Name: "path", EntryNodeID: 131, IXNodeID: 132, EntryPoolID: entry.ID, IXPoolID: ix.ID}
	if err := r.CreateIXPathProfile(profile); err != nil {
		t.Fatal(err)
	}
	route := &model.IXRoute{PathProfileID: profile.ID, Name: "landing", LandingHost: "203.0.113.10", LandingPort: 443, ProtocolMode: "tcp"}
	if err := r.CreateIXRoute(route); err != nil {
		t.Fatal(err)
	}

	profile.Name = "renamed"
	profile.Status = 0
	if err := r.UpdateIXPathProfile(profile); err != nil {
		t.Fatalf("name/status-only update should remain safe: %v", err)
	}
	profile.EntryPoolID = entryReplacement.ID
	if err := r.UpdateIXPathProfile(profile); err == nil {
		t.Fatal("active route did not protect its PathProfile identity")
	}
	stored, err := r.GetIXPathProfile(profile.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.EntryPoolID != entry.ID || stored.Name != "renamed" || stored.Status != 0 {
		t.Fatalf("unsafe update partially changed profile: %+v", stored)
	}
}

func TestValidateIXPathProfileForRouteRequiresActiveCompatibleResources(t *testing.T) {
	r := openIXPortTestRepo(t)
	entry := createIXTestPool(t, r, 141, "entry", IXPoolKindFull, 1, 65535)
	ix := createIXTestPool(t, r, 142, "ix", IXPoolKindNAT, 38000, 38099)
	resource := &model.IXPortResource{
		PoolID: ix.ID, NodeID: 142, PublicPort: 38001, InternalPort: 30001,
		TCPMapped: true, UDPMapped: false, VerificationStatus: "verified",
	}
	if err := r.SaveIXPortResource(resource); err != nil {
		t.Fatal(err)
	}
	profile := &model.IXPathProfile{Name: "scoped", EntryNodeID: 141, IXNodeID: 142, EntryPoolID: entry.ID, IXPoolID: ix.ID}
	if err := r.CreateIXPathProfile(profile); err != nil {
		t.Fatal(err)
	}
	if _, err := r.ValidateIXPathProfileForRoute(profile.ID, []string{"tcp"}); err != nil {
		t.Fatalf("TCP-compatible path rejected: %v", err)
	}
	if _, err := r.ValidateIXPathProfileForRoute(profile.ID, []string{"udp"}); err == nil {
		t.Fatal("UDP route accepted a TCP-only IX resource")
	}

	profile.Status = 0
	if err := r.UpdateIXPathProfile(profile); err != nil {
		t.Fatal(err)
	}
	if _, err := r.ValidateIXPathProfileForRoute(profile.ID, []string{"tcp"}); err == nil {
		t.Fatal("disabled PathProfile remained available")
	}

	profile.Status = 1
	if err := r.UpdateIXPathProfile(profile); err != nil {
		t.Fatal(err)
	}
	ix.Status = 0
	if err := r.UpdateIXPortPool(ix); err != nil {
		t.Fatal(err)
	}
	if _, err := r.ValidateIXPathProfileForRoute(profile.ID, []string{"tcp"}); err == nil {
		t.Fatal("disabled IX pool remained available")
	}
}
