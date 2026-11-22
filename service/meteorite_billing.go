package service

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
)

// meteorite: the only change to how the gateway bills (D-063). It bills the tokens the upstream
// reports, minus the average tokens that upstream model injects into every prompt, which the
// verification backend measures and stores on the channel.
const MeteoriteMinChargedInputTokens = 10

// MeteoriteChargedInputTokens is the input tokens a request is charged for: the reported count
// less the injection, never below 10 tokens and never above the reported count.
func MeteoriteChargedInputTokens(reported, injection int) int {
	if injection <= 0 || reported <= 0 {
		return reported
	}
	charged := max(reported-injection, MeteoriteMinChargedInputTokens)
	return min(charged, reported)
}

// meteoriteAdjustUsage returns the usage to bill and a log line when the deduction changed it.
// The reported usage is not modified.
func meteoriteAdjustUsage(relayInfo *relaycommon.RelayInfo, usage *dto.Usage) (*dto.Usage, *common.Message) {
	if usage == nil || relayInfo == nil || relayInfo.ChannelMeta == nil {
		return usage, nil
	}
	injection := relayInfo.ChannelSetting.MeteoriteInjectionTokens
	charged := MeteoriteChargedInputTokens(usage.PromptTokens, injection)
	if charged == usage.PromptTokens {
		return usage, nil
	}
	adjusted := *usage
	adjusted.PromptTokens = charged
	adjusted.TotalTokens = charged + usage.CompletionTokens
	note := common.NewMessage("Input tokens reported {{reported}}, deduction {{deduction}}, charged {{charged}}", map[string]any{
		"reported":  usage.PromptTokens,
		"deduction": usage.PromptTokens - charged,
		"charged":   charged,
	})
	return &adjusted, note
}

// MeteoriteRelayErrorStatus maps a relay error to the HTTP status users see: a request refused
// for lack of balance is 402, not 403.
func MeteoriteRelayErrorStatus(err *types.NewAPIError) int {
	if err.GetErrorCode() == types.ErrorCodeInsufficientUserQuota {
		return 402
	}
	return err.StatusCode
}

// MeteoriteCoolDownOnUpstreamFailure rests the channel for 60 seconds when an upstream error
// (a transport failure, a timeout, HTTP 5xx, 408 or 429) failed the request. A request the user
// got wrong (other 4xx) or a local rejection does not count against the channel.
func MeteoriteCoolDownOnUpstreamFailure(channelId int, err *types.NewAPIError) {
	if err == nil || channelId <= 0 {
		return
	}
	failedUpstream := err.GetErrorType() != types.ErrorTypeNewAPIError &&
		(err.StatusCode >= 500 || err.StatusCode == 408 || err.StatusCode == 429)
	if types.IsChannelError(err) || err.GetErrorCode() == types.ErrorCodeDoRequestFailed || failedUpstream {
		model.MeteoriteCoolDownChannel(channelId)
	}
}
