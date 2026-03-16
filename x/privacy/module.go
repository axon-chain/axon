package privacy

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/axon-chain/axon/x/privacy/keeper"
	"github.com/axon-chain/axon/x/privacy/types"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
	_ module.HasGenesis     = AppModule{}
)

type AppModuleBasic struct{}

func (AppModuleBasic) Name() string { return types.ModuleName }

func (AppModuleBasic) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {}

func (AppModuleBasic) RegisterInterfaces(_ codectypes.InterfaceRegistry) {}

func (AppModuleBasic) RegisterGRPCGatewayRoutes(_ client.Context, _ *runtime.ServeMux) {}

func (AppModuleBasic) DefaultGenesis(_ codec.JSONCodec) json.RawMessage {
	bz, _ := json.Marshal(GenesisState{Params: types.DefaultParams()})
	return bz
}

func (AppModuleBasic) ValidateGenesis(_ codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	if bz == nil {
		return nil
	}
	var gs GenesisState
	if err := json.Unmarshal(bz, &gs); err != nil {
		return err
	}
	return gs.Params.Validate()
}

type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

func NewAppModule(k keeper.Keeper) AppModule {
	return AppModule{keeper: k}
}

func (AppModule) RegisterServices(module.Configurator) {}

func (am AppModule) InitGenesis(ctx sdk.Context, _ codec.JSONCodec, bz json.RawMessage) {
	var gs GenesisState
	if bz != nil {
		if err := json.Unmarshal(bz, &gs); err == nil && gs.Params.Validate() == nil {
			am.keeper.SetParams(ctx, gs.Params)
			return
		}
	}
	am.keeper.SetParams(ctx, types.DefaultParams())
}

func (am AppModule) ExportGenesis(ctx sdk.Context, _ codec.JSONCodec) json.RawMessage {
	gs := GenesisState{Params: am.keeper.GetParams(ctx)}
	bz, _ := json.Marshal(gs)
	return bz
}

func (AppModule) ConsensusVersion() uint64 { return 1 }
func (AppModule) IsOnePerModuleType()      {}
func (AppModule) IsAppModule()             {}

type GenesisState struct {
	Params types.Params `json:"params"`
}
