package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegister{}, "axon/agent/MsgRegister", nil)
	cdc.RegisterConcrete(&MsgAddStake{}, "axon/agent/MsgAddStake", nil)
	cdc.RegisterConcrete(&MsgReduceStake{}, "axon/agent/MsgReduceStake", nil)
	cdc.RegisterConcrete(&MsgClaimReducedStake{}, "axon/agent/MsgClaimReducedStake", nil)
	cdc.RegisterConcrete(&MsgUpdateAgent{}, "axon/agent/MsgUpdateAgent", nil)
	cdc.RegisterConcrete(&MsgHeartbeat{}, "axon/agent/MsgHeartbeat", nil)
	cdc.RegisterConcrete(&MsgDeregister{}, "axon/agent/MsgDeregister", nil)
	cdc.RegisterConcrete(&MsgSubmitAIChallengeResponse{}, "axon/agent/MsgSubmitAIChallengeResponse", nil)
	cdc.RegisterConcrete(&MsgRevealAIChallengeResponse{}, "axon/agent/MsgRevealAIChallengeResponse", nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRegister{},
		&MsgAddStake{},
		&MsgReduceStake{},
		&MsgClaimReducedStake{},
		&MsgUpdateAgent{},
		&MsgHeartbeat{},
		&MsgDeregister{},
		&MsgSubmitAIChallengeResponse{},
		&MsgRevealAIChallengeResponse{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &Msg_ServiceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(types.NewInterfaceRegistry())
)
