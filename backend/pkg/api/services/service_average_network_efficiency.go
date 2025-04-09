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
	"golang.org/x/sync/errgroup"
)

// TODO: As a service this will not scale well as it is running once on every instance of the api.
// Instead of service this should be moved to the exporter.

var currentEfficiencyInfo atomic.Pointer[EfficiencyData]

func (s *Services) startEfficiencyDataService(wg *sync.WaitGroup) {
	o := sync.Once{}
	for {
		startTime := time.Now()
		delay := time.Duration(utils.Config.Chain.ClConfig.SlotsPerEpoch*utils.Config.Chain.ClConfig.SecondsPerSlot) * time.Second
		statusReporter := services.NewStatusReporter(constants.Event_ApiServiceAvgEfficiency, constants.Default, delay)
		statusReporter.Report(constants.Running, nil)
		err := s.updateEfficiencyData() // TODO: only update data if something has changed (new head epoch)
		if err != nil {
			log.Error(err, "error updating average network efficiency data", 0)
			statusReporter.Report(constants.Failure, map[string]string{"error": err.Error()})
			delay = 10 * time.Second
		} else {
			log.Infof("=== average network efficiency data updated in %s", time.Since(startTime))
			statusReporter.Report(constants.Success, map[string]string{"took": time.Since(startTime).String(), "took_raw": fmt.Sprintf("%v", time.Since(startTime).Milliseconds())})
			o.Do(func() {
				wg.Done()
			})
		}
		utils.ConstantTimeDelay(startTime, delay)
	}
}

func (s *Services) updateEfficiencyData() error {
	efficiencyInfo := s.initEfficiencyInfo()
	efficiencyMutex := &sync.RWMutex{}

	setEfficiencyData := func(period enums.TimePeriod) error {
		tableName, err := period.Table()
		if err != nil {
			return err
		}

		var queryResult struct {
			TotalEfficiency       sql.NullFloat64 `db:"total_efficiency"`
			AttestationEfficiency sql.NullFloat64 `db:"attestation_efficiency"`
			ProposalEfficiency    sql.NullFloat64 `db:"proposal_efficiency"`
			SyncEfficiency        sql.NullFloat64 `db:"sync_efficiency"`
		}

		ds := goqu.Dialect("postgres").
			From(goqu.L(fmt.Sprintf(`%s AS r FINAL`, tableName))).
			Select(
				goqu.L("SUM(efficiency_dividend::decimal) / NULLIF(SUM(efficiency_divisor::decimal), 0)").As("total_efficiency"),
				goqu.L("SUM(efficiency_attestations_dividend::decimal) / NULLIF(SUM(efficiency_attestations_divisor::decimal), 0)").As("attestation_efficiency"),
				goqu.L("SUM(efficiency_proposals_dividend::decimal) / NULLIF(SUM(efficiency_proposals_divisor::decimal), 0)").As("proposal_efficiency"),
				goqu.L("SUM(efficiency_sync_dividend::decimal) / NULLIF(SUM(efficiency_sync_divisor::decimal), 0)").As("sync_efficiency"),
			)

		query, args, err := ds.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %v", err)
		}

		err = s.clickhouseReader.Get(&queryResult, query, args...)
		if err != nil {
			return err
		}

		efficiencyMutex.Lock()
		efficiencyInfo.TotalEfficiency[period] = queryResult.TotalEfficiency
		efficiencyInfo.AttestationEfficiency[period] = queryResult.AttestationEfficiency
		efficiencyInfo.ProposalEfficiency[period] = queryResult.ProposalEfficiency
		efficiencyInfo.SyncEfficiency[period] = queryResult.SyncEfficiency
		efficiencyMutex.Unlock()

		return nil
	}

	// create waiting group for concurrency
	wg := &errgroup.Group{}

	wg.Go(func() error {
		return setEfficiencyData(enums.TimePeriods.Last1h)
	})
	wg.Go(func() error {
		return setEfficiencyData(enums.TimePeriods.Last24h)
	})
	wg.Go(func() error {
		return setEfficiencyData(enums.TimePeriods.Last7d)
	})
	wg.Go(func() error {
		return setEfficiencyData(enums.TimePeriods.Last30d)
	})
	wg.Go(func() error {
		return setEfficiencyData(enums.TimePeriods.AllTime)
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
	efficiencyInfo.TotalEfficiency = make(map[enums.TimePeriod]sql.NullFloat64)
	efficiencyInfo.AttestationEfficiency = make(map[enums.TimePeriod]sql.NullFloat64)
	efficiencyInfo.ProposalEfficiency = make(map[enums.TimePeriod]sql.NullFloat64)
	efficiencyInfo.SyncEfficiency = make(map[enums.TimePeriod]sql.NullFloat64)
	return &efficiencyInfo
}

type EfficiencyData struct {
	TotalEfficiency       map[enums.TimePeriod]sql.NullFloat64 // period -> efficiency
	AttestationEfficiency map[enums.TimePeriod]sql.NullFloat64 // period -> efficiency
	ProposalEfficiency    map[enums.TimePeriod]sql.NullFloat64 // period -> efficiency
	SyncEfficiency        map[enums.TimePeriod]sql.NullFloat64 // period -> efficiency
}
