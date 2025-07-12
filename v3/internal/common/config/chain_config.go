package config

import (
	"github.com/gobitfly/beaconchain-api/internal/log"
	"github.com/spf13/viper"
)

// Chain
// This should be kept as simple and high-level as possible.
type Chain struct {
	Name    ChainName
	ChainId uint64
}

// ChainName
// Config for defining various chain names
type ChainName string

const (
	Mainnet  ChainName = "mainnet"
	Holesky  ChainName = "holesky"
	Optimism ChainName = "optimism"
)

type ChainConfig struct {
}

func (chain Chain) LoadChainConfig() {
	// "configs/chain/default.yaml"
	viper.AddConfigPath("configs/chain")
	viper.SetConfigName("default")
	viper.SetConfigType("yaml")

	// Read the default config file
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	// Now load in the override config file. It replaces anything which exists in both
	viper.SetConfigName(string(chain.Name))
	err = viper.MergeInConfig()
	if err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	// Optionally read from environment variables (e.g., override with ENV vars)
	viper.AutomaticEnv()

	logDebugConfigKeys()
}
