package modules

import (
	"database/sql"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/consapi/network"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	edb "github.com/gobitfly/beaconchain/pkg/exporter/db"
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/go-bitfield"
	"golang.org/x/sync/errgroup"
)

// -------------- DEBUG FLAGS ----------------
const debugAggregateMidEveryEpoch = true                             // prod: false
const debugTargetBackfillEpoch = uint64(0)                           // prod: 0
const debugSetBackfillCompleted = true                               // prod: true
const debugSkipOldEpochClear = false                                 // prod: false
const debugAddToColumnEngine = false                                 // prod: true?
const debugAggregateRollingWindowsDuringBackfillUTCBoundEpoch = true // prod: true

// ----------- END OF DEBUG FLAGS ------------

const epochFetchParallelism = 6
const epochWriteParallelism = 3
const databaseAggregationParallelism = 3

type dashboardData struct {
	ModuleContext
	log               ModuleLog
	signingDomain     []byte
	epochWriter       *epochWriter
	epochToTotal      *epochToTotalAggregator
	epochToHour       *epochToHourAggregator
	hourToDay         *hourToDayAggregator
	dayUp             *dayUpAggregator
	headEpochQueue    chan uint64
	backFillCompleted bool
}

func NewDashboardDataModule(moduleContext ModuleContext) ModuleInterface {
	temp := &dashboardData{
		ModuleContext: moduleContext,
	}
	temp.log = ModuleLog{module: temp}
	temp.epochWriter = newEpochWriter(temp)
	temp.epochToTotal = newEpochToTotalAggregator(temp)
	temp.epochToHour = newEpochToHourAggregator(temp)
	temp.hourToDay = newHourToDayAggregator(temp)
	temp.dayUp = newDayUpAggregator(temp)
	temp.headEpochQueue = make(chan uint64, 100)
	temp.backFillCompleted = false
	return temp
}

func (d *dashboardData) Init() error {
	go func() {
		_, err := db.AlloyWriter.Exec("SET work_mem TO '128MB';")
		if err != nil {
			d.log.Fatal(err, "failed to set work_mem", 0)
		}
		for {
			var upToEpochPtr *uint64 = nil    // nil will backfill back to head
			if debugTargetBackfillEpoch > 0 { // todo remove
				upToEpoch := debugTargetBackfillEpoch
				upToEpochPtr = &upToEpoch
			}

			done, err := d.backfillHeadEpochData(upToEpochPtr)
			if err != nil {
				d.log.Fatal(err, "failed to backfill epoch data", 0)
			}

			if done {
				d.log.Infof("dashboard data up to date, starting head export")
				if debugSetBackfillCompleted { // todo remove
					utils.SendMessage(fmt.Sprintf("🎉🎉🎉 v2 Dashboard %s - Reached head, exporting from head now", utils.Config.Chain.Name), &utils.Config.InternalAlerts)
					d.backFillCompleted = true
				}
				break
			}
			time.Sleep(1 * time.Second)
		}

		d.processHeadQueue()
	}()

	return nil
}

func (d *dashboardData) processHeadQueue() {
	for {
		epoch := <-d.headEpochQueue

		startTime := time.Now()
		d.log.Infof("exporting dashboard epoch data for epoch %d", epoch)
		stage := 0
		for { // retry this epoch until no errors occur
			currentFinalizedEpoch, err := d.CL.GetFinalityCheckpoints("finalized")
			if err != nil {
				d.log.Error(err, "failed to get finalized checkpoint", 0)
				time.Sleep(time.Second * 10)
				continue
			}

			if stage <= 0 {
				targetEpoch := epoch - 1
				_, err := d.backfillHeadEpochData(&targetEpoch)
				if err != nil {
					d.log.Error(err, "failed to backfill head epoch data", 0, map[string]interface{}{"epoch": epoch})
					time.Sleep(time.Second * 10)
					continue
				}
				stage = 1
			}

			if stage <= 1 {
				err := d.exportEpochAndTails(epoch)
				if err != nil {
					d.log.Error(err, "failed to backfill tail epoch data", 0, map[string]interface{}{"epoch": epoch})
					time.Sleep(time.Second * 10)
					continue
				}
				stage = 2
			}

			if stage <= 2 {
				err := d.aggregatePerEpoch(true, debugAggregateMidEveryEpoch || currentFinalizedEpoch.Data.Finalized.Epoch <= epoch+1, false)
				if err != nil {
					d.log.Error(err, "failed to aggregate", 0, map[string]interface{}{"epoch": epoch})
					time.Sleep(time.Second * 10)
					continue
				}
				stage = 3
			}

			break
		}

		d.log.Infof("completed dashboard epoch data for epoch %d in %v", epoch, time.Since(startTime))
	}
}

// returns epochs between start and end that are missing in the database, arguments are inclusive
func getMissingEpochsBetween(start, end int64) ([]uint64, error) {
	if end < start {
		return nil, nil
	}
	missingEpochs := make([]uint64, 0)
	for epoch := start; epoch <= end; epoch++ {
		hasEpoch, err := edb.HasDashboardDataForEpoch(uint64(epoch))
		if err != nil {
			return nil, errors.Wrap(err, "failed to get epoch")
		}
		if !hasEpoch {
			missingEpochs = append(missingEpochs, uint64(epoch))
		}
	}
	return missingEpochs, nil
}

// exports the provided headEpoch plus any tail epochs that are needed for rolling aggregation
// fE a tail epoch for rolling 1 day aggregation (225 epochs) for head 227 on ethereum would correspond to two tail epochs [0,1]
func (d *dashboardData) exportEpochAndTails(headEpoch uint64) error {
	// for 24h aggregation
	missingTails, err := d.hourToDay.getMissingRolling24TailEpochs(headEpoch)
	if err != nil {
		return errors.Wrap(err, "failed to get missing 24h tail epochs")
	}

	d.log.Infof("missing 24h: %v", missingTails)

	// day aggregation
	daysMissingTails, err := d.dayUp.getMissingRollingDayTailEpochs(headEpoch)
	if err != nil {
		return errors.Wrap(err, "failed to get missing day tail epochs")
	}

	dayMissingHeads, err := d.dayUp.getMissingRollingDayHeadEpochs(headEpoch)
	if err != nil {
		return errors.Wrap(err, "failed to get missing day head epochs")
	}

	// merge
	missingTails = append(missingTails, deduplicate(append(daysMissingTails, dayMissingHeads...))...)

	if len(missingTails) > 10 {
		d.log.Infof("This might take a bit longer than usual as exporter is catching up quite a lot old epochs, usually happens after downtime or after initial sync")
	}

	// sort asc
	sort.Slice(missingTails, func(i, j int) bool {
		return missingTails[i] < missingTails[j]
	})

	hasHeadAlreadyExported, err := edb.HasDashboardDataForEpoch(headEpoch)
	if err != nil {
		return errors.Wrap(err, "failed to check if head epoch has dashboard data")
	}

	// append head
	if !hasHeadAlreadyExported {
		missingTails = append(missingTails, headEpoch)
		d.log.Infof("fetch missing tail/head epochs: %v | fetch head: %d", missingTails, headEpoch)
	} else {
		d.log.Infof("fetch missing tail/head epochs: %v | fetch head: -", missingTails)
	}

	var nextDataChan chan []DataEpoch = make(chan []DataEpoch, 1)
	go func() {
		d.epochDataFetcher(missingTails, 0, epochFetchParallelism, nextDataChan)
	}()

	for {
		datas := <-nextDataChan

		d.writeEpochDatas(datas)

		// has written last entry in gaps
		if containsEpoch(datas, missingTails[len(missingTails)-1]) {
			break
		}
	}

	d.log.Infof("backfilling tail epochs for aggregation finished")

	return nil
}

// fetches and processes epoch data and provides them via the nextDataChan
// expects ordered epochs in ascending order
func (d *dashboardData) epochDataFetcher(epochs []uint64, epochTailCutOff uint64, epochFetchParallelism int, nextDataChan chan []DataEpoch) {
	// group epochs into parallel worker groups
	groups := getEpochParallelGroups(epochs, epochFetchParallelism)
	numberOfEpochsToFetch := len(epochs)
	epochsFetched := 0

	for _, gapGroup := range groups {
		errGroup := &errgroup.Group{}

		datas := make([]DataEpoch, 0, epochFetchParallelism)
		start := time.Now()

		// Step 1: fetch epoch data raw
		for _, gap := range gapGroup.Epochs {
			gap := gap
			errGroup.Go(func() error {
				for {
					//backfill if needed, skip backfilling older than RetainEpochDuration
					if gap < epochTailCutOff {
						continue
					}

					// just in case we ask again before exporting since some time may have been passed
					hasEpoch, err := edb.HasDashboardDataForEpoch(gap)
					if err != nil {
						d.log.Error(err, "failed to check if epoch has dashboard data", 0, map[string]interface{}{"epoch": gap})
						time.Sleep(time.Second * 10)
						continue
					}
					if hasEpoch {
						time.Sleep(time.Second * 1)
						continue
					}

					d.log.Infof("epoch data fetcher, retrieving data for epoch %d", gap)

					// for sequential improve performance by skipping some calls for all epochs that are not the start epoch
					// and provide the startBalance and the missed slots for every following epoch in this sequence
					// with the data from the previous epoch. Do not do this for non sequential working groups
					data, err := d.GetEpochDataRaw(gap, gap != gapGroup.Epochs[0] && gapGroup.Sequential)
					if err != nil {
						d.log.Error(err, "failed to get epoch data", 0, map[string]interface{}{"epoch": gap})
						time.Sleep(time.Second * 10)
						continue
					}

					datas = append(datas, DataEpoch{
						DataRaw: data,
						Epoch:   gap,
					})

					break
				}
				return nil
			})
		}

		_ = errGroup.Wait() // no need to catch error since it will retry unless all clear without errors

		// sort datas first, epoch asc
		sort.Slice(datas, func(i, j int) bool {
			return datas[i].Epoch < datas[j].Epoch
		})

		// Half Step: for sequential since we used skipSerialCalls we must provide startBalance and missedSlots for every epoch except FromEpoch
		// with the data from the previous epoch. This is done to improve performance by skipping some calls
		if gapGroup.Sequential {
			// provide data from the previous epochs
			for i := 1; i < len(datas); i++ {
				datas[i].DataRaw.startBalances = datas[i-1].DataRaw.endBalances

				for slot := range datas[i-1].DataRaw.missedslots {
					datas[i].DataRaw.missedslots[slot] = true
				}
			}
		}

		// Step 2: process data
		errGroup = &errgroup.Group{}
		errGroup.SetLimit(epochWriteParallelism / 2) // mitigate short ram spike
		for i := 0; i < len(datas); i++ {
			i := i
			errGroup.Go(func() error {
				for {
					d.log.Infof("epoch data fetcher, processing data for epoch %d", datas[i].Epoch)
					start := time.Now()

					result, err := d.ProcessEpochData(datas[i].DataRaw, d.signingDomain)
					if err != nil {
						d.log.Error(err, "failed to process epoch data", 0, map[string]interface{}{"epoch": datas[i].Epoch})
						time.Sleep(time.Second * 10)
						continue
					}

					datas[i].Data = result
					datas[i].DataRaw = nil // clear raw data, not needed any more

					d.log.Infof("epoch data fetcher, processed data for epoch %d in %v", datas[i].Epoch, time.Since(start))
					break
				}
				return nil
			})
		}

		_ = errGroup.Wait() // no need to catch error since it will retry unless all clear without errors

		epochsFetched += len(datas)
		remaining := numberOfEpochsToFetch - epochsFetched
		remainingTimeEst := time.Duration(0)
		if remaining > 0 {
			remainingTimeEst = time.Duration(time.Since(start).Nanoseconds() / int64(len(datas)) * int64(remaining))
		}

		d.log.Infof("epoch data fetcher, fetched %v epochs %v in %v. Remaining: %v (%v)", len(datas), gapGroup.Epochs, time.Since(start), remaining, remainingTimeEst)

		nextDataChan <- datas
	}
}

// breaks epoch down in groups of size parallelism
// for example epochs: 1,2,3,5,7,9,10,11,12,13,20,30 with parallelism = 4 breaks down to groups:
// 0 = 1,2,3 sequential true
// 1 = 5,7 sequential false
// 2 = 9,10,11,12 sequential true
// 3 = 13,20,30 sequential false
// expects ordered epochs in ascending order
func getEpochParallelGroups(epochs []uint64, parallelism int) []EpochParallelGroup {
	parallelGroups := make([]EpochParallelGroup, 0, len(epochs)/4)

	// 1. Group sequential epochs in parallel groups
	for i := 0; i < len(epochs); i++ {
		group := EpochParallelGroup{
			Epochs:     []uint64{epochs[i]},
			Sequential: true,
		}

		var j int
		for j = 1; j < parallelism && i+j < len(epochs); j++ {
			if group.Epochs[len(group.Epochs)-1]+1 == epochs[i+j] { // only group sequential epochs
				group.Epochs = append(group.Epochs, epochs[i+j])
			} else {
				break
			}
		}
		i += j - 1

		parallelGroups = append(parallelGroups, group)
	}

	groups := make([]EpochParallelGroup, 0)

	// 2. Find non sequential (len(Epochs) == 1) epochs and group them as non sequential groups
	for i := 0; i < len(parallelGroups); i++ {
		group := parallelGroups[i]

		if len(group.Epochs) == 1 {
			var j int
			for j = 1; j < parallelism && i+j < len(parallelGroups); j++ {
				nextGroup := parallelGroups[i+j]
				if len(nextGroup.Epochs) == 1 {
					group.Sequential = false
					group.Epochs = append(group.Epochs, nextGroup.Epochs[0])
				} else {
					break
				}
			}
			i += j - 1
		}

		groups = append(groups, group)
	}

	return groups
}

// can be used to start a backfill up to epoch
// returns true if there was nothing to backfill, otherwise returns false
// if upToEpoch is nil, it will backfill until the latest finalized epoch
func (d *dashboardData) backfillHeadEpochData(upToEpoch *uint64) (bool, error) {
	if upToEpoch == nil {
		res, err := d.CL.GetFinalityCheckpoints("finalized")
		if err != nil {
			return false, errors.Wrap(err, "failed to get finalized checkpoint")
		}
		if utils.IsByteArrayAllZero(res.Data.Finalized.Root) {
			return false, errors.New("network not finalized yet")
		}
		upToEpoch = &res.Data.Finalized.Epoch
	}

	latestExportedEpoch, err := edb.GetLatestDashboardEpoch()
	if err != nil {
		return false, errors.Wrap(err, "failed to get latest dashboard epoch")
	}

	gaps, err := edb.GetDashboardEpochGapsBetween(*upToEpoch, int64(latestExportedEpoch))
	if err != nil {
		return false, errors.Wrap(err, "failed to get epoch gaps")
	}

	if len(gaps) > 0 {
		if latestExportedEpoch > 0 {
			err = d.aggregatePerEpoch(true, false, true)
			if err != nil {
				return false, errors.Wrap(err, "failed to aggregate")
			}
		}

		d.log.Infof("Epoch dashboard data %d gaps found, backfilling gaps in the range fom epoch %d to %d", len(gaps), gaps[0], gaps[len(gaps)-1])

		cutOff := latestExportedEpoch - d.epochWriter.getRetentionEpochDuration()
		if d.epochWriter.getRetentionEpochDuration() > latestExportedEpoch {
			cutOff = 0
		}
		var nextDataChan chan []DataEpoch = make(chan []DataEpoch, 1)
		go func() {
			d.epochDataFetcher(gaps, cutOff, epochFetchParallelism, nextDataChan)
		}()

		for {
			d.log.Info("storage waiting for data from fetcher")
			datas := <-nextDataChan

			done := containsEpoch(datas, gaps[len(gaps)-1])
			lastEpoch := datas[len(datas)-1].Epoch

			// If one epoch containing an UTC boundary epoch we do a bootstrap aggregation for the rolling windows
			// since this an exact utc bounds (and hence also an hour bounds) we do not need any tail epochs to calculate the rolling window
			if debugAggregateRollingWindowsDuringBackfillUTCBoundEpoch {
				for i, data := range datas {
					if _, utcEndBound := getDayAggregateBounds(data.Epoch); utcEndBound-1 == data.Epoch {
						d.log.Infof("BOOTSTRAPPING ROLLING WINDOWS! Epoch %d contains an UTC boundary epoch", data.Epoch)
						// write subset
						d.writeEpochDatas(datas[:i+1])
						for {
							err = d.aggregatePerEpoch(true, true, false)
							if err != nil {
								d.log.Error(err, "backfill, failed to aggregate", 0, map[string]interface{}{"epoch start": datas[0].Epoch, "epoch end": datas[len(datas)-1].Epoch})
								time.Sleep(time.Second * 10)
								continue
							}
							break
						}

						{
							utcDayStart, _ := getDayAggregateBounds(data.Epoch - 1)
							utcDay := utils.EpochToTime(utcDayStart).Format("2006-01-02")
							utils.SendMessage(fmt.Sprintf("🗡🧙‍♂️ v2 Dashboard %s - Completed UTC day `%s` & Updated Rolling tables (24h, 7d, 30d, 90d) to epoch %v", utils.Config.Chain.Name, utcDay, data.Epoch), &utils.Config.InternalAlerts)
						}

						if len(datas) > i+1 {
							datas = datas[i+1:]
						} else {
							datas = nil
						}
						break
					}
				}
			}

			if len(datas) > 0 {
				d.log.Info("storage got data, writing epoch data")
				d.writeEpochDatas(datas)

				d.log.Info("storage writing done, aggregate")
				for {
					err = d.aggregatePerEpoch(false, false, false)
					if err != nil {
						d.log.Error(err, "backfill, failed to aggregate", 0, map[string]interface{}{"epoch start": datas[0].Epoch, "epoch end": lastEpoch})
						time.Sleep(time.Second * 10)
						continue
					}
					break
				}
				d.log.InfoWithFields(map[string]interface{}{"epoch start": datas[0].Epoch, "epoch end": lastEpoch}, "backfill, aggregated epoch data")
			}

			if lastEpoch%225 < epochFetchParallelism {
				upToEpoch := *upToEpoch
				utils.SendMessage(fmt.Sprintf("<:stonks:820252887094394901> v2 Dashboard %s - Epoch progress %d/%d [%.2f%%]", utils.Config.Chain.Name, lastEpoch, upToEpoch, float64(lastEpoch*100)/float64(upToEpoch)), &utils.Config.InternalAlerts)
			}

			// has written last entry in gaps
			if done {
				break
			}
		}
	}
	return true, nil
}

func containsEpoch(d []DataEpoch, epoch uint64) bool {
	for i := 0; i < len(d); i++ {
		if d[i].Epoch == epoch {
			return true
		}
	}
	return false
}

// stores all passed epoch data, blocks until all data is written without error
func (d *dashboardData) writeEpochDatas(datas []DataEpoch) {
	errGroup := &errgroup.Group{}
	errGroup.SetLimit(epochWriteParallelism)
	for i := 0; i < len(datas); i++ {
		data := datas[i]
		errGroup.Go(func() error {
			for {
				d.log.Infof("storage, writing epoch data for epoch %v", data.Epoch)
				start := time.Now()

				// retry this epoch until no errors occur
				err := d.epochWriter.WriteEpochData(data.Epoch, data.Data)
				if err != nil {
					d.log.Error(err, "storage, failed to write epoch data", 0, map[string]interface{}{"epoch": data.Epoch})
					time.Sleep(time.Second * 10)
					continue
				}

				d.log.Infof("storage, wrote epoch data %d in %v", data.Epoch, time.Since(start))

				break
			}
			return nil
		})
	}

	_ = errGroup.Wait() // no errors to handle since it will retry until it resolves without err
}

var lastExportedHour uint64 = ^uint64(0)

// Contains all aggregation logic that should happen for every new exported epoch
// forceAggregate triggers an aggregation, use this when calling on head.
// updateRollingWindows specifies whether we should update rolling windows
func (d *dashboardData) aggregatePerEpoch(forceAggregate bool, updateRollingWindows bool, preventClearOldEpochs bool) error {
	currentExportedEpoch, err := edb.GetLatestDashboardEpoch()
	if err != nil {
		return errors.Wrap(err, "failed to get last exported epoch")
	}
	currentStartBound, _ := getHourAggregateBounds(currentExportedEpoch)

	// Performance improvement for backfilling, no need to aggregate day after each epoch, we can update once per hour
	if forceAggregate || currentStartBound != lastExportedHour {
		start := time.Now()
		defer func() {
			d.log.Infof("all of epoch based aggregation took %v", time.Since(start))
		}()

		// important to do this before hour aggregate as hour aggregate deletes old epochs
		errGroup := &errgroup.Group{}
		errGroup.SetLimit(databaseAggregationParallelism)
		errGroup.Go(func() error {
			err := d.epochToTotal.aggregateTotal(currentExportedEpoch)
			if err != nil {
				return errors.Wrap(err, "failed to aggregate total")
			}
			return nil
		})

		// below: so this could be parallel IF we dont need to bootstrap. Room for improvement?
		errGroup.Go(func() error { // can run in parallel with aggregateRollingWindows as long as no bootstrap is required, otherwise must be sequential
			err := d.epochToHour.aggregate1h(currentExportedEpoch) // will aggregate last hour too if it hasn't completed yet
			if err != nil {
				return errors.Wrap(err, "failed to aggregate 1h")
			}
			return nil
		})

		errGroup.Go(func() error { // can run in parallel with aggregateRollingWindowsas long as no bootstrap is required, otherwise must be sequential
			err := d.hourToDay.dayAggregate(currentExportedEpoch)
			if err != nil {
				return errors.Wrap(err, "failed to aggregate day")
			}
			return nil
		})

		err = errGroup.Wait()
		if err != nil {
			return errors.Wrap(err, "failed to aggregate")
		}

		lastExportedHour = currentStartBound

		if updateRollingWindows {
			// todo you could add it to the err group above IF no bootstrap is needed.

			err = d.aggregateRollingWindows(currentExportedEpoch)
			if err != nil {
				return errors.Wrap(err, "failed to aggregate rolling windows")
			}
		}

		d.log.Infof("cleaning old epochs")

		if !preventClearOldEpochs {
			err = d.epochWriter.clearOldEpochs(int64(currentExportedEpoch - d.epochWriter.getRetentionEpochDuration()))
			if err != nil {
				return errors.Wrap(err, "failed to clear old epochs")
			}
		}

		// clear old hourly aggregated epochs, do not remove epochs from epoch table here as these are needed for Mid aggregation
		err = d.epochToHour.clearOldHourAggregations(int64(currentExportedEpoch - d.epochToHour.getHourRetentionDurationEpochs()))
		if err != nil {
			return errors.Wrap(err, "failed to clear old hours")
		}
	}

	return nil
}

// This function contains more heavy aggregation like rolling 7d, 30d, 90d
// This function assumes that epoch aggregation is finished before calling THOUGH it could run in parallel
// as long as the rolling tables do not require a bootstrap
func (d *dashboardData) aggregateRollingWindows(currentExportedEpoch uint64) error {
	start := time.Now()
	defer func() {
		d.log.Infof("all of mid aggregation took %v", time.Since(start))
	}()

	errGroup := &errgroup.Group{}
	errGroup.SetLimit(databaseAggregationParallelism)

	errGroup.Go(func() error {
		err := d.hourToDay.rolling24hAggregate(currentExportedEpoch)
		if err != nil {
			return errors.Wrap(err, "failed to rolling 24h aggregate")
		}
		d.log.Infof("finished dayAggregate rolling 24h")
		return nil
	})

	errGroup.Go(func() error {
		err := d.dayUp.rolling7dAggregate(currentExportedEpoch)
		if err != nil {
			return errors.Wrap(err, "failed to aggregate 7d")
		}
		return nil
	})

	errGroup.Go(func() error {
		err := d.dayUp.rolling30dAggregate(currentExportedEpoch)
		if err != nil {
			return errors.Wrap(err, "failed to aggregate 30d")
		}
		return nil
	})

	errGroup.Go(func() error {
		err := d.dayUp.rolling90dAggregate(currentExportedEpoch)
		if err != nil {
			return errors.Wrap(err, "failed to aggregate 90d")
		}
		return nil
	})

	err := errGroup.Wait()
	if err != nil {
		return errors.Wrap(err, "failed to aggregate")
	}

	return nil
}

func (d *dashboardData) OnFinalizedCheckpoint(_ *constypes.StandardFinalizedCheckpointResponse) error {
	if !d.backFillCompleted {
		return nil // nothing to do, wait for backfill to finish
	}

	// Note that "StandardFinalizedCheckpointResponse" event contains the current justified epoch, not the finalized one
	// An epoch becomes finalized once the next epoch gets justified
	// Hence we just listen for new justified epochs here and fetch the latest finalized one from the node
	// Do not assume event.Epoch -1 is finalized by default as it could be that it is not justified
	res, err := d.CL.GetFinalityCheckpoints("finalized")
	if err != nil {
		return err
	}

	latestExported, err := edb.GetLatestDashboardEpoch()
	if err != nil {
		return err
	}

	if latestExported != 0 {
		if res.Data.Finalized.Epoch <= latestExported {
			d.log.Infof("dashboard epoch data already exported for epoch %d", res.Data.Finalized.Epoch)
			return nil
		}
	}

	d.headEpochQueue <- res.Data.Finalized.Epoch

	return nil
}

func (d *dashboardData) GetName() string {
	return "Dashboard-Data"
}

func (d *dashboardData) OnHead(event *constypes.StandardEventHeadResponse) error {
	return nil
}

func (d *dashboardData) OnChainReorg(event *constypes.StandardEventChainReorg) error {
	return nil
}

func (d *dashboardData) GetEpochDataRaw(epoch uint64, skipSerialCalls bool) (*Data, error) {
	data, err := d.getData(epoch, utils.Config.Chain.ClConfig.SlotsPerEpoch, skipSerialCalls)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, errors.New("can not get data")
	}

	return data, nil
}

func (d *dashboardData) ProcessEpochData(data *Data, domain []byte) ([]*validatorDashboardDataRow, error) {
	if d.signingDomain == nil {
		domain, err := utils.GetSigningDomain()
		if err != nil {
			return nil, err
		}
		d.signingDomain = domain
	}

	return d.process(data, domain)
}

type Data struct {
	startBalances            *constypes.StandardValidatorsResponse
	endBalances              *constypes.StandardValidatorsResponse
	proposerAssignments      *constypes.StandardProposerAssignmentsResponse
	syncCommitteeAssignments *constypes.StandardSyncCommitteesResponse
	attestationRewards       *constypes.StandardAttestationRewardsResponse
	beaconBlockData          map[uint64]*constypes.StandardBeaconSlotResponse
	beaconBlockRewardData    map[uint64]*constypes.StandardBlockRewardsResponse
	syncCommitteeRewardData  map[uint64]*constypes.StandardSyncCommitteeRewardsResponse
	attestationAssignments   map[uint64]uint32
	missedslots              map[uint64]bool
	genesis                  bool
}

// Data for a single validator
// use skipSerialCalls = false if you are not sure what you are doing. This flag is mainly
// to gain performance improvements when exporting a couple sequential epochs in a row
func (d *dashboardData) getData(epoch, slotsPerEpoch uint64, skipSerialCalls bool) (*Data, error) {
	var result Data
	var err error

	firstSlotOfEpoch := epoch * slotsPerEpoch

	firstSlotOfPreviousEpoch := int64(firstSlotOfEpoch) - 1
	lastSlotOfEpoch := firstSlotOfEpoch + slotsPerEpoch - 1

	result.beaconBlockData = make(map[uint64]*constypes.StandardBeaconSlotResponse, slotsPerEpoch)
	result.beaconBlockRewardData = make(map[uint64]*constypes.StandardBlockRewardsResponse, slotsPerEpoch)
	result.syncCommitteeRewardData = make(map[uint64]*constypes.StandardSyncCommitteeRewardsResponse, slotsPerEpoch)
	result.attestationAssignments = make(map[uint64]uint32)
	result.missedslots = make(map[uint64]bool, slotsPerEpoch*2)

	cl := d.CL

	errGroup := &errgroup.Group{}
	errGroup.SetLimit(5)

	totalStart := time.Now()

	errGroup.Go(func() error {
		// retrieve proposer assignments for the epoch in order to attribute missed slots
		start := time.Now()
		result.proposerAssignments, err = cl.GetPropoalAssignments(epoch)
		if err != nil {
			d.log.Error(err, "can not get proposer assignments", 0, map[string]interface{}{"epoch": epoch})
			return err
		}
		d.log.Debugf("retrieved proposer assignments in %v", time.Since(start))
		return nil
	})

	errGroup.Go(func() error {
		// retrieve sync committee assignments for the epoch in order to attribute missed sync assignments
		start := time.Now()
		result.syncCommitteeAssignments, err = cl.GetSyncCommitteesAssignments(epoch, int64(firstSlotOfEpoch))
		if err != nil {
			d.log.Error(err, "can not get sync committee assignments", 0, map[string]interface{}{"epoch": epoch})
			return err
		}
		d.log.Debugf("retrieved sync committee assignments in %v", time.Since(start))
		return nil
	})

	// As of dencun you can attest up until the end of the following epoch
	min := lastSlotOfEpoch - utils.Config.Chain.ClConfig.SlotsPerEpoch
	if lastSlotOfEpoch < utils.Config.Chain.ClConfig.SlotsPerEpoch {
		min = 0
	}

	aaMutex := &sync.Mutex{}
	// executes twice, one with lastSlotOf this epoch and then lastSlotOf last epoch
	for slot := lastSlotOfEpoch; slot >= min; slot -= utils.Config.Chain.ClConfig.SlotsPerEpoch {
		slot := slot
		errGroup.Go(func() error {
			data, err := cl.GetCommittees(slot, nil, nil, nil)
			if err != nil {
				d.log.Error(err, "can not get attestation assignments", 0, map[string]interface{}{"slot": slot})
				return err
			}

			for _, committee := range data.Data {
				for i, valIndex := range committee.Validators {
					k := utils.FormatAttestorAssignmentKeyLowMem(committee.Slot, uint16(committee.Index), uint32(i))
					aaMutex.Lock()
					result.attestationAssignments[k] = uint32(valIndex)
					aaMutex.Unlock()
				}
			}
			return nil
		})
		if slot < utils.Config.Chain.ClConfig.SlotsPerEpoch {
			break // special case for epoch 0
		}
	}

	errGroup.Go(func() error {
		// attestation rewards
		start := time.Now()
		result.attestationRewards, err = cl.GetAttestationRewards(epoch)
		if err != nil {
			d.log.Error(err, "can not get attestation rewards", 0, map[string]interface{}{"epoch": epoch})
			return err
		}
		d.log.Debugf("retrieved attestation rewards data in %v", time.Since(start))
		return nil
	})

	errGroup.Go(func() error {
		// retrieve the validator balances at the end of the epoch
		start := time.Now()
		result.endBalances, err = cl.GetValidators(lastSlotOfEpoch, nil, nil)
		if err != nil {
			d.log.Error(err, "can not get validators balances", 0, map[string]interface{}{"lastSlotOfEpoch": lastSlotOfEpoch})
			return err
		}
		d.log.Debugf("retrieved end balances using state at slot %d in %v", lastSlotOfEpoch, time.Since(start))
		return nil
	})

	// if this flag is used the caller must provide startBalance from the previous epoch themselves
	// as well as providing the missedslots data from the previous epoch
	if !skipSerialCalls {
		errGroup.Go(func() error {
			// retrieve the validator balances at the start of the epoch
			start := time.Now()
			if firstSlotOfPreviousEpoch < 0 {
				result.startBalances, err = d.CL.GetValidators("genesis", nil, nil)
				result.genesis = true
			} else {
				result.startBalances, err = d.CL.GetValidators(firstSlotOfPreviousEpoch, nil, nil)
			}
			if err != nil {
				d.log.Error(err, "can not get validators balances", 0, map[string]interface{}{"firstSlotOfPreviousEpoch": firstSlotOfPreviousEpoch})
				return err
			}
			d.log.Debugf("retrieved start balances using state at slot %d in %v", firstSlotOfPreviousEpoch, time.Since(start))
			return nil
		})

		errGroup.Go(func() error {
			start := time.Now()

			if firstSlotOfEpoch > slotsPerEpoch { // handle case for first epoch
				// get missed slots of last epoch for optimal inclusion distance
				for slot := firstSlotOfEpoch - slotsPerEpoch; slot <= lastSlotOfEpoch-slotsPerEpoch; slot++ {
					_, err := cl.GetBlockHeader(slot)
					if err != nil {
						httpErr, _ := network.SpecificError(err)
						if httpErr != nil && httpErr.StatusCode == 404 {
							result.missedslots[slot] = true
							continue // missed
						}
						d.log.Fatal(err, "can not get block data", 0, map[string]interface{}{"slot": slot})
						continue
					}
				}
			}
			d.log.Debugf("retrieved missed slots in %v", time.Since(start))
			return nil
		})
	}

	mutex := &sync.Mutex{}
	for slot := firstSlotOfEpoch; slot <= lastSlotOfEpoch; slot++ {
		slot := slot
		errGroup.Go(func() error {
			// retrieve the data for all blocks that were proposed in this epoch
			block, err := cl.GetSlot(slot)
			if err != nil {
				httpErr, _ := network.SpecificError(err)
				if httpErr != nil && httpErr.StatusCode == 404 {
					mutex.Lock()
					result.missedslots[slot] = true
					mutex.Unlock()
					return nil // missed
				}

				return err
			}
			if len(block.Data.Message.StateRoot) == 0 {
				// todo better network handling, if 404 just log info, else log error
				d.log.Error(err, "can not get block data", 0, map[string]interface{}{"slot": slot})
				return errors.New("can not get block data")
			}

			mutex.Lock()
			result.beaconBlockData[slot] = block
			mutex.Unlock()

			blockReward, err := cl.GetPropoalRewards(slot)
			if err != nil {
				d.log.Error(err, "can not get block reward data", 0, map[string]interface{}{"slot": slot})
				return err
			}
			mutex.Lock()
			result.beaconBlockRewardData[slot] = blockReward
			mutex.Unlock()

			syncRewards, err := cl.GetSyncRewards(slot)
			if err != nil {
				d.log.Error(err, "can not get sync committee reward data", 0, map[string]interface{}{"slot": slot})
				return err
			}
			mutex.Lock()
			result.syncCommitteeRewardData[slot] = syncRewards
			mutex.Unlock()

			return nil
		})
	}

	err = errGroup.Wait()
	if err != nil {
		return nil, err
	}

	d.log.Infof("retrieved all data for epoch %d in %v", epoch, time.Since(totalStart))

	return &result, nil
}

func (d *dashboardData) process(data *Data, domain []byte) ([]*validatorDashboardDataRow, error) {
	validatorsData := make([]*validatorDashboardDataRow, len(data.endBalances.Data))

	idealAttestationRewards := make(map[int64]int)
	for i, idealReward := range data.attestationRewards.Data.IdealRewards {
		idealAttestationRewards[idealReward.EffectiveBalance] = i
	}

	pubkeyToIndexMapEnd := make(map[string]int64, len(validatorsData))
	pubkeyToIndexMapStart := make(map[string]int64, len(validatorsData))
	activeCount := 0
	// write start & end balances and slashed status
	for i := 0; i < len(validatorsData); i++ {
		validatorsData[i] = &validatorDashboardDataRow{}
		if i < len(data.startBalances.Data) {
			validatorsData[i].BalanceStart = data.startBalances.Data[i].Balance
			pubkeyToIndexMapStart[string(data.startBalances.Data[i].Validator.Pubkey)] = int64(i)

			if data.startBalances.Data[i].Status.IsActive() {
				activeCount++
				validatorsData[i].AttestationsScheduled = sql.NullInt16{Int16: 1, Valid: true}
			}

			if data.genesis {
				validatorsData[i].DepositsCount = sql.NullInt16{Int16: 1, Valid: true}
				validatorsData[i].DepositsAmount = sql.NullInt64{Int64: int64(data.startBalances.Data[i].Validator.EffectiveBalance), Valid: true}
			}
		} else {
			pubkeyToIndexMapEnd[string(data.endBalances.Data[i].Validator.Pubkey)] = int64(i)
		}

		validatorsData[i].BalanceEnd = data.endBalances.Data[i].Balance
		validatorsData[i].Slashed = data.endBalances.Data[i].Validator.Slashed
	}

	// slotsPerSyncCommittee :=  * float64(utils.Config.Chain.ClConfig.SlotsPerEpoch)
	for validator_index := range validatorsData {
		validatorsData[validator_index].BlockChance = float64(utils.Config.Chain.ClConfig.SlotsPerEpoch) / float64(activeCount)
	}

	size := uint64(len(validatorsData))
	sizeInt := int64(len(validatorsData))
	size32 := uint32(len(validatorsData))

	// write scheduled block data
	for _, proposerAssignment := range data.proposerAssignments.Data {
		proposerIndex := proposerAssignment.ValidatorIndex
		if proposerIndex >= size {
			return nil, errors.New("proposer index out of range")
		}
		validatorsData[proposerIndex].BlockScheduled.Int16++
		validatorsData[proposerIndex].BlockScheduled.Valid = true
	}

	// write scheduled sync committee data
	for _, validator := range data.syncCommitteeAssignments.Data.Validators {
		validatorIndex := int64(validator)
		if validatorIndex >= sizeInt {
			return nil, errors.New("proposer index out of range")
		}
		validatorsData[validatorIndex].SyncScheduled.Int16 = int16(len(data.beaconBlockData)) // take into account missed slots
		validatorsData[validatorIndex].SyncScheduled.Valid = true
	}

	// write proposer rewards data
	for _, reward := range data.beaconBlockRewardData {
		if reward.Data.ProposerIndex >= size {
			return nil, errors.New("proposer index out of range")
		}
		validatorsData[reward.Data.ProposerIndex].BlocksClReward.Int64 += reward.Data.Attestations + reward.Data.AttesterSlashings + reward.Data.ProposerSlashings + reward.Data.SyncAggregate
		validatorsData[reward.Data.ProposerIndex].BlocksClReward.Valid = true

		validatorsData[reward.Data.ProposerIndex].SlasherRewards.Int64 += reward.Data.AttesterSlashings + reward.Data.ProposerSlashings
		if reward.Data.AttesterSlashings+reward.Data.ProposerSlashings > 0 {
			validatorsData[reward.Data.ProposerIndex].SlasherRewards.Valid = true
		}
	}

	// write sync committee reward data & sync committee execution stats
	for _, rewards := range data.syncCommitteeRewardData {
		for _, reward := range rewards.Data {
			validator_index := reward.ValidatorIndex
			if validator_index >= size {
				return nil, errors.New("proposer index out of range")
			}
			syncReward := reward.Reward
			validatorsData[validator_index].SyncReward.Int64 += syncReward
			validatorsData[validator_index].SyncReward.Valid = true

			if syncReward > 0 {
				validatorsData[validator_index].SyncExecuted.Int16++
				validatorsData[validator_index].SyncExecuted.Valid = true
			}
		}
	}

	// write block specific data
	for _, block := range data.beaconBlockData {
		validatorsData[block.Data.Message.ProposerIndex].BlocksProposed.Int16++
		validatorsData[block.Data.Message.ProposerIndex].BlocksProposed.Valid = true
		validatorsData[block.Data.Message.ProposerIndex].LastSubmittedDutyEpoch = sql.NullInt32{Int32: int32(block.Data.Message.Slot / utils.Config.Chain.ClConfig.SlotsPerEpoch), Valid: true}

		for depositIndex, depositData := range block.Data.Message.Body.Deposits {
			// TODO: properly verify that deposit is valid:
			// if signature is valid I count the deposit towards the balance
			// if signature is invalid and the validator was in the state at the beginning of the epoch I count the deposit towards the balance
			// if signature is invalid and the validator was NOT in the state at the beginning of the epoch and there were no valid deposits in the block prior I DO NOT count the deposit towards the balance
			// if signature is invalid and the validator was NOT in the state at the beginning of the epoch and there was a VALID deposit in the blocks prior I DO COUNT the deposit towards the balance

			err := utils.VerifyDepositSignature(&phase0.DepositData{
				PublicKey:             phase0.BLSPubKey(depositData.Data.Pubkey),
				WithdrawalCredentials: depositData.Data.WithdrawalCredentials,
				Amount:                phase0.Gwei(depositData.Data.Amount),
				Signature:             phase0.BLSSignature(depositData.Data.Signature),
			}, domain)

			if err != nil {
				d.log.Error(fmt.Errorf("deposit at index %d in slot %v is invalid: %v (signature: %s)", depositIndex, block.Data.Message.Slot, err, depositData.Data.Signature), "", 0)

				// if the validator hat a valid deposit prior to the current one, count the invalid towards the balance
				if validatorsData[pubkeyToIndexMapEnd[string(depositData.Data.Pubkey)]].DepositsCount.Int16 > 0 {
					d.log.Infof("validator had a valid deposit in some earlier block of the epoch, count the invalid towards the balance")
				} else if _, ok := pubkeyToIndexMapStart[string(depositData.Data.Pubkey)]; ok {
					d.log.Infof("validator had a valid deposit in some block prior to the current epoch, count the invalid towards the balance")
				} else {
					d.log.Infof("validator did not have a prior valid deposit, do not count the invalid towards the balance")
					continue
				}
			}

			validator_index := pubkeyToIndexMapEnd[string(depositData.Data.Pubkey)]
			if validator_index >= sizeInt {
				return nil, errors.New("proposer index out of range")
			}

			validatorsData[validator_index].DepositsAmount.Int64 += int64(depositData.Data.Amount)
			validatorsData[validator_index].DepositsAmount.Valid = true

			validatorsData[validator_index].DepositsCount.Int16++
			validatorsData[validator_index].DepositsCount.Valid = true
		}

		for _, withdrawal := range block.Data.Message.Body.ExecutionPayload.Withdrawals {
			validator_index := withdrawal.ValidatorIndex

			if validator_index >= size {
				return nil, errors.New("proposer index out of range")
			}
			validatorsData[validator_index].WithdrawalsAmount.Int64 += int64(withdrawal.Amount)
			validatorsData[validator_index].WithdrawalsAmount.Valid = true

			validatorsData[validator_index].WithdrawalsCount.Int16++
			validatorsData[validator_index].WithdrawalsCount.Valid = true
		}

		for _, attestation := range block.Data.Message.Body.Attestations {
			aggregationBits := bitfield.Bitlist(attestation.AggregationBits)

			for i := uint64(0); i < aggregationBits.Len(); i++ {
				if aggregationBits.BitAt(i) {
					validator_index, found := data.attestationAssignments[utils.FormatAttestorAssignmentKeyLowMem(attestation.Data.Slot, attestation.Data.Index, uint32(i))]
					if !found { // This should never happen!
						d.log.Error(fmt.Errorf("validator not found in attestation assignments"), "validator not found in attestation assignments", 0, map[string]interface{}{"slot": attestation.Data.Slot, "index": attestation.Data.Index, "i": i})
						return nil, fmt.Errorf("validator not found in attestation assignments")
					}
					if validator_index >= size32 {
						return nil, errors.New("proposer index out of range")
					}

					validatorsData[validator_index].InclusionDelaySum = sql.NullInt32{
						Int32: int32(block.Data.Message.Slot - attestation.Data.Slot - 1),
						Valid: true,
					}

					validatorsData[validator_index].LastSubmittedDutyEpoch = sql.NullInt32{Int32: int32(attestation.Data.Slot / utils.Config.Chain.ClConfig.SlotsPerEpoch), Valid: true}

					optimalInclusionDistance := 0
					for i := attestation.Data.Slot + 1; i < block.Data.Message.Slot; i++ {
						if _, ok := data.missedslots[i]; ok {
							optimalInclusionDistance++
						} else {
							break
						}
					}

					validatorsData[validator_index].OptimalInclusionDelay = sql.NullInt32{Int32: int32(optimalInclusionDistance), Valid: true}
				}
			}
		}

		// attester slashings "done" by proposer (slashed by)
		for _, data := range block.Data.Message.Body.AttesterSlashings {
			slashedIndices := data.GetSlashedIndices()
			for _, index := range slashedIndices {
				validatorsData[index].SlashedBy = sql.NullInt32{Int32: int32(block.Data.Message.ProposerIndex), Valid: true}
				validatorsData[index].Slashed = true
				validatorsData[index].SlashedViolation = sql.NullInt16{Int16: SLASHED_VIOLATION_ATTESTATION, Valid: true}
			}
		}

		// proposer slashings "done" by proposer (slashed by)
		for _, data := range block.Data.Message.Body.ProposerSlashings {
			validatorsData[data.SignedHeader1.Message.ProposerIndex].SlashedBy = sql.NullInt32{Int32: int32(block.Data.Message.ProposerIndex), Valid: true}
			validatorsData[data.SignedHeader1.Message.ProposerIndex].Slashed = true
			validatorsData[data.SignedHeader1.Message.ProposerIndex].SlashedViolation = sql.NullInt16{Int16: SLASHED_VIOLATION_PROPOSER, Valid: true}
		}
	}

	// write attestation rewards data
	for _, attestationReward := range data.attestationRewards.Data.TotalRewards {
		validator_index := attestationReward.ValidatorIndex
		if validator_index >= size {
			return nil, errors.New("proposer index out of range")
		}

		validatorsData[validator_index].AttestationsHeadReward = sql.NullInt32{Int32: attestationReward.Head, Valid: true}
		validatorsData[validator_index].AttestationsSourceReward = sql.NullInt32{Int32: attestationReward.Source, Valid: true}
		validatorsData[validator_index].AttestationsTargetReward = sql.NullInt32{Int32: attestationReward.Target, Valid: true}
		validatorsData[validator_index].AttestationsInactivityPenalty = sql.NullInt32{Int32: attestationReward.Inactivity, Valid: true}
		validatorsData[validator_index].AttestationsInclusionsReward = sql.NullInt32{Int32: attestationReward.InclusionDelay, Valid: true}
		validatorsData[validator_index].AttestationReward = sql.NullInt64{
			Int64: int64(attestationReward.Head + attestationReward.Source + attestationReward.Target + attestationReward.Inactivity + attestationReward.InclusionDelay),
			Valid: true,
		}
		idealRewardsOfValidator := data.attestationRewards.Data.IdealRewards[idealAttestationRewards[int64(data.startBalances.Data[validator_index].Validator.EffectiveBalance)]]
		validatorsData[validator_index].AttestationsIdealHeadReward = sql.NullInt32{Int32: idealRewardsOfValidator.Head, Valid: true}
		validatorsData[validator_index].AttestationsIdealTargetReward = sql.NullInt32{Int32: idealRewardsOfValidator.Target, Valid: true}
		validatorsData[validator_index].AttestationsIdealSourceReward = sql.NullInt32{Int32: idealRewardsOfValidator.Source, Valid: true}
		validatorsData[validator_index].AttestationsIdealInactivityPenalty = sql.NullInt32{Int32: idealRewardsOfValidator.Inactivity, Valid: true}
		validatorsData[validator_index].AttestationsIdealInclusionsReward = sql.NullInt32{Int32: idealRewardsOfValidator.InclusionDelay, Valid: true}

		validatorsData[validator_index].AttestationIdealReward = sql.NullInt64{
			Int64: int64(idealRewardsOfValidator.Head + idealRewardsOfValidator.Source + idealRewardsOfValidator.Target + idealRewardsOfValidator.Inactivity + idealRewardsOfValidator.InclusionDelay),
			Valid: true,
		}

		if attestationReward.Head > 0 {
			validatorsData[validator_index].AttestationHeadExecuted = sql.NullInt16{Int16: 1, Valid: true}
			validatorsData[validator_index].AttestationsExecuted = sql.NullInt16{Int16: 1, Valid: true}
		}
		if attestationReward.Source > 0 {
			validatorsData[validator_index].AttestationSourceExecuted = sql.NullInt16{Int16: 1, Valid: true}
			validatorsData[validator_index].AttestationsExecuted = sql.NullInt16{Int16: 1, Valid: true}
		}
		if attestationReward.Target > 0 {
			validatorsData[validator_index].AttestationTargetExecuted = sql.NullInt16{Int16: 1, Valid: true}
			validatorsData[validator_index].AttestationsExecuted = sql.NullInt16{Int16: 1, Valid: true}
		}
	}

	return validatorsData, nil
}

func parseEpochRange(pattern, partition string) (uint64, uint64, error) {
	// Compile the regular expression pattern
	regex := regexp.MustCompile(pattern)

	// Find the matches in the partition string
	matches := regex.FindStringSubmatch(partition)

	// Check if the partition string matches the pattern
	if len(matches) != 3 {
		return 0, 0, fmt.Errorf("invalid partition string: %s", partition)
	}

	// Parse the epoch range values
	epochFrom, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse epoch from: %w", err)
	}

	epochTo, err := strconv.ParseUint(matches[2], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse epoch to: %w", err)
	}

	return epochFrom, epochTo, nil
}

type DataEpoch struct {
	DataRaw *Data
	Data    []*validatorDashboardDataRow
	Epoch   uint64
}

type EpochParallelGroup struct {
	Epochs     []uint64
	Sequential bool
}

type validatorDashboardDataRow struct {
	AttestationsSourceReward           sql.NullInt32 //done
	AttestationsTargetReward           sql.NullInt32 //done
	AttestationsHeadReward             sql.NullInt32 //done
	AttestationsInactivityPenalty      sql.NullInt32 //done
	AttestationsInclusionsReward       sql.NullInt32 //done
	AttestationReward                  sql.NullInt64 //done
	AttestationsIdealSourceReward      sql.NullInt32 //done
	AttestationsIdealTargetReward      sql.NullInt32 //done
	AttestationsIdealHeadReward        sql.NullInt32 //done
	AttestationsIdealInactivityPenalty sql.NullInt32 //done
	AttestationsIdealInclusionsReward  sql.NullInt32 //done
	AttestationIdealReward             sql.NullInt64 //done

	AttestationsScheduled     sql.NullInt16 //done
	AttestationsExecuted      sql.NullInt16 //done
	AttestationHeadExecuted   sql.NullInt16 //done
	AttestationSourceExecuted sql.NullInt16 //done
	AttestationTargetExecuted sql.NullInt16 //done

	LastSubmittedDutyEpoch sql.NullInt32 // does not include sync committee duty slots

	BlockScheduled sql.NullInt16 // done
	BlocksProposed sql.NullInt16 // done
	BlockChance    float64       // done

	BlocksClReward sql.NullInt64 // done

	SyncScheduled sql.NullInt16 // done
	SyncExecuted  sql.NullInt16 // done
	SyncReward    sql.NullInt64 // done

	SlasherRewards   sql.NullInt64 // done
	Slashed          bool          // done
	SlashedBy        sql.NullInt32 // done
	SlashedViolation sql.NullInt16 // done

	BalanceStart uint64 // done
	BalanceEnd   uint64 // done

	DepositsCount  sql.NullInt16 // done
	DepositsAmount sql.NullInt64 // done

	WithdrawalsCount  sql.NullInt16 // done
	WithdrawalsAmount sql.NullInt64 // done

	InclusionDelaySum     sql.NullInt32 // done
	OptimalInclusionDelay sql.NullInt32 // done
}

const SLASHED_VIOLATION_ATTESTATION = 1
const SLASHED_VIOLATION_PROPOSER = 2
