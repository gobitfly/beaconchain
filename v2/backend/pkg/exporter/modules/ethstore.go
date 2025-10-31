package modules

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/services"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"

	ethstore "github.com/gobitfly/eth.store"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

type EthStoreExporter struct {
	DB             *sqlx.DB
	BNAddress      string
	ENAddress      string
	UpdateInverval time.Duration
	ErrorInterval  time.Duration
	Sleep          time.Duration
}

// start exporting of eth.store into db
func StartEthStoreExporter(bnAddress string, enAddress string, updateInterval, errorInterval, sleepInterval time.Duration, startDayReexport, endDayReexport int64, concurrency, receiptsMode int) {
	log.Infof("starting eth.store exporter")
	ese := &EthStoreExporter{
		DB:             db.WriterDb,
		BNAddress:      bnAddress,
		ENAddress:      enAddress,
		UpdateInverval: updateInterval,
		ErrorInterval:  errorInterval,
		Sleep:          sleepInterval,
	}
	// set sane defaults if config is not set
	if ese.UpdateInverval == 0 {
		ese.UpdateInverval = time.Minute
	}
	if ese.ErrorInterval == 0 {
		ese.ErrorInterval = time.Second * 10
	}
	if ese.Sleep == 0 {
		ese.Sleep = time.Minute
	}

	// Reexport days if specified
	if startDayReexport != -1 && endDayReexport != -1 {
		for day := startDayReexport; day <= endDayReexport; day++ {
			err := ese.reexportDay(strconv.FormatInt(day, 10), concurrency, receiptsMode)
			if err != nil {
				log.Error(err, fmt.Sprintf("error reexporting eth.store day %d in database", day), 0)
				return
			}
		}
		return
	}

	ese.Run(concurrency, receiptsMode)
}

func (ese *EthStoreExporter) reexportDay(day string, concurrency, receiptsMode int) error {
	conn, err := ese.DB.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("error retrieving raw sql connection: %w", err)
	}
	defer conn.Close()

	err = conn.Raw(func(driverConn interface{}) error {
		conn := driverConn.(*stdlib.Conn).Conn()

		pgxdecimal.Register(conn.TypeMap())
		tx, err := conn.Begin(context.Background())
		if err != nil {
			return err
		}
		defer func() {
			err := tx.Rollback(context.Background())
			if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
				log.Error(err, "error rolling back transaction", 0)
			}
		}()

		err = ese.prepareClearDayTx(tx, day)
		if err != nil {
			return err
		}

		err = ese.prepareExportDayTx(tx, day, concurrency, receiptsMode)
		if err != nil {
			return err
		}

		err = tx.Commit(context.Background())
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error during ethstore data export: %w", err)
	}

	return nil
}

func (ese *EthStoreExporter) exportDay(day string, concurrency, receiptsMode int) error {
	conn, err := ese.DB.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("error retrieving raw sql connection: %w", err)
	}
	defer conn.Close()

	err = conn.Raw(func(driverConn interface{}) error {
		conn := driverConn.(*stdlib.Conn).Conn()

		pgxdecimal.Register(conn.TypeMap())
		tx, err := conn.Begin(context.Background())
		if err != nil {
			return err
		}
		defer func() {
			err := tx.Rollback(context.Background())
			if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
				log.Error(err, "error rolling back transaction", 0)
			}
		}()

		err = ese.prepareExportDayTx(tx, day, concurrency, receiptsMode)
		if err != nil {
			return err
		}

		err = tx.Commit(context.Background())
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error during ethsore data export: %w", err)
	}

	return nil
}

func (ese *EthStoreExporter) prepareClearDayTx(tx pgx.Tx, day string) error {
	log.Infof("removing data for day %s", day)
	dayInt, err := strconv.ParseInt(day, 10, 64)
	if err != nil {
		return err
	}
	_, err = tx.Exec(context.Background(), `DELETE FROM eth_store_stats WHERE day = $1`, dayInt)
	return err
}

func (ese *EthStoreExporter) prepareExportDayTx(tx pgx.Tx, day string, concurrency, receiptsMode int) error {
	ethStoreDay, validators, err := ese.getStoreDay(day, concurrency, receiptsMode)
	if err != nil {
		return err
	}

	type EthStoreDayWrapper struct {
		ValidatorIndex uint64
		Data           *ethstore.Day
	}

	validatorsArr := make([]*EthStoreDayWrapper, 0, len(validators))
	for validatorIndex, data := range validators {
		validatorsArr = append(validatorsArr, &EthStoreDayWrapper{
			ValidatorIndex: validatorIndex,
			Data:           data,
		})
	}

	log.Infof("inserting validator data in bulk")
	_, err = tx.CopyFrom(context.Background(), pgx.Identifier{"eth_store_stats"}, []string{
		"day",
		"validator",
		"effective_balances_sum_wei",
		"start_balances_sum_wei",
		"end_balances_sum_wei",
		"deposits_sum_wei",
		"tx_fees_sum_wei",
		"consensus_rewards_sum_wei",
		"total_rewards_wei",
		"apr",
	}, pgx.CopyFromSlice(len(validatorsArr), func(i int) ([]interface{}, error) {
		return []interface{}{
			validatorsArr[i].Data.Day,
			validatorsArr[i].ValidatorIndex,
			validatorsArr[i].Data.EffectiveBalanceGwei.Mul(decimal.NewFromInt(1e9)),
			validatorsArr[i].Data.StartBalanceGwei.Mul(decimal.NewFromInt(1e9)),
			validatorsArr[i].Data.EndBalanceGwei.Mul(decimal.NewFromInt(1e9)),
			validatorsArr[i].Data.DepositsSumGwei.Mul(decimal.NewFromInt(1e9)),
			validatorsArr[i].Data.TxFeesSumWei,
			validatorsArr[i].Data.ConsensusRewardsGwei.Mul(decimal.NewFromInt(1e9)),
			validatorsArr[i].Data.TotalRewardsWei,
			validatorsArr[i].Data.Apr,
		}, nil
	}))

	if err != nil {
		return err
	}

	log.Infof("inserting aggregated day data data")
	_, err = tx.Exec(context.Background(), `
	INSERT INTO eth_store_stats (day, validator, effective_balances_sum_wei, start_balances_sum_wei, end_balances_sum_wei, deposits_sum_wei, tx_fees_sum_wei, consensus_rewards_sum_wei, total_rewards_wei, apr)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		ethStoreDay.Day,
		-1,
		ethStoreDay.EffectiveBalanceGwei.Mul(decimal.NewFromInt(1e9)),
		ethStoreDay.StartBalanceGwei.Mul(decimal.NewFromInt(1e9)),
		ethStoreDay.EndBalanceGwei.Mul(decimal.NewFromInt(1e9)),
		ethStoreDay.DepositsSumGwei.Mul(decimal.NewFromInt(1e9)),
		ethStoreDay.TxFeesSumWei,
		ethStoreDay.ConsensusRewardsGwei.Mul(decimal.NewFromInt(1e9)),
		ethStoreDay.TotalRewardsWei,
		ethStoreDay.Apr,
	)
	if err != nil {
		return err
	}

	log.Infof("inserting historical pool performance data")
	_, err = tx.Exec(context.Background(), `
		insert into historical_pool_performance
		select
			eth_store_stats.day,
			coalesce(validator_pool.pool, 'Unknown'),
			count(*) as validators,
			sum(effective_balances_sum_wei) as effective_balances_sum_wei,
			sum(start_balances_sum_wei) as start_balances_sum_wei,
			sum(end_balances_sum_wei) as end_balances_sum_wei,
			sum(deposits_sum_wei) as deposits_sum_wei,
			sum(tx_fees_sum_wei) as tx_fees_sum_wei,
			sum(consensus_rewards_sum_wei) as consensus_rewards_sum_wei,
			sum(total_rewards_wei) as total_rewards_wei,
			avg(eth_store_stats.apr) as apr
		from validators
		left join validator_pool on validators.pubkey = validator_pool.publickey
		inner join eth_store_stats on validators.validatorindex = eth_store_stats.validator
		where day = $1
		group by validator_pool.pool, eth_store_stats.day
		on conflict (day, pool) do update set
			day                         = excluded.day,
			pool                        = excluded.pool,
			validators                  = excluded.validators,
			effective_balances_sum_wei  = excluded.effective_balances_sum_wei,
			start_balances_sum_wei      = excluded.start_balances_sum_wei,
			end_balances_sum_wei        = excluded.end_balances_sum_wei,
			deposits_sum_wei            = excluded.deposits_sum_wei,
			tx_fees_sum_wei             = excluded.tx_fees_sum_wei,
			consensus_rewards_sum_wei   = excluded.consensus_rewards_sum_wei,
			total_rewards_wei           = excluded.total_rewards_wei,
			apr                         = excluded.apr`,
		ethStoreDay.Day)

	return err
}

func (ese *EthStoreExporter) getStoreDay(day string, concurrency, receiptsMode int) (*ethstore.Day, map[uint64]*ethstore.Day, error) {
	log.Infof("retrieving eth.store for day %v", day)
	return ethstore.Calculate(context.Background(), ese.BNAddress, ese.ENAddress, day, concurrency, receiptsMode)
}

func (ese *EthStoreExporter) Run(concurrency, receiptsMode int) {
	t := time.NewTicker(ese.UpdateInverval)
	defer t.Stop()
DBCHECK:
	for {
		// get latest eth.store day
		latestFinalizedEpoch, err := db.GetLatestFinalizedEpoch()
		if err != nil {
			log.Error(err, "error retrieving latest finalized epoch from db", 0)
			time.Sleep(ese.ErrorInterval)
			continue
		}

		if latestFinalizedEpoch == 0 {
			log.Error(err, "error retrieved 0 as latest finalized epoch from the db", 0)
			time.Sleep(ese.ErrorInterval)
			continue
		}
		latestDay := utils.DayOfSlot(latestFinalizedEpoch*utils.Config.Chain.ClConfig.SlotsPerEpoch) - 1

		log.Infof("latest day is %v", latestDay)
		// count rows of eth.store days in db
		var ethStoreDayCount uint64
		err = ese.DB.Get(&ethStoreDayCount, `
				SELECT COUNT(*)
				FROM eth_store_stats WHERE validator = -1`)
		if err != nil {
			log.Error(err, "error retrieving eth.store days count from db", 0)
			time.Sleep(ese.ErrorInterval)
			continue
		}

		log.Infof("ethStoreDayCount is %v", ethStoreDayCount)

		if ethStoreDayCount <= latestDay {
			// db is incomplete
			// init export map, set every day to true
			daysToExport := make(map[uint64]bool)
			for i := uint64(0); i <= latestDay; i++ {
				daysToExport[i] = true
			}

			// set every existing day in db to false in export map
			if ethStoreDayCount > 0 {
				var ethStoreDays []types.EthStoreDay
				err = ese.DB.Select(&ethStoreDays, `
						SELECT day
						FROM eth_store_stats WHERE validator = -1`)
				if err != nil {
					log.Error(err, "error retrieving eth.store days from db", 0)
					time.Sleep(ese.ErrorInterval)
					continue
				}
				for _, ethStoreDay := range ethStoreDays {
					daysToExport[ethStoreDay.Day] = false
				}
			}
			daysToExportArray := make([]uint64, 0, len(daysToExport))
			for dayToExport, shouldExport := range daysToExport {
				if shouldExport {
					daysToExportArray = append(daysToExportArray, dayToExport)
				}
			}

			sort.Slice(daysToExportArray, func(i, j int) bool {
				return daysToExportArray[i] > daysToExportArray[j]
			})
			// export missing days
			for _, dayToExport := range daysToExportArray {
				err = ese.exportDay(strconv.FormatUint(dayToExport, 10), concurrency, receiptsMode)
				if err != nil {
					log.Error(err, fmt.Sprintf("error exporting eth.store day %d into database", dayToExport), 0)
					time.Sleep(ese.ErrorInterval)
					continue DBCHECK
				}
				log.Infof("exported eth.store day %d into db", dayToExport)
				if ethStoreDayCount < latestDay {
					// more than 1 day is being exported, sleep for duration specified in config
					time.Sleep(ese.Sleep)
				}
			}
		}

		services.ReportStatus("ethstoreExporter", "Running", nil)
		<-t.C
	}
}
