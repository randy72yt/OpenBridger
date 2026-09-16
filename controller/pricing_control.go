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

func GetPricingOffers(c *gin.Context) {
	offers, err := model.ListUpstreamModelOffers(c.Query("model"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, offers)
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
	proposals, err := service.ListPricingProposalReviews(c.Query("status"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, proposals)
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

func GetPricingRisks(c *gin.Context) {
	risks, err := service.PricingRisks()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, risks)
}
