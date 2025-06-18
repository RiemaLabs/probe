// APACHE NOTICE
// Sourced with modifications from https://github.com/strangelove-ventures/lens
package client

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	sdkTypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	authz "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

var (
	// Provides a default set of AppModuleBasics that are included in the ChainClientConfig
	// This is used to provide a default set of modules that will be used for protobuf registration and in-app decoding of RPC responses
	DefaultModuleBasics = []module.AppModuleBasic{
		auth.AppModuleBasic{},
		authz.AppModuleBasic{},
		bank.AppModuleBasic{},
		gov.AppModuleBasic{},
		crisis.AppModuleBasic{},
		distribution.AppModuleBasic{},
		mint.AppModuleBasic{},
		params.AppModuleBasic{},
		slashing.AppModuleBasic{},
		staking.AppModuleBasic{},
		vesting.AppModuleBasic{},
		wasm.AppModuleBasic{},
	}
)

type ChainClientConfig struct {
	Key                   string                  `json:"key" yaml:"key"`
	ChainID               string                  `json:"chain-id" yaml:"chain-id"`
	RPCAddr               string                  `json:"rpc-addr" yaml:"rpc-addr"`
	AccountPrefix         string                  `json:"account-prefix" yaml:"account-prefix"`
	KeyringBackend        string                  `json:"keyring-backend" yaml:"keyring-backend"`
	KeyDirectory          string                  `json:"key-directory" yaml:"key-directory"`
	Debug                 bool                    `json:"debug" yaml:"debug"`
	Timeout               string                  `json:"timeout" yaml:"timeout"`
	OutputFormat          string                  `json:"output-format" yaml:"output-format"`
	Modules               []module.AppModuleBasic `json:"-" yaml:"-"`
	CustomMsgTypeRegistry map[string]sdkTypes.Msg `json:"-" yaml:"-"`
}
