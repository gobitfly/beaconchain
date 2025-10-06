package commands

import (
	"flag"
	"fmt"
	"time"

	"github.com/gobitfly/beaconchain/cmd/misc/misctypes"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

type MigrateLegacyBalancesCommand struct {
	FlagSet *flag.FlagSet
	Config  migrateLegacyBalancesCommandConfig
}

type migrateLegacyBalancesCommandConfig struct {
	StartEpoch uint64
	EndEpoch   uint64
}

func (s *MigrateLegacyBalancesCommand) ParseCommandOptions() {
	s.FlagSet.Uint64Var(&s.Config.StartEpoch, "legacy-balances-start-epoch", 0, "Start epoch for the balance migration")
	s.FlagSet.Uint64Var(&s.Config.EndEpoch, "legacy-balances-end-epoch", 0, "End epoch for the balance migration")
}

func (s *MigrateLegacyBalancesCommand) Requires() misctypes.Requires {
	return misctypes.Requires{
		ClickhouseDBs: true,
		ClNode:        true,
	}
}

func (s *MigrateLegacyBalancesCommand) Run(clClient *rpc.LighthouseClient) error {
	if s.Config.EndEpoch == 0 {
		s.showHelp()
		return fmt.Errorf("end-epoch is required")
	}

	log.Infof("command: migrate-legacy-balances start-epoch=%d end-epoch=%d", s.Config.StartEpoch, s.Config.EndEpoch)

	for epoch := s.Config.StartEpoch; epoch <= s.Config.EndEpoch; epoch++ {
		for attempt := 1; attempt <= 5; attempt++ {
			state, err := clClient.GetValidatorState(epoch)
			if err != nil {
				log.Error(err, fmt.Sprintf("error getting balances for epoch %d (attempt %d/5): %v", epoch, attempt, err), 0)
				if attempt == 5 {
					return fmt.Errorf("failed to get balances for epoch %d after %d attempts: %w", epoch, attempt, err)
				}
				time.Sleep(5 * time.Second)
				continue
			}

			validators := make([]*types.Validator, 0, len(state.Data))
			for _, validator := range state.Data {
				validators = append(validators, &types.Validator{
					Index:            validator.Index,
					Balance:          validator.Balance,
					EffectiveBalance: validator.Validator.EffectiveBalance,
				})
			}

			log.Infof("writing epoch %d to clickhouse with %d validators", epoch, len(state.Data))
			err = db.SaveLegacyValidatorBalancesToClickhouse(epoch, validators)
			if err != nil {
				log.Error(err, fmt.Sprintf("error saving legacy balances for epoch %d (attempt %d/5): %v", epoch, attempt, err), 0)
				if attempt == 5 {
					return fmt.Errorf("error saving legacy balances for epoch %d after %d attempts: %w", epoch, attempt, err)
				}
				time.Sleep(5 * time.Second)
				continue
			}

			// success for this epoch
			break
		}
	}
	return nil
}

func (s *MigrateLegacyBalancesCommand) showHelp() {
	log.Infof("Usage: migrate_legacy_balances [options]")
	log.Infof("Options:")
	log.Infof("  --legacy-balances-start-epoch int\tStart epoch for the balance migration")
	log.Infof("  --legacy-balances-end-epoch int\tEnd epoch for the balance migration")
}
