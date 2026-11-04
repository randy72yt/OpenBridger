package model

import (
	"errors"
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
	UpstreamGroup      string   `json:"upstream_group" gorm:"type:varchar(64);not null;default:''"`
	Currency           string   `json:"currency" gorm:"type:varchar(16);not null"`
	SourceType         string   `json:"source_type" gorm:"type:varchar(32);not null"`
	SourceURL          string   `json:"source_url" gorm:"type:varchar(512);not null"`
	SourceVersion      string   `json:"source_version" gorm:"type:varchar(128);not null"`
	SuccessRateBPS     int      `json:"success_rate_bps" gorm:"not null"`
	CollectedAt        int64    `json:"collected_at" gorm:"bigint;index"`
	ExpiresAt          int64    `json:"expires_at" gorm:"bigint;index"`
	Enabled            bool     `json:"enabled" gorm:"not null"`
	CreatedAt          int64    `json:"created_at" gorm:"bigint"`
	UpdatedAt          int64    `json:"updated_at" gorm:"bigint"`
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
	ID                     int64  `json:"id" gorm:"primaryKey"`
	PublicModel            string `json:"public_model" gorm:"type:varchar(191);uniqueIndex:idx_price_policy,priority:1"`
	ServiceTier            string `json:"service_tier" gorm:"type:varchar(32);uniqueIndex:idx_price_policy,priority:2"`
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
	CreatedAt              int64  `json:"created_at" gorm:"bigint"`
	UpdatedAt              int64  `json:"updated_at" gorm:"bigint"`
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
		Currency:           offer.Currency, SourceVersion: offer.SourceVersion, CollectedAt: offer.CollectedAt,
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
