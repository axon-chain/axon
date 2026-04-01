package app

import (
	"context"
	"encoding/json"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	feemarketkeeper "github.com/cosmos/evm/x/feemarket/keeper"

	agentkeeper "github.com/axon-chain/axon/x/agent/keeper"
	agenttypes "github.com/axon-chain/axon/x/agent/types"
)

var (
	_ module.AppModuleBasic = AgentAppModule{}
	_ module.HasABCIGenesis = AgentAppModule{}
	_ module.HasName        = AgentAppModule{}
)

type AgentAppModule struct {
	cdc             codec.Codec
	keeper          agentkeeper.Keeper
	bankKeeper      bankkeeper.Keeper
	feeMarketKeeper feemarketkeeper.Keeper
}

func NewAgentAppModule(cdc codec.Codec, keeper agentkeeper.Keeper, bankKeeper bankkeeper.Keeper, fmKeeper feemarketkeeper.Keeper) AgentAppModule {
	return AgentAppModule{cdc: cdc, keeper: keeper, bankKeeper: bankKeeper, feeMarketKeeper: fmKeeper}
}

func (am AgentAppModule) Name() string { return agenttypes.ModuleName }

func (am AgentAppModule) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	agenttypes.RegisterCodec(cdc)
}

func (am AgentAppModule) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	agenttypes.RegisterInterfaces(registry)
}

func (am AgentAppModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(agenttypes.DefaultGenesis())
}

func (am AgentAppModule) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var gs agenttypes.GenesisState
	if err := cdc.UnmarshalJSON(bz, &gs); err != nil {
		return err
	}
	return gs.Validate()
}

func (am AgentAppModule) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := agenttypes.RegisterQueryHandlerClient(context.Background(), mux, agenttypes.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

func (am AgentAppModule) RegisterServices(cfg module.Configurator) {
	agenttypes.RegisterMsgServer(cfg.MsgServer(), agentkeeper.NewMsgServerImpl(am.keeper))
	agenttypes.RegisterQueryServer(cfg.QueryServer(), agentkeeper.NewQueryServerImpl(am.keeper))
}

func (am AgentAppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var gs agenttypes.GenesisState
	cdc.MustUnmarshalJSON(data, &gs)
	if err := am.keeper.SetParams(ctx, gs.Params); err != nil {
		panic(fmt.Sprintf("invalid agent module params in genesis: %v", err))
	}
	for _, agent := range gs.Agents {
		am.keeper.SetAgent(ctx, agent)
		am.keeper.BootstrapLegacyReputation(ctx, agent.Address, agent.Reputation)
	}

	var extra genesisExtra
	if err := json.Unmarshal(data, &extra); err != nil {
		panic(fmt.Sprintf("failed to parse genesis extra fields: %v — this would reset mint caps and cause token over-issuance", err))
	}
	if extra.TotalBlockRewardsMinted != "" {
		v, ok := sdkmath.NewIntFromString(extra.TotalBlockRewardsMinted)
		if !ok {
			panic(fmt.Sprintf("failed to parse total_block_rewards_minted %q — refusing to start with reset mint cap", extra.TotalBlockRewardsMinted))
		}
		am.keeper.SetTotalBlockRewardsMinted(ctx, v)
	}
	if extra.TotalContributionMinted != "" {
		v, ok := sdkmath.NewIntFromString(extra.TotalContributionMinted)
		if !ok {
			panic(fmt.Sprintf("failed to parse total_contribution_minted %q — refusing to start with reset mint cap", extra.TotalContributionMinted))
		}
		am.keeper.SetTotalContributionMinted(ctx, v)
	}
	if extra.LastProcessedEpoch > 0 {
		am.keeper.SetLastProcessedEpoch(ctx, extra.LastProcessedEpoch)
	}
	if extra.RewardPool != "" {
		v, ok := sdkmath.NewIntFromString(extra.RewardPool)
		if ok && v.IsPositive() {
			am.keeper.SetRewardPool(ctx, sdk.NewCoin("aaxon", v))
		}
	}
	if extra.ContributionPool != "" {
		v, ok := sdkmath.NewIntFromString(extra.ContributionPool)
		if ok && v.IsPositive() {
			am.keeper.SetContributionPool(ctx, sdk.NewCoin("aaxon", v))
		}
	}
	if len(extra.ContractDeployers) > 0 {
		am.keeper.ImportContractDeployers(ctx, extra.ContractDeployers)
	}
	if len(extra.AIBonuses) > 0 {
		am.keeper.ImportAIBonuses(ctx, extra.AIBonuses)
	}
	if len(extra.Wallets) > 0 {
		am.keeper.ImportWalletData(ctx, extra.Wallets)
	}
	return nil
}

func (am AgentAppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	params := am.keeper.GetParams(ctx)
	agents := am.keeper.GetAllAgents(ctx)
	gs := agenttypes.GenesisState{
		Params: params,
		Agents: agents,
	}
	base := cdc.MustMarshalJSON(&gs)

	extra := genesisExtra{
		TotalBlockRewardsMinted: am.keeper.GetTotalBlockRewardsMinted(ctx).String(),
		TotalContributionMinted: am.keeper.GetTotalContributionMinted(ctx).String(),
		LastProcessedEpoch:      am.keeper.GetLastProcessedEpoch(ctx),
		RewardPool:              am.keeper.GetRewardPool(ctx).Amount.String(),
		ContributionPool:        am.keeper.GetContributionPool(ctx).Amount.String(),
		ContractDeployers:       am.keeper.ExportContractDeployers(ctx),
		AIBonuses:               am.keeper.ExportAIBonuses(ctx),
		Wallets:                 am.keeper.ExportWalletData(ctx),
	}
	extraBz, _ := json.Marshal(extra)

	merged := mergeJSON(base, extraBz)
	return merged
}

func (am AgentAppModule) ConsensusVersion() uint64 { return 1 }

func (am AgentAppModule) BeginBlock(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	BurnCollectedFees(sdkCtx, am.bankKeeper, am.feeMarketKeeper)
	am.keeper.BeginBlocker(sdkCtx)
	return nil
}

func (am AgentAppModule) EndBlock(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	am.keeper.EndBlocker(sdkCtx)
	return nil
}

func (am AgentAppModule) IsOnePerModuleType() {}
func (am AgentAppModule) IsAppModule()        {}

type genesisExtra struct {
	TotalBlockRewardsMinted string            `json:"total_block_rewards_minted,omitempty"`
	TotalContributionMinted string            `json:"total_contribution_minted,omitempty"`
	LastProcessedEpoch      uint64            `json:"last_processed_epoch,omitempty"`
	RewardPool              string            `json:"reward_pool,omitempty"`
	ContributionPool        string            `json:"contribution_pool,omitempty"`
	ContractDeployers       map[string]string `json:"contract_deployers,omitempty"`
	AIBonuses               map[string]int64  `json:"ai_bonuses,omitempty"`
	Wallets                 map[string]string `json:"wallets,omitempty"`
}

func mergeJSON(base, overlay json.RawMessage) json.RawMessage {
	var bMap, oMap map[string]json.RawMessage
	if err := json.Unmarshal(base, &bMap); err != nil {
		return base
	}
	if err := json.Unmarshal(overlay, &oMap); err != nil {
		return base
	}
	for k, v := range oMap {
		bMap[k] = v
	}
	merged, err := json.Marshal(bMap)
	if err != nil {
		return base
	}
	return merged
}
