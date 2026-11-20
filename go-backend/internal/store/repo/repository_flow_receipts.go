package repo

import (
	"errors"
	"sort"
	"strings"
	"time"

	"go-backend/internal/store/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func validFlowReportID(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) != 32 {
		return false
	}
	for _, ch := range value {
		if (ch < '0' || ch > '9') && (ch < 'a' || ch > 'f') {
			return false
		}
	}
	return true
}

func (r *Repository) ApplyDurableFlowReport(
	nodeID int64,
	reportID string,
	flowDeltas []FlowUploadCounterDelta,
	ixDeltas []IXRouteFlowDelta,
	quotaUsage map[int64]int64,
	peerShareUsage map[int64]int64,
	now time.Time,
) (bool, map[int64]*model.UserQuotaView, error) {
	if r == nil || r.db == nil {
		return false, nil, errors.New("repository not initialized")
	}
	reportID = strings.TrimSpace(reportID)
	if nodeID <= 0 || !validFlowReportID(reportID) {
		return false, nil, errors.New("invalid durable flow report identity")
	}
	quotaViews := map[int64]*model.UserQuotaView{}
	applied := false
	err := r.db.Transaction(func(tx *gorm.DB) error {
		receipt := &model.FlowUploadReceipt{NodeID: nodeID, ReportID: reportID, CreatedTime: now.UnixMilli()}
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(receipt)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}
		if err := applyFlowUploadDeltasTx(tx, flowDeltas); err != nil {
			return err
		}
		if err := addIXRouteFlowBatchTx(tx, nodeID, ixDeltas); err != nil {
			return err
		}
		var err error
		quotaViews, err = r.addUserQuotaUsageBatchTx(tx, quotaUsage, now)
		if err != nil {
			return err
		}
		shareIDs := make([]int64, 0, len(peerShareUsage))
		for shareID := range peerShareUsage {
			shareIDs = append(shareIDs, shareID)
		}
		sort.Slice(shareIDs, func(i, j int) bool { return shareIDs[i] < shareIDs[j] })
		for _, shareID := range shareIDs {
			delta := peerShareUsage[shareID]
			if shareID <= 0 || delta <= 0 {
				continue
			}
			if err := tx.Model(&model.PeerShare{}).Where("id = ? AND node_id = ?", shareID, nodeID).UpdateColumns(map[string]interface{}{
				"current_flow": gorm.Expr("current_flow + ?", delta),
				"updated_time": now.UnixMilli(),
			}).Error; err != nil {
				return err
			}
		}
		applied = true
		return nil
	})
	return applied, quotaViews, err
}

func (r *Repository) PurgeFlowUploadReceiptsBefore(cutoff int64) error {
	if r == nil || r.db == nil {
		return errors.New("repository not initialized")
	}
	return r.db.Where("created_time < ?", cutoff).Delete(&model.FlowUploadReceipt{}).Error
}
