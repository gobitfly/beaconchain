package modules

import (
	"database/sql"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/consapi/network"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	edb "github.com/gobitfly/beaconchain/pkg/exporter/db"
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/go-bitfield"
	"golang.org/x/sync/errgroup"
)

// -------------- DEBUG FLAGS ----------------
// Normally rolling aggregation is only done when headEpochQueue exporter is near head so the exporter can catch up faster if behind, but for debugging purposes we can force it to be done every epoch
const debugAggregateMidEveryEpoch = false // prod: false

// If set to 0 exporter will backfill to node finalized head, use any other value to backfill up to that specific epoch
const debugTargetBackfillEpoch = uint64(0) // prod: 0

// Once backfill is done the exporter will start to listen for new head epochs to export. You can disable this behavior by setting this flag to false
const debugSetBackfillCompleted = true // prod: true

// Old epoch data is cleared from the database to save space and improve performance. This can be disabled for debugging purposes
const debugSkipOldEpochClear = false // prod: false

// If set to true some tables like the epoch based table or hourly based table will be manually added to AlloyDBs Column Engine
const debugAddToColumnEngine = false // prod: true?

// During backfill we can attempt to bootstrap the rolling tables on each UTC boundary day (as no tail fetching is needed here). So setting this will update the rolling tables every
// 225 epochs (for ETH mainnet) but at the slight cost of increased aggregation time for this particular boundary epoch.
const debugAggregateRollingWindowsDuringBackfillUTCBoundEpoch = true // prod: true

const debugDeadlockBandaid = true // prod: fix root cause then set to false

// This flag can be used to force a bootstrap of the rolling tables. This is done once, after the bootstrap completes it switches back to off and normal rolling aggregation.
// Can be used to fix a corrupted rolling table.
var debugForceBootstrapRollingTables = false // prod: false

// ----------- END OF DEBUG FLAGS ------------

// How many epochs will be fetched in parallel from the node (relevant for backfill and rolling tail fetching). We are fetching the head epoch and
// one epoch for each rolling table (tail), so if you want to fetch all epochs in one go (and your node can handle that) set this to at least 5.
const epochFetchParallelism = 5

// Fetching one epoch consists of multiple calls. You can define how many concurrent calls each epoch fetch will do. Keep in mind that
// the total number of concurrent requests is epochFetchParallelism * epochFetchParallelismWithinEpoch
const epochFetchParallelismWithinEpoch = 6

// How many epochs will be written in parallel to the database
const epochWriteParallelism = 4

// How many epoch aggregations will be executed in parallel (e.g. total, hour, day, each rolling table)
const databaseAggregationParallelism = 4

// How many epochFetchParallelism iterations will be written before a new aggregation will be triggered during backfill. This can speed up backfill as writing epochs to db is fast and we can delay
// aggregation for a couple iterations. Don't set too high or else epoch table will grow to large and will be a bottleneck.
// Set to 0 to disable and write after every iteration. Recommended value for this is 1 or maybe 2.
// Try increasing this one by one if node_fetch_time < agg_and_storage_time until it targets roughly agg_and_storage_time = node_fetch_time
// Try 0 if agg_and_storage_time is < node_fetch_time
const backfillMaxUnaggregatedIterations = 1

type dashboardData struct {
	ModuleContext
	log               ModuleLog
	signingDomain     []byte
	epochWriter       *epochWriter
	epochToTotal      *epochToTotalAggregator
	epochToHour       *epochToHourAggregator
	epochToDay        *epochToDayAggregator
	dayUp             *dayUpAggregator
	headEpochQueue    chan uint64
	backFillCompleted bool
	responseCache     ResponseCache
}

func NewDashboardDataModule(moduleContext ModuleContext) ModuleInterface {
	temp := &dashboardData{
		ModuleContext: moduleContext,
	}
	temp.log = ModuleLog{module: temp}

	// When a new epoch gets exported the very first step is to export it to the db via epochWriter
	temp.epochWriter = newEpochWriter(temp)

	// Then those aggregators below use the epoch data to aggregate it into the respective tables
	temp.epochToTotal = newEpochToTotalAggregator(temp)
	temp.epochToHour = newEpochToHourAggregator(temp)
	temp.epochToDay = newEpochToDayAggregator(temp)

	// Once an epoch is aggregated to its respective UTC day, we can use the UTC day table to aggregate up to the rolling window tables (7d, 30d, 90d)
	temp.dayUp = newDayUpAggregator(temp)

	// This channel is used to queue up epochs from chain head that need to be exported
	temp.headEpochQueue = make(chan uint64, 100)

	// Indicates whether the initial backfill - which is checked when starting the exporter - has completed
	// and the exporter can start listening for new head epochs to be processed
	temp.backFillCompleted = false

	temp.responseCache = ResponseCache{
		cache: make(map[string]any),
	}
	return temp
}

func (d *dashboardData) Init() error {
	go func() {
		_, err := db.AlloyWriter.Exec("SET work_mem TO '128MB';")
		if err != nil {
			d.log.Fatal(err, "failed to set work_mem", 0)
		}

		start := time.Now()
		for {
			var upToEpochPtr *uint64 = nil // nil will backfill back to head
			if debugTargetBackfillEpoch > 0 {
				upToEpoch := debugTargetBackfillEpoch
				upToEpochPtr = &upToEpoch
			}

			result, err := d.backfillHeadEpochData(upToEpochPtr)
			if err != nil {
				d.log.Error(err, "failed to backfill epoch data", 0)
				metrics.Errors.WithLabelValues("exporter_v2dash_backfill_fail").Inc()
				time.Sleep(10 * time.Second)
				continue
			}

			if result.BackfilledToHead {
				d.log.Infof("dashboard data up to date, starting head export")
				if debugSetBackfillCompleted {
					if time.Since(start) > time.Hour {
						utils.SendMessage(fmt.Sprintf("🎉🎉🎉 v2 Dashboard %s - Reached head, exporting from head now", utils.Config.Chain.Name), &utils.Config.InternalAlerts)
					}
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
	reachedHead := false
	for {
		epoch := <-d.headEpochQueue

		// After initial sync or long downtime first head processing might take a long time, so by the time we finished
		// the queue might have filled up significantly. To get back on head more quickly we skip some epochs and let the backfill handle those
		// before processing the more recent epoch
		for len(d.headEpochQueue) > 1 {
			epoch = <-d.headEpochQueue
		}
		if len(d.headEpochQueue) == 0 && !reachedHead {
			d.log.Infof("exporter is at head of the chain")
			reachedHead = true
		}

		startTime := time.Now()
		d.log.Infof("exporting dashboard epoch data for epoch %d", epoch)
		stage := 0
		doRollingAggregate := false
		for { // retry this epoch until no errors occur
			currentFinalizedEpoch, err := d.CL.GetFinalityCheckpoints("head")
			if err != nil {
				d.log.Error(err, "failed to get finalized checkpoint", 0)
				metrics.Errors.WithLabelValues("exporter_v2dash_node_get_finalize_fail").Inc()
				time.Sleep(time.Second * 10)
				continue
			}

			// Back fill to epoch -1 if necessary
			var backfillResult backfillResult
			if stage <= 0 {
				targetEpoch := epoch - 1
				backfillResult, err = d.backfillHeadEpochData(&targetEpoch)
				if err != nil {
					d.log.Error(err, "failed to backfill head epoch data", 0, map[string]interface{}{"epoch": epoch})
					metrics.Errors.WithLabelValues("exporter_v2dash_backfill_fail").Inc()
					time.Sleep(time.Second * 10)
					continue
				}
				stage = 1
			}

			// Get epoch data from node and write to database
			if stage <= 1 {
				doRollingAggregate = currentFinalizedEpoch.Data.Finalized.Epoch <= epoch+1 // only near head
				err := d.exportEpochAndTails(epoch, debugAggregateMidEveryEpoch || doRollingAggregate)
				if err != nil {
					d.log.Error(err, "failed to export epoch tail data", 0, map[string]interface{}{"epoch": epoch})
					metrics.Errors.WithLabelValues("exporter_v2dash_export_epoch_tail_fail").Inc()
					time.Sleep(time.Second * 10)
					continue
				}
				stage = 2
			}

			// Run aggregations
			if stage <= 2 {
				err := d.aggregatePerEpoch(debugAggregateMidEveryEpoch || doRollingAggregate, backfillResult.DidPerformBackfill) // keep epoch data if backfill was needed
				if err != nil {
					d.log.Error(err, "failed to aggregate", 0, map[string]interface{}{"epoch": epoch})
					metrics.Errors.WithLabelValues("exporter_v2dash_agg_fail").Inc()
					time.Sleep(time.Second * 10)
					continue
				}
				stage = 3
			}

			break
		}

		d.log.Infof("[time] completed dashboard epoch data for epoch %d in %v", epoch, time.Since(startTime))
	}
}

// exports the provided headEpoch plus any tail epochs that are needed for rolling aggregation
// fE a tail epoch for rolling 1 day aggregation (225 epochs) for head 227 on ethereum would correspond to two tail epochs [0,1]
func (d *dashboardData) exportEpochAndTails(headEpoch uint64, fetchRollingTails bool) error {
	missingTails := make([]uint64, 0)
	var err error
	if fetchRollingTails {
		// for 24h aggregation
		missingTails, err = d.epochToDay.getMissingRolling24TailEpochs(headEpoch)
		if err != nil {
			return errors.Wrap(err, "failed to get missing 24h tail epochs")
		}

		d.log.Infof("missing 24h tails: %v", missingTails)

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
		missingTails = append(missingTails, utils.Deduplicate(append(daysMissingTails, dayMissingHeads...))...)

		if len(missingTails) > 10 {
			d.log.Infof("This might take a bit longer than usual as exporter is catching up quite a lot old epochs, usually happens after downtime or after initial sync")
		}

		// sort asc
		sort.Slice(missingTails, func(i, j int) bool {
			return missingTails[i] < missingTails[j]
		})
	}

	hasHeadAlreadyExported, err := edb.HasDashboardDataForEpoch(headEpoch)
	if err != nil {
		return errors.Wrap(err, "failed to check if head epoch has dashboard data")
	}

	// append head
	if !hasHeadAlreadyExported {
		missingTails = append(missingTails, headEpoch)
		d.log.Infof("fetch missing tail/head epochs: %v | fetch head: %d", len(missingTails)-1, headEpoch)
	} else {
		if len(missingTails) == 0 {
			return nil // nothing to do
		}
		d.log.Infof("fetch missing tail/head epochs: %v | fetch head: -", len(missingTails))
	}

	var nextDataChan chan []DataEpochProcessed = make(chan []DataEpochProcessed, 1)
	go func() {
		d.epochDataFetcher(missingTails, epochFetchParallelism, nextDataChan)
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
func (d *dashboardData) epochDataFetcher(epochs []uint64, epochFetchParallelism int, nextDataChan chan []DataEpochProcessed) {
	// group epochs into parallel worker groups
	groups := getEpochParallelGroups(epochs, epochFetchParallelism)
	numberOfEpochsToFetch := len(epochs)
	epochsFetched := 0

	for _, gapGroup := range groups {
		errGroup := &errgroup.Group{}

		datas := make([]*Data, 0, epochFetchParallelism)
		start := time.Now()

		// identifies unique sync periods in this epoch group
		var syncCommitteePeriods = make(map[uint64]bool)

		// Step 1: fetch epoch data raw
		for _, gap := range gapGroup.Epochs {
			gap := gap
			syncCommitteePeriods[utils.SyncPeriodOfEpoch(gap)] = true

			errGroup.Go(func() error {
				for {
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
						metrics.Errors.WithLabelValues("exporter_v2dash_node_fail").Inc()
						time.Sleep(time.Second * 10)
						continue
					}

					datas = append(datas, data)

					break
				}
				return nil
			})
		}

		// Step 2: fetch sync committee data
		d.getSyncCommitteesData(errGroup, syncCommitteePeriods)

		_ = errGroup.Wait() // no need to catch error since it will retry unless all clear without errors

		// Clear old sync committee cache entries that are not relevant for this group
		d.clearOldCache(syncCommitteePeriods)

		// sort datas first, epoch asc
		sort.Slice(datas, func(i, j int) bool {
			return datas[i].epoch < datas[j].epoch
		})

		// Half Step: for sequential since we used skipSerialCalls we must provide startBalance and missedSlots for every epoch except FromEpoch
		// with the data from the previous epoch. This is done to improve performance by skipping some calls
		if gapGroup.Sequential {
			// provide data from the previous epochs
			for i := 1; i < len(datas); i++ {
				datas[i].lastEpochStateEnd = datas[i-1].currentEpochStateEnd

				for slot := range datas[i-1].missedslots {
					datas[i].missedslots[slot] = true
				}
			}
		}

		processed := make([]DataEpochProcessed, len(datas))

		// Step 2: process data
		errGroup = &errgroup.Group{}
		errGroup.SetLimit(int(math.Max(epochWriteParallelism/2, 2.0))) // mitigate short ram spike
		for i := 0; i < len(datas); i++ {
			i := i
			errGroup.Go(func() error {
				for {
					d.log.Infof("epoch data fetcher, processing data for epoch %d", datas[i].epoch)
					start := time.Now()

					result, err := d.ProcessEpochData(datas[i])
					if err != nil {
						d.log.Error(err, "failed to process epoch data", 0, map[string]interface{}{"epoch": datas[i].epoch})
						time.Sleep(time.Second * 10)
						continue
					}

					err = storeClBlockRewards(datas[i].beaconBlockRewardData)
					if err != nil {
						d.log.Error(err, "failed to store cl block rewards", 0, map[string]interface{}{"epoch": datas[i].epoch})
						time.Sleep(time.Second * 10)
						continue
					}

					processed[i] = DataEpochProcessed{
						Epoch: datas[i].epoch,
						Data:  result,
					}

					d.log.Infof("epoch data fetcher, processed data for epoch %d in %v", datas[i].epoch, time.Since(start))
					break
				}
				return nil
			})
		}

		_ = errGroup.Wait() // no need to catch error since it will retry unless all clear without errors
		datas = nil         // clear raw data, not needed any more

		{
			epochsFetched += len(processed)
			remaining := numberOfEpochsToFetch - epochsFetched
			remainingTimeEst := time.Duration(0)
			if remaining > 0 {
				remainingTimeEst = time.Duration(time.Since(start).Nanoseconds() / int64(len(processed)) * int64(remaining))
			}

			d.log.Infof("[time] epoch data fetcher, fetched %v epochs %v in %v. Remaining: %v (%v)", len(processed), gapGroup.Epochs, time.Since(start), remaining, remainingTimeEst)
			metrics.TaskDuration.WithLabelValues("exporter_v2dash_fetch_epochs").Observe(time.Since(start).Seconds())
			metrics.TaskDuration.WithLabelValues("exporter_v2dash_fetch_epochs_per_epochs").Observe(time.Since(start).Seconds() / float64(len(processed)))
		}

		nextDataChan <- processed
	}
}

// Fetches sync committee assignments of provided periods
func (d *dashboardData) getSyncCommitteesData(errGroup *errgroup.Group, syncCommitteePeriods map[uint64]bool) {
	for syncPeriod := range syncCommitteePeriods {
		syncPeriod := syncPeriod
		// -- Get current sync committee members and cache it
		{
			if found := d.responseCache.GetSyncCommittee(syncPeriod); found == nil {
				errGroup.Go(func() error {
					for {
						start := time.Now()
						data, err := d.CL.GetSyncCommitteesAssignments(nil, utils.FirstEpochOfSyncPeriod(syncPeriod)*utils.Config.Chain.ClConfig.SlotsPerEpoch)
						if err != nil {
							d.log.Error(err, "cannot get sync committee assignments", 0, map[string]interface{}{"syncPeriod": syncPeriod})
							metrics.Errors.WithLabelValues("exporter_v2dash_node_committee_fail").Inc()
							time.Sleep(time.Second * 10)
							continue
						}
						d.responseCache.SetSyncCommittee(syncPeriod, data)
						d.log.Infof("retrieved sync committee members for sync period %d in %v", syncPeriod, time.Since(start))
						break
					}
					return nil
				})
			}
		}
	}
}

func (d *dashboardData) clearOldCache(syncCommitteePeriods map[uint64]bool) {
	// delete old sync committee election cache entries
	for key := range d.responseCache.cache {
		stillNeeded := false

		if strings.Contains(key, RawSyncCommitteeCacheKey) {
			for syncPeriod := range syncCommitteePeriods {
				syncCommitteeCacheKey := d.responseCache.GetSyncCommitteeCacheKey(syncPeriod)
				if key == syncCommitteeCacheKey {
					stillNeeded = true
					break
				}
			}
		}

		if !stillNeeded {
			delete(d.responseCache.cache, key)
			d.log.Infof("deleted stale sync committee cache entry %s", key)
		}
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

var unaggregatedWrites = 0

type backfillResult struct {
	BackfilledToHead   bool // if backfill finished and current chain state is head (only set of backfill to head was requested, so only when upToEpoch = null)
	DidPerformBackfill bool // whether a backfill was performed at all, false if backfill was not necessary
}

// can be used to start a backfill up to epoch
// if upToEpoch is nil, it will backfill until the latest finalized epoch
func (d *dashboardData) backfillHeadEpochData(upToEpoch *uint64) (backfillResult, error) {
	var result = backfillResult{}
	backfillToChainFinalizedHead := upToEpoch == nil
	if upToEpoch == nil {
		res, err := d.CL.GetFinalityCheckpoints("head")
		if err != nil {
			return result, errors.Wrap(err, "failed to get finalized checkpoint")
		}
		if utils.IsByteArrayAllZero(res.Data.Finalized.Root) {
			return result, errors.New("network not finalized yet")
		}
		upToEpoch = &res.Data.Finalized.Epoch
		d.log.Infof("backfilling head epoch data up to epoch %d", *upToEpoch)
	}

	latestExportedEpoch, err := edb.GetLatestDashboardEpoch()
	if err != nil {
		return result, errors.Wrap(err, "failed to get latest dashboard epoch")
	}

	// An unclean shutdown can occur when exporter is shut down in between writing epoch data
	// while each epoch is its own transaction and hence writes the complete epoch or nothing,
	// it could happen that epochs have been written out of order due to the parallel nature of the exporter.
	// Meaning that there is a gap in the last ~epochFetchParallelism epochs
	{
		uncleanShutdownGaps, err := edb.GetMissingEpochsBetween(int64(latestExportedEpoch-epochFetchParallelism), int64(latestExportedEpoch+1))
		if err != nil {
			return result, errors.Wrap(err, "failed to get epoch gaps")
		}

		if latestExportedEpoch > 0 && len(uncleanShutdownGaps) > 0 {
			d.log.Infof("Unclean shutdown detected, backfilling missing epochs %d", uncleanShutdownGaps)
			var nextDataChan chan []DataEpochProcessed = make(chan []DataEpochProcessed, 1)
			go func() {
				d.epochDataFetcher(uncleanShutdownGaps, epochFetchParallelism, nextDataChan)
			}()

			for {
				datas := <-nextDataChan
				done := containsEpoch(datas, uncleanShutdownGaps[len(uncleanShutdownGaps)-1])
				d.writeEpochDatas(datas)
				if done {
					break
				}
			}
			d.log.Infof("Fixed unclean shutdown gaps")
		}
	}

	// more epoch partitions than expected would indicate that a rolling aggregation backfill was interrupted or failed.
	// We can use ancientEpochsPresent to keep the ancient epochs for a bit longer to prevent repeating fetching work.
	var ancientEpochsPresent bool
	{
		partitions, err := edb.GetPartitionNamesOfTable(edb.EpochWriterTableName)
		if err != nil {
			return result, errors.Wrap(err, "failed to get partitions")
		}

		epochsInDb := len(partitions) * PartitionEpochWidth
		epochsExpectedInDb := int(float64(d.epochWriter.getRetentionEpochDuration()) * 1.2) // 20% buffer
		maxAncientEpochs := int(float64(4*utils.EpochsPerDay()) * 0.75)                     // 18h (24h x 4 rolling tables, 75%). Makes sure we keep not too many which would degrade db performance
		ancientEpochsPresent = epochsInDb > epochsExpectedInDb && epochsInDb < maxAncientEpochs
		d.log.Infof("Checked for ancient epochs. Epochs in db: %d, expected: %d, maxAncientEpochs: %d, ancientEpochsPresent: %v", epochsInDb, epochsExpectedInDb, maxAncientEpochs, ancientEpochsPresent)
	}

	gaps, err := edb.GetMissingEpochsBetween(int64(latestExportedEpoch), int64(*upToEpoch+1))
	if err != nil {
		return result, errors.Wrap(err, "failed to get epoch gaps")
	}

	if len(gaps) > 0 {
		// Aggregate (non rolling) before exporting more epochs
		// This is just a precaution so that the aggregated epochs tables are up to date
		// before exporting new epochs
		if latestExportedEpoch > 0 {
			err = d.aggregatePerEpoch(false, true)
			if err != nil {
				return result, errors.Wrap(err, "failed to aggregate")
			}
		}

		d.log.Infof("Epoch dashboard data %d gaps found, backfilling gaps in the range fom epoch %d to %d", len(gaps), gaps[0], gaps[len(gaps)-1])

		// get epochs data
		var nextDataChan chan []DataEpochProcessed = make(chan []DataEpochProcessed, 1)
		go func() {
			d.epochDataFetcher(gaps, epochFetchParallelism, nextDataChan)
		}()

		// save epochs data
		for {
			d.log.Info("storage waiting for data from fetcher")
			datas := <-nextDataChan

			done := containsEpoch(datas, gaps[len(gaps)-1]) // if the last epoch to fetch is in the result set, mark as job completed
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
							err = d.aggregatePerEpoch(true, false) // if we do a bootstrap rolling we don't have to prevent any epoch cleanup
							if err != nil {
								metrics.Errors.WithLabelValues("exporter_v2dash_agg_per_epoch_fail").Inc()
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
				unaggregatedWrites += len(datas)

				if unaggregatedWrites > backfillMaxUnaggregatedIterations*epochFetchParallelism || lastEpoch%225 < epochFetchParallelism {
					unaggregatedWrites = 0
					d.log.Info("storage writing done, aggregate")
					for {
						// This handles the case where we fetch a lot of ancient epochs, something goes wrong, export restarts, backfills (here) to head,
						// but delete all the ancient epochs along the way. And then the head export must fetch them again to do the rollings.
						// So we keep them in the backfill export and let head export clean them up after done work.
						var preventClearOldEpochs bool

						// We target max 24h epochs to be in db to prevent performance degradation.
						// ancientEpochsPresent targets max 18h + adding 6h (1/4 of 24h) = 24h
						isSmallBackfill := len(gaps) < int(utils.EpochsPerDay()/4)

						// prevent cleaning old epochs in case head export got interrupted and we are already done in this backfill iteration
						// or prevent if previous rolling aggregation has been interrupted.
						preventClearOldEpochs = done || ancientEpochsPresent && isSmallBackfill

						err = d.aggregatePerEpoch(false, preventClearOldEpochs)
						if err != nil {
							d.log.Error(err, "backfill, failed to aggregate", 0, map[string]interface{}{"epoch start": datas[0].Epoch, "epoch end": lastEpoch})
							metrics.Errors.WithLabelValues("exporter_v2dash_agg_per_epoch_fail").Inc()
							time.Sleep(time.Second * 10)
							continue
						}
						break
					}
					d.log.InfoWithFields(map[string]interface{}{"epoch start": datas[0].Epoch, "epoch end": lastEpoch}, "backfill, aggregated epoch data")
				}

				metrics.State.WithLabelValues("exporter_v2dash_last_exported_epoch").Set(float64(lastEpoch))
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

	result.DidPerformBackfill = len(gaps) > 0

	// Return with "complete" only if task was to sync to chain finalized head and we finished
	if backfillToChainFinalizedHead {
		res, err := d.CL.GetFinalityCheckpoints("head")
		if err != nil {
			return result, errors.Wrap(err, "failed to get finalized checkpoint")
		}
		if utils.IsByteArrayAllZero(res.Data.Finalized.Root) {
			return result, errors.New("network not finalized yet")
		}

		result.BackfilledToHead = res.Data.Finalized.Epoch-1 <= *upToEpoch
		return result, nil
	}

	return result, nil
}

func containsEpoch(d []DataEpochProcessed, epoch uint64) bool {
	for i := 0; i < len(d); i++ {
		if d[i].Epoch == epoch {
			return true
		}
	}
	return false
}

// stores all passed epoch data, blocks until all data is written without error
func (d *dashboardData) writeEpochDatas(datas []DataEpochProcessed) {
	totalStart := time.Now()
	defer func() {
		d.log.Infof("[time] storage, wrote all epoch data in %v", time.Since(totalStart))
		metrics.TaskDuration.WithLabelValues("exporter_v2dash_write_epochs").Observe(time.Since(totalStart).Seconds())
	}()

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

// Contains all aggregation logic that should happen for every new exported epoch
// forceAggregate triggers an aggregation, use this when calling on head.
// updateRollingWindows specifies whether we should update rolling windows
func (d *dashboardData) aggregatePerEpoch(updateRollingWindows bool, preventClearOldEpochs bool) error {
	currentExportedEpoch, err := edb.GetLatestDashboardEpoch()
	if err != nil {
		return errors.Wrap(err, "failed to get last exported epoch")
	}

	// Performance improvement for backfilling, no need to aggregate day after each epoch, we can update once per hour
	start := time.Now()

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
		err := d.epochToDay.dayAggregate(currentExportedEpoch)
		if err != nil {
			return errors.Wrap(err, "failed to aggregate day")
		}
		return nil
	})

	err = errGroup.Wait()
	if err != nil {
		metrics.Errors.WithLabelValues("exporter_v2dash_agg_non_rolling_fail").Inc()
		return errors.Wrap(err, "failed to aggregate")
	}
	d.log.Infof("[time] all of epoch based aggregation took %v", time.Since(start))
	metrics.TaskDuration.WithLabelValues("exporter_v2dash_agg_non_rolling").Observe(time.Since(start).Seconds())

	if updateRollingWindows {
		// todo you could add it to the err group above IF no bootstrap is needed.

		err = d.aggregateRollingWindows(currentExportedEpoch)
		if err != nil {
			metrics.Errors.WithLabelValues("exporter_v2dash_agg_non_fail").Inc()
			return errors.Wrap(err, "failed to aggregate rolling windows")
		}

		debugForceBootstrapRollingTables = false // reset flag after first run

		err = refreshMaterializedSlashedByCounts()
		if err != nil {
			return errors.Wrap(err, "failed to refresh slashed by counts")
		}
	}

	if !preventClearOldEpochs {
		d.log.Infof("cleaning old epochs")
		err = d.epochWriter.clearOldEpochs(int64(currentExportedEpoch - d.epochWriter.getRetentionEpochDuration()))
		if err != nil {
			return errors.Wrap(err, "failed to clear old epochs")
		}
	}

	metrics.State.WithLabelValues("exporter_v2dash_last_exported_epoch").Set(float64(currentExportedEpoch))

	// clear old hourly aggregated epochs, do not remove epochs from epoch table here as these are needed for Mid aggregation
	err = d.epochToHour.clearOldHourAggregations(int64(currentExportedEpoch - d.epochToHour.getHourRetentionDurationEpochs()))
	if err != nil {
		return errors.Wrap(err, "failed to clear old hours")
	}

	err = d.epochToDay.clearOldDayAggregations(d.epochToDay.getDayRetentionDurationDays())
	if err != nil {
		return errors.Wrap(err, "failed to clear old days")
	}

	return nil
}

// This function contains more heavy aggregation like rolling 7d, 30d, 90d
// This function assumes that epoch aggregation is finished before calling THOUGH it could run in parallel
// as long as the rolling tables do not require a bootstrap
func (d *dashboardData) aggregateRollingWindows(currentExportedEpoch uint64) error {
	start := time.Now()
	defer func() {
		d.log.Infof("[time] all of mid aggregation took %v", time.Since(start))
		metrics.TaskDuration.WithLabelValues("exporter_v2dash_agg_roling").Observe(time.Since(start).Seconds())
	}()

	errGroup := &errgroup.Group{}
	errGroup.SetLimit(databaseAggregationParallelism)

	errGroup.Go(func() error {
		err := d.epochToDay.rolling24hAggregate(currentExportedEpoch)
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
	res, err := d.CL.GetFinalityCheckpoints("head")
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

	metrics.State.WithLabelValues("exporter_v2dash_last_finalized_epoch").Set(float64(res.Data.Finalized.Epoch))

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

func (d *dashboardData) ProcessEpochData(data *Data) ([]*validatorDashboardDataRow, error) {
	if d.signingDomain == nil {
		domain, err := utils.GetSigningDomain()
		if err != nil {
			return nil, err
		}
		d.log.Infof("initialized signing domain to %x", domain)
		d.signingDomain = domain
	}

	return d.process(data, d.signingDomain)
}

func isPartitionAttached(pTable string, partition string) (bool, error) {
	var attached bool
	err := db.AlloyWriter.QueryRow(fmt.Sprintf(`
	SELECT EXISTS (
		SELECT 1
		FROM pg_partitioned_table pgt
		JOIN pg_inherits pi ON pgt.partrelid = pi.inhparent
		JOIN pg_class pc ON pc.oid = pi.inhrelid
		WHERE pgt.partrelid = '%s'::regclass
		AND pc.relname = '%s'
	)
	`,
		pTable, partition,
	)).Scan(&attached)

	if err != nil {
		return false, errors.Wrap(err, "failed to check if partition is already attached")
	}

	return attached, nil
}

// Returns the epoch where the sync committee election for the given epoch took place
func getSyncCommitteeElectionEpochOf(period uint64) uint64 {
	syncElectionPeriod := int64(period) - 1
	firstEpoch := utils.FirstEpochOfSyncPeriod(uint64(syncElectionPeriod))
	if syncElectionPeriod < 0 || firstEpoch < utils.Config.Chain.ClConfig.AltairForkEpoch {
		// first sync committee is identical to second sync committee
		// See https://github.com/ethereum/consensus-specs/blob/dev/specs/altair/fork.md#upgrading-the-state
		return utils.Config.Chain.ClConfig.AltairForkEpoch
	}
	return firstEpoch
}

type Data struct {
	lastEpochStateEnd       *constypes.StandardValidatorsResponse
	currentEpochStateEnd    *constypes.StandardValidatorsResponse
	proposerAssignments     *constypes.StandardProposerAssignmentsResponse
	attestationRewards      []constypes.AttestationReward
	idealAttestationRewards map[int64]constypes.AttestationIdealReward // effective-balance -> ideal reward
	beaconBlockData         map[uint64]*constypes.StandardBeaconSlotResponse
	beaconBlockRewardData   map[uint64]*constypes.StandardBlockRewardsResponse
	syncCommitteeRewardData map[uint64]*constypes.StandardSyncCommitteeRewardsResponse
	attestationAssignments  map[uint64]uint32
	missedslots             map[uint64]bool
	genesis                 bool
	epoch                   uint64

	// Contains the validator state of the epoch where the current sync committee election took place
	syncCommitteeElectedState *constypes.StandardValidatorsResponse
}

const MAX_EFFECTIVE_BALANCE = 32e9

// Data for a single validator
// use skipSerialCalls = false if you are not sure what you are doing. This flag is mainly
// to gain performance improvements when exporting a couple sequential epochs in a row
func (d *dashboardData) getData(epoch, slotsPerEpoch uint64, skipSerialCalls bool) (*Data, error) {
	var result Data
	var err error

	firstSlotOfEpoch := epoch * slotsPerEpoch

	lastSlotOfPreviousEpoch := int64(firstSlotOfEpoch) - 1
	lastSlotOfEpoch := firstSlotOfEpoch + slotsPerEpoch - 1

	result.beaconBlockData = make(map[uint64]*constypes.StandardBeaconSlotResponse, slotsPerEpoch)
	result.beaconBlockRewardData = make(map[uint64]*constypes.StandardBlockRewardsResponse, slotsPerEpoch)
	result.syncCommitteeRewardData = make(map[uint64]*constypes.StandardSyncCommitteeRewardsResponse, slotsPerEpoch)
	result.attestationAssignments = make(map[uint64]uint32)
	result.idealAttestationRewards = make(map[int64]constypes.AttestationIdealReward)
	result.missedslots = make(map[uint64]bool, slotsPerEpoch*2)
	result.epoch = epoch

	cl := d.CL

	errGroup := &errgroup.Group{}
	errGroup.SetLimit(epochFetchParallelismWithinEpoch)

	totalStart := time.Now()

	errGroup.Go(func() error {
		// retrieve proposer assignments for the epoch in order to attribute missed slots
		start := time.Now()
		var err error
		result.proposerAssignments, err = cl.GetPropoalAssignments(epoch)
		if err != nil {
			d.log.Error(err, "can not get proposer assignments", 0, map[string]interface{}{"epoch": epoch})
			return err
		}
		d.log.Debugf("retrieved proposer assignments in %v", time.Since(start))
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
		data, err := cl.GetAttestationRewards(epoch)
		if err != nil {
			d.log.Error(err, "can not get attestation rewards", 0, map[string]interface{}{"epoch": epoch})
			return err
		}

		for _, idealReward := range data.Data.IdealRewards {
			if _, ok := result.idealAttestationRewards[idealReward.EffectiveBalance]; !ok {
				result.idealAttestationRewards[idealReward.EffectiveBalance] = idealReward
			}
		}

		result.attestationRewards = data.Data.TotalRewards

		d.log.Debugf("retrieved attestation rewards data in %v", time.Since(start))
		return nil
	})

	errGroup.Go(func() error {
		// retrieve the validator balances at the end of the epoch
		start := time.Now()
		var err error
		result.currentEpochStateEnd, err = cl.GetValidators(lastSlotOfEpoch, nil, nil)
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
			var err error
			if lastSlotOfPreviousEpoch < 0 {
				result.lastEpochStateEnd, err = d.CL.GetValidators("genesis", nil, nil)
				result.genesis = true
			} else {
				result.lastEpochStateEnd, err = d.CL.GetValidators(lastSlotOfPreviousEpoch, nil, nil)
			}
			if err != nil {
				d.log.Error(err, "can not get validators balances", 0, map[string]interface{}{"lastSlotOfPreviousEpoch": lastSlotOfPreviousEpoch})
				return err
			}
			d.log.Debugf("retrieved start balances using state at slot %d in %v", lastSlotOfPreviousEpoch, time.Since(start))
			return nil
		})

		errGroup.Go(func() error {
			start := time.Now()

			if firstSlotOfEpoch > slotsPerEpoch { // handle case for first epoch
				// get missed slots of last epoch for optimal inclusion distance
				for slot := firstSlotOfEpoch - slotsPerEpoch; slot <= lastSlotOfEpoch-slotsPerEpoch; slot++ {
					_, err := cl.GetBlockHeader(slot)
					if err != nil {
						httpErr := network.SpecificError(err)
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
				httpErr := network.SpecificError(err)
				if httpErr != nil && httpErr.StatusCode == 404 {
					mutex.Lock()
					result.missedslots[slot] = true
					mutex.Unlock()
					return nil // missed
				}

				return err
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

	currentSyncPeriod := utils.SyncPeriodOfEpoch(epoch)
	if epoch == utils.FirstEpochOfSyncPeriod(currentSyncPeriod) {
		syncCommitteElectionEpoch := getSyncCommitteeElectionEpochOf(currentSyncPeriod)
		syncCommitteeElectedInSlot := syncCommitteElectionEpoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
		errGroup.Go(func() error {
			start := time.Now()
			var err error
			result.syncCommitteeElectedState, err = d.CL.GetValidators(syncCommitteeElectedInSlot, nil, []constypes.ValidatorStatus{constypes.Active})
			if err != nil {
				d.log.Error(err, "can not get sync committee election state", 0, map[string]interface{}{"slot": syncCommitteeElectedInSlot})
				return err
			}
			d.log.Infof("retrieved validator state for sync committee period %d (election state is epoch %d) in %v", currentSyncPeriod, syncCommitteElectionEpoch, time.Since(start))

			return nil
		})
	}

	err = errGroup.Wait()
	if err != nil {
		return nil, err
	}

	d.log.Infof("[time] retrieved all data for epoch %d in %v", epoch, time.Since(totalStart))

	return &result, nil
}

func (d *dashboardData) process(data *Data, domain []byte) ([]*validatorDashboardDataRow, error) {
	validatorsData := make([]*validatorDashboardDataRow, len(data.currentEpochStateEnd.Data))

	pubkeyToIndexMapNewlyActivatedValidators := make(map[string]int64, 8)
	pubkeyToIndexMapOldValidators := make(map[string]int64, len(validatorsData))
	activeTotalEffectiveBalanceETH := int64(0)

	currentSyncPeriod := utils.SyncPeriodOfEpoch(data.epoch)

	postAltair := data.epoch >= utils.Config.Chain.ClConfig.AltairForkEpoch

	// write start & end balances and slashed status
	for i := 0; i < len(validatorsData); i++ {
		if uint64(i) != data.currentEpochStateEnd.Data[i].Index { // sanity assumption that i = validator index
			return nil, errors.New("validator index mismatch")
		}

		validatorsData[i] = &validatorDashboardDataRow{}
		if i < len(data.lastEpochStateEnd.Data) {
			validatorsData[i].BalanceStart = data.lastEpochStateEnd.Data[i].Balance
			pubkeyToIndexMapOldValidators[string(data.lastEpochStateEnd.Data[i].Validator.Pubkey)] = int64(i)

			if data.genesis { // Add genesis deposits
				validatorsData[i].DepositsCount = utils.NullInt16(1)
				validatorsData[i].DepositsAmount = utils.NullInt64(int64(data.lastEpochStateEnd.Data[i].Validator.EffectiveBalance))
			}
		} else {
			pubkeyToIndexMapNewlyActivatedValidators[string(data.currentEpochStateEnd.Data[i].Validator.Pubkey)] = int64(i) // validators that become active in the epoch
		}

		if data.currentEpochStateEnd.Data[i].Status.IsActive() {
			activeTotalEffectiveBalanceETH += int64(data.currentEpochStateEnd.Data[i].Validator.EffectiveBalance / 1e9)
			validatorsData[i].AttestationsScheduled = utils.NullInt16(1)
		}

		validatorsData[i].BalanceEnd = data.currentEpochStateEnd.Data[i].Balance
		validatorsData[i].Slashed = data.currentEpochStateEnd.Data[i].Validator.Slashed
	}

	// Expected Block Proposal
	for _, valData := range data.currentEpochStateEnd.Data {
		if valData.Status.IsActive() {
			// See https://github.com/ethereum/annotated-spec/blob/master/phase0/beacon-chain.md#compute_proposer_index
			proposalChance := float64(valData.Validator.EffectiveBalance/1e9) / float64(activeTotalEffectiveBalanceETH)
			validatorsData[valData.Index].BlocksExpectedThisEpoch = proposalChance * float64(utils.Config.Chain.ClConfig.SlotsPerEpoch)
		}
	}

	// Expected Sync Committees
	// Get the total effective balance from the state where the current sync committee was elected
	// And then calculate the chance of being in the sync committee from that state
	if data.epoch == utils.FirstEpochOfSyncPeriod(currentSyncPeriod) {
		if data.syncCommitteeElectedState == nil && postAltair {
			return nil, errors.New("sync committee election state not found")
		}

		syncCommitteeElectionStateTotalEffectiveBalanceETH := int64(0)
		for _, valData := range data.syncCommitteeElectedState.Data {
			if valData.Status.IsActive() {
				syncCommitteeElectionStateTotalEffectiveBalanceETH += int64(valData.Validator.EffectiveBalance / 1e9)
			}
		}

		for _, valData := range data.syncCommitteeElectedState.Data {
			if valData.Status.IsActive() {
				// See https://github.com/ethereum/annotated-spec/blob/master/altair/beacon-chain.md#get_sync_committee_indices
				// Note that this formula is not 100% the chance as defined in the spec, but after running simulations we found
				// it being precise enough for our purposes with an error margin of less than 0.003%
				syncChance := float64(valData.Validator.EffectiveBalance/1e9) / float64(syncCommitteeElectionStateTotalEffectiveBalanceETH)
				validatorsData[valData.Index].SyncCommitteesExpectedThisPeriod = syncChance * float64(utils.Config.Chain.ClConfig.SyncCommitteeSize)
			}
		}
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

	syncCommitteeAssignments := d.responseCache.GetSyncCommittee(currentSyncPeriod)
	if syncCommitteeAssignments == nil && postAltair {
		return nil, errors.New("sync committee assignments not found")
	}

	// write scheduled sync committee data
	for _, validator := range syncCommitteeAssignments.Data.Validators {
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

		validatorsData[reward.Data.ProposerIndex].BlocksClSyncAggregateReward.Int64 += reward.Data.SyncAggregate
		validatorsData[reward.Data.ProposerIndex].BlocksClSyncAggregateReward.Valid = true

		validatorsData[reward.Data.ProposerIndex].BlocksClAttestestationsReward.Int64 += reward.Data.Attestations
		validatorsData[reward.Data.ProposerIndex].BlocksClAttestestationsReward.Valid = true

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
		if block.Data.Message.Slot != 0 { // Special case to exclude genesis block as Validator 0 did not propose it
			validatorsData[block.Data.Message.ProposerIndex].BlocksProposed.Int16++
			validatorsData[block.Data.Message.ProposerIndex].BlocksProposed.Valid = true
			validatorsData[block.Data.Message.ProposerIndex].LastSubmittedDutyEpoch = utils.NullInt32(int32(block.Data.Message.Slot / utils.Config.Chain.ClConfig.SlotsPerEpoch))
		}

		for depositIndex, depositData := range block.Data.Message.Body.Deposits {
			// Properly verify that deposit is valid:
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
				d.log.Error(fmt.Errorf("deposit at index %d in slot %v is invalid: %v (domain: %x, PublicKey: %s, WithdrawalCredentials: %s, Amount: %d, Signature: %s)",
					depositIndex, block.Data.Message.Slot, err, domain, depositData.Data.Pubkey, depositData.Data.WithdrawalCredentials, depositData.Data.Amount, depositData.Data.Signature), "", 0)

				// if the validator hat a valid deposit prior to the current one, count the invalid towards the balance
				if validatorsData[pubkeyToIndexMapNewlyActivatedValidators[string(depositData.Data.Pubkey)]].DepositsCount.Int16 > 0 {
					d.log.Infof("validator had a valid deposit in some earlier block of the epoch, count the invalid towards the balance")
				} else if _, ok := pubkeyToIndexMapOldValidators[string(depositData.Data.Pubkey)]; ok {
					d.log.Infof("validator had a valid deposit in some block prior to the current epoch, count the invalid towards the balance")
				} else {
					d.log.Infof("validator did not have a prior valid deposit, do not count the invalid towards the balance")
					continue
				}
			}

			validator_index, ok := pubkeyToIndexMapNewlyActivatedValidators[string(depositData.Data.Pubkey)]
			if !ok {
				validator_index, ok = pubkeyToIndexMapOldValidators[string(depositData.Data.Pubkey)]
				if !ok {
					return nil, errors.New("proposer index out of range")
				}
			}
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

					validatorsData[validator_index].InclusionDelaySum = utils.NullInt16(int16(block.Data.Message.Slot - attestation.Data.Slot - 1))

					validatorsData[validator_index].LastSubmittedDutyEpoch = utils.NullInt32(int32(attestation.Data.Slot / utils.Config.Chain.ClConfig.SlotsPerEpoch))

					optimalInclusionDistance := 0
					for i := attestation.Data.Slot + 1; i < block.Data.Message.Slot; i++ {
						if _, ok := data.missedslots[i]; ok {
							optimalInclusionDistance++
						} else {
							break
						}
					}

					validatorsData[validator_index].OptimalInclusionDelay = utils.NullInt16(int16(optimalInclusionDistance))
				}
			}
		}

		// attester slashings "done" by proposer (slashed by)
		for _, data := range block.Data.Message.Body.AttesterSlashings {
			slashedIndices := data.GetSlashedIndices()
			for _, index := range slashedIndices {
				validatorsData[index].SlashedBy = utils.NullInt32(int32(block.Data.Message.ProposerIndex))
				validatorsData[index].Slashed = true
				validatorsData[index].SlashedViolation = utils.NullInt16(SLASHED_VIOLATION_ATTESTATION)
			}
		}

		// proposer slashings "done" by proposer (slashed by)
		for _, data := range block.Data.Message.Body.ProposerSlashings {
			validatorsData[data.SignedHeader1.Message.ProposerIndex].SlashedBy = utils.NullInt32(int32(block.Data.Message.ProposerIndex))
			validatorsData[data.SignedHeader1.Message.ProposerIndex].Slashed = true
			validatorsData[data.SignedHeader1.Message.ProposerIndex].SlashedViolation = utils.NullInt16(SLASHED_VIOLATION_PROPOSER)
		}
	}

	// write attestation rewards data
	for _, attestationReward := range data.attestationRewards {
		validator_index := attestationReward.ValidatorIndex
		if validator_index >= size {
			return nil, errors.New("proposer index out of range")
		}

		validatorsData[validator_index].AttestationsHeadReward = utils.NullInt32(attestationReward.Head)
		validatorsData[validator_index].AttestationsSourceReward = utils.NullInt32(attestationReward.Source)
		validatorsData[validator_index].AttestationsTargetReward = utils.NullInt32(attestationReward.Target)
		validatorsData[validator_index].AttestationsInactivityPenalty = utils.NullInt32(attestationReward.Inactivity)
		validatorsData[validator_index].AttestationsInclusionsReward = utils.NullInt32(attestationReward.InclusionDelay)
		validatorsData[validator_index].AttestationReward = utils.NullInt64(int64(attestationReward.Head + attestationReward.Source + attestationReward.Target + attestationReward.Inactivity + attestationReward.InclusionDelay))

		idealRewardsOfValidator, ok := data.idealAttestationRewards[int64(data.currentEpochStateEnd.Data[validator_index].Validator.EffectiveBalance)]
		if !ok {
			return nil, errors.New("ideal reward not found")
		}

		validatorsData[validator_index].AttestationsIdealHeadReward = utils.NullInt32(idealRewardsOfValidator.Head)
		validatorsData[validator_index].AttestationsIdealTargetReward = utils.NullInt32(idealRewardsOfValidator.Target)
		validatorsData[validator_index].AttestationsIdealSourceReward = utils.NullInt32(idealRewardsOfValidator.Source)
		validatorsData[validator_index].AttestationsIdealInactivityPenalty = utils.NullInt32(idealRewardsOfValidator.Inactivity)
		validatorsData[validator_index].AttestationsIdealInclusionsReward = utils.NullInt32(idealRewardsOfValidator.InclusionDelay)
		validatorsData[validator_index].AttestationIdealReward = utils.NullInt64(int64(idealRewardsOfValidator.Head + idealRewardsOfValidator.Source + idealRewardsOfValidator.Target + idealRewardsOfValidator.Inactivity + idealRewardsOfValidator.InclusionDelay))

		if attestationReward.Head > 0 {
			validatorsData[validator_index].AttestationHeadExecuted = utils.NullInt16(1)
			validatorsData[validator_index].AttestationsExecuted = utils.NullInt16(1)
		}
		if attestationReward.Source > 0 {
			validatorsData[validator_index].AttestationSourceExecuted = utils.NullInt16(1)
			validatorsData[validator_index].AttestationsExecuted = utils.NullInt16(1)
		}
		if attestationReward.Target > 0 {
			validatorsData[validator_index].AttestationTargetExecuted = utils.NullInt16(1)
			validatorsData[validator_index].AttestationsExecuted = utils.NullInt16(1)
		}
	}

	return validatorsData, nil
}

func storeClBlockRewards(data map[uint64]*constypes.StandardBlockRewardsResponse) error {
	tx, err := db.AlloyWriter.Beginx()
	if err != nil {
		return errors.Wrap(err, "failed to start cl blocks transaction")
	}
	defer utils.Rollback(tx)

	for slot, rewards := range data {
		_, err := tx.Exec(`
			INSERT INTO consensus_payloads (slot, cl_attestations_reward, cl_sync_aggregate_reward, cl_slashing_inclusion_reward)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (slot) DO NOTHING
		`, slot, rewards.Data.Attestations, rewards.Data.SyncAggregate, rewards.Data.AttesterSlashings+rewards.Data.ProposerSlashings)
		if err != nil {
			return errors.Wrap(err, "failed to insert cl blocks data")
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit cl blocks transaction")
	}

	return nil
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
	Epoch   uint64
}

type DataEpochProcessed struct {
	Data  []*validatorDashboardDataRow
	Epoch uint64
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

	BlockScheduled                   sql.NullInt16 // done
	BlocksProposed                   sql.NullInt16 // done
	BlocksExpectedThisEpoch          float64       // done
	SyncCommitteesExpectedThisPeriod float64       // done

	BlocksClReward                sql.NullInt64 // done
	BlocksClAttestestationsReward sql.NullInt64 // done
	BlocksClSyncAggregateReward   sql.NullInt64 // done

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

	InclusionDelaySum     sql.NullInt16 // done
	OptimalInclusionDelay sql.NullInt16 // done
}

const SLASHED_VIOLATION_ATTESTATION = 1
const SLASHED_VIOLATION_PROPOSER = 2

const RawSyncCommitteeCacheKey = "sync_committee_period_"

type ResponseCache struct {
	cache map[string]any
}

func (r *ResponseCache) SetSyncCommittee(period uint64, data *constypes.StandardSyncCommitteesResponse) {
	r.cache[r.GetSyncCommitteeCacheKey(period)] = data
}

func (r *ResponseCache) GetSyncCommittee(period uint64) *constypes.StandardSyncCommitteesResponse {
	temp, ok := r.cache[r.GetSyncCommitteeCacheKey(period)]
	if !ok {
		return nil
	}
	return temp.(*constypes.StandardSyncCommitteesResponse)
}

func (r *ResponseCache) GetSyncCommitteeCacheKey(period uint64) string {
	return fmt.Sprintf("%s%d", RawSyncCommitteeCacheKey, period)
}

func refreshMaterializedSlashedByCounts() error {
	tx, err := db.AlloyWriter.Beginx()
	if err != nil {
		return errors.Wrap(err, "failed to start transaction")
	}
	defer utils.Rollback(tx)

	_, err = tx.Exec(`
		CREATE MATERIALIZED VIEW IF NOT EXISTS validator_dashboard_data_rolling_total_slashedby_count AS
		SELECT slashed_by, COUNT(*) as slashed_amount
		FROM validator_dashboard_data_rolling_total
		WHERE slashed = true AND slashed_by IS NOT NULL
		GROUP BY slashed_by;
		CREATE INDEX IF NOT EXISTS idx_validator_dashboard_data_rolling_total_slashedby_count_slashed_by ON validator_dashboard_data_rolling_total_slashedby_count(slashed_by);

		CREATE MATERIALIZED VIEW IF NOT EXISTS validator_dashboard_data_rolling_daily_slashedby_count AS
		SELECT slashed_by, COUNT(*) as slashed_amount
		FROM validator_dashboard_data_rolling_daily
		WHERE slashed = true AND slashed_by IS NOT NULL
		GROUP BY slashed_by;
		CREATE INDEX IF NOT EXISTS idx_validator_dashboard_data_rolling_daily_slashedby_count_slashed_by ON validator_dashboard_data_rolling_daily_slashedby_count(slashed_by);

		CREATE MATERIALIZED VIEW IF NOT EXISTS validator_dashboard_data_rolling_weekly_slashedby_count AS
		SELECT slashed_by, COUNT(*) as slashed_amount
		FROM validator_dashboard_data_rolling_weekly
		WHERE slashed = true AND slashed_by IS NOT NULL
		GROUP BY slashed_by;
		CREATE INDEX IF NOT EXISTS idx_validator_dashboard_data_rolling_weekly_slashedby_count_slashed_by ON validator_dashboard_data_rolling_weekly_slashedby_count(slashed_by);

		CREATE MATERIALIZED VIEW IF NOT EXISTS validator_dashboard_data_rolling_monthly_slashedby_count AS
		SELECT slashed_by, COUNT(*) as slashed_amount
		FROM validator_dashboard_data_rolling_monthly
		WHERE slashed = true AND slashed_by IS NOT NULL
		GROUP BY slashed_by;
		CREATE INDEX IF NOT EXISTS idx_validator_dashboard_data_rolling_monthly_slashedby_count_slashed_by ON validator_dashboard_data_rolling_monthly_slashedby_count(slashed_by);

		CREATE MATERIALIZED VIEW IF NOT EXISTS validator_dashboard_data_epoch_slashedby_count AS
		SELECT epoch, slashed_by, COUNT(*) as slashed_amount
		FROM validator_dashboard_data_epoch
		WHERE slashed = true AND slashed_by IS NOT NULL
		GROUP BY epoch, slashed_by;
		CREATE INDEX IF NOT EXISTS idx_validator_dashboard_data_epoch_slashedby_count_epoch_slashed_by ON validator_dashboard_data_epoch_slashedby_count(epoch, slashed_by);

		REFRESH MATERIALIZED VIEW validator_dashboard_data_rolling_total_slashedby_count;
		REFRESH MATERIALIZED VIEW validator_dashboard_data_rolling_daily_slashedby_count;
		REFRESH MATERIALIZED VIEW validator_dashboard_data_rolling_weekly_slashedby_count;
		REFRESH MATERIALIZED VIEW validator_dashboard_data_rolling_monthly_slashedby_count;
		REFRESH MATERIALIZED VIEW validator_dashboard_data_epoch_slashedby_count;
	`)

	if err != nil {
		return errors.Wrap(err, "failed to refresh materialized views")
	}
	return tx.Commit()
}

// Can be used to backfill old missing cl block rewards
// Commented out since this is a one time operation, kept in in case we need it again
// func (d *dashboardData) backfillCLBlockRewards() {
// 	upTo := 1731488
// 	startFrom := 1328805
// 	batchSize := 8
// 	parallelization := 8

// 	blocksChan := make(chan map[uint64]*constypes.StandardBlockRewardsResponse, 1)

// 	err := db.AlloyWriter.Get(&startFrom, "SELECT last_slot FROM meta_slot_export")
// 	if err != nil {
// 		d.log.Error(err, "failed to get last slot from meta_slot_export", 0)
// 		return
// 	}

// 	// re export 317201 +- 20000

// 	go func() {
// 		for i := startFrom; i < upTo+batchSize; i += batchSize {
// 			batched := map[uint64]*constypes.StandardBlockRewardsResponse{}
// 			mutex := &sync.Mutex{}

// 			errgroup := &errgroup.Group{}
// 			errgroup.SetLimit(parallelization)

// 			for slot := i; slot < i+batchSize; slot++ {
// 				slot := slot
// 				errgroup.Go(func() error {
// 					blockReward, err := d.CL.GetPropoalRewards(slot)
// 					if err != nil {
// 						httpErr := network.SpecificError(err)
// 						if httpErr != nil && httpErr.StatusCode == 404 {
// 							return nil
// 						}

// 						d.log.Error(err, "can not get block reward data", 0, map[string]interface{}{"slot": slot})
// 						return err
// 					}

// 					mutex.Lock()
// 					batched[uint64(slot)] = blockReward
// 					mutex.Unlock()

// 					return nil
// 				})
// 			}

// 			err := errgroup.Wait()
// 			if err != nil {
// 				d.log.Error(err, "failed to backfill cl block rewards", 0)
// 				close(blocksChan)
// 			}

// 			blocksChan <- batched
// 		}
// 	}()

// 	go func() {
// 		for blockReward := range blocksChan {
// 			for {
// 				err := storeClBlockRewards(blockReward)
// 				if err != nil {
// 					d.log.Error(err, "failed to store cl block rewards", 0)
// 					continue
// 				}
// 				break
// 			}
// 			highestSlot := uint64(0)
// 			for slot := range blockReward {
// 				if slot > highestSlot {
// 					highestSlot = slot
// 				}
// 			}

// 			if highestSlot%100 < uint64(batchSize) {
// 				d.log.Infof("processed blocks, height: %d", highestSlot)
// 			}

// 			if highestSlot%10000 < uint64(batchSize) {
// 				_, err := db.AlloyWriter.Exec("UPDATE meta_slot_export SET last_slot = $1", highestSlot)
// 				if err != nil {
// 					d.log.Error(err, "failed to update last slot in meta_slot_export", 0)
// 				}
// 			}
// 		}
// 	}()
// }
