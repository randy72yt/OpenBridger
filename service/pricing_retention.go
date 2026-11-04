package service

import (
	"context"
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

// pricingRetentionHandler bounds the growth of the pricing tables.
//
//	upstream_cost_snapshots : append-only, kept for PRICING_SNAPSHOT_RETENTION_DAYS (30)
//	upstream_model_offers   : rows expired longer than PRICING_EXPIRED_OFFER_RETENTION_DAYS (7) are deleted
//
// Deleting snapshots only removes history; live offers, proposals and published
// prices are untouched.
type pricingRetentionHandler struct{}

func (pricingRetentionHandler) Type() string { return model.SystemTaskTypePricingRetention }
func (pricingRetentionHandler) Enabled() bool {
	return PricingControlEnabled()
}
func (pricingRetentionHandler) Interval() time.Duration {
	hours := common.GetEnvOrDefault("PRICING_RETENTION_INTERVAL_HOURS", 24)
	if hours < 1 {
		hours = 1
	}
	return time.Duration(hours) * time.Hour
}
func (pricingRetentionHandler) NewPayload() any { return map[string]any{"scheduled": true} }

func (pricingRetentionHandler) Run(ctx context.Context, task *model.SystemTask, runnerID string) {
	snapshotDays := common.GetEnvOrDefault("PRICING_SNAPSHOT_RETENTION_DAYS", 30)
	if snapshotDays < 1 {
		snapshotDays = 1
	}
	offerDays := common.GetEnvOrDefault("PRICING_EXPIRED_OFFER_RETENTION_DAYS", 7)
	if offerDays < 1 {
		offerDays = 1
	}
	now := common.GetTimestamp()
	snapshots, err := model.PruneUpstreamCostSnapshots(now - int64(snapshotDays)*86400)
	if err != nil {
		failSystemTask(task, runnerID, err)
		return
	}
	offers, err := model.PruneExpiredUpstreamModelOffers(now - int64(offerDays)*86400)
	if err != nil {
		failSystemTask(task, runnerID, err)
		return
	}
	result := map[string]any{
		"deleted_snapshots":  snapshots,
		"deleted_offers":     offers,
		"snapshot_keep_days": snapshotDays,
		"offer_keep_days":    offerDays,
	}
	if err := model.FinishSystemTask(task.TaskID, runnerID, model.SystemTaskStatusSucceeded, result, ""); err != nil {
		logSystemTaskLockError(ctx, task, err)
	}
}

func init() { RegisterSystemTaskHandler(pricingRetentionHandler{}) }

// DescribePricingRetention is used by the console to show what the cleanup job
// will do before an operator trusts it.
func DescribePricingRetention() string {
	return fmt.Sprintf("snapshots kept %d days, expired offers kept %d days",
		common.GetEnvOrDefault("PRICING_SNAPSHOT_RETENTION_DAYS", 30),
		common.GetEnvOrDefault("PRICING_EXPIRED_OFFER_RETENTION_DAYS", 7))
}
