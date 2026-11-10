package perfmetrics

import (
	"net/http"
	"strings"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/QuantumNous/new-api/setting/perf_metrics_setting"
)

// Outcome classifies a relay result for the availability caliber: successes and
// server failures count towards public availability, client-excluded results do
// not (they are admin-visible only).
type Outcome int

const (
	OutcomeSuccess Outcome = iota
	OutcomeServerFailure
	OutcomeClientExcluded
)

// Sample sources. Public numbers are built from real samples only (plus probe
// once public_v2 is enabled); internal and client_error samples land in
// perf_metrics_aux for the admin view.
const (
	SourceReal        = "real"
	SourceProbe       = "probe"
	SourceInternal    = "internal"
	SourceClientError = "client_error"
)

// excludedErrorCodes are request-side problems that do not indicate model
// availability: bad client parameters, insufficient quota, auth failures and
// content moderation.
var excludedErrorCodes = map[types.ErrorCode]struct{}{
	types.ErrorCodeInvalidRequest:            {},
	types.ErrorCodeBadRequestBody:            {},
	types.ErrorCodeReadRequestBodyFailed:     {},
	types.ErrorCodeConvertRequestFailed:      {},
	types.ErrorCodeJsonMarshalFailed:         {},
	types.ErrorCodeCountTokenFailed:          {},
	types.ErrorCodeInsufficientUserQuota:     {},
	types.ErrorCodePreConsumeTokenQuotaFailed: {},
	types.ErrorCodeAccessDenied:              {},
	types.ErrorCodeSensitiveWordsDetected:    {},
	types.ErrorCodePromptBlocked:             {},
	// A missing built-in price is our configuration problem, but surfacing it in
	// public availability would mark sellable models as down; it stays
	// admin-only together with the other excluded sources.
	types.ErrorCodeModelPriceError: {},
}

// serverFailureErrorCodes are gateway/upstream problems we own: transport
// failures, channel misconfiguration and routing failures.
var serverFailureErrorCodes = map[types.ErrorCode]struct{}{
	types.ErrorCodeDoRequestFailed:            {},
	types.ErrorCodeChannelResponseTimeExceeded: {},
	types.ErrorCodeReadResponseBodyFailed:     {},
	types.ErrorCodeEmptyResponse:              {},
	types.ErrorCodeBadResponse:                {},
	types.ErrorCodeBadResponseBody:            {},
	types.ErrorCodeGetChannelFailed:           {},
	types.ErrorCodeChannelNoAvailableKey:      {},
	types.ErrorCodeChannelInvalidKey:          {},
	types.ErrorCodeChannelModelMappedError:    {},
	types.ErrorCodeChannelAwsClientError:      {},
	types.ErrorCodeModelNotFound:              {},
}

// upstreamErrorTypes mark errors whose status code came from the upstream
// provider rather than the gateway itself.
var upstreamErrorTypes = map[types.ErrorType]struct{}{
	types.ErrorTypeOpenAIError:   {},
	types.ErrorTypeClaudeError:   {},
	types.ErrorTypeGeminiError:   {},
	types.ErrorTypeUpstreamError: {},
}

// ClassifyRelayError maps a failed relay to the availability caliber per the
// failure classification table (plan section 3). Unknown errors default to
// server failure so unclassified outages stay visible in the numbers.
func ClassifyRelayError(err *types.NewAPIError) Outcome {
	if err == nil {
		return OutcomeSuccess
	}

	code := err.GetErrorCode()
	if _, ok := excludedErrorCodes[code]; ok {
		return OutcomeClientExcluded
	}
	if strings.HasPrefix(string(code), "violation_fee.") {
		return OutcomeClientExcluded
	}
	if _, ok := serverFailureErrorCodes[code]; ok {
		return OutcomeServerFailure
	}

	statusCode := err.StatusCode
	errorType := err.GetErrorType()
	if statusCode >= 500 {
		return OutcomeServerFailure
	}
	if statusCode == http.StatusTooManyRequests && errorType != types.ErrorTypeNewAPIError {
		return OutcomeServerFailure
	}
	if _, upstream := upstreamErrorTypes[errorType]; upstream {
		switch statusCode {
		case http.StatusBadRequest, http.StatusUnprocessableEntity:
			return OutcomeClientExcluded
		case http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
			return OutcomeServerFailure
		}
	}
	if errorType == types.ErrorTypeNewAPIError &&
		(statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden) {
		return OutcomeClientExcluded
	}
	return OutcomeServerFailure
}

// SourceForRelay decides which sample bucket a relay belongs to. The probe
// token wins over the internal list; playground traffic is always internal.
func SourceForRelay(info *relaycommon.RelayInfo, tokenName string) string {
	if info == nil {
		return SourceReal
	}
	cfg := perf_metrics_setting.GetSetting()
	if cfg.ProbeTokenName != "" && tokenName == cfg.ProbeTokenName {
		return SourceProbe
	}
	if info.IsPlayground {
		return SourceInternal
	}
	if _, ok := cfg.InternalTokenSet()[tokenName]; ok {
		return SourceInternal
	}
	return SourceReal
}
