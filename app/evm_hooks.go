package app

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	evmtypes "github.com/cosmos/evm/x/vm/types"

	agentkeeper "github.com/axon-chain/axon/x/agent/keeper"
)

const (
	// DeployBurnAmount is 10 AXON = 10 * 10^18 aaxon
	DeployBurnDenom = "aaxon"
	DeployBurnAxon  = 10
	DeployBurnExp   = 18
)

var _ evmtypes.EvmHooks = DeployBurnHook{}

// DeployBurnHook burns 10 AXON from the deployer when a contract is created,
// and tracks the deployment for contribution rewards.
//
// Limitation: only top-level deployments (receipt.ContractAddress) trigger the
// burn. Internal CREATE/CREATE2 by factory contracts bypass it. Mitigated by
// factory itself paying the burn + normal gas-fee burning. A future upgrade
// may add EVM tracing to catch internal creations.
type DeployBurnHook struct {
	bankKeeper  bankkeeper.Keeper
	agentKeeper agentkeeper.Keeper
}

func NewDeployBurnHook(bk bankkeeper.Keeper, ak agentkeeper.Keeper) DeployBurnHook {
	return DeployBurnHook{bankKeeper: bk, agentKeeper: ak}
}

func (h DeployBurnHook) PostTxProcessing(
	ctx sdk.Context,
	sender common.Address,
	msg core.Message,
	receipt *ethtypes.Receipt,
) error {
	if h.agentKeeper.IsV110UpgradeActivated(ctx) {
		h.agentKeeper.RecordEvidenceTxHash(ctx, receipt.TxHash)
	}

	senderAccAddr := sdk.AccAddress(sender.Bytes())
	isDeployment := receipt.ContractAddress != (common.Address{})

	if isDeployment {
		// 10 AXON = 10 * 10^18 aaxon
		burnAmount := sdkmath.NewInt(DeployBurnAxon).Mul(sdkmath.NewIntWithDecimal(1, DeployBurnExp))
		burnCoin := sdk.NewCoin(DeployBurnDenom, burnAmount)

		balance := h.bankKeeper.GetBalance(ctx, senderAccAddr, DeployBurnDenom)
		if balance.Amount.LT(burnAmount) {
			return fmt.Errorf("insufficient balance for contract deployment burn: need %s %s, have %s", burnAmount.String(), DeployBurnDenom, balance.Amount.String())
		}

		if err := h.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAccAddr, evmtypes.ModuleName, sdk.NewCoins(burnCoin)); err != nil {
			return err
		}
		if err := h.bankKeeper.BurnCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(burnCoin)); err != nil {
			return err
		}

		// Record deployer mapping for contribution reward attribution
		contractAddrStr := sdk.AccAddress(receipt.ContractAddress.Bytes()).String()
		h.agentKeeper.SetContractDeployer(ctx, contractAddrStr, senderAccAddr.String())

		if h.agentKeeper.IsAgent(ctx, senderAccAddr.String()) {
			h.agentKeeper.IncrementDeployCount(ctx, senderAccAddr.String())
		}

		ctx.EventManager().EmitEvent(sdk.NewEvent(
			"contract_deploy_burn",
			sdk.NewAttribute("deployer", senderAccAddr.String()),
			sdk.NewAttribute("contract", receipt.ContractAddress.Hex()),
			sdk.NewAttribute("burned", burnCoin.String()),
		))

		return nil
	}

	// Track contract calls for contribution rewards (whitepaper §8.4).
	// Credit goes to the contract deployer (not the contract address or caller).
	// Self-calls excluded: sender must differ from deployer.
	if msg.To != nil && len(receipt.Logs) > 0 {
		contractAddr := sdk.AccAddress(msg.To.Bytes()).String()
		deployer := h.agentKeeper.GetContractDeployer(ctx, contractAddr)
		if deployer != "" && deployer != senderAccAddr.String() && h.agentKeeper.IsAgent(ctx, deployer) {
			h.agentKeeper.IncrementContractCalls(ctx, deployer)
		}
	}

	return nil
}
