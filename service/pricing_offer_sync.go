package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/pkg/billingexpr"
	"github.com/QuantumNous/new-api/setting/billing_setting"
)

const pricingOfferSyncResponseLimit = 10 << 20

type PricingOfferSyncChannelResult struct {
	ChannelID int      `json:"channel_id"`
	Imported  int      `json:"imported"`
	Kept      int      `json:"kept,omitempty"`
	Skipped   int      `json:"skipped"`
	Warnings  []string `json:"warnings,omitempty"`
	Error     string   `json:"error,omitempty"`
}

type PricingOfferSyncResult struct {
	Imported         int                             `json:"imported"`
	Kept             int                             `json:"kept,omitempty"`
	Skipped          int                             `json:"skipped"`
	CreatedProposals int                             `json:"created_proposals,omitempty"`
	AutoPublish      *PricingAutoPublishStats        `json:"auto_publish,omitempty"`
	Channels         []PricingOfferSyncChannelResult `json:"channels"`
	Trigger          string                          `json:"trigger,omitempty"`
}

type upstreamPricingResponse struct {
	Success         bool                          `json:"success"`
	Message         string                        `json:"message"`
	PricingVersion  string                        `json:"pricing_version"`
	GroupRatio      map[string]float64            `json:"group_ratio"`
	ModelGroupRatio map[string]map[string]float64 `json:"model_group_ratio"`
	AutoGroups      []string                      `json:"auto_groups"`
	Data            []upstreamPriceItem           `json:"data"`
}

type upstreamPriceItem struct {
	ModelName       string   `json:"model_name"`
	QuotaType       int      `json:"quota_type"`
	ModelRatio      float64  `json:"model_ratio"`
	ModelPrice      float64  `json:"model_price"`
	CompletionRatio float64  `json:"completion_ratio"`
	CacheRatio      *float64 `json:"cache_ratio"`
	EnableGroups    []string `json:"enable_groups"`
	BillingMode     string   `json:"billing_mode"`
	BillingExpr     string   `json:"billing_expr"`
	PricingVersion  string   `json:"pricing_version"`
}

func SyncPricingOffersFromChannels(ctx context.Context, channelIDs []int) (PricingOfferSyncResult, error) {
	warnPricingOfferTTL(ctx)
	if len(channelIDs) == 0 {
		policies, err := model.ListModelPricePolicies()
		if err != nil {
			return PricingOfferSyncResult{}, err
		}
		seen := make(map[int]struct{}, len(policies)*2)
		for i := range policies {
			if !policies[i].Enabled {
				continue
			}
			seen[policies[i].PrimaryChannelID] = struct{}{}
			if policies[i].BackupChannelID > 0 {
				seen[policies[i].BackupChannelID] = struct{}{}
			}
		}
		channelIDs = make([]int, 0, len(seen))
		for id := range seen {
			if id > 0 {
				channelIDs = append(channelIDs, id)
			}
		}
		sort.Ints(channelIDs)
	}
	if len(channelIDs) == 0 {
		return PricingOfferSyncResult{}, errors.New("no pricing channels configured")
	}

	channels, err := model.GetChannelsByIds(channelIDs)
	if err != nil {
		return PricingOfferSyncResult{}, err
	}
	byID := make(map[int]*model.Channel, len(channels))
	for i := range channels {
		byID[channels[i].Id] = channels[i]
	}

	result := PricingOfferSyncResult{Channels: make([]PricingOfferSyncChannelResult, 0, len(channelIDs))}
	for _, channelID := range channelIDs {
		channelResult := PricingOfferSyncChannelResult{ChannelID: channelID}
		channel := byID[channelID]
		if channel == nil {
			channelResult.Error = "channel not found"
			result.Channels = append(result.Channels, channelResult)
			continue
		}
		offers, warnings, syncErr := fetchPricingOffersForChannel(ctx, channel)
		channelResult.Warnings = warnings
		channelResult.Skipped = len(warnings)
		result.Skipped += channelResult.Skipped
		if syncErr != nil {
			channelResult.Error = syncErr.Error()
			result.Channels = append(result.Channels, channelResult)
			continue
		}
		if len(offers) == 0 {
			channelResult.Error = "upstream returned no importable token prices"
			result.Channels = append(result.Channels, channelResult)
			continue
		}
		stats, importErr := ValidateAndImportPricingOffersWithStats(offers)
		if importErr != nil {
			channelResult.Error = importErr.Error()
			result.Channels = append(result.Channels, channelResult)
			continue
		}
		channelResult.Imported = stats.Applied
		channelResult.Kept = stats.Kept
		result.Imported += stats.Applied
		result.Kept += stats.Kept
		result.Channels = append(result.Channels, channelResult)
	}
	if result.Imported == 0 && result.Kept == 0 {
		return result, errors.New("pricing sync did not import any offers")
	}
	return result, nil
}

const defaultUpstreamPricingPath = "/api/pricing"

// upstreamPricingEndpoint resolves where a channel's price list lives.
// A channel-level override (dto.ChannelSettings.pricing_endpoint) wins over
// the historical hard-coded "/api/pricing", so the dynamic pricing collector
// and the admin "sync endpoint" selector finally agree on one URL.
func upstreamPricingEndpoint(channel *model.Channel, baseURL *url.URL) (*url.URL, error) {
	endpoint := strings.TrimSpace(channel.GetSetting().PricingEndpoint)
	if endpoint == "" {
		endpoint = defaultUpstreamPricingPath
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid pricing endpoint %q: %w", endpoint, err)
	}
	if parsed.IsAbs() {
		parsed.RawQuery = ""
		parsed.Fragment = ""
		return parsed, nil
	}
	if !strings.HasPrefix(endpoint, "/") {
		return nil, fmt.Errorf("pricing endpoint %q must be absolute or start with /", endpoint)
	}
	resolved := *baseURL
	resolved.Path = endpoint
	resolved.RawQuery = ""
	resolved.Fragment = ""
	return &resolved, nil
}

func fetchPricingOffersForChannel(ctx context.Context, channel *model.Channel) ([]model.UpstreamModelOffer, []string, error) {
	baseURL, err := url.Parse(strings.TrimSpace(channel.GetBaseURL()))
	if err != nil || baseURL.Scheme == "" || baseURL.Host == "" {
		return nil, nil, errors.New("channel has no valid upstream base URL")
	}
	pricingURL, err := upstreamPricingEndpoint(channel, baseURL)
	if err != nil {
		return nil, nil, err
	}

	key, _, keyErr := channel.GetNextEnabledKey()
	if keyErr != nil || strings.TrimSpace(key) == "" {
		return nil, nil, errors.New("channel has no enabled API key")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pricingURL.String(), nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(key))
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("upstream pricing returned %s", resp.Status)
	}
	var payload upstreamPricingResponse
	if err := common.DecodeJson(io.LimitReader(resp.Body, pricingOfferSyncResponseLimit), &payload); err != nil {
		return nil, nil, err
	}
	if !payload.Success {
		if payload.Message == "" {
			payload.Message = "upstream pricing request failed"
		}
		return nil, nil, errors.New(payload.Message)
	}

	allowedModels := make(map[string]struct{})
	for _, name := range strings.Split(channel.Models, ",") {
		if name = strings.TrimSpace(name); name != "" {
			allowedModels[name] = struct{}{}
		}
	}
	modelMapping := make(map[string]string)
	if raw := strings.TrimSpace(channel.GetModelMapping()); raw != "" {
		if err := common.UnmarshalJsonStr(raw, &modelMapping); err != nil {
			return nil, nil, fmt.Errorf("invalid channel model mapping: %w", err)
		}
	}
	publicByUpstream := make(map[string]string, len(modelMapping))
	for publicName, upstreamName := range modelMapping {
		if _, exists := publicByUpstream[upstreamName]; !exists {
			publicByUpstream[upstreamName] = publicName
		}
	}

	policy := pricingGroupRatioPolicy()
	now := common.GetTimestamp()
	ttl := pricingOfferTTLSeconds()
	offers := make([]model.UpstreamModelOffer, 0, len(payload.Data))
	warnings := make([]string, 0)
	for i := range payload.Data {
		item := payload.Data[i]
		publicModel := item.ModelName
		if mapped := publicByUpstream[item.ModelName]; mapped != "" {
			publicModel = mapped
		}
		_, upstreamAllowed := allowedModels[item.ModelName]
		_, publicAllowed := allowedModels[publicModel]
		if len(allowedModels) > 0 && !upstreamAllowed && !publicAllowed {
			continue
		}
		groupRatio, warningsForModel := resolveUpstreamGroupRatio(&payload, item, policy)
		if len(warningsForModel) > 0 {
			warnings = append(warnings, warningsForModel...)
			continue
		}

		inputCost, outputCost, cacheReadCost, priceErr := normalizedUpstreamTokenCosts(item)
		if priceErr != nil {
			warnings = append(warnings, item.ModelName+": "+priceErr.Error())
			continue
		}
		version := item.PricingVersion
		if version == "" {
			version = payload.PricingVersion
		}
		ratio := groupRatio.Ratio
		offers = append(offers, model.UpstreamModelOffer{
			ChannelID: channel.Id, UpstreamModel: item.ModelName, PublicModel: publicModel,
			InputCost: inputCost, OutputCost: outputCost, CacheReadCost: cacheReadCost,
			UpstreamGroupRatio: &ratio, UpstreamGroup: groupRatio.Group, Currency: "USD", SourceType: "new-api-pricing",
			SourceURL: pricingURL.String(), SourceVersion: version, SuccessRateBPS: 10000,
			CollectedAt: now, ExpiresAt: now + ttl, Enabled: true,
		})
	}
	return offers, warnings, nil
}

type upstreamRatioChoice struct {
	Group string
	Ratio float64
	// Scoped is true when the ratio came from the model-level table
	// (model_group_ratio[model][group]) rather than the top-level group_ratio.
	Scoped bool
	// MinRatio and MaxRatio bound the group across both ratio layers. They are
	// what the max/min policies compare, so an uncertain upstream can never
	// silently understate procurement cost.
	MinRatio float64
	MaxRatio float64
}

const (
	pricingGroupRatioPolicyKeyGroup = "key_group"
	pricingGroupRatioPolicyMax      = "max"
	pricingGroupRatioPolicyMin      = "min"
	pricingGroupRatioPolicyMedian   = "median"
)

// pricingGroupRatioPolicy selects how the settlement multiplier is chosen when
// an upstream model is served by several groups. A static /api/pricing response
// does not reveal which group a runtime request will settle on, so operators
// must state the assumption explicitly.
//
// key_group follows the provider's own auto_groups order, max/min bound the
// uncertainty, and median is the neutral estimate. Median is the default
// because measuring real traffic showed per-model ratios spanning 0.002-1.0 on
// the same response: max over-priced models by up to 15x while the private
// settlement group sat near the median. Calibration supersedes this guess as
// soon as a model has traffic.
func pricingGroupRatioPolicy() string {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("PRICING_GROUP_RATIO_POLICY"))) {
	case pricingGroupRatioPolicyMax:
		return pricingGroupRatioPolicyMax
	case pricingGroupRatioPolicyMin:
		return pricingGroupRatioPolicyMin
	case pricingGroupRatioPolicyKeyGroup:
		return pricingGroupRatioPolicyKeyGroup
	default:
		return pricingGroupRatioPolicyMedian
	}
}

// ratioFor resolves one group's multiplier, preferring the model-level
// override table when the upstream provides it. MinRatio/MaxRatio keep both
// layers so the max/min policies can stay conservative without guessing which
// layer the upstream actually bills with.
func (payload *upstreamPricingResponse) ratioFor(modelName, group string) (upstreamRatioChoice, bool) {
	if group == "" {
		return upstreamRatioChoice{}, false
	}
	topRatio, hasTop := payload.GroupRatio[group]
	hasTop = hasTop && topRatio > 0 && validPriceNumber(topRatio)
	scopedRatio, hasScoped := 0.0, false
	if byGroup, ok := payload.ModelGroupRatio[modelName]; ok {
		if ratio, exists := byGroup[group]; exists && ratio > 0 && validPriceNumber(ratio) {
			scopedRatio, hasScoped = ratio, true
		}
	}
	switch {
	case hasScoped && hasTop:
		return upstreamRatioChoice{
			Group: group, Ratio: scopedRatio, Scoped: true,
			MinRatio: math.Min(scopedRatio, topRatio), MaxRatio: math.Max(scopedRatio, topRatio),
		}, true
	case hasScoped:
		return upstreamRatioChoice{Group: group, Ratio: scopedRatio, Scoped: true, MinRatio: scopedRatio, MaxRatio: scopedRatio}, true
	case hasTop:
		return upstreamRatioChoice{Group: group, Ratio: topRatio, MinRatio: topRatio, MaxRatio: topRatio}, true
	}
	return upstreamRatioChoice{}, false
}

// resolveUpstreamGroupRatio picks the procurement multiplier for one model.
//
// key_group keeps the upstream auto-group order (the key's own candidates),
// max/min widen the choice to every group that can serve the model. max is the
// conservative default for keys whose settlement group cannot be proven,
// because understating cost silently eats margin.
func resolveUpstreamGroupRatio(payload *upstreamPricingResponse, item upstreamPriceItem, policy string) (upstreamRatioChoice, []string) {
	candidates := make([]upstreamRatioChoice, 0, len(item.EnableGroups))
	seen := make(map[string]struct{}, len(item.EnableGroups))
	for _, group := range item.EnableGroups {
		group = strings.TrimSpace(group)
		if group == "" {
			continue
		}
		if _, dup := seen[group]; dup {
			continue
		}
		seen[group] = struct{}{}
		if choice, ok := payload.ratioFor(item.ModelName, group); ok {
			candidates = append(candidates, choice)
		}
	}

	switch policy {
	case pricingGroupRatioPolicyMax, pricingGroupRatioPolicyMin:
		if len(candidates) == 0 {
			return upstreamRatioChoice{}, []string{ratioWarning(item, payload, candidates)}
		}
		chosen := candidates[0]
		best := chosen.MaxRatio
		if policy == pricingGroupRatioPolicyMin {
			best = chosen.MinRatio
		}
		for _, candidate := range candidates[1:] {
			if policy == pricingGroupRatioPolicyMax && candidate.MaxRatio > best {
				chosen, best = candidate, candidate.MaxRatio
			}
			if policy == pricingGroupRatioPolicyMin && candidate.MinRatio < best {
				chosen, best = candidate, candidate.MinRatio
			}
		}
		chosen.Ratio = best
		return chosen, nil
	case pricingGroupRatioPolicyMedian:
		// The provider never publishes which group a model settles through, so
		// neither the maximum nor the minimum is right in general: max overprices
		// a cheap route, min under-prices an expensive one. The median is the
		// neutral estimate, and calibration corrects it once traffic exists.
		if len(candidates) == 0 {
			return upstreamRatioChoice{}, []string{ratioWarning(item, payload, candidates)}
		}
		values := make([]float64, 0, len(candidates))
		for _, candidate := range candidates {
			values = append(values, candidate.Ratio)
		}
		sort.Float64s(values)
		median := values[len(values)/2]
		if len(values)%2 == 0 {
			median = (values[len(values)/2-1] + values[len(values)/2]) / 2
		}
		chosen := candidates[0]
		bestDelta := math.Abs(chosen.Ratio - median)
		for _, candidate := range candidates[1:] {
			if delta := math.Abs(candidate.Ratio - median); delta < bestDelta {
				chosen, bestDelta = candidate, delta
			}
		}
		chosen.Ratio = median
		return chosen, nil
	default:
		for _, autoGroup := range payload.AutoGroups {
			for _, candidate := range candidates {
				if candidate.Group == autoGroup {
					return candidate, nil
				}
			}
		}
		return upstreamRatioChoice{}, []string{ratioWarning(item, payload, candidates)}
	}
}

func ratioWarning(item upstreamPriceItem, payload *upstreamPricingResponse, candidates []upstreamRatioChoice) string {
	available := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		available = append(available, fmt.Sprintf("%s=%s", candidate.Group, trimFloat(candidate.Ratio)))
	}
	if len(available) == 0 {
		available = append(available, "none")
	}
	return fmt.Sprintf("%s: no usable upstream group ratio (enable_groups=[%s], auto_groups=[%s], candidates=[%s])",
		item.ModelName, strings.Join(item.EnableGroups, ","), strings.Join(payload.AutoGroups, ","), strings.Join(available, ","))
}

func trimFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

// pricingOfferTTLSeconds is the lifetime of a collected offer. It must be at
// least as long as the sync interval, otherwise the inventory is empty and
// every policy lands in the risk list between two successful syncs.
func pricingOfferTTLSeconds() int64 {
	minutes := common.GetEnvOrDefault("PRICING_OFFER_TTL_MINUTES", 120)
	if minutes < 15 {
		minutes = 15
	}
	return int64(minutes) * 60
}

var pricingOfferTTLWarningOnce sync.Once

func warnPricingOfferTTL(ctx context.Context) {
	pricingOfferTTLWarningOnce.Do(func() {
		ttlMinutes := common.GetEnvOrDefault("PRICING_OFFER_TTL_MINUTES", 120)
		intervalMinutes := common.GetEnvOrDefault("PRICING_OFFER_SYNC_INTERVAL_MINUTES", 60)
		if intervalMinutes < 15 {
			intervalMinutes = 15
		}
		if ttlMinutes < intervalMinutes {
			logger.LogWarn(ctx, fmt.Sprintf(
				"PRICING_OFFER_TTL_MINUTES=%d is shorter than PRICING_OFFER_SYNC_INTERVAL_MINUTES=%d; offers will expire before the next sync, leaving the cost inventory empty and every policy in the risk list",
				ttlMinutes, intervalMinutes))
		}
		if ttlMinutes < 15 {
			logger.LogWarn(ctx, "PRICING_OFFER_TTL_MINUTES below 15 is clamped to 15")
		}
	})
}

type pricingOfferSyncHandler struct{}

func (pricingOfferSyncHandler) Type() string { return model.SystemTaskTypePricingOfferSync }
func (pricingOfferSyncHandler) Enabled() bool {
	return strings.EqualFold(os.Getenv("PRICING_CONTROL_ENABLED"), "true")
}
func (pricingOfferSyncHandler) Interval() time.Duration {
	minutes := common.GetEnvOrDefault("PRICING_OFFER_SYNC_INTERVAL_MINUTES", 60)
	if minutes < 15 {
		minutes = 15
	}
	return time.Duration(minutes) * time.Minute
}
func (pricingOfferSyncHandler) NewPayload() any { return map[string]any{"scheduled": true} }
func (pricingOfferSyncHandler) Run(ctx context.Context, task *model.SystemTask, runnerID string) {
	result, err := RunPricingOfferSync(ctx, nil, "scheduled")
	if err != nil {
		if result.AutoPublish != nil {
			logger.LogWarn(ctx, fmt.Sprintf("pricing auto-publish failed: %s", err.Error()))
			if finishErr := model.FinishSystemTask(task.TaskID, runnerID, model.SystemTaskStatusFailed, result, err.Error()); finishErr != nil {
				logSystemTaskLockError(ctx, task, finishErr)
			}
			return
		}
		failSystemTask(task, runnerID, err)
		return
	}
	if err := model.FinishSystemTask(task.TaskID, runnerID, model.SystemTaskStatusSucceeded, result, ""); err != nil {
		logSystemTaskLockError(ctx, task, err)
	}
}

// RunPricingOfferSync is the single entry point for both the manual button and
// the scheduled task: collect, recalculate, optionally auto-publish, then
// persist the run so the console can show the last result after a refresh.
func RunPricingOfferSync(ctx context.Context, channelIDs []int, trigger string) (PricingOfferSyncResult, error) {
	startedAt := common.GetTimestamp()
	result, err := SyncPricingOffersFromChannels(ctx, channelIDs)
	result.Trigger = trigger
	if err == nil {
		err = CompletePricingOfferSync(ctx, &result)
	}
	recordPricingSyncRun(ctx, trigger, startedAt, &result, err)
	return result, err
}

func recordPricingSyncRun(ctx context.Context, trigger string, startedAt int64, result *PricingOfferSyncResult, syncErr error) {
	status := "succeeded"
	message := ""
	if syncErr != nil {
		status = "failed"
		message = syncErr.Error()
		if result.Imported > 0 || result.Kept > 0 {
			status = "partial"
		}
	}
	channels := ""
	if encoded, err := common.Marshal(result.Channels); err == nil {
		channels = string(encoded)
	}
	run := &model.PricingSyncRun{
		Trigger: trigger, Status: status,
		Imported: result.Imported, Kept: result.Kept, Skipped: result.Skipped,
		CreatedProposals: result.CreatedProposals, Channels: channels, ErrorMessage: message,
		StartedAt: startedAt, FinishedAt: common.GetTimestamp(),
	}
	if err := model.RecordPricingSyncRun(run); err != nil {
		logger.LogWarn(ctx, fmt.Sprintf("failed to record pricing sync run: %s", err.Error()))
	}
}

// CompletePricingOfferSync recalculates proposals after a successful import.
// Both the manual button and the scheduled task go through this function so the
// same payload always produces the same outcome. Selling prices change only
// when PRICING_AUTO_PUBLISH=true; the separate pricing_recalculate task stays
// proposal-only.
func CompletePricingOfferSync(ctx context.Context, result *PricingOfferSyncResult) error {
	created, ids, err := recalculatePricingProposals(ctx)
	if err != nil {
		return err
	}
	result.CreatedProposals = created
	if !pricingAutoPublishEnabled() {
		return nil
	}
	stats := AutoPublishEligibleProposals(ctx, ids)
	result.AutoPublish = &stats
	if stats.Error != "" {
		return errors.New(stats.Error)
	}
	return nil
}

func init() { RegisterSystemTaskHandler(pricingOfferSyncHandler{}) }

func normalizedUpstreamTokenCosts(item upstreamPriceItem) (float64, float64, float64, error) {
	if item.QuotaType == 1 {
		return 0, 0, 0, errors.New("fixed-price model is not token-priced")
	}
	if item.BillingMode == billing_setting.BillingModeTieredExpr {
		if strings.TrimSpace(item.BillingExpr) == "" {
			return 0, 0, 0, errors.New("tiered pricing expression is empty")
		}
		input, _, err := billingexpr.RunExpr(item.BillingExpr, billingexpr.TokenParams{P: 1, Len: 1})
		if err != nil {
			return 0, 0, 0, fmt.Errorf("input price expression failed: %w", err)
		}
		output, _, err := billingexpr.RunExpr(item.BillingExpr, billingexpr.TokenParams{C: 1})
		if err != nil {
			return 0, 0, 0, fmt.Errorf("output price expression failed: %w", err)
		}
		cache, _, err := billingexpr.RunExpr(item.BillingExpr, billingexpr.TokenParams{CR: 1, Len: 1})
		if err != nil {
			return 0, 0, 0, fmt.Errorf("cache price expression failed: %w", err)
		}
		if input <= 0 || output <= 0 || !validPriceNumber(input) || !validPriceNumber(output) || !validPriceNumber(cache) {
			return 0, 0, 0, errors.New("pricing expression did not produce positive token costs")
		}
		return input, output, cache, nil
	}
	completionRatio := item.CompletionRatio
	if completionRatio <= 0 {
		completionRatio = 1
	}
	input := item.ModelRatio * 2
	output := input * completionRatio
	cache := 0.0
	if item.CacheRatio != nil {
		cache = input * *item.CacheRatio
	}
	if input <= 0 || output <= 0 || !validPriceNumber(input) || !validPriceNumber(output) || !validPriceNumber(cache) {
		return 0, 0, 0, errors.New("ratio pricing did not produce positive token costs")
	}
	return input, output, cache, nil
}
