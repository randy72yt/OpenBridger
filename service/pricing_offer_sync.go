package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/pkg/billingexpr"
	"github.com/QuantumNous/new-api/setting/billing_setting"
)

const pricingOfferSyncResponseLimit = 10 << 20

type PricingOfferSyncChannelResult struct {
	ChannelID int      `json:"channel_id"`
	Imported  int      `json:"imported"`
	Skipped   int      `json:"skipped"`
	Warnings  []string `json:"warnings,omitempty"`
	Error     string   `json:"error,omitempty"`
}

type PricingOfferSyncResult struct {
	Imported         int                             `json:"imported"`
	Skipped          int                             `json:"skipped"`
	CreatedProposals int                             `json:"created_proposals,omitempty"`
	Channels         []PricingOfferSyncChannelResult `json:"channels"`
}

type upstreamPricingResponse struct {
	Success        bool                `json:"success"`
	Message        string              `json:"message"`
	PricingVersion string              `json:"pricing_version"`
	GroupRatio     map[string]float64  `json:"group_ratio"`
	AutoGroups     []string            `json:"auto_groups"`
	Data           []upstreamPriceItem `json:"data"`
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
		if err := ValidateAndImportPricingOffers(offers); err != nil {
			channelResult.Error = err.Error()
			result.Channels = append(result.Channels, channelResult)
			continue
		}
		channelResult.Imported = len(offers)
		result.Imported += len(offers)
		result.Channels = append(result.Channels, channelResult)
	}
	if result.Imported == 0 {
		return result, errors.New("pricing sync did not import any offers")
	}
	return result, nil
}

func fetchPricingOffersForChannel(ctx context.Context, channel *model.Channel) ([]model.UpstreamModelOffer, []string, error) {
	baseURL, err := url.Parse(strings.TrimSpace(channel.GetBaseURL()))
	if err != nil || baseURL.Scheme == "" || baseURL.Host == "" {
		return nil, nil, errors.New("channel has no valid upstream base URL")
	}
	baseURL.Path = "/api/pricing"
	baseURL.RawQuery = ""
	baseURL.Fragment = ""

	key, _, keyErr := channel.GetNextEnabledKey()
	if keyErr != nil || strings.TrimSpace(key) == "" {
		return nil, nil, errors.New("channel has no enabled API key")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL.String(), nil)
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

	now := common.GetTimestamp()
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
		group := ""
		for _, candidate := range payload.AutoGroups {
			if common.StringsContains(item.EnableGroups, candidate) {
				group = candidate
				break
			}
		}
		groupRatio, ok := payload.GroupRatio[group]
		if !ok || groupRatio <= 0 || !validPriceNumber(groupRatio) {
			warnings = append(warnings, item.ModelName+": no positive auto-group ratio")
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
		ratio := groupRatio
		offers = append(offers, model.UpstreamModelOffer{
			ChannelID: channel.Id, UpstreamModel: item.ModelName, PublicModel: publicModel,
			InputCost: inputCost, OutputCost: outputCost, CacheReadCost: cacheReadCost,
			UpstreamGroupRatio: &ratio, Currency: "USD", SourceType: "new-api-pricing",
			SourceURL: baseURL.String(), SourceVersion: version, SuccessRateBPS: 10000,
			CollectedAt: now, ExpiresAt: now + 2*60*60, Enabled: true,
		})
	}
	return offers, warnings, nil
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
	result, err := SyncPricingOffersFromChannels(ctx, nil)
	if err != nil {
		failSystemTask(task, runnerID, err)
		return
	}
	result.CreatedProposals, err = RecalculatePricingProposals(ctx)
	if err != nil {
		failSystemTask(task, runnerID, err)
		return
	}
	if err := model.FinishSystemTask(task.TaskID, runnerID, model.SystemTaskStatusSucceeded, result, ""); err != nil {
		logSystemTaskLockError(ctx, task, err)
	}
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
