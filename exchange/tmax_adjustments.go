package exchange

import (
	"context"
	"github.com/prebid/prebid-server/config"
	"time"
)

type TmaxAdjustmentsPreprocessed struct {
	BidderNetworkLatencyBuffer     uint
	PBSResponsePreparationDuration uint
	BidderResponseDurationMin      uint

	IsEnforced bool
}

func ProcessTMaxAdjustments(adjustmentsConfig config.TmaxAdjustments) *TmaxAdjustmentsPreprocessed {
	if !adjustmentsConfig.Enabled {
		return nil
	}

	isEnforced := adjustmentsConfig.BidderResponseDurationMin != 0 &&
		(adjustmentsConfig.BidderNetworkLatencyBuffer != 0 || adjustmentsConfig.PBSResponsePreparationDuration != 0)

	tmax := &TmaxAdjustmentsPreprocessed{
		BidderNetworkLatencyBuffer:     adjustmentsConfig.BidderNetworkLatencyBuffer,
		PBSResponsePreparationDuration: adjustmentsConfig.PBSResponsePreparationDuration,
		BidderResponseDurationMin:      adjustmentsConfig.BidderResponseDurationMin,
		IsEnforced:                     isEnforced,
	}

	return tmax
}

type bidderTmaxContext interface {
	Deadline() (deadline time.Time, ok bool)
	RemainingDurationMS(deadline time.Time) int64
}
type bidderTmaxCtx struct{ ctx context.Context }

func (b *bidderTmaxCtx) RemainingDurationMS(deadline time.Time) int64 {
	return time.Until(deadline).Milliseconds()
}
func (b *bidderTmaxCtx) Deadline() (deadline time.Time, ok bool) {
	deadline, ok = b.ctx.Deadline()
	return
}

func getBidderTmax(ctx bidderTmaxContext, requestTmaxMS int64, tmaxAdjustments *TmaxAdjustmentsPreprocessed) int64 {
	if tmaxAdjustments.IsEnforced {
		if deadline, ok := ctx.Deadline(); ok {

			return ctx.RemainingDurationMS(deadline) - int64(tmaxAdjustments.BidderNetworkLatencyBuffer) - int64(tmaxAdjustments.PBSResponsePreparationDuration)
		}
	}
	return requestTmaxMS
}
