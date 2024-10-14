package commands

import (
	"flag"
	"fmt"

	"github.com/gobitfly/beaconchain/cmd/misc/misctypes"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
)

type ClickhouseDataValidatorCommand struct {
	FlagSet *flag.FlagSet
	Config  clickhouseDataValidatorCommandConfig
}

type clickhouseDataValidatorCommandConfig struct {
	DryRun            bool
	Force             bool // bypass summary confirm
	BundleURL         string
	BundleVersionCode int64
	NativeVersionCode int64
	TargetInstalls    int64
}

func (s *ClickhouseDataValidatorCommand) ParseCommandOptions() {
	// s.FlagSet.Int64Var(&s.Config.BundleVersionCode, "version-code", 0, "Version code of that bundle (Default: Next)")
	// s.FlagSet.Int64Var(&s.Config.NativeVersionCode, "min-native-version", 0, "Minimum required native version (Default: Current)")
	// s.FlagSet.Int64Var(&s.Config.TargetInstalls, "target-installs", -1, "How many people to roll out to (Default: All)")
	// s.FlagSet.StringVar(&s.Config.BundleURL, "bundle-url", "", "URL to bundle that contains the update, bundle.zip")
	// s.FlagSet.BoolVar(&s.Config.Force, "force", false, "Skips summary and confirmation")
}

func (s *ClickhouseDataValidatorCommand) Requires() misctypes.Requires {
	return misctypes.Requires{
		Bigtable:   true,
		Redis:      true,
		NetworkDBs: true,
		Clickhouse: true,
	}
}

func (s *ClickhouseDataValidatorCommand) Run() error {
	err := s.verifyTable("validator_dashboard_data_rolling_24h", 1000)
	if err != nil {
		return err
	}

	return nil
}

func (s *ClickhouseDataValidatorCommand) verifyTable(table string, validatorsToCheck int) error {
	var err error

	retSyncCommitteeCheck := []struct {
		ValidatorIndex uint64 `db:"validator_index"`
		StartEpoch     uint64 `db:"epoch_start"`
		EndEpoch       uint64 `db:"epoch_end"`
		SyncScheduled  uint64 `db:"sync_scheduled"`
		SyncExecuted   uint64 `db:"sync_executed"`
	}{}

	err = db.ClickHouseReader.Select(&retSyncCommitteeCheck, `SELECT validator_index, epoch_start, epoch_end, sync_scheduled, sync_executed FROM `+table+` final WHERE sync_scheduled > 0 ORDER BY RAND() LIMIT $1`, validatorsToCheck)
	if err != nil {
		return err
	}

	missedBlocksCache := make(map[string]uint64)
	for _, r := range retSyncCommitteeCheck {
		log.Infof("checking sync committee data for validator %d", r.ValidatorIndex)
		if r.SyncExecuted > r.SyncScheduled {
			return fmt.Errorf("validator %d has more executed more sync tasks than scheduled in table %s", r.ValidatorIndex, table)
		}
		syncData, err := db.BigtableClient.GetValidatorSyncDutiesHistory([]uint64{r.ValidatorIndex}, r.StartEpoch*32, r.EndEpoch*32+31)
		if err != nil {
			return err
		}
		syncScheduled := uint64(0)
		syncExecuted := uint64(0)
		for validator, duties := range syncData {
			for _, duty := range duties {
				log.Debugf("bigtable: %d has a sync duty in epoch %d with status %d", validator, duty.Slot/32, duty.Status)
				syncScheduled++
				if duty.Status == 1 {
					syncExecuted++
				}
			}
		}

		// retrieve the number of missed blocks in the range
		key := fmt.Sprintf("%d-%d", r.StartEpoch, r.EndEpoch)
		_, ok := missedBlocksCache[key]
		if !ok {
			missedBlocksFromDb := uint64(0)
			err = db.ReaderDb.Get(&missedBlocksFromDb, `SELECT COUNT(*) FROM blocks WHERE status != '1' AND epoch BETWEEN $1 AND $2`, r.StartEpoch, r.EndEpoch)
			if err != nil {
				return err
			}
			missedBlocksCache[key] = missedBlocksFromDb
		}
		missedBlocks := missedBlocksCache[key]

		if syncScheduled != r.SyncScheduled+missedBlocks {
			return fmt.Errorf("validator %d has %d scheduled sync duties in clickhouse but %d in bigtable between epoch %d and %d (with %d missed slots)", r.ValidatorIndex, r.SyncScheduled, syncScheduled, r.StartEpoch, r.EndEpoch, missedBlocks)
		}
		if syncExecuted != r.SyncExecuted {
			return fmt.Errorf("validator %d has %d executed sync duties in clickhouse but %d in bigtable between epoch %d and %d", r.ValidatorIndex, r.SyncExecuted, syncExecuted, r.StartEpoch, r.EndEpoch)
		}
	}

	retAttestationCheck := []struct {
		ValidatorIndex        uint64 `db:"validator_index"`
		StartEpoch            uint64 `db:"epoch_start"`
		EndEpoch              uint64 `db:"epoch_end"`
		AttestationsScheduled uint64 `db:"attestations_scheduled"`
		AttestationsExecuted  uint64 `db:"attestations_executed"`
	}{}

	err = db.ClickHouseReader.Select(&retAttestationCheck, `SELECT validator_index, epoch_start, epoch_end, attestations_scheduled, attestations_executed FROM `+table+` final ORDER BY RAND() LIMIT $1`, validatorsToCheck)
	if err != nil {
		return err
	}
	for _, r := range retAttestationCheck {
		log.Infof("checking attestation data for validator %d", r.ValidatorIndex)
		if r.AttestationsExecuted > r.AttestationsScheduled {
			return fmt.Errorf("validator %d has more executed attestations than scheduled in table %s", r.ValidatorIndex, table)
		}

		totalAttestations := r.AttestationsScheduled
		missedAttestations := totalAttestations - r.AttestationsExecuted

		missedAttestaionsBigtable, err := db.BigtableClient.GetValidatorMissedAttestationHistory([]uint64{r.ValidatorIndex}, r.StartEpoch, r.EndEpoch)
		if err != nil {
			return err
		}
		missedCount := uint64(0)
		for validator, atts := range missedAttestaionsBigtable {
			for slot := range atts {
				log.Infof("bigtable: %d missed an attestation in epoch %d", validator, slot/32)
				missedCount++
			}
		}
		if missedCount != missedAttestations {
			return fmt.Errorf("validator %d has %d missed attestations in clickhouse but %d in bigtable", r.ValidatorIndex, missedAttestations, missedCount)
		}
	}

	retProposalsCheck := []struct {
		ValidatorIndex  uint64 `db:"validator_index"`
		StartEpoch      uint64 `db:"epoch_start"`
		EndEpoch        uint64 `db:"epoch_end"`
		BlocksScheduled uint64 `db:"blocks_scheduled"`
		BlocksProposed  uint64 `db:"blocks_proposed"`
	}{}

	err = db.ClickHouseReader.Select(&retProposalsCheck, `SELECT validator_index, epoch_start, epoch_end, blocks_scheduled, blocks_proposed FROM `+table+` final WHERE blocks_scheduled > 0 ORDER BY RAND() LIMIT $1`, validatorsToCheck)
	if err != nil {
		return err
	}

	for _, r := range retProposalsCheck {
		log.Infof("checking proposal data for validator %d", r.ValidatorIndex)
		if r.BlocksProposed > r.BlocksScheduled {
			return fmt.Errorf("validator %d has more proposed blocks than scheduled in table %s", r.ValidatorIndex, table)
		}
		proposalData, err := db.BigtableClient.GetValidatorProposalHistory([]uint64{r.ValidatorIndex}, r.StartEpoch, r.EndEpoch)
		if err != nil {
			return err
		}
		blocksScheduled := uint64(0)
		blocksProposed := uint64(0)
		for validator, blocks := range proposalData {
			for _, block := range blocks {
				log.Infof("bigtable: %d proposed a block in epoch %d with status %d", validator, block.Slot/32, block.Status)
				blocksScheduled++
				if block.Status == 1 {
					blocksProposed++
				}
			}
		}

		if blocksScheduled != r.BlocksScheduled {
			return fmt.Errorf("validator %d has %d scheduled blocks in clickhouse but %d in bigtable", r.ValidatorIndex, r.BlocksScheduled, blocksScheduled)
		}
		if blocksProposed != r.BlocksProposed {
			return fmt.Errorf("validator %d has %d proposed blocks in clickhouse but %d in bigtable", r.ValidatorIndex, r.BlocksProposed, blocksProposed)
		}
	}
	return nil
}
