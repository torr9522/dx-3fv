package repo

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-backend/internal/store/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ixRouteRuntimeSnapshot struct {
	ID            int64  `json:"id"`
	RouteID       int64  `json:"routeId"`
	NodeID        int64  `json:"nodeId"`
	Role          string `json:"role"`
	Protocol      string `json:"protocol"`
	ServiceName   string `json:"serviceName"`
	Config        string `json:"config"`
	DesiredState  string `json:"desiredState"`
	ObservedState string `json:"observedState"`
	Applied       bool   `json:"applied"`
	LastError     string `json:"lastError"`
	Generation    int64  `json:"generation"`
	CreatedTime   int64  `json:"createdTime"`
	UpdatedTime   int64  `json:"updatedTime"`
}

func snapshotIXRouteRuntimes(items []model.IXRouteRuntime) []ixRouteRuntimeSnapshot {
	out := make([]ixRouteRuntimeSnapshot, 0, len(items))
	for _, item := range items {
		out = append(out, ixRouteRuntimeSnapshot{
			ID: item.ID, RouteID: item.RouteID, NodeID: item.NodeID, Role: item.Role,
			Protocol: item.Protocol, ServiceName: item.ServiceName, Config: item.Config,
			DesiredState: item.DesiredState, ObservedState: item.ObservedState,
			Applied: item.Applied, LastError: item.LastError, Generation: item.Generation,
			CreatedTime: item.CreatedTime, UpdatedTime: item.UpdatedTime,
		})
	}
	return out
}

func restoreIXRouteRuntimes(items []ixRouteRuntimeSnapshot) []model.IXRouteRuntime {
	out := make([]model.IXRouteRuntime, 0, len(items))
	for _, item := range items {
		out = append(out, model.IXRouteRuntime{
			ID: item.ID, RouteID: item.RouteID, NodeID: item.NodeID, Role: item.Role,
			Protocol: item.Protocol, ServiceName: item.ServiceName, Config: item.Config,
			DesiredState: item.DesiredState, ObservedState: item.ObservedState,
			Applied: item.Applied, LastError: item.LastError, Generation: item.Generation,
			CreatedTime: item.CreatedTime, UpdatedTime: item.UpdatedTime,
		})
	}
	return out
}

func (r *Repository) BeginIXRouteMigration(routeID int64) (*model.IXRouteMigration, error) {
	if r == nil || r.db == nil || routeID <= 0 {
		return nil, errors.New("invalid route migration")
	}
	now := time.Now().UnixMilli()
	item := &model.IXRouteMigration{RouteID: routeID, ActiveKey: fmt.Sprintf("route:%d", routeID), State: "preparing", ReleaseLeaseIDsJSON: "[]", CreatedTime: now, UpdatedTime: now}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var route model.IXRoute
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&route, routeID).Error; err != nil {
			return err
		}
		if route.DesiredState == "deleted" {
			return errors.New("deleted route cannot be edited")
		}
		if err := tx.Create(item).Error; err != nil {
			if isUniqueConstraintError(err) {
				return errors.New("route already has an active migration")
			}
			return err
		}
		return nil
	})
	return item, err
}

func (r *Repository) SaveIXRouteMigrationPlan(item *model.IXRouteMigration, oldRoute, newRoute *model.IXRoute, oldRuntimes, newRuntimes []model.IXRouteRuntime, releaseLeaseIDs []int64) error {
	if r == nil || r.db == nil || item == nil || item.ID <= 0 {
		return errors.New("invalid route migration")
	}
	encode := func(value interface{}) (string, error) {
		data, err := json.Marshal(value)
		return string(data), err
	}
	oldRouteJSON, err := encode(oldRoute)
	if err != nil {
		return err
	}
	newRouteJSON, err := encode(newRoute)
	if err != nil {
		return err
	}
	oldRuntimesJSON, err := encode(snapshotIXRouteRuntimes(oldRuntimes))
	if err != nil {
		return err
	}
	newRuntimesJSON, err := encode(snapshotIXRouteRuntimes(newRuntimes))
	if err != nil {
		return err
	}
	releaseJSON, err := encode(releaseLeaseIDs)
	if err != nil {
		return err
	}
	return r.db.Model(&model.IXRouteMigration{}).Where("id = ? AND active_key = ?", item.ID, fmt.Sprintf("route:%d", item.RouteID)).Updates(map[string]interface{}{
		"state": "ready", "old_route_json": oldRouteJSON, "new_route_json": newRouteJSON,
		"old_runtimes_json": oldRuntimesJSON, "new_runtimes_json": newRuntimesJSON,
		"release_lease_ids_json": releaseJSON, "updated_time": time.Now().UnixMilli(),
	}).Error
}

func (r *Repository) SetIXRouteMigrationState(id int64, state, lastError string) error {
	return r.db.Model(&model.IXRouteMigration{}).Where("id = ?", id).Updates(map[string]interface{}{
		"state": strings.TrimSpace(state), "last_error": strings.TrimSpace(lastError), "updated_time": time.Now().UnixMilli(),
	}).Error
}

func (r *Repository) GetActiveIXRouteMigration(routeID int64) (*model.IXRouteMigration, error) {
	var item model.IXRouteMigration
	err := r.db.Where("active_key = ?", fmt.Sprintf("route:%d", routeID)).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &item, err
}

func DecodeIXRouteMigrationPlan(item *model.IXRouteMigration) (*model.IXRoute, *model.IXRoute, []model.IXRouteRuntime, []model.IXRouteRuntime, []int64, error) {
	if item == nil {
		return nil, nil, nil, nil, nil, errors.New("missing route migration")
	}
	var oldRoute, newRoute model.IXRoute
	var oldRuntimeSnapshots, newRuntimeSnapshots []ixRouteRuntimeSnapshot
	var releaseIDs []int64
	for _, part := range []struct {
		raw string
		out interface{}
	}{
		{item.OldRouteJSON, &oldRoute}, {item.NewRouteJSON, &newRoute},
		{item.OldRuntimesJSON, &oldRuntimeSnapshots}, {item.NewRuntimesJSON, &newRuntimeSnapshots},
		{item.ReleaseLeaseIDsJSON, &releaseIDs},
	} {
		if strings.TrimSpace(part.raw) == "" {
			return nil, nil, nil, nil, nil, errors.New("route migration plan is incomplete")
		}
		if err := json.Unmarshal([]byte(part.raw), part.out); err != nil {
			return nil, nil, nil, nil, nil, err
		}
	}
	return &oldRoute, &newRoute, restoreIXRouteRuntimes(oldRuntimeSnapshots), restoreIXRouteRuntimes(newRuntimeSnapshots), releaseIDs, nil
}

func releaseIXLeaseIDsTx(tx *gorm.DB, ids []int64, now int64) error {
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		var lease model.IXPortLease
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&lease, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return err
		}
		if lease.State == "released" {
			continue
		}
		if err := tx.Model(&model.IXPortLease{}).Where("id = ?", id).Updates(map[string]interface{}{
			"lease_key": fmt.Sprintf("released:%d:%d", id, now), "state": "released",
			"release_requested": now, "released_time": now, "updated_time": now,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) RollbackIXRouteMigration(item *model.IXRouteMigration, lastError string) error {
	if r == nil || r.db == nil || item == nil || item.ID <= 0 {
		return errors.New("invalid route migration")
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		var leases []model.IXPortLease
		if err := tx.Where("owner_id = ? AND owner_type IN ? AND state <> ?", item.ID, []string{"ix_route_migration_entry", "ix_route_migration_ix"}, "released").Find(&leases).Error; err != nil {
			return err
		}
		ids := make([]int64, 0, len(leases))
		for _, lease := range leases {
			ids = append(ids, lease.ID)
		}
		now := time.Now().UnixMilli()
		if err := releaseIXLeaseIDsTx(tx, ids, now); err != nil {
			return err
		}
		return tx.Model(&model.IXRouteMigration{}).Where("id = ?", item.ID).Updates(map[string]interface{}{
			"active_key": fmt.Sprintf("done:%d:%d", item.ID, now), "state": "rolled_back",
			"last_error": strings.TrimSpace(lastError), "updated_time": now,
		}).Error
	})
}

func (r *Repository) CommitIXRouteMigration(item *model.IXRouteMigration) error {
	oldRoute, newRoute, _, newRuntimes, releaseIDs, err := DecodeIXRouteMigrationPlan(item)
	if err != nil {
		return err
	}
	if oldRoute.ID != item.RouteID || newRoute.ID != item.RouteID || len(newRuntimes) == 0 {
		return errors.New("route migration identity mismatch")
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		var current model.IXRoute
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&current, item.RouteID).Error; err != nil {
			return err
		}
		if current.UpdatedTime != oldRoute.UpdatedTime {
			return errors.New("route changed while migration was active")
		}
		now := time.Now().UnixMilli()
		if err := tx.Model(&model.IXRoute{}).Where("id = ?", item.RouteID).Updates(map[string]interface{}{
			"name": newRoute.Name, "entry_port": newRoute.EntryPort,
			"ix_public_port": newRoute.IXPublicPort, "ix_internal_port": newRoute.IXInternalPort,
			"landing_host": newRoute.LandingHost, "landing_port": newRoute.LandingPort,
			"protocol_mode": newRoute.ProtocolMode, "desired_state": "active",
			"observed_state": "active", "last_error": "", "updated_time": now,
		}).Error; err != nil {
			return err
		}
		if err := tx.Where("route_id = ?", item.RouteID).Delete(&model.IXRouteRuntime{}).Error; err != nil {
			return err
		}
		for i := range newRuntimes {
			newRuntimes[i].ID = 0
			newRuntimes[i].ObservedState = "active"
			newRuntimes[i].Applied = true
			newRuntimes[i].LastError = ""
			newRuntimes[i].UpdatedTime = now
		}
		if err := tx.Create(&newRuntimes).Error; err != nil {
			return err
		}
		for from, to := range map[string]string{"ix_route_migration_entry": "ix_route_entry", "ix_route_migration_ix": "ix_route_ix"} {
			if err := tx.Model(&model.IXPortLease{}).Where("owner_type = ? AND owner_id = ? AND state = ?", from, item.ID, "active").Updates(map[string]interface{}{"owner_type": to, "owner_id": item.RouteID, "updated_time": now}).Error; err != nil {
				return err
			}
		}
		if err := releaseIXLeaseIDsTx(tx, releaseIDs, now); err != nil {
			return err
		}
		return tx.Model(&model.IXRouteMigration{}).Where("id = ?", item.ID).Updates(map[string]interface{}{
			"active_key": fmt.Sprintf("done:%d:%d", item.ID, now), "state": "complete", "last_error": "", "updated_time": now,
		}).Error
	})
}
