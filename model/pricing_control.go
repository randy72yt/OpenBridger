package model

import (
	"errors"

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
	Currency           string   `json:"currency" gorm:"type:varchar(16);not null"`
	SourceVersion      string   `json:"source_version" gorm:"type:varchar(128);not null"`
	CollectedAt        int64    `json:"collected_at" gorm:"bigint;index"`
}

type ModelPricePolicy struct {
	ID                     int64  `json:"id" gorm:"primaryKey"`
	PublicModel            string `json:"public_model" gorm:"type:varchar(191);uniqueIndex:idx_price_policy,priority:1"`
	ServiceTier            string `json:"service_tier" gorm:"type:varchar(32);uniqueIndex:idx_price_policy,priority:2"`
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
	if offer == nil {
		return errors.New("offer is required")
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		return upsertUpstreamModelOffer(tx, offer)
	})
}

func ImportUpstreamModelOffers(offers []UpstreamModelOffer) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		for i := range offers {
			if err := upsertUpstreamModelOffer(tx, &offers[i]); err != nil {
				return err
			}
		}
		return nil
	})
}

func upsertUpstreamModelOffer(tx *gorm.DB, offer *UpstreamModelOffer) error {
	offer.UpdatedAt = common.GetTimestamp()
	if err := tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "channel_id"}, {Name: "upstream_model"}, {Name: "public_model"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"input_cost", "output_cost", "cache_read_cost", "upstream_group_ratio", "currency", "source_type",
			"source_url", "source_version", "success_rate_bps", "collected_at", "expires_at", "enabled", "updated_at",
		}),
	}).Create(offer).Error; err != nil {
		return err
	}
	var current UpstreamModelOffer
	if err := tx.Where("channel_id = ? AND upstream_model = ? AND public_model = ?", offer.ChannelID, offer.UpstreamModel, offer.PublicModel).First(&current).Error; err != nil {
		return err
	}
	offer.ID = current.ID
	return tx.Create(&UpstreamCostSnapshot{
		OfferID: current.ID, ChannelID: offer.ChannelID, PublicModel: offer.PublicModel,
		InputCost: offer.InputCost, OutputCost: offer.OutputCost, CacheReadCost: offer.CacheReadCost,
		UpstreamGroupRatio: offer.UpstreamGroupRatio,
		Currency:           offer.Currency, SourceVersion: offer.SourceVersion, CollectedAt: offer.CollectedAt,
	}).Error
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

func UpsertModelPricePolicy(policy *ModelPricePolicy) error {
	if policy == nil {
		return errors.New("policy is required")
	}
	policy.UpdatedAt = common.GetTimestamp()
	return DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "public_model"}, {Name: "service_tier"}},
		DoUpdates: clause.AssignmentColumns([]string{"primary_channel_id", "backup_channel_id", "target_margin_bps", "minimum_margin_bps", "overhead_bps", "fallback_probability_bps", "auto_publish_change_bps", "price_locked", "enabled", "updated_at"}),
	}).Create(policy).Error
}

func ListModelPricePolicies() ([]ModelPricePolicy, error) {
	var policies []ModelPricePolicy
	err := DB.Order("public_model, service_tier").Find(&policies).Error
	return policies, err
}

func ListModelPriceProposals(status string) ([]ModelPriceProposal, error) {
	var proposals []ModelPriceProposal
	query := DB.Order("id desc").Limit(500)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Find(&proposals).Error
	return proposals, err
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
