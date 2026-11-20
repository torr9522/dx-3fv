package repo

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"go-backend/internal/store/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IXRouteDiagnosisEvidence struct {
	Diagnosis model.IXRouteDiagnosis `json:"diagnosis"`
	Result    interface{}            `json:"result"`
}

func validateIXPathProfile(db *gorm.DB, profile *model.IXPathProfile) error {
	if profile == nil || strings.TrimSpace(profile.Name) == "" {
		return errors.New("path name is required")
	}
	if profile.Status != 0 && profile.Status != 1 {
		return errors.New("invalid path status")
	}
	if profile.EntryNodeID <= 0 || profile.IXNodeID <= 0 || profile.EntryNodeID == profile.IXNodeID {
		return errors.New("entry and IX must be different Node identities")
	}
	var nodeCount int64
	if err := db.Model(&model.Node{}).Where("id IN ?", []int64{profile.EntryNodeID, profile.IXNodeID}).Count(&nodeCount).Error; err != nil {
		return err
	}
	if nodeCount != 2 {
		return errors.New("path node not found")
	}
	var entryPool model.IXPortPool
	if err := db.First(&entryPool, profile.EntryPoolID).Error; err != nil {
		return err
	}
	var ixPool model.IXPortPool
	if err := db.First(&ixPool, profile.IXPoolID).Error; err != nil {
		return err
	}
	if entryPool.NodeID != profile.EntryNodeID {
		return errors.New("entry pool does not belong to entry Node identity")
	}
	if ixPool.NodeID != profile.IXNodeID {
		return errors.New("IX pool does not belong to IX Node identity")
	}
	if entryPool.Kind != IXPoolKindFull {
		return errors.New("entry Node requires a full listener-port policy")
	}
	return nil
}

func (r *Repository) CreateIXPathProfile(profile *model.IXPathProfile) error {
	if r == nil || r.db == nil {
		return errors.New("repository not initialized")
	}
	profile.Status = 1
	if err := validateIXPathProfile(r.db, profile); err != nil {
		return err
	}
	now := time.Now().UnixMilli()
	profile.Name = strings.TrimSpace(profile.Name)
	profile.CreatedTime = now
	profile.UpdatedTime = now
	return r.db.Create(profile).Error
}

func (r *Repository) UpdateIXPathProfile(profile *model.IXPathProfile) error {
	if r == nil || r.db == nil || profile == nil || profile.ID <= 0 {
		return errors.New("invalid path profile")
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		var current model.IXPathProfile
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&current, profile.ID).Error; err != nil {
			return err
		}
		if err := validateIXPathProfile(tx, profile); err != nil {
			return err
		}
		var activeRoutes int64
		if err := tx.Model(&model.IXRoute{}).Where("path_profile_id = ? AND desired_state <> ?", profile.ID, "deleted").Count(&activeRoutes).Error; err != nil {
			return err
		}
		identityChanged := current.EntryNodeID != profile.EntryNodeID || current.IXNodeID != profile.IXNodeID || current.EntryPoolID != profile.EntryPoolID || current.IXPoolID != profile.IXPoolID
		if activeRoutes > 0 && identityChanged {
			return errors.New("path Node and pool identities cannot change while active businesses exist")
		}
		return tx.Model(&model.IXPathProfile{}).Where("id = ?", profile.ID).Updates(map[string]interface{}{
			"name": strings.TrimSpace(profile.Name), "entry_node_id": profile.EntryNodeID,
			"ix_node_id": profile.IXNodeID, "entry_pool_id": profile.EntryPoolID,
			"ix_pool_id": profile.IXPoolID, "status": profile.Status,
			"updated_time": time.Now().UnixMilli(),
		}).Error
	})
}

func (r *Repository) GetIXPathProfile(id int64) (*model.IXPathProfile, error) {
	var profile model.IXPathProfile
	err := r.db.First(&profile, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &profile, err
}

func (r *Repository) ListIXPathProfiles() ([]model.IXPathProfile, error) {
	var profiles []model.IXPathProfile
	err := r.db.Order("id ASC").Find(&profiles).Error
	return profiles, err
}

func (r *Repository) DeleteIXPathProfile(id int64) error {
	if r == nil || r.db == nil || id <= 0 {
		return errors.New("invalid path profile")
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.IXRoute{}).Where("path_profile_id = ? AND desired_state <> ?", id, "deleted").Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("path profile is still used by active businesses")
		}
		return tx.Delete(&model.IXPathProfile{}, id).Error
	})
}

func (r *Repository) CreateIXRoute(route *model.IXRoute) error {
	if r == nil || r.db == nil || route == nil {
		return errors.New("repository not initialized")
	}
	if route.PathProfileID <= 0 || strings.TrimSpace(route.Name) == "" || strings.TrimSpace(route.LandingHost) == "" || route.LandingPort <= 0 || route.LandingPort > 65535 {
		return errors.New("invalid multiport business")
	}
	if route.ProtocolMode != "tcp" && route.ProtocolMode != "udp" && route.ProtocolMode != "tcp_udp" {
		return errors.New("invalid protocol mode")
	}
	now := time.Now().UnixMilli()
	route.Name = strings.TrimSpace(route.Name)
	route.LandingHost = strings.TrimSpace(route.LandingHost)
	route.DesiredState = "active"
	route.ObservedState = "pending"
	route.CreatedTime = now
	route.UpdatedTime = now
	return r.db.Transaction(func(tx *gorm.DB) error {
		var profile model.IXPathProfile
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&profile, route.PathProfileID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("path profile is unavailable")
			}
			return err
		}
		if profile.Status != 1 {
			return errors.New("path profile is unavailable")
		}
		return tx.Create(route).Error
	})
}

func (r *Repository) GetIXRoute(id int64) (*model.IXRoute, error) {
	var route model.IXRoute
	err := r.db.First(&route, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &route, err
}

func (r *Repository) ListIXRoutes() ([]model.IXRoute, error) {
	var routes []model.IXRoute
	err := r.db.Where("desired_state <> ?", "deleted").Order("path_profile_id ASC, id ASC").Find(&routes).Error
	return routes, err
}

func (r *Repository) SaveIXRouteDiagnosis(item *model.IXRouteDiagnosis, result interface{}) error {
	if r == nil || r.db == nil || item == nil || item.RouteID <= 0 {
		return errors.New("invalid route diagnosis")
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		return err
	}
	item.ResultJSON = string(encoded)
	return r.db.Create(item).Error
}

func (r *Repository) GetLatestIXRouteDiagnosis(routeID int64) (*IXRouteDiagnosisEvidence, error) {
	if r == nil || r.db == nil || routeID <= 0 {
		return nil, errors.New("invalid route diagnosis")
	}
	var item model.IXRouteDiagnosis
	err := r.db.Where("route_id = ?", routeID).Order("id DESC").First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var result interface{}
	if err := json.Unmarshal([]byte(item.ResultJSON), &result); err != nil {
		return nil, err
	}
	return &IXRouteDiagnosisEvidence{Diagnosis: item, Result: result}, nil
}

func (r *Repository) UpdateIXRouteAllocation(id int64, entryPort, ixPublicPort, ixInternalPort int) error {
	return r.db.Model(&model.IXRoute{}).Where("id = ?", id).Updates(map[string]interface{}{
		"entry_port": entryPort, "ix_public_port": ixPublicPort, "ix_internal_port": ixInternalPort,
		"updated_time": time.Now().UnixMilli(),
	}).Error
}

func (r *Repository) SetIXRouteState(id int64, desired, observed, lastError string) error {
	updates := map[string]interface{}{"updated_time": time.Now().UnixMilli()}
	if strings.TrimSpace(desired) != "" {
		updates["desired_state"] = strings.TrimSpace(desired)
	}
	if strings.TrimSpace(observed) != "" {
		updates["observed_state"] = strings.TrimSpace(observed)
	}
	updates["last_error"] = strings.TrimSpace(lastError)
	return r.db.Model(&model.IXRoute{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) SaveIXRouteRuntimes(items []model.IXRouteRuntime) error {
	if len(items) == 0 {
		return errors.New("runtime desired state is empty")
	}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "route_id"}, {Name: "node_id"}, {Name: "role"}, {Name: "protocol"}},
		DoUpdates: clause.AssignmentColumns([]string{"service_name", "config", "desired_state", "observed_state", "applied", "last_error", "generation", "updated_time"}),
	}).Create(&items).Error
}

func (r *Repository) ListIXRouteRuntimes(routeID int64) ([]model.IXRouteRuntime, error) {
	var items []model.IXRouteRuntime
	err := r.db.Where("route_id = ?", routeID).Order("role DESC, protocol ASC").Find(&items).Error
	return items, err
}

func (r *Repository) ListIXRouteRuntimesByNode(nodeID int64) ([]model.IXRouteRuntime, error) {
	var items []model.IXRouteRuntime
	err := r.db.Where("node_id = ?", nodeID).Order("route_id ASC, role DESC, protocol ASC").Find(&items).Error
	return items, err
}

func (r *Repository) MarkIXRouteRuntimeObserved(id int64, state, lastError string, applied bool) error {
	return r.db.Model(&model.IXRouteRuntime{}).Where("id = ?", id).Updates(map[string]interface{}{
		"observed_state": state, "last_error": strings.TrimSpace(lastError), "applied": applied, "updated_time": time.Now().UnixMilli(),
	}).Error
}

func (r *Repository) SetIXRouteRuntimesDesired(routeID int64, desired string) error {
	return r.db.Model(&model.IXRouteRuntime{}).Where("route_id = ?", routeID).Updates(map[string]interface{}{
		"desired_state": desired, "observed_state": "pending", "applied": false, "last_error": "", "updated_time": time.Now().UnixMilli(),
	}).Error
}

func (r *Repository) ListPendingIXRouteIDs() ([]int64, error) {
	var ids []int64
	err := r.db.Model(&model.IXRoute{}).
		Where("observed_state IN ?", []string{"pending", "degraded", "error", "cleanup_pending"}).
		Order("id ASC").Pluck("id", &ids).Error
	if err != nil {
		return nil, err
	}
	var migrationIDs []int64
	if err := r.db.Model(&model.IXRouteMigration{}).Where("active_key LIKE ?", "route:%").Pluck("route_id", &migrationIDs).Error; err != nil {
		return nil, err
	}
	seen := make(map[int64]bool, len(ids)+len(migrationIDs))
	out := make([]int64, 0, len(ids)+len(migrationIDs))
	for _, id := range append(ids, migrationIDs...) {
		if id > 0 && !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out, nil
}

func (r *Repository) ListIXRouteIDsByNode(nodeID int64) ([]int64, error) {
	var ids []int64
	err := r.db.Model(&model.IXRouteRuntime{}).Where("node_id = ?", nodeID).Distinct().Order("route_id ASC").Pluck("route_id", &ids).Error
	return ids, err
}

func (r *Repository) AddIXRouteFlowBatch(nodeID int64, deltas []IXRouteFlowDelta) error {
	if r == nil || r.db == nil {
		return errors.New("repository not initialized")
	}
	if nodeID <= 0 || len(deltas) == 0 {
		return nil
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		return addIXRouteFlowBatchTx(tx, nodeID, deltas)
	})
}

func addIXRouteFlowBatchTx(tx *gorm.DB, nodeID int64, deltas []IXRouteFlowDelta) error {
	if tx == nil {
		return errors.New("database unavailable")
	}
	for _, delta := range deltas {
		if delta.RouteID <= 0 || delta.InFlow < 0 || delta.OutFlow < 0 {
			continue
		}
		var owned int64
		if err := tx.Model(&model.IXRouteRuntime{}).Where("route_id = ? AND node_id = ? AND role = ?", delta.RouteID, nodeID, "entry").Count(&owned).Error; err != nil {
			return err
		}
		if owned == 0 {
			continue
		}
		if err := tx.Model(&model.IXRoute{}).Where("id = ?", delta.RouteID).UpdateColumns(map[string]interface{}{
			"in_flow":  gorm.Expr("in_flow + ?", delta.InFlow),
			"out_flow": gorm.Expr("out_flow + ?", delta.OutFlow),
		}).Error; err != nil {
			return err
		}
	}
	return nil
}
