package config

import (
	clienthelpers "cosmossdk.io/client/v2/helpers"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	cosmosevmserverconfig "github.com/cosmos/evm/server/config"
)

const (
	AxonDenom  = "aaxon" // atto-AXON (18 decimals)
	HumanDenom = "AXON"
	EVMChainID = uint64(8210) // Axon default EVM chain ID
)

func MustGetDefaultNodeHome() string {
	defaultNodeHome, err := clienthelpers.GetNodeHomeDirectory(".axond")
	if err != nil {
		panic(err)
	}
	return defaultNodeHome
}

func InitAppConfig() (string, interface{}) {
	srvCfg := serverconfig.DefaultConfig()
	srvCfg.MinGasPrices = "1000000000" + AxonDenom

	evmCfg := cosmosevmserverconfig.DefaultEVMConfig()
	evmCfg.EVMChainID = EVMChainID

	customAppConfig := EVMAppConfig{
		Config:  *srvCfg,
		EVM:     *evmCfg,
		JSONRPC: *cosmosevmserverconfig.DefaultJSONRPCConfig(),
		TLS:     *cosmosevmserverconfig.DefaultTLSConfig(),
	}

	return EVMAppTemplate, customAppConfig
}

type EVMAppConfig struct {
	serverconfig.Config

	EVM     cosmosevmserverconfig.EVMConfig
	JSONRPC cosmosevmserverconfig.JSONRPCConfig
	TLS     cosmosevmserverconfig.TLSConfig
}

const EVMAppTemplate = serverconfig.DefaultConfigTemplate + cosmosevmserverconfig.DefaultEVMConfigTemplate
