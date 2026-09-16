package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
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
		if offer.SuccessRateBPS < 0 || offer.SuccessRateBPS > 10000 {
			return fmt.Errorf("%w at item %d", ErrPricingOfferInvalid, i)
		}
		if offer.CollectedAt <= 0 {
			offer.CollectedAt = now
		}
		if offer.ExpiresAt <= offer.CollectedAt {
			offer.ExpiresAt = offer.CollectedAt + 24*60*60
		}
		if err := model.UpsertUpstreamModelOffer(offer); err != nil {
			return err
		}
	}
	return nil
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
	policies, err := model.ListModelPricePolicies()
	if err != nil {
		return 0, err
	}
	created := 0
	for i := range policies {
		if ctx != nil && ctx.Err() != nil {
			return created, ctx.Err()
		}
		policy := policies[i]
		if !policy.Enabled || policy.PriceLocked {
			continue
		}
		proposal, err := buildPriceProposal(&policy)
		if err != nil {
			if errors.Is(err, ErrPricingOfferMissing) {
				continue
			}
			return created, err
		}
		if err := model.ReplacePendingModelPriceProposal(proposal); err != nil {
			return created, err
		}
		created++
	}
	return created, nil
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
	var offer model.UpstreamModelOffer
	err := model.DB.Where("public_model = ? AND channel_id = ? AND enabled = ? AND expires_at > ?", publicModel, channelID, true, now).
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
	result := model.DB.Model(&model.ModelPriceProposal{}).
		Where("id = ? AND status = ?", id, model.PricingProposalPending).
		Updates(map[string]any{"status": model.PricingProposalApproved, "approved_by": operatorID, "approved_at": common.GetTimestamp(), "updated_at": common.GetTimestamp()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return ErrPricingProposalState
	}
	return nil
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
	expression := fmt.Sprintf("tier(\"base\", p * %s + c * %s)", formatPricingValue(proposal.ProposedInputPrice), formatPricingValue(proposal.ProposedOutputPrice))
	change := model.ModelPricingChange{
		ModelName: proposal.PublicModel, ExpectedVersion: proposal.PricingVersion,
		Pricing: model.PricingValues{"billing_setting.billing_mode": "tiered_expr", "billing_setting.billing_expr": expression},
	}
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
		if _, err := currentOffer(policy.PublicModel, policy.PrimaryChannelID, now); err != nil {
			risks = append(risks, map[string]any{"public_model": policy.PublicModel, "service_tier": policy.ServiceTier, "code": "primary_offer_missing_or_expired"})
		}
		if policy.BackupChannelID > 0 {
			if _, err := currentOffer(policy.PublicModel, policy.BackupChannelID, now); err != nil {
				risks = append(risks, map[string]any{"public_model": policy.PublicModel, "service_tier": policy.ServiceTier, "code": "backup_offer_missing_or_expired"})
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
