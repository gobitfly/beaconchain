package config

import (
	"github.com/gobitfly/beaconchain-api/internal/log"
	"github.com/spf13/viper"
)

// This should be kept as simple and high-level as possible.
type Chain struct {
	Name      ChainName
	NetworkId uint64
	ChainId   uint64
}

// Config for defining various chain names
type ChainName string

const (
	Mainnet  ChainName = "mainnet"
	Holesky  ChainName = "holesky"
	Optimism ChainName = "optimism"
)

func LoadChainConfig(chain ChainName) {
	// "configs/chain/default.yaml"
	viper.AddConfigPath("configs/chain")
	viper.SetConfigName("default")
	viper.SetConfigType("yaml")

	// Read the default config file
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("Error reading config file, %s", err)
	}

	// Now load in the override config file. It replaces anything which exists in both
	viper.SetConfigName(string(chain))
	err = viper.MergeInConfig()
	if err != nil {
		log.Fatal("Error reading config file, %s", err)
	}

	// Optionally read from environment variables (e.g., override with ENV vars)
	viper.AutomaticEnv()

	logDebugConfigKeys()
}
