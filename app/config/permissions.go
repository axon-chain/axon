package config

import (
	"maps"
	"sort"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	cosmosevmutils "github.com/cosmos/evm/utils"
	erc20types "github.com/cosmos/evm/x/erc20/types"
	feemarkettypes "github.com/cosmos/evm/x/feemarket/types"
	vmtypes "github.com/cosmos/evm/x/vm/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	corevm "github.com/ethereum/go-ethereum/core/vm"

	agenttypes "github.com/axon-chain/axon/x/agent/types"
	privacytypes "github.com/axon-chain/axon/x/privacy/types"
)

func BlockedAddresses() map[string]bool {
	blockedAddrs := make(map[string]bool)

	maccPerms := GetMaccPerms()
	accs := make([]string, 0, len(maccPerms))
	for acc := range maccPerms {
		accs = append(accs, acc)
	}
	sort.Strings(accs)

	for _, acc := range accs {
		blockedAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	axonPrecompiles := map[string]bool{
		"0x0000000000000000000000000000000000000801": true,
		"0x0000000000000000000000000000000000000802": true,
		"0x0000000000000000000000000000000000000803": true,
		"0x0000000000000000000000000000000000000807": true,
		"0x0000000000000000000000000000000000000810": true,
		"0x0000000000000000000000000000000000000811": true,
		"0x0000000000000000000000000000000000000812": true,
		"0x0000000000000000000000000000000000000813": true,
	}

	blockedPrecompilesHex := vmtypes.AvailableStaticPrecompiles
	for _, addr := range corevm.PrecompiledAddressesPrague {
		blockedPrecompilesHex = append(blockedPrecompilesHex, addr.Hex())
	}

	for _, precompile := range blockedPrecompilesHex {
		if axonPrecompiles[precompile] {
			continue
		}
		blockedAddrs[cosmosevmutils.Bech32StringFromHexAddress(precompile)] = true
	}

	return blockedAddrs
}

var maccPerms = map[string][]string{
	authtypes.FeeCollectorName:     {authtypes.Burner},
	distrtypes.ModuleName:          nil,
	transfertypes.ModuleName:       {authtypes.Minter, authtypes.Burner},
	minttypes.ModuleName:           {authtypes.Minter},
	stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
	stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
	govtypes.ModuleName:            {authtypes.Burner},

	// Cosmos EVM modules
	vmtypes.ModuleName:        {authtypes.Minter, authtypes.Burner},
	feemarkettypes.ModuleName: nil,
	erc20types.ModuleName:     {authtypes.Minter, authtypes.Burner},

	// Axon custom modules
	agenttypes.ModuleName:   {authtypes.Minter, authtypes.Burner},
	privacytypes.ModuleName: nil,
}

func GetMaccPerms() map[string][]string {
	return maps.Clone(maccPerms)
}
