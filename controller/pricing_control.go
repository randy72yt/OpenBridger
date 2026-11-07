package controller

import (
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
)

func ImportPricingOffers(c *gin.Context) {
	var request struct {
		Offers []model.UpstreamModelOffer `json:"offers"`
	}
	if err := common.DecodeJson(c.Request.Body, &request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	if err := service.ValidateAndImportPricingOffers(request.Offers); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	recordManageAudit(c, "pricing.offers.import", map[string]interface{}{"count": len(request.Offers)})
	common.ApiSuccess(c, gin.H{"imported": len(request.Offers)})
}

func SyncPricingOffers(c *gin.Context) {
	var request struct {
		ChannelIDs []int `json:"channel_ids"`
	}
	if c.Request.ContentLength > 0 {
		if err := common.DecodeJson(c.Request.Body, &request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
			return
		}
	}
	result, err := service.RunPricingOfferSync(c.Request.Context(), request.ChannelIDs, "manual")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error(), "data": result})
		return
	}
	recordManageAudit(c, "pricing.offers.sync", map[string]interface{}{"channel_ids": request.ChannelIDs, "imported": result.Imported})
	common.ApiSuccess(c, result)
}

func GetPricingOffers(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	channelID, _ := strconv.Atoi(c.Query("channel_id"))
	page, err := model.ListUpstreamModelOffersPage(model.UpstreamModelOfferQuery{
		PublicModel:    c.Query("model"),
		Keyword:        c.Query("q"),
		SourceType:     c.Query("source_type"),
		ChannelID:      channelID,
		IncludeExpired: c.Query("include_expired") == "true",
		Limit:          limit,
		Offset:         offset,
	})
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, page)
}

func UpsertPricingPolicy(c *gin.Context) {
	var policy model.ModelPricePolicy
	if err := common.DecodeJson(c.Request.Body, &policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	if err := service.ValidateAndUpsertPricePolicy(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	recordManageAudit(c, "pricing.policy.update", map[string]interface{}{"model": policy.PublicModel, "tier": policy.ServiceTier})
	common.ApiSuccess(c, policy)
}

func GetPricingPolicies(c *gin.Context) {
	policies, err := model.ListModelPricePolicies()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, policies)
}

func RecalculatePricing(c *gin.Context) {
	task, created, err := service.EnqueueSystemTask(model.SystemTaskTypePricingRecalculate, map[string]any{"manual": true})
	if err != nil {
		common.ApiError(c, err)
		return
	}
	recordManageAudit(c, "pricing.recalculate", map[string]interface{}{"task_id": task.TaskID, "created": created})
	common.ApiSuccess(c, gin.H{"task": task.ToResponse(), "created": created})
}

func GetPricingProposals(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	reviews, total, err := service.ListPricingProposalReviewsPage(c.Query("status"), limit, offset)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"items": reviews, "total": total, "limit": limit, "offset": offset})
}

func ApprovePricingProposal(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.ApiErrorMsg(c, "invalid proposal id")
		return
	}
	if err := service.ApprovePricingProposal(id, c.GetInt("id")); err != nil {
		common.ApiError(c, err)
		return
	}
	recordManageAudit(c, "pricing.proposal.approve", map[string]interface{}{"id": id})
	common.ApiSuccess(c, nil)
}

func RejectPricingProposal(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.ApiErrorMsg(c, "invalid proposal id")
		return
	}
	if err := service.RejectPricingProposal(id); err != nil {
		common.ApiError(c, err)
		return
	}
	recordManageAudit(c, "pricing.proposal.reject", map[string]interface{}{"id": id})
	common.ApiSuccess(c, nil)
}

func PublishPricingProposal(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.ApiErrorMsg(c, "invalid proposal id")
		return
	}
	if err := service.PublishPricingProposal(id); err != nil {
		common.ApiError(c, err)
		return
	}
	recordManageAudit(c, "pricing.proposal.publish", map[string]interface{}{"id": id})
	common.ApiSuccess(c, nil)
}

func BatchUpdatePricingProposals(c *gin.Context) {
	var request struct {
		Action string  `json:"action"`
		IDs    []int64 `json:"ids"`
	}
	if err := common.DecodeJson(c.Request.Body, &request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	if len(request.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ids is required"})
		return
	}
	var result *service.PricingProposalBatchResult
	switch request.Action {
	case "approve":
		result = service.BatchApprovePricingProposals(request.IDs, c.GetInt("id"))
	case "reject":
		result = service.BatchRejectPricingProposals(request.IDs)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "action must be approve or reject"})
		return
	}
	recordManageAudit(c, "pricing.proposal.batch_"+request.Action, map[string]interface{}{"ids": request.IDs, "succeeded": len(result.Succeeded), "failed": len(result.Failed)})
	common.ApiSuccess(c, result)
}

func DisablePricingPolicy(c *gin.Context) {
	var request struct {
		PublicModel string `json:"public_model"`
		ServiceTier string `json:"service_tier"`
	}
	if err := common.DecodeJson(c.Request.Body, &request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	if request.ServiceTier == "" {
		request.ServiceTier = "default"
	}
	disabled, err := service.DisablePricingPolicyByModel(request.PublicModel, request.ServiceTier)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	recordManageAudit(c, "pricing.policy.disable", map[string]interface{}{"model": request.PublicModel, "tier": request.ServiceTier})
	common.ApiSuccess(c, gin.H{"disabled": disabled})
}

func GetPricingRisks(c *gin.Context) {
	risks, err := service.PricingRisks()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, risks)
}

func GetPricingSyncRuns(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	runs, err := model.ListPricingSyncRuns(limit)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	latest, err := model.LatestPricingSyncRun()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"items": runs, "latest": latest})
}

func GetPricingCalibrations(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	channelID, _ := strconv.Atoi(c.Query("channel_id"))
	rows, total, err := model.ListUpstreamCostCalibrationsPage(c.Query("q"), channelID, limit, offset)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	summary, err := service.PricingCalibrationSummary()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"items": rows, "total": total, "limit": limit, "offset": offset, "summary": summary})
}

func SyncPricingCalibrations(c *gin.Context) {
	var request struct {
		ChannelIDs []int `json:"channel_ids"`
	}
	if c.Request.ContentLength > 0 {
		if err := common.DecodeJson(c.Request.Body, &request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
			return
		}
	}
	result, err := service.RunPricingCalibration(c.Request.Context(), request.ChannelIDs, "manual")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error(), "data": result})
		return
	}
	recordManageAudit(c, "pricing.calibration.sync", map[string]interface{}{
		"channel_ids": request.ChannelIDs, "applied": result.Applied, "skipped": result.Skipped,
	})
	common.ApiSuccess(c, result)
}

// PinPricingCalibration lets an operator store a ratio measured by hand, which
// is the only option for a model with too little traffic to calibrate.
func PinPricingCalibration(c *gin.Context) {
	var request struct {
		ChannelID   int     `json:"channel_id"`
		PublicModel string  `json:"public_model"`
		Ratio       float64 `json:"effective_ratio"`
		Reason      string  `json:"reason"`
	}
	if err := common.DecodeJson(c.Request.Body, &request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	if request.Ratio <= 0 {
		if err := service.ClearPinnedUpstreamCostCalibration(request.ChannelID, request.PublicModel); err != nil {
			common.ApiError(c, err)
			return
		}
		recordManageAudit(c, "pricing.calibration.clear", map[string]interface{}{
			"channel_id": request.ChannelID, "model": request.PublicModel,
		})
		common.ApiSuccess(c, nil)
		return
	}
	if err := service.PinUpstreamCostCalibration(request.ChannelID, request.PublicModel, request.Ratio, request.Reason); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	recordManageAudit(c, "pricing.calibration.pin", map[string]interface{}{
		"channel_id": request.ChannelID, "model": request.PublicModel, "ratio": request.Ratio,
	})
	common.ApiSuccess(c, nil)
}
