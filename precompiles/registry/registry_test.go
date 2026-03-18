package registry

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/holiman/uint256"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestUseLegacyMutationSenderCompat(t *testing.T) {
	tests := []struct {
		name    string
		chainID string
		height  int64
		want    bool
	}{
		{name: "mainnet before cutoff", chainID: mainnetChainID, height: 18392, want: true},
		{name: "mainnet after cutoff", chainID: mainnetChainID, height: 18393, want: false},
		{name: "other chain before cutoff", chainID: "axon-local-1", height: 100, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := sdk.Context{}.WithChainID(tt.chainID).WithBlockHeight(tt.height)
			if got := useLegacyMutationSenderCompat(ctx); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestResolveMutationSenderPrefersImmediateCaller(t *testing.T) {
	caller := common.HexToAddress("0x1111111111111111111111111111111111111111")
	origin := common.HexToAddress("0x2222222222222222222222222222222222222222")
	contract := vm.NewContract(caller, address, uint256.NewInt(0), 0, nil)
	evm := &vm.EVM{TxContext: vm.TxContext{Origin: origin}}

	got := resolveMutationSender(contract, evm)
	want := sdk.AccAddress(caller.Bytes())
	if !got.Equals(want) {
		t.Fatalf("expected sender %s, got %s", want.String(), got.String())
	}
}

func TestResolveMutationSenderFallsBackToOrigin(t *testing.T) {
	origin := common.HexToAddress("0x2222222222222222222222222222222222222222")
	contract := vm.NewContract(common.Address{}, address, uint256.NewInt(0), 0, nil)
	evm := &vm.EVM{TxContext: vm.TxContext{Origin: origin}}

	got := resolveMutationSender(contract, evm)
	want := sdk.AccAddress(origin.Bytes())
	if !got.Equals(want) {
		t.Fatalf("expected sender %s, got %s", want.String(), got.String())
	}
}

func TestResolveMutationSenderReturnsEmptyAddressWithoutContext(t *testing.T) {
	contract := vm.NewContract(common.Address{}, address, uint256.NewInt(0), 0, nil)

	got := resolveMutationSender(contract, nil)
	if len(got) != 0 {
		t.Fatalf("expected empty sender, got %s", got.String())
	}
}

func TestResolveRegisterMutationSenderUsesLegacyBehaviorOnHistoricalMainnet(t *testing.T) {
	caller := common.HexToAddress("0x1111111111111111111111111111111111111111")
	origin := common.HexToAddress("0x2222222222222222222222222222222222222222")
	contract := vm.NewContract(caller, address, uint256.NewInt(0), 0, nil)
	evm := &vm.EVM{TxContext: vm.TxContext{Origin: origin}}
	ctx := sdk.Context{}.WithChainID(mainnetChainID).WithBlockHeight(18392)

	got := (Precompile{}).resolveRegisterMutationSender(ctx, evm, contract)
	want := sdk.AccAddress(origin.Bytes())
	if !got.Equals(want) {
		t.Fatalf("expected sender %s, got %s", want.String(), got.String())
	}
}

func TestResolveRegisterMutationSenderUsesNewBehaviorAfterCutoff(t *testing.T) {
	caller := common.HexToAddress("0x1111111111111111111111111111111111111111")
	origin := common.HexToAddress("0x2222222222222222222222222222222222222222")
	contract := vm.NewContract(caller, address, uint256.NewInt(0), 0, nil)
	evm := &vm.EVM{TxContext: vm.TxContext{Origin: origin}}
	ctx := sdk.Context{}.WithChainID(mainnetChainID).WithBlockHeight(18393)

	got := (Precompile{}).resolveRegisterMutationSender(ctx, evm, contract)
	want := sdk.AccAddress(caller.Bytes())
	if !got.Equals(want) {
		t.Fatalf("expected sender %s, got %s", want.String(), got.String())
	}
}
