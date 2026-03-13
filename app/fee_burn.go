package app

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	feemarketkeeper "github.com/cosmos/evm/x/feemarket/keeper"
)

// BurnCollectedFees implements whitepaper §8.5 / §8.6:
//   - Base Fee  → 100% burned (deflationary)
//   - Priority Fee → left in FeeCollector for x/distribution → proposer
//
// This MUST run before x/distribution's BeginBlocker so that distribution
// only distributes the remaining priority fees to the proposer/validators.
//
// When EIP-1559 is active (NoBaseFee=false), base fee dominates gas costs
// (typically 70-95% of effectiveGasPrice). Without per-tx tracking we use
// 80% as a conservative estimate for the base-fee share. When NoBaseFee=true
// (development mode), we burn 50% as a rough approximation.
func BurnCollectedFees(ctx sdk.Context, bankKeeper bankkeeper.Keeper, fmKeeper feemarketkeeper.Keeper) {
	feeCollectorAddr := authtypes.NewModuleAddress(authtypes.FeeCollectorName)
	balance := bankKeeper.GetBalance(ctx, feeCollectorAddr, "aaxon")

	if !balance.IsPositive() {
		return
	}

	fmParams := fmKeeper.GetParams(ctx)

	var burnPercent int64
	if fmParams.NoBaseFee {
		burnPercent = 50
	} else {
		baseFee := fmParams.BaseFee
		minGasPrice := fmParams.MinGasPrice
		if baseFee.IsPositive() && minGasPrice.IsPositive() {
			// Dynamic ratio: baseFee / minGasPrice, capped at [50, 95]
			ratio := baseFee.MulInt64(100).Quo(minGasPrice).TruncateInt64()
			if ratio < 50 {
				ratio = 50
			}
			if ratio > 95 {
				ratio = 95
			}
			burnPercent = ratio
		} else {
			burnPercent = 80
		}
	}

	burnAmount := balance.Amount.MulRaw(burnPercent).QuoRaw(100)
	if burnAmount.IsZero() {
		return
	}

	burnCoin := sdk.NewCoin("aaxon", burnAmount)
	if err := bankKeeper.BurnCoins(ctx, authtypes.FeeCollectorName, sdk.NewCoins(burnCoin)); err != nil {
		ctx.Logger().Error("failed to burn gas fees", "error", err)
		return
	}

	remaining := balance.Sub(burnCoin)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"gas_fee_burned",
		sdk.NewAttribute("burned", burnCoin.String()),
		sdk.NewAttribute("remaining_for_proposer", remaining.String()),
		sdk.NewAttribute("burn_percent", fmt.Sprintf("%d", burnPercent)),
		sdk.NewAttribute("block", fmt.Sprintf("%d", ctx.BlockHeight())),
	))
}
