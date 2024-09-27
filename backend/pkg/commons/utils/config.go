package utils

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"os"
	"slices"
	"strings"

	"github.com/ethereum/go-ethereum/params"
	"github.com/gobitfly/beaconchain/pkg/commons/config"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/consapi"
	"github.com/kelseyhightower/envconfig"

	//nolint:depguard
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var Config *types.Config

func readConfigFile(cfg *types.Config, path string) error {
	if path == "" {
		return yaml.Unmarshal([]byte(config.DefaultConfigYml), cfg)
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening config file %v: %v", path, err)
	}

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		return fmt.Errorf("error decoding config file %v: %v", path, err)
	}

	return nil
}

func readConfigEnv(cfg *types.Config) error {
	return envconfig.Process("", cfg)
}

func readConfigSecrets(cfg *types.Config) error {
	return ProcessSecrets(cfg)
}

func confSanityCheck(cfg *types.Config) {
	if cfg.Chain.ClConfig.SlotsPerEpoch == 0 || cfg.Chain.ClConfig.SecondsPerSlot == 0 {
		log.Fatal(nil, "invalid chain configuration specified, you must specify the slots per epoch, seconds per slot and genesis timestamp in the config file", 0)
	}
}

func ReadConfig(cfg *types.Config, path string) error {
	configPathFromEnv := os.Getenv("BEACONCHAIN_CONFIG")

	if configPathFromEnv != "" { // allow the location of the config file to be passed via env args
		path = configPathFromEnv
	}
	if strings.HasPrefix(path, "projects/") {
		x, err := AccessSecretVersion(path)
		if err != nil {
			return fmt.Errorf("error getting config from secret store: %v", err)
		}
		err = yaml.Unmarshal([]byte(*x), cfg)
		if err != nil {
			return fmt.Errorf("error decoding config file %v: %v", path, err)
		}

		log.Infof("seeded config file from secret store")
	} else {
		err := readConfigFile(cfg, path)
		if err != nil {
			return err
		}
	}

	err := readConfigEnv(cfg)
	if err != nil {
		return err
	}
	err = readConfigSecrets(cfg)
	if err != nil {
		return err
	}

	if cfg.Frontend.SiteBrand == "" {
		cfg.Frontend.SiteBrand = "beaconcha.in"
	}

	err = setCLConfig(cfg)
	if err != nil {
		return err
	}

	err = setELConfig(cfg)
	if err != nil {
		return err
	}

	cfg.Chain.Name = cfg.Chain.ClConfig.ConfigName

	// match DeploymentType to development, staging, production. if its empty fallback to development
	validTypes := []string{"development", "development_noisy", "staging", "production"}
	if cfg.DeploymentType == "" {
		log.Warn("DeploymentType not set, defaulting to development")
		cfg.DeploymentType = validTypes[0]
	}
	if !slices.Contains(validTypes, cfg.DeploymentType) {
		log.Fatal(fmt.Errorf("invalid DeploymentType: %v (valid types: %v)", cfg.DeploymentType, validTypes), "", 0)
	}

	if cfg.Chain.GenesisTimestamp == 0 {
		switch cfg.Chain.Name {
		case "mainnet":
			cfg.Chain.GenesisTimestamp = 1606824023
		case "prater":
			cfg.Chain.GenesisTimestamp = 1616508000
		case "sepolia":
			cfg.Chain.GenesisTimestamp = 1655733600
		case "zhejiang":
			cfg.Chain.GenesisTimestamp = 1675263600
		case "gnosis":
			cfg.Chain.GenesisTimestamp = 1638993340
		case "holesky":
			cfg.Chain.GenesisTimestamp = 1695902400
		default:
			return fmt.Errorf("tried to set known genesis-timestamp, but unknown chain-name")
		}
	}

	if cfg.Chain.GenesisValidatorsRoot == "" {
		switch cfg.Chain.Name {
		case "mainnet":
			cfg.Chain.GenesisValidatorsRoot = "0x4b363db94e286120d76eb905340fdd4e54bfe9f06bf33ff6cf5ad27f511bfe95"
		case "prater":
			cfg.Chain.GenesisValidatorsRoot = "0x043db0d9a83813551ee2f33450d23797757d430911a9320530ad8a0eabc43efb"
		case "sepolia":
			cfg.Chain.GenesisValidatorsRoot = "0xd8ea171f3c94aea21ebc42a1ed61052acf3f9209c00e4efbaaddac09ed9b8078"
		case "zhejiang":
			cfg.Chain.GenesisValidatorsRoot = "0x53a92d8f2bb1d85f62d16a156e6ebcd1bcaba652d0900b2c2f387826f3481f6f"
		case "gnosis":
			cfg.Chain.GenesisValidatorsRoot = "0xf5dcb5564e829aab27264b9becd5dfaa017085611224cb3036f573368dbb9d47"
		case "holesky":
			cfg.Chain.GenesisValidatorsRoot = "0x9143aa7c615a7f7115e2b6aac319c03529df8242ae705fba9df39b79c59fa8b1"
		default:
			return fmt.Errorf("tried to set known genesis-validators-root, but unknown chain-name")
		}
	}

	if cfg.Chain.DomainBLSToExecutionChange == "" {
		cfg.Chain.DomainBLSToExecutionChange = "0x0A000000"
	}
	if cfg.Chain.DomainVoluntaryExit == "" {
		cfg.Chain.DomainVoluntaryExit = "0x04000000"
	}

	if cfg.Frontend.ClCurrency == "" {
		switch cfg.Chain.Name {
		case "gnosis":
			cfg.Frontend.MainCurrency = "GNO"
			cfg.Frontend.ClCurrency = "mGNO"
			cfg.Frontend.ClCurrencyDecimals = 18
			cfg.Frontend.ClCurrencyDivisor = 1e9
		default:
			cfg.Frontend.MainCurrency = "ETH"
			cfg.Frontend.ClCurrency = "ETH"
			cfg.Frontend.ClCurrencyDecimals = 18
			cfg.Frontend.ClCurrencyDivisor = 1e9
		}
	}

	if cfg.Frontend.ElCurrency == "" {
		switch cfg.Chain.Name {
		case "gnosis":
			cfg.Frontend.ElCurrency = "xDAI"
			cfg.Frontend.ElCurrencyDecimals = 18
			cfg.Frontend.ElCurrencyDivisor = 1e18
		default:
			cfg.Frontend.ElCurrency = "ETH"
			cfg.Frontend.ElCurrencyDecimals = 18
			cfg.Frontend.ElCurrencyDivisor = 1e18
		}
	}

	if cfg.Frontend.SiteTitle == "" {
		cfg.Frontend.SiteTitle = "Open Source Ethereum Explorer"
	}

	if cfg.Frontend.Keywords == "" {
		cfg.Frontend.Keywords = "open source ethereum block explorer, ethereum block explorer, beacon chain explorer, ethereum blockchain explorer"
	}

	if cfg.Frontend.Ratelimits.FreeDay == 0 {
		cfg.Frontend.Ratelimits.FreeDay = 30000
	}
	if cfg.Frontend.Ratelimits.FreeMonth == 0 {
		cfg.Frontend.Ratelimits.FreeMonth = 30000
	}
	if cfg.Frontend.Ratelimits.SapphierDay == 0 {
		cfg.Frontend.Ratelimits.SapphierDay = 100000
	}
	if cfg.Frontend.Ratelimits.SapphierMonth == 0 {
		cfg.Frontend.Ratelimits.SapphierMonth = 500000
	}
	if cfg.Frontend.Ratelimits.EmeraldDay == 0 {
		cfg.Frontend.Ratelimits.EmeraldDay = 200000
	}
	if cfg.Frontend.Ratelimits.EmeraldMonth == 0 {
		cfg.Frontend.Ratelimits.EmeraldMonth = 1000000
	}
	if cfg.Frontend.Ratelimits.DiamondDay == 0 {
		cfg.Frontend.Ratelimits.DiamondDay = 6000000
	}
	if cfg.Frontend.Ratelimits.DiamondMonth == 0 {
		cfg.Frontend.Ratelimits.DiamondMonth = 6000000
	}

	if cfg.Chain.Id != 0 {
		switch cfg.Chain.Name {
		case "mainnet", "ethereum":
			cfg.Chain.Id = 1
		case "prater", "goerli":
			cfg.Chain.Id = 5
		case "holesky":
			cfg.Chain.Id = 17000
		case "sepolia":
			cfg.Chain.Id = 11155111
		case "gnosis":
			cfg.Chain.Id = 100
		}
	}

	// we check for machine chain id just for safety
	if cfg.Chain.Id != 0 && cfg.Chain.Id != cfg.Chain.ClConfig.DepositChainID {
		log.Fatal(fmt.Errorf("cfg.Chain.Id != cfg.Chain.ClConfig.DepositChainID: %v != %v", cfg.Chain.Id, cfg.Chain.ClConfig.DepositChainID), "", 0)
	}

	cfg.Chain.Id = cfg.Chain.ClConfig.DepositChainID

	if cfg.RedisSessionStoreEndpoint == "" && cfg.RedisCacheEndpoint != "" {
		log.Warnf("using RedisCacheEndpoint %s as RedisSessionStoreEndpoint as no dedicated RedisSessionStoreEndpoint was provided", cfg.RedisCacheEndpoint)
		cfg.RedisSessionStoreEndpoint = cfg.RedisCacheEndpoint
	}

	confSanityCheck(cfg)

	log.InfoWithFields(log.Fields{
		"genesisTimestamp":       cfg.Chain.GenesisTimestamp,
		"genesisValidatorsRoot":  cfg.Chain.GenesisValidatorsRoot,
		"configName":             cfg.Chain.ClConfig.ConfigName,
		"depositChainID":         cfg.Chain.ClConfig.DepositChainID,
		"depositNetworkID":       cfg.Chain.ClConfig.DepositNetworkID,
		"depositContractAddress": cfg.Chain.ClConfig.DepositContractAddress,
		"clCurrency":             cfg.Frontend.ClCurrency,
		"elCurrency":             cfg.Frontend.ElCurrency,
		"mainCurrency":           cfg.Frontend.MainCurrency,
	}, "did init config")

	Config = cfg
	return nil
}

func setELConfig(cfg *types.Config) error {
	var err error
	type MinimalELConfig struct {
		ByzantiumBlock      *big.Int `yaml:"BYZANTIUM_FORK_BLOCK,omitempty"`      // Byzantium switch block (nil = no fork, 0 = already on byzantium)
		ConstantinopleBlock *big.Int `yaml:"CONSTANTINOPLE_FORK_BLOCK,omitempty"` // Constantinople switch block (nil = no fork, 0 = already activated)
	}
	if cfg.Chain.ElConfigPath == "" {
		minimalCfg := MinimalELConfig{}
		switch cfg.Chain.Name {
		case "mainnet":
			err = yaml.Unmarshal([]byte(config.MainnetChainYml), &minimalCfg)
		case "prater":
			err = yaml.Unmarshal([]byte(config.PraterChainYml), &minimalCfg)
		case "ropsten":
			err = yaml.Unmarshal([]byte(config.RopstenChainYml), &minimalCfg)
		case "sepolia":
			err = yaml.Unmarshal([]byte(config.SepoliaChainYml), &minimalCfg)
		case "gnosis":
			err = yaml.Unmarshal([]byte(config.GnosisChainYml), &minimalCfg)
		case "holesky":
			err = yaml.Unmarshal([]byte(config.HoleskyChainYml), &minimalCfg)
		default:
			return fmt.Errorf("tried to set known chain-config, but unknown chain-name: %v (path: %v)", cfg.Chain.Name, cfg.Chain.ElConfigPath)
		}
		if err != nil {
			return err
		}
		if minimalCfg.ByzantiumBlock == nil {
			minimalCfg.ByzantiumBlock = big.NewInt(0)
		}
		if minimalCfg.ConstantinopleBlock == nil {
			minimalCfg.ConstantinopleBlock = big.NewInt(0)
		}
		cfg.Chain.ElConfig = &params.ChainConfig{
			ChainID:             big.NewInt(int64(cfg.Chain.Id)),
			ByzantiumBlock:      minimalCfg.ByzantiumBlock,
			ConstantinopleBlock: minimalCfg.ConstantinopleBlock,
		}
	} else {
		f, err := os.Open(cfg.Chain.ElConfigPath)
		if err != nil {
			return fmt.Errorf("error opening EL Chain Config file %v: %w", cfg.Chain.ElConfigPath, err)
		}
		var chainConfig *params.ChainConfig
		decoder := json.NewDecoder(f)
		err = decoder.Decode(&chainConfig)
		if err != nil {
			return fmt.Errorf("error decoding EL Chain Config file %v: %v", cfg.Chain.ElConfigPath, err)
		}
		cfg.Chain.ElConfig = chainConfig
	}
	return nil
}

func setCLConfig(cfg *types.Config) error {
	var err error
	if cfg.Chain.ClConfigPath == "" {
		// var prysmParamsConfig *prysmParams.BeaconChainConfig
		switch cfg.Chain.Name {
		case "mainnet":
			err = yaml.Unmarshal([]byte(config.MainnetChainYml), &cfg.Chain.ClConfig)
		case "prater":
			err = yaml.Unmarshal([]byte(config.PraterChainYml), &cfg.Chain.ClConfig)
		case "ropsten":
			err = yaml.Unmarshal([]byte(config.RopstenChainYml), &cfg.Chain.ClConfig)
		case "sepolia":
			err = yaml.Unmarshal([]byte(config.SepoliaChainYml), &cfg.Chain.ClConfig)
		case "gnosis":
			err = yaml.Unmarshal([]byte(config.GnosisChainYml), &cfg.Chain.ClConfig)
		case "holesky":
			err = yaml.Unmarshal([]byte(config.HoleskyChainYml), &cfg.Chain.ClConfig)
		default:
			return fmt.Errorf("tried to set known chain-config, but unknown chain-name: %v (path: %v)", cfg.Chain.Name, cfg.Chain.ClConfigPath)
		}
		if err != nil {
			return err
		}
		// err = prysmParams.SetActive(prysmParamsConfig)
		// if err != nil {
		// 	return fmt.Errorf("error setting chainConfig (%v) for prysmParams: %w", cfg.Chain.Name, err)
		// }
	} else if cfg.Chain.ClConfigPath == "node" {
		nodeEndpoint := fmt.Sprintf("http://%s", net.JoinHostPort(cfg.Indexer.Node.Host, cfg.Indexer.Node.Port))
		client := consapi.NewClient(nodeEndpoint)

		jr, err := client.GetSpec()
		if err != nil {
			return err
		}

		maxForkEpoch := uint64(18446744073709551615)

		if jr.Data.AltairForkEpoch == nil {
			log.Warnf("AltairForkEpoch not set, defaulting to maxForkEpoch")
			jr.Data.AltairForkEpoch = &maxForkEpoch
		}
		if jr.Data.BellatrixForkEpoch == nil {
			log.Warnf("BellatrixForkEpoch not set, defaulting to maxForkEpoch")
			jr.Data.BellatrixForkEpoch = &maxForkEpoch
		}
		if jr.Data.CapellaForkEpoch == nil {
			log.Warnf("CapellaForkEpoch not set, defaulting to maxForkEpoch")
			jr.Data.CapellaForkEpoch = &maxForkEpoch
		}
		if jr.Data.DenebForkEpoch == nil {
			log.Warnf("DenebForkEpoch not set, defaulting to maxForkEpoch")
			jr.Data.DenebForkEpoch = &maxForkEpoch
		}

		chainCfg := types.ClChainConfig{
			PresetBase:                              jr.Data.PresetBase,
			ConfigName:                              jr.Data.ConfigName,
			TerminalTotalDifficulty:                 jr.Data.TerminalTotalDifficulty,
			TerminalBlockHash:                       jr.Data.TerminalBlockHash,
			TerminalBlockHashActivationEpoch:        jr.Data.TerminalBlockHashActivationEpoch,
			MinGenesisActiveValidatorCount:          uint64(jr.Data.MinGenesisActiveValidatorCount),
			MinGenesisTime:                          jr.Data.MinGenesisTime,
			GenesisForkVersion:                      jr.Data.GenesisForkVersion,
			GenesisDelay:                            uint64(jr.Data.GenesisDelay),
			AltairForkVersion:                       jr.Data.AltairForkVersion,
			AltairForkEpoch:                         *jr.Data.AltairForkEpoch,
			BellatrixForkVersion:                    jr.Data.BellatrixForkVersion,
			BellatrixForkEpoch:                      *jr.Data.BellatrixForkEpoch,
			CappellaForkVersion:                     jr.Data.CapellaForkVersion,
			CappellaForkEpoch:                       *jr.Data.CapellaForkEpoch,
			DenebForkVersion:                        jr.Data.DenebForkVersion,
			DenebForkEpoch:                          *jr.Data.DenebForkEpoch,
			SecondsPerSlot:                          uint64(jr.Data.SecondsPerSlot),
			SecondsPerEth1Block:                     uint64(jr.Data.SecondsPerEth1Block),
			MinValidatorWithdrawabilityDelay:        uint64(jr.Data.MinValidatorWithdrawabilityDelay),
			ShardCommitteePeriod:                    uint64(jr.Data.ShardCommitteePeriod),
			Eth1FollowDistance:                      uint64(jr.Data.Eth1FollowDistance),
			InactivityScoreBias:                     uint64(jr.Data.InactivityScoreBias),
			InactivityScoreRecoveryRate:             uint64(jr.Data.InactivityScoreRecoveryRate),
			EjectionBalance:                         uint64(jr.Data.EjectionBalance),
			MinPerEpochChurnLimit:                   uint64(jr.Data.MinPerEpochChurnLimit),
			ChurnLimitQuotient:                      uint64(jr.Data.ChurnLimitQuotient),
			ProposerScoreBoost:                      uint64(jr.Data.ProposerScoreBoost),
			DepositChainID:                          uint64(jr.Data.DepositChainID),
			DepositNetworkID:                        uint64(jr.Data.DepositNetworkID),
			DepositContractAddress:                  jr.Data.DepositContractAddress,
			MaxCommitteesPerSlot:                    uint64(jr.Data.MaxCommitteesPerSlot),
			TargetCommitteeSize:                     uint64(jr.Data.TargetCommitteeSize),
			MaxValidatorsPerCommittee:               uint64(jr.Data.TargetCommitteeSize),
			ShuffleRoundCount:                       uint64(jr.Data.ShuffleRoundCount),
			HysteresisQuotient:                      uint64(jr.Data.HysteresisQuotient),
			HysteresisDownwardMultiplier:            uint64(jr.Data.HysteresisDownwardMultiplier),
			HysteresisUpwardMultiplier:              uint64(jr.Data.HysteresisUpwardMultiplier),
			SafeSlotsToUpdateJustified:              uint64(jr.Data.SafeSlotsToUpdateJustified),
			MinDepositAmount:                        uint64(jr.Data.MinDepositAmount),
			MaxEffectiveBalance:                     uint64(jr.Data.MaxEffectiveBalance),
			EffectiveBalanceIncrement:               uint64(jr.Data.EffectiveBalanceIncrement),
			MinAttestationInclusionDelay:            uint64(jr.Data.MinAttestationInclusionDelay),
			SlotsPerEpoch:                           uint64(jr.Data.SlotsPerEpoch),
			MinSeedLookahead:                        uint64(jr.Data.MinSeedLookahead),
			MaxSeedLookahead:                        uint64(jr.Data.MaxSeedLookahead),
			EpochsPerEth1VotingPeriod:               uint64(jr.Data.EpochsPerEth1VotingPeriod),
			SlotsPerHistoricalRoot:                  uint64(jr.Data.SlotsPerHistoricalRoot),
			MinEpochsToInactivityPenalty:            uint64(jr.Data.MinEpochsToInactivityPenalty),
			EpochsPerHistoricalVector:               uint64(jr.Data.EpochsPerHistoricalVector),
			EpochsPerSlashingsVector:                uint64(jr.Data.EpochsPerSlashingsVector),
			HistoricalRootsLimit:                    uint64(jr.Data.HistoricalRootsLimit),
			ValidatorRegistryLimit:                  uint64(jr.Data.ValidatorRegistryLimit),
			BaseRewardFactor:                        uint64(jr.Data.BaseRewardFactor),
			WhistleblowerRewardQuotient:             uint64(jr.Data.WhistleblowerRewardQuotient),
			ProposerRewardQuotient:                  uint64(jr.Data.ProposerRewardQuotient),
			InactivityPenaltyQuotient:               uint64(jr.Data.InactivityPenaltyQuotient),
			MinSlashingPenaltyQuotient:              uint64(jr.Data.MinSlashingPenaltyQuotient),
			ProportionalSlashingMultiplier:          uint64(jr.Data.ProportionalSlashingMultiplier),
			MaxProposerSlashings:                    uint64(jr.Data.MaxProposerSlashings),
			MaxAttesterSlashings:                    uint64(jr.Data.MaxAttesterSlashings),
			MaxAttestations:                         uint64(jr.Data.MaxAttestations),
			MaxDeposits:                             uint64(jr.Data.MaxDeposits),
			MaxVoluntaryExits:                       uint64(jr.Data.MaxVoluntaryExits),
			InvactivityPenaltyQuotientAltair:        uint64(jr.Data.InactivityPenaltyQuotientAltair),
			MinSlashingPenaltyQuotientAltair:        uint64(jr.Data.MinSlashingPenaltyQuotientAltair),
			ProportionalSlashingMultiplierAltair:    uint64(jr.Data.ProportionalSlashingMultiplierAltair),
			SyncCommitteeSize:                       uint64(jr.Data.SyncCommitteeSize),
			EpochsPerSyncCommitteePeriod:            uint64(jr.Data.EpochsPerSyncCommitteePeriod),
			MinSyncCommitteeParticipants:            uint64(jr.Data.MinSyncCommitteeParticipants),
			InvactivityPenaltyQuotientBellatrix:     uint64(jr.Data.InactivityPenaltyQuotientBellatrix),
			MinSlashingPenaltyQuotientBellatrix:     uint64(jr.Data.MinSlashingPenaltyQuotientBellatrix),
			ProportionalSlashingMultiplierBellatrix: uint64(jr.Data.ProportionalSlashingMultiplierBellatrix),
			MaxBytesPerTransaction:                  uint64(jr.Data.MaxBytesPerTransaction),
			MaxTransactionsPerPayload:               uint64(jr.Data.MaxTransactionsPerPayload),
			BytesPerLogsBloom:                       uint64(jr.Data.BytesPerLogsBloom),
			MaxExtraDataBytes:                       uint64(jr.Data.MaxExtraDataBytes),
			MaxWithdrawalsPerPayload:                uint64(jr.Data.MaxWithdrawalsPerPayload),
			MaxValidatorsPerWithdrawalSweep:         uint64(jr.Data.MaxValidatorsPerWithdrawalsSweep),
			MaxBlsToExecutionChange:                 uint64(jr.Data.MaxBlsToExecutionChanges),
		}

		cfg.Chain.ClConfig = chainCfg

		gtr, err := client.GetGenesis()
		if err != nil {
			return err
		}

		cfg.Chain.GenesisTimestamp = mustParseUint(gtr.Data.GenesisTime)
		cfg.Chain.GenesisValidatorsRoot = gtr.Data.GenesisValidatorsRoot

		log.Infof("loaded chain config from node with genesis time %s", gtr.Data.GenesisTime)
	} else {
		f, err := os.Open(cfg.Chain.ClConfigPath)
		if err != nil {
			return fmt.Errorf("error opening Chain Config file %v: %w", cfg.Chain.ClConfigPath, err)
		}
		var chainConfig *types.ClChainConfig
		decoder := yaml.NewDecoder(f)
		err = decoder.Decode(&chainConfig)
		if err != nil {
			return fmt.Errorf("error decoding Chain Config file %v: %v", cfg.Chain.ClConfigPath, err)
		}
		cfg.Chain.ClConfig = *chainConfig
	}

	// rewrite to match to allow trace as well
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "trace":
		logrus.SetLevel(logrus.TraceLevel)
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "fatal":
		logrus.SetLevel(logrus.FatalLevel)
	case "panic":
		logrus.SetLevel(logrus.PanicLevel)
	}

	return nil
}
