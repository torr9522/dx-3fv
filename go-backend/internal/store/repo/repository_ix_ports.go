package repo

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"go-backend/internal/store/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	IXPoolKindFull = "full"
	IXPoolKindNAT  = "nat"
)

type IXPortAllocationRequest struct {
	PoolID        int64
	RequestedPort int
	Protocols     []string
	OwnerType     string
	OwnerID       int64
	Manual        bool
}

type IXPortAllocation struct {
	Resource model.IXPortResource
	Leases   []model.IXPortLease
}

func normalizeIXProtocols(protocols []string) ([]string, error) {
	seen := map[string]bool{}
	for _, protocol := range protocols {
		protocol = strings.ToLower(strings.TrimSpace(protocol))
		if protocol != "tcp" && protocol != "udp" {
			return nil, fmt.Errorf("unsupported protocol %q", protocol)
		}
		seen[protocol] = true
	}
	if len(seen) == 0 {
		return nil, errors.New("at least one protocol is required")
	}
	out := make([]string, 0, len(seen))
	for protocol := range seen {
		out = append(out, protocol)
	}
	sort.Strings(out)
	return out, nil
}

func activeIXLeaseKey(nodeID int64, port int, protocol string) string {
	return fmt.Sprintf("%d:%d:%s", nodeID, port, protocol)
}

func (r *Repository) CreateIXPortPool(pool *model.IXPortPool) error {
	if r == nil || r.db == nil || pool == nil {
		return errors.New("repository not initialized")
	}
	pool.Name = strings.TrimSpace(pool.Name)
	pool.Kind = strings.ToLower(strings.TrimSpace(pool.Kind))
	if pool.NodeID <= 0 || pool.Name == "" {
		return errors.New("node and pool name are required")
	}
	node, err := r.GetNodeByID(pool.NodeID)
	if err != nil {
		return err
	}
	if node == nil {
		return errors.New("port pool node not found")
	}
	if pool.Kind != IXPoolKindFull && pool.Kind != IXPoolKindNAT {
		return errors.New("pool kind must be full or nat")
	}
	if pool.Kind == IXPoolKindFull {
		// Direct nodes such as GZ2 have a full listener-port policy. Keep the
		// eligibility range authoritative here so non-HTTP callers cannot turn
		// it into a synthetic provider-style narrow pool.
		pool.RangeStart, pool.RangeEnd = 1, 65535
	}
	if pool.RangeStart <= 0 {
		pool.RangeStart = 1
	}
	if pool.RangeEnd <= 0 {
		pool.RangeEnd = 65535
	}
	if pool.RangeStart > pool.RangeEnd || pool.RangeEnd > 65535 {
		return errors.New("invalid port range")
	}
	if pool.NextPort < pool.RangeStart || pool.NextPort > pool.RangeEnd {
		pool.NextPort = pool.RangeStart
	}
	now := time.Now().UnixMilli()
	pool.CreatedTime = now
	pool.UpdatedTime = now
	if pool.Status == 0 {
		pool.Status = 1
	}
	return r.db.Create(pool).Error
}

func (r *Repository) GetIXPortPool(id int64) (*model.IXPortPool, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("repository not initialized")
	}
	var pool model.IXPortPool
	err := r.db.First(&pool, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &pool, err
}

func (r *Repository) UpdateIXPortPool(update *model.IXPortPool) error {
	if r == nil || r.db == nil || update == nil || update.ID <= 0 {
		return errors.New("invalid port pool")
	}
	current, err := r.GetIXPortPool(update.ID)
	if err != nil {
		return err
	}
	if current == nil {
		return errors.New("port pool not found")
	}
	name := strings.TrimSpace(update.Name)
	if name == "" {
		return errors.New("pool name is required")
	}
	start, end := update.RangeStart, update.RangeEnd
	if current.Kind == IXPoolKindFull {
		start, end = 1, 65535
	}
	if start <= 0 || end <= 0 || start > end || end > 65535 {
		return errors.New("invalid port range")
	}
	status := update.Status
	if status != 0 && status != 1 {
		return errors.New("invalid pool status")
	}
	now := time.Now().UnixMilli()
	return r.db.Model(&model.IXPortPool{}).Where("id = ?", current.ID).Updates(map[string]interface{}{
		"name": name, "range_start": start, "range_end": end, "status": status, "updated_time": now,
	}).Error
}

func (r *Repository) SaveIXPortResource(resource *model.IXPortResource) error {
	if r == nil || r.db == nil || resource == nil {
		return errors.New("repository not initialized")
	}
	if resource.PoolID <= 0 || resource.NodeID <= 0 || resource.InternalPort <= 0 || resource.InternalPort > 65535 {
		return errors.New("invalid port resource")
	}
	pool, err := r.GetIXPortPool(resource.PoolID)
	if err != nil {
		return err
	}
	if pool == nil || pool.NodeID != resource.NodeID {
		return errors.New("resource Node identity does not match its pool")
	}
	if resource.PublicPort == 0 {
		resource.PublicPort = resource.InternalPort
	}
	if resource.PublicPort < 1 || resource.PublicPort > 65535 {
		return errors.New("invalid public port")
	}
	if pool.Kind == IXPoolKindNAT && (resource.PublicPort < pool.RangeStart || resource.PublicPort > pool.RangeEnd) {
		return errors.New("public port is outside the declared NAT range")
	}
	if resource.InternalPort == 22 {
		resource.Excluded = true
		if strings.TrimSpace(resource.ReservedReason) == "" {
			resource.ReservedReason = "SSH"
		}
	}
	resource.VerificationStatus = strings.ToLower(strings.TrimSpace(resource.VerificationStatus))
	if resource.VerificationStatus == "" {
		resource.VerificationStatus = "unverified"
	}
	now := time.Now().UnixMilli()
	if resource.CreatedTime == 0 {
		resource.CreatedTime = now
	}
	resource.UpdatedTime = now
	return r.db.Save(resource).Error
}

func (r *Repository) GetIXPortResource(id int64) (*model.IXPortResource, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("repository not initialized")
	}
	var resource model.IXPortResource
	err := r.db.First(&resource, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &resource, err
}

func (r *Repository) ListIXPortPools() ([]model.IXPortPool, error) {
	var pools []model.IXPortPool
	err := r.db.Order("node_id ASC, id ASC").Find(&pools).Error
	return pools, err
}

func (r *Repository) ListIXPortResources(poolID int64) ([]model.IXPortResource, error) {
	var resources []model.IXPortResource
	err := r.db.Where("pool_id = ?", poolID).Order("public_port ASC, internal_port ASC").Find(&resources).Error
	return resources, err
}

func (r *Repository) ListIXPortLeases(ownerType string, ownerID int64) ([]model.IXPortLease, error) {
	var leases []model.IXPortLease
	query := r.db.Order("id ASC")
	if strings.TrimSpace(ownerType) != "" {
		query = query.Where("owner_type = ?", strings.TrimSpace(ownerType))
	}
	if ownerID > 0 {
		query = query.Where("owner_id = ?", ownerID)
	}
	err := query.Find(&leases).Error
	return leases, err
}

func (r *Repository) AllocateIXPorts(req IXPortAllocationRequest) (*IXPortAllocation, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("repository not initialized")
	}
	protocols, err := normalizeIXProtocols(req.Protocols)
	if err != nil {
		return nil, err
	}
	if req.PoolID <= 0 || req.OwnerID <= 0 || strings.TrimSpace(req.OwnerType) == "" {
		return nil, errors.New("pool and owner are required")
	}
	if req.RequestedPort < 0 || req.RequestedPort > 65535 {
		return nil, errors.New("invalid requested port")
	}

	var allocated IXPortAllocation
	err = r.db.Transaction(func(tx *gorm.DB) error {
		var pool model.IXPortPool
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&pool, req.PoolID).Error; err != nil {
			return err
		}
		if pool.Status != 1 {
			return errors.New("port pool is disabled")
		}

		candidates, err := ixPortCandidates(tx, &pool, req.RequestedPort, protocols)
		if err != nil {
			return err
		}
		for i := range candidates {
			resource := candidates[i]
			leases := make([]model.IXPortLease, 0, len(protocols))
			now := time.Now().UnixMilli()
			for _, protocol := range protocols {
				leases = append(leases, model.IXPortLease{
					PoolID:       pool.ID,
					ResourceID:   resource.ID,
					NodeID:       pool.NodeID,
					InternalPort: resource.InternalPort,
					Protocol:     protocol,
					LeaseKey:     activeIXLeaseKey(pool.NodeID, resource.InternalPort, protocol),
					OwnerType:    strings.TrimSpace(req.OwnerType),
					OwnerID:      req.OwnerID,
					Manual:       req.Manual || req.RequestedPort > 0,
					State:        "active",
					CreatedTime:  now,
					UpdatedTime:  now,
				})
			}
			if err := tx.Create(&leases).Error; err != nil {
				if isUniqueConstraintError(err) && req.RequestedPort == 0 {
					continue
				}
				if isUniqueConstraintError(err) {
					return errors.New("requested port is already leased")
				}
				return err
			}
			allocated = IXPortAllocation{Resource: resource, Leases: leases}
			next := resource.InternalPort + 1
			if next > pool.RangeEnd {
				next = pool.RangeStart
			}
			return tx.Model(&model.IXPortPool{}).Where("id = ?", pool.ID).Updates(map[string]interface{}{"next_port": next, "updated_time": now}).Error
		}
		if req.RequestedPort > 0 {
			return errors.New("requested port is unavailable")
		}
		return errors.New("port pool has no allocatable resource")
	})
	if err != nil {
		return nil, err
	}
	return &allocated, nil
}

func ixPortCandidates(tx *gorm.DB, pool *model.IXPortPool, requested int, protocols []string) ([]model.IXPortResource, error) {
	if pool.Kind == IXPoolKindNAT {
		query := tx.Where("pool_id = ? AND excluded = ? AND verification_status = ?", pool.ID, false, "verified")
		if requested > 0 {
			query = query.Where("public_port = ?", requested)
		}
		if containsIXProtocol(protocols, "tcp") {
			query = query.Where("tcp_mapped = ?", true)
		}
		if containsIXProtocol(protocols, "udp") {
			query = query.Where("udp_mapped = ?", true)
		}
		var resources []model.IXPortResource
		if err := query.Order("public_port ASC, id ASC").Find(&resources).Error; err != nil {
			return nil, err
		}
		available := resources[:0]
		for i := range resources {
			occupied, err := ixPortAlreadyOwned(tx, pool.NodeID, resources[i].InternalPort, protocols)
			if err != nil {
				return nil, err
			}
			if !occupied {
				available = append(available, resources[i])
			}
		}
		return available, nil
	}

	start := pool.NextPort
	if start < pool.RangeStart || start > pool.RangeEnd {
		start = pool.RangeStart
	}
	ports := make([]int, 0, pool.RangeEnd-pool.RangeStart+1)
	if requested > 0 {
		if requested < pool.RangeStart || requested > pool.RangeEnd {
			return nil, errors.New("requested port is outside the eligible range")
		}
		ports = append(ports, requested)
	} else {
		for port := start; port <= pool.RangeEnd; port++ {
			ports = append(ports, port)
		}
		for port := pool.RangeStart; port < start; port++ {
			ports = append(ports, port)
		}
	}

	for _, port := range ports {
		if port == 22 {
			continue
		}
		var resource model.IXPortResource
		err := tx.Where("node_id = ? AND internal_port = ?", pool.NodeID, port).First(&resource).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			now := time.Now().UnixMilli()
			resource = model.IXPortResource{PoolID: pool.ID, NodeID: pool.NodeID, PublicPort: port, InternalPort: port, TCPMapped: true, UDPMapped: true, VerificationStatus: "local", CreatedTime: now, UpdatedTime: now}
			if err = tx.Create(&resource).Error; err != nil {
				if isUniqueConstraintError(err) {
					if err = tx.Where("node_id = ? AND internal_port = ?", pool.NodeID, port).First(&resource).Error; err != nil {
						return nil, err
					}
				} else {
					return nil, err
				}
			}
		} else if err != nil {
			return nil, err
		}
		if resource.PoolID != pool.ID || resource.Excluded {
			continue
		}
		occupied, err := ixPortAlreadyOwned(tx, pool.NodeID, resource.InternalPort, protocols)
		if err != nil {
			return nil, err
		}
		if occupied {
			continue
		}
		// AllocateIXPorts holds the pool row lock for this transaction and only
		// consumes one candidate. Returning immediately avoids materializing and
		// querying the entire 1-65535 full-pool range for every allocation.
		return []model.IXPortResource{resource}, nil
	}
	return nil, nil
}

func ixPortAlreadyOwned(tx *gorm.DB, nodeID int64, port int, protocols []string) (bool, error) {
	var count int64
	if err := tx.Model(&model.ForwardPort{}).Where("node_id = ? AND port = ?", nodeID, port).Count(&count).Error; err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	if err := tx.Model(&model.PeerShareRuntime{}).Where("node_id = ? AND port = ? AND status = ?", nodeID, port, 1).Count(&count).Error; err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	if err := tx.Model(&model.ChainTunnel{}).Where("node_id = ? AND port = ?", nodeID, port).Count(&count).Error; err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	for _, protocol := range protocols {
		if err := tx.Model(&model.IXPortObservation{}).
			Where("node_id = ? AND internal_port = ? AND protocol = ? AND state IN ?", nodeID, port, protocol, []string{"listening", "in_use", "conflict"}).
			Count(&count).Error; err != nil {
			return false, err
		}
		if count > 0 {
			return true, nil
		}
	}
	return false, nil
}

func containsIXProtocol(protocols []string, expected string) bool {
	for _, protocol := range protocols {
		if protocol == expected {
			return true
		}
	}
	return false
}

func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unique constraint") || strings.Contains(message, "duplicate key") || strings.Contains(message, "is not unique")
}

func (r *Repository) RequestIXPortLeaseRelease(ownerType string, ownerID int64) error {
	if r == nil || r.db == nil {
		return errors.New("repository not initialized")
	}
	now := time.Now().UnixMilli()
	return r.db.Model(&model.IXPortLease{}).
		Where("owner_type = ? AND owner_id = ? AND state = ?", strings.TrimSpace(ownerType), ownerID, "active").
		Updates(map[string]interface{}{"state": "pending_release", "release_requested": now, "updated_time": now}).Error
}

func (r *Repository) CompleteIXPortLeaseRelease(ownerType string, ownerID int64) error {
	if r == nil || r.db == nil {
		return errors.New("repository not initialized")
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		var leases []model.IXPortLease
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("owner_type = ? AND owner_id = ? AND state = ?", strings.TrimSpace(ownerType), ownerID, "pending_release").Find(&leases).Error; err != nil {
			return err
		}
		now := time.Now().UnixMilli()
		for _, lease := range leases {
			tombstone := fmt.Sprintf("released:%d:%d", lease.ID, now)
			if err := tx.Model(&model.IXPortLease{}).Where("id = ?", lease.ID).Updates(map[string]interface{}{
				"lease_key": tombstone, "state": "released", "released_time": now, "updated_time": now,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) UpsertIXPortObservation(observation *model.IXPortObservation) error {
	if r == nil || r.db == nil || observation == nil {
		return errors.New("repository not initialized")
	}
	if observation.NodeID <= 0 || observation.InternalPort <= 0 || (observation.Protocol != "tcp" && observation.Protocol != "udp") {
		return errors.New("invalid port observation")
	}
	observation.ObservedTime = time.Now().UnixMilli()
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "node_id"}, {Name: "internal_port"}, {Name: "protocol"}},
		DoUpdates: clause.AssignmentColumns([]string{"runtime_name", "state", "error", "observed_time"}),
	}).Create(observation).Error
}
