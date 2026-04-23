package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axon-chain/axon/x/agent/types"
)

func (k Keeper) HandleCallService(ctx sdk.Context, sender, serviceId string, inputData []byte, payment sdk.Coin) ([]byte, error) {
	return k.CallService(ctx, sender, serviceId, inputData, payment)
}

func (k Keeper) HandleCreateTask(ctx sdk.Context, requester, title, description string, caps []string, budget sdk.Coin, deadline int64) (string, error) {
	return k.CreateTask(ctx, requester, title, description, caps, budget, deadline)
}

func (k Keeper) HandleCancelTask(ctx sdk.Context, requester, taskId string) error {
	return k.CancelTask(ctx, requester, taskId)
}

func (k Keeper) HandleCompleteTask(ctx sdk.Context, agentAddr, taskId, completionData string) error {
	return k.CompleteTask(ctx, agentAddr, taskId, completionData)
}

func (k Keeper) HandleCallTool(ctx sdk.Context, caller, toolId string, inputData []byte, payment sdk.Coin) ([]byte, error) {
	return k.CallTool(ctx, caller, toolId, inputData, payment)
}

func (k Keeper) HandleRegisterTool(ctx sdk.Context, agentAddr, name, description, inputSchema, outputSchema string, price sdk.Coin, isPublic bool) (string, error) {
	return k.RegisterTool(ctx, agentAddr, name, description, inputSchema, outputSchema, price, isPublic)
}

func (k Keeper) HandleSubmitBid(ctx sdk.Context, agentAddr, taskId, proposal string, price sdk.Coin) error {
	return k.SubmitBid(ctx, agentAddr, taskId, proposal, price)
}

func (k Keeper) HandleSelectBid(ctx sdk.Context, requester, taskId, agentAddr string) error {
	return k.SelectBid(ctx, requester, taskId, agentAddr)
}

func (k Keeper) HandleRegisterService(ctx sdk.Context, agentAddr, name, description string, caps, inputTypes, outputTypes []string, price sdk.Coin, endpoint string) (string, error) {
	return k.RegisterService(ctx, agentAddr, name, description, caps, inputTypes, outputTypes, price, endpoint)
}

func (k Keeper) HandleUpdateService(ctx sdk.Context, agentAddr, serviceId, name, description string, price sdk.Coin, endpoint string) error {
	return k.UpdateService(ctx, agentAddr, serviceId, name, description, price, endpoint)
}

func (k Keeper) HandleDisableService(ctx sdk.Context, agentAddr, serviceId string) error {
	return k.DisableService(ctx, agentAddr, serviceId)
}

func (k Keeper) PayAgent(ctx sdk.Context, agentAddr string, amount sdk.Coin) error {
	addr, err := sdk.AccAddressFromBech32(agentAddr)
	if err != nil {
		return err
	}
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, sdk.NewCoins(amount))
}