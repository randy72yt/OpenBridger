package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/pkg/billingexpr"
	"gorm.io/gorm"
)

type PricingProposalReview struct {
	model.ModelPriceProposal
	PrimaryChannelName   string   `json:"primary_channel_name"`
	BackupChannelName    string   `json:"backup_channel_name"`
	PrimaryCollectedAt   int64    `json:"primary_collected_at"`
	PrimaryExpiresAt     int64    `json:"primary_expires_at"`
	BackupCollectedAt    int64    `json:"backup_collected_at"`
	BackupExpiresAt      int64    `json:"backup_expires_at"`
	SourceType           string   `json:"source_type"`
	SourceVersion        string   `json:"source_version"`
	CurrentInputPrice    *float64 `json:"current_input_price"`
	CurrentOutputPrice   *float64 `json:"current_output_price"`
	InputPriceChangeBPS  *int     `json:"input_price_change_bps"`
	OutputPriceChangeBPS *int     `json:"output_price_change_bps"`
	WorstMarginBPS       int      `json:"worst_margin_bps"`
	PricingChanged       bool     `json:"pricing_changed"`
}

func ListPricingProposalReviews(status string) ([]PricingProposalReview, error) {
	proposals, err := model.ListModelPriceProposals(status)
	if err != nil {
		return nil, err
	}
	policies, err := model.ListModelPricePolicies()
	if err != nil {
		return nil, err
	}
	offers, err := model.ListUpstreamModelOffers("")
	if err != nil {
		return nil, err
	}
	policyByID := make(map[int64]model.ModelPricePolicy, len(policies))
	channelIDs := make([]int, 0, len(policies)*2)
	for i := range policies {
		policyByID[policies[i].ID] = policies[i]
		channelIDs = append(channelIDs, policies[i].PrimaryChannelID)
		if policies[i].BackupChannelID > 0 {
			channelIDs = append(channelIDs, policies[i].BackupChannelID)
		}
	}
	channelNames := make(map[int]string)
	if len(channelIDs) > 0 {
		var channels []model.Channel
		if err := model.DB.Select("id", "name").Where("id IN ?", channelIDs).Find(&channels).Error; err != nil {
			return nil, err
		}
		for i := range channels {
			channelNames[channels[i].Id] = channels[i].Name
		}
	}
	offerByModelChannel := make(map[string]model.UpstreamModelOffer, len(offers))
	for i := range offers {
		key := fmt.Sprintf("%s\x00%d", offers[i].PublicModel, offers[i].ChannelID)
		offerByModelChannel[key] = offers[i]
	}
	modelNames := make([]string, 0, len(proposals))
	seenModels := make(map[string]bool, len(proposals))
	for i := range proposals {
		if !seenModels[proposals[i].PublicModel] {
			modelNames = append(modelNames, proposals[i].PublicModel)
			seenModels[proposals[i].PublicModel] = true
		}
	}
	pricing, err := model.GetModelPricingSnapshot(modelNames)
	if err != nil {
		return nil, err
	}
	pricingByModel := make(map[string]model.ModelPricingEntry, len(pricing.Entries))
	for i := range pricing.Entries {
		pricingByModel[pricing.Entries[i].ModelName] = pricing.Entries[i]
	}

	reviews := make([]PricingProposalReview, 0, len(proposals))
	for i := range proposals {
		proposal := proposals[i]
		review := PricingProposalReview{ModelPriceProposal: proposal}
		policy, ok := policyByID[proposal.PolicyID]
		if ok {
			review.PrimaryChannelName = channelNames[policy.PrimaryChannelID]
			review.BackupChannelName = channelNames[policy.BackupChannelID]
			primary := offerByModelChannel[fmt.Sprintf("%s\x00%d", proposal.PublicModel, policy.PrimaryChannelID)]
			review.PrimaryCollectedAt = primary.CollectedAt
			review.PrimaryExpiresAt = primary.ExpiresAt
			review.SourceType = primary.SourceType
			review.SourceVersion = primary.SourceVersion
			if policy.BackupChannelID > 0 {
				backup := offerByModelChannel[fmt.Sprintf("%s\x00%d", proposal.PublicModel, policy.BackupChannelID)]
				review.BackupCollectedAt = backup.CollectedAt
				review.BackupExpiresAt = backup.ExpiresAt
			}
		}
		review.WorstMarginBPS = minimumPriceMarginBPS(proposal.ProposedInputPrice, proposal.ProposedOutputPrice, proposal.StressInputCost, proposal.StressOutputCost)
		if entry, exists := pricingByModel[proposal.PublicModel]; exists {
			review.PricingChanged = entry.Version != proposal.PricingVersion
			expression, expressionOK := entry.Effective["billing_setting.billing_expr"].(string)
			if expressionOK {
				inputPrice, _, inputErr := billingexpr.RunExpr(expression, billingexpr.TokenParams{P: 1, Len: 1})
				outputPrice, _, outputErr := billingexpr.RunExpr(expression, billingexpr.TokenParams{C: 1})
				if inputErr == nil && outputErr == nil && validPriceNumber(inputPrice) && validPriceNumber(outputPrice) {
					inputPrice = roundPricingValue(inputPrice)
					outputPrice = roundPricingValue(outputPrice)
					review.CurrentInputPrice = &inputPrice
					review.CurrentOutputPrice = &outputPrice
					review.InputPriceChangeBPS = priceChangeBPS(inputPrice, proposal.ProposedInputPrice)
					review.OutputPriceChangeBPS = priceChangeBPS(outputPrice, proposal.ProposedOutputPrice)
				}
			}
		}
		reviews = append(reviews, review)
	}
	return reviews, nil
}

func minimumPriceMarginBPS(inputPrice, outputPrice, inputCost, outputCost float64) int {
	if inputPrice <= 0 || outputPrice <= 0 {
		return 0
	}
	inputMargin := int(math.Round((1 - inputCost/inputPrice) * 10000))
	outputMargin := int(math.Round((1 - outputCost/outputPrice) * 10000))
	return min(inputMargin, outputMargin)
}

func priceChangeBPS(current, proposed float64) *int {
	if current <= 0 {
		return nil
	}
	change := int(math.Round((proposed/current - 1) * 10000))
	return &change
}

var (
	ErrPricingOfferInvalid  = errors.New("invalid upstream pricing offer")
	ErrPricingPolicyInvalid = errors.New("invalid pricing policy")
	ErrPricingProposalState = errors.New("invalid pricing proposal state")
	ErrPricingOfferMissing  = errors.New("required upstream pricing offer is missing or expired")
	ErrPricingCostInvalid   = errors.New("upstream pricing offer has no positive effective cost")
	ErrPricingProposalStale = errors.New("pricing proposal is stale; refresh offers and recalculate")
)

func ValidateAndImportPricingOffers(offers []model.UpstreamModelOffer) error {
	if len(offers) == 0 {
		return ErrPricingOfferInvalid
	}
	now := common.GetTimestamp()
	for i := range offers {
		offer := &offers[i]
		offer.UpstreamModel = strings.TrimSpace(offer.UpstreamModel)
		offer.PublicModel = strings.TrimSpace(offer.PublicModel)
		offer.Currency = strings.ToUpper(strings.TrimSpace(offer.Currency))
		offer.SourceType = strings.TrimSpace(offer.SourceType)
		if offer.ChannelID <= 0 || offer.UpstreamModel == "" || offer.PublicModel == "" || offer.Currency != "USD" || offer.SourceType == "" {
			return fmt.Errorf("%w at item %d", ErrPricingOfferInvalid, i)
		}
		if !validPriceNumber(offer.InputCost) || !validPriceNumber(offer.OutputCost) || !validPriceNumber(offer.CacheReadCost) {
			return fmt.Errorf("%w at item %d", ErrPricingOfferInvalid, i)
		}
		if offer.UpstreamGroupRatio == nil || !validPriceNumber(*offer.UpstreamGroupRatio) || *offer.UpstreamGroupRatio <= 0 {
			return fmt.Errorf("%w at item %d", ErrPricingOfferInvalid, i)
		}
		if offer.SuccessRateBPS < 0 || offer.SuccessRateBPS > 10000 {
			return fmt.Errorf("%w at item %d", ErrPricingOfferInvalid, i)
		}
		if offer.CollectedAt <= 0 {
			offer.CollectedAt = now
		}
		if offer.ExpiresAt <= offer.CollectedAt {
			offer.ExpiresAt = offer.CollectedAt + 24*60*60
		}
	}
	return model.ImportUpstreamModelOffers(offers)
}

func ValidateAndUpsertPricePolicy(policy *model.ModelPricePolicy) error {
	if policy == nil {
		return ErrPricingPolicyInvalid
	}
	policy.PublicModel = strings.TrimSpace(policy.PublicModel)
	policy.ServiceTier = strings.TrimSpace(policy.ServiceTier)
	if policy.ServiceTier == "" {
		policy.ServiceTier = "default"
	}
	if policy.PublicModel == "" || policy.PrimaryChannelID <= 0 || policy.BackupChannelID < 0 {
		return ErrPricingPolicyInvalid
	}
	if policy.TargetMarginBPS <= 0 || policy.TargetMarginBPS >= 10000 || policy.MinimumMarginBPS < 0 || policy.MinimumMarginBPS >= policy.TargetMarginBPS {
		return ErrPricingPolicyInvalid
	}
	if policy.OverheadBPS < 0 || policy.OverheadBPS > 5000 || policy.FallbackProbabilityBPS < 0 || policy.FallbackProbabilityBPS > 10000 || policy.AutoPublishChangeBPS < 0 || policy.AutoPublishChangeBPS > 5000 {
		return ErrPricingPolicyInvalid
	}
	return model.UpsertModelPricePolicy(policy)
}

func RecalculatePricingProposals(ctx context.Context) (int, error) {
	created, _, err := recalculatePricingProposals(ctx)
	return created, err
}

func recalculatePricingProposals(ctx context.Context) (int, []int64, error) {
	policies, err := model.ListModelPricePolicies()
	if err != nil {
		return 0, nil, err
	}
	created := 0
	ids := make([]int64, 0)
	for i := range policies {
		if ctx != nil && ctx.Err() != nil {
			return created, ids, ctx.Err()
		}
		policy := policies[i]
		if !policy.Enabled || policy.PriceLocked {
			continue
		}
		proposal, err := buildPriceProposal(&policy)
		if err != nil {
			if errors.Is(err, ErrPricingOfferMissing) || errors.Is(err, ErrPricingCostInvalid) {
				continue
			}
			return created, ids, err
		}
		if err := model.ReplacePendingModelPriceProposal(proposal); err != nil {
			return created, ids, err
		}
		if proposal.ID <= 0 {
			return created, ids, errors.New("pricing proposal id was not assigned")
		}
		ids = append(ids, proposal.ID)
		created++
	}
	return created, ids, nil
}

func buildPriceProposal(policy *model.ModelPricePolicy) (*model.ModelPriceProposal, error) {
	now := common.GetTimestamp()
	primary, err := currentOffer(policy.PublicModel, policy.PrimaryChannelID, now)
	if err != nil {
		return nil, err
	}
	var backup *model.UpstreamModelOffer
	if policy.BackupChannelID > 0 {
		backup, err = currentOffer(policy.PublicModel, policy.BackupChannelID, now)
		if err != nil {
			return nil, err
		}
	}
	if primary.UpstreamGroupRatio == nil || !validPriceNumber(*primary.UpstreamGroupRatio) {
		return nil, fmt.Errorf("%w: model=%s channel=%d", ErrPricingCostInvalid, policy.PublicModel, policy.PrimaryChannelID)
	}
	primaryMultiplier := *primary.UpstreamGroupRatio
	if !validPriceNumber(primary.InputCost) || !validPriceNumber(primary.OutputCost) || !validPriceNumber(primary.CacheReadCost) ||
		primary.InputCost <= 0 || primary.OutputCost <= 0 || primaryMultiplier <= 0 {
		return nil, fmt.Errorf("%w: model=%s channel=%d", ErrPricingCostInvalid, policy.PublicModel, policy.PrimaryChannelID)
	}
	primary.InputCost *= primaryMultiplier
	primary.OutputCost *= primaryMultiplier
	primary.CacheReadCost *= primaryMultiplier
	if backup != nil {
		if backup.UpstreamGroupRatio == nil || !validPriceNumber(*backup.UpstreamGroupRatio) ||
			!validPriceNumber(backup.InputCost) || !validPriceNumber(backup.OutputCost) || !validPriceNumber(backup.CacheReadCost) {
			return nil, fmt.Errorf("%w: model=%s channel=%d", ErrPricingCostInvalid, policy.PublicModel, policy.BackupChannelID)
		}
		backupMultiplier := *backup.UpstreamGroupRatio
		if backupMultiplier <= 0 {
			return nil, fmt.Errorf("%w: model=%s channel=%d", ErrPricingCostInvalid, policy.PublicModel, policy.BackupChannelID)
		}
		backup.InputCost *= backupMultiplier
		backup.OutputCost *= backupMultiplier
		backup.CacheReadCost *= backupMultiplier
	}
	fallback := float64(policy.FallbackProbabilityBPS) / 10000
	inputCost := primary.InputCost
	outputCost := primary.OutputCost
	stressInput := primary.InputCost
	stressOutput := primary.OutputCost
	if backup != nil {
		inputCost = primary.InputCost*(1-fallback) + backup.InputCost*fallback
		outputCost = primary.OutputCost*(1-fallback) + backup.OutputCost*fallback
		stressInput = backup.InputCost
		stressOutput = backup.OutputCost
	}
	overhead := 1 + float64(policy.OverheadBPS)/10000
	inputCost *= overhead
	outputCost *= overhead
	stressInput *= overhead
	stressOutput *= overhead
	targetDenominator := 1 - float64(policy.TargetMarginBPS)/10000
	minimumDenominator := 1 - float64(policy.MinimumMarginBPS)/10000
	inputPrice := inputCost / targetDenominator
	outputPrice := outputCost / targetDenominator
	if policy.ServiceTier == "stable" && backup != nil {
		inputPrice = math.Max(inputPrice, stressInput/minimumDenominator)
		outputPrice = math.Max(outputPrice, stressOutput/minimumDenominator)
	}
	inputPrice = roundPricingValue(inputPrice)
	outputPrice = roundPricingValue(outputPrice)
	margin := policy.TargetMarginBPS
	if inputPrice > 0 && outputPrice > 0 {
		inputMargin := int(math.Round((1 - inputCost/inputPrice) * 10000))
		outputMargin := int(math.Round((1 - outputCost/outputPrice) * 10000))
		margin = min(inputMargin, outputMargin)
	}
	snapshot, err := model.GetModelPricingSnapshot([]string{policy.PublicModel})
	if err != nil {
		return nil, err
	}
	version := snapshot.EmptyVersion
	if len(snapshot.Entries) == 1 {
		version = snapshot.Entries[0].Version
	}
	return &model.ModelPriceProposal{
		PublicModel: policy.PublicModel, ServiceTier: policy.ServiceTier, PolicyID: policy.ID,
		ExpectedInputCost: roundPricingValue(inputCost), ExpectedOutputCost: roundPricingValue(outputCost),
		ProposedInputPrice: inputPrice, ProposedOutputPrice: outputPrice,
		StressInputCost: roundPricingValue(stressInput), StressOutputCost: roundPricingValue(stressOutput),
		ExpectedMarginBPS: margin, Status: model.PricingProposalPending,
		Reason:         fmt.Sprintf("target_margin_bps=%d fallback_probability_bps=%d overhead_bps=%d", policy.TargetMarginBPS, policy.FallbackProbabilityBPS, policy.OverheadBPS),
		PricingVersion: version,
	}, nil
}

func currentOffer(publicModel string, channelID int, now int64) (*model.UpstreamModelOffer, error) {
	return currentOfferWithDB(model.DB, publicModel, channelID, now)
}

func currentOfferWithDB(db *gorm.DB, publicModel string, channelID int, now int64) (*model.UpstreamModelOffer, error) {
	var offer model.UpstreamModelOffer
	err := db.Where("public_model = ? AND channel_id = ? AND enabled = ? AND expires_at > ?", publicModel, channelID, true, now).
		Order("collected_at desc").First(&offer).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: model=%s channel=%d", ErrPricingOfferMissing, publicModel, channelID)
		}
		return nil, err
	}
	return &offer, nil
}

func ApprovePricingProposal(id int64, operatorID int) error {
	if id <= 0 || operatorID <= 0 {
		return ErrPricingProposalState
	}
	var proposal model.ModelPriceProposal
	if err := model.DB.Where("id = ? AND status = ?", id, model.PricingProposalPending).First(&proposal).Error; err != nil {
		return ErrPricingProposalState
	}
	snapshot, err := model.GetModelPricingSnapshot([]string{proposal.PublicModel})
	if err != nil {
		return err
	}
	version := snapshot.EmptyVersion
	if len(snapshot.Entries) == 1 {
		version = snapshot.Entries[0].Version
	}
	if version != proposal.PricingVersion {
		return fmt.Errorf("%w: current selling price changed", ErrPricingProposalStale)
	}
	return model.DB.Transaction(func(tx *gorm.DB) error {
		var policy model.ModelPricePolicy
		if err := tx.Where("id = ?", proposal.PolicyID).First(&policy).Error; err != nil {
			return fmt.Errorf("%w: pricing policy no longer exists", ErrPricingProposalStale)
		}
		now := common.GetTimestamp()
		if _, err := currentOfferWithDB(tx, proposal.PublicModel, policy.PrimaryChannelID, now); err != nil {
			return fmt.Errorf("%w: primary upstream offer is missing, disabled, or expired", ErrPricingProposalStale)
		}
		if policy.BackupChannelID > 0 {
			if _, err := currentOfferWithDB(tx, proposal.PublicModel, policy.BackupChannelID, now); err != nil {
				return fmt.Errorf("%w: backup upstream offer is missing, disabled, or expired", ErrPricingProposalStale)
			}
		}
		result := tx.Model(&model.ModelPriceProposal{}).
			Where("id = ? AND status = ?", id, model.PricingProposalPending).
			Updates(map[string]any{"status": model.PricingProposalApproved, "approved_by": operatorID, "approved_at": now, "updated_at": now})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrPricingProposalState
		}
		return nil
	})
}

func RejectPricingProposal(id int64) error {
	result := model.DB.Model(&model.ModelPriceProposal{}).
		Where("id = ? AND status IN ?", id, []string{model.PricingProposalPending, model.PricingProposalApproved}).
		Updates(map[string]any{"status": model.PricingProposalRejected, "updated_at": common.GetTimestamp()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return ErrPricingProposalState
	}
	return nil
}

func PublishPricingProposal(id int64) error {
	var proposal model.ModelPriceProposal
	if err := model.DB.Where("id = ? AND status = ?", id, model.PricingProposalApproved).First(&proposal).Error; err != nil {
		return ErrPricingProposalState
	}
	if proposal.ServiceTier != "default" {
		return errors.New("only the default service tier can publish a global model price")
	}
	change := pricingChangeForProposal(&proposal)
	if err := model.UpdateModelPricing([]model.ModelPricingChange{change}); err != nil {
		return err
	}
	result := model.DB.Model(&model.ModelPriceProposal{}).
		Where("id = ? AND status = ?", id, model.PricingProposalApproved).
		Updates(map[string]any{"status": model.PricingProposalPublished, "published_at": common.GetTimestamp(), "updated_at": common.GetTimestamp()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return ErrPricingProposalState
	}
	return nil
}

func PricingRisks() ([]map[string]any, error) {
	policies, err := model.ListModelPricePolicies()
	if err != nil {
		return nil, err
	}
	now := common.GetTimestamp()
	risks := make([]map[string]any, 0)
	for i := range policies {
		policy := policies[i]
		if !policy.Enabled {
			continue
		}
		primary, err := currentOffer(policy.PublicModel, policy.PrimaryChannelID, now)
		if err != nil {
			risks = append(risks, map[string]any{"public_model": policy.PublicModel, "service_tier": policy.ServiceTier, "code": "primary_offer_missing_or_expired"})
		} else if !validPriceNumber(primary.InputCost) || !validPriceNumber(primary.OutputCost) ||
			primary.InputCost <= 0 || primary.OutputCost <= 0 || primary.UpstreamGroupRatio == nil ||
			!validPriceNumber(*primary.UpstreamGroupRatio) || *primary.UpstreamGroupRatio <= 0 {
			risks = append(risks, map[string]any{"public_model": policy.PublicModel, "service_tier": policy.ServiceTier, "code": "primary_offer_cost_invalid"})
		}
		if policy.BackupChannelID > 0 {
			backup, err := currentOffer(policy.PublicModel, policy.BackupChannelID, now)
			if err != nil {
				risks = append(risks, map[string]any{"public_model": policy.PublicModel, "service_tier": policy.ServiceTier, "code": "backup_offer_missing_or_expired"})
			} else if !validPriceNumber(backup.InputCost) || !validPriceNumber(backup.OutputCost) ||
				backup.UpstreamGroupRatio == nil || !validPriceNumber(*backup.UpstreamGroupRatio) || *backup.UpstreamGroupRatio <= 0 {
				risks = append(risks, map[string]any{"public_model": policy.PublicModel, "service_tier": policy.ServiceTier, "code": "backup_offer_cost_invalid"})
			}
		}
	}
	return risks, nil
}

func validPriceNumber(value float64) bool {
	return value >= 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}
func roundPricingValue(value float64) float64 { return math.Round(value*1e8) / 1e8 }
func formatPricingValue(value float64) string {
	return strconv.FormatFloat(roundPricingValue(value), 'f', -1, 64)
}

func pricingChangeForProposal(proposal *model.ModelPriceProposal) model.ModelPricingChange {
	expression := fmt.Sprintf("tier(\"base\", p * %s + c * %s)", formatPricingValue(proposal.ProposedInputPrice), formatPricingValue(proposal.ProposedOutputPrice))
	return model.ModelPricingChange{
		ModelName: proposal.PublicModel, ExpectedVersion: proposal.PricingVersion,
		Pricing: model.PricingValues{"billing_setting.billing_mode": "tiered_expr", "billing_setting.billing_expr": expression},
	}
}

// simpleBaseTierExpr matches the expression PublishPricingProposal writes.
// Anything with cache, tools, or extra tiers stays on manual review.
var simpleBaseTierExpr = regexp.MustCompile(`^tier\("base",\s*p\s*\*\s*[0-9]+(?:\.[0-9]+)?\s*\+\s*c\s*\*\s*[0-9]+(?:\.[0-9]+)?\s*\)$`)

var autoPublishBlockedPricingKeys = []string{
	"AudioCompletionRatio", "AudioRatio", "CacheRatio", "CreateCacheRatio", "ImageRatio",
}

var errPricingAutoPublishSkip = errors.New("pricing auto-publish skipped")

// PricingAutoPublishStats is attached to a pricing_offer_sync task result.
type PricingAutoPublishStats struct {
	Published int                      `json:"published"`
	Skipped   int                      `json:"skipped"`
	Skips     []PricingAutoPublishSkip `json:"skips,omitempty"`
	Error     string                   `json:"error,omitempty"`
}

type PricingAutoPublishSkip struct {
	ProposalID  int64  `json:"proposal_id"`
	PublicModel string `json:"public_model,omitempty"`
	Reason      string `json:"reason"`
}

func pricingAutoPublishEnabled() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("PRICING_AUTO_PUBLISH")), "true")
}

// pricingAutoPublishVerifySamples returns how many recent cost snapshots must
// agree with the live offer before an unattended publish is allowed. Values
// below 2 disable the multi-sample verification entirely (legacy behavior).
func pricingAutoPublishVerifySamples() int {
	samples := common.GetEnvOrDefault("PRICING_AUTO_PUBLISH_VERIFY_SAMPLES", 3)
	if samples < 0 {
		return 0
	}
	return samples
}

// pricingAutoPublishVerifyToleranceBPS is the allowed cost deviation between
// the live offer and each verified snapshot, in basis points.
func pricingAutoPublishVerifyToleranceBPS() int {
	tolerance := common.GetEnvOrDefault("PRICING_AUTO_PUBLISH_VERIFY_TOLERANCE_BPS", 500)
	if tolerance < 0 {
		return 0
	}
	return tolerance
}

// pricingCostVerifyState reports whether multi-sample cost verification is
// active and whether the policy's channels passed it.
type pricingCostVerifyState struct {
	Required bool
	Passed   bool
}

func verifyPricingCosts(tx *gorm.DB, policy *model.ModelPricePolicy) pricingCostVerifyState {
	state := pricingCostVerifyState{}
	samples := pricingAutoPublishVerifySamples()
	if samples < 2 {
		return state
	}
	state.Required = true
	tolerance := pricingAutoPublishVerifyToleranceBPS()
	now := common.GetTimestamp()
	channelIDs := []int{policy.PrimaryChannelID}
	if policy.BackupChannelID > 0 && policy.BackupChannelID != policy.PrimaryChannelID {
		channelIDs = append(channelIDs, policy.BackupChannelID)
	}
	for _, channelID := range channelIDs {
		offer, err := currentOfferWithDB(tx, policy.PublicModel, channelID, now)
		if err != nil {
			return state
		}
		snapshots, err := model.RecentUpstreamCostSnapshots(tx, channelID, policy.PublicModel, samples)
		if err != nil || len(snapshots) < samples {
			return state
		}
		for _, snapshot := range snapshots {
			if !costsWithinTolerance(offer.InputCost, snapshot.InputCost, tolerance) ||
				!costsWithinTolerance(offer.OutputCost, snapshot.OutputCost, tolerance) {
				return state
			}
		}
	}
	state.Passed = true
	return state
}

func costsWithinTolerance(current, sampled float64, toleranceBPS int) bool {
	if math.Abs(sampled) <= 1e-12 {
		return math.Abs(current-sampled) <= 1e-12
	}
	return math.Abs(current-sampled) <= math.Abs(sampled)*float64(toleranceBPS)/10000
}

// pricingCoversStressCost is the absolute floor for unattended pricing: the
// proposed prices must cover the worst-case (stress) procurement cost even
// when the policy's minimum margin is zero.
func pricingCoversStressCost(proposal *model.ModelPriceProposal) bool {
	return roundPricingValue(proposal.ProposedInputPrice) >= roundPricingValue(proposal.StressInputCost) &&
		roundPricingValue(proposal.ProposedOutputPrice) >= roundPricingValue(proposal.StressOutputCost)
}

// AutoPublishEligibleProposals publishes pending default-tier proposals that
// pass the unattended gates. Each proposal commits in its own transaction.
// Proposals that fail a gate stay pending.
func AutoPublishEligibleProposals(ctx context.Context, proposalIDs []int64) PricingAutoPublishStats {
	stats := PricingAutoPublishStats{}
	if len(proposalIDs) == 0 {
		return stats
	}
	operator, err := pricingAutoPublishOperator()
	if err != nil {
		stats.Skipped = len(proposalIDs)
		stats.Error = err.Error()
		return stats
	}
	for i, id := range proposalIDs {
		if ctx != nil && ctx.Err() != nil {
			stats.Error = ctx.Err().Error()
			stats.Skipped += len(proposalIDs) - i
			return stats
		}
		published, skip, err := autoPublishProposal(id, operator)
		if err != nil {
			stats.Error = err.Error()
			stats.Skipped += len(proposalIDs) - i
			return stats
		}
		if skip.Reason != "" {
			stats.Skipped++
			stats.Skips = append(stats.Skips, skip)
			continue
		}
		if published != nil {
			stats.Published++
			recordPricingAutoPublishAudit(operator, "pricing.proposal.approve", published.ID)
			recordPricingAutoPublishAudit(operator, "pricing.proposal.publish", published.ID)
		}
	}
	return stats
}

func pricingAutoPublishOperator() (*model.User, error) {
	raw := strings.TrimSpace(os.Getenv("PRICING_AUTO_PUBLISH_OPERATOR_ID"))
	if raw == "" {
		return nil, errors.New("PRICING_AUTO_PUBLISH_OPERATOR_ID is required when PRICING_AUTO_PUBLISH=true")
	}
	id, err := strconv.Atoi(raw)
	if err != nil || id <= 0 {
		return nil, errors.New("PRICING_AUTO_PUBLISH_OPERATOR_ID must be a positive user id")
	}
	user, err := model.GetUserById(id, false)
	if err != nil {
		return nil, fmt.Errorf("PRICING_AUTO_PUBLISH_OPERATOR_ID: %w", err)
	}
	if user.Role != common.RoleRootUser || user.Status != common.UserStatusEnabled {
		return nil, errors.New("PRICING_AUTO_PUBLISH_OPERATOR_ID must be an enabled root user")
	}
	return user, nil
}

func autoPublishProposal(id int64, operator *model.User) (*model.ModelPriceProposal, PricingAutoPublishSkip, error) {
	skip := PricingAutoPublishSkip{ProposalID: id}
	var published *model.ModelPriceProposal
	err := model.MutateModelPricing(func(tx *gorm.DB, values map[string]map[string]any) error {
		proposal, err := model.LoadModelPriceProposalForUpdate(tx, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				skip.Reason = "not_pending"
				return errPricingAutoPublishSkip
			}
			return err
		}
		skip.PublicModel = proposal.PublicModel
		if proposal.Status != model.PricingProposalPending {
			skip.Reason = "not_pending"
			return errPricingAutoPublishSkip
		}
		policy, err := model.LoadModelPricePolicyForUpdate(tx, proposal.PolicyID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				skip.Reason = "policy_unavailable"
				return errPricingAutoPublishSkip
			}
			return err
		}
		offersOK := pricingOffersCurrent(tx, &policy)
		costVerify := verifyPricingCosts(tx, &policy)
		reason := autoPublishSkipReason(&policy, &proposal, model.EffectiveModelPricing(values, proposal.PublicModel), offersOK, costVerify)
		if reason != "" {
			skip.Reason = reason
			return errPricingAutoPublishSkip
		}
		if err := model.RejectOtherApprovedPriceProposals(tx, proposal.PublicModel, proposal.ID); err != nil {
			return err
		}
		change := pricingChangeForProposal(&proposal)
		if err := model.ApplyModelPricingChange(values, change); err != nil {
			if errors.Is(err, model.ErrModelPricingConflict) {
				skip.Reason = "stale_version"
				return errPricingAutoPublishSkip
			}
			return err
		}
		now := common.GetTimestamp()
		result := tx.Model(&model.ModelPriceProposal{}).
			Where("id = ? AND status = ?", proposal.ID, model.PricingProposalPending).
			Updates(map[string]any{
				"status":       model.PricingProposalPublished,
				"approved_by":  operator.Id,
				"approved_at":  now,
				"published_at": now,
				"updated_at":   now,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrPricingProposalState
		}
		published = &proposal
		return nil
	})
	if errors.Is(err, errPricingAutoPublishSkip) {
		return nil, skip, nil
	}
	if err != nil {
		return nil, PricingAutoPublishSkip{}, err
	}
	return published, PricingAutoPublishSkip{}, nil
}

func pricingOffersCurrent(tx *gorm.DB, policy *model.ModelPricePolicy) bool {
	now := common.GetTimestamp()
	if _, err := currentOfferWithDB(tx, policy.PublicModel, policy.PrimaryChannelID, now); err != nil {
		return false
	}
	if policy.BackupChannelID > 0 {
		if _, err := currentOfferWithDB(tx, policy.PublicModel, policy.BackupChannelID, now); err != nil {
			return false
		}
	}
	return true
}

func autoPublishSkipReason(policy *model.ModelPricePolicy, proposal *model.ModelPriceProposal, effective model.PricingValues, offersOK bool, costVerify pricingCostVerifyState) string {
	if policy == nil || policy.ID == 0 || policy.PublicModel != proposal.PublicModel {
		return "policy_unavailable"
	}
	if !policy.Enabled {
		return "policy_disabled"
	}
	if policy.PriceLocked {
		return "price_locked"
	}
	if proposal.ServiceTier != "default" || policy.ServiceTier != "default" {
		return "service_tier"
	}
	if !offersOK {
		return "offer_stale"
	}
	if policy.AutoPublishChangeBPS <= 0 {
		return "auto_publish_disabled"
	}
	if !pricingShapeAllowsAutoPublish(effective) {
		return "pricing_shape"
	}
	expression, _ := effective["billing_setting.billing_expr"].(string)
	currentInput, currentOutput, priced := liveTokenPrices(expression)
	proposedInput := roundPricingValue(proposal.ProposedInputPrice)
	proposedOutput := roundPricingValue(proposal.ProposedOutputPrice)
	if priced && proposedInput == currentInput && proposedOutput == currentOutput {
		return "unchanged"
	}
	// A price cut is only unattended-safe when the cost drop is confirmed by
	// multiple consistent samples; otherwise it stays on manual review.
	if priced && (proposedInput < currentInput || proposedOutput < currentOutput) && !(costVerify.Required && costVerify.Passed) {
		return "price_cut"
	}
	if !priced ||
		!priceChangeWithinBPS(currentInput, proposedInput, policy.AutoPublishChangeBPS) ||
		!priceChangeWithinBPS(currentOutput, proposedOutput, policy.AutoPublishChangeBPS) {
		return "change_band"
	}
	if proposal.ExpectedMarginBPS < policy.MinimumMarginBPS {
		return "margin"
	}
	if !pricingCoversStressCost(proposal) {
		return "below_cost"
	}
	if costVerify.Required && !costVerify.Passed {
		return "cost_unverified"
	}
	return ""
}

func pricingShapeAllowsAutoPublish(effective model.PricingValues) bool {
	for _, key := range autoPublishBlockedPricingKeys {
		if _, exists := effective[key]; exists {
			return false
		}
	}
	expression, _ := effective["billing_setting.billing_expr"].(string)
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return true
	}
	return simpleBaseTierExpr.MatchString(expression)
}

func liveTokenPrices(expression string) (float64, float64, bool) {
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return 0, 0, false
	}
	input, _, inputErr := billingexpr.RunExpr(expression, billingexpr.TokenParams{P: 1, Len: 1})
	output, _, outputErr := billingexpr.RunExpr(expression, billingexpr.TokenParams{C: 1})
	if inputErr != nil || outputErr != nil || !validPriceNumber(input) || !validPriceNumber(output) || input <= 0 || output <= 0 {
		return 0, 0, false
	}
	return roundPricingValue(input), roundPricingValue(output), true
}

func priceChangeWithinBPS(current, proposed float64, limit int) bool {
	if current <= 0 || limit <= 0 || !validPriceNumber(current) || !validPriceNumber(proposed) {
		return false
	}
	change := int(math.Round((proposed/current - 1) * 10000))
	if change < 0 {
		change = -change
	}
	return change <= limit
}

func recordPricingAutoPublishAudit(operator *model.User, action string, proposalID int64) {
	params := map[string]interface{}{"id": proposalID, "auto": true}
	content := action
	switch action {
	case "pricing.proposal.approve":
		content = fmt.Sprintf("Approved pricing proposal %d", proposalID)
	case "pricing.proposal.publish":
		content = fmt.Sprintf("Published pricing proposal %d", proposalID)
	}
	model.RecordOperationAuditLog(operator.Id, operator.Role, content, "", action, params, &model.AuditAdminInfo{
		AdminID:       operator.Id,
		AdminUsername: operator.Username,
		AdminRole:     operator.Role,
		AuthMethod:    "system_task",
	}, nil)
}

type pricingRecalculateHandler struct{}

func (pricingRecalculateHandler) Type() string { return model.SystemTaskTypePricingRecalculate }
func (pricingRecalculateHandler) Enabled() bool {
	return strings.EqualFold(os.Getenv("PRICING_CONTROL_ENABLED"), "true")
}
func (pricingRecalculateHandler) Interval() time.Duration {
	minutes := common.GetEnvOrDefault("PRICING_RECALCULATE_INTERVAL_MINUTES", 60)
	if minutes < 5 {
		minutes = 5
	}
	return time.Duration(minutes) * time.Minute
}
func (pricingRecalculateHandler) NewPayload() any { return map[string]any{"scheduled": true} }
func (pricingRecalculateHandler) Run(ctx context.Context, task *model.SystemTask, runnerID string) {
	created, err := RecalculatePricingProposals(ctx)
	if err != nil {
		failSystemTask(task, runnerID, err)
		return
	}
	if err := model.FinishSystemTask(task.TaskID, runnerID, model.SystemTaskStatusSucceeded, map[string]int{"created_proposals": created}, ""); err != nil {
		logSystemTaskLockError(ctx, task, err)
	}
}

func init() { RegisterSystemTaskHandler(pricingRecalculateHandler{}) }
