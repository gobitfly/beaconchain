package services

import (
	"database/sql"
	"sync"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// TODO: As a service this will not scale well as it is running once on every instance of the api.
// Instead of service this should be moved to the exporter.

var currentEfficiencyInfo *EfficiencyData
var currentEfficiencyMutex = &sync.RWMutex{}

func (s *Services) startEfficiencyDataService() {
	for {
		startTime := time.Now()
		delay := time.Duration(utils.Config.Chain.ClConfig.SlotsPerEpoch*utils.Config.Chain.ClConfig.SecondsPerSlot) * time.Second
		err := s.updateEfficiencyData() // TODO: only update data if something has changed (new head epoch)
		if err != nil {
			log.Error(err, "error updating average network efficiency data", 0)
			delay = 10 * time.Second
		} else {
			log.Infof("=== average network efficiency data updated in %s", time.Since(startTime))
		}
		utils.ConstantTimeDelay(startTime, delay)
	}
}

func (s *Services) updateEfficiencyData() error {
	var efficiencyInfo *EfficiencyData
	efficiencyMutex := &sync.RWMutex{}

	if currentEfficiencyInfo == nil {
		efficiencyInfo = s.initEfficiencyInfo()
	}

	setEfficiencyData := func(tableName string, period enums.TimePeriod) error {
		var queryResult struct {
			AttestationEfficiency sql.NullFloat64 `db:"attestation_efficiency"`
			ProposerEfficiency    sql.NullFloat64 `db:"proposer_efficiency"`
			SyncEfficiency        sql.NullFloat64 `db:"sync_efficiency"`
		}

		ds := goqu.
			Select(
				goqu.L("COALESCE(SUM(attestations_reward), 0)::decimal / NULLIF(SUM(attestations_ideal_reward)::decimal, 0) AS attestation_efficiency"),
				goqu.L("COALESCE(SUM(blocks_proposed), 0)::decimal / NULLIF(SUM(blocks_scheduled)::decimal, 0) AS proposer_efficiency"),
				goqu.L("COALESCE(SUM(sync_executed), 0)::decimal / NULLIF(SUM(sync_scheduled)::decimal, 0) AS sync_efficiency")).
			From(goqu.T(tableName))

		query, args, err := ds.ToSQL()
		if err != nil {
			return err
		}

		err = s.alloyReader.Get(&queryResult, query, args...)
		if err != nil {
			return err
		}

		efficiencyMutex.Lock()
		efficiencyInfo.AttestationEfficiency[period] = queryResult.AttestationEfficiency
		efficiencyInfo.ProposalEfficiency[period] = queryResult.ProposerEfficiency
		efficiencyInfo.SyncEfficiency[period] = queryResult.SyncEfficiency
		efficiencyMutex.Unlock()

		return nil
	}

	// create waiting group for concurrency
	wg := &errgroup.Group{}

	wg.Go(func() error {
		err := setEfficiencyData("validator_dashboard_data_rolling_total", enums.TimePeriods.AllTime)
		return err
	})
	wg.Go(func() error {
		err := setEfficiencyData("validator_dashboard_data_rolling_daily", enums.TimePeriods.Last24h)
		return err
	})
	wg.Go(func() error {
		err := setEfficiencyData("validator_dashboard_data_rolling_weekly", enums.TimePeriods.Last7d)
		return err
	})
	wg.Go(func() error {
		err := setEfficiencyData("validator_dashboard_data_rolling_monthly", enums.TimePeriods.Last30d)
		return err
	})

	err := wg.Wait()
	if err != nil {
		return err
	}

	// update currentEfficiencyInfo
	currentEfficiencyMutex.Lock()
	if currentEfficiencyInfo == nil { // info on first iteration
		log.Infof("== average network efficiency data updater initialized ==")
	}
	currentEfficiencyInfo = efficiencyInfo
	currentEfficiencyMutex.Unlock()

	return nil
}

// GetCurrentEfficiencyInfo returns the current efficiency info and a function to release the lock
// Call release lock after you are done with accessing the data, otherwise it will block the efficiency service from updating
func (s *Services) GetCurrentEfficiencyInfo() (*EfficiencyData, func(), error) {
	currentEfficiencyMutex.RLock()

	if currentEfficiencyInfo == nil {
		return nil, currentEfficiencyMutex.RUnlock, errors.New("waiting for efficiencyInfo to be initialized")
	}

	return currentEfficiencyInfo, currentEfficiencyMutex.RUnlock, nil
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
