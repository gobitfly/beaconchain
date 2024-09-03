package services

import (
	"database/sql"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
	"github.com/gobitfly/beaconchain/pkg/monitoring/services"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
)

// TODO: As a service this will not scale well as it is running once on every instance of the api.
// Instead of service this should be moved to the exporter.

var currentEfficiencyInfo atomic.Pointer[EfficiencyData]

func (s *Services) startEfficiencyDataService() {
	for {
		startTime := time.Now()
		delay := time.Duration(utils.Config.Chain.ClConfig.SlotsPerEpoch*utils.Config.Chain.ClConfig.SecondsPerSlot) * time.Second
		r := services.NewStatusReport("api_service_avg_efficiency", constants.Default, delay)
		err := s.updateEfficiencyData() // TODO: only update data if something has changed (new head epoch)
		r(constants.Running, nil)
		if err != nil {
			log.Error(err, "error updating average network efficiency data", 0)
			r(constants.Failure, map[string]string{"error": err.Error()})
			delay = 10 * time.Second
		} else {
			log.Infof("=== average network efficiency data updated in %s", time.Since(startTime))
			r(constants.Success, map[string]string{"took": time.Since(startTime).String()})
		}
		utils.ConstantTimeDelay(startTime, delay)
	}
}

func (s *Services) updateEfficiencyData() error {
	efficiencyInfo := s.initEfficiencyInfo()
	efficiencyMutex := &sync.RWMutex{}

	setEfficiencyData := func(tableName string, period enums.TimePeriod) error {
		var queryResult struct {
			AttestationReward      decimal.Decimal `db:"attestations_reward"`
			AttestationIdealReward decimal.Decimal `db:"attestations_ideal_reward"`
			BlocksProposed         uint64          `db:"blocks_proposed"`
			BlocksScheduled        uint64          `db:"blocks_scheduled"`
			SyncExecuted           uint64          `db:"sync_executed"`
			SyncScheduled          uint64          `db:"sync_scheduled"`
		}

		ds := goqu.Dialect("postgres").
			From(goqu.L(fmt.Sprintf(`%s AS r FINAL`, tableName))).
			Select(
				goqu.L("COALESCE(SUM(r.attestations_reward)::decimal, 0) AS attestations_reward"),
				goqu.L("COALESCE(SUM(r.attestations_ideal_reward)::decimal, 0) AS attestations_ideal_reward"),
				goqu.L("COALESCE(SUM(r.blocks_proposed), 0) AS blocks_proposed"),
				goqu.L("COALESCE(SUM(r.blocks_scheduled), 0) AS blocks_scheduled"),
				goqu.L("COALESCE(SUM(r.sync_executed), 0) AS sync_executed"),
				goqu.L("COALESCE(SUM(r.sync_scheduled), 0) AS sync_scheduled"))

		query, args, err := ds.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %v", err)
		}

		err = s.clickhouseReader.Get(&queryResult, query, args...)
		if err != nil {
			return err
		}

		var attestationEfficiency, proposerEfficiency, syncEfficiency sql.NullFloat64
		if !queryResult.AttestationIdealReward.IsZero() {
			attestationEfficiency.Float64 = queryResult.AttestationReward.Div(queryResult.AttestationIdealReward).InexactFloat64()
			attestationEfficiency.Valid = true
		}
		if queryResult.BlocksScheduled > 0 {
			proposerEfficiency.Float64 = float64(queryResult.BlocksProposed) / float64(queryResult.BlocksScheduled)
			proposerEfficiency.Valid = true
		}
		if queryResult.SyncScheduled > 0 {
			syncEfficiency.Float64 = float64(queryResult.SyncExecuted) / float64(queryResult.SyncScheduled)
			syncEfficiency.Valid = true
		}

		efficiencyMutex.Lock()
		efficiencyInfo.AttestationEfficiency[period] = attestationEfficiency
		efficiencyInfo.ProposalEfficiency[period] = proposerEfficiency
		efficiencyInfo.SyncEfficiency[period] = syncEfficiency
		efficiencyMutex.Unlock()

		return nil
	}

	// create waiting group for concurrency
	wg := &errgroup.Group{}

	wg.Go(func() error {
		err := setEfficiencyData("validator_dashboard_data_rolling_1h", enums.TimePeriods.Last1h)
		return err
	})
	wg.Go(func() error {
		err := setEfficiencyData("validator_dashboard_data_rolling_24h", enums.TimePeriods.Last24h)
		return err
	})
	wg.Go(func() error {
		err := setEfficiencyData("validator_dashboard_data_rolling_7d", enums.TimePeriods.Last7d)
		return err
	})
	wg.Go(func() error {
		err := setEfficiencyData("validator_dashboard_data_rolling_30d", enums.TimePeriods.Last30d)
		return err
	})
	wg.Go(func() error {
		err := setEfficiencyData("validator_dashboard_data_rolling_total", enums.TimePeriods.AllTime)
		return err
	})

	err := wg.Wait()
	if err != nil {
		return err
	}

	// update currentEfficiencyInfo
	if currentEfficiencyInfo.Load() == nil { // info on first iteration
		log.Infof("== average network efficiency data updater initialized ==")
	}
	currentEfficiencyInfo.Store(efficiencyInfo)

	return nil
}

// GetCurrentEfficiencyInfo returns the current efficiency info and a function to release the lock
// Call release lock after you are done with accessing the data, otherwise it will block the efficiency service from updating
func (s *Services) GetCurrentEfficiencyInfo() (*EfficiencyData, error) {
	if currentEfficiencyInfo.Load() == nil {
		return nil, fmt.Errorf("%w: efficiencyInfo", ErrWaiting)
	}

	return currentEfficiencyInfo.Load(), nil
}

func (s *Services) initEfficiencyInfo() *EfficiencyData {
	efficiencyInfo := EfficiencyData{}
	efficiencyInfo.AttestationEfficiency = make(map[enums.TimePeriod]sql.NullFloat64)
	efficiencyInfo.ProposalEfficiency = make(map[enums.TimePeriod]sql.NullFloat64)
	efficiencyInfo.SyncEfficiency = make(map[enums.TimePeriod]sql.NullFloat64)
	return &efficiencyInfo
}

type EfficiencyData struct {
	AttestationEfficiency map[enums.TimePeriod]sql.NullFloat64 // period -> efficiency
	ProposalEfficiency    map[enums.TimePeriod]sql.NullFloat64 // period -> efficiency
	SyncEfficiency        map[enums.TimePeriod]sql.NullFloat64 // period -> efficiency
}
