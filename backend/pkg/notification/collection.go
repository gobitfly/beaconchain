package notification

import (
	"database/sql"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	gcp_bigtable "cloud.google.com/go/bigtable"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/ethclients"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/services"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	"github.com/gobitfly/beaconchain/pkg/exporter/modules"
	"github.com/lib/pq"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/shopspring/decimal"
)

func InitNotificationCollector(pubkeyCachePath string) {
	err := initPubkeyCache(pubkeyCachePath)
	if err != nil {
		log.Fatal(err, "error initializing pubkey cache path for notifications", 0)
	}

	go ethclients.Init()

	go notificationCollector()
}

// the notificationCollector is responsible for collecting & queuing notifications
// it is epoch based and will only collect notification for a given epoch once
// notifications are collected in ascending epoch order
// the epochs_notified sql table is used to keep track of already notified epochs
// before collecting notifications several db consistency checks are done
func notificationCollector() {
	var once sync.Once
	once.Do(func() {
		gob.Register(&ValidatorProposalNotification{})
		gob.Register(&ValidatorUpcomingProposalNotification{})
		gob.Register(&ValidatorGroupEfficiencyNotification{})
		gob.Register(&ValidatorAttestationNotification{})
		gob.Register(&ValidatorIsOfflineNotification{})
		gob.Register(&ValidatorIsOnlineNotification{})
		gob.Register(&ValidatorGotSlashedNotification{})
		gob.Register(&ValidatorWithdrawalNotification{})
		gob.Register(&NetworkNotification{})
		gob.Register(&RocketpoolNotification{})
		gob.Register(&MonitorMachineNotification{})
		gob.Register(&TaxReportNotification{})
		gob.Register(&EthClientNotification{})
		gob.Register(&SyncCommitteeSoonNotification{})
		gob.Register(&GasAboveThresholdNotification{})
		gob.Register(&GasBelowThresholdNotification{})
	})

	mc, err := modules.GetModuleContext()
	if err != nil {
		log.Fatal(err, "error getting module context", 0)
	}

	go func() {
		log.Infof("starting head notification collector")
		for ; ; time.Sleep(time.Second * 30) {
			// get the head epoch
			head, err := mc.ConsClient.GetChainHead()
			if err != nil {
				log.Error(err, "error getting chain head", 0)
				continue
			}

			headEpoch := head.HeadEpoch

			var lastNotifiedEpoch uint64
			err = db.WriterDb.Get(&lastNotifiedEpoch, "SELECT COUNT(*) FROM epochs_notified_head WHERE epoch = $1 AND event_name = $2", headEpoch, types.ValidatorUpcomingProposalEventName)

			if err != nil {
				log.Error(err, fmt.Sprintf("error checking if upcoming block proposal notifications for epoch %v have already been collected", headEpoch), 0)
				continue
			}

			if lastNotifiedEpoch > 0 {
				log.Warnf("head epoch notifications for epoch %v have already been collected", headEpoch)
				continue
			}

			notifications, err := collectHeadNotifications(mc, headEpoch)
			if err != nil {
				log.Error(err, "error collecting head notifications", 0)
			}

			_, err = db.WriterDb.Exec("INSERT INTO epochs_notified_head (epoch, event_name, senton) VALUES ($1, $2, NOW())", headEpoch, types.ValidatorUpcomingProposalEventName)
			if err != nil {
				log.Error(err, "error marking head notification status for epoch in db", 0)
				continue
			}

			if len(notifications) > 0 {
				err = queueNotifications(headEpoch, notifications)
				if err != nil {
					log.Error(err, "error queuing head notifications", 0)
				}
			}
		}
	}()

	for {
		latestFinalizedEpoch := cache.LatestFinalizedEpoch.Get()

		if latestFinalizedEpoch < 4 {
			log.Error(nil, "pausing notifications until at least 5 epochs have been exported into the db", 0)
			time.Sleep(time.Minute)
			continue
		}

		var lastNotifiedEpoch uint64
		err := db.WriterDb.Get(&lastNotifiedEpoch, "SELECT COALESCE(MAX(epoch), 0) FROM epochs_notified")

		if err != nil {
			log.Error(err, "error retrieving last notified epoch from the db", 0)
			time.Sleep(time.Minute)
			continue
		}

		log.Infof("latest finalized epoch is %v, latest notified epoch is %v", latestFinalizedEpoch, lastNotifiedEpoch)

		if latestFinalizedEpoch < lastNotifiedEpoch {
			log.Error(nil, "notification consistency error, lastest finalized epoch is lower than the last notified epoch!", 0)
			time.Sleep(time.Minute)
			continue
		}

		if latestFinalizedEpoch-lastNotifiedEpoch > 5 {
			log.Infof("last notified epoch is more than 5 epochs behind the last finalized epoch, limiting lookback to last 5 epochs")
			lastNotifiedEpoch = latestFinalizedEpoch - 5
		}

		for epoch := lastNotifiedEpoch + 1; epoch <= latestFinalizedEpoch; epoch++ {
			var exported uint64
			err := db.WriterDb.Get(&exported, "SELECT COUNT(*) FROM epochs WHERE epoch <= $1 AND epoch >= $2", epoch, epoch-3)
			if err != nil {
				log.Error(err, "error retrieving export status of epoch", 0, log.Fields{"epoch": epoch})
				services.ReportStatus("notification-collector", "Error", nil)
				break
			}

			if exported != 4 {
				log.Error(nil, "epoch notification consistency error, epochs are not all yet exported into the db", 0, log.Fields{"epoch start": epoch, "epoch end": epoch - 3, "exported": exported, "wanted": 4})
			}

			start := time.Now()
			log.Infof("collecting notifications for epoch %v", epoch)

			// Network DB Notifications (network related)
			notifications, err := collectNotifications(mc, epoch)

			if err != nil {
				log.Error(err, "error collection notifications", 0)
				services.ReportStatus("notification-collector", "Error", nil)
				break
			}

			_, err = db.WriterDb.Exec("INSERT INTO epochs_notified VALUES ($1, NOW())", epoch)
			if err != nil {
				log.Error(err, "error marking notification status for epoch %v in db: %v", 0, log.Fields{"epoch": epoch})
				services.ReportStatus("notification-collector", "Error", nil)
				break
			}

			err = queueNotifications(epoch, notifications) // this caused the collected notifications to be queued and sent
			if err != nil {
				log.Error(err, fmt.Sprintf("error queuing notifications for epoch %v in db", epoch), 0)
				services.ReportStatus("notification-collector", "Error", nil)
				break
			}

			// Network DB Notifications (user related, must only run on one instance ever!!!!)
			if utils.Config.Notifications.UserDBNotifications {
				log.Infof("collecting user db notifications")
				userNotifications, err := collectUserDbNotifications(epoch)
				if err != nil {
					log.Error(err, "error collection user db notifications", 0)
					services.ReportStatus("notification-collector", "Error", nil)
					time.Sleep(time.Minute * 2)
					continue
				}

				err = queueNotifications(epoch, userNotifications)
				if err != nil {
					log.Error(err, fmt.Sprintf("error queuing user notifications for epoch %v in db", epoch), 0)
					services.ReportStatus("notification-collector", "Error", nil)
					time.Sleep(time.Minute * 2)
					continue
				}
			}

			log.InfoWithFields(log.Fields{"notifications": len(notifications), "duration": time.Since(start), "epoch": epoch}, "notifications completed")

			metrics.TaskDuration.WithLabelValues("service_notifications").Observe(time.Since(start).Seconds())
		}

		services.ReportStatus("notification-collector", "Running", nil)
		time.Sleep(time.Second * 10)
	}
}

func collectHeadNotifications(mc modules.ModuleContext, headEpoch uint64) (types.NotificationsPerUserId, error) {
	notificationsByUserID := types.NotificationsPerUserId{}
	start := time.Now()
	err := collectUpcomingBlockProposalNotifications(notificationsByUserID, mc, headEpoch)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_collect_upcoming_block_proposal").Inc()
		return nil, fmt.Errorf("error collecting upcoming block proposal notifications: %v", err)
	}
	log.Infof("collecting upcoming block proposal notifications took: %v", time.Since(start))

	return notificationsByUserID, nil
}

func collectUpcomingBlockProposalNotifications(notificationsByUserID types.NotificationsPerUserId, mc modules.ModuleContext, headEpoch uint64) (err error) {
	nextEpoch := headEpoch + 1
	log.Infof("collecting upcoming block proposal notifications for epoch %v (head epoch is %d)", nextEpoch, headEpoch)

	if utils.EpochToTime(nextEpoch).Before(time.Now()) {
		log.Error(fmt.Errorf("error upcoming block proposal notifications for epoch %v are already in the past", nextEpoch), "", 0)
		return nil
	}

	assignments, err := mc.CL.GetPropoalAssignments(nextEpoch)
	if err != nil {
		return fmt.Errorf("error getting proposal assignments: %w", err)
	}

	subs, err := GetSubsForEventFilter(types.ValidatorUpcomingProposalEventName, "", nil, nil)
	if err != nil {
		return fmt.Errorf("error getting subscriptions for upcoming block proposal notifications: %w", err)
	}

	log.Infof("retrieved %d subscriptions for upcoming block proposal notifications", len(subs))
	if len(subs) == 0 {
		return nil
	}

	for _, assignment := range assignments.Data {
		log.Infof("upcoming block proposal for validator %d in slot %d", assignment.ValidatorIndex, assignment.Slot)
		for _, sub := range subs[hex.EncodeToString(assignment.Pubkey)] {
			if sub.UserID == nil || sub.ID == nil {
				return fmt.Errorf("error expected userId and subId to be defined but got user: %v, sub: %v", sub.UserID, sub.ID)
			}

			log.Infof("creating %v notification for validator %v in epoch %v (dashboard: %v)", sub.EventName, assignment.ValidatorIndex, nextEpoch, sub.DashboardId != nil)
			n := &ValidatorUpcomingProposalNotification{
				NotificationBaseImpl: types.NotificationBaseImpl{
					SubscriptionID:     *sub.ID,
					UserID:             *sub.UserID,
					Epoch:              nextEpoch,
					EventName:          sub.EventName,
					EventFilter:        hex.EncodeToString(assignment.Pubkey),
					DashboardId:        sub.DashboardId,
					DashboardName:      sub.DashboardName,
					DashboardGroupId:   sub.DashboardGroupId,
					DashboardGroupName: sub.DashboardGroupName,
				},
				ValidatorIndex: assignment.ValidatorIndex,
				Slot:           uint64(assignment.Slot),
			}
			notificationsByUserID.AddNotification(n)
			metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()
		}
	}
	return nil
}

func collectNotifications(mc modules.ModuleContext, epoch uint64) (types.NotificationsPerUserId, error) {
	notificationsByUserID := types.NotificationsPerUserId{}
	start := time.Now()
	var err error
	var dbIsCoherent bool

	// do a consistency check to make sure that we have all the data we need in the db
	err = db.WriterDb.Get(&dbIsCoherent, `
		SELECT
			NOT (array[false] && array_agg(is_coherent)) AS is_coherent
		FROM (
			SELECT
				epoch - 1 = lead(epoch) OVER (ORDER BY epoch DESC) AS is_coherent
			FROM epochs
			ORDER BY epoch DESC
			LIMIT 2^14
		) coherency`)

	if err != nil {
		log.Error(err, "error doing epochs table coherence check", 0)
		return nil, err
	}
	if !dbIsCoherent {
		log.Error(nil, "epochs coherence check failed, aborting", 0)
		return nil, fmt.Errorf("epochs coherence check failed, aborting")
	}

	log.Infof("started collecting notifications")
	// The following functions will collect the notifications and add them to the
	// notificationsByUserID map. The notifications will be queued and sent later
	// by the notification sender process
	err = collectGasPriceNotifications(notificationsByUserID, epoch)
	if err != nil {
		metrics.Errors.WithLabelValues("network_gas_threshold").Inc()
		return nil, fmt.Errorf("error collecting network_gas_threshold notifications: %v", err)
	}
	log.Infof("collecting network gas threshold notifications took: %v", time.Since(start))

	err = collectGroupEfficiencyNotifications(notificationsByUserID, epoch, mc)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_collect_group_efficiency").Inc()
		return nil, fmt.Errorf("error collecting validator_group_efficiency notifications: %v", err)
	}
	log.Infof("collecting group efficiency notifications took: %v", time.Since(start))

	err = collectAttestationAndOfflineValidatorNotifications(notificationsByUserID, epoch)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_collect_missed_attestation").Inc()
		return nil, fmt.Errorf("error collecting validator_attestation_missed notifications: %v", err)
	}
	log.Infof("collecting attestation & offline notifications took: %v", time.Since(start))

	err = collectBlockProposalNotifications(notificationsByUserID, 1, types.ValidatorExecutedProposalEventName, epoch)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_collect_executed_block_proposal").Inc()
		return nil, fmt.Errorf("error collecting validator_proposal_submitted notifications: %v", err)
	}
	log.Infof("collecting block proposal proposed notifications took: %v", time.Since(start))

	err = collectBlockProposalNotifications(notificationsByUserID, 2, types.ValidatorMissedProposalEventName, epoch)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_collect_missed_block_proposal").Inc()
		return nil, fmt.Errorf("error collecting validator_proposal_missed notifications: %v", err)
	}
	log.Infof("collecting block proposal missed notifications took: %v", time.Since(start))

	err = collectBlockProposalNotifications(notificationsByUserID, 3, types.ValidatorMissedProposalEventName, epoch)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_collect_missed_orphaned_block_proposal").Inc()
		return nil, fmt.Errorf("error collecting validator_proposal_missed notifications for orphaned slots: %w", err)
	}
	log.Infof("collecting block proposal missed notifications for orphaned slots took: %v", time.Since(start))

	err = collectValidatorGotSlashedNotifications(notificationsByUserID, epoch)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_collect_validator_got_slashed").Inc()
		return nil, fmt.Errorf("error collecting validator_got_slashed notifications: %v", err)
	}
	log.Infof("collecting validator got slashed notifications took: %v", time.Since(start))

	err = collectWithdrawalNotifications(notificationsByUserID, epoch)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_collect_validator_withdrawal").Inc()
		return nil, fmt.Errorf("error collecting withdrawal notifications: %v", err)
	}
	log.Infof("collecting withdrawal notifications took: %v", time.Since(start))

	err = collectNetworkNotifications(notificationsByUserID)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_collect_network").Inc()
		return nil, fmt.Errorf("error collecting network notifications: %v", err)
	}
	log.Infof("collecting network notifications took: %v", time.Since(start))

	// Rocketpool
	{
		var ts int64
		err = db.ReaderDb.Get(&ts, `SELECT id FROM rocketpool_network_stats LIMIT 1;`)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				log.Infof("skipped the collecting of rocketpool notifications, because rocketpool_network_stats is empty")
			} else {
				metrics.Errors.WithLabelValues("notifications_collect_rocketpool_notifications").Inc()
				return nil, fmt.Errorf("error collecting rocketpool notifications: %v", err)
			}
		} else {
			err = collectRocketpoolComissionNotifications(notificationsByUserID)
			if err != nil {
				//nolint:misspell
				metrics.Errors.WithLabelValues("notifications_collect_rocketpool_comission").Inc()
				return nil, fmt.Errorf("error collecting rocketpool commission: %v", err)
			}
			log.Infof("collecting rocketpool commissions took: %v", time.Since(start))

			err = collectRocketpoolRewardClaimRoundNotifications(notificationsByUserID)
			if err != nil {
				metrics.Errors.WithLabelValues("notifications_collect_rocketpool_reward_claim").Inc()
				return nil, fmt.Errorf("error collecting new rocketpool claim round: %v", err)
			}
			log.Infof("collecting rocketpool claim round took: %v", time.Since(start))

			err = collectRocketpoolRPLCollateralNotifications(notificationsByUserID, types.RocketpoolCollateralMaxReachedEventName, epoch)
			if err != nil {
				metrics.Errors.WithLabelValues("notifications_collect_rocketpool_rpl_collateral_max_reached").Inc()
				return nil, fmt.Errorf("error collecting rocketpool max collateral: %v", err)
			}
			log.Infof("collecting rocketpool max collateral took: %v", time.Since(start))

			err = collectRocketpoolRPLCollateralNotifications(notificationsByUserID, types.RocketpoolCollateralMinReachedEventName, epoch)
			if err != nil {
				metrics.Errors.WithLabelValues("notifications_collect_rocketpool_rpl_collateral_min_reached").Inc()
				return nil, fmt.Errorf("error collecting rocketpool min collateral: %v", err)
			}
			log.Infof("collecting rocketpool min collateral took: %v", time.Since(start))
		}
	}

	err = collectSyncCommitteeNotifications(notificationsByUserID, epoch)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_collect_sync_committee").Inc()
		return nil, fmt.Errorf("error collecting sync committee: %v", err)
	}
	log.Infof("collecting sync committee took: %v", time.Since(start))

	return notificationsByUserID, nil
}

func collectUserDbNotifications(epoch uint64) (types.NotificationsPerUserId, error) {
	notificationsByUserID := types.NotificationsPerUserId{}
	var err error

	// Monitoring (premium): machine offline
	err = collectMonitoringMachineOffline(notificationsByUserID, epoch)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_collect_monitoring_machine_offline").Inc()
		return nil, fmt.Errorf("error collecting Eth client offline notifications: %v", err)
	}

	// Monitoring (premium): disk full warnings
	err = collectMonitoringMachineDiskAlmostFull(notificationsByUserID, epoch)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_collect_monitoring_machine_disk_almost_full").Inc()
		return nil, fmt.Errorf("error collecting Eth client disk full notifications: %v", err)
	}

	// Monitoring (premium): cpu load
	err = collectMonitoringMachineCPULoad(notificationsByUserID, epoch)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_collect_monitoring_machine_cpu_load").Inc()
		return nil, fmt.Errorf("error collecting Eth client cpu notifications: %v", err)
	}

	// Monitoring (premium): ram
	err = collectMonitoringMachineMemoryUsage(notificationsByUserID, epoch)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_collect_monitoring_machine_memory_usage").Inc()
		return nil, fmt.Errorf("error collecting Eth client memory notifications: %v", err)
	}

	// New ETH clients
	err = collectEthClientNotifications(notificationsByUserID)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_collect_eth_client").Inc()
		return nil, fmt.Errorf("error collecting Eth client notifications: %v", err)
	}

	//Tax Report
	err = collectTaxReportNotificationNotifications(notificationsByUserID)
	if err != nil {
		metrics.Errors.WithLabelValues("notifications_collect_tax_report").Inc()
		return nil, fmt.Errorf("error collecting tax report notifications: %v", err)
	}

	return notificationsByUserID, nil
}

func collectGasPriceNotifications(notificationsByUserID types.NotificationsPerUserId, epoch uint64) error {
	// retrieve gas price history
	ts := time.Now()
	duration := time.Minute * 10
	gasPrices, err := db.BigtableClient.GetGasNowHistory(ts, ts.Add(duration*-1))
	if err != nil {
		return fmt.Errorf("error getting gas price history: %w", err)
	}

	if len(gasPrices) == 0 {
		log.Warnf("no gas price data found for epoch %v", epoch)
		return nil
	}

	averageGasPrice := decimal.NewFromInt(0)
	for _, gasPrice := range gasPrices {
		log.Infof("gas price for %v: %v", gasPrice.Ts, gasPrice.Rapid)
		averageGasPrice = averageGasPrice.Add(decimal.NewFromBigInt(gasPrice.Rapid, 0))
	}
	averageGasPrice = averageGasPrice.Div(decimal.NewFromInt(int64(len(gasPrices)))).Div(decimal.NewFromInt(params.GWei))

	log.Infof("average gas price is %f GWei", averageGasPrice.InexactFloat64())

	// retrieve subscriptions
	subMapAboveThreshold, err := GetSubsForEventFilter(types.NetworkGasAboveThresholdEventName, "", nil, nil)
	if err != nil {
		return fmt.Errorf("error getting subscriptions for gas price above threshold: %w", err)
	}

	for _, subs := range subMapAboveThreshold {
		for _, sub := range subs {
			if averageGasPrice.GreaterThan(decimal.NewFromFloat(sub.EventThreshold)) {
				n := &GasAboveThresholdNotification{
					NotificationBaseImpl: types.NotificationBaseImpl{
						SubscriptionID:     *sub.ID,
						UserID:             *sub.UserID,
						Epoch:              sub.CreatedEpoch,
						EventFilter:        sub.EventFilter,
						EventName:          sub.EventName,
						DashboardId:        sub.DashboardId,
						DashboardName:      sub.DashboardName,
						DashboardGroupId:   sub.DashboardGroupId,
						DashboardGroupName: sub.DashboardGroupName,
					},
					AverageGasPrice:   averageGasPrice.InexactFloat64(),
					ThresholdGasPrice: sub.EventThreshold,
				}

				notificationsByUserID.AddNotification(n)
				metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()

				log.Infof("created %s notification for user %d", n.GetInfo(types.NotifciationFormatText), n.GetUserId())
			}
		}
	}

	subMapBelowThreshold, err := GetSubsForEventFilter(types.NetworkGasBelowThresholdEventName, "", nil, nil)
	if err != nil {
		return fmt.Errorf("error getting subscriptions for gas price below threshold: %w", err)
	}
	for _, subs := range subMapBelowThreshold {
		for _, sub := range subs {
			if averageGasPrice.LessThan(decimal.NewFromFloat(sub.EventThreshold)) {
				n := &GasBelowThresholdNotification{
					NotificationBaseImpl: types.NotificationBaseImpl{
						SubscriptionID:     *sub.ID,
						UserID:             *sub.UserID,
						Epoch:              sub.CreatedEpoch,
						EventFilter:        sub.EventFilter,
						EventName:          sub.EventName,
						DashboardId:        sub.DashboardId,
						DashboardName:      sub.DashboardName,
						DashboardGroupId:   sub.DashboardGroupId,
						DashboardGroupName: sub.DashboardGroupName,
					},
					AverageGasPrice:   averageGasPrice.InexactFloat64(),
					ThresholdGasPrice: sub.EventThreshold,
				}

				notificationsByUserID.AddNotification(n)
				metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()
				log.Infof("created %s notification for user %d", n.GetInfo(types.NotifciationFormatText), n.GetUserId())
			}
		}
	}
	return nil
}

func collectGroupEfficiencyNotifications(notificationsByUserID types.NotificationsPerUserId, epoch uint64, mc modules.ModuleContext) error {
	type dbResult struct {
		ValidatorIndex         uint64          `db:"validator_index"`
		AttestationReward      decimal.Decimal `db:"attestations_reward"`
		AttestationIdealReward decimal.Decimal `db:"attestations_ideal_reward"`
		BlocksProposed         uint64          `db:"blocks_proposed"`
		BlocksScheduled        uint64          `db:"blocks_scheduled"`
		SyncExecuted           uint64          `db:"sync_executed"`
		SyncScheduled          uint64          `db:"sync_scheduled"`
	}

	// retrieve rewards for the epoch
	log.Info("retrieving validator metadata")
	validators, err := mc.CL.GetValidators(epoch*utils.Config.Chain.ClConfig.SlotsPerEpoch, nil, []constypes.ValidatorStatus{constypes.Active})
	if err != nil {
		return fmt.Errorf("error getting validators: %w", err)
	}
	effectiveBalanceMap := make(map[uint64]uint64)
	activeValidatorsMap := make(map[uint64]struct{})
	for _, validator := range validators.Data {
		effectiveBalanceMap[validator.Index] = validator.Validator.EffectiveBalance
		activeValidatorsMap[validator.Index] = struct{}{}
	}
	log.Info("retrieving attestation reward data")
	attestationRewards, err := mc.CL.GetAttestationRewards(epoch)
	if err != nil {
		return fmt.Errorf("error getting attestation rewards: %w", err)
	}

	efficiencyMap := make(map[types.ValidatorIndex]*dbResult, len(attestationRewards.Data.TotalRewards))

	idealRewardsMap := make(map[uint64]decimal.Decimal)
	for _, reward := range attestationRewards.Data.IdealRewards {
		idealRewardsMap[uint64(reward.EffectiveBalance)] = decimal.NewFromInt(int64(reward.Head) + int64(reward.Target) + int64(reward.Source) + int64(reward.InclusionDelay) + int64(reward.Inactivity))
	}
	for _, reward := range attestationRewards.Data.TotalRewards {
		efficiencyMap[types.ValidatorIndex(reward.ValidatorIndex)] = &dbResult{
			ValidatorIndex:         reward.ValidatorIndex,
			AttestationReward:      decimal.NewFromInt(int64(reward.Head) + int64(reward.Target) + int64(reward.Source) + int64(reward.InclusionDelay) + int64(reward.Inactivity)),
			AttestationIdealReward: idealRewardsMap[effectiveBalanceMap[reward.ValidatorIndex]],
		}
	}

	log.Info("retrieving block proposal data")
	proposalAssignments, err := mc.CL.GetPropoalAssignments(epoch)
	if err != nil {
		return fmt.Errorf("error getting proposal assignments: %w", err)
	}
	for _, assignment := range proposalAssignments.Data {
		if _, ok := efficiencyMap[types.ValidatorIndex(assignment.ValidatorIndex)]; !ok {
			efficiencyMap[types.ValidatorIndex(assignment.ValidatorIndex)] = &dbResult{
				ValidatorIndex:         assignment.ValidatorIndex,
				AttestationReward:      decimal.Decimal{},
				AttestationIdealReward: decimal.Decimal{},
			}
		}
		efficiencyMap[types.ValidatorIndex(assignment.ValidatorIndex)].BlocksScheduled++
	}

	syncAssignments, err := mc.CL.GetSyncCommitteesAssignments(nil, epoch*utils.Config.Chain.ClConfig.SlotsPerEpoch)
	if err != nil {
		return fmt.Errorf("error getting sync committee assignments: %w", err)
	}

	for slot := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch; slot < (epoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch; slot++ {
		log.Infof("retrieving data for slot %v", slot)
		s, err := mc.CL.GetSlot(slot)
		if err != nil && strings.Contains(err.Error(), "NOT_FOUND") {
			continue
		} else if err != nil {
			return fmt.Errorf("error getting block header for slot %v: %w", slot, err)
		}
		efficiencyMap[types.ValidatorIndex(s.Data.Message.ProposerIndex)].BlocksProposed++

		for i, validatorIndex := range syncAssignments.Data.Validators {
			if _, ok := efficiencyMap[types.ValidatorIndex(validatorIndex)]; !ok {
				efficiencyMap[types.ValidatorIndex(validatorIndex)] = &dbResult{
					ValidatorIndex:         uint64(validatorIndex),
					AttestationReward:      decimal.Decimal{},
					AttestationIdealReward: decimal.Decimal{},
				}
			}
			efficiencyMap[types.ValidatorIndex(validatorIndex)].SyncScheduled++

			if utils.BitAtVector(s.Data.Message.Body.SyncAggregate.SyncCommitteeBits, i) {
				efficiencyMap[types.ValidatorIndex(validatorIndex)].SyncExecuted++
			}
		}
	}

	subMap, err := GetSubsForEventFilter(types.ValidatorGroupEfficiencyEventName, "", nil, nil)
	if err != nil {
		return fmt.Errorf("error getting subscriptions for (missed) block proposals %w", err)
	}

	// create a lookup map for the dashboard & groups
	type groupDetails struct {
		Validators   []types.ValidatorIndex
		Subscription *types.Subscription
	}
	dashboardMap := make(map[types.UserId]map[types.DashboardId]map[types.DashboardGroupId]*groupDetails)

	for _, subs := range subMap {
		for _, sub := range subs {
			if sub.DashboardId == nil || sub.DashboardGroupId == nil {
				continue
			}
			userId := *sub.UserID
			dashboardId := types.DashboardId(*sub.DashboardId)
			groupId := types.DashboardGroupId(*sub.DashboardGroupId)
			if _, ok := dashboardMap[userId]; !ok {
				dashboardMap[userId] = make(map[types.DashboardId]map[types.DashboardGroupId]*groupDetails)
			}
			if _, ok := dashboardMap[userId][dashboardId]; !ok {
				dashboardMap[userId][dashboardId] = make(map[types.DashboardGroupId]*groupDetails)
			}
			if _, ok := dashboardMap[userId][dashboardId][groupId]; !ok {
				dashboardMap[userId][dashboardId][groupId] = &groupDetails{
					Validators: []types.ValidatorIndex{},
				}
			}
			if sub.EventFilter != "" {
				pubkeyDecoded, err := hex.DecodeString(sub.EventFilter)
				if err != nil {
					return fmt.Errorf("error decoding pubkey %v: %w", sub.EventFilter, err)
				}
				validatorIndex, err := GetIndexForPubkey(pubkeyDecoded)
				if err != nil {
					return fmt.Errorf("error getting validator index for pubkey %v: %w", sub.EventFilter, err)
				}
				dashboardMap[userId][dashboardId][groupId].Validators = append(dashboardMap[*sub.UserID][dashboardId][groupId].Validators, types.ValidatorIndex(validatorIndex))
			}
			dashboardMap[userId][dashboardId][groupId].Subscription = sub
		}
	}

	// The commented code below can be used to validate data retrieved from the node against
	// data in clickhouse
	// var queryResult []*dbResult
	// clickhouseTable := "validator_dashboard_data_epoch"
	// // retrieve efficiency data for the epoch
	// log.Infof("retrieving efficiency data for epoch %v", epoch)
	// ds := goqu.Dialect("postgres").
	// 	From(goqu.L(fmt.Sprintf(`%s AS r`, clickhouseTable))).
	// 	Select(
	// 		goqu.L("validator_index"),
	// 		goqu.L("COALESCE(r.attestations_reward, 0) AS attestations_reward"),
	// 		goqu.L("COALESCE(r.attestations_ideal_reward, 0) AS attestations_ideal_reward"),
	// 		goqu.L("COALESCE(r.blocks_proposed, 0) AS blocks_proposed"),
	// 		goqu.L("COALESCE(r.blocks_scheduled, 0) AS blocks_scheduled"),
	// 		goqu.L("COALESCE(r.sync_executed, 0) AS sync_executed"),
	// 		goqu.L("COALESCE(r.sync_scheduled, 0) AS sync_scheduled")).
	// 	Where(goqu.L("r.epoch_timestamp = ?", utils.EpochToTime(epoch)))
	// query, args, err := ds.Prepared(true).ToSQL()
	// if err != nil {
	// 	return fmt.Errorf("error preparing query: %v", err)
	// }

	// err = db.ClickHouseReader.Select(&queryResult, query, args...)
	// if err != nil {
	// 	return fmt.Errorf("error retrieving data from table %s: %v", clickhouseTable, err)
	// }

	// if len(queryResult) == 0 {
	// 	return fmt.Errorf("no efficiency data found for epoch %v", epoch)
	// }

	// log.Infof("retrieved %v efficiency data rows", len(queryResult))

	// for _, row := range queryResult {
	// 	if _, ok := activeValidatorsMap[row.ValidatorIndex]; !ok {
	// 		continue
	// 	}
	// 	existing := efficiencyMap[types.ValidatorIndex(row.ValidatorIndex)]

	// 	if existing == nil {
	// 		existing = &dbResult{
	// 			ValidatorIndex:         row.ValidatorIndex,
	// 			AttestationReward:      decimal.Decimal{},
	// 			AttestationIdealReward: decimal.Decimal{},
	// 		}
	// 	}
	// 	if !existing.AttestationIdealReward.Equal(row.AttestationIdealReward) {
	// 		log.Fatal(fmt.Errorf("ideal reward mismatch for validator %v: %v != %v", row.ValidatorIndex, existing.AttestationIdealReward, row.AttestationIdealReward), "ideal reward mismatch", 0)
	// 	}
	// 	if !existing.AttestationReward.Equal(row.AttestationReward) {
	// 		log.Fatal(fmt.Errorf("attestation reward mismatch for validator %v: %v != %v", row.ValidatorIndex, existing.AttestationReward, row.AttestationReward), "attestation reward mismatch", 0)
	// 	}
	// 	if existing.BlocksProposed != row.BlocksProposed {
	// 		log.Fatal(fmt.Errorf("blocks proposed mismatch for validator %v: %v != %v", row.ValidatorIndex, existing.BlocksProposed, row.BlocksProposed), "blocks proposed mismatch", 0)
	// 	}
	// 	if existing.BlocksScheduled != row.BlocksScheduled {
	// 		log.Fatal(fmt.Errorf("blocks scheduled mismatch for validator %v: %v != %v", row.ValidatorIndex, existing.BlocksScheduled, row.BlocksScheduled), "blocks scheduled mismatch", 0)
	// 	}
	// 	if existing.SyncExecuted != row.SyncExecuted {
	// 		log.Fatal(fmt.Errorf("sync executed mismatch for validator %v: %v != %v", row.ValidatorIndex, existing.SyncExecuted, row.SyncExecuted), "sync executed mismatch", 0)
	// 	}
	// 	if existing.SyncScheduled != row.SyncScheduled {
	// 		log.Fatal(fmt.Errorf("sync scheduled mismatch for validator %v: %v != %v", row.ValidatorIndex, existing.SyncScheduled, row.SyncScheduled), "sync scheduled mismatch", 0)
	// 	}
	// 	efficiencyMap[types.ValidatorIndex(row.ValidatorIndex)] = row
	// }

	for userId, dashboards := range dashboardMap {
		for dashboardId, groups := range dashboards {
			for groupId, groupDetails := range groups {
				attestationReward := decimal.Decimal{}
				attestationIdealReward := decimal.Decimal{}
				blocksProposed := uint64(0)
				blocksScheduled := uint64(0)
				syncExecuted := uint64(0)
				syncScheduled := uint64(0)

				for _, validatorIndex := range groupDetails.Validators {
					if row, ok := efficiencyMap[validatorIndex]; ok {
						attestationReward = attestationReward.Add(row.AttestationReward)
						attestationIdealReward = attestationIdealReward.Add(row.AttestationIdealReward)
						blocksProposed += row.BlocksProposed
						blocksScheduled += row.BlocksScheduled
						syncExecuted += row.SyncExecuted
						syncScheduled += row.SyncScheduled
					}
				}

				var attestationEfficiency, proposerEfficiency, syncEfficiency sql.NullFloat64

				if !attestationIdealReward.IsZero() {
					attestationEfficiency.Float64 = attestationReward.Div(attestationIdealReward).InexactFloat64()
					attestationEfficiency.Valid = true
				}
				if blocksScheduled > 0 {
					proposerEfficiency.Float64 = float64(blocksProposed) / float64(blocksScheduled)
					proposerEfficiency.Valid = true
				}
				if syncScheduled > 0 {
					syncEfficiency.Float64 = float64(syncExecuted) / float64(syncScheduled)
					syncEfficiency.Valid = true
				}

				efficiency := utils.CalculateTotalEfficiency(attestationEfficiency, proposerEfficiency, syncEfficiency) / 100

				log.Infof("efficiency: %v, threshold: %v", efficiency*100, groupDetails.Subscription.EventThreshold*100)

				if efficiency < groupDetails.Subscription.EventThreshold {
					log.Infof("creating group efficiency notification for user %v, dashboard %v, group %v in epoch %v", userId, dashboardId, groupId, epoch)
					n := &ValidatorGroupEfficiencyNotification{
						NotificationBaseImpl: types.NotificationBaseImpl{
							SubscriptionID:     *groupDetails.Subscription.ID,
							UserID:             *groupDetails.Subscription.UserID,
							Epoch:              epoch,
							EventName:          groupDetails.Subscription.EventName,
							EventFilter:        "-",
							DashboardId:        groupDetails.Subscription.DashboardId,
							DashboardName:      groupDetails.Subscription.DashboardName,
							DashboardGroupId:   groupDetails.Subscription.DashboardGroupId,
							DashboardGroupName: groupDetails.Subscription.DashboardGroupName,
						},
						Threshold:  groupDetails.Subscription.EventThreshold,
						Efficiency: efficiency,
					}
					notificationsByUserID.AddNotification(n)
					metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()
				}
			}
		}
	}

	log.Info("done collecting group efficiency notifications")

	return nil
}
func collectBlockProposalNotifications(notificationsByUserID types.NotificationsPerUserId, status uint64, eventName types.EventName, epoch uint64) error {
	type dbResult struct {
		Proposer      uint64 `db:"proposer"`
		Status        uint64 `db:"status"`
		Slot          uint64 `db:"slot"`
		ExecBlock     uint64 `db:"exec_block_number"`
		ExecRewardETH float64
	}

	subMap, err := GetSubsForEventFilter(eventName, "", nil, nil)
	if err != nil {
		return fmt.Errorf("error getting subscriptions for (missed) block proposals %w", err)
	}

	events := make([]dbResult, 0)
	err = db.WriterDb.Select(&events, "SELECT slot, proposer, status, COALESCE(exec_block_number, 0) AS exec_block_number FROM blocks WHERE epoch = $1 AND status = $2", epoch, fmt.Sprintf("%d", status))
	if err != nil {
		return fmt.Errorf("error retrieving slots for epoch %v: %w", epoch, err)
	}

	log.Infof("retrieved %v events", len(events))

	// Get Execution reward for proposed blocks
	if status == 1 { // if proposed
		var blockList = []uint64{}
		for _, data := range events {
			if data.ExecBlock != 0 {
				blockList = append(blockList, data.ExecBlock)
			}
		}

		if len(blockList) > 0 {
			blocks, err := db.BigtableClient.GetBlocksIndexedMultiple(blockList, 10000)
			if err != nil {
				log.Error(err, "error loading blocks from bigtable", 0, log.Fields{"blockList": blockList})
				return err
			}
			var execBlockNrToExecBlockMap = map[uint64]*types.Eth1BlockIndexed{}
			for _, block := range blocks {
				execBlockNrToExecBlockMap[block.GetNumber()] = block
			}
			relaysData, err := db.GetRelayDataForIndexedBlocks(blocks)
			if err != nil {
				return err
			}

			for j := 0; j < len(events); j++ {
				execData, found := execBlockNrToExecBlockMap[events[j].ExecBlock]
				if found {
					reward := utils.Eth1TotalReward(execData)
					relayData, found := relaysData[common.BytesToHash(execData.Hash)]
					if found {
						reward = relayData.MevBribe.BigInt()
					}
					events[j].ExecRewardETH = float64(int64(eth.WeiToEth(reward)*100000)) / 100000
				}
			}
		}
	}

	for _, event := range events {
		pubkey, err := GetPubkeyForIndex(event.Proposer)
		if err != nil {
			log.Error(err, "error retrieving pubkey for validator", 0, map[string]interface{}{"validator": event.Proposer})
			continue
		}
		subscribers, ok := subMap[hex.EncodeToString(pubkey)]
		if !ok {
			continue
		}
		for _, sub := range subscribers {
			if sub.UserID == nil || sub.ID == nil {
				return fmt.Errorf("error expected userId and subId to be defined but got user: %v, sub: %v", sub.UserID, sub.ID)
			}
			if sub.LastEpoch != nil {
				lastSentEpoch := *sub.LastEpoch
				if lastSentEpoch >= epoch || epoch < sub.CreatedEpoch {
					continue
				}
			}
			log.Infof("creating %v notification for validator %v in epoch %v (dashboard: %v)", sub.EventName, event.Proposer, epoch, sub.DashboardId != nil)
			n := &ValidatorProposalNotification{
				NotificationBaseImpl: types.NotificationBaseImpl{
					SubscriptionID:     *sub.ID,
					UserID:             *sub.UserID,
					Epoch:              epoch,
					EventName:          sub.EventName,
					EventFilter:        hex.EncodeToString(pubkey),
					DashboardId:        sub.DashboardId,
					DashboardName:      sub.DashboardName,
					DashboardGroupId:   sub.DashboardGroupId,
					DashboardGroupName: sub.DashboardGroupName,
				},
				ValidatorIndex: event.Proposer,
				Status:         event.Status,
				Reward:         event.ExecRewardETH,
				Slot:           event.Slot,
				Block:          event.ExecBlock,
			}
			notificationsByUserID.AddNotification(n)
			metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()
		}
	}

	return nil
}

// collectAttestationAndOfflineValidatorNotifications collects notifications for missed attestations and offline validators
func collectAttestationAndOfflineValidatorNotifications(notificationsByUserID types.NotificationsPerUserId, epoch uint64) error {
	// Retrieve subscriptions for missed attestations
	subMapAttestationMissed, err := GetSubsForEventFilter(types.ValidatorMissedAttestationEventName, "", nil, nil)
	if err != nil {
		return fmt.Errorf("error getting subscriptions for missted attestations %w", err)
	}

	type dbResult struct {
		ValidatorIndex uint64 `db:"validatorindex"`
		Epoch          uint64 `db:"epoch"`
		Status         uint64 `db:"status"`
		EventFilter    []byte `db:"pubkey"`
	}

	// get attestations for all validators for the last 4 epochs
	// we need 4 epochs so that can detect the online / offline status of validators
	validators, err := db.GetValidatorIndices()
	if err != nil {
		return err
	}

	// this reads the submitted attestations for the last 4 epochs
	participationPerEpoch, err := db.GetValidatorAttestationHistoryForNotifications(epoch-3, epoch)
	if err != nil {
		return fmt.Errorf("error getting validator attestations from db %w", err)
	}

	log.Infof("retrieved validator attestation history data")

	events := make([]dbResult, 0)

	epochAttested := make(map[types.Epoch]uint64)
	epochTotal := make(map[types.Epoch]uint64)
	for currentEpoch, participation := range participationPerEpoch {
		for validatorIndex, participated := range participation {
			epochTotal[currentEpoch] = epochTotal[currentEpoch] + 1 // count the total attestations for each epoch

			if !participated {
				pubkey, err := GetPubkeyForIndex(uint64(validatorIndex))
				if err == nil {
					if currentEpoch != types.Epoch(epoch) || subMapAttestationMissed[hex.EncodeToString(pubkey)] == nil {
						continue
					}

					events = append(events, dbResult{
						ValidatorIndex: uint64(validatorIndex),
						Epoch:          uint64(currentEpoch),
						Status:         0,
						EventFilter:    pubkey,
					})
				} else {
					log.Error(err, "error retrieving pubkey for validator", 0, map[string]interface{}{"validatorIndex": validatorIndex})
				}
			} else {
				epochAttested[currentEpoch] = epochAttested[currentEpoch] + 1 // count the total attested attestation for each epoch (exlude missing)
			}
		}
	}

	// process missed attestation events
	for _, event := range events {
		subscribers, ok := subMapAttestationMissed[hex.EncodeToString(event.EventFilter)]
		if !ok {
			return fmt.Errorf("error event returned that does not exist: %x", event.EventFilter)
		}
		for _, sub := range subscribers {
			if sub.UserID == nil || sub.ID == nil {
				return fmt.Errorf("error expected userId and subId to be defined but got user: %v, sub: %v", sub.UserID, sub.ID)
			}
			if sub.LastEpoch != nil {
				lastSentEpoch := *sub.LastEpoch
				if lastSentEpoch >= event.Epoch || event.Epoch < sub.CreatedEpoch {
					// log.Infof("skipping creating %v for validator %v (lastSentEpoch: %v, createdEpoch: %v)", types.ValidatorMissedAttestationEventName, event.ValidatorIndex, lastSentEpoch, sub.CreatedEpoch)
					continue
				}
			}

			//log.Infof("creating %v notification for validator %v in epoch %v (dashboard: %v)", sub.EventName, event.ValidatorIndex, event.Epoch, sub.DashboardId != nil)
			n := &ValidatorAttestationNotification{
				NotificationBaseImpl: types.NotificationBaseImpl{
					SubscriptionID:     *sub.ID,
					UserID:             *sub.UserID,
					Epoch:              event.Epoch,
					EventName:          sub.EventName,
					EventFilter:        hex.EncodeToString(event.EventFilter),
					DashboardId:        sub.DashboardId,
					DashboardName:      sub.DashboardName,
					DashboardGroupId:   sub.DashboardGroupId,
					DashboardGroupName: sub.DashboardGroupName,
				},
				ValidatorIndex: event.ValidatorIndex,
				Status:         event.Status,
			}
			notificationsByUserID.AddNotification(n)
			metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()
		}
	}

	// detect online & offline validators
	type indexPubkeyPair struct {
		Index  uint64
		Pubkey []byte
	}
	var offlineValidators []*indexPubkeyPair
	var onlineValidators []*indexPubkeyPair

	epochNMinus1 := types.Epoch(epoch - 1)
	epochNMinus2 := types.Epoch(epoch - 2)
	epochNMinus3 := types.Epoch(epoch - 3)

	if epochTotal[types.Epoch(epoch)] == 0 {
		return fmt.Errorf("consistency error, did not retrieve attestation data for epoch %v", epoch)
	}
	if epochTotal[epochNMinus1] == 0 {
		return fmt.Errorf("consistency error, did not retrieve attestation data for epoch %v", epochNMinus1)
	}
	if epochTotal[epochNMinus2] == 0 {
		return fmt.Errorf("consistency error, did not retrieve attestation data for epoch %v", epochNMinus2)
	}
	if epochTotal[epochNMinus3] == 0 {
		return fmt.Errorf("consistency error, did not retrieve attestation data for epoch %v", epochNMinus3)
	}

	if epochAttested[types.Epoch(epoch)]*100/epochTotal[types.Epoch(epoch)] < 60 {
		return fmt.Errorf("consistency error, did receive more than 60%% of missed attestation in epoch %v (total: %v, attested: %v)", epoch, epochTotal[types.Epoch(epoch)], epochAttested[types.Epoch(epoch)])
	}
	if epochAttested[epochNMinus1]*100/epochTotal[epochNMinus1] < 60 {
		return fmt.Errorf("consistency error, did receive more than 60%% of missed attestation in epoch %v (total: %v, attested: %v)", epochNMinus1, epochTotal[epochNMinus1], epochAttested[epochNMinus1])
	}
	if epochAttested[epochNMinus2]*100/epochTotal[epochNMinus2] < 60 {
		return fmt.Errorf("consistency error, did receive more than 60%% of missed attestation in epoch %v (total: %v, attested: %v)", epochNMinus2, epochTotal[epochNMinus2], epochAttested[epochNMinus2])
	}
	if epochAttested[epochNMinus3]*100/epochTotal[epochNMinus3] < 60 {
		return fmt.Errorf("consistency error, did receive more than 60%% of missed attestation in epoch %v (total: %v, attested: %v)", epochNMinus3, epochTotal[epochNMinus3], epochAttested[epochNMinus3])
	}

	for _, validator := range validators {
		if participationPerEpoch[epochNMinus3][types.ValidatorIndex(validator)] && !participationPerEpoch[epochNMinus2][types.ValidatorIndex(validator)] && !participationPerEpoch[epochNMinus1][types.ValidatorIndex(validator)] && !participationPerEpoch[types.Epoch(epoch)][types.ValidatorIndex(validator)] {
			//log.Infof("validator %v detected as offline in epoch %v (did not attest since epoch %v)", validator, epoch, epochNMinus2)
			pubkey, err := GetPubkeyForIndex(validator)
			if err != nil {
				return err
			}
			offlineValidators = append(offlineValidators, &indexPubkeyPair{Index: validator, Pubkey: pubkey})
		}

		if !participationPerEpoch[epochNMinus3][types.ValidatorIndex(validator)] && !participationPerEpoch[epochNMinus2][types.ValidatorIndex(validator)] && !participationPerEpoch[epochNMinus1][types.ValidatorIndex(validator)] && participationPerEpoch[types.Epoch(epoch)][types.ValidatorIndex(validator)] {
			//log.Infof("validator %v detected as online in epoch %v (attested again in epoch %v)", validator, epoch, epoch)
			pubkey, err := GetPubkeyForIndex(validator)
			if err != nil {
				return err
			}
			onlineValidators = append(onlineValidators, &indexPubkeyPair{Index: validator, Pubkey: pubkey})
		}
	}

	offlineValidatorsLimit := 5000
	if utils.Config.Notifications.OfflineDetectionLimit != 0 {
		offlineValidatorsLimit = utils.Config.Notifications.OfflineDetectionLimit
	}

	onlineValidatorsLimit := 5000
	if utils.Config.Notifications.OnlineDetectionLimit != 0 {
		onlineValidatorsLimit = utils.Config.Notifications.OnlineDetectionLimit
	}

	if len(offlineValidators) > offlineValidatorsLimit {
		return fmt.Errorf("retrieved more than %v offline validators notifications: %v, exiting", offlineValidatorsLimit, len(offlineValidators))
	}

	if len(onlineValidators) > onlineValidatorsLimit {
		return fmt.Errorf("retrieved more than %v online validators notifications: %v, exiting", onlineValidatorsLimit, len(onlineValidators))
	}

	subMapOnlineOffline, err := GetSubsForEventFilter(types.ValidatorIsOfflineEventName, "", nil, nil)
	if err != nil {
		return fmt.Errorf("failed to get subs for %v: %v", types.ValidatorIsOfflineEventName, err)
	}

	for _, validator := range offlineValidators {
		t := hex.EncodeToString(validator.Pubkey)
		subs := subMapOnlineOffline[t]
		for _, sub := range subs {
			if sub.UserID == nil || sub.ID == nil {
				return fmt.Errorf("error expected userId and subId to be defined but got user: %v, sub: %v", sub.UserID, sub.ID)
			}
			log.Infof("new event: validator %v detected as offline since epoch %v", validator.Index, epoch)

			n := &ValidatorIsOfflineNotification{
				NotificationBaseImpl: types.NotificationBaseImpl{
					SubscriptionID:     *sub.ID,
					Epoch:              epoch,
					EventName:          sub.EventName,
					LatestState:        fmt.Sprint(epoch - 2), // first epoch the validator stopped attesting
					EventFilter:        hex.EncodeToString(validator.Pubkey),
					UserID:             *sub.UserID,
					DashboardId:        sub.DashboardId,
					DashboardName:      sub.DashboardName,
					DashboardGroupId:   sub.DashboardGroupId,
					DashboardGroupName: sub.DashboardGroupName,
				},
				ValidatorIndex: validator.Index,
			}

			notificationsByUserID.AddNotification(n)
			metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()
		}
	}

	for _, validator := range onlineValidators {
		t := hex.EncodeToString(validator.Pubkey)
		subs := subMapOnlineOffline[t]
		for _, sub := range subs {
			if sub.UserID == nil || sub.ID == nil {
				return fmt.Errorf("error expected userId and subId to be defined but got user: %v, sub: %v", sub.UserID, sub.ID)
			}

			log.Infof("new event: validator %v detected as online again at epoch %v", validator.Index, epoch)

			n := &ValidatorIsOnlineNotification{
				NotificationBaseImpl: types.NotificationBaseImpl{
					SubscriptionID:     *sub.ID,
					UserID:             *sub.UserID,
					Epoch:              epoch,
					EventName:          types.ValidatorIsOnlineEventName,
					EventFilter:        hex.EncodeToString(validator.Pubkey),
					LatestState:        "-",
					DashboardId:        sub.DashboardId,
					DashboardName:      sub.DashboardName,
					DashboardGroupId:   sub.DashboardGroupId,
					DashboardGroupName: sub.DashboardGroupName,
				},
				ValidatorIndex: validator.Index,
			}

			notificationsByUserID.AddNotification(n)
			metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()
		}
	}

	// subMapGroupOnlineOffline, err := GetSubsForEventFilter(types.ValidatorGroupIsOfflineEventName, "", nil, nil, validatorDashboardConfig)
	// if err != nil {
	// 	return fmt.Errorf("failed to get subs for %v: %v", types.ValidatorGroupIsOfflineEventName, err)
	// }
	// type groupSummary struct {
	// 	Online       int
	// 	Offline      int
	// 	Subscription types.Subscription
	// }
	// dashboardGroups := make(map[types.DashboardId]map[types.DashboardGroupId]*groupSummary)

	// for _, validator := range offlineValidators {
	// 	t := hex.EncodeToString(validator.Pubkey)
	// 	subs := subMapGroupOnlineOffline[t]
	// 	for _, sub := range subs {
	// 		if sub.UserID == nil || sub.ID == nil {
	// 			return fmt.Errorf("error expected userId and subId to be defined but got user: %v, sub: %v", sub.UserID, sub.ID)
	// 		}

	// 		if sub.DashboardId == nil || sub.DashboardGroupId == nil {
	// 			continue
	// 		}
	// 		dashboardId := types.DashboardId(*sub.DashboardId)
	// 		dashboardGroupId := types.DashboardGroupId(*sub.DashboardGroupId)

	// 		if dashboardGroups[dashboardId] == nil {
	// 			dashboardGroups[dashboardId] = make(map[types.DashboardGroupId]*groupSummary)
	// 			dashboardGroups[dashboardId][dashboardGroupId] = &groupSummary{
	// 				Subscription: sub,
	// 			}
	// 		}
	// 		dashboardGroups[dashboardId][dashboardGroupId].Offline++
	// 	}
	// }
	// for _, validator := range onlineValidators {
	// 	t := hex.EncodeToString(validator.Pubkey)
	// 	subs := subMapGroupOnlineOffline[t]
	// 	for _, sub := range subs {
	// 		if sub.UserID == nil || sub.ID == nil {
	// 			return fmt.Errorf("error expected userId and subId to be defined but got user: %v, sub: %v", sub.UserID, sub.ID)
	// 		}

	// 		if sub.DashboardId == nil || sub.DashboardGroupId == nil {
	// 			continue
	// 		}
	// 		dashboardId := types.DashboardId(*sub.DashboardId)
	// 		dashboardGroupId := types.DashboardGroupId(*sub.DashboardGroupId)

	// 		if dashboardGroups[dashboardId] == nil {
	// 			dashboardGroups[dashboardId] = make(map[types.DashboardGroupId]*groupSummary)
	// 			dashboardGroups[dashboardId][dashboardGroupId] = &groupSummary{
	// 				Subscription: sub,
	// 			}
	// 		}
	// 		dashboardGroups[dashboardId][dashboardGroupId].Online++
	// 	}
	// }

	// for dashboardId, groups := range dashboardGroups {
	// 	for groupId, data := range groups {
	// 		totalGroupValidators := len(validatorDashboardConfig.DashboardsById[dashboardId].Groups[groupId].Validators)

	// 		offlinePercentage := float64(data.Offline) / float64(totalGroupValidators) * float64(100)

	// 		if offlinePercentage >= data.Subscription.EventThreshold {
	// 			log.Infof("new event: group %v detected as offline in epoch %v (offline: %v, total: %v, percentage: %v, threshold: %v)", groupId, epoch, data.Offline, totalGroupValidators, offlinePercentage, data.Subscription.EventThreshold)
	// 		}
	// 	}
	// }
	return nil
}

func collectValidatorGotSlashedNotifications(notificationsByUserID types.NotificationsPerUserId, epoch uint64) error {
	dbResult, err := db.GetValidatorsGotSlashed(epoch)
	if err != nil {
		return fmt.Errorf("error getting slashed validators from database, err: %w", err)
	}
	pubkeyToSlashingInfoMap := make(map[string]*types.SlashingInfo)
	for _, event := range dbResult {
		pubkeyStr := hex.EncodeToString(event.SlashedValidatorPubkey)
		pubkeyToSlashingInfoMap[pubkeyStr] = event
	}

	subscribedUsers, err := GetSubsForEventFilter(types.ValidatorGotSlashedEventName, "", nil, nil)
	if err != nil {
		return fmt.Errorf("failed to get subs for %v: %v", types.ValidatorGotSlashedEventName, err)
	}

	for _, subs := range subscribedUsers {
		for _, sub := range subs {
			event := pubkeyToSlashingInfoMap[sub.EventFilter]
			if event == nil { // pubkey has not been slashed
				//log.Error(fmt.Errorf("error retrieving slashing info for public key %s", sub.EventFilter), "", 0)
				continue
			}
			log.Infof("creating %v notification for validator %v in epoch %v", event.Reason, sub.EventFilter, epoch)

			n := &ValidatorGotSlashedNotification{
				NotificationBaseImpl: types.NotificationBaseImpl{
					SubscriptionID:     *sub.ID,
					UserID:             *sub.UserID,
					Epoch:              epoch,
					EventFilter:        sub.EventFilter,
					EventName:          sub.EventName,
					DashboardId:        sub.DashboardId,
					DashboardName:      sub.DashboardName,
					DashboardGroupId:   sub.DashboardGroupId,
					DashboardGroupName: sub.DashboardGroupName,
				},
				Slasher:        event.SlasherIndex,
				Reason:         event.Reason,
				ValidatorIndex: event.SlashedValidatorIndex,
			}
			notificationsByUserID.AddNotification(n)
			metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()
		}
	}

	return nil
}

// collectWithdrawalNotifications collects all notifications validator withdrawals
func collectWithdrawalNotifications(notificationsByUserID types.NotificationsPerUserId, epoch uint64) error {
	// get all users that are subscribed to this event (scale: a few thousand rows depending on how many users we have)
	subMap, err := GetSubsForEventFilter(types.ValidatorReceivedWithdrawalEventName, "", nil, nil)
	if err != nil {
		return fmt.Errorf("error getting subscriptions for missed attestations %w", err)
	}

	// get all the withdrawal events for a specific epoch. Will be at most X per slot (currently 16 on mainnet, which is 32 * 16 per epoch; 512 rows).
	events, err := db.GetEpochWithdrawals(epoch)
	if err != nil {
		return fmt.Errorf("error getting withdrawals from database, err: %w", err)
	}

	log.Infof("retrieved %v events", len(events))
	for _, event := range events {
		subscribers, ok := subMap[hex.EncodeToString(event.Pubkey)]
		if ok {
			for _, sub := range subscribers {
				if sub.UserID == nil || sub.ID == nil {
					return fmt.Errorf("error expected userId and subId to be defined but got user: %v, sub: %v", sub.UserID, sub.ID)
				}
				if sub.LastEpoch != nil {
					lastSentEpoch := *sub.LastEpoch
					if lastSentEpoch >= epoch || epoch < sub.CreatedEpoch {
						continue
					}
				}
				log.Infof("creating %v notification for validator %v in epoch %v", types.ValidatorReceivedWithdrawalEventName, event.ValidatorIndex, epoch)
				n := &ValidatorWithdrawalNotification{
					NotificationBaseImpl: types.NotificationBaseImpl{
						SubscriptionID:     *sub.ID,
						UserID:             *sub.UserID,
						EventFilter:        hex.EncodeToString(event.Pubkey),
						EventName:          sub.EventName,
						DashboardId:        sub.DashboardId,
						DashboardName:      sub.DashboardName,
						DashboardGroupId:   sub.DashboardGroupId,
						DashboardGroupName: sub.DashboardGroupName,
						Epoch:              epoch,
					},
					ValidatorIndex: event.ValidatorIndex,
					Epoch:          epoch,
					Slot:           event.Slot,
					Amount:         event.Amount,
					Address:        event.Address,
				}
				notificationsByUserID.AddNotification(n)
				metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()
			}
		}
	}

	return nil
}

func collectEthClientNotifications(notificationsByUserID types.NotificationsPerUserId) error {
	updatedClients := ethclients.GetUpdatedClients() //only check if there are new updates
	for _, client := range updatedClients {
		// err := db.FrontendWriterDB.Select(&dbResult, `
		// 	SELECT us.id, us.user_id, us.created_epoch, us.event_filter, ENCODE(us.unsubscribe_hash, 'hex') AS unsubscribe_hash
		// 	FROM users_subscriptions AS us
		// 	WHERE
		// 		us.event_name=$1
		// 	AND
		// 		us.event_filter=$2
		// 	AND
		// 		((us.last_sent_ts <= NOW() - INTERVAL '2 DAY' AND TO_TIMESTAMP($3) > us.last_sent_ts) OR us.last_sent_ts IS NULL)
		// 	`,
		// 	eventName, strings.ToLower(client.Name), client.Date.Unix()) // was last notification sent 2 days ago for this client

		dbResult, err := GetSubsForEventFilter(
			types.EthClientUpdateEventName,
			"((last_sent_ts <= NOW() - INTERVAL '2 DAY' AND TO_TIMESTAMP(?) > last_sent_ts) OR last_sent_ts IS NULL)",
			[]interface{}{client.Date.Unix()},
			[]string{strings.ToLower(client.Name)})
		if err != nil {
			return err
		}

		for _, subs := range dbResult {
			for _, sub := range subs {
				n := &EthClientNotification{
					NotificationBaseImpl: types.NotificationBaseImpl{
						SubscriptionID:     *sub.ID,
						UserID:             *sub.UserID,
						Epoch:              sub.CreatedEpoch,
						EventFilter:        sub.EventFilter,
						EventName:          sub.EventName,
						DashboardId:        sub.DashboardId,
						DashboardName:      sub.DashboardName,
						DashboardGroupId:   sub.DashboardGroupId,
						DashboardGroupName: sub.DashboardGroupName,
					},
					EthClient: client.Name,
					Url:       client.Url,
					Version:   client.Version,
				}
				notificationsByUserID.AddNotification(n)
				metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()
			}
		}
	}
	return nil
}

func collectMonitoringMachineOffline(notificationsByUserID types.NotificationsPerUserId, epoch uint64) error {
	nowTs := time.Now().Unix()
	return collectMonitoringMachine(notificationsByUserID, types.MonitoringMachineOfflineEventName, 120,
		// notify condition
		func(subscribeData *types.Subscription, machineData *types.MachineMetricSystemUser) bool {
			if machineData.CurrentDataInsertTs < nowTs-10*60 && machineData.CurrentDataInsertTs > nowTs-90*60 {
				return true
			}
			return false
		},
		epoch,
	)
}

func isMachineDataRecent(machineData *types.MachineMetricSystemUser) bool {
	nowTs := time.Now().Unix()
	return machineData.CurrentDataInsertTs >= nowTs-60*60
}

func collectMonitoringMachineDiskAlmostFull(notificationsByUserID types.NotificationsPerUserId, epoch uint64) error {
	return collectMonitoringMachine(notificationsByUserID, types.MonitoringMachineDiskAlmostFullEventName, 750,
		// notify condition
		func(subscribeData *types.Subscription, machineData *types.MachineMetricSystemUser) bool {
			if !isMachineDataRecent(machineData) {
				return false
			}

			percentFree := float64(machineData.CurrentData.DiskNodeBytesFree) / float64(machineData.CurrentData.DiskNodeBytesTotal+1)
			return percentFree < subscribeData.EventThreshold
		},
		epoch,
	)
}

func collectMonitoringMachineCPULoad(notificationsByUserID types.NotificationsPerUserId, epoch uint64) error {
	return collectMonitoringMachine(notificationsByUserID, types.MonitoringMachineCpuLoadEventName, 10,
		// notify condition
		func(subscribeData *types.Subscription, machineData *types.MachineMetricSystemUser) bool {
			if !isMachineDataRecent(machineData) {
				return false
			}

			if machineData.FiveMinuteOldData == nil { // no compare data found (5 min old data)
				return false
			}

			idle := float64(machineData.CurrentData.CpuNodeIdleSecondsTotal) - float64(machineData.FiveMinuteOldData.CpuNodeIdleSecondsTotal)
			total := float64(machineData.CurrentData.CpuNodeSystemSecondsTotal) - float64(machineData.FiveMinuteOldData.CpuNodeSystemSecondsTotal)
			percentLoad := float64(1) - (idle / total)

			return percentLoad > subscribeData.EventThreshold
		},
		epoch,
	)
}

func collectMonitoringMachineMemoryUsage(notificationsByUserID types.NotificationsPerUserId, epoch uint64) error {
	return collectMonitoringMachine(notificationsByUserID, types.MonitoringMachineMemoryUsageEventName, 10,
		// notify condition
		func(subscribeData *types.Subscription, machineData *types.MachineMetricSystemUser) bool {
			if !isMachineDataRecent(machineData) {
				return false
			}

			memFree := float64(machineData.CurrentData.MemoryNodeBytesFree) + float64(machineData.CurrentData.MemoryNodeBytesCached) + float64(machineData.CurrentData.MemoryNodeBytesBuffers)
			memTotal := float64(machineData.CurrentData.MemoryNodeBytesTotal)
			memUsage := float64(1) - (memFree / memTotal)

			return memUsage > subscribeData.EventThreshold
		},
		epoch,
	)
}

var isFirstNotificationCheck = true

func collectMonitoringMachine(
	notificationsByUserID types.NotificationsPerUserId,
	eventName types.EventName,
	epochWaitInBetween int,
	notifyConditionFulfilled func(subscribeData *types.Subscription, machineData *types.MachineMetricSystemUser) bool,
	epoch uint64,
) error {
	// event_filter == machine name

	dbResult, err := GetSubsForEventFilter(
		eventName,
		"(created_epoch <= ? AND (last_sent_epoch < ? OR last_sent_epoch IS NULL))",
		[]interface{}{epoch, int64(epoch) - int64(epochWaitInBetween)},
		nil,
	)

	// TODO: clarify why we need grouping here?!
	// err := db.FrontendWriterDB.Select(&allSubscribed,
	// 	`SELECT
	// 		us.user_id,
	// 		max(us.id) AS id,
	// 		ENCODE((array_agg(us.unsubscribe_hash))[1], 'hex') AS unsubscribe_hash,
	// 		event_filter,
	// 		COALESCE(event_threshold, 0) AS event_threshold
	// 	FROM users_subscriptions us
	// 	WHERE us.event_name = $1 AND us.created_epoch <= $2
	// 	AND (us.last_sent_epoch < ($2 - $3) OR us.last_sent_epoch IS NULL)
	// 	group by us.user_id, event_filter, event_threshold`,
	// 	eventName, epoch, epochWaitInBetween)
	if err != nil {
		return err
	}

	rowKeys := gcp_bigtable.RowList{}
	totalSubscribed := 0
	for _, data := range dbResult {
		for _, sub := range data {
			rowKeys = append(rowKeys, db.BigtableClient.GetMachineRowKey(*sub.UserID, "system", sub.EventFilter))
			totalSubscribed++
		}
	}

	machineDataOfSubscribed, err := db.BigtableClient.GetMachineMetricsForNotifications(rowKeys)
	if err != nil {
		return err
	}

	var result []*types.Subscription
	for _, data := range dbResult {
		for _, sub := range data {
			localData := sub // Create a local copy of the data variable
			machineMap, found := machineDataOfSubscribed[*localData.UserID]
			if !found {
				continue
			}
			currentMachineData, found := machineMap[localData.EventFilter]
			if !found {
				continue
			}

			//logrus.Infof("currentMachineData %v | %v | %v | %v", currentMachine.CurrentDataInsertTs, currentMachine.CompareDataInsertTs, currentMachine.UserID, currentMachine.Machine)
			if notifyConditionFulfilled(localData, currentMachineData) {
				result = append(result, localData)
			}
		}
	}

	subThreshold := uint64(10)
	if utils.Config.Notifications.MachineEventThreshold != 0 {
		subThreshold = utils.Config.Notifications.MachineEventThreshold
	}

	subFirstRatioThreshold := 0.3
	if utils.Config.Notifications.MachineEventFirstRatioThreshold != 0 {
		subFirstRatioThreshold = utils.Config.Notifications.MachineEventFirstRatioThreshold
	}

	subSecondRatioThreshold := 0.9
	if utils.Config.Notifications.MachineEventSecondRatioThreshold != 0 {
		subSecondRatioThreshold = utils.Config.Notifications.MachineEventSecondRatioThreshold
	}

	var subScriptionCount uint64
	err = db.FrontendWriterDB.Get(&subScriptionCount,
		`SELECT
			COUNT(DISTINCT user_id)
			FROM users_subscriptions
			WHERE event_name = $1`,
		eventName)
	if err != nil {
		return err
	}

	// If there are too few users subscribed to this event, we always send the notifications
	if subScriptionCount >= subThreshold {
		subRatioThreshold := subSecondRatioThreshold
		// For the machine offline check we do a low threshold check first and the next time a high threshold check
		if isFirstNotificationCheck && eventName == types.MonitoringMachineOfflineEventName {
			subRatioThreshold = subFirstRatioThreshold
			isFirstNotificationCheck = false
		}
		if float64(len(result))/float64(totalSubscribed) >= subRatioThreshold {
			log.Error(nil, fmt.Errorf("error too many users would be notified concerning: %v", eventName), 0)
			return nil
		}
	}

	for _, r := range result {
		n := &MonitorMachineNotification{
			NotificationBaseImpl: types.NotificationBaseImpl{
				SubscriptionID:     *r.ID,
				UserID:             *r.UserID,
				EventName:          r.EventName,
				Epoch:              epoch,
				EventFilter:        r.EventFilter,
				DashboardId:        r.DashboardId,
				DashboardName:      r.DashboardName,
				DashboardGroupId:   r.DashboardGroupId,
				DashboardGroupName: r.DashboardGroupName,
			},
			MachineName:    r.EventFilter,
			EventThreshold: r.EventThreshold,
		}
		//logrus.Infof("notify %v %v", eventName, n)
		notificationsByUserID.AddNotification(n)
		metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()
	}

	if eventName == types.MonitoringMachineOfflineEventName {
		// Notifications will be sent, reset the flag
		isFirstNotificationCheck = true
	}

	return nil
}

func collectTaxReportNotificationNotifications(notificationsByUserID types.NotificationsPerUserId) error {
	lastStatsDay, err := cache.LatestExportedStatisticDay.GetOrDefault(db.GetLastExportedStatisticDay)

	if err != nil {
		return err
	}
	//Check that the last day of the month is already exported
	tNow := time.Now()
	firstDayOfMonth := time.Date(tNow.Year(), tNow.Month(), 1, 0, 0, 0, 0, time.UTC)
	if utils.TimeToDay(uint64(firstDayOfMonth.Unix())) > lastStatsDay {
		return nil
	}

	// err = db.FrontendWriterDB.Select(&dbResult, `
	// 		SELECT us.id, us.user_id, us.created_epoch, us.event_filter, ENCODE(us.unsubscribe_hash, 'hex') AS unsubscribe_hash
	// 		FROM users_subscriptions AS us
	// 		WHERE us.event_name=$1 AND (us.last_sent_ts < $2 OR (us.last_sent_ts IS NULL AND us.created_ts < $2));
	// 		`,
	// 	name, firstDayOfMonth)

	dbResults, err := GetSubsForEventFilter(
		types.TaxReportEventName,
		"(last_sent_ts < ? OR (last_sent_ts IS NULL AND created_ts < ?))",
		[]interface{}{firstDayOfMonth, firstDayOfMonth},
		nil,
	)
	if err != nil {
		return err
	}

	for _, subs := range dbResults {
		for _, sub := range subs {
			n := &TaxReportNotification{
				NotificationBaseImpl: types.NotificationBaseImpl{
					SubscriptionID:     *sub.ID,
					UserID:             *sub.UserID,
					Epoch:              sub.CreatedEpoch,
					EventFilter:        sub.EventFilter,
					EventName:          sub.EventName,
					DashboardId:        sub.DashboardId,
					DashboardName:      sub.DashboardName,
					DashboardGroupId:   sub.DashboardGroupId,
					DashboardGroupName: sub.DashboardGroupName,
				},
			}
			notificationsByUserID.AddNotification(n)
			metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()
		}
	}

	return nil
}

func collectNetworkNotifications(notificationsByUserID types.NotificationsPerUserId) error {
	count := 0
	err := db.WriterDb.Get(&count, `
		SELECT count(ts) FROM network_liveness WHERE (headepoch-finalizedepoch) > 3 AND ts > now() - interval '60 minutes';
	`)

	if err != nil {
		return err
	}

	if count > 0 {
		// err := db.FrontendWriterDB.Select(&dbResult, `
		// 	SELECT us.id, us.user_id, us.created_epoch, us.event_filter, ENCODE(us.unsubscribe_hash, 'hex') AS unsubscribe_hash
		// 	FROM users_subscriptions AS us
		// 	WHERE us.event_name=$1 AND (us.last_sent_ts <= NOW() - INTERVAL '1 hour' OR us.last_sent_ts IS NULL);
		// 	`,
		// 	utils.GetNetwork()+":"+string(eventName))

		dbResult, err := GetSubsForEventFilter(
			types.NetworkLivenessIncreasedEventName,
			"(last_sent_ts <= NOW() - INTERVAL '1 hour' OR last_sent_ts IS NULL)",
			nil,
			nil,
		)
		if err != nil {
			return err
		}

		for _, subs := range dbResult {
			for _, sub := range subs {
				n := &NetworkNotification{
					NotificationBaseImpl: types.NotificationBaseImpl{
						SubscriptionID:     *sub.ID,
						UserID:             *sub.UserID,
						Epoch:              sub.CreatedEpoch,
						EventFilter:        sub.EventFilter,
						EventName:          sub.EventName,
						DashboardId:        sub.DashboardId,
						DashboardName:      sub.DashboardName,
						DashboardGroupId:   sub.DashboardGroupId,
						DashboardGroupName: sub.DashboardGroupName,
					},
				}

				notificationsByUserID.AddNotification(n)
				metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()
			}
		}
	}

	return nil
}

func collectRocketpoolComissionNotifications(notificationsByUserID types.NotificationsPerUserId) error {
	fee := 0.0
	err := db.WriterDb.Get(&fee, `
		select current_node_fee from rocketpool_network_stats order by id desc LIMIT 1;
	`)

	if err != nil {
		return err
	}

	if fee > 0 {
		// err := db.FrontendWriterDB.Select(&dbResult, `
		// 	SELECT us.id, us.user_id, us.created_epoch, us.event_filter, ENCODE(us.unsubscribe_hash, 'hex') AS unsubscribe_hash
		// 	FROM users_subscriptions AS us
		// 	WHERE us.event_name=$1 AND (us.last_sent_ts <= NOW() - INTERVAL '8 hours' OR us.last_sent_ts IS NULL) AND (us.event_threshold <= $2 OR (us.event_threshold < 0 AND us.event_threshold * -1 >= $2));
		// 	`,
		// 	utils.GetNetwork()+":"+string(eventName), fee)

		dbResult, err := GetSubsForEventFilter(
			types.RocketpoolCommissionThresholdEventName,
			"(last_sent_ts <= NOW() - INTERVAL '8 hours' OR last_sent_ts IS NULL) AND (event_threshold <= ? OR (event_threshold < 0 AND event_threshold * -1 >= ?))",
			[]interface{}{fee, fee},
			nil,
		)
		if err != nil {
			return err
		}

		for _, subs := range dbResult {
			for _, sub := range subs {
				n := &RocketpoolNotification{
					NotificationBaseImpl: types.NotificationBaseImpl{
						SubscriptionID:     *sub.ID,
						UserID:             *sub.UserID,
						Epoch:              sub.CreatedEpoch,
						EventFilter:        sub.EventFilter,
						EventName:          sub.EventName,
						DashboardId:        sub.DashboardId,
						DashboardName:      sub.DashboardName,
						DashboardGroupId:   sub.DashboardGroupId,
						DashboardGroupName: sub.DashboardGroupName,
					},
					ExtraData: strconv.FormatInt(int64(fee*100), 10) + "%",
				}

				notificationsByUserID.AddNotification(n)
				metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()
			}
		}
	}

	return nil
}

func collectRocketpoolRewardClaimRoundNotifications(notificationsByUserID types.NotificationsPerUserId) error {
	var ts int64
	err := db.WriterDb.Get(&ts, `
		select date_part('epoch', claim_interval_time_start)::int from rocketpool_network_stats order by id desc LIMIT 1;
	`)

	if err != nil {
		return err
	}

	if ts+3*60*60 > time.Now().Unix() {
		// var dbResult []*types.Subscription

		// err := db.FrontendWriterDB.Select(&dbResult, `
		// 	SELECT us.id, us.user_id, us.created_epoch, us.event_filter, ENCODE(us.unsubscribe_hash, 'hex') AS unsubscribe_hash
		// 	FROM users_subscriptions AS us
		// 	WHERE us.event_name=$1 AND (us.last_sent_ts <= NOW() - INTERVAL '5 hours' OR us.last_sent_ts IS NULL);
		// 	`,
		// 	utils.GetNetwork()+":"+string(eventName))

		dbResult, err := GetSubsForEventFilter(
			types.RocketpoolNewClaimRoundStartedEventName,
			"(last_sent_ts <= NOW() - INTERVAL '5 hours' OR last_sent_ts IS NULL)",
			nil,
			nil,
		)
		if err != nil {
			return err
		}

		for _, subs := range dbResult {
			for _, sub := range subs {
				n := &RocketpoolNotification{
					NotificationBaseImpl: types.NotificationBaseImpl{
						SubscriptionID:     *sub.ID,
						UserID:             *sub.UserID,
						Epoch:              sub.CreatedEpoch,
						EventFilter:        sub.EventFilter,
						EventName:          sub.EventName,
						DashboardId:        sub.DashboardId,
						DashboardName:      sub.DashboardName,
						DashboardGroupId:   sub.DashboardGroupId,
						DashboardGroupName: sub.DashboardGroupName,
					},
				}

				notificationsByUserID.AddNotification(n)
				metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()
			}
		}
	}

	return nil
}

func collectRocketpoolRPLCollateralNotifications(notificationsByUserID types.NotificationsPerUserId, eventName types.EventName, epoch uint64) error {
	subMap, err := GetSubsForEventFilter(
		eventName,
		"(last_sent_ts <= NOW() - INTERVAL '24 hours' OR last_sent_ts IS NULL)", // send out this notification type only once per day
		nil,
		nil)
	if err != nil {
		return fmt.Errorf("error getting subscriptions for RocketpoolRPLCollateral %w", err)
	}

	type dbResult struct {
		Address     []byte
		RPLStake    BigFloat `db:"rpl_stake"`
		RPLStakeMin BigFloat `db:"min_rpl_stake"`
		RPLStakeMax BigFloat `db:"max_rpl_stake"`
	}

	// filter nodes with no minipools (anymore) because they have min/max stake of 0
	// TODO properly remove notification entry from db
	stakeInfoPerNode := make([]dbResult, 0)
	batchSize := 5000
	keys := make([][]byte, 0, batchSize)
	for pubkey := range subMap {
		b, err := hex.DecodeString(pubkey)
		if err != nil {
			log.Error(err, fmt.Sprintf("error decoding pubkey %s", pubkey), 0)
			continue
		}
		keys = append(keys, b)

		if len(keys) > batchSize {
			var partial []dbResult

			err = db.WriterDb.Select(&partial, `
			SELECT address, rpl_stake, min_rpl_stake, max_rpl_stake
			FROM rocketpool_nodes
			WHERE address = ANY($1) AND min_rpl_stake != 0 AND max_rpl_stake != 0`, pq.ByteaArray(keys))
			if err != nil {
				return err
			}
			stakeInfoPerNode = append(stakeInfoPerNode, partial...)
			keys = make([][]byte, 0, batchSize)
		}
	}
	if len(keys) > 0 {
		var partial []dbResult

		// filter nodes with no minipools (anymore) because they have min/max stake of 0
		// TODO properly remove notification entry from db
		err = db.WriterDb.Select(&partial, `
		SELECT address, rpl_stake, min_rpl_stake, max_rpl_stake
		FROM rocketpool_nodes
		WHERE address = ANY($1) AND min_rpl_stake != 0 AND max_rpl_stake != 0`, pq.ByteaArray(keys))
		if err != nil {
			return err
		}
		stakeInfoPerNode = append(stakeInfoPerNode, partial...)
	}

	// factor in network-wide min/max collat ratio. Since LEB8 they are not directly correlated anymore (ratio of bonded to borrowed ETH), so we need either min or max
	// however this is dynamic and might be changed in the future; Should extend rocketpool_network_stats to include min/max collateral values!
	minRPLCollatRatio := bigFloat(0.1) // bigFloat it to save some memory re-allocations
	maxRPLCollatRatio := bigFloat(1.5)
	// temporary helper (modifying values in dbResult directly would be bad style)
	nodeCollatRatioHelper := bigFloat(0)

	for _, r := range stakeInfoPerNode {
		subs, ok := subMap[hex.EncodeToString(r.Address)]
		if !ok {
			continue
		}
		sub := subs[0] // RPL min/max collateral notifications are always unique per user
		var alertConditionMet bool

		// according to app logic, sub.EventThreshold can be +- [0.9 to 1.5] for CollateralMax after manually changed by the user
		// this corresponds to a collateral range of 140% to 200% currently shown in the app UI; so +- 0.5 allows us to compare to the actual collat ratio
		// for CollateralMin it  can be 1.0 to 4.0 if manually changed, to represent 10% to 40%
		// 0 in both cases if not modified
		var threshold float64 = sub.EventThreshold
		if threshold == 0 {
			threshold = 1.0 // default case
		}
		inverse := false
		if eventName == types.RocketpoolCollateralMaxReachedEventName {
			if threshold < 0 {
				threshold *= -1
			} else {
				inverse = true
			}
			threshold += 0.5

			// 100% (of bonded eth)
			nodeCollatRatioHelper.Quo(r.RPLStakeMax.bigFloat(), maxRPLCollatRatio)
		} else {
			threshold /= 10.0

			// 100% (of borrowed eth)
			nodeCollatRatioHelper.Quo(r.RPLStakeMin.bigFloat(), minRPLCollatRatio)
		}

		nodeCollatRatio, _ := nodeCollatRatioHelper.Quo(r.RPLStake.bigFloat(), nodeCollatRatioHelper).Float64()

		alertConditionMet = nodeCollatRatio <= threshold
		if inverse {
			// handle special case for max collateral: notify if *above* selected amount
			alertConditionMet = !alertConditionMet
		}

		if !alertConditionMet {
			continue
		}

		if sub.LastEpoch != nil {
			lastSentEpoch := *sub.LastEpoch
			if lastSentEpoch >= epoch-225 || epoch < sub.CreatedEpoch {
				continue
			}
		}

		n := &RocketpoolNotification{
			NotificationBaseImpl: types.NotificationBaseImpl{
				SubscriptionID:     *sub.ID,
				UserID:             *sub.UserID,
				Epoch:              epoch,
				EventFilter:        sub.EventFilter,
				EventName:          sub.EventName,
				DashboardId:        sub.DashboardId,
				DashboardName:      sub.DashboardName,
				DashboardGroupId:   sub.DashboardGroupId,
				DashboardGroupName: sub.DashboardGroupName,
			},
			ExtraData: strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", threshold*100), "0"), "."),
		}

		notificationsByUserID.AddNotification(n)
		metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()
	}

	return nil
}

func collectSyncCommitteeNotifications(notificationsByUserID types.NotificationsPerUserId, epoch uint64) error {
	slotsPerSyncCommittee := utils.SlotsPerSyncCommittee()
	currentPeriod := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch / slotsPerSyncCommittee
	nextPeriod := currentPeriod + 1

	var validators []struct {
		PubKey string `db:"pubkey"`
		Index  uint64 `db:"validatorindex"`
	}
	err := db.WriterDb.Select(&validators, `SELECT ENCODE(pubkey, 'hex') AS pubkey, validators.validatorindex FROM sync_committees LEFT JOIN validators ON validators.validatorindex = sync_committees.validatorindex WHERE period = $1`, nextPeriod)

	if err != nil {
		return err
	}

	if len(validators) <= 0 {
		return nil
	}

	var mapping map[string]uint64 = make(map[string]uint64)
	for _, val := range validators {
		mapping[val.PubKey] = val.Index
	}

	dbResult, err := GetSubsForEventFilter(types.SyncCommitteeSoonEventName, "(last_sent_ts <= NOW() - INTERVAL '26 hours' OR last_sent_ts IS NULL)", nil, nil)

	if err != nil {
		return err
	}

	for pubkey := range mapping {
		subs, ok := dbResult[pubkey]
		if ok {
			for _, sub := range subs {
				n := &SyncCommitteeSoonNotification{
					NotificationBaseImpl: types.NotificationBaseImpl{
						SubscriptionID:     *sub.ID,
						UserID:             *sub.UserID,
						Epoch:              epoch,
						EventFilter:        sub.EventFilter,
						EventName:          sub.EventName,
						DashboardId:        sub.DashboardId,
						DashboardName:      sub.DashboardName,
						DashboardGroupId:   sub.DashboardGroupId,
						DashboardGroupName: sub.DashboardGroupName,
					},
					ValidatorIndex: mapping[sub.EventFilter],
					StartEpoch:     nextPeriod * utils.Config.Chain.ClConfig.EpochsPerSyncCommitteePeriod,
					EndEpoch:       (nextPeriod + 1) * utils.Config.Chain.ClConfig.EpochsPerSyncCommitteePeriod,
				}
				notificationsByUserID.AddNotification(n)
				metrics.NotificationsCollected.WithLabelValues(string(n.GetEventName())).Inc()
			}
		}
	}

	return nil
}

func getSyncCommitteeSoonInfo(format types.NotificationFormat, ns map[types.EventFilter]types.Notification) string {
	validators := []uint64{}
	var startEpoch, endEpoch uint64
	var inTime time.Duration

	i := 0
	for _, n := range ns {
		n, ok := n.(*SyncCommitteeSoonNotification)
		if !ok {
			log.Error(nil, "Sync committee notification not of type syncCommitteeSoonNotification", 0)
			return ""
		}

		validators = append(validators, n.ValidatorIndex)
		if i == 0 {
			// startEpoch, endEpoch and inTime must be the same for all validators
			startEpoch = n.StartEpoch
			endEpoch = n.EndEpoch

			inTime = time.Until(utils.EpochToTime(startEpoch)).Round(time.Second)
		}
		i++
	}

	if len(validators) > 0 {
		startEpochFormatted := formatEpochLink(format, startEpoch)
		endEpochFormatted := formatEpochLink(format, endEpoch)
		validatorsInfo := ""
		if len(validators) == 1 {
			vali := formatValidatorLink(format, validators[0])
			validatorsInfo = fmt.Sprintf(`Your validator %s has been elected to be part of the next sync committee.`, vali)
		} else {
			validatorsText := ""
			for i, validator := range validators {
				vali := formatValidatorLink(format, validator)
				if i < len(validators)-1 {
					validatorsText += fmt.Sprintf("%v, ", vali)
				} else {
					validatorsText += fmt.Sprintf("and %v", vali)
				}
			}
			validatorsInfo = fmt.Sprintf(`Your validators %s have been elected to be part of the next sync committee.`, validatorsText)
		}
		return fmt.Sprintf(`%s The additional duties start at epoch %v, which is in %v and will last for about a day until epoch %v.`, validatorsInfo, startEpochFormatted, inTime, endEpochFormatted)
	}

	return ""
}
