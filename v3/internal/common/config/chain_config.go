package config

import (
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/gobitfly/beaconchain-backend/internal/log"
	"github.com/spf13/viper"
)

// ChainConfig
// This should be kept as simple and high-level as possible.
type ChainConfig struct {
	ID               uint64 `mapstructure:"CHAIN_ID"`
	GenesisTimestamp int    `mapstructure:"GENESIS_TIMESTAMP"`
	SecondsPerSlot   int    `mapstructure:"SECONDS_PER_SLOT"`
	SlotsPerEpoch    int    `mapstructure:"SLOTS_PER_EPOCH"`
}

type ChainConfigs map[domain.Chain]ChainConfig

func LoadChainConfigs() ChainConfigs {
	return ChainConfigs{
		domain.ChainMainnet: LoadChainConfig("mainnet"),
		domain.ChainHoodi:   LoadChainConfig("hoodi"),
	}
}

func LoadChainConfig(name string) ChainConfig {
	viper := viper.New()
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
	viper.SetConfigName(name)
	err = viper.MergeInConfig()
	if err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	// Optionally read from environment variables (e.g., override with ENV vars)
	viper.AutomaticEnv()

	logDebugConfigKeys()

	var chain ChainConfig
	err = viper.UnmarshalKey("ChainSpec", &chain)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	return chain
}
