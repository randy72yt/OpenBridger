package model

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	PricingProposalPending   = "pending"
	PricingProposalApproved  = "approved"
	PricingProposalPublished = "published"
	PricingProposalRejected  = "rejected"
)

const (
	// PricePolicyScopeGlobal marks the default policy row. It supplies margin
	// and overhead defaults only; it never generates proposals by itself.
	PricePolicyScopeGlobal = "global"
	PricePolicyScopeModel  = "model"
	// GlobalPricePolicyModel is the public_model value of the global row. The
	// (public_model, service_tier) unique index guarantees a single row.
	GlobalPricePolicyModel = "*"
)

// UpstreamModelOffer is the latest normalized procurement cost for one model
// on one channel. Prices use the same per-million billing unit as the public
// billing expression; the original currency and source remain auditable.
type UpstreamModelOffer struct {
	ID                 int64    `json:"id" gorm:"primaryKey"`
	ChannelID          int      `json:"channel_id" gorm:"uniqueIndex:idx_pricing_offer,priority:1;index"`
	UpstreamModel      string   `json:"upstream_model" gorm:"type:varchar(191);uniqueIndex:idx_pricing_offer,priority:2"`
	PublicModel        string   `json:"public_model" gorm:"type:varchar(191);uniqueIndex:idx_pricing_offer,priority:3;index"`
	InputCost          float64  `json:"input_cost" gorm:"type:decimal(20,8);not null"`
	OutputCost         float64  `json:"output_cost" gorm:"type:decimal(20,8);not null"`
	CacheReadCost      float64  `json:"cache_read_cost" gorm:"type:decimal(20,8);not null"`
	UpstreamGroupRatio *float64 `json:"upstream_group_ratio" gorm:"type:decimal(20,8)"`
	// UpstreamGroup records which upstream group produced UpstreamGroupRatio,
	// so a cost can be traced back to the settlement group it came from.
	UpstreamGroup  string `json:"upstream_group" gorm:"type:varchar(64);not null;default:''"`
	Currency       string `json:"currency" gorm:"type:varchar(16);not null"`
	SourceType     string `json:"source_type" gorm:"type:varchar(32);not null"`
	SourceURL      string `json:"source_url" gorm:"type:varchar(512);not null"`
	SourceVersion  string `json:"source_version" gorm:"type:varchar(128);not null"`
	SuccessRateBPS int    `json:"success_rate_bps" gorm:"not null"`
	CollectedAt    int64  `json:"collected_at" gorm:"bigint;index"`
	ExpiresAt      int64  `json:"expires_at" gorm:"bigint;index"`
	Enabled        bool   `json:"enabled" gorm:"not null"`
	CreatedAt      int64  `json:"created_at" gorm:"bigint"`
	UpdatedAt      int64  `json:"updated_at" gorm:"bigint"`
}

type UpstreamCostSnapshot struct {
	ID                 int64    `json:"id" gorm:"primaryKey"`
	OfferID            int64    `json:"offer_id" gorm:"index"`
	ChannelID          int      `json:"channel_id" gorm:"index"`
	PublicModel        string   `json:"public_model" gorm:"type:varchar(191);index"`
	InputCost          float64  `json:"input_cost" gorm:"type:decimal(20,8);not null"`
	OutputCost         float64  `json:"output_cost" gorm:"type:decimal(20,8);not null"`
	CacheReadCost      float64  `json:"cache_read_cost" gorm:"type:decimal(20,8);not null"`
	UpstreamGroupRatio *float64 `json:"upstream_group_ratio" gorm:"type:decimal(20,8)"`
	UpstreamGroup      string   `json:"upstream_group" gorm:"type:varchar(64);not null;default:''"`
	Currency           string   `json:"currency" gorm:"type:varchar(16);not null"`
	SourceVersion      string   `json:"source_version" gorm:"type:varchar(128);not null"`
	CollectedAt        int64    `json:"collected_at" gorm:"bigint;index"`
}

// RecentUpstreamCostSnapshots returns the newest snapshots for one model on
// one channel (newest first), used to verify cost stability across syncs.
func RecentUpstreamCostSnapshots(db *gorm.DB, channelID int, publicModel string, limit int) ([]UpstreamCostSnapshot, error) {
	if db == nil || limit <= 0 {
		return nil, nil
	}
	var snapshots []UpstreamCostSnapshot
	err := db.Where("channel_id = ? AND public_model = ?", channelID, publicModel).
		Order("collected_at desc, id desc").
		Limit(limit).
		Find(&snapshots).Error
	return snapshots, err
}

type ModelPricePolicy struct {
	ID          int64  `json:"id" gorm:"primaryKey"`
	PublicModel string `json:"public_model" gorm:"type:varchar(191);uniqueIndex:idx_price_policy,priority:1"`
	ServiceTier string `json:"service_tier" gorm:"type:varchar(32);uniqueIndex:idx_price_policy,priority:2"`
	// Scope is "global" for the single default row and "model" otherwise (D4).
	Scope                  string `json:"scope" gorm:"type:varchar(16);not null;default:'model'"`
	PrimaryChannelID       int    `json:"primary_channel_id" gorm:"index"`
	BackupChannelID        int    `json:"backup_channel_id" gorm:"index"`
	TargetMarginBPS        int    `json:"target_margin_bps"`
	MinimumMarginBPS       int    `json:"minimum_margin_bps"`
	OverheadBPS            int    `json:"overhead_bps"`
	FallbackProbabilityBPS int    `json:"fallback_probability_bps"`
	AutoPublishChangeBPS   int    `json:"auto_publish_change_bps"`
	PriceLocked            bool   `json:"price_locked"`
	Enabled                bool   `json:"enabled"`
	// Missing-offer tracking: MissingSince is the first successful sync that
	// found no current offer anywhere, MissingChecks counts consecutive
	// verified misses. A valid offer on any channel resets both to zero.
	MissingSince  int64 `json:"missing_since" gorm:"bigint"`
	MissingChecks int   `json:"missing_checks"`
	CreatedAt     int64 `json:"created_at" gorm:"bigint"`
	UpdatedAt     int64 `json:"updated_at" gorm:"bigint"`
}

// IsGlobal reports whether this row is the default policy. Older rows have an
// empty scope and are therefore treated as model-scoped.
func (policy *ModelPricePolicy) IsGlobal() bool {
	return policy != nil && (policy.Scope == PricePolicyScopeGlobal || policy.PublicModel == GlobalPricePolicyModel)
}

// ApplyGlobalDefaults fills unset (zero) knobs from the global policy. A model
// row keeps every value it explicitly set, so overriding one model never leaks
// into the others (D4).
func (policy *ModelPricePolicy) ApplyGlobalDefaults(global *ModelPricePolicy) {
	if policy == nil || global == nil || global.ID == 0 || !global.Enabled || policy.IsGlobal() {
		return
	}
	if policy.TargetMarginBPS <= 0 {
		policy.TargetMarginBPS = global.TargetMarginBPS
	}
	if policy.MinimumMarginBPS <= 0 {
		policy.MinimumMarginBPS = global.MinimumMarginBPS
	}
	if policy.OverheadBPS <= 0 {
		policy.OverheadBPS = global.OverheadBPS
	}
	if policy.FallbackProbabilityBPS <= 0 {
		policy.FallbackProbabilityBPS = global.FallbackProbabilityBPS
	}
	if policy.AutoPublishChangeBPS <= 0 {
		policy.AutoPublishChangeBPS = global.AutoPublishChangeBPS
	}
}

// PricingSyncRun is the durable record of one upstream price collection. The
// console reads the latest row so a refresh no longer loses the sync summary.
type PricingSyncRun struct {
	ID               int64  `json:"id" gorm:"primaryKey"`
	Trigger          string `json:"trigger" gorm:"type:varchar(32);index"`
	Status           string `json:"status" gorm:"type:varchar(32);index"`
	Imported         int    `json:"imported"`
	Kept             int    `json:"kept"`
	Skipped          int    `json:"skipped"`
	CreatedProposals int    `json:"created_proposals"`
	Channels         string `json:"channels" gorm:"type:text"`
	ErrorMessage     string `json:"error,omitempty" gorm:"type:text"`
	StartedAt        int64  `json:"started_at" gorm:"bigint;index"`
	FinishedAt       int64  `json:"finished_at" gorm:"bigint"`
}

type ModelPriceProposal struct {
	ID                  int64   `json:"id" gorm:"primaryKey"`
	PublicModel         string  `json:"public_model" gorm:"type:varchar(191);index"`
	ServiceTier         string  `json:"service_tier" gorm:"type:varchar(32);index"`
	PolicyID            int64   `json:"policy_id" gorm:"index"`
	ExpectedInputCost   float64 `json:"expected_input_cost" gorm:"type:decimal(20,8);not null"`
	ExpectedOutputCost  float64 `json:"expected_output_cost" gorm:"type:decimal(20,8);not null"`
	ProposedInputPrice  float64 `json:"proposed_input_price" gorm:"type:decimal(20,8);not null"`
	ProposedOutputPrice float64 `json:"proposed_output_price" gorm:"type:decimal(20,8);not null"`
	StressInputCost     float64 `json:"stress_input_cost" gorm:"type:decimal(20,8);not null"`
	StressOutputCost    float64 `json:"stress_output_cost" gorm:"type:decimal(20,8);not null"`
	ExpectedMarginBPS   int     `json:"expected_margin_bps"`
	Status              string  `json:"status" gorm:"type:varchar(32);index"`
	Reason              string  `json:"reason" gorm:"type:text"`
	PricingVersion      string  `json:"pricing_version" gorm:"type:varchar(128)"`
	ApprovedBy          int     `json:"approved_by"`
	ApprovedAt          int64   `json:"approved_at" gorm:"bigint"`
	PublishedAt         int64   `json:"published_at" gorm:"bigint"`
	CreatedAt           int64   `json:"created_at" gorm:"bigint;index"`
	UpdatedAt           int64   `json:"updated_at" gorm:"bigint"`
}

func (offer *UpstreamModelOffer) BeforeCreate(_ *gorm.DB) error {
	now := common.GetTimestamp()
	if offer.CreatedAt == 0 {
		offer.CreatedAt = now
	}
	if offer.UpdatedAt == 0 {
		offer.UpdatedAt = now
	}
	return nil
}

func (policy *ModelPricePolicy) BeforeCreate(_ *gorm.DB) error {
	now := common.GetTimestamp()
	if policy.CreatedAt == 0 {
		policy.CreatedAt = now
	}
	if policy.UpdatedAt == 0 {
		policy.UpdatedAt = now
	}
	return nil
}

func (proposal *ModelPriceProposal) BeforeCreate(_ *gorm.DB) error {
	now := common.GetTimestamp()
	if proposal.CreatedAt == 0 {
		proposal.CreatedAt = now
	}
	if proposal.UpdatedAt == 0 {
		proposal.UpdatedAt = now
	}
	return nil
}

func UpsertUpstreamModelOffer(offer *UpstreamModelOffer) error {
	if err := validateUpstreamModelOfferWrite(offer); err != nil {
		return err
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		_, err := upsertUpstreamModelOffer(tx, offer)
		return err
	})
}

// PricingOfferImportStats reports how many offers were written and how many
// were rejected because a higher-priority, still-valid source already owns the
// same (channel, upstream model, public model) row.
type PricingOfferImportStats struct {
	Applied int `json:"applied"`
	Kept    int `json:"kept"`
}

// pricingOfferSourcePriority orders cost sources: automated upstream
// collection always wins over a manual import (decision D3).
func pricingOfferSourcePriority(sourceType string) int {
	switch strings.ToLower(strings.TrimSpace(sourceType)) {
	case "new-api-pricing":
		return 100
	case "manual":
		return 50
	default:
		return 10
	}
}

func ImportUpstreamModelOffers(offers []UpstreamModelOffer) error {
	_, err := ImportUpstreamModelOffersWithStats(offers)
	return err
}

func ImportUpstreamModelOffersWithStats(offers []UpstreamModelOffer) (PricingOfferImportStats, error) {
	stats := PricingOfferImportStats{}
	if len(offers) == 0 {
		return stats, errors.New("offers are required")
	}
	for i := range offers {
		if err := validateUpstreamModelOfferWrite(&offers[i]); err != nil {
			return stats, err
		}
	}
	err := DB.Transaction(func(tx *gorm.DB) error {
		for i := range offers {
			applied, err := upsertUpstreamModelOffer(tx, &offers[i])
			if err != nil {
				return err
			}
			if applied {
				stats.Applied++
			} else {
				stats.Kept++
			}
		}
		return nil
	})
	if err != nil {
		return PricingOfferImportStats{}, err
	}
	return stats, nil
}

func validateUpstreamModelOfferWrite(offer *UpstreamModelOffer) error {
	if offer == nil {
		return errors.New("offer is required")
	}
	validNumber := func(value float64) bool {
		return value >= 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
	}
	if offer.ChannelID <= 0 || strings.TrimSpace(offer.UpstreamModel) == "" || strings.TrimSpace(offer.PublicModel) == "" ||
		strings.ToUpper(strings.TrimSpace(offer.Currency)) != "USD" || strings.TrimSpace(offer.SourceType) == "" {
		return errors.New("invalid upstream pricing offer")
	}
	if !validNumber(offer.InputCost) || !validNumber(offer.OutputCost) || !validNumber(offer.CacheReadCost) ||
		offer.UpstreamGroupRatio == nil || !validNumber(*offer.UpstreamGroupRatio) || *offer.UpstreamGroupRatio <= 0 {
		return errors.New("invalid upstream pricing offer costs")
	}
	if offer.SuccessRateBPS < 0 || offer.SuccessRateBPS > 10000 {
		return errors.New("invalid upstream pricing offer success rate")
	}
	return nil
}

// upsertUpstreamModelOffer writes one offer and appends a cost snapshot.
// It returns false without writing when an unexpired row from a
// higher-priority source already owns the same key (D3: automated collection
// wins over manual import).
func upsertUpstreamModelOffer(tx *gorm.DB, offer *UpstreamModelOffer) (bool, error) {
	now := common.GetTimestamp()
	var existing UpstreamModelOffer
	err := tx.Where("channel_id = ? AND upstream_model = ? AND public_model = ?", offer.ChannelID, offer.UpstreamModel, offer.PublicModel).
		First(&existing).Error
	if err == nil {
		if existing.Enabled && existing.ExpiresAt > now &&
			pricingOfferSourcePriority(existing.SourceType) > pricingOfferSourcePriority(offer.SourceType) {
			return false, nil
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	offer.UpdatedAt = now
	if err := tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "channel_id"}, {Name: "upstream_model"}, {Name: "public_model"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"input_cost", "output_cost", "cache_read_cost", "upstream_group_ratio", "upstream_group", "currency", "source_type",
			"source_url", "source_version", "success_rate_bps", "collected_at", "expires_at", "enabled", "updated_at",
		}),
	}).Create(offer).Error; err != nil {
		return false, err
	}
	var current UpstreamModelOffer
	if err := tx.Where("channel_id = ? AND upstream_model = ? AND public_model = ?", offer.ChannelID, offer.UpstreamModel, offer.PublicModel).First(&current).Error; err != nil {
		return false, err
	}
	offer.ID = current.ID
	if err := tx.Create(&UpstreamCostSnapshot{
		OfferID: current.ID, ChannelID: offer.ChannelID, PublicModel: offer.PublicModel,
		InputCost: offer.InputCost, OutputCost: offer.OutputCost, CacheReadCost: offer.CacheReadCost,
		UpstreamGroupRatio: offer.UpstreamGroupRatio, UpstreamGroup: offer.UpstreamGroup,
		Currency: offer.Currency, SourceVersion: offer.SourceVersion, CollectedAt: offer.CollectedAt,
	}).Error; err != nil {
		return false, err
	}
	return true, nil
}

func ListUpstreamModelOffers(publicModel string) ([]UpstreamModelOffer, error) {
	var offers []UpstreamModelOffer
	query := DB.Order("public_model, channel_id")
	if publicModel != "" {
		query = query.Where("public_model = ?", publicModel)
	}
	err := query.Find(&offers).Error
	return offers, err
}

// UpstreamModelOfferQuery filters the cost inventory. Empty IncludeExpired
// keeps only offers that are enabled and not past their expiry, which is what
// the console shows by default.
type UpstreamModelOfferQuery struct {
	PublicModel    string
	Keyword        string
	SourceType     string
	ChannelID      int
	IncludeExpired bool
	Limit          int
	Offset         int
}

type UpstreamModelOfferPage struct {
	Items  []UpstreamModelOffer `json:"items"`
	Total  int64                `json:"total"`
	Limit  int                  `json:"limit"`
	Offset int                  `json:"offset"`
}

const (
	DefaultPricingOfferPageSize = 50
	MaxPricingOfferPageSize     = 200
)

func normalizePricingPageSize(limit int) int {
	if limit <= 0 {
		return DefaultPricingOfferPageSize
	}
	if limit > MaxPricingOfferPageSize {
		return MaxPricingOfferPageSize
	}
	return limit
}

func ListUpstreamModelOffersPage(query UpstreamModelOfferQuery) (UpstreamModelOfferPage, error) {
	page := UpstreamModelOfferPage{Limit: normalizePricingPageSize(query.Limit)}
	if query.Offset < 0 {
		query.Offset = 0
	}
	page.Offset = query.Offset
	db := DB.Model(&UpstreamModelOffer{})
	if query.PublicModel != "" {
		db = db.Where("public_model = ?", query.PublicModel)
	}
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("public_model LIKE ? OR upstream_model LIKE ?", like, like)
	}
	if query.SourceType != "" {
		db = db.Where("source_type = ?", query.SourceType)
	}
	if query.ChannelID > 0 {
		db = db.Where("channel_id = ?", query.ChannelID)
	}
	if !query.IncludeExpired {
		db = db.Where("enabled = ? AND expires_at > ?", true, common.GetTimestamp())
	}
	if err := db.Count(&page.Total).Error; err != nil {
		return page, err
	}
	if err := db.Order("public_model, channel_id").
		Limit(page.Limit).Offset(page.Offset).
		Find(&page.Items).Error; err != nil {
		return page, err
	}
	return page, nil
}

func UpsertModelPricePolicy(policy *ModelPricePolicy) error {
	if policy == nil {
		return errors.New("policy is required")
	}
	if policy.Scope == "" {
		if policy.IsGlobal() {
			policy.Scope = PricePolicyScopeGlobal
		} else {
			policy.Scope = PricePolicyScopeModel
		}
	}
	policy.UpdatedAt = common.GetTimestamp()
	return DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "public_model"}, {Name: "service_tier"}},
		DoUpdates: clause.AssignmentColumns([]string{"scope", "primary_channel_id", "backup_channel_id", "target_margin_bps", "minimum_margin_bps", "overhead_bps", "fallback_probability_bps", "auto_publish_change_bps", "price_locked", "enabled", "updated_at"}),
	}).Create(policy).Error
}

// GlobalPricePolicy returns the enabled default policy, or nil when none exists.
func GlobalPricePolicy() (*ModelPricePolicy, error) {
	var policies []ModelPricePolicy
	if err := DB.Where("public_model = ? OR scope = ?", GlobalPricePolicyModel, PricePolicyScopeGlobal).Find(&policies).Error; err != nil {
		return nil, err
	}
	for i := range policies {
		if policies[i].Enabled {
			return &policies[i], nil
		}
	}
	return nil, nil
}

func ListModelPricePolicies() ([]ModelPricePolicy, error) {
	var policies []ModelPricePolicy
	err := DB.Order("public_model, service_tier").Find(&policies).Error
	return policies, err
}

// ListModelPriceProposals is the unbounded internal listing. Console reads go
// through ListModelPriceProposalsPage so a growing table cannot truncate
// silently.
func ListModelPriceProposals(status string) ([]ModelPriceProposal, error) {
	var proposals []ModelPriceProposal
	query := DB.Order("id desc")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Find(&proposals).Error
	return proposals, err
}

// ListModelPriceProposalsPage returns one page of proposals plus the total
// count for the same filter. A non-positive limit falls back to the default
// page size, so callers no longer rely on a silent hard cap.
func ListModelPriceProposalsPage(status string, limit, offset int) ([]ModelPriceProposal, int64, error) {
	pageSize := normalizePricingPageSize(limit)
	if offset < 0 {
		offset = 0
	}
	countQuery := DB.Model(&ModelPriceProposal{})
	if status != "" {
		countQuery = countQuery.Where("status = ?", status)
	}
	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var proposals []ModelPriceProposal
	query := DB.Order("id desc").Limit(pageSize).Offset(offset)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Find(&proposals).Error; err != nil {
		return nil, 0, err
	}
	return proposals, total, nil
}

func LoadModelPriceProposalForUpdate(tx *gorm.DB, id int64) (ModelPriceProposal, error) {
	var proposal ModelPriceProposal
	err := lockForUpdate(tx).Where("id = ?", id).First(&proposal).Error
	return proposal, err
}

func LoadModelPricePolicyForUpdate(tx *gorm.DB, id int64) (ModelPricePolicy, error) {
	var policy ModelPricePolicy
	err := lockForUpdate(tx).Where("id = ?", id).First(&policy).Error
	return policy, err
}

// RejectOtherApprovedPriceProposals retires unpublished approvals for one
// public model so an older approval cannot be published over a newer price.
func RejectOtherApprovedPriceProposals(tx *gorm.DB, publicModel string, exceptID int64) error {
	if tx == nil || strings.TrimSpace(publicModel) == "" {
		return errors.New("proposal is required")
	}
	return tx.Model(&ModelPriceProposal{}).
		Where("public_model = ? AND status = ? AND id <> ?", publicModel, PricingProposalApproved, exceptID).
		Updates(map[string]any{"status": PricingProposalRejected, "updated_at": common.GetTimestamp()}).Error
}

// LatestPendingModelPriceProposal returns the newest pending proposal for a
// model+tier. Recalculation uses it to skip inserting a duplicate when neither
// the expected costs nor the proposed prices moved.
func LatestPendingModelPriceProposal(publicModel, serviceTier string) (*ModelPriceProposal, bool) {
	var proposal ModelPriceProposal
	err := DB.Where("public_model = ? AND service_tier = ? AND status = ?", publicModel, serviceTier, PricingProposalPending).
		Order("id desc").First(&proposal).Error
	if err != nil {
		return nil, false
	}
	return &proposal, true
}

// PruneModelPriceProposals deletes proposals with the given statuses whose
// updated_at precedes the cutoff. Rejected rows are pure churn history, so the
// retention job bounds them by time; approved/published rows are the pricing
// audit trail and are pruned only by an explicit caller.
func PruneModelPriceProposals(statuses []string, cutoff int64) (int64, error) {
	if len(statuses) == 0 || cutoff <= 0 {
		return 0, errors.New("proposal prune requires statuses and a cutoff")
	}
	result := DB.Where("status IN ? AND updated_at < ?", statuses, cutoff).Delete(&ModelPriceProposal{})
	return result.RowsAffected, result.Error
}

// DisableModelPricePolicy disables the model-scoped pricing policy so the
// model stops being repriced. It reports whether a row was actually disabled.
func DisableModelPricePolicy(publicModel, serviceTier string) (bool, error) {
	if publicModel == GlobalPricePolicyModel || serviceTier == PricePolicyScopeGlobal {
		return false, errors.New("the global pricing policy cannot be disabled here")
	}
	result := DB.Model(&ModelPricePolicy{}).
		Where("public_model = ? AND service_tier = ? AND enabled = ?", publicModel, serviceTier, true).
		Updates(map[string]any{"enabled": false, "missing_since": 0, "missing_checks": 0, "updated_at": common.GetTimestamp()})
	return result.RowsAffected > 0, result.Error
}

// ChannelsWithCurrentOffer lists channel IDs that currently hold an unexpired
// offer for the model. Missing-offer verification cross-checks every channel
// before concluding that the model is gone upstream.
func ChannelsWithCurrentOffer(publicModel string, now int64) ([]int, error) {
	channelIDs := make([]int, 0)
	err := DB.Model(&UpstreamModelOffer{}).
		Where("public_model = ? AND enabled = ? AND expires_at > ?", publicModel, true, now).
		Distinct("channel_id").
		Pluck("channel_id", &channelIDs).Error
	return channelIDs, err
}

func ReplacePendingModelPriceProposal(proposal *ModelPriceProposal) error {
	if proposal == nil {
		return errors.New("proposal is required")
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		now := common.GetTimestamp()
		if err := tx.Model(&ModelPriceProposal{}).
			Where("public_model = ? AND service_tier = ? AND status = ?", proposal.PublicModel, proposal.ServiceTier, PricingProposalPending).
			Updates(map[string]any{"status": PricingProposalRejected, "updated_at": now}).Error; err != nil {
			return err
		}
		return tx.Create(proposal).Error
	})
}

// PruneUpstreamCostSnapshots deletes snapshots collected before the cutoff.
// Snapshots stay append-only because auto-publish verification reads the most
// recent samples, so growth is bounded by time instead of de-duplication.
func PruneUpstreamCostSnapshots(cutoff int64) (int64, error) {
	if cutoff <= 0 {
		return 0, errors.New("snapshot cutoff is required")
	}
	result := DB.Where("collected_at < ?", cutoff).Delete(&UpstreamCostSnapshot{})
	return result.RowsAffected, result.Error
}

// PruneExpiredUpstreamModelOffers deletes offers that expired before the
// cutoff, so retired channels stop leaving dead rows in the inventory.
func PruneExpiredUpstreamModelOffers(cutoff int64) (int64, error) {
	if cutoff <= 0 {
		return 0, errors.New("offer cutoff is required")
	}
	result := DB.Where("expires_at < ?", cutoff).Delete(&UpstreamModelOffer{})
	return result.RowsAffected, result.Error
}

func RecordPricingSyncRun(run *PricingSyncRun) error {
	if run == nil {
		return errors.New("sync run is required")
	}
	if run.StartedAt <= 0 {
		run.StartedAt = common.GetTimestamp()
	}
	if run.FinishedAt <= 0 {
		run.FinishedAt = common.GetTimestamp()
	}
	return DB.Create(run).Error
}

func LatestPricingSyncRun() (*PricingSyncRun, error) {
	var run PricingSyncRun
	err := DB.Order("id desc").First(&run).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &run, nil
}

func ListPricingSyncRuns(limit int) ([]PricingSyncRun, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	var runs []PricingSyncRun
	err := DB.Order("id desc").Limit(limit).Find(&runs).Error
	return runs, err
}

// PrunePricingSyncRuns keeps the audit trail bounded.
func PrunePricingSyncRuns(cutoff int64) (int64, error) {
	if cutoff <= 0 {
		return 0, errors.New("sync run cutoff is required")
	}
	result := DB.Where("started_at < ?", cutoff).Delete(&PricingSyncRun{})
	return result.RowsAffected, result.Error
}

const (
	// PricingCalibrationSourceUpstreamLog marks a factor derived from the
	// upstream provider's own consumption log. It is the only source that
	// reflects what the provider actually charged us.
	PricingCalibrationSourceUpstreamLog = "upstream_log"
	// PricingCalibrationSourceManual marks a factor a human pinned on purpose.
	// A manual row is never overwritten by the automatic calibration.
	PricingCalibrationSourceManual = "manual"
	// PricingCalibrationSourceSeed marks the one-shot factor imported from a
	// measured sample before the automatic loop had any data.
	PricingCalibrationSourceSeed = "seed"
)

// UpstreamCostCalibration records the end-to-end factor between what an upstream
// provider actually billed us and what our own price table predicts for the very
// same tokens:
//
//	factor = (quota / quota_per_unit) / (tokens × our published unit price)
//
// Upstream pricing is not reconstructable from `/api/pricing` alone. Real
// traffic showed the same model settling through a group whose ratio is absent
// from the public table, cache tokens discounted by a per-request ratio, a
// night-time factor, tiered expressions and a currency divisor. Rather than
// modelling each of those, the factor absorbs all of them at once: multiplying
// our base cost by the factor reproduces the provider's invoice exactly.
//
// The factor is a property of (channel, model) and is re-derived from real
// traffic, so a provider changing its markup is picked up without a code change.
type UpstreamCostCalibration struct {
	ID          int64  `json:"id" gorm:"primaryKey"`
	ChannelID   int    `json:"channel_id" gorm:"uniqueIndex:idx_upstream_cost_calibration,priority:1;index"`
	PublicModel string `json:"public_model" gorm:"type:varchar(191);uniqueIndex:idx_upstream_cost_calibration,priority:2;index"`
	// UpstreamModel is the name the provider used, kept for audit because the
	// same public model can be served under a different upstream name.
	UpstreamModel string `json:"upstream_model" gorm:"type:varchar(191);not null;default:''"`
	// EffectiveRatio replaces UpstreamModelOffer.UpstreamGroupRatio when pricing.
	EffectiveRatio float64 `json:"effective_ratio" gorm:"type:decimal(20,10);not null"`
	// PolicyRatio is what the group-ratio policy would have produced, kept so
	// the UI can show how far the policy is off.
	PolicyRatio float64 `json:"policy_ratio" gorm:"type:decimal(20,10);not null;default:0"`
	// ObservedGroup is the settlement group seen in the upstream log. It is the
	// only way to learn which group a model actually settles through.
	ObservedGroup string  `json:"observed_group" gorm:"type:varchar(64);not null;default:''"`
	Source        string  `json:"source" gorm:"type:varchar(32);not null"`
	Samples       int     `json:"samples" gorm:"not null;default:0"`
	BilledUSD     float64 `json:"billed_usd" gorm:"type:decimal(20,8);not null;default:0"`
	TheoryUSD     float64 `json:"theory_usd" gorm:"type:decimal(20,8);not null;default:0"`
	// DispersionPct is the median absolute deviation of the per-request factor,
	// as a percentage of the median. A high value means traffic is mixing
	// billing modes (for example standard and long-context tiers), so a single
	// factor cannot describe them all.
	DispersionPct float64 `json:"dispersion_pct" gorm:"type:decimal(10,4);not null;default:0"`
	Confidence    float64 `json:"confidence" gorm:"type:decimal(6,4);not null;default:0"`
	WindowStart   int64   `json:"window_start" gorm:"bigint;not null;default:0"`
	WindowEnd     int64   `json:"window_end" gorm:"bigint;not null;default:0;index"`
	Applied       bool    `json:"applied" gorm:"not null;default:false"`
	Reason        string  `json:"reason" gorm:"type:varchar(255);not null;default:''"`
	CreatedAt     int64   `json:"created_at" gorm:"bigint"`
	UpdatedAt     int64   `json:"updated_at" gorm:"bigint"`
}

func (UpstreamCostCalibration) TableName() string {
	return "upstream_cost_calibrations"
}

// UpstreamCostCalibrationKey is the lookup key used by the pricing path.
func UpstreamCostCalibrationKey(channelID int, publicModel string) string {
	return fmt.Sprintf("%d|%s", channelID, strings.ToLower(strings.TrimSpace(publicModel)))
}

func (cal *UpstreamCostCalibration) IsUsable(now, staleSeconds int64) bool {
	if cal == nil || !cal.Applied {
		return false
	}
	if cal.EffectiveRatio <= 0 || math.IsNaN(cal.EffectiveRatio) || math.IsInf(cal.EffectiveRatio, 0) {
		return false
	}
	if staleSeconds > 0 && now-cal.WindowEnd > staleSeconds {
		return false
	}
	return true
}

func UpsertUpstreamCostCalibration(cal *UpstreamCostCalibration) error {
	if cal == nil {
		return errors.New("calibration is required")
	}
	cal.PublicModel = strings.TrimSpace(cal.PublicModel)
	if cal.ChannelID <= 0 || cal.PublicModel == "" {
		return errors.New("calibration channel and model are required")
	}
	now := common.GetTimestamp()
	if cal.CreatedAt <= 0 {
		cal.CreatedAt = now
	}
	cal.UpdatedAt = now
	return DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "channel_id"}, {Name: "public_model"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"upstream_model", "effective_ratio", "policy_ratio", "observed_group",
			"source", "samples", "billed_usd", "theory_usd", "dispersion_pct",
			"confidence", "window_start", "window_end", "applied", "reason", "updated_at",
		}),
	}).Create(cal).Error
}

func ListUpstreamCostCalibrations() ([]UpstreamCostCalibration, error) {
	var rows []UpstreamCostCalibration
	err := DB.Order("public_model asc").Find(&rows).Error
	return rows, err
}

// UpstreamCostCalibrationMap loads every calibration in one query. The pricing
// path calls this once per recalculation instead of per model.
func UpstreamCostCalibrationMap() (map[string]UpstreamCostCalibration, error) {
	rows, err := ListUpstreamCostCalibrations()
	if err != nil {
		return nil, err
	}
	out := make(map[string]UpstreamCostCalibration, len(rows))
	for _, row := range rows {
		out[UpstreamCostCalibrationKey(row.ChannelID, row.PublicModel)] = row
	}
	return out, nil
}

func FindUpstreamCostCalibration(channelID int, publicModel string) (*UpstreamCostCalibration, error) {
	var cal UpstreamCostCalibration
	err := DB.Where("channel_id = ? AND public_model = ?", channelID, strings.TrimSpace(publicModel)).
		First(&cal).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &cal, nil
}

// PruneUpstreamCostCalibrations drops rows for models that stopped being
// calibrated, so a decommissioned model cannot linger as a stale override.
func PruneUpstreamCostCalibrations(cutoff int64) (int64, error) {
	if cutoff <= 0 {
		return 0, errors.New("calibration cutoff is required")
	}
	result := DB.Where("updated_at < ? AND source <> ?", cutoff, PricingCalibrationSourceManual).
		Delete(&UpstreamCostCalibration{})
	return result.RowsAffected, result.Error
}

// ListCurrentUpstreamModelOffers returns the live offers of one channel, newest
// first per public model. Calibration needs the exact unit prices the pricing
// path uses, so it must read offers and not the upstream price list.
func ListCurrentUpstreamModelOffers(channelID int, now int64) ([]UpstreamModelOffer, error) {
	query := DB.Where("enabled = ?", true)
	if channelID > 0 {
		query = query.Where("channel_id = ?", channelID)
	}
	if now > 0 {
		query = query.Where("expires_at > ?", now)
	}
	var rows []UpstreamModelOffer
	err := query.Order("public_model asc, collected_at desc").Find(&rows).Error
	return rows, err
}

// ListChannelsWithPricingOffers lists the channels the pricing control plane
// actually manages, so calibration never walks unrelated channels.
func ListChannelsWithPricingOffers() ([]int, error) {
	var ids []int
	err := DB.Model(&UpstreamModelOffer{}).Distinct("channel_id").Pluck("channel_id", &ids).Error
	return ids, err
}

// DeleteUpstreamCostCalibration removes one calibration row, optionally limited
// to a single source so an automatic row cannot be deleted by mistake.
func DeleteUpstreamCostCalibration(channelID int, publicModel, source string) error {
	query := DB.Where("channel_id = ? AND public_model = ?", channelID, strings.TrimSpace(publicModel))
	if source != "" {
		query = query.Where("source = ?", source)
	}
	return query.Delete(&UpstreamCostCalibration{}).Error
}

// ListUpstreamCostCalibrationsPage pages the calibration table for the console.
func ListUpstreamCostCalibrationsPage(keyword string, channelID, limit, offset int) ([]UpstreamCostCalibration, int64, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	query := DB.Model(&UpstreamCostCalibration{})
	if trimmed := strings.TrimSpace(keyword); trimmed != "" {
		like := "%" + trimmed + "%"
		query = query.Where("public_model LIKE ? OR upstream_model LIKE ? OR observed_group LIKE ?", like, like, like)
	}
	if channelID > 0 {
		query = query.Where("channel_id = ?", channelID)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []UpstreamCostCalibration
	err := query.Order("applied desc, public_model asc").Limit(limit).Offset(offset).Find(&rows).Error
	return rows, total, err
}
