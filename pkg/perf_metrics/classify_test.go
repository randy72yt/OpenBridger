package perfmetrics

import (
	"errors"
	"testing"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/stretchr/testify/assert"
)

func TestClassifyRelayError(t *testing.T) {
	upstream := func(statusCode int) *types.NewAPIError {
		return types.WithOpenAIError(types.OpenAIError{Message: "upstream said no"}, statusCode)
	}
	local := func(code types.ErrorCode, statusCode int) *types.NewAPIError {
		return types.NewError(errors.New("local error"), code, types.ErrOptionWithStatusCode(statusCode))
	}

	cases := []struct {
		name string
		err  *types.NewAPIError
		want Outcome
	}{
		{"nil error is success", nil, OutcomeSuccess},
		// status-code driven
		{"upstream 500", upstream(500), OutcomeServerFailure},
		{"upstream 503", upstream(503), OutcomeServerFailure},
		{"upstream 429 is upstream rate limit", upstream(429), OutcomeServerFailure},
		{"local 429 stays conservative failure", local("", 429), OutcomeServerFailure},
		{"upstream 400 is client error", upstream(400), OutcomeClientExcluded},
		{"upstream 422 is client error", upstream(422), OutcomeClientExcluded},
		{"upstream 401 is our key problem", upstream(401), OutcomeServerFailure},
		{"upstream 403 is our key problem", upstream(403), OutcomeServerFailure},
		{"upstream 404 is our mapping problem", upstream(404), OutcomeServerFailure},
		{
			"claude upstream 400 is client error",
			types.WithClaudeError(types.ClaudeError{Type: "invalid_request_error"}, 400),
			OutcomeClientExcluded,
		},
		{"local 401 is auth", local("", 401), OutcomeClientExcluded},
		{"local 403 is auth", local("", 403), OutcomeClientExcluded},
		{"unknown code and status defaults to failure", local("some_new_code", 200), OutcomeServerFailure},
		// excluded codes (status forced to 200 so only the code path decides)
		{"invalid_request", local(types.ErrorCodeInvalidRequest, 200), OutcomeClientExcluded},
		{"bad_request_body", local(types.ErrorCodeBadRequestBody, 200), OutcomeClientExcluded},
		{"read_request_body_failed", local(types.ErrorCodeReadRequestBodyFailed, 200), OutcomeClientExcluded},
		{"convert_request_failed", local(types.ErrorCodeConvertRequestFailed, 200), OutcomeClientExcluded},
		{"json_marshal_failed", local(types.ErrorCodeJsonMarshalFailed, 200), OutcomeClientExcluded},
		{"count_token_failed", local(types.ErrorCodeCountTokenFailed, 200), OutcomeClientExcluded},
		{"insufficient_user_quota", local(types.ErrorCodeInsufficientUserQuota, 200), OutcomeClientExcluded},
		{"pre_consume_token_quota_failed", local(types.ErrorCodePreConsumeTokenQuotaFailed, 200), OutcomeClientExcluded},
		{"access_denied", local(types.ErrorCodeAccessDenied, 200), OutcomeClientExcluded},
		{"sensitive_words_detected", local(types.ErrorCodeSensitiveWordsDetected, 200), OutcomeClientExcluded},
		{"prompt_blocked", local(types.ErrorCodePromptBlocked, 200), OutcomeClientExcluded},
		{"violation_fee prefix", local(types.ErrorCodeViolationFeeGrokCSAM, 200), OutcomeClientExcluded},
		// model_price_error is our config problem but must stay out of public
		// availability (admin-only caliber)
		{"model_price_error excluded from public", local(types.ErrorCodeModelPriceError, 200), OutcomeClientExcluded},
		// server-failure codes (status forced to 200 so only the code path decides)
		{"do_request_failed", local(types.ErrorCodeDoRequestFailed, 200), OutcomeServerFailure},
		{"channel:response_time_exceeded", local(types.ErrorCodeChannelResponseTimeExceeded, 200), OutcomeServerFailure},
		{"read_response_body_failed", local(types.ErrorCodeReadResponseBodyFailed, 200), OutcomeServerFailure},
		{"empty_response", local(types.ErrorCodeEmptyResponse, 200), OutcomeServerFailure},
		{"bad_response", local(types.ErrorCodeBadResponse, 200), OutcomeServerFailure},
		{"bad_response_body", local(types.ErrorCodeBadResponseBody, 200), OutcomeServerFailure},
		{"get_channel_failed", local(types.ErrorCodeGetChannelFailed, 200), OutcomeServerFailure},
		{"channel:no_available_key", local(types.ErrorCodeChannelNoAvailableKey, 200), OutcomeServerFailure},
		{"channel:invalid_key", local(types.ErrorCodeChannelInvalidKey, 200), OutcomeServerFailure},
		{"channel:model_mapped_error", local(types.ErrorCodeChannelModelMappedError, 200), OutcomeServerFailure},
		{"channel:aws_client_error", local(types.ErrorCodeChannelAwsClientError, 200), OutcomeServerFailure},
		{"model_not_found", local(types.ErrorCodeModelNotFound, 200), OutcomeServerFailure},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, ClassifyRelayError(tc.err))
		})
	}
}

func TestSourceForRelay(t *testing.T) {
	cases := []struct {
		name       string
		playground bool
		tokenName  string
		adminOwned bool
		want       string
	}{
		{"admin-owned probe token wins over internal list", false, "ob-probe", true, SourceProbe},
		{"admin-owned internal token by name", false, "yuntao-main", true, SourceInternal},
		{"regular user token", false, "some-user-token", false, SourceReal},
		{"regular user token named ob-probe stays real", false, "ob-probe", false, SourceReal},
		{"regular user token named yuntao-main stays real", false, "yuntao-main", false, SourceReal},
		{"empty token name", false, "", false, SourceReal},
		{"empty token name admin-owned", false, "", true, SourceReal},
		{"playground is always internal", true, "some-user-token", false, SourceInternal},
		{"playground with admin probe token still probe", true, "ob-probe", true, SourceProbe},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			relayInfo := &relaycommon.RelayInfo{IsPlayground: tc.playground}
			token := relayTokenInfo{name: tc.tokenName, adminOwned: tc.adminOwned}
			assert.Equal(t, tc.want, SourceForRelay(relayInfo, token))
		})
	}
}
