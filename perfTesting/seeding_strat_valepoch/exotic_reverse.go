package seeding_strat_valepoch

import (
	"fmt"
	"math"
	"perftesting/db"
	"perftesting/seeding"
)

// Reversed exotic, in theory does not really makes sense but I wanted to see if my assumptions about performance is correct

type SeederPartitionExoticReverse struct {
	NumberOfEpochPartitions int
	NumberOfValiPartitions  int
}

func GetSeederPartitionExoticReverse(tableName string, noOfEpochPartitions, notOfValiPartitions int, columnarEngine bool) *seeding.Seeder {
	return getValiEpochSeeder(tableName, columnarEngine, &SeederPartitionExoticReverse{
		NumberOfEpochPartitions: noOfEpochPartitions,
		NumberOfValiPartitions:  notOfValiPartitions,
	})
}

func (conf *SeederPartitionExoticReverse) CreateSchema(s *seeding.Seeder) error {
	_, err := db.DB.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			validatorindex BIGINT,
			epoch BIGINT,
			attestations_source_reward BIGINT,
			attestations_target_reward BIGINT,
			attestations_head_reward BIGINT,
			attestations_inactivity_reward BIGINT,
			attestations_inclusion_reward BIGINT,
			attestations_reward BIGINT,
			attestations_ideal_source_reward BIGINT,
			attestations_ideal_target_reward BIGINT,
			attestations_ideal_head_reward BIGINT,
			attestations_ideal_inactivity_reward BIGINT,
			attestations_ideal_inclusion_reward BIGINT,
			attestations_ideal_reward BIGINT,
			blocks_scheduled INTEGER,
			blocks_proposed INTEGER,
			blocks_cl_reward BIGINT,
			blocks_el_reward NUMERIC,
			sync_scheduled INTEGER,
			sync_executed INTEGER,
			sync_rewards BIGINT,
			slashed BOOLEAN,
			balance_start BIGINT,
			balance_end BIGINT,
			deposits_count BIGINT,
			deposits_amount BIGINT,
			withdrawals_count BIGINT,
			withdrawals_amount BIGINT
		) PARTITION BY hash(validatorindex)
	`, s.TableName))
	if err != nil {
		return err
	}

	_, err = db.DB.Exec(fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS %s_validatorindex ON %[1]s (validatorindex, epoch)
	`, s.TableName))
	if err != nil {
		return err
	}

	partRange := int(math.Ceil(float64(s.EpochsInDB) / float64(conf.NumberOfEpochPartitions)))

	for i := 0; i < conf.NumberOfValiPartitions; i++ {
		partName := fmt.Sprintf("%s_v_%d", s.TableName, i)
		partitionCreate := fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %[2]s PARTITION OF %[1]s
				FOR VALUES WITH (MODULUS %[4]d, REMAINDER %[3]d)
				PARTITION BY range(epoch)
		`, s.TableName, partName, i, conf.NumberOfValiPartitions)

		_, err = db.DB.Exec(partitionCreate)
		if err != nil {
			fmt.Printf("here")
			return err
		}

		for j := 0; j < conf.NumberOfEpochPartitions; j++ {
			partitionCreate := fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s_e_%d PARTITION OF %[1]s
				FOR VALUES FROM (%[3]d) TO (%[4]d)
		`, partName, j, j*partRange, (j+1)*partRange)

			_, err = db.DB.Exec(partitionCreate)
			if err != nil {
				fmt.Printf("or here?")
				return err
			}

			// Column engine leaves
			if s.ColumnEngine {
				fmt.Printf("adding column engine to %s_e_%d\n", partName, j)
				err = s.AddToColumnEngine(fmt.Sprintf("%s_e_%d", partName, j), "attestations_head_reward,attestations_source_reward,attestations_target_reward,blocks_cl_reward,epoch,validatorindex")
				if err != nil {
					return err
				}
			}

		}

	}

	return nil
}
