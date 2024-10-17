package modules

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"
	"sync"
	"time"

	ch "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/consapi/network"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	edb "github.com/gobitfly/beaconchain/pkg/exporter/db"
	"github.com/google/uuid"

	//"github.com/fjl/memsize/memsizeui"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
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

// ----------- END OF DEBUG FLAGS ------------

// How many epochs will be fetched in parallel from the node (relevant for backfill and rolling tail fetching). We are fetching the head epoch and
// one epoch for each rolling table (tail), so if you want to fetch all epochs in one go (and your node can handle that) set this to at least 5.
var epochFetchParallelism = 12

// Fetching one epoch consists of multiple calls. You can define how many concurrent calls each epoch fetch will do. Keep in mind that
// the total number of concurrent requests is epochF<etchParallelism * epochFetchParallelismWithinFetch
const epochFetchParallelismWithinFetch = 8 * 5

// How many epochs will be written in parallel to the database
const epochWriteParallelism = 5

// How many epochs get written to a single batch
const epochWriteBatchSize = 1

// how many mantainEpochs will be executed in parallel
const databaseEpochMaintainParallelism = 12 / 6

// How many epoch aggregations will be executed in parallel (e.g. total, hour, day, each rolling table)
const databaseAggregationParallelism = 2

// How many epochFetchParallelism iterations will be written before a new aggregation will be triggered during backfill. This can speed up backfill as writing epochs to db is fast and we can delay
// aggregation for a couple iterations. Don't set too high or else epoch table will grow to large and will be a bottleneck.
// Set to 0 to disable and write after every iteration. Recommended value for this is 1 or maybe 2.
// Try increasing this one by one if node_fetch_time < agg_and_storage_time until it targets roughly agg_and_storage_time = node_fetch_time
// Try 0 if agg_and_storage_time is < node_fetch_time
const backfillMaxUnaggregatedIterations = 1

func getNodeId[T int64 | uint64](a T) int {
	return int(a) % len(utils.Config.Indexer.Node)
}

type dashboardData struct {
	ModuleContext
	log               ModuleLog
	signingDomain     []byte
	epochWriter       *epochWriter
	headEpochQueue    chan uint64
	backFillCompleted bool
	phase0HotfixMutex sync.Mutex
}

func NewDashboardDataModule(moduleContext ModuleContext) ModuleInterface {
	temp := &dashboardData{
		ModuleContext: moduleContext,
	}
	temp.log = ModuleLog{module: temp}
	if strings.ToLower(os.Getenv("LOG_LEVEL")) == "debug" {
		logrus.SetLevel(logrus.DebugLevel)
	}
	if utils.Config.FetchEpochsOverwrite != 0 {
		epochFetchParallelism = utils.Config.FetchEpochsOverwrite
	}
	// When a new epoch gets exported the very first step is to export it to the db via epochWriter
	temp.epochWriter = newEpochWriter(temp)

	// Then those aggregators below use the epoch data to aggregate it into the respective tables

	// This channel is used to queue up epochs from chain head that need to be exported
	temp.headEpochQueue = make(chan uint64, 100)

	// Indicates whether the initial backfill - which is checked when starting the exporter - has completed
	// and the exporter can start listening for new head epochs to be processed
	temp.backFillCompleted = false

	return temp
}

type Task struct {
	UUID     uuid.UUID `db:"uuid"`
	Hostname string    `db:"hostname"`
	Priority int64     `db:"priority"`
	StartTs  time.Time `db:"start_ts"`
	EndTs    time.Time `db:"end_ts"`
	Status   string    `db:"status"`
}

func (d *dashboardData) Init() error {
	for {
		// get tasks
		tasks, err := d.getTasksFromDb()
		if err != nil {
			d.log.Error(err, "failed to get tasks from db", 0)
			time.Sleep(10 * time.Second)
			continue
		}
		// debug log tasks
		for _, task := range tasks {
			// nicely formatted log message. you can see the schema above
			d.log.Infof("uuid: %s, priority: %d, start_ts: %s, end_ts: %s, status: %s", task.UUID, task.Priority, task.StartTs, task.EndTs, task.Status)
		}

		// get next todo task - task that is not marked as complete
		var nextTask *Task
		for _, task := range tasks {
			if task.Status != "completed" {
				nextTask = &task
				break
			}
		}
		if nextTask == nil {
			d.log.Warnf("no tasks to do, idling and checking again in 60 seconds")
			time.Sleep(60 * time.Second)
			continue
		}
		if nextTask.Status == "running" {
			// we will will overtake this task, we assume that only one exporter ever runs per hostname
			d.log.Warnf("task %s is already running, taking over and continuing", nextTask.UUID)
		} else {
			// update
			_ = db.ClickHouseWriter.MustExec("ALTER TABLE _exporter_tasks UPDATE status = 'running' WHERE uuid = ?", nextTask.UUID)
		}
		// start export using startTs and endTs
		err = d.backfillHeadEpochData(nextTask.StartTs, nextTask.EndTs)
		if err != nil {
			// try again in 10 seconds
			d.log.Error(err, "failed to backfill epoch data", 0)
			time.Sleep(10 * time.Second)
			continue
		}
		// update task to completed
		_ = db.ClickHouseWriter.MustExec("ALTER TABLE _exporter_tasks UPDATE status = 'completed' WHERE uuid = ?", nextTask.UUID)
	}
	/*
		go func() {
			d.processHeadQueue()
		}()
		go func() {
			start := time.Now()
			for {
				var upToEpochPtr *uint64 = nil // nil will backfill back to head
				if debugTargetBackfillEpoch > 0 {
					upToEpoch := debugTargetBackfillEpoch
					upToEpochPtr = &upToEpoch
				}

				done, err := d.backfillHeadEpochData(upToEpochPtr)
				if err != nil {
					d.log.Error(err, "failed to backfill epoch data", 0)
					metrics.Errors.WithLabelValues("exporter_v25dash_backfill_fail").Inc()
					time.Sleep(10 * time.Second)
					continue
				}

				if done {
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
		}()
	*/

	return nil
}

/*

func (d *dashboardData) processHeadQueue() {
	reachedHead := false
	for {
		epoch := <-d.headEpochQueue

		// skip any intermediate epochs
		for len(d.headEpochQueue) > 1 {
			epoch = <-d.headEpochQueue
		}
		if d.backFillCompleted {
			if len(d.headEpochQueue) == 0 && !reachedHead {
				d.log.Infof("exporter is at head of the chain")
				reachedHead = true
			}

			startTime := time.Now()
			d.log.Infof("exporting dashboard epoch data for epoch %d", epoch)
			for {
				_, err := d.backfillHeadEpochData(&epoch)
				if err != nil {
					d.log.Error(err, "failed to backfill head epoch data", 0, map[string]interface{}{"epoch": epoch})
					metrics.Errors.WithLabelValues("exporter_v25dash_backfill_fail").Inc()
					time.Sleep(time.Second * 10)
					continue
				}
				// update rollings using refreshRollings
				err = d.refreshAllRollings(epoch)
				if err != nil {
					d.log.Error(err, "failed to refresh rollings", 0, map[string]interface{}{"epoch": epoch})
					metrics.Errors.WithLabelValues("exporter_v25dash_refresh_rollings_fail").Inc()
					time.Sleep(time.Second * 10)
					continue
				}

				break
			}

			d.log.Infof("[time] completed dashboard epoch data for epoch %d in %v", epoch, time.Since(startTime))
			metrics.State.WithLabelValues("exporter_v25dash_last_exported_epoch").Set(float64(epoch))
		}
		// we do maintenance tasks even if backfill is not completed
		s := time.Now()
		utils.SendMessage(fmt.Sprintf("🔧 v2.5 Dashboard %s - Starting maintenance tasks", utils.Config.Chain.Name), &utils.Config.InternalAlerts)
		var staleEpochs []uint64
		for {
			var err error
			staleEpochs, err = edb.GetStaleEpochs(epoch, 450) // 2 days, for mainnet.
			if err != nil {
				d.log.Error(err, "failed to get stale epochs", 0)
				metrics.Errors.WithLabelValues("exporter_v25dash_stale_epochs_fail").Inc()
				time.Sleep(time.Second * 10)
				continue
			}
			break
		}
		utils.SendMessage(fmt.Sprintf("🔧 v2.5 Dashboard %s - Found stale epochs: `%v`", utils.Config.Chain.Name, staleEpochs), &utils.Config.InternalAlerts)
		for {
			err := d.maintainEpochs(staleEpochs)
			if err != nil {
				d.log.Error(err, "failed to maintain epochs", 0)
				metrics.Errors.WithLabelValues("exporter_v25dash_maintenance_fail").Inc()
				time.Sleep(time.Second * 10)
				continue
			}
			break
		}
		d.log.Infof("[time] completed maintenance tasks in %v", time.Since(s))
		metrics.TaskDuration.WithLabelValues("exporter_v25dash_maintenance").Observe(time.Since(s).Seconds())
		utils.SendMessage(fmt.Sprintf("🔧 v2.5 Dashboard %s - Completed maintenance tasks in %s", utils.Config.Chain.Name, time.Since(s)), &utils.Config.InternalAlerts)

	}
	d.log.Fatal(errors.Errorf("head epoch queue closed"), "head epoch queue closed", 0)
}
*/

// fetches and processes epoch data and provides them via the nextDataChan
// expects ordered epochs in ascending order
func (d *dashboardData) epochDataFetcher(epochs []uint64, epochFetchParallelism int, nextDataChan chan []db.VDBDataEpochColumns) {
	// group epochs into parallel worker groups
	groups := getEpochParallelGroups(epochs, epochFetchParallelism)
	for _, gapGroup := range groups {
		var rawData MultiEpochData
		for {
			rawData = NewMultiEpochData(len(gapGroup.Epochs))

			errGroup := &errgroup.Group{}

			start := time.Now()

			errGroup.Go(func() error {
				err := d.getDataForEpochRange(gapGroup.Epochs[0], gapGroup.Epochs[len(gapGroup.Epochs)-1], &rawData)
				if err != nil {
					d.log.Error(err, "failed to get epoch data", 0, map[string]interface{}{"epoch": gapGroup.Epochs})
					metrics.Errors.WithLabelValues("exporter_v25dash_node_fail").Inc()
					time.Sleep(time.Second * 10)
					return err
				}
				d.log.Infof("epoch data fetcher, retrieved data for epochs %v", gapGroup.Epochs)
				return nil
			})

			err := errGroup.Wait()
			d.log.Infof("epoch data fetcher, fetched data for epochs %v in %v (thats %v per epoch)", gapGroup.Epochs, time.Since(start), time.Since(start)/time.Duration(len(gapGroup.Epochs)))
			metrics.TaskDuration.WithLabelValues("exporter_v25dash_node_fetch").Observe(time.Since(start).Seconds())
			metrics.TaskDuration.WithLabelValues("exporter_v25dash_node_fetch_per_epoch").Observe((time.Since(start) / time.Duration(len(gapGroup.Epochs))).Seconds())

			if err != nil {
				d.log.Error(err, "failed to fetch epoch data", 0, map[string]interface{}{"epoch": gapGroup.Epochs})
				metrics.Errors.WithLabelValues("exporter_v25dash_node_fail").Inc()
				continue
			}
			break
		}
		//msg := conutils.PrintFieldSizes(rawData)
		//utils.SendMessage(fmt.Sprintf("📊 v2.5 Dashboard %s - Fetched data for epochs %v\n```\n%s```", utils.Config.Chain.Name, gapGroup.Epochs, msg), &utils.Config.InternalAlerts)
		// process
		var data []db.VDBDataEpochColumns
		for {
			l := len(gapGroup.Epochs) / epochWriteBatchSize
			if l == 0 {
				l = 1
			}
			data = make([]db.VDBDataEpochColumns, l)
			err := d.processRunner(&rawData, &data)
			if err != nil {
				d.log.Error(err, "failed to process epoch data", 0, map[string]interface{}{"epoch": gapGroup.Epochs})
				metrics.Errors.WithLabelValues("exporter_v25dash_process_fail").Inc()
				time.Sleep(time.Second * 10)
				continue
			}
			break
		}
		//msg = conutils.PrintFieldSizes(data)
		//utils.SendMessage(fmt.Sprintf("📊 v2.5 Dashboard %s - Processed data for epochs %v\n```\n%s```", utils.Config.Chain.Name, gapGroup.Epochs, msg), &utils.Config.InternalAlerts)
		nextDataChan <- data
	}
}

// breaks epoch down in groups of size parallelism
// for example epochs: 1,2,3,5,7,9,10,11,12,13,20,30 with parallelism = 4 breaks down to groups:
// 0 = 1,2,3
// 1 = 5
// 2 = 7
// 3 = 9,10,11,12
// 4 = 14
// 5 = 20
// 6 = 30
// expects ordered epochs in ascending order
func getEpochParallelGroups(epochs []uint64, parallelism int) []EpochParallelGroup {
	parallelGroups := make([]EpochParallelGroup, 0, len(epochs)/4)

	// 1. Group sequential epochs in parallel groups
	for i := 0; i < len(epochs); i++ {
		group := EpochParallelGroup{
			Epochs: []uint64{epochs[i]},
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

	return parallelGroups
}

func (d *dashboardData) maintainGroupsTask(startEpoch, endEpoch uint64, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		stale, err := edb.GetStaleEpochs(startEpoch, endEpoch, 1000)
		if err != nil {
			d.log.Error(err, "failed to get stale epochs", 0)
			time.Sleep(10 * time.Second)
			continue
		}
		// limit to x parallel runs
		g := &errgroup.Group{}
		g.SetLimit(6)
		d.log.Infof("maintainGroupsTask, found %d stale epochs", len(stale))
		for _, e := range stale {
			epoch := e
			g.Go(func() error {
				d.log.Infof("maintainGroupsTask, maintaining epoch %d", epoch)
				for {
					err = d.maintainEpochGroup([]uint64{epoch})
					if err != nil {
						d.log.Error(err, "failed to maintain epoch", 0, map[string]interface{}{"epoch": epoch})
						utils.SendMessage(fmt.Sprintf("🔧 v2.5 Dashboard %s - Failed to maintain epoch %d", utils.Config.Chain.Name, epoch), &utils.Config.InternalAlerts)
						time.Sleep(10 * time.Second)
						continue
					}
					break
				}
				d.log.Infof("maintainGroupsTask, done maintaining epoch %d", epoch)

				return nil
			})
		}
		g.Wait()
		// check if there are the expected amount of epochs in the final table
		// if not, repeat
		// if yes, break
		is_done, err := edb.IsEpochsInFinalTable(startEpoch, endEpoch)
		if !is_done || err != nil {
			if err != nil {
				d.log.Error(err, "failed to check if epochs are in final table", 0)
			}
			if len(stale) != 0 {
				d.log.Infof("maintainGroupsTask, epochs %d to %d not in final table, repeating maintenance (did %d this loop)", startEpoch, endEpoch, len(stale))
				utils.SendMessage(fmt.Sprintf("🔧 v2.5 Dashboard %s - Epochs %d to %d not in final table, repeating maintenance (did %d this loop)", utils.Config.Chain.Name, startEpoch, endEpoch, len(stale)), &utils.Config.InternalAlerts)
				time.Sleep(30 * time.Second)
			} else {
				d.log.Infof("maintainGroupsTask, epochs %d to %d not in final table, repeating maintenance", startEpoch, endEpoch)
				time.Sleep(10 * time.Second)
			}
			continue
		}
		d.log.Infof("maintainGroupsTask, epochs %d to %d are in final table", startEpoch, endEpoch)
		break
	}

}

// can be used to start a backfill up to epoch
// returns true if there was nothing to backfill, otherwise returns false
// if upToEpoch is nil, it will backfill until the latest finalized epoch
func (d *dashboardData) backfillHeadEpochData(startTs time.Time, endTs time.Time) error {
	// rough sketch of the message to report
	// hostnanme (chain) - Starting backfill between startTs and endTs
	// utils.SendMessage(fmt.Sprintf("🔙 v2.5 Dashboard %s - Starting backfill between `%s` and `%s`", utils.Config.Chain.Name, startTs, endTs), &utils.Config.InternalAlerts)
	backfillToChainFinalizedHead := endTs.After(time.Now())
	wg := sync.WaitGroup{}
	wg.Add(1)
	// turn startTs and endTs into epochs
	var startEpoch, endEpoch int64
	startEpoch = utils.TimeToEpoch(startTs)
	if backfillToChainFinalizedHead {
		res, err := d.CL[0].GetFinalityCheckpoints("head")
		if err != nil {
			return errors.Wrap(err, "failed to get finalized checkpoint")
		}
		if utils.IsByteArrayAllZero(res.Data.Finalized.Root) {
			return errors.New("network not finalized yet")
		}
		d.log.Infof("node reported finalized epoch %d", res.Data.Finalized.Epoch)
		res.Data.Finalized.Epoch -= 2 // backfill up to the last finalized epoch. checkpoint is one epoch after, so we need to backfill one less
		endEpoch = int64(res.Data.Finalized.Epoch)
		d.headEpochQueue <- res.Data.Finalized.Epoch
		d.log.Infof("backfilling head epoch data up to epoch %d", endEpoch)
	} else {
		endEpoch = utils.TimeToEpoch(endTs)
	}
	go d.maintainGroupsTask(uint64(startEpoch), uint64(endEpoch), &wg)
	gaps, err := edb.GetMissingEpochsBetween(startEpoch, endEpoch)
	if err != nil {
		return errors.Wrap(err, "failed to get epoch gaps")
	}

	if len(gaps) > 0 {
		d.log.Infof("Epoch dashboard data %d gaps found, backfilling gaps in the range fom epoch %d to %d", len(gaps), gaps[0], gaps[len(gaps)-1])

		// get epochs data
		var nextDataChan chan []db.VDBDataEpochColumns = make(chan []db.VDBDataEpochColumns, 0)
		go func() {
			d.epochDataFetcher(gaps, epochFetchParallelism, nextDataChan)
		}()

		// save epochs data
		for {
			d.log.Info("storage waiting for data from fetcher")
			datas := <-nextDataChan

			done := containsEpoch(datas, gaps[len(gaps)-1]) // if the last epoch to fetch is in the result set, mark as job completed

			if len(datas) > 0 {
				d.log.Info("storage got data, writing epoch data")
				d.writeEpochDatas(datas)
			}
			if done {
				break
			}
		}
	}

	// wait on maintenance tasks
	wg.Wait()
	if backfillToChainFinalizedHead {
		// never done, try again in a bit
		return errors.New("head mode, try again in a bit")
	}

	gaps, err = edb.GetMissingEpochsBetween(startEpoch, endEpoch)
	if err != nil {
		return errors.Wrap(err, "failed to get epoch gaps")
	}
	if len(gaps) > 0 {
		return errors.New("backfill not completed")
	}
	if backfillToChainFinalizedHead {
		// never done, try again in a bit
		return errors.New("head mode, try again in a bit")
	}

	// debug message to discord
	utils.SendMessage(fmt.Sprintf("🔙 v2.5 Dashboard %s - Completed backfill between `%s` and `%s`", utils.Config.Chain.Name, startTs, endTs), &utils.Config.InternalAlerts)
	return nil
}

func containsEpoch(d []db.VDBDataEpochColumns, epoch uint64) bool {
	for i := 0; i < len(d); i++ {
		for j := 0; j < len(d[i].EpochsContained); j++ {
			if d[i].EpochsContained[j] == epoch {
				return true
			}
		}
	}
	return false
}

var EpochsWritten int
var FirstEpochWritten *time.Time

func getBatchId(epoch int64) int {
	// 3 nodes, split so each node has 1 insert
	// when we fetch 9 this results in 3 epoch batches
	epochWidth := float64(9) / 3
	return int(float64(epoch) / epochWidth)
}

// stores all passed epoch data, blocks until all data is written without error<<<<
func (d *dashboardData) writeEpochDatas(datas []db.VDBDataEpochColumns) {
	totalStart := time.Now()
	defer func() {
		d.log.Infof("[time] storage, wrote all epoch data in %v", time.Since(totalStart))
		metrics.TaskDuration.WithLabelValues("exporter_v25dash_write_epochs").Observe(time.Since(totalStart).Seconds())
	}()
	if FirstEpochWritten == nil {
		now := time.Now()
		FirstEpochWritten = &now
	}
	debugTimes := sync.Map{}

	start := time.Now()
	errGroup := &errgroup.Group{}
	errGroup.SetLimit(epochWriteParallelism)
	for i := 0; i < len(datas); i++ {
		data := datas[i]
		errGroup.Go(func() error {
			for {
				d.log.Infof("storage, writing epoch data for epochs %v", data.EpochsContained)
				start := time.Now()

				// retry this epoch until no errors occur
				err := d.epochWriter.WriteEpochsData(data.EpochsContained, &data)
				if err != nil {
					d.log.Error(err, "storage, failed to write epoch data", 0, map[string]interface{}{"epoch": data.EpochsContained})
					time.Sleep(time.Second * 10)
					continue
				}

				d.log.Infof("storage, wrote epoch data %d in %v", data.EpochsContained, time.Since(start))
				// check if epoch % 225 true, if yes send message
				// 225 epochs is the amount of epochs in a day
				for _, epoch := range data.EpochsContained {
					EpochsWritten++
					if epoch%225 == 0 {
						utils.SendMessage(fmt.Sprintf("<:PauseChamp:771100439599906886>v2.5 Dashboard %s - Completed UTC day `%s` %s/epoch avg", utils.Config.Chain.Name, utils.EpochToTime(epoch).Format("2006-01-02"), time.Since(*FirstEpochWritten)/time.Duration(EpochsWritten)), &utils.Config.InternalAlerts)
						EpochsWritten = 0
						tmpnow := time.Now()
						FirstEpochWritten = &tmpnow
					}
				}

				break
			}
			return nil
		})
	}

	_ = errGroup.Wait() // no errors to handle since it will retry until it resolves without err
	debugTimes.Store("writing", time.Since(start))
	// debug message to discord
	var msg string
	// sort by debug time seen
	sortedKeys := make([]string, 0)
	debugTimes.Range(
		func(key, value any) bool {
			sortedKeys = append(sortedKeys, key.(string))
			return true
		})
	slices.SortFunc(sortedKeys,
		func(a, b string) int {
			av, _ := debugTimes.Load(a)
			bv, _ := debugTimes.Load(b)
			return int((av.(time.Duration) - bv.(time.Duration)).Nanoseconds())
		})
	for _, k := range sortedKeys {
		v, _ := debugTimes.Load(k)
		msg += fmt.Sprintf("%s: %v\n", k,
			v)
	}
	d.log.Infof("debug times:\n%s", msg)
	// send message
	utils.SendMessage(fmt.Sprintf("Debug Times writeEpochData\n```\n%s```", msg), &utils.Config.InternalAlerts)

}

func (d *dashboardData) OnFinalizedCheckpoint(t *constypes.StandardFinalizedCheckpointResponse) error {
	d.log.Infof("finalized checkpoint %v (backfill completed: %v)", t, d.backFillCompleted)
	// random sleep to hit race condition due to load balancer
	time.Sleep(12 * time.Second)

	// Note that "StandardFinalizedCheckpointResponse" event contains the current justified epoch, not the finalized one
	// An epoch becomes finalized once the next epoch gets justified
	// Hence we just listen for new justified epochs here and fetch the latest finalized one from the node
	// Do not assume event.Epoch -1 is finalized by default as it could be that it is not justified
	res, err := d.CL[0].GetFinalityCheckpoints("head")
	if err != nil {
		return err
	}

	latestExported, err := edb.GetLatestDashboardEpoch()
	if err != nil {
		return err
	}

	if latestExported != 0 {
		if res.Data.Finalized.Epoch-2 <= latestExported {
			d.log.Infof("dashboard epoch data already exported for epoch %d", res.Data.Finalized.Epoch)
			return nil
		}
	}
	metrics.State.WithLabelValues("exporter_v25dash_last_finalized_epoch").Set(float64(res.Data.Finalized.Epoch))
	metrics.State.WithLabelValues("exporter_v25dash_safe_to_export_epoch").Set(float64(res.Data.Finalized.Epoch - 2))

	d.headEpochQueue <- res.Data.Finalized.Epoch - 2

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

type MultiEpochData struct {
	// needs sorting
	epochBasedData struct {
		epochs          []uint64
		tarIndices      []int
		tarOffsets      []int
		validatorStates map[int64]constypes.LightStandardValidatorsResponse // epoch => state
		rewards         struct {
			attestationRewards      map[uint64][]constypes.AttestationReward               // epoch => validator index => reward
			attestationIdealRewards map[uint64]map[uint64]constypes.AttestationIdealReward // epoch => effective balance => reward
		}
	}
	validatorBasedData struct {
		// mapping pubkey => validator index
		validatorIndices map[string]uint64
	}
	syncPeriodBasedData struct {
		// sync committee period => assignments
		SyncAssignments map[uint64][]uint64
		// sync committee period => state
		SyncStateEffectiveBalances map[uint64][]uint64
	}
	slotBasedData struct {
		blocks      map[uint64]constypes.LightAnySignedBlock // slotOffset => block, if nil = missed. will include blocks for one more epoch than needed because attestations can be included an epoch later
		assignments struct {
			attestationAssignments map[uint64][][]uint64 // slotOffset => committee index => validator index
			blockAssignments       map[uint64]uint64     // slotOffset => validator index
		}
		rewards struct {
			syncCommitteeRewards map[uint64]constypes.StandardSyncCommitteeRewardsResponse // slotOffset => sync committee rewards
			blockRewards         map[uint64]constypes.StandardBlockRewardsResponse         // slotOffset => block reward data
		}
	}
}

// factory
func NewMultiEpochData(epochCount int) MultiEpochData {
	// allocate all maps
	data := MultiEpochData{}
	data.epochBasedData.validatorStates = make(map[int64]constypes.LightStandardValidatorsResponse, epochCount)
	data.epochBasedData.tarIndices = make([]int, epochCount)
	data.epochBasedData.tarOffsets = make([]int, epochCount)
	data.epochBasedData.rewards.attestationRewards = make(map[uint64][]constypes.AttestationReward, epochCount)
	data.epochBasedData.rewards.attestationIdealRewards = make(map[uint64]map[uint64]constypes.AttestationIdealReward, epochCount)
	slotCount := epochCount * int(utils.Config.Chain.ClConfig.SlotsPerEpoch)
	data.slotBasedData.blocks = make(map[uint64]constypes.LightAnySignedBlock, slotCount)
	data.slotBasedData.assignments.attestationAssignments = make(map[uint64][][]uint64, slotCount)
	data.slotBasedData.assignments.blockAssignments = make(map[uint64]uint64, slotCount)
	data.slotBasedData.rewards.syncCommitteeRewards = make(map[uint64]constypes.StandardSyncCommitteeRewardsResponse, slotCount)
	data.slotBasedData.rewards.blockRewards = make(map[uint64]constypes.StandardBlockRewardsResponse, slotCount)
	data.validatorBasedData.validatorIndices = make(map[string]uint64)
	return data
}

// takes a chan of strings that define what dependenices have been processed
func (d *dashboardData) processRunner(data *MultiEpochData, tar *[]db.VDBDataEpochColumns) error {
	d.log.Info("starting processRunner")
	var err error
	latestTarIndex := 0
	bulkValiCount := 0
	bulkEpochsContained := make([]uint64, 0)
	// prepare tar
	for i, epoch := range data.epochBasedData.epochs {
		tarIndex := int(epoch-data.epochBasedData.epochs[0]) / epochWriteBatchSize
		// debug log
		d.log.Infof("epoch %d, tarIndex %d", epoch, tarIndex)
		if tarIndex > latestTarIndex {
			d.log.Infof("allocating tar index %d with %d entries", latestTarIndex, bulkValiCount)
			(*tar)[latestTarIndex], err = db.NewVDBDataEpochColumns(bulkValiCount)
			if err != nil {
				return errors.Wrap(err, "failed to allocate new VDBDataEpochColumns")
			}
			(*tar)[latestTarIndex].EpochsContained = bulkEpochsContained
			latestTarIndex = tarIndex
			bulkEpochsContained = make([]uint64, 0)
			bulkValiCount = 0
		}
		d.log.Infof("adding epoch %d to tar index %d", epoch, tarIndex)
		bulkEpochsContained = append(bulkEpochsContained, epoch)
		data.epochBasedData.tarIndices[i] = tarIndex
		data.epochBasedData.tarOffsets[i] = bulkValiCount
		bulkValiCount += len(data.epochBasedData.validatorStates[int64(epoch)].Data)
	}
	if bulkValiCount > 0 {
		d.log.Infof("leftover valis, allocating tar index %d with %d entries", latestTarIndex, bulkValiCount)
		(*tar)[latestTarIndex], err = db.NewVDBDataEpochColumns(bulkValiCount)
		if err != nil {
			return errors.Wrap(err, "failed to allocate new VDBDataEpochColumns")
		}
		(*tar)[latestTarIndex].EpochsContained = bulkEpochsContained
	}
	// sanity check that the first tar has epochWriteBatchSize epochscontained
	if len((*tar)[0].EpochsContained) != epochWriteBatchSize {
		d.log.Infof("first tar epochs contained: %v", (*tar)[0].EpochsContained)
		return fmt.Errorf("first tar has %d epochs contained, expected %d", len((*tar)[0].EpochsContained), epochWriteBatchSize)
	}
	d.log.Infof("prepared tar with %d entries", len(*tar))
	start := time.Now()
	debugTimes := sync.Map{}
	// errgroup
	g := &errgroup.Group{}
	g.Go(func() error {
		start := time.Now()
		err := d.processValidatorStates(data, tar)
		if err != nil {
			return fmt.Errorf("error in processValidatorStates: %w", err)
		}
		d.log.Infof("processed validator states in %v", time.Since(start))
		debugTimes.Store("validatorStates", time.Since(start))
		return nil
	})
	g.Go(func() error {
		start := time.Now()
		err := d.processScheduledAttestations(data, tar)
		if err != nil {
			return fmt.Errorf("error in processScheduledAttestations: %w", err)
		}
		d.log.Infof("processed scheduled attestations in %v", time.Since(start))
		debugTimes.Store("scheduledAttestations", time.Since(start))
		return nil
	})
	g.Go(func() error {
		start := time.Now()
		err := d.processBlocks(data, tar)
		if err != nil {
			return fmt.Errorf("error in processBlocks: %w", err)
		}
		d.log.Infof("processed blocks in %v", time.Since(start))
		debugTimes.Store("blocks", time.Since(start))
		return nil
	})
	g.Go(func() error {
		start := time.Now()
		err := d.processDeposits(data, tar)
		if err != nil {
			return fmt.Errorf("error in processDeposits: %w", err)
		}
		d.log.Infof("processed deposits in %v", time.Since(start))
		debugTimes.Store("deposits", time.Since(start))
		return nil
	})
	// force sequential operation of attestation rewards and proposal rewards
	if data.epochBasedData.epochs[len(data.epochBasedData.epochs)-1] < utils.Config.Chain.ClConfig.AltairForkEpoch {
		d.phase0HotfixMutex.Lock()
	}
	g.Go(func() error {
		start := time.Now()
		err := d.processAttestationRewards(data, tar)
		if err != nil {
			return fmt.Errorf("error in processAttestationRewards: %w", err)
		}
		d.log.Infof("processed attestation rewards in %v", time.Since(start))
		debugTimes.Store("attestationRewards", time.Since(start))
		return nil
	})
	g.Go(func() error {
		start := time.Now()
		err := d.processWithdrawals(data, tar)
		if err != nil {
			return fmt.Errorf("error in processWithdrawals: %w", err)
		}
		d.log.Infof("processed withdrawals in %v", time.Since(start))
		debugTimes.Store("withdrawals", time.Since(start))
		return nil
	})
	g.Go(func() error {
		start := time.Now()
		err := d.processAttestations(data, tar)
		if err != nil {
			return fmt.Errorf("error in processAttestations: %w", err)
		}
		d.log.Infof("processed attestations in %v", time.Since(start))
		debugTimes.Store("attestations", time.Since(start))
		return nil
	})
	// processExpectedSyncPeriods
	g.Go(func() error {
		start := time.Now()
		err := d.processExpectedSyncPeriods(data, tar)
		if err != nil {
			return fmt.Errorf("error in processExpectedSyncPeriods: %w", err)
		}
		d.log.Infof("processed expected sync periods in %v", time.Since(start))
		debugTimes.Store("expectedSyncPeriods", time.Since(start))
		return nil
	})
	// processSyncVotes
	g.Go(func() error {
		start := time.Now()
		err := d.processSyncVotes(data, tar)
		if err != nil {
			return fmt.Errorf("error in processSyncVotes: %w", err)
		}
		d.log.Infof("processed sync votes in %v", time.Since(start))
		debugTimes.Store("syncVotes", time.Since(start))
		return nil
	})
	// processProposalRewards
	g.Go(func() error {
		start := time.Now()
		err := d.processProposalRewards(data, tar)
		if err != nil {
			return fmt.Errorf("error in processProposalRewards: %w", err)
		}
		d.log.Infof("processed proposal rewards in %v", time.Since(start))
		debugTimes.Store("proposalRewards", time.Since(start))
		return nil
	})
	// processSyncCommitteeRewards
	g.Go(func() error {
		start := time.Now()
		err := d.processSyncCommitteeRewards(data, tar)
		if err != nil {
			return fmt.Errorf("error in processSyncCommitteeRewards: %w", err)
		}
		d.log.Infof("processed sync committee rewards in %v", time.Since(start))
		debugTimes.Store("syncCommitteeRewards", time.Since(start))
		return nil
	})
	// processBlocksExpected
	g.Go(func() error {
		start := time.Now()
		err := d.processBlocksExpected(data, tar)
		if err != nil {
			return fmt.Errorf("error in processBlocksExpected: %w", err)
		}
		d.log.Infof("processed blocks expected in %v", time.Since(start))
		debugTimes.Store("blocksExpected", time.Since(start))
		return nil
	})

	d.log.Info("waiting for all processes to finish")
	err = g.Wait()
	if err != nil {
		return fmt.Errorf("error in processRunner: %w", err)
	}
	d.log.Infof("all processes finished in %v (that's %v per epoch)", time.Since(start), time.Since(start)/time.Duration(len(data.epochBasedData.epochs)))
	metrics.TaskDuration.WithLabelValues("exporter_v25dash_process_runner").Observe(time.Since(start).Seconds())
	metrics.TaskDuration.WithLabelValues("exporter_v25dash_process_runner_per_epoch").Observe((time.Since(start) / time.Duration(len(data.epochBasedData.epochs))).Seconds())
	// debug message to discord
	var msg string
	// sort by debug time seen
	sortedKeys := make([]string, 0)
	debugTimes.Range(
		func(key, value any) bool {
			sortedKeys = append(sortedKeys, key.(string))
			return true
		})
	slices.SortFunc(sortedKeys,
		func(a, b string) int {
			av, _ := debugTimes.Load(a)
			bv, _ := debugTimes.Load(b)
			return int((av.(time.Duration) - bv.(time.Duration)).Nanoseconds())
		})
	for _, k := range sortedKeys {
		v, _ := debugTimes.Load(k)
		msg += fmt.Sprintf("%s: %v\n", k,
			v)
		// also expose over metrics
		metrics.TaskDuration.WithLabelValues("exporter_v25dash_processing_" + k).Observe(v.(time.Duration).Seconds())
	}
	d.log.Infof("debug times:\n%s", msg)
	// send message
	utils.SendMessage(fmt.Sprintf("Debug Times processRunner\n```\n%s```", msg), &utils.Config.InternalAlerts)

	return nil
}

func (d *dashboardData) processValidatorStates(data *MultiEpochData, tar *[]db.VDBDataEpochColumns) error {
	g := &errgroup.Group{}
	for i, e := range data.epochBasedData.epochs {
		epoch := e
		iEpoch := int64(epoch)
		tI := data.epochBasedData.tarIndices[i]
		tO := data.epochBasedData.tarOffsets[i]
		startSate := data.epochBasedData.validatorStates[iEpoch-1]
		endState := data.epochBasedData.validatorStates[iEpoch]
		ts := utils.EpochToTime(uint64(epoch))
		g.Go(func() error {
			for j := range endState.Data {
				(*tar)[tI].ValidatorIndex[tO+j] = uint64(j)
				(*tar)[tI].Epoch[tO+j] = iEpoch
				(*tar)[tI].EpochTimestamp[tO+j] = &ts
				(*tar)[tI].BalanceEnd[tO+j] = int64(endState.Data[j].Balance)
				(*tar)[tI].BalanceEffectiveEnd[tO+j] = int64(endState.Data[j].EffectiveBalance)
				// do NOT set the slashed flag here. it is set by the block processing
			}
			if epoch == 0 {
				// we do not set the start state for epoch 0, this is because validators that join
				// after the genesis epoch will always have a start balance of 0 (they didn't exist
				// in the previous epoch yet) and we want to emulate the same behavior
				return nil
			}
			for j := range startSate.Data {
				(*tar)[tI].BalanceStart[tO+j] = int64(startSate.Data[j].Balance)
				(*tar)[tI].BalanceEffectiveStart[tO+j] = int64(startSate.Data[j].EffectiveBalance)
			}
			return nil
		})
	}
	return g.Wait()
}

func (d *dashboardData) processScheduledAttestations(data *MultiEpochData, tar *[]db.VDBDataEpochColumns) error {
	g := &errgroup.Group{}
	for i, e := range data.epochBasedData.epochs {
		epoch := e
		tI := data.epochBasedData.tarIndices[i]
		tO := data.epochBasedData.tarOffsets[i]
		startSlot := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
		endSlot := startSlot + utils.Config.Chain.ClConfig.SlotsPerEpoch
		g.Go(func() error {
			// pre-init should not be required, is pointer and defaults to null
			for slot := startSlot; slot < endSlot; slot++ {
				// fetch from attestati	on assignments
				if _, ok := data.slotBasedData.assignments.attestationAssignments[slot]; !ok {
					// error, should never happen, we should always have assignments
					return fmt.Errorf("no attestation assignments for slot %d", slot)
				}
				for committee, validatorIndices := range data.slotBasedData.assignments.attestationAssignments[slot] {
					for committeeIndex, validatorIndex := range validatorIndices {
						(*tar)[tI].AttestationsScheduled[tO+int(validatorIndex)]++
						(*tar)[tI].AttestationAssignmentsSlot[tO+int(validatorIndex)] = append(
							(*tar)[tI].AttestationAssignmentsSlot[tO+int(validatorIndex)],
							int64(slot),
						)
						(*tar)[tI].AttestationAssignmentsCommittee[tO+int(validatorIndex)] = append(
							(*tar)[tI].AttestationAssignmentsCommittee[tO+int(validatorIndex)],
							int64(committee),
						)
						(*tar)[tI].AttestationAssignmentsIndex[tO+int(validatorIndex)] = append(
							(*tar)[tI].AttestationAssignmentsIndex[tO+int(validatorIndex)],
							int64(committeeIndex),
						)
					}
				}
			}
			return nil
		})
	}
	return g.Wait()
}

func (d *dashboardData) processBlocks(data *MultiEpochData, tar *[]db.VDBDataEpochColumns) error {
	g := &errgroup.Group{}
	for i, e := range data.epochBasedData.epochs {
		epoch := e
		tI := data.epochBasedData.tarIndices[i]
		tO := data.epochBasedData.tarOffsets[i]
		startSlot := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
		endSlot := startSlot + utils.Config.Chain.ClConfig.SlotsPerEpoch
		g.Go(func() error {
			for slot := startSlot; slot < endSlot; slot++ {
				// skip slot 0, as nobody proposed it
				if slot == 0 {
					continue
				}
				proposer := data.slotBasedData.assignments.blockAssignments[slot]
				(*tar)[tI].BlocksStatusSlot[tO+int(proposer)] = append((*tar)[tI].BlocksStatusSlot[tO+int(proposer)], int64(slot))

				if _, ok := data.slotBasedData.blocks[slot]; !ok {
					(*tar)[tI].BlocksStatusProposed[tO+int(proposer)] = append((*tar)[tI].BlocksStatusProposed[tO+int(proposer)], false)
					continue
				}

				(*tar)[tI].BlocksStatusProposed[tO+int(proposer)] = append((*tar)[tI].BlocksStatusProposed[tO+int(proposer)], true)

				for _, index := range data.slotBasedData.blocks[slot].SlashedIndices {
					(*tar)[tI].Slashed[tO+int(index)] = true
					(*tar)[tI].BlocksSlashingCount[uint64(tO)+data.slotBasedData.blocks[slot].ProposerIndex]++
				}
			}
			return nil
		})
	}
	return g.Wait()
}

func (d *dashboardData) processDeposits(data *MultiEpochData, tar *[]db.VDBDataEpochColumns) error {
	if d.signingDomain == nil {
		domain, err := utils.GetSigningDomain()
		if err != nil {
			return err
		}
		d.signingDomain = domain
	}
	g := &errgroup.Group{}
	for i, e := range data.epochBasedData.epochs {
		epoch := e
		tI := data.epochBasedData.tarIndices[i]
		tO := data.epochBasedData.tarOffsets[i]
		// genesis deposits
		if epoch == 0 {
			start := time.Now()
			for j := range data.epochBasedData.validatorStates[0].Data {
				(*tar)[tI].DepositsCount[tO+j] = 1
				(*tar)[tI].DepositsAmount[tO+j] = int64(data.epochBasedData.validatorStates[0].Data[j].Balance)
			}
			d.log.Infof("processed genesis deposits in %v", time.Since(start))
		}
		startSlot := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
		endSlot := startSlot + utils.Config.Chain.ClConfig.SlotsPerEpoch
		g.Go(func() error {
			for jj := startSlot; jj < endSlot; jj++ {
				slot := jj
				if _, ok := data.slotBasedData.blocks[slot]; !ok {
					// nothing to do
					continue
				}

				for depositIndex, depositData := range data.slotBasedData.blocks[slot].Deposits {
					index, indexExists := data.validatorBasedData.validatorIndices[string(depositData.Data.Pubkey)]
					if !indexExists {
						// skip
						d.log.Infof("validator not found for deposit at index %d in slot %v", depositIndex, data.slotBasedData.blocks[slot].Slot)
						continue
					}
					err := utils.VerifyDepositSignature(&phase0.DepositData{
						PublicKey:             phase0.BLSPubKey(depositData.Data.Pubkey),
						WithdrawalCredentials: depositData.Data.WithdrawalCredentials,
						Amount:                phase0.Gwei(depositData.Data.Amount),
						Signature:             phase0.BLSSignature(depositData.Data.Signature),
					}, d.signingDomain)
					if err != nil {
						d.log.Error(fmt.Errorf("deposit at index %d in slot %v is invalid: %v (signature: %s)", depositIndex, data.slotBasedData.blocks[slot].Slot, err, depositData.Data.Signature), "", 0)
						// this fails if there was no valid deposits in our entire epoch range
						// so we can skip this deposit
						if !indexExists {
							d.log.Infof("validator did not have a valid deposit within epoch range - assuming discard of deposit")
							continue
						}
						// validator has been created. check if it already exists in the current epoch. if yes, all deposits can be assumed as valid
						// check if it existed in the previous state
						var existedPreviousEpoch bool
						if uint64(len(data.epochBasedData.validatorStates[int64(epoch)].Data)) > index {
							existedPreviousEpoch = true
						}

						if (*tar)[tI].DepositsCount[uint64(tO)+index] == 0 && !existedPreviousEpoch {
							d.log.Infof("validator did not have a valid deposit before current deposit within epoch and did not exist in previous epoch - assuming discard of deposit")
							continue
						}
					}
					(*tar)[tI].DepositsAmount[uint64(tO)+index] += int64(depositData.Data.Amount)

					(*tar)[tI].DepositsCount[uint64(tO)+index]++
					d.log.Infof("processed deposit at index %d in slot %v", depositIndex, data.slotBasedData.blocks[slot].Slot)

				}
			}
			return nil
		})
	}

	return g.Wait()
}

func (d *dashboardData) processAttestationRewards(data *MultiEpochData, tar *[]db.VDBDataEpochColumns) error {
	if data.epochBasedData.epochs[len(data.epochBasedData.epochs)-1] < utils.Config.Chain.ClConfig.AltairForkEpoch {
		d.phase0HotfixMutex.Lock()
		defer d.phase0HotfixMutex.Unlock()
	}
	g := &errgroup.Group{}
	for i, e := range data.epochBasedData.epochs {
		epoch := e
		tI := data.epochBasedData.tarIndices[i]
		tO := data.epochBasedData.tarOffsets[i]
		if data.epochBasedData.rewards.attestationRewards[epoch] == nil {
			return fmt.Errorf("no rewards for epoch %d", epoch)
		}
		g.Go(func() error {
			// calculate max per committee x attestation slot
			// array of slotsperepoch length, then array of committee per slot length. no maps because maps are slow
			hyperlocalizedMax := make([][]int64, int(utils.Config.Chain.ClConfig.SlotsPerEpoch))
			if utils.Config.Chain.ClConfig.TargetCommitteeSize == 0 {
				utils.Config.Chain.ClConfig.TargetCommitteeSize = 64
			}
			// pre-init committees
			for slot := 0; slot < int(utils.Config.Chain.ClConfig.SlotsPerEpoch); slot++ {
				hyperlocalizedMax[slot] = make([]int64, utils.Config.Chain.ClConfig.TargetCommitteeSize)
			}
			// validator_index => [slot, committee] mapping
			validatorSlotMap := make([]*struct {
				Slot      int64
				Committee int64
			}, len(data.epochBasedData.validatorStates[int64(epoch)].Data))
			startSlot := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
			endSlot := startSlot + utils.Config.Chain.ClConfig.SlotsPerEpoch
			for slot := startSlot; slot < endSlot; slot++ {
				if _, ok := data.slotBasedData.assignments.attestationAssignments[slot]; !ok {
					return fmt.Errorf("no attestation assignments for slot %d", slot)
				}
				for committee, validatorIndices := range data.slotBasedData.assignments.attestationAssignments[slot] {
					for _, validatorIndex := range validatorIndices {
						validatorSlotMap[validatorIndex] = &struct {
							Slot      int64
							Committee int64
						}{
							Slot:      int64(slot - startSlot),
							Committee: int64(committee),
						}
					}
				}
			}

			for _, ar := range data.epochBasedData.rewards.attestationRewards[epoch] {
				// there is a bug in lighthouse which causes the rewards api to return rewards for validators
				// that have been deposited but arent active yet for some reason
				// we can safely ignore these
				if ar.ValidatorIndex >= uint64(len(validatorSlotMap)) {
					//d.log.Infof("skipping reward for validator %d in epoch %d", ar.ValidatorIndex, epoch)
					continue
				}
				valiIndextO := uint64(tO) + ar.ValidatorIndex
				// ideal rewards
				idealReward, ok := data.epochBasedData.rewards.attestationIdealRewards[epoch][data.epochBasedData.validatorStates[int64(epoch)].Data[ar.ValidatorIndex].EffectiveBalance]
				if !ok {
					return fmt.Errorf("no ideal reward for validator %d in epoch %d", valiIndextO, epoch)
				}
				(*tar)[tI].AttestationsIdealHeadReward[valiIndextO] = int64(idealReward.Head)
				(*tar)[tI].AttestationsIdealSourceReward[valiIndextO] = int64(idealReward.Source)
				(*tar)[tI].AttestationsIdealTargetReward[valiIndextO] = int64(idealReward.Target)
				(*tar)[tI].AttestationsIdealInclusionReward[valiIndextO] = int64(idealReward.InclusionDelay)
				(*tar)[tI].AttestationsIdealInactivityReward[valiIndextO] = int64(idealReward.Inactivity)

				(*tar)[tI].AttestationsHeadReward[valiIndextO] = int64(ar.Head)
				(*tar)[tI].AttestationsSourceReward[valiIndextO] = int64(ar.Source)
				(*tar)[tI].AttestationsTargetReward[valiIndextO] = int64(ar.Target)
				// phase0 hotfix - cap inclusion delay reward at ideal inclusion delay reward
				if epoch < utils.Config.Chain.ClConfig.AltairForkEpoch && int64(ar.InclusionDelay) > int64(idealReward.InclusionDelay) {
					ar.InclusionDelay = idealReward.InclusionDelay
				}
				(*tar)[tI].AttestationsInclusionReward[valiIndextO] = int64(ar.InclusionDelay)
				(*tar)[tI].AttestationsInactivityReward[valiIndextO] = int64(ar.Inactivity)
				// total
				total := int64(0)
				for _, r := range []int64{int64(ar.Head), int64(ar.Source), int64(ar.Target), int64(ar.InclusionDelay), int64(ar.Inactivity)} {
					total += r
				}
				// generate hyper localized max
				attData := validatorSlotMap[ar.ValidatorIndex]
				if attData == nil {
					// happens when the validator has been added to state but doesnt have a duty yet or has been slashed
					// fine to ignore
					continue
				}
				if hyperlocalizedMax[attData.Slot][attData.Committee] < total {
					hyperlocalizedMax[attData.Slot][attData.Committee] = total
				}
			}
			// localize to slot
			localizedMax := make([]int64, int(utils.Config.Chain.ClConfig.SlotsPerEpoch))
			for slot := 0; slot < int(utils.Config.Chain.ClConfig.SlotsPerEpoch); slot++ {
				r := int64(0)
				for _, d := range hyperlocalizedMax[slot] {
					if d > r {
						r = d
					}
				}
				localizedMax[slot] = r
			}
			// write hyper localized max by iterating over all validators and looking up the max
			for _, ar := range data.epochBasedData.rewards.attestationRewards[epoch] {
				if ar.ValidatorIndex >= uint64(len(validatorSlotMap)) {
					continue
				}
				valiSlot := validatorSlotMap[ar.ValidatorIndex]

				// happens when the validator has been added to state but doesnt have a duty yet or has been slashed
				// fine to ignore
				if valiSlot == nil {
					continue
				}
				valiIndextO := uint64(tO) + ar.ValidatorIndex
				(*tar)[tI].AttestationsHyperLocalizedMaxReward[valiIndextO] = hyperlocalizedMax[valiSlot.Slot][valiSlot.Committee]
				(*tar)[tI].AttestationsLocalizedMaxReward[valiIndextO] = localizedMax[valiSlot.Slot]
			}
			return nil
		})
	}
	return g.Wait()
}

// withdrawals
func (d *dashboardData) processWithdrawals(data *MultiEpochData, tar *[]db.VDBDataEpochColumns) error {
	g := &errgroup.Group{}
	for i, e := range data.epochBasedData.epochs {
		epoch := e
		if epoch < utils.Config.Chain.ClConfig.CapellaForkEpoch {
			d.log.Infof("skipping withdrawals for epoch %d (before capella)", epoch)
			// no withdrawals before cappella
			continue
		}
		tI := data.epochBasedData.tarIndices[i]
		tO := uint64(data.epochBasedData.tarOffsets[i])
		startSlot := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
		endSlot := startSlot + utils.Config.Chain.ClConfig.SlotsPerEpoch
		g.Go(func() error {
			for j := startSlot; j < endSlot; j++ {
				if _, ok := data.slotBasedData.blocks[j]; !ok {
					// nothing to do
					continue
				}
				for _, withdrawal := range data.slotBasedData.blocks[j].Withdrawals {
					(*tar)[tI].WithdrawalsAmount[tO+withdrawal.ValidatorIndex] = int64(withdrawal.Amount)
					(*tar)[tI].WithdrawalsCount[tO+withdrawal.ValidatorIndex]++
				}
			}
			return nil
		})
	}
	return g.Wait()
}

// Function to filter an array of integers using a bit mask with reduced allocations
func filterArrayUsingBitMask(arr []uint64, bitmask []byte) []uint64 {
	result := make([]uint64, 0, len(arr))

	for i := 0; i < len(arr); i++ {
		byteIndex := i / 8
		bitIndex := i % 8
		if byteIndex < len(bitmask) {
			if (bitmask[byteIndex] & (1 << bitIndex)) != 0 {
				result = append(result, arr[i])
			}
		}
	}

	return result
}

func IntegerSquareRoot(n uint64) uint64 {
	x := n
	y := (x + 1) / 2
	for y < x {
		x = y
		y = (x + n/x) / 2
	}
	return x
}

// attestations
func (d *dashboardData) processAttestations(data *MultiEpochData, tar *[]db.VDBDataEpochColumns) error {
	g := &errgroup.Group{}
	blockRoots := make(map[uint64]hexutil.Bytes)
	blockValidityMap := make(map[uint64]int64, len(data.slotBasedData.blocks))
	// loop through slots. if a slot is missing reuse the previous slot. skip if the first slot is missing just skip
	var lastBlockHash *hexutil.Bytes
	var previousValidHash hexutil.Bytes
	var toBeFilled []uint64
	for _, e := range data.epochBasedData.epochs {
		epoch := e
		start := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
		end := start + utils.Config.Chain.ClConfig.SlotsPerEpoch
		for j := start; j < end; j++ {
			if _, ok := data.slotBasedData.blocks[j]; !ok {
				blockValidityMap[j] = 0
				// check if we have a previous block
				if lastBlockHash == nil {
					toBeFilled = append(toBeFilled, j)
					continue
				}
				blockRoots[j] = *lastBlockHash
				continue
			}
			blockValidityMap[j] = 1
			if lastBlockHash == nil {
				previousValidHash = data.slotBasedData.blocks[j].ParentRoot
			}
			a := data.slotBasedData.blocks[j].BlockRoot
			lastBlockHash = &a
			blockRoots[j] = *lastBlockHash
		}
	}
	// do blockValidityMap for n+1 epoch data
	lookaheadEpoch := data.epochBasedData.epochs[len(data.epochBasedData.epochs)-1] + 1
	start := lookaheadEpoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
	end := start + utils.Config.Chain.ClConfig.SlotsPerEpoch
	for j := start; j < end; j++ {
		if _, ok := data.slotBasedData.blocks[j]; !ok {
			blockValidityMap[j] = 0
			continue
		}
		blockValidityMap[j] = 1
	}
	if lastBlockHash == nil {
		return fmt.Errorf("no valid slots found")
	}
	// fill toBeFilled
	for _, slot := range toBeFilled {
		blockRoots[slot] = previousValidHash
	}

	for i, e := range data.epochBasedData.epochs {
		epoch := e
		//debugCounters := make(map[string]int)
		tI := data.epochBasedData.tarIndices[i]
		tO := uint64(data.epochBasedData.tarOffsets[i])
		startSlot := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
		endSlot := startSlot + utils.Config.Chain.ClConfig.SlotsPerEpoch
		squareSlotsPerEpoch := IntegerSquareRoot(utils.Config.Chain.ClConfig.SlotsPerEpoch)
		//debugCounters := make(map[string]int)
		g.Go(func() error {
			start := time.Now()
			for j := startSlot; j < startSlot+(utils.Config.Chain.ClConfig.SlotsPerEpoch*2); j++ {
				if _, ok := data.slotBasedData.blocks[j]; !ok {
					// nothing to do
					continue
				}
				// d.log.Infof("processing attestations for epoch %d in slot %d", epoch, j)
				// attestations
				for _, att := range data.slotBasedData.blocks[j].Attestations {
					// ignore if slot is not within our epoch of interest
					if att.Data.Slot < startSlot || att.Data.Slot >= endSlot {
						//d.log.Infof("ignoring attestation in slot %d because its for a different epoch %d while we are processing epoch %d", att.Data.Slot, att.Data.Slot/utils.Config.Chain.ClConfig.SlotsPerEpoch, epoch)
						//debugCounters["skipped_different_epoch"]++
						continue
					} else {
						//debugCounters["processed_attestations"]++
						//d.log.Infof("processing attestation in slot %d during epoch %d", att.Data.Slot, epoch)
					}
					// precalculate integer squareroot of slots per epoch
					v := filterArrayUsingBitMask(data.slotBasedData.assignments.attestationAssignments[att.Data.Slot][att.Data.Index], att.AggregationBits)
					if len(v) == 0 {
						//debugCounters["skipped_no_validators"]++
						continue
					}
					inclusion_delay := int64(j - att.Data.Slot)
					if inclusion_delay < 1 {
						return fmt.Errorf("inclusion delay is less than 1 in slot %d for attestation of slot %d", j, att.Data.Slot)
					}
					// https://eips.ethereum.org/EIPS/eip-7045
					if epoch < utils.Config.Chain.ClConfig.AltairForkEpoch && inclusion_delay > 32 {
						//debugCounters["skipped_inclusion_delay"]++
						continue
					}
					optimalInclusionDelay := int64(0)
					for k := att.Data.Slot + 1; k < j; k++ {
						c, ok := blockValidityMap[k]
						if !ok {
							return fmt.Errorf("no block validity found for slot %d", k)
						}
						optimalInclusionDelay += c
					}
					is_matching_source := true // enforced by the spec
					br, ok := blockRoots[att.Data.Target.Epoch*utils.Config.Chain.ClConfig.SlotsPerEpoch]
					if !ok {
						return fmt.Errorf("no block root found for epoch %d", att.Data.Target.Epoch)
					}
					is_matching_target := bytes.Equal(att.Data.Target.Root, br)
					br, ok = blockRoots[att.Data.Slot]
					if !ok {
						return fmt.Errorf("no block root found for slot %d", att.Data.Slot)
					}
					is_matching_head := is_matching_target && bytes.Equal(att.Data.BeaconBlockRoot, br)
					sourceValue := int64(0)
					targetValue := int64(0)
					headValue := int64(0)
					switch utils.ForkVersionAtEpoch(epoch).CurrentVersion {
					case utils.Config.Chain.ClConfig.GenesisForkVersion:
						// genesis, head source and target only care about having accurate hashes
						if is_matching_source {
							sourceValue = 1
						}
						if is_matching_target {
							targetValue = 1
						}
						if is_matching_head {
							headValue = 1
						}
					default:
						if is_matching_source && inclusion_delay <= int64(squareSlotsPerEpoch) {
							sourceValue = 1
						}
						if is_matching_target { // 32 slot filter on target for pre altair forks is done above
							targetValue = 1
						}
						if is_matching_head && inclusion_delay == 1 {
							headValue = 1
						}
					}

					for _, valiIndex := range v {
						valiIndex += tO
						// check if it was already processed. if yes skip
						if (*tar)[tI].AttestationsObserved[valiIndex] > 0 {
							continue
						}
						// executed
						(*tar)[tI].AttestationsObserved[valiIndex]++
						// inclusion delay
						(*tar)[tI].InclusionDelaySum[valiIndex] += inclusion_delay - 1
						(*tar)[tI].OptimalInclusionDelaySum[valiIndex] += optimalInclusionDelay
						// metrics
						(*tar)[tI].AttestationsSourceExecuted[valiIndex] += sourceValue
						if is_matching_source {
							(*tar)[tI].AttestationsSourceMatched[valiIndex]++
						}
						(*tar)[tI].AttestationsTargetExecuted[valiIndex] += targetValue
						if is_matching_target {
							(*tar)[tI].AttestationsTargetMatched[valiIndex]++
						}
						(*tar)[tI].AttestationsHeadExecuted[valiIndex] += headValue
						if is_matching_head {
							(*tar)[tI].AttestationsHeadMatched[valiIndex]++
						}
					}
				}
			}
			d.log.Infof("processed attestations in epoch %d in %v", epoch, time.Since(start))
			/*
				fmt.Println("debug counters:")
				for k, v := range debugCounters {
					fmt.Printf("%s: %d (%.2f%%)\n", k, v, float64(v)/float64(debugCounters["executed"])*100)
				}
			*/
			return nil
		})
	}
	return g.Wait()
}

// sync odds
func (d *dashboardData) processExpectedSyncPeriods(data *MultiEpochData, tar *[]db.VDBDataEpochColumns) error {
	g := &errgroup.Group{}
	for i, e := range data.epochBasedData.epochs {
		epoch := e
		if utils.FirstEpochOfSyncPeriod(utils.SyncPeriodOfEpoch(epoch)) != epoch {
			// d.log.Infof("skipping epoch %d as it is not the first epoch of a sync period", epoch)
			continue
		}
		tI := data.epochBasedData.tarIndices[i]
		tO := uint64(data.epochBasedData.tarOffsets[i])
		g.Go(func() error {
			iEpoch := int64(epoch)

			totalEffective := int64(0)
			for _, valData := range data.epochBasedData.validatorStates[iEpoch].Data {
				if valData.Status.IsActive() {
					totalEffective += int64(valData.EffectiveBalance / 1e9)
				}
			}
			fTotalEffective := float64(totalEffective)
			fCommitteeSize := float64(utils.Config.Chain.ClConfig.SyncCommitteeSize)

			for _, valData := range data.epochBasedData.validatorStates[iEpoch].Data {
				if valData.Status.IsActive() {
					// See https://github.com/ethereum/annotated-spec/blob/master/altair/beacon-chain.md#get_sync_committee_indices
					// Note that this formula is not 100% the chance as defined in the spec, but after running simulations we found
					// it being precise enough for our purposes with an error margin of less than 0.003%
					syncChance := float64(valData.EffectiveBalance/1e9) / fTotalEffective
					(*tar)[tI].SyncCommitteesExpected[tO+valData.Index] = syncChance * fCommitteeSize
				}
			}

			return nil
		})
	}
	return g.Wait()
}

// blocks expected
func (d *dashboardData) processBlocksExpected(data *MultiEpochData, tar *[]db.VDBDataEpochColumns) error {
	g := &errgroup.Group{}
	for i, e := range data.epochBasedData.epochs {
		epoch := e
		tI := data.epochBasedData.tarIndices[i]
		tO := uint64(data.epochBasedData.tarOffsets[i])
		g.Go(func() error {
			iEpoch := int64(epoch)
			totalEffective := int64(0)
			for _, valData := range data.epochBasedData.validatorStates[iEpoch].Data {
				if valData.Status.IsActive() {
					totalEffective += int64(valData.EffectiveBalance / 1e9)
				}
			}
			fTotalEffective := float64(totalEffective)
			fSlotsPerEpoch := float64(utils.Config.Chain.ClConfig.SlotsPerEpoch)
			if epoch == 0 {
				// cant propose slot 0
				fSlotsPerEpoch--
			}
			for _, valData := range data.epochBasedData.validatorStates[iEpoch].Data {
				if valData.Status.IsActive() {
					// See https://github.com/ethereum/annotated-spec/blob/master/phase0/beacon-chain.md#compute_proposer_index
					proposalChance := float64(valData.EffectiveBalance/1e9) / fTotalEffective
					(*tar)[tI].BlocksExpected[tO+valData.Index] = proposalChance * fSlotsPerEpoch
				}
			}
			return nil

		})
	}
	return g.Wait()
}

func (d *dashboardData) processSyncVotes(data *MultiEpochData, tar *[]db.VDBDataEpochColumns) error {
	g := &errgroup.Group{}
	for i, e := range data.epochBasedData.epochs {
		epoch := e
		if epoch < utils.Config.Chain.ClConfig.AltairForkEpoch {
			d.log.Infof("skipping sync votes for epoch %d (before altair)", epoch)
			// no sync votes before altair
			continue
		}
		syncPeriod := utils.SyncPeriodOfEpoch(epoch)
		iSyncPeriod := int64(syncPeriod)
		tI := data.epochBasedData.tarIndices[i]
		tO := uint64(data.epochBasedData.tarOffsets[i])
		startSlot := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
		endSlot := startSlot + utils.Config.Chain.ClConfig.SlotsPerEpoch
		g.Go(func() error {
			// safety check, check if we have the sync period data
			if _, ok := data.syncPeriodBasedData.SyncAssignments[syncPeriod]; !ok {
				return fmt.Errorf("no sync assignments for sync period %d", syncPeriod)
			}
			for i, valiIndex := range data.syncPeriodBasedData.SyncAssignments[syncPeriod] {
				(*tar)[tI].SyncCommitteeAssignmentsPeriod[tO+valiIndex] = append((*tar)[tI].SyncCommitteeAssignmentsPeriod[tO+valiIndex], iSyncPeriod)
				(*tar)[tI].SyncCommitteeAssignmentsIndex[tO+valiIndex] = append((*tar)[tI].SyncCommitteeAssignmentsIndex[tO+valiIndex], int64(i))
			}
			for j := startSlot; j < endSlot; j++ {
				for _, valiIndex := range data.syncPeriodBasedData.SyncAssignments[syncPeriod] {
					(*tar)[tI].SyncStatusSlot[tO+valiIndex] = append((*tar)[tI].SyncStatusSlot[tO+valiIndex], int64(j))
					(*tar)[tI].SyncStatusExecuted[tO+valiIndex] = append((*tar)[tI].SyncStatusExecuted[tO+valiIndex], false)
				}
				if _, ok := data.slotBasedData.blocks[j]; !ok {
					// nothing to do
					continue
				}
				// you cant vote for slot 0
				if j == 0 {
					continue
				}
				// sync votes
				v := filterArrayUsingBitMask(
					data.syncPeriodBasedData.SyncAssignments[syncPeriod],
					data.slotBasedData.blocks[j].SyncAggregate.SyncCommitteeBits)
				for _, valiIndex := range v {
					(*tar)[tI].SyncStatusExecuted[tO+valiIndex][len((*tar)[tI].SyncStatusExecuted[tO+valiIndex])-1] = true
				}
				for _, valiIndex := range data.syncPeriodBasedData.SyncAssignments[syncPeriod] {
					(*tar)[tI].SyncScheduled[tO+valiIndex]++
				}
			}
			return nil
		})
	}
	return g.Wait()
}

func (d *dashboardData) processProposalRewards(data *MultiEpochData, tar *[]db.VDBDataEpochColumns) error {
	if data.epochBasedData.epochs[len(data.epochBasedData.epochs)-1] < utils.Config.Chain.ClConfig.AltairForkEpoch {
		defer d.phase0HotfixMutex.Unlock()
	}
	g := &errgroup.Group{}
	buffer := utils.Config.Chain.ClConfig.SlotsPerEpoch / 2
	for i, e := range data.epochBasedData.epochs {
		epoch := e
		tI := data.epochBasedData.tarIndices[i]
		tO := uint64(data.epochBasedData.tarOffsets[i])
		startSlot := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
		endSlot := startSlot + utils.Config.Chain.ClConfig.SlotsPerEpoch
		g.Go(func() error {
			for j := startSlot; j < endSlot; j++ {
				// calculate median reward
				medianStartSlot := uint64(0)
				if j >= buffer {
					medianStartSlot = j - buffer
				}
				medianEndSlot := j + buffer
				// actually, lets do median instead
				medianArray := make([]int64, 0)
				for k := medianStartSlot; k < medianEndSlot; k++ {
					if r, ok := data.slotBasedData.rewards.blockRewards[k]; ok {
						rewards := r.Data
						reward := rewards.Attestations + rewards.AttesterSlashings + rewards.ProposerSlashings + rewards.SyncAggregate
						medianArray = append(medianArray, reward)
					}
				}
				if len(medianArray) == 0 {
					// no rewards in buffer
					// lets just bork ourselves to be safe
					return fmt.Errorf("no proposal rewards in buffer for slot %d", j)
				}
				// calculate median
				slices.Sort(medianArray)
				median := int64(0)
				if len(medianArray)%2 == 0 {
					// even
					median = (medianArray[len(medianArray)/2-1] + medianArray[len(medianArray)/2]) / 2
				} else {
					// odd
					median = medianArray[len(medianArray)/2]
				}

				if _, ok := data.slotBasedData.rewards.blockRewards[j]; !ok {
					// assign to validator that missed the slot
					proposer := data.slotBasedData.assignments.blockAssignments[j]
					(*tar)[tI].BlocksClMissedMedianReward[tO+proposer] += median
					continue
				}
				rewards := data.slotBasedData.rewards.blockRewards[j].Data
				(*tar)[tI].BlockRewardsSlot[tO+rewards.ProposerIndex] = append((*tar)[tI].BlockRewardsSlot[tO+rewards.ProposerIndex], int64(j))
				(*tar)[tI].BlockRewardsAttestationsReward[tO+rewards.ProposerIndex] = append((*tar)[tI].BlockRewardsAttestationsReward[tO+rewards.ProposerIndex], int64(rewards.Attestations))
				(*tar)[tI].BlockRewardsSyncAggregateReward[tO+rewards.ProposerIndex] = append((*tar)[tI].BlockRewardsSyncAggregateReward[tO+rewards.ProposerIndex], int64(rewards.SyncAggregate))
				(*tar)[tI].BlockRewardsSlasherReward[tO+rewards.ProposerIndex] = append((*tar)[tI].BlockRewardsSlasherReward[tO+rewards.ProposerIndex], int64(rewards.AttesterSlashings+rewards.ProposerSlashings))
				reward := rewards.Attestations + rewards.AttesterSlashings + rewards.ProposerSlashings + rewards.SyncAggregate
				if reward < median {
					(*tar)[tI].BlocksClMissedMedianReward[tO+rewards.ProposerIndex] += median - reward
				}
				if epoch < utils.Config.Chain.ClConfig.AltairForkEpoch {
					// hotfix for phase0 blocks
					if (*data).epochBasedData.rewards.attestationRewards[epoch][rewards.ProposerIndex].InclusionDelay > int32(reward) {
						(*data).epochBasedData.rewards.attestationRewards[epoch][rewards.ProposerIndex].InclusionDelay -= int32(reward)
					}
				}
			}
			return nil
		})
	}
	return g.Wait()
}

// sync rewards
func (d *dashboardData) processSyncCommitteeRewards(data *MultiEpochData, tar *[]db.VDBDataEpochColumns) error {
	g := &errgroup.Group{}
	for i, e := range data.epochBasedData.epochs {
		epoch := e
		if epoch < utils.Config.Chain.ClConfig.AltairForkEpoch {
			d.log.Infof("skipping sync rewards for epoch %d (before altair)", epoch)
			// no sync rewards before altair
			continue
		}
		tI := data.epochBasedData.tarIndices[i]
		tO := uint64(data.epochBasedData.tarOffsets[i])
		startSlot := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
		endSlot := startSlot + utils.Config.Chain.ClConfig.SlotsPerEpoch
		g.Go(func() error {
			maxRewards := int64(0)
			for j := startSlot; j < endSlot; j++ {
				if _, ok := data.slotBasedData.blocks[j]; !ok {
					// no data is expected for missed blocks
					continue
				}
				for _, reward := range data.slotBasedData.rewards.syncCommitteeRewards[j].Data {
					(*tar)[tI].SyncRewardsSlot[tO+reward.ValidatorIndex] = append((*tar)[tI].SyncRewardsSlot[tO+reward.ValidatorIndex], int64(j))
					(*tar)[tI].SyncRewardsReward[tO+reward.ValidatorIndex] = append((*tar)[tI].SyncRewardsReward[tO+reward.ValidatorIndex], int64(reward.Reward))
					if reward.Reward > maxRewards {
						maxRewards = reward.Reward
					}
				}
			}
			if maxRewards <= 0 {
				// lets just bork ourselves to be safe
				return fmt.Errorf("no rewards in epoch %d", epoch)
			}
			for j := startSlot; j < endSlot; j++ {
				for _, reward := range data.slotBasedData.rewards.syncCommitteeRewards[j].Data {
					(*tar)[tI].SyncLocalizedMaxReward[tO+reward.ValidatorIndex] += maxRewards
				}
			}
			return nil
		})
	}
	return g.Wait()
}

func (d *dashboardData) getDataForEpochRange(epochStart, epochEnd uint64, tar *MultiEpochData) error {
	// data <=|
	//		  |<= (epochBasedData) <=|
	//		  |				 		 |<= sorted(states) <=|
	//		  |<= (syncBasedData)  <=… 					  |<= state(n)
	//		  | 				 	  					  |…
	//		  |<= (slotBasedData)  <=|
	//		  						 |<= sorted(blocks) <=|
	//		  						 |
	g1 := &errgroup.Group{}
	// g1.SetLimit(epochFetchParallelism)
	// prefill epochBasedData.epochs
	for i := epochStart; i <= epochEnd; i++ {
		tar.epochBasedData.epochs = append(tar.epochBasedData.epochs, i)
	}
	heavyRequestsSemMap := sync.Map{}
	weights := make(map[string]int64)
	weights["heavy"] = 8   // 8 parallel requests
	weights["medium"] = 18 // 12 parallel requests
	weights["light"] = 128 // 32 parallel requests
	orderedKeyList := []string{"heavy", "medium", "light"}
	for k, v := range weights {
		a := semaphore.NewWeighted(v)
		heavyRequestsSemMap.Store(k, a)
	}
	// debug timer that prints the size of the queue for each node every 10 seconds
	// should be stopped once function is done
	timer := time.NewTicker(3 * time.Second)
	defer timer.Stop()
	go func() {
		for {
			_, ok := <-timer.C
			if !ok {
				return
			}
			for _, k := range orderedKeyList {
				heavyRequestsSem, _ := heavyRequestsSemMap.Load(k)
				// read cur, size, len(waiters) using reflection
				v := reflect.ValueOf(heavyRequestsSem)
				cur := v.Elem().FieldByName("cur").Int()
				size := v.Elem().FieldByName("size").Int()
				// waiters is a struct that has a len field
				waiters := v.Elem().FieldByName("waiters").FieldByName("len").Int()
				d.log.Infof("%s: cur: %d, size: %d, waiters: %d", k, cur, size, waiters)
			}
		}
	}()

	debugTimes := sync.Map{}
	// epoch based Data
	g1.Go(func() error {
		start := time.Now()
		// get states
		g2 := &errgroup.Group{}
		slots := make([]uint64, 0)
		// first slot of the first epoch
		firstEpochToFetch := epochStart
		/*
			if firstEpochToFetch > 0 {
				firstEpochToFetch--
			}
		*/
		for i := firstEpochToFetch; i <= epochEnd+1; i++ {
			if i == 0 {
				slots = append(slots, 0)
				continue
			}
			slots = append(slots, uint64(i)*utils.Config.Chain.ClConfig.SlotsPerEpoch-1)
		}
		writeMutex := &sync.Mutex{}
		d.log.Infof("fetching states for epochs %d to %d using slots %v", epochStart, epochEnd, slots)
		tar.epochBasedData.validatorStates = make(map[int64]constypes.LightStandardValidatorsResponse, len(slots))
		startEpoch := int64(epochStart) - 1
		for i, s := range slots {
			slot := uint64(s)
			virtualEpoch := startEpoch + int64(i)
			g2.Go(func() error {
				// aquiring semaphore
				nodeId := getNodeId(virtualEpoch)
				heavyRequestsSem, _ := heavyRequestsSemMap.Load("heavy")
				err := heavyRequestsSem.(*semaphore.Weighted).Acquire(context.Background(), 1)
				if err != nil {
					return err
				}
				defer heavyRequestsSem.(*semaphore.Weighted).Release(1)
				d.log.Infof("fetching validator state at slot %d", slot)
				start := time.Now()
				var valis *constypes.StandardValidatorsResponse
				if slot == 0 {
					valis, err = d.CL[nodeId].GetValidators("genesis", nil, nil)
				} else {
					valis, err = d.CL[nodeId].GetValidators(slot, nil, nil)
				}
				d.log.Infof("retrieved validator state at slot %d in %v", slot, time.Since(start))
				if err != nil {
					d.log.Error(err, "can not get validators state", 0, map[string]interface{}{"slot": slot})
					return err
				}
				// convert to light validators
				var lightValis constypes.LightStandardValidatorsResponse
				lightValis.Data = make([]constypes.LightStandardValidator, len(valis.Data))
				for i, val := range valis.Data {
					lightValis.Data[i] = constypes.LightStandardValidator{
						Index:            val.Index,
						Balance:          val.Balance,
						Status:           val.Status,
						Pubkey:           val.Validator.Pubkey,
						EffectiveBalance: val.Validator.EffectiveBalance,
						Slashed:          val.Validator.Slashed,
					}
				}
				writeMutex.Lock()
				tar.epochBasedData.validatorStates[virtualEpoch] = lightValis
				// quick update validatorBasedData.validatorIndices
				for _, val := range lightValis.Data {
					tar.validatorBasedData.validatorIndices[string(val.Pubkey)] = val.Index
				}
				writeMutex.Unlock()
				// free up memory
				valis = nil
				d.log.Infof("fetched validator state at slot %d in %v", slot, time.Since(start))
				return nil
			})
		}
		err := g2.Wait()
		if err != nil {
			return fmt.Errorf("error in epochBasedData: %w", err)
		}
		d.log.Infof("fetched states for epochs %d to %d in %v", epochStart, epochEnd, time.Since(start))
		// add to debug
		debugTimes.Store("epochBasedData", time.Since(start))
		return nil
	})
	// syncPeriodBasedData
	g1.Go(func() error {
		start := time.Now()
		// get sync committee assignments
		g2 := &errgroup.Group{}
		syncPeriodAssignmentsToFetch := make([]uint64, 0)
		snycPeriodStatesToFetch := make([]uint64, 0)
		for i := epochStart; i <= epochEnd; i++ {
			if i < utils.Config.Chain.ClConfig.AltairForkEpoch {
				d.log.Infof("skipping sync committee assignments for epoch %d (before altair)", i)
				// no sync committee assignments before altair
				continue
			}
			syncPeriod := utils.SyncPeriodOfEpoch(i)
			// if we dont have the assignment yet fetch it
			if len(syncPeriodAssignmentsToFetch) == 0 || syncPeriodAssignmentsToFetch[len(syncPeriodAssignmentsToFetch)-1] != syncPeriod {
				syncPeriodAssignmentsToFetch = append(syncPeriodAssignmentsToFetch, syncPeriod)
			}
			if utils.FirstEpochOfSyncPeriod(syncPeriod) == i {
				snycPeriodStatesToFetch = append(snycPeriodStatesToFetch, syncPeriod)
			}
		}
		d.log.Infof("fetching sync committee assignments and states for sync periods %v", syncPeriodAssignmentsToFetch)
		writeMutex := &sync.Mutex{}
		tar.syncPeriodBasedData.SyncAssignments = make(map[uint64][]uint64, len(syncPeriodAssignmentsToFetch))
		tar.syncPeriodBasedData.SyncStateEffectiveBalances = make(map[uint64][]uint64, len(snycPeriodStatesToFetch))
		// assignments
		for _, s := range syncPeriodAssignmentsToFetch {
			syncPeriod := s
			g2.Go(func() error {
				// aquiring semaphore
				nodeId := getNodeId(syncPeriod)
				heavyRequestsSem, _ := heavyRequestsSemMap.Load("medium")
				err := heavyRequestsSem.(*semaphore.Weighted).Acquire(context.Background(), 1)
				if err != nil {
					return err
				}
				defer heavyRequestsSem.(*semaphore.Weighted).Release(1)
				start := time.Now()
				relevantSlot := utils.FirstEpochOfSyncPeriod(syncPeriod) * utils.Config.Chain.ClConfig.SlotsPerEpoch
				assignments, err := d.CL[nodeId].GetSyncCommitteesAssignments(nil, relevantSlot)
				if err != nil {
					d.log.Error(err, "can not get sync committee assignments", 0, map[string]interface{}{"syncPeriod": syncPeriod})
					return err
				}
				writeMutex.Lock()
				tar.syncPeriodBasedData.SyncAssignments[syncPeriod] = make([]uint64, len(assignments.Data.Validators))
				for i, a := range assignments.Data.Validators {
					tar.syncPeriodBasedData.SyncAssignments[syncPeriod][i] = uint64(a)
				}
				writeMutex.Unlock()
				d.log.Infof("fetched sync committee assignments for sync period %d in %v", syncPeriod, time.Since(start))
				return nil
			})
		}
		// states
		for _, s := range snycPeriodStatesToFetch {
			syncPeriod := s
			g2.Go(func() error {
				slot := utils.FirstEpochOfSyncPeriod(syncPeriod) * utils.Config.Chain.ClConfig.SlotsPerEpoch
				// aquiring semaphore
				nodeId := getNodeId(slot / utils.Config.Chain.ClConfig.SlotsPerEpoch)
				heavyRequestsSem, _ := heavyRequestsSemMap.Load("heavy")
				err := heavyRequestsSem.(*semaphore.Weighted).Acquire(context.Background(), 1)
				if err != nil {
					return err
				}
				defer heavyRequestsSem.(*semaphore.Weighted).Release(1)
				start := time.Now()
				valis, err := d.CL[nodeId].GetValidators(slot, nil, nil)
				if err != nil {
					d.log.Error(err, "can not get sync committee state", 0, map[string]interface{}{"syncPeriod": syncPeriod})
					return err
				}
				// convert to light validators
				dat := make([]uint64, len(valis.Data))
				for i, val := range valis.Data {
					if val.Status.IsActive() {
						dat[i] = val.Validator.EffectiveBalance
					}
				}
				writeMutex.Lock()
				tar.syncPeriodBasedData.SyncStateEffectiveBalances[syncPeriod] = dat
				writeMutex.Unlock()
				d.log.Infof("fetched sync committee state for sync period %d in %v", syncPeriod, time.Since(start))
				return nil
			})
		}
		err := g2.Wait()
		if err != nil {
			return fmt.Errorf("error in syncPeriodBasedData: %w", err)
		}
		d.log.Infof("fetched sync committee assignments and states for sync periods %v in %v", syncPeriodAssignmentsToFetch, time.Since(start))
		// add to debug
		debugTimes.Store("syncPeriodBasedData", time.Since(start))
		return nil
	})

	// blocks
	g1.Go(func() error {
		start := time.Now()
		// get blocks
		g2 := &errgroup.Group{}
		slots := make([]uint64, 0)
		// first slot of the previous epoch
		firstSlotToFetch := (epochStart) * utils.Config.Chain.ClConfig.SlotsPerEpoch
		if epochStart == 0 {
			firstSlotToFetch = 0
		}
		lastSlotToFetch := ((epochEnd + 2) * utils.Config.Chain.ClConfig.SlotsPerEpoch) - 1
		for i := firstSlotToFetch; i <= lastSlotToFetch; i++ {
			slots = append(slots, uint64(i))
		}
		writeMutex := &sync.Mutex{}
		d.log.Infof("fetching blocks for slots %d to %d", firstSlotToFetch, lastSlotToFetch)
		tar.slotBasedData.blocks = make(map[uint64]constypes.LightAnySignedBlock, len(slots))
		for _, s := range slots {
			slot := uint64(s)
			epoch := slot / utils.Config.Chain.ClConfig.SlotsPerEpoch
			g2.Go(func() error {
				// d.log.Infof("fetching block at slot %d", slot)
				//start := time.Now()
				heavyRequestsSem, _ := heavyRequestsSemMap.Load("light")
				err := heavyRequestsSem.(*semaphore.Weighted).Acquire(context.Background(), 1)
				if err != nil {
					return err
				}
				defer heavyRequestsSem.(*semaphore.Weighted).Release(1)

				block, err := d.CL[getNodeId(slot)].GetSlot(slot)
				if err != nil {
					httpErr := network.SpecificError(err)
					if httpErr != nil && httpErr.StatusCode == 404 {
						//d.log.Infof("no block at slot %d", slot)
						return nil
					}
					d.log.Error(err, "can not get block", 0, map[string]interface{}{"slot": slot})
					return err
				}
				// header
				header, err := d.CL[getNodeId(slot)].GetBlockHeader(slot)
				if err != nil {
					d.log.Error(err, "can not get block header", 0, map[string]interface{}{"slot": slot})
					return err
				}
				var lightBlock constypes.LightAnySignedBlock
				lightBlock.Slot = block.Data.Message.Slot
				lightBlock.BlockRoot = header.Data.Root
				lightBlock.ParentRoot = header.Data.Header.Message.ParentRoot
				lightBlock.ProposerIndex = block.Data.Message.ProposerIndex
				lightBlock.Attestations = block.Data.Message.Body.Attestations
				// deposits
				lightBlock.Deposits = append(lightBlock.Deposits, block.Data.Message.Body.Deposits...)
				// withdrawals
				if epoch >= utils.Config.Chain.ClConfig.CapellaForkEpoch {
					for _, w := range block.Data.Message.Body.ExecutionPayload.Withdrawals {
						lightBlock.Withdrawals = append(lightBlock.Withdrawals, constypes.LightWithdrawal{
							Amount:         w.Amount,
							ValidatorIndex: w.ValidatorIndex,
						})
					}
				}
				// AttesterSlashings
				for _, s := range block.Data.Message.Body.AttesterSlashings {
					lightBlock.SlashedIndices = append(lightBlock.SlashedIndices, s.GetSlashedIndices()...)
				}
				// ProposerSlashings
				for _, s := range block.Data.Message.Body.ProposerSlashings {
					lightBlock.SlashedIndices = append(lightBlock.SlashedIndices, s.SignedHeader1.Message.ProposerIndex)
				}
				if epoch >= utils.Config.Chain.ClConfig.AltairForkEpoch {
					// sync
					lightBlock.SyncAggregate = block.Data.Message.Body.SyncAggregate
				}
				// free up memory
				block = nil

				writeMutex.Lock()
				tar.slotBasedData.blocks[slot] = lightBlock
				writeMutex.Unlock()
				//d.log.Infof("fetched block at slot %d in %v", slot, time.Since(start))
				return nil
			})
		}
		err := g2.Wait()
		if err != nil {
			return fmt.Errorf("error in slotBasedData: %w", err)
		}
		// add to debug
		debugTimes.Store("slotBasedData", time.Since(start))
		return nil
	})
	// block rewards
	g1.Go(func() error {
		start := time.Now()
		// get block rewards
		g2 := &errgroup.Group{}
		writeMutex := &sync.Mutex{}
		// we will fetch more than the requested epoch range because there is an "median cl reward" column for missed proposals
		// slots:
		buffer := utils.Config.Chain.ClConfig.SlotsPerEpoch / 2
		firstSlotToFetch := (epochStart) * utils.Config.Chain.ClConfig.SlotsPerEpoch
		if firstSlotToFetch >= buffer {
			firstSlotToFetch -= buffer
		}
		lastSlotToFetch := ((epochEnd + 1) * utils.Config.Chain.ClConfig.SlotsPerEpoch) + buffer - 1
		for i := firstSlotToFetch; i <= lastSlotToFetch; i++ {
			slot := i
			g2.Go(func() error {
				// aquiring semaphore
				nodeId := getNodeId(slot)
				heavyRequestsSem, _ := heavyRequestsSemMap.Load("light")
				err := heavyRequestsSem.(*semaphore.Weighted).Acquire(context.Background(), 1)
				if err != nil {
					return err
				}
				defer heavyRequestsSem.(*semaphore.Weighted).Release(1)
				//start := time.Now()
				data, err := d.CL[nodeId].GetPropoalRewards(slot)
				if err != nil {
					httpErr := network.SpecificError(err)
					if httpErr != nil && httpErr.StatusCode == 404 {
						d.log.Infof("no block rewards for slot %d", slot)
						return nil
					}
					d.log.Error(err, "can not get block rewards", 0, map[string]interface{}{"slot": slot})
					return err
				}
				writeMutex.Lock()
				tar.slotBasedData.rewards.blockRewards[slot] = *data
				writeMutex.Unlock()
				// d.log.Infof("fetched block rewards for slot %d in %v", slot, time.Since(start))
				return nil
			})
		}
		err := g2.Wait()
		if err != nil {
			return fmt.Errorf("error in block rewards: %w", err)
		}
		// add to debug
		debugTimes.Store("blockRewards", time.Since(start))
		return nil
	})
	// GetSyncRewards
	g1.Go(func() error {
		start := time.Now()
		// get sync rewards
		g2 := &errgroup.Group{}
		writeMutex := &sync.Mutex{}
		firstSlotToFetch := (epochStart) * utils.Config.Chain.ClConfig.SlotsPerEpoch
		lastSlotToFetch := ((epochEnd + 1) * utils.Config.Chain.ClConfig.SlotsPerEpoch) - 1
		for i := firstSlotToFetch; i <= lastSlotToFetch; i++ {
			slot := i
			epoch := slot / utils.Config.Chain.ClConfig.SlotsPerEpoch
			// check if slot is post hardfork
			if epoch < utils.Config.Chain.ClConfig.AltairForkEpoch {
				d.log.Infof("skipping sync rewards for slot %d (before altair)", slot)
				continue
			}
			g2.Go(func() error {
				// aquiring semaphore
				nodeId := getNodeId(slot)
				heavyRequestsSem, _ := heavyRequestsSemMap.Load("medium")
				err := heavyRequestsSem.(*semaphore.Weighted).Acquire(context.Background(), 1)
				if err != nil {
					return err
				}
				defer heavyRequestsSem.(*semaphore.Weighted).Release(1)
				//start := time.Now()
				data, err := d.CL[nodeId].GetSyncRewards(slot)
				if err != nil {
					httpErr := network.SpecificError(err)
					if httpErr != nil && httpErr.StatusCode == 404 {
						d.log.Infof("no sync rewards for slot %d", slot)
						return nil
					}
					d.log.Error(err, "can not get sync rewards", 0, map[string]interface{}{"slot": slot})
					return err
				}
				writeMutex.Lock()
				tar.slotBasedData.rewards.syncCommitteeRewards[slot] = *data
				writeMutex.Unlock()
				// d.log.Infof("fetched sync rewards for slot %d in %v", slot, time.Since(start))
				return nil
			})
		}
		err := g2.Wait()
		if err != nil {
			return fmt.Errorf("error in sync rewards: %w", err)
		}
		// add to debug
		debugTimes.Store("syncRewards", time.Since(start))
		return nil
	})
	// block assignments
	g1.Go(func() error {
		start := time.Now()
		// get block assignments
		g2 := &errgroup.Group{}
		g2.SetLimit(epochFetchParallelismWithinFetch)
		writeMutex := &sync.Mutex{}
		for e := epochStart; e <= epochEnd; e++ {
			epoch := e
			g2.Go(func() error {
				heavyRequestsSem, _ := heavyRequestsSemMap.Load("medium")
				err := heavyRequestsSem.(*semaphore.Weighted).Acquire(context.Background(), 1)
				if err != nil {
					return err
				}
				defer heavyRequestsSem.(*semaphore.Weighted).Release(1)
				//start := time.Now()
				start := time.Now()
				data, err := d.CL[getNodeId(epoch)].GetPropoalAssignments(epoch)
				if err != nil {
					d.log.Error(err, "can not get block assignments", 0, map[string]interface{}{"epoch": epoch})
					return err
				}
				writeMutex.Lock()
				for _, p := range data.Data {
					tar.slotBasedData.assignments.blockAssignments[uint64(p.Slot)] = p.ValidatorIndex
				}
				writeMutex.Unlock()
				d.log.Infof("fetched block assignments for epoch %d in %v", epoch, time.Since(start))
				return nil
			})
		}
		err := g2.Wait()
		if err != nil {
			return fmt.Errorf("error in block assignments: %w", err)
		}
		// add to debug
		debugTimes.Store("blockAssignments", time.Since(start))
		return nil
	})

	// attestation rewards
	g1.Go(func() error {
		start := time.Now()
		// get attestation rewards
		g2 := &errgroup.Group{}
		writeMutex := &sync.Mutex{}
		// once per epoch, no extra epochs needed
		for e := epochStart; e <= epochEnd; e++ {
			epoch := e
			g2.Go(func() error {
				// aquiring semaphore
				nodeId := getNodeId(epoch)
				heavyRequestsSem, _ := heavyRequestsSemMap.Load("heavy")
				err := heavyRequestsSem.(*semaphore.Weighted).Acquire(context.Background(), 1)
				if err != nil {
					return err
				}
				defer heavyRequestsSem.(*semaphore.Weighted).Release(1)
				start := time.Now()
				data, err := d.CL[nodeId].GetAttestationRewards(epoch)
				if err != nil {
					d.log.Error(err, "can not get attestation rewards", 0, map[string]interface{}{"epoch": epoch})
					return err
				}
				// ideal
				ideal := make(map[uint64]constypes.AttestationIdealReward)
				for _, idealReward := range data.Data.IdealRewards {
					ideal[uint64(idealReward.EffectiveBalance)] = idealReward
				}
				writeMutex.Lock()
				tar.epochBasedData.rewards.attestationRewards[epoch] = data.Data.TotalRewards
				tar.epochBasedData.rewards.attestationIdealRewards[epoch] = ideal
				writeMutex.Unlock()
				d.log.Infof("fetched attestation rewards for epoch %d in %v", epoch, time.Since(start))
				return nil
			})

		}
		err := g2.Wait()
		if err != nil {
			return fmt.Errorf("error in attestation rewards: %w", err)
		}
		debugTimes.Store("attestationRewards", time.Since(start))
		return nil
	})
	// attestation assignments
	g1.Go(func() error {
		start := time.Now()
		// get attestation assignments
		g2 := &errgroup.Group{}
		writeMutex := &sync.Mutex{}
		for e := epochStart; e <= epochEnd; e++ {
			epoch := e
			g2.Go(func() error {
				// fetch assignment using last fetchSlot in epoch. somehow thats faster than using the first fetchSlot. dont ask why
				fetchSlot := (epoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch - 1
				nodeId := getNodeId(epoch)
				heavyRequestsSem, _ := heavyRequestsSemMap.Load("heavy")
				err := heavyRequestsSem.(*semaphore.Weighted).Acquire(context.Background(), 1)
				if err != nil {
					return err
				}
				defer heavyRequestsSem.(*semaphore.Weighted).Release(1)
				start := time.Now()
				data, err := d.CL[nodeId].GetCommittees(fetchSlot, nil, nil, nil)
				if err != nil {
					d.log.Error(err, "can not get attestation assignments", 0, map[string]interface{}{"slot": fetchSlot})
					return err
				}
				d.log.Infof("retrieved attestation assignments for epoch %d in %v", epoch, time.Since(start))
				writeMutex.Lock()
				for _, committee := range data.Data {
					// todo replace with single alloc variant that uses config values (config has 0 when the code hits here)
					// preallocate
					if _, ok := tar.slotBasedData.assignments.attestationAssignments[committee.Slot]; !ok {
						tar.slotBasedData.assignments.attestationAssignments[committee.Slot] = make([][]uint64, committee.Index+1)
					}
					// if not long enough, extend
					if l := len(tar.slotBasedData.assignments.attestationAssignments[committee.Slot]); l < int(committee.Index)+1 {
						tar.slotBasedData.assignments.attestationAssignments[committee.Slot] = append(
							tar.slotBasedData.assignments.attestationAssignments[committee.Slot],
							make([][]uint64, int(committee.Index)+1-l)...,
						)
					}
					// if not enough space for validators, allocate
					if l := len(tar.slotBasedData.assignments.attestationAssignments[committee.Slot][committee.Index]); l < len(committee.Validators) {
						tar.slotBasedData.assignments.attestationAssignments[committee.Slot][committee.Index] = make([]uint64, len(committee.Validators))
					}
					// ass
					for i, valIndex := range committee.Validators {
						tar.slotBasedData.assignments.attestationAssignments[committee.Slot][committee.Index][i] = uint64(valIndex)
					}
				}
				writeMutex.Unlock()
				d.log.Infof("fetched attestation assignments for epoch %d in %v", epoch, time.Since(start))
				return nil
			})
		}
		err := g2.Wait()
		if err != nil {
			return fmt.Errorf("error in attestation assignments: %w", err)
		}
		debugTimes.Store("attestationAssignments", time.Since(start))
		return nil
	})
	// d.log.Infof("[time] epoch data fetcher, fetched %v epochs %v in %v. Remaining: %v (%v)", len(processed), gapGroup.Epochs, time.Since(start), remaining, remainingTimeEst)
	//metrics.TaskDuration.WithLabelValues("exporter_v2dash_fetch_epochs").Observe(time.Since(start).Seconds())
	//metrics.TaskDuration.WithLabelValues("exporter_v2dash_fetch_epochs_per_epochs").Observe(time.Since(start).Seconds() / float64(len(processed)))

	// lets finish for now
	err := g1.Wait()
	if err != nil {
		return fmt.Errorf("error in getDataForEpochRange: %w", err)
	}
	// debug message to discord
	var msg string
	// sort by debug time seen
	sortedKeys := make([]string, 0)
	debugTimes.Range(
		func(key, value any) bool {
			sortedKeys = append(sortedKeys, key.(string))
			return true
		})
	slices.SortFunc(sortedKeys,
		func(a, b string) int {
			av, _ := debugTimes.Load(a)
			bv, _ := debugTimes.Load(b)
			return int((av.(time.Duration) - bv.(time.Duration)).Nanoseconds())
		})
	for _, k := range sortedKeys {
		v, _ := debugTimes.Load(k)
		msg += fmt.Sprintf("%s: %v\n", k,
			v)
		// also expose over metrics
		metrics.TaskDuration.WithLabelValues("exporter_v25dash_fetch_" + k).Observe(v.(time.Duration).Seconds())
	}
	d.log.Infof("debug times:\n%s", msg)
	// send message
	utils.SendMessage(fmt.Sprintf("Debug Times getDataForEpochRange %d to %d\n```\n%s```", epochStart, epochEnd, msg), &utils.Config.InternalAlerts)
	return nil
}

type EpochParallelGroup struct {
	Epochs     []uint64
	Sequential bool
}

func (d *dashboardData) refreshAllRollings(epoch uint64) error {
	start := time.Now()
	d.log.Infof("refreshing all rollings")
	tables := [][]string{
		{"validator_dashboard_data_rolling_1h", "1 hour"},
		{"validator_dashboard_data_rolling_24h", "24 hours"},
		{"validator_dashboard_data_rolling_7d", "7 days"},
		{"validator_dashboard_data_rolling_30d", "30 days"},
		{"validator_dashboard_data_rolling_90d", "90 days"},
		{"validator_dashboard_data_rolling_total", "20 years"},
	}
	g := &errgroup.Group{}
	g.SetLimit(databaseAggregationParallelism)
	debugTimes := sync.Map{}
	for _, t := range tables {
		targetTable := t[0]
		interval := t[1]
		g.Go(func() error {
			start := time.Now()
			err := d.refreshRollings(targetTable, interval)
			if err != nil {
				return err
			}
			d.log.Infof("refreshed rolling %s in %v", targetTable, time.Since(start))
			debugTimes.Store(targetTable, time.Since(start))
			metrics.State.WithLabelValues(fmt.Sprintf("exporter_v25dash_rolling_%s_end_epoch", targetTable)).Set(float64(epoch))
			return nil
		})
	}
	err := g.Wait()
	if err != nil {
		return fmt.Errorf("error in refreshAllRollings: %w", err)
	}
	// debug message to discord
	var msg string
	// sort by debug time seen
	sortedKeys := make([]string, 0)
	debugTimes.Range(
		func(key, value any) bool {
			sortedKeys = append(sortedKeys, key.(string))
			return true
		})
	slices.SortFunc(sortedKeys,
		func(a, b string) int {
			av, _ := debugTimes.Load(a)
			bv, _ := debugTimes.Load(b)
			return int((av.(time.Duration) - bv.(time.Duration)).Nanoseconds())
		})
	for _, k := range sortedKeys {
		v, _ := debugTimes.Load(k)
		msg += fmt.Sprintf("%s: %v\n", k,
			v)
		// also expose over metrics
		metrics.TaskDuration.WithLabelValues("exporter_v25dash_roll_" + k).Observe(v.(time.Duration).Seconds())
	}
	d.log.Infof("debug times:\n%s", msg)
	// send message
	utils.SendMessage(fmt.Sprintf("Debug Times refreshAllRollings\n```\n%s```", msg), &utils.Config.InternalAlerts)
	d.log.Infof("refreshed all rollings in %v", time.Since(start))
	metrics.TaskDuration.WithLabelValues("exporter_v25dash_refresh_all_rollings").Observe(time.Since(start).Seconds())
	return nil
}

func (d *dashboardData) refreshRollings(targetTable string, interval string) error {
	d.log.Infof("refreshing rolling %s", targetTable)
	query := fmt.Sprintf(`
		with
			min_ts as (
				select
					max(epoch_timestamp) - Interval %[1]s as t
				from
					mainnet_legacy.validator_dashboard_data_epoch
			),
			monthly_aggregates as (
				select
					min(month) as t,
					validator_index,
					min(epoch_start) AS epoch_start,
					max(epoch_end) AS epoch_end,
					argMinStateMerge (balance_start) AS balance_start,
					argMaxStateMerge (balance_end) AS balance_end,
					min(balance_min) AS balance_min,
					max(balance_max) AS balance_max,
					sum(deposits_count) AS deposits_count,
					sum(deposits_amount) AS deposits_amount,
					sum(withdrawals_count) AS withdrawals_count,
					sum(withdrawals_amount) AS withdrawals_amount,
					sum(attestations_scheduled) AS attestations_scheduled,
					sum(attestations_executed) AS attestations_executed,
					sum(attestation_head_executed) AS attestation_head_executed,
					sum(attestation_source_executed) AS attestation_source_executed,
					sum(attestation_target_executed) AS attestation_target_executed,
					sum(attestations_reward) AS attestations_reward,
					sum(attestations_reward_rewards_only) AS attestations_reward_rewards_only,
					sum(attestations_reward_penalties_only) AS attestations_reward_penalties_only,
					sum(attestations_source_reward) AS attestations_source_reward,
					sum(attestations_target_reward) AS attestations_target_reward,
					sum(attestations_head_reward) AS attestations_head_reward,
					sum(attestations_inactivity_reward) AS attestations_inactivity_reward,
					sum(attestations_inclusion_reward) AS attestations_inclusion_reward,
					sum(attestations_ideal_reward) AS attestations_ideal_reward,
					sum(attestations_ideal_source_reward) AS attestations_ideal_source_reward,
					sum(attestations_ideal_target_reward) AS attestations_ideal_target_reward,
					sum(attestations_ideal_head_reward) AS attestations_ideal_head_reward,
					sum(attestations_ideal_inactivity_reward) AS attestations_ideal_inactivity_reward,
					sum(attestations_ideal_inclusion_reward) AS attestations_ideal_inclusion_reward,
					sum(inclusion_delay_sum) AS inclusion_delay_sum,
					sum(optimal_inclusion_delay_sum) AS optimal_inclusion_delay_sum,
					sum(blocks_scheduled) AS blocks_scheduled,
					sum(blocks_proposed) AS blocks_proposed,
					sum(blocks_cl_reward) AS blocks_cl_reward,
					sum(blocks_cl_attestations_reward) AS blocks_cl_attestations_reward,
					sum(blocks_cl_sync_aggregate_reward) AS blocks_cl_sync_aggregate_reward,
					sum(blocks_cl_slasher_reward) AS blocks_cl_slasher_reward,
					sum(blocks_slashing_count) AS blocks_slashing_count,
					sum(blocks_expected) AS blocks_expected,
					sum(sync_scheduled) AS sync_scheduled,
					sum(sync_executed) AS sync_executed,
					sum(sync_rewards) AS sync_rewards,
					sum(sync_committees_expected) AS sync_committees_expected,
					max(slashed) AS slashed,
					max(last_executed_duty_epoch) AS last_executed_duty_epoch,
					max(last_scheduled_sync_epoch) AS last_scheduled_sync_epoch,
					max(last_scheduled_block_epoch) AS last_scheduled_block_epoch
				from
					validator_dashboard_data_monthly foo
				where
					month > (
						select
							t
						from
							min_ts
					)
				group by
					validator_index
			),
			monthly_floor as (
				select
					min(t) as t
				from
					(
						select
							min(toNullable (t)) as t
						from
							monthly_aggregates
						union all
						select
							max(epoch_timestamp) as t
						from
							validator_dashboard_data_epoch
					)
			),
			daily_aggregates as (
				SELECT
					MIN(day) AS t,
					validator_index,
					min(epoch_start) AS epoch_start,
					max(epoch_end) AS epoch_end,
					argMinStateMerge (balance_start) AS balance_start,
					argMaxStateMerge (balance_end) AS balance_end,
					min(balance_min) AS balance_min,
					max(balance_max) AS balance_max,
					sum(deposits_count) AS deposits_count,
					sum(deposits_amount) AS deposits_amount,
					sum(withdrawals_count) AS withdrawals_count,
					sum(withdrawals_amount) AS withdrawals_amount,
					sum(attestations_scheduled) AS attestations_scheduled,
					sum(attestations_executed) AS attestations_executed,
					sum(attestation_head_executed) AS attestation_head_executed,
					sum(attestation_source_executed) AS attestation_source_executed,
					sum(attestation_target_executed) AS attestation_target_executed,
					sum(attestations_reward) AS attestations_reward,
					sum(attestations_reward_rewards_only) AS attestations_reward_rewards_only,
					sum(attestations_reward_penalties_only) AS attestations_reward_penalties_only,
					sum(attestations_source_reward) AS attestations_source_reward,
					sum(attestations_target_reward) AS attestations_target_reward,
					sum(attestations_head_reward) AS attestations_head_reward,
					sum(attestations_inactivity_reward) AS attestations_inactivity_reward,
					sum(attestations_inclusion_reward) AS attestations_inclusion_reward,
					sum(attestations_ideal_reward) AS attestations_ideal_reward,
					sum(attestations_ideal_source_reward) AS attestations_ideal_source_reward,
					sum(attestations_ideal_target_reward) AS attestations_ideal_target_reward,
					sum(attestations_ideal_head_reward) AS attestations_ideal_head_reward,
					sum(attestations_ideal_inactivity_reward) AS attestations_ideal_inactivity_reward,
					sum(attestations_ideal_inclusion_reward) AS attestations_ideal_inclusion_reward,
					sum(inclusion_delay_sum) AS inclusion_delay_sum,
					sum(optimal_inclusion_delay_sum) AS optimal_inclusion_delay_sum,
					sum(blocks_scheduled) AS blocks_scheduled,
					sum(blocks_proposed) AS blocks_proposed,
					sum(blocks_cl_reward) AS blocks_cl_reward,
					sum(blocks_cl_attestations_reward) AS blocks_cl_attestations_reward,
					sum(blocks_cl_sync_aggregate_reward) AS blocks_cl_sync_aggregate_reward,
					sum(blocks_cl_slasher_reward) AS blocks_cl_slasher_reward,
					sum(blocks_slashing_count) AS blocks_slashing_count,
					sum(blocks_expected) AS blocks_expected,
					sum(sync_scheduled) AS sync_scheduled,
					sum(sync_executed) AS sync_executed,
					sum(sync_rewards) AS sync_rewards,
					sum(sync_committees_expected) AS sync_committees_expected,
					max(slashed) AS slashed,
					max(last_executed_duty_epoch) AS last_executed_duty_epoch,
					max(last_scheduled_sync_epoch) AS last_scheduled_sync_epoch,
					max(last_scheduled_block_epoch) AS last_scheduled_block_epoch
				FROM
					validator_dashboard_data_daily foo
				WHERE
					day > (
						select
							t
						from
							min_ts
					)
					AND day < (
						select
							t
						from
							monthly_floor
					)
				GROUP BY
					validator_index
			),
			daily_floor as (
				select
					min(t) as t
				from
					(
						SELECT
							min(toNullable (t)) as t
						from
							daily_aggregates
						union all
						select
							t
						from
							monthly_floor
					)
			),
			hourly_aggregates as (
				SELECT
					MIN(hour) AS t,
					validator_index,
					min(epoch_start) AS epoch_start,
					max(epoch_end) AS epoch_end,
					argMinStateMerge (balance_start) AS balance_start,
					argMaxStateMerge (balance_end) AS balance_end,
					min(balance_min) AS balance_min,
					max(balance_max) AS balance_max,
					sum(deposits_count) AS deposits_count,
					sum(deposits_amount) AS deposits_amount,
					sum(withdrawals_count) AS withdrawals_count,
					sum(withdrawals_amount) AS withdrawals_amount,
					sum(attestations_scheduled) AS attestations_scheduled,
					sum(attestations_executed) AS attestations_executed,
					sum(attestation_head_executed) AS attestation_head_executed,
					sum(attestation_source_executed) AS attestation_source_executed,
					sum(attestation_target_executed) AS attestation_target_executed,
					sum(attestations_reward) AS attestations_reward,
					sum(attestations_reward_rewards_only) AS attestations_reward_rewards_only,
					sum(attestations_reward_penalties_only) AS attestations_reward_penalties_only,
					sum(attestations_source_reward) AS attestations_source_reward,
					sum(attestations_target_reward) AS attestations_target_reward,
					sum(attestations_head_reward) AS attestations_head_reward,
					sum(attestations_inactivity_reward) AS attestations_inactivity_reward,
					sum(attestations_inclusion_reward) AS attestations_inclusion_reward,
					sum(attestations_ideal_reward) AS attestations_ideal_reward,
					sum(attestations_ideal_source_reward) AS attestations_ideal_source_reward,
					sum(attestations_ideal_target_reward) AS attestations_ideal_target_reward,
					sum(attestations_ideal_head_reward) AS attestations_ideal_head_reward,
					sum(attestations_ideal_inactivity_reward) AS attestations_ideal_inactivity_reward,
					sum(attestations_ideal_inclusion_reward) AS attestations_ideal_inclusion_reward,
					sum(inclusion_delay_sum) AS inclusion_delay_sum,
					sum(optimal_inclusion_delay_sum) AS optimal_inclusion_delay_sum,
					sum(blocks_scheduled) AS blocks_scheduled,
					sum(blocks_proposed) AS blocks_proposed,
					sum(blocks_cl_reward) AS blocks_cl_reward,
					sum(blocks_cl_attestations_reward) AS blocks_cl_attestations_reward,
					sum(blocks_cl_sync_aggregate_reward) AS blocks_cl_sync_aggregate_reward,
					sum(blocks_cl_slasher_reward) AS blocks_cl_slasher_reward,
					sum(blocks_slashing_count) AS blocks_slashing_count,
					sum(blocks_expected) AS blocks_expected,
					sum(sync_scheduled) AS sync_scheduled,
					sum(sync_executed) AS sync_executed,
					sum(sync_rewards) AS sync_rewards,
					sum(sync_committees_expected) AS sync_committees_expected,
					max(slashed) AS slashed,
					max(last_executed_duty_epoch) AS last_executed_duty_epoch,
					max(last_scheduled_sync_epoch) AS last_scheduled_sync_epoch,
					max(last_scheduled_block_epoch) AS last_scheduled_block_epoch
				FROM
					validator_dashboard_data_hourly foo
				WHERE
					hour > (
						select
							t
						from
							min_ts
					)
					AND hour < (
						select
							t
						from
							daily_floor
					)
				GROUP BY
					validator_index
			),
			hourly_floor as (
				select
					min(t) as t
				from
					(
						SELECT
							min(toNullable (t)) as t
						from
							hourly_aggregates
						union all
						SELECT
							t
						from
							daily_floor
					)
			),
			epochly_aggregates as (
				SELECT
					MIN(epoch_timestamp) AS t,
					validator_index,
					min(foo.epoch) AS epoch_start,
					max(foo.epoch) AS epoch_end,
					argMinState (foo.balance_start, foo.epoch) AS balance_start,
					argMaxState (foo.balance_end, foo.epoch) AS balance_end,
					least (min(foo.balance_start), min(foo.balance_end)) AS balance_min,
					greatest (max(foo.balance_start), max(foo.balance_end)) AS balance_max,
					sum(deposits_count) AS deposits_count,
					sum(deposits_amount) AS deposits_amount,
					sum(withdrawals_count) AS withdrawals_count,
					sum(withdrawals_amount) AS withdrawals_amount,
					sum(attestations_scheduled) AS attestations_scheduled,
					sum(attestations_executed) AS attestations_executed,
					sum(attestation_head_executed) AS attestation_head_executed,
					sum(attestation_source_executed) AS attestation_source_executed,
					sum(attestation_target_executed) AS attestation_target_executed,
					sum(foo.attestations_reward) AS attestations_reward,
					sumIf (
						foo.attestations_reward,
						foo.attestations_reward > 0
					) AS attestations_reward_rewards_only,
					sumIf (
						foo.attestations_reward,
						foo.attestations_reward < 0
					) AS attestations_reward_penalties_only,
					sum(attestations_source_reward) AS attestations_source_reward,
					sum(attestations_target_reward) AS attestations_target_reward,
					sum(attestations_head_reward) AS attestations_head_reward,
					sum(attestations_inactivity_reward) AS attestations_inactivity_reward,
					sum(attestations_inclusion_reward) AS attestations_inclusion_reward,
					sum(attestations_ideal_reward) AS attestations_ideal_reward,
					sum(attestations_ideal_source_reward) AS attestations_ideal_source_reward,
					sum(attestations_ideal_target_reward) AS attestations_ideal_target_reward,
					sum(attestations_ideal_head_reward) AS attestations_ideal_head_reward,
					sum(attestations_ideal_inactivity_reward) AS attestations_ideal_inactivity_reward,
					sum(attestations_ideal_inclusion_reward) AS attestations_ideal_inclusion_reward,
					sum(inclusion_delay_sum) AS inclusion_delay_sum,
					sum(optimal_inclusion_delay_sum) AS optimal_inclusion_delay_sum,
					sum(blocks_scheduled) AS blocks_scheduled,
					sum(blocks_proposed) AS blocks_proposed,
					sum(blocks_cl_reward) AS blocks_cl_reward,
					sum(blocks_cl_attestations_reward) AS blocks_cl_attestations_reward,
					sum(blocks_cl_sync_aggregate_reward) AS blocks_cl_sync_aggregate_reward,
					sum(blocks_cl_slasher_reward) AS blocks_cl_slasher_reward,
					sum(blocks_slashing_count) AS blocks_slashing_count,
					sum(blocks_expected) AS blocks_expected,
					sum(sync_scheduled) AS sync_scheduled,
					sum(sync_executed) AS sync_executed,
					sum(sync_rewards) AS sync_rewards,
					sum(sync_committees_expected) AS sync_committees_expected,
					max(slashed) AS slashed,
					maxIfOrNull (
						foo.epoch,
						(foo.blocks_proposed != 0)
						OR (foo.sync_executed != 0)
						OR (foo.attestations_executed != 0)
					) AS last_executed_duty_epoch,
					maxIfOrNull (foo.epoch, foo.sync_scheduled != 0) AS last_scheduled_sync_epoch,
					maxIfOrNull (foo.epoch, foo.blocks_proposed != 0) AS last_scheduled_block_epoch
				FROM
					validator_dashboard_data_epoch foo
				WHERE
					epoch_timestamp > (
						select
							t
						from
							min_ts
					)
					AND epoch_timestamp < (
						select
							t
						from
							hourly_floor
					)
				GROUP BY
					validator_index
			),
			finals as (
				select
					validator_index,
					min(epoch_start) AS epoch_start,
					max(epoch_end) AS epoch_end,
					argMinMerge (balance_start) AS balance_start,
					argMaxMerge (balance_end) AS balance_end,
					min(balance_min) AS balance_min,
					max(balance_max) AS balance_max,
					sum(deposits_count) AS deposits_count,
					sum(deposits_amount) AS deposits_amount,
					sum(withdrawals_count) AS withdrawals_count,
					sum(withdrawals_amount) AS withdrawals_amount,
					sum(attestations_scheduled) AS attestations_scheduled,
					sum(attestations_executed) AS attestations_executed,
					sum(attestation_head_executed) AS attestation_head_executed,
					sum(attestation_source_executed) AS attestation_source_executed,
					sum(attestation_target_executed) AS attestation_target_executed,
					sum(attestations_reward) AS attestations_reward,
					sum(attestations_reward_rewards_only) AS attestations_reward_rewards_only,
					sum(attestations_reward_penalties_only) AS attestations_reward_penalties_only,
					sum(attestations_source_reward) AS attestations_source_reward,
					sum(attestations_target_reward) AS attestations_target_reward,
					sum(attestations_head_reward) AS attestations_head_reward,
					sum(attestations_inactivity_reward) AS attestations_inactivity_reward,
					sum(attestations_inclusion_reward) AS attestations_inclusion_reward,
					sum(attestations_ideal_reward) AS attestations_ideal_reward,
					sum(attestations_ideal_source_reward) AS attestations_ideal_source_reward,
					sum(attestations_ideal_target_reward) AS attestations_ideal_target_reward,
					sum(attestations_ideal_head_reward) AS attestations_ideal_head_reward,
					sum(attestations_ideal_inactivity_reward) AS attestations_ideal_inactivity_reward,
					sum(attestations_ideal_inclusion_reward) AS attestations_ideal_inclusion_reward,
					sum(inclusion_delay_sum) AS inclusion_delay_sum,
					sum(optimal_inclusion_delay_sum) AS optimal_inclusion_delay_sum,
					sum(blocks_scheduled) AS blocks_scheduled,
					sum(blocks_proposed) AS blocks_proposed,
					sum(blocks_cl_reward) AS blocks_cl_reward,
					sum(blocks_cl_attestations_reward) AS blocks_cl_attestations_reward,
					sum(blocks_cl_sync_aggregate_reward) AS blocks_cl_sync_aggregate_reward,
					sum(blocks_cl_slasher_reward) AS blocks_cl_slasher_reward,
					sum(blocks_slashing_count) AS blocks_slashing_count,
					sum(blocks_expected) AS blocks_expected,
					sum(sync_scheduled) AS sync_scheduled,
					sum(sync_executed) AS sync_executed,
					sum(sync_rewards) AS sync_rewards,
					sum(sync_committees_expected) AS sync_committees_expected,
					max(slashed) AS slashed,
					max(last_executed_duty_epoch) AS last_executed_duty_epoch,
					max(last_scheduled_sync_epoch) AS last_scheduled_sync_epoch,
					max(last_scheduled_block_epoch) AS last_scheduled_block_epoch
				from
					(
						select
							*
						from
							monthly_aggregates
						UNION ALL
						select
							*
						from
							daily_aggregates
						UNION ALL
						select
							*
						from
							hourly_aggregates
						UNION ALL
						select
							*
						from
							epochly_aggregates
					) foo
				group by
					validator_index
			)
		select
			NOW() as version,
			*
		from
			finals
	`, interval)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	err := db.ClickHouseNativeWriter.Exec(ctx, fmt.Sprintf("INSERT INTO %s %s", targetTable, query))
	if err != nil {
		return fmt.Errorf("error in refreshRollings: %w", err)
	}
	return nil
}

func (d *dashboardData) maintainEpochs(staleEpochs []uint64) error {
	if len(staleEpochs) == 0 {
		return nil
	}
	// get gaps in the final table against the _unsafe epoch table, limited to 10 epochs to restrict load for now
	// enforce that the gaps match our expected epoch bulk size, fatal if not
	GROUP_SIZE := 1
	// ignore the last group as it is expected to be partial - we will never maintain partial groups
	lastGroup := uint64(staleEpochs[len(staleEpochs)-1]) / uint64(GROUP_SIZE)
	groups := make(map[uint64][]uint64)
	for _, e := range staleEpochs {
		groupId := e / uint64(GROUP_SIZE)
		if _, ok := groups[groupId]; !ok {
			groups[groupId] = make([]uint64, 0)
		}
		groups[groupId] = append(groups[groupId], e)
	}
	// check if the last group is incomplete. if so, remove it
	if len(groups[lastGroup]) != GROUP_SIZE {
		delete(groups, lastGroup)
	}
	// enforce that each group is exactly group size
	for groupId, group := range groups {
		if len(group) != GROUP_SIZE {
			// for now lets try doing the full group again
			d.log.Warnf("epoch group %d is not of size %d, doing full group", groupId, GROUP_SIZE)
			g := make([]uint64, GROUP_SIZE)
			for j := 0; j < GROUP_SIZE; j++ {
				g[j] = groupId*uint64(GROUP_SIZE) + uint64(j)
			}
			groups[groupId] = g
		}
	}

	eg := &errgroup.Group{}
	eg.SetLimit(databaseEpochMaintainParallelism)
	if len(groups) > 0 {
		d.log.Infof("maintaining groups: %v", groups)
		values := maps.Values(groups)
		slices.SortFunc(values, func(a, b []uint64) int {
			return int(a[0] - b[0])
		})
		for _, group := range values {
			targetGroup := group
			eg.Go(func() error {
				ss := time.Now()
				err := d.maintainEpochGroup(targetGroup)
				if err != nil {
					utils.SendMessage(fmt.Sprintf("error maintaining epoch group %d: \n`%v`", targetGroup, err), &utils.Config.InternalAlerts)
					return err
				}
				utils.SendMessage(fmt.Sprintf("maintained epoch group %d in %v", targetGroup, time.Since(ss)), &utils.Config.InternalAlerts)
				d.log.Infof("maintained epoch group %d", targetGroup)
				return nil
			})
		}
	}
	err := eg.Wait()
	if err != nil {
		return fmt.Errorf("error in maintainEpochs: %w", err)
	}
	return nil
}

func (d *dashboardData) maintainEpochGroup(targetGroup []uint64) error {
	abortCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	ctx := ch.Context(abortCtx, ch.WithSettings(ch.Settings{
		"insert_deduplication_token":    fmt.Sprintf("epoch_%d_%d", targetGroup[0], targetGroup[len(targetGroup)-1]),
		"select_sequential_consistency": 1,
	}), ch.WithLogs(func(l *ch.Log) {
		d.log.Infof("clickhouse log: %s", l.Text)
	}))
	err := db.ClickHouseNativeWriter.Exec(ctx,
		fmt.Sprintf(`
		insert into %s
		select
			* EXCEPT _inserted_at
		from
			%s FINAL
		where
			epoch_timestamp >= $1 and epoch_timestamp <= $2 
	`, edb.FinalEpochsTableName, edb.UnsafeEpochsTableName),
		utils.EpochToTime(targetGroup[0]),
		utils.EpochToTime(targetGroup[len(targetGroup)-1]))
	if err != nil {
		return fmt.Errorf("error in maintainEpochGroup: %w", err)
	}
	return nil
}

func getHostname() string {
	hostname, _ := os.Hostname()
	return hostname + utils.Config.HostNameSuffix
}

// func to get tasks for hostname from db
func (d *dashboardData) getTasksFromDb() (tasks []Task, err error) {
	hostname := getHostname()
	// query to get tasks for hostname
	// table: _exporter_tasks
	err = db.ClickHouseReader.Select(&tasks, `
		SELECT
			*
			FROM
			_exporter_tasks
			WHERE
			hostname = {hostname:String}
			ORDER BY priority DESC, start_ts DESC`, // highest prio first, newest first
		ch.Named("hostname", hostname))
	if err != nil {
		// wrap
		return nil, fmt.Errorf("error in getTasksFromDb: %w", err)
	}
	return tasks, nil
}
