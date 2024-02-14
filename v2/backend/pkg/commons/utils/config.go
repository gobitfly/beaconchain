package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/carlmjohnson/requests"
	"github.com/ethereum/go-ethereum/params"
	"github.com/gobitfly/beaconchain/pkg/commons/config"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/kelseyhightower/envconfig"
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

		logrus.Infof("seeded config file from secret store")
	} else {
		err := readConfigFile(cfg, path)
		if err != nil {
			return err
		}
	}

	readConfigEnv(cfg)
	err := readConfigSecrets(cfg)
	if err != nil {
		return err
	}

	if cfg.Frontend.SiteBrand == "" {
		cfg.Frontend.SiteBrand = "beaconcha.in"
	}

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
		nodeEndpoint := fmt.Sprintf("http://%s:%s", cfg.Indexer.Node.Host, cfg.Indexer.Node.Port)

		jr := &types.ConfigJsonResponse{}

		err := requests.
			URL(nodeEndpoint + "/eth/v1/config/spec").
			ToJSON(jr).
			Fetch(context.Background())

		if err != nil {
			return err
		}

		chainCfg := types.ClChainConfig{
			PresetBase:                              jr.Data.PresetBase,
			ConfigName:                              jr.Data.ConfigName,
			TerminalTotalDifficulty:                 jr.Data.TerminalTotalDifficulty,
			TerminalBlockHash:                       jr.Data.TerminalBlockHash,
			TerminalBlockHashActivationEpoch:        mustParseUint(jr.Data.TerminalBlockHashActivationEpoch),
			MinGenesisActiveValidatorCount:          mustParseUint(jr.Data.MinGenesisActiveValidatorCount),
			MinGenesisTime:                          int64(mustParseUint(jr.Data.MinGenesisTime)),
			GenesisForkVersion:                      jr.Data.GenesisForkVersion,
			GenesisDelay:                            mustParseUint(jr.Data.GenesisDelay),
			AltairForkVersion:                       jr.Data.AltairForkVersion,
			AltairForkEpoch:                         mustParseUint(jr.Data.AltairForkEpoch),
			BellatrixForkVersion:                    jr.Data.BellatrixForkVersion,
			BellatrixForkEpoch:                      mustParseUint(jr.Data.BellatrixForkEpoch),
			CappellaForkVersion:                     jr.Data.CapellaForkVersion,
			CappellaForkEpoch:                       mustParseUint(jr.Data.CapellaForkEpoch),
			DenebForkVersion:                        jr.Data.DenebForkVersion,
			DenebForkEpoch:                          mustParseUint(jr.Data.DenebForkEpoch),
			SecondsPerSlot:                          mustParseUint(jr.Data.SecondsPerSlot),
			SecondsPerEth1Block:                     mustParseUint(jr.Data.SecondsPerEth1Block),
			MinValidatorWithdrawabilityDelay:        mustParseUint(jr.Data.MinValidatorWithdrawabilityDelay),
			ShardCommitteePeriod:                    mustParseUint(jr.Data.ShardCommitteePeriod),
			Eth1FollowDistance:                      mustParseUint(jr.Data.Eth1FollowDistance),
			InactivityScoreBias:                     mustParseUint(jr.Data.InactivityScoreBias),
			InactivityScoreRecoveryRate:             mustParseUint(jr.Data.InactivityScoreRecoveryRate),
			EjectionBalance:                         mustParseUint(jr.Data.EjectionBalance),
			MinPerEpochChurnLimit:                   mustParseUint(jr.Data.MinPerEpochChurnLimit),
			ChurnLimitQuotient:                      mustParseUint(jr.Data.ChurnLimitQuotient),
			ProposerScoreBoost:                      mustParseUint(jr.Data.ProposerScoreBoost),
			DepositChainID:                          mustParseUint(jr.Data.DepositChainID),
			DepositNetworkID:                        mustParseUint(jr.Data.DepositNetworkID),
			DepositContractAddress:                  jr.Data.DepositContractAddress,
			MaxCommitteesPerSlot:                    mustParseUint(jr.Data.MaxCommitteesPerSlot),
			TargetCommitteeSize:                     mustParseUint(jr.Data.TargetCommitteeSize),
			MaxValidatorsPerCommittee:               mustParseUint(jr.Data.TargetCommitteeSize),
			ShuffleRoundCount:                       mustParseUint(jr.Data.ShuffleRoundCount),
			HysteresisQuotient:                      mustParseUint(jr.Data.HysteresisQuotient),
			HysteresisDownwardMultiplier:            mustParseUint(jr.Data.HysteresisDownwardMultiplier),
			HysteresisUpwardMultiplier:              mustParseUint(jr.Data.HysteresisUpwardMultiplier),
			SafeSlotsToUpdateJustified:              mustParseUint(jr.Data.SafeSlotsToUpdateJustified),
			MinDepositAmount:                        mustParseUint(jr.Data.MinDepositAmount),
			MaxEffectiveBalance:                     mustParseUint(jr.Data.MaxEffectiveBalance),
			EffectiveBalanceIncrement:               mustParseUint(jr.Data.EffectiveBalanceIncrement),
			MinAttestationInclusionDelay:            mustParseUint(jr.Data.MinAttestationInclusionDelay),
			SlotsPerEpoch:                           mustParseUint(jr.Data.SlotsPerEpoch),
			MinSeedLookahead:                        mustParseUint(jr.Data.MinSeedLookahead),
			MaxSeedLookahead:                        mustParseUint(jr.Data.MaxSeedLookahead),
			EpochsPerEth1VotingPeriod:               mustParseUint(jr.Data.EpochsPerEth1VotingPeriod),
			SlotsPerHistoricalRoot:                  mustParseUint(jr.Data.SlotsPerHistoricalRoot),
			MinEpochsToInactivityPenalty:            mustParseUint(jr.Data.MinEpochsToInactivityPenalty),
			EpochsPerHistoricalVector:               mustParseUint(jr.Data.EpochsPerHistoricalVector),
			EpochsPerSlashingsVector:                mustParseUint(jr.Data.EpochsPerSlashingsVector),
			HistoricalRootsLimit:                    mustParseUint(jr.Data.HistoricalRootsLimit),
			ValidatorRegistryLimit:                  mustParseUint(jr.Data.ValidatorRegistryLimit),
			BaseRewardFactor:                        mustParseUint(jr.Data.BaseRewardFactor),
			WhistleblowerRewardQuotient:             mustParseUint(jr.Data.WhistleblowerRewardQuotient),
			ProposerRewardQuotient:                  mustParseUint(jr.Data.ProposerRewardQuotient),
			InactivityPenaltyQuotient:               mustParseUint(jr.Data.InactivityPenaltyQuotient),
			MinSlashingPenaltyQuotient:              mustParseUint(jr.Data.MinSlashingPenaltyQuotient),
			ProportionalSlashingMultiplier:          mustParseUint(jr.Data.ProportionalSlashingMultiplier),
			MaxProposerSlashings:                    mustParseUint(jr.Data.MaxProposerSlashings),
			MaxAttesterSlashings:                    mustParseUint(jr.Data.MaxAttesterSlashings),
			MaxAttestations:                         mustParseUint(jr.Data.MaxAttestations),
			MaxDeposits:                             mustParseUint(jr.Data.MaxDeposits),
			MaxVoluntaryExits:                       mustParseUint(jr.Data.MaxVoluntaryExits),
			InvactivityPenaltyQuotientAltair:        mustParseUint(jr.Data.InactivityPenaltyQuotientAltair),
			MinSlashingPenaltyQuotientAltair:        mustParseUint(jr.Data.MinSlashingPenaltyQuotientAltair),
			ProportionalSlashingMultiplierAltair:    mustParseUint(jr.Data.ProportionalSlashingMultiplierAltair),
			SyncCommitteeSize:                       mustParseUint(jr.Data.SyncCommitteeSize),
			EpochsPerSyncCommitteePeriod:            mustParseUint(jr.Data.EpochsPerSyncCommitteePeriod),
			MinSyncCommitteeParticipants:            mustParseUint(jr.Data.MinSyncCommitteeParticipants),
			InvactivityPenaltyQuotientBellatrix:     mustParseUint(jr.Data.InactivityPenaltyQuotientBellatrix),
			MinSlashingPenaltyQuotientBellatrix:     mustParseUint(jr.Data.MinSlashingPenaltyQuotientBellatrix),
			ProportionalSlashingMultiplierBellatrix: mustParseUint(jr.Data.ProportionalSlashingMultiplierBellatrix),
			MaxBytesPerTransaction:                  mustParseUint(jr.Data.MaxBytesPerTransaction),
			MaxTransactionsPerPayload:               mustParseUint(jr.Data.MaxTransactionsPerPayload),
			BytesPerLogsBloom:                       mustParseUint(jr.Data.BytesPerLogsBloom),
			MaxExtraDataBytes:                       mustParseUint(jr.Data.MaxExtraDataBytes),
			MaxWithdrawalsPerPayload:                mustParseUint(jr.Data.MaxWithdrawalsPerPayload),
			MaxValidatorsPerWithdrawalSweep:         mustParseUint(jr.Data.MaxValidatorsPerWithdrawalsSweep),
			MaxBlsToExecutionChange:                 mustParseUint(jr.Data.MaxBlsToExecutionChanges),
		}

		if jr.Data.AltairForkEpoch == "" {
			chainCfg.AltairForkEpoch = 18446744073709551615
		}
		if jr.Data.BellatrixForkEpoch == "" {
			chainCfg.BellatrixForkEpoch = 18446744073709551615
		}
		if jr.Data.CapellaForkEpoch == "" {
			chainCfg.CappellaForkEpoch = 18446744073709551615
		}
		if jr.Data.DenebForkEpoch == "" {
			chainCfg.DenebForkEpoch = 18446744073709551615
		}

		cfg.Chain.ClConfig = chainCfg

		type GenesisResponse struct {
			Data struct {
				GenesisTime           string `json:"genesis_time"`
				GenesisValidatorsRoot string `json:"genesis_validators_root"`
				GenesisForkVersion    string `json:"genesis_fork_version"`
			} `json:"data"`
		}

		gtr := &GenesisResponse{}

		err = requests.
			URL(nodeEndpoint + "/eth/v1/beacon/genesis").
			ToJSON(gtr).
			Fetch(context.Background())

		if err != nil {
			return err
		}

		cfg.Chain.GenesisTimestamp = mustParseUint(gtr.Data.GenesisTime)
		cfg.Chain.GenesisValidatorsRoot = gtr.Data.GenesisValidatorsRoot

		logrus.Infof("loaded chain config from node with genesis time %s", gtr.Data.GenesisTime)
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

	cfg.Chain.Name = cfg.Chain.ClConfig.ConfigName

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
		logrus.Fatalf("cfg.Chain.Id != cfg.Chain.ClConfig.DepositChainID: %v != %v", cfg.Chain.Id, cfg.Chain.ClConfig.DepositChainID)
	}

	cfg.Chain.Id = cfg.Chain.ClConfig.DepositChainID

	if cfg.RedisSessionStoreEndpoint == "" && cfg.RedisCacheEndpoint != "" {
		logrus.Infof("using RedisCacheEndpoint %s as RedisSessionStoreEndpoint as no dedicated RedisSessionStoreEndpoint was provided", cfg.RedisCacheEndpoint)
		cfg.RedisSessionStoreEndpoint = cfg.RedisCacheEndpoint
	}

	logrus.WithFields(logrus.Fields{
		"genesisTimestamp":       cfg.Chain.GenesisTimestamp,
		"genesisValidatorsRoot":  cfg.Chain.GenesisValidatorsRoot,
		"configName":             cfg.Chain.ClConfig.ConfigName,
		"depositChainID":         cfg.Chain.ClConfig.DepositChainID,
		"depositNetworkID":       cfg.Chain.ClConfig.DepositNetworkID,
		"depositContractAddress": cfg.Chain.ClConfig.DepositContractAddress,
		"clCurrency":             cfg.Frontend.ClCurrency,
		"elCurrency":             cfg.Frontend.ElCurrency,
		"mainCurrency":           cfg.Frontend.MainCurrency,
	}).Infof("did init config")

	return nil
}
