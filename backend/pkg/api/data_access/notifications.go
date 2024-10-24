package dataaccess

import (
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/gob"
	"fmt"
	"io"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/params"
	"github.com/go-redis/redis/v8"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/notification"
	n "github.com/gobitfly/beaconchain/pkg/notification"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
)

type NotificationsRepository interface {
	GetNotificationOverview(ctx context.Context, userId uint64) (*t.NotificationOverviewData, error)

	GetDashboardNotifications(ctx context.Context, userId uint64, chainIds []uint64, cursor string, colSort t.Sort[enums.NotificationDashboardsColumn], search string, limit uint64) ([]t.NotificationDashboardsTableRow, *t.Paging, error)
	// depending on how notifications are implemented, we may need to use something other than `notificationId` for identifying the notification
	GetValidatorDashboardNotificationDetails(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, epoch uint64, search string) (*t.NotificationValidatorDashboardDetail, error)
	GetAccountDashboardNotificationDetails(ctx context.Context, dashboardId uint64, groupId uint64, epoch uint64, search string) (*t.NotificationAccountDashboardDetail, error)

	GetMachineNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationMachinesColumn], search string, limit uint64) ([]t.NotificationMachinesTableRow, *t.Paging, error)
	GetClientNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationClientsColumn], search string, limit uint64) ([]t.NotificationClientsTableRow, *t.Paging, error)
	GetRocketPoolNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationRocketPoolColumn], search string, limit uint64) ([]t.NotificationRocketPoolTableRow, *t.Paging, error)
	GetNetworkNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationNetworksColumn], limit uint64) ([]t.NotificationNetworksTableRow, *t.Paging, error)

	GetNotificationSettings(ctx context.Context, userId uint64) (*t.NotificationSettings, error)
	GetNotificationSettingsDefaultValues(ctx context.Context) (*t.NotificationSettingsDefaultValues, error)
	UpdateNotificationSettingsGeneral(ctx context.Context, userId uint64, settings t.NotificationSettingsGeneral) error
	UpdateNotificationSettingsNetworks(ctx context.Context, userId uint64, chainId uint64, settings t.NotificationSettingsNetwork) error
	UpdateNotificationSettingsPairedDevice(ctx context.Context, userId uint64, pairedDeviceId uint64, name string, IsNotificationsEnabled bool) error
	DeleteNotificationSettingsPairedDevice(ctx context.Context, userId uint64, pairedDeviceId uint64) error
	UpdateNotificationSettingsClients(ctx context.Context, userId uint64, clientId uint64, IsSubscribed bool) (*t.NotificationSettingsClient, error)
	GetNotificationSettingsDashboards(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationSettingsDashboardColumn], search string, limit uint64) ([]t.NotificationSettingsDashboardsTableRow, *t.Paging, error)
	UpdateNotificationSettingsValidatorDashboard(ctx context.Context, userId uint64, dashboardId t.VDBIdPrimary, groupId uint64, settings t.NotificationSettingsValidatorDashboard) error
	UpdateNotificationSettingsAccountDashboard(ctx context.Context, userId uint64, dashboardId t.VDBIdPrimary, groupId uint64, settings t.NotificationSettingsAccountDashboard) error

	QueueTestEmailNotification(ctx context.Context, userId uint64) error
	QueueTestPushNotification(ctx context.Context, userId uint64) error
	QueueTestWebhookNotification(ctx context.Context, userId uint64, webhookUrl string, isDiscordWebhook bool) error
}

func (*DataAccessService) registerNotificationInterfaceTypes() {
	var once sync.Once
	once.Do(func() {
		gob.Register(&n.ValidatorProposalNotification{})
		gob.Register(&n.ValidatorUpcomingProposalNotification{})
		gob.Register(&n.ValidatorGroupEfficiencyNotification{})
		gob.Register(&n.ValidatorAttestationNotification{})
		gob.Register(&n.ValidatorIsOfflineNotification{})
		gob.Register(&n.ValidatorIsOnlineNotification{})
		gob.Register(&n.ValidatorGotSlashedNotification{})
		gob.Register(&n.ValidatorWithdrawalNotification{})
		gob.Register(&n.NetworkNotification{})
		gob.Register(&n.RocketpoolNotification{})
		gob.Register(&n.MonitorMachineNotification{})
		gob.Register(&n.TaxReportNotification{})
		gob.Register(&n.EthClientNotification{})
		gob.Register(&n.SyncCommitteeSoonNotification{})
	})
}

const (
	ValidatorDashboardEventPrefix string = "vdb"
	AccountDashboardEventPrefix   string = "adb"

	GroupEfficiencyBelowThresholdDefault     float64 = 0.95
	MaxCollateralThresholdDefault            float64 = 1.0
	MinCollateralThresholdDefault            float64 = 0.2
	ERC20TokenTransfersValueThresholdDefault float64 = 0.1

	MachineStorageUsageThresholdDefault float64 = 0.9
	MachineCpuUsageThresholdDefault     float64 = 0.6
	MachineMemoryUsageThresholdDefault  float64 = 0.8

	GasAboveThresholdDefault          float64 = 950
	GasBelowThresholdDefault          float64 = 150
	ParticipationRateThresholdDefault float64 = 0.8
)

func (d *DataAccessService) GetNotificationOverview(ctx context.Context, userId uint64) (*t.NotificationOverviewData, error) {
	response := t.NotificationOverviewData{}
	eg := errgroup.Group{}

	// enabled channels
	eg.Go(func() error {
		var channels []struct {
			Channel string `db:"channel"`
			Active  bool   `db:"active"`
		}

		err := d.userReader.SelectContext(ctx, &channels, `SELECT channel, active FROM users_notification_channels WHERE user_id = $1`, userId)
		if err != nil {
			return err
		}

		for _, channel := range channels {
			switch channel.Channel {
			case "email":
				response.IsEmailNotificationsEnabled = channel.Active
			case "push":
				response.IsPushNotificationsEnabled = channel.Active
			}
		}
		return nil
	})

	// most notified groups
	latestSlot, err := d.GetLatestSlot(ctx)
	if err != nil {
		return nil, err
	}
	epoch30dAgo := utils.TimeToEpoch(utils.EpochToTime(utils.EpochOfSlot(latestSlot)).Add(time.Duration(-30) * time.Hour * 24))
	getMostNotifiedGroups := func(historyTable, groupsTable string) ([3]string, error) {
		query := goqu.Dialect("postgres").
			From(goqu.T(historyTable).As("history")).
			Select(
				goqu.I("history.dashboard_id"),
				goqu.I("history.group_id"),
			).
			Where(
				goqu.Ex{"history.user_id": userId},
				goqu.I("history.epoch").Gt(epoch30dAgo),
			).
			GroupBy(
				goqu.I("history.dashboard_id"),
				goqu.I("history.group_id"),
			).
			Order(
				goqu.L("COUNT(*)").Desc(),
			).
			Limit(3)

		// join result with names
		query = goqu.Dialect("postgres").
			Select(goqu.L("COALESCE(name, 'default') AS name")).
			From(query.As("history")).
			LeftJoin(goqu.I(groupsTable).As("groups"), goqu.On(
				goqu.Ex{"groups.dashboard_id": goqu.I("history.dashboard_id")},
				goqu.Ex{"groups.id": goqu.I("history.group_id")},
			))

		mostNotifiedGroups := [3]string{}
		querySql, args, err := query.Prepared(true).ToSQL()
		if err != nil {
			return mostNotifiedGroups, fmt.Errorf("failed to prepare getMostNotifiedGroups query: %w", err)
		}
		res := []string{}
		err = d.alloyReader.SelectContext(ctx, &res, querySql, args...)
		if err != nil {
			return mostNotifiedGroups, fmt.Errorf("failed to execute getMostNotifiedGroups query: %w", err)
		}
		copy(mostNotifiedGroups[:], res)
		return mostNotifiedGroups, err
	}

	eg.Go(func() error {
		var err error
		response.VDBMostNotifiedGroups, err = getMostNotifiedGroups("users_val_dashboards_notifications_history", "users_val_dashboards_groups")
		return err
	})
	// TODO account dashboards
	/*eg.Go(func() error {
		var err error
		response.VDBMostNotifiedGroups, err = getMostNotifiedGroups("users_acc_dashboards_notifications_history", "users_acc_dashboards_groups")
		return err
	})*/

	// 24h counts
	eg.Go(func() error {
		var err error
		day := time.Now().Truncate(utils.Day).Unix()
		getMessageCount := func(prefix string) (uint64, error) {
			key := fmt.Sprintf("%s:%d:%d", prefix, userId, day)
			res := d.persistentRedisDbClient.Get(ctx, key)
			if res.Err() == redis.Nil {
				return 0, nil
			} else if res.Err() != nil {
				return 0, res.Err()
			}
			return res.Uint64()
		}
		response.Last24hEmailCount, err = getMessageCount("n_mails")
		if err != nil {
			return err
		}
		response.Last24hPushCount, err = getMessageCount("n_push")
		if err != nil {
			return err
		}
		response.Last24hWebhookCount, err = getMessageCount("n_webhook")
		return err
	})

	// subscription counts
	eg.Go(func() error {
		networks, err := d.GetAllNetworks()
		if err != nil {
			return err
		}

		whereNetwork := ""
		for _, network := range networks {
			if len(whereNetwork) > 0 {
				whereNetwork += " OR "
			}
			whereNetwork += "event_name like '" + network.NotificationsName + ":rocketpool_%' OR event_name like '" + network.NotificationsName + ":network_%'"
		}

		query := goqu.Dialect("postgres").
			From("users_subscriptions").
			Select(
				goqu.L("count(*) FILTER (WHERE event_filter like 'vdb:%')").As("vdb_subscriptions_count"),
				goqu.L("count(*) FILTER (WHERE event_filter like 'adb:%')").As("adb_subscriptions_count"),
				goqu.L("count(*) FILTER (WHERE event_name like 'monitoring_%')").As("machines_subscription_count"),
				goqu.L("count(*) FILTER (WHERE event_name = 'eth_client_update')").As("clients_subscription_count"),
				// not sure if there's a better way in goqu
				goqu.L("count(*) FILTER (WHERE "+whereNetwork+")").As("networks_subscription_count"),
			).
			Where(goqu.Ex{
				"user_id": userId,
			})

		querySql, args, err := query.Prepared(true).ToSQL()
		if err != nil {
			return err
		}

		err = d.userReader.GetContext(ctx, &response, querySql, args...)
		return err
	})

	err = eg.Wait()
	return &response, err
}

func (d *DataAccessService) GetDashboardNotifications(ctx context.Context, userId uint64, chainIds []uint64, cursor string, colSort t.Sort[enums.NotificationDashboardsColumn], search string, limit uint64) ([]t.NotificationDashboardsTableRow, *t.Paging, error) {
	response := []t.NotificationDashboardsTableRow{}
	var err error

	var currentCursor t.NotificationsDashboardsCursor
	if cursor != "" {
		if currentCursor, err = utils.StringToCursor[t.NotificationsDashboardsCursor](cursor); err != nil {
			return nil, nil, fmt.Errorf("failed to parse passed cursor as NotificationsDashboardsCursor: %w", err)
		}
	}

	// validator query
	vdbQuery := goqu.Dialect("postgres").
		From(goqu.T("users_val_dashboards_notifications_history").As("uvdnh")).
		Select(
			goqu.L("false").As("is_account_dashboard"),
			goqu.I("uvd.network").As("chain_id"),
			goqu.I("uvdnh.epoch"),
			goqu.I("uvd.id").As("dashboard_id"),
			goqu.I("uvd.name").As("dashboard_name"),
			goqu.I("uvdg.id").As("group_id"),
			goqu.I("uvdg.name").As("group_name"),
			goqu.SUM("uvdnh.event_count").As("entity_count"),
			goqu.L("ARRAY_AGG(DISTINCT event_type)").As("event_types"),
		).
		InnerJoin(goqu.T("users_val_dashboards").As("uvd"), goqu.On(
			goqu.Ex{"uvd.id": goqu.I("uvdnh.dashboard_id")})).
		InnerJoin(goqu.T("users_val_dashboards_groups").As("uvdg"), goqu.On(
			goqu.Ex{"uvdg.id": goqu.I("uvdnh.group_id")},
			goqu.Ex{"uvdg.dashboard_id": goqu.I("uvd.id")},
		)).
		Where(
			goqu.Ex{"uvd.user_id": userId},
		).
		GroupBy(
			goqu.I("uvdnh.epoch"),
			goqu.I("uvd.network"),
			goqu.I("uvd.id"),
			goqu.I("uvdg.id"),
			goqu.I("uvdg.name"),
		)

	if chainIds != nil {
		vdbQuery = vdbQuery.Where(
			goqu.L("uvd.network = ANY(?)", pq.Array(chainIds)),
		)
	}

	// TODO account dashboards
	/*adbQuery := goqu.Dialect("postgres").
		From(goqu.T("adb_notifications_history").As("anh")).
		Select(
			goqu.L("true").As("is_account_dashboard"),
			goqu.I("anh.network").As("chain_id"),
			goqu.I("anh.epoch"),
			goqu.I("uad.id").As("dashboard_id"),
			goqu.I("uad.name").As("dashboard_name"),
			goqu.I("uadg.id").As("group_id"),
			goqu.I("uadg.name").As("group_name"),
			goqu.SUM("anh.event_count").As("entity_count"),
			goqu.L("ARRAY_AGG(DISTINCT event_type)").As("event_types"),
		).
		InnerJoin(goqu.T("users_acc_dashboards").As("uad"), goqu.On(
			goqu.Ex{"uad.id": goqu.I("anh.dashboard_id"),
			})).
		InnerJoin(goqu.T("users_acc_dashboards_groups").As("uadg"), goqu.On(
			goqu.Ex{"uadg.id": goqu.I("anh.group_id"),
			goqu.Ex{"uadg.dashboard_id": goqu.I("uad.id")},
			})).
		Where(
			goqu.Ex{"uad.user_id": userId},
			goqu.L("anh.network = ANY(?)", pq.Array(chainIds)),
		).
		GroupBy(
			goqu.I("anh.epoch"),
			goqu.I("anh.network"),
			goqu.I("uad.id"),
			goqu.I("uadg.id"),
			goqu.I("uadg.name"),
		)

	unionQuery := vdbQuery.Union(adbQuery)*/
	unionQuery := goqu.From(vdbQuery)

	// sorting
	defaultColumns := []t.SortColumn{
		{Column: enums.NotificationsDashboardsColumns.Timestamp.ToString(), Desc: true, Offset: currentCursor.Epoch},
		{Column: enums.NotificationsDashboardsColumns.DashboardName.ToString(), Desc: false, Offset: currentCursor.DashboardName},
		{Column: enums.NotificationsDashboardsColumns.DashboardId.ToString(), Desc: false, Offset: currentCursor.DashboardId},
		{Column: enums.NotificationsDashboardsColumns.GroupName.ToString(), Desc: false, Offset: currentCursor.GroupName},
		{Column: enums.NotificationsDashboardsColumns.GroupId.ToString(), Desc: false, Offset: currentCursor.GroupId},
		{Column: enums.NotificationsDashboardsColumns.ChainId.ToString(), Desc: true, Offset: currentCursor.ChainId},
	}
	order, directions := applySortAndPagination(defaultColumns, t.SortColumn{Column: colSort.Column.ToString(), Desc: colSort.Desc}, currentCursor.GenericCursor)
	unionQuery = unionQuery.Order(order...)
	if directions != nil {
		unionQuery = unionQuery.Where(directions)
	}

	// search
	searchName := regexp.MustCompile(`^[a-zA-Z0-9_\-.\ ]+$`).MatchString(search)
	if searchName {
		searchLower := strings.ToLower(strings.Replace(search, "_", "\\_", -1)) + "%"
		unionQuery = unionQuery.Where(exp.NewExpressionList(
			exp.OrType,
			goqu.L("LOWER(?)", goqu.I("dashboard_name")).Like(searchLower),
			goqu.L("LOWER(?)", goqu.I("group_name")).Like(searchLower),
		))
	}
	unionQuery = unionQuery.Limit(uint(limit + 1))

	query, args, err := unionQuery.ToSQL()
	if err != nil {
		return nil, nil, err
	}
	err = d.alloyReader.SelectContext(ctx, &response, query, args...)
	if err != nil {
		return nil, nil, err
	}

	moreDataFlag := len(response) > int(limit)
	if moreDataFlag {
		response = response[:len(response)-1]
	}
	if currentCursor.IsReverse() {
		slices.Reverse(response)
	}
	if !moreDataFlag && !currentCursor.IsValid() {
		// No paging required
		return response, &t.Paging{}, nil
	}
	paging, err := utils.GetPagingFromData(response, currentCursor, moreDataFlag)
	if err != nil {
		return nil, nil, err
	}
	return response, paging, nil
}

func (d *DataAccessService) GetValidatorDashboardNotificationDetails(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, epoch uint64, search string) (*t.NotificationValidatorDashboardDetail, error) {
	notificationDetails := t.NotificationValidatorDashboardDetail{
		ValidatorOffline:         []uint64{},
		ProposalMissed:           []t.IndexSlots{},
		ProposalDone:             []t.IndexBlocks{},
		UpcomingProposals:        []t.IndexSlots{},
		Slashed:                  []uint64{},
		SyncCommittee:            []uint64{},
		AttestationMissed:        []t.IndexEpoch{},
		Withdrawal:               []t.NotificationEventWithdrawal{},
		ValidatorOfflineReminder: []uint64{},
		ValidatorBackOnline:      []t.NotificationEventValidatorBackOnline{},
		MinimumCollateralReached: []t.Address{},
		MaximumCollateralReached: []t.Address{},
	}

	var searchIndices []uint64
	// TODO move to api layer
	searchIndicesStrings := strings.Split(search, ",")
	for _, searchIndex := range searchIndicesStrings {
		idx, err := strconv.Atoi(searchIndex)
		if err == nil {
			searchIndices = append(searchIndices, uint64(idx))
		}
	}

	searchEnabled := len(searchIndices) > 0
	searchIndexSet := make(map[uint64]bool)
	for _, searchIndex := range searchIndices {
		searchIndexSet[searchIndex] = true
	}

	// -------------------------------------
	// dashboard and group name
	query := `SELECT
		uvd.name AS dashboard_name,
		uvdg.name AS group_name
	FROM
		users_val_dashboards uvd
	INNER JOIN
		users_val_dashboards_groups uvdg ON uvdg.dashboard_id = uvd.id
	WHERE uvd.id = $1 AND uvdg.id = $2`
	err := d.alloyReader.GetContext(ctx, &notificationDetails, query, dashboardId, groupId)
	if err != nil {
		if err == sql.ErrNoRows {
			return &notificationDetails, nil
		}
		return nil, err
	}
	if notificationDetails.GroupName == "" {
		notificationDetails.GroupName = t.DefaultGroupName
	}
	if notificationDetails.DashboardName == "" {
		notificationDetails.DashboardName = t.DefaultDashboardName
	}

	// -------------------------------------
	// retrieve notification events
	eventTypesEncodedList := [][]byte{}
	query = `SELECT details FROM users_val_dashboards_notifications_history WHERE dashboard_id = $1 AND group_id = $2 AND epoch = $3`
	err = d.alloyReader.SelectContext(ctx, &eventTypesEncodedList, query, dashboardId, groupId, epoch)
	if err != nil {
		return nil, err
	}
	if len(eventTypesEncodedList) == 0 {
		return &notificationDetails, nil
	}

	latestBlocks, err := d.GetLatestBlockHeightsForEpoch(ctx, epoch)
	if err != nil {
		return nil, fmt.Errorf("error getting latest block height: %w", err)
	}

	type ProposalInfo struct {
		Proposed  []uint64
		Scheduled []uint64
		Missed    []uint64
	}

	proposalsInfo := make(map[t.VDBValidator]*ProposalInfo)
	addressMapping := make(map[string]*t.Address)
	contractStatusRequests := make([]db.ContractInteractionAtRequest, 0)
	for _, eventTypesEncoded := range eventTypesEncodedList {
		buf := bytes.NewBuffer(eventTypesEncoded)
		gz, err := gzip.NewReader(buf)
		if err != nil {
			return nil, err
		}
		defer gz.Close()

		// might need to loop if we get memory issues
		eventTypes, err := io.ReadAll(gz)
		if err != nil {
			return nil, err
		}

		decoder := gob.NewDecoder(bytes.NewReader(eventTypes))

		notifications := []types.Notification{}
		err = decoder.Decode(&notifications)
		if err != nil {
			return nil, err
		}

		for _, notification := range notifications {
			switch notification.GetEventName() {
			case types.ValidatorMissedProposalEventName, types.ValidatorExecutedProposalEventName /*, types.ValidatorScheduledProposalEventName*/ :
				// aggregate proposals
				curNotification, ok := notification.(*n.ValidatorProposalNotification)
				if !ok {
					return nil, fmt.Errorf("failed to cast notification to ValidatorProposalNotification")
				}
				if searchEnabled && !searchIndexSet[curNotification.ValidatorIndex] {
					continue
				}
				if _, ok := proposalsInfo[curNotification.ValidatorIndex]; !ok {
					proposalsInfo[curNotification.ValidatorIndex] = &ProposalInfo{}
				}
				prop := proposalsInfo[curNotification.ValidatorIndex]
				switch curNotification.Status {
				case 0:
					prop.Scheduled = append(prop.Scheduled, curNotification.Slot)
				case 1:
					prop.Proposed = append(prop.Proposed, curNotification.Block)
				case 2:
					prop.Missed = append(prop.Missed, curNotification.Slot)
				}
			case types.ValidatorMissedAttestationEventName:
				curNotification, ok := notification.(*n.ValidatorAttestationNotification)
				if !ok {
					return nil, fmt.Errorf("failed to cast notification to ValidatorAttestationNotification")
				}
				if searchEnabled && !searchIndexSet[curNotification.ValidatorIndex] {
					continue
				}
				if curNotification.Status != 0 {
					continue
				}
				notificationDetails.AttestationMissed = append(notificationDetails.AttestationMissed, t.IndexEpoch{Index: curNotification.ValidatorIndex, Epoch: curNotification.Epoch})
			case types.ValidatorUpcomingProposalEventName:
				curNotification, ok := notification.(*n.ValidatorUpcomingProposalNotification)
				if !ok {
					return nil, fmt.Errorf("failed to cast notification to ValidatorUpcomingProposalNotification")
				}
				if searchEnabled && !searchIndexSet[curNotification.ValidatorIndex] {
					continue
				}
				notificationDetails.UpcomingProposals = append(notificationDetails.UpcomingProposals, t.IndexSlots{Index: curNotification.ValidatorIndex, Slots: []uint64{curNotification.Slot}})
			case types.ValidatorGotSlashedEventName:
				curNotification, ok := notification.(*n.ValidatorGotSlashedNotification)
				if !ok {
					return nil, fmt.Errorf("failed to cast notification to ValidatorGotSlashedNotification")
				}
				if searchEnabled && !searchIndexSet[curNotification.ValidatorIndex] {
					continue
				}
				notificationDetails.Slashed = append(notificationDetails.Slashed, curNotification.ValidatorIndex)
			case types.ValidatorIsOfflineEventName:
				curNotification, ok := notification.(*n.ValidatorIsOfflineNotification)
				if !ok {
					return nil, fmt.Errorf("failed to cast notification to ValidatorIsOfflineNotification")
				}
				if searchEnabled && !searchIndexSet[curNotification.ValidatorIndex] {
					continue
				}
				notificationDetails.ValidatorOffline = append(notificationDetails.ValidatorOffline, curNotification.ValidatorIndex)
				// TODO not present in backend yet
				//notificationDetails.ValidatorOfflineReminder = ...
			case types.ValidatorIsOnlineEventName:
				curNotification, ok := notification.(*n.ValidatorIsOnlineNotification)
				if !ok {
					return nil, fmt.Errorf("failed to cast notification to ValidatorIsOnlineNotification")
				}
				if searchEnabled && !searchIndexSet[curNotification.ValidatorIndex] {
					continue
				}
				notificationDetails.ValidatorBackOnline = append(notificationDetails.ValidatorBackOnline, t.NotificationEventValidatorBackOnline{Index: curNotification.ValidatorIndex, EpochCount: curNotification.Epoch})
			case types.ValidatorReceivedWithdrawalEventName:
				curNotification, ok := notification.(*n.ValidatorWithdrawalNotification)
				if !ok {
					return nil, fmt.Errorf("failed to cast notification to ValidatorWithdrawalNotification")
				}
				if searchEnabled && !searchIndexSet[curNotification.ValidatorIndex] {
					continue
				}
				// incorrect formatting TODO rework the Address and ContractInteractionAtRequest types to use clear string formatting (or prob go-ethereum common.Address)
				contractStatusRequests = append(contractStatusRequests, db.ContractInteractionAtRequest{
					Address:  fmt.Sprintf("%x", curNotification.Address),
					Block:    int64(latestBlocks[curNotification.Slot%utils.Config.Chain.ClConfig.SlotsPerEpoch]),
					TxIdx:    -1,
					TraceIdx: -1,
				})
				addr := t.Address{Hash: t.Hash(hexutil.Encode(curNotification.Address))}
				addressMapping[hexutil.Encode(curNotification.Address)] = &addr
				notificationDetails.Withdrawal = append(notificationDetails.Withdrawal, t.NotificationEventWithdrawal{
					Index:   curNotification.ValidatorIndex,
					Amount:  decimal.NewFromUint64(curNotification.Amount).Mul(decimal.NewFromFloat(params.GWei)), // Amounts have to be in WEI
					Address: addr,
				})
			case types.NetworkLivenessIncreasedEventName,
				types.EthClientUpdateEventName,
				types.MonitoringMachineOfflineEventName,
				types.MonitoringMachineDiskAlmostFullEventName,
				types.MonitoringMachineCpuLoadEventName,
				types.MonitoringMachineMemoryUsageEventName,
				types.TaxReportEventName:
				// not vdb notifications, skip
			case types.ValidatorDidSlashEventName:
			case types.RocketpoolCommissionThresholdEventName,
				types.RocketpoolNewClaimRoundStartedEventName:
				// these could maybe returned later (?)
			case types.RocketpoolCollateralMinReachedEventName, types.RocketpoolCollateralMaxReachedEventName:
				_, ok := notification.(*n.RocketpoolNotification)
				if !ok {
					return nil, fmt.Errorf("failed to cast notification to RocketpoolNotification")
				}
				addr := t.Address{Hash: t.Hash(notification.GetEventFilter()), IsContract: true}
				addressMapping[notification.GetEventFilter()] = &addr
				if notification.GetEventName() == types.RocketpoolCollateralMinReachedEventName {
					notificationDetails.MinimumCollateralReached = append(notificationDetails.MinimumCollateralReached, addr)
				} else {
					notificationDetails.MaximumCollateralReached = append(notificationDetails.MaximumCollateralReached, addr)
				}
			case types.SyncCommitteeSoonEventName:
				curNotification, ok := notification.(*n.SyncCommitteeSoonNotification)
				if !ok {
					return nil, fmt.Errorf("failed to cast notification to SyncCommitteeSoonNotification")
				}
				if searchEnabled && !searchIndexSet[curNotification.ValidatorIndex] {
					continue
				}
				notificationDetails.SyncCommittee = append(notificationDetails.SyncCommittee, curNotification.ValidatorIndex)
			default:
				log.Debugf("Unhandled notification type: %s", notification.GetEventName())
			}
		}
	}

	// fill proposals
	for validatorIndex, proposalInfo := range proposalsInfo {
		if len(proposalInfo.Proposed) > 0 {
			notificationDetails.ProposalDone = append(notificationDetails.ProposalDone, t.IndexBlocks{Index: validatorIndex, Blocks: proposalInfo.Proposed})
		}
		if len(proposalInfo.Scheduled) > 0 {
			notificationDetails.UpcomingProposals = append(notificationDetails.UpcomingProposals, t.IndexSlots{Index: validatorIndex, Slots: proposalInfo.Scheduled})
		}
		if len(proposalInfo.Missed) > 0 {
			notificationDetails.ProposalMissed = append(notificationDetails.ProposalMissed, t.IndexSlots{Index: validatorIndex, Slots: proposalInfo.Missed})
		}
	}

	// fill addresses
	if err := d.GetNamesAndEnsForAddresses(ctx, addressMapping); err != nil {
		return nil, err
	}
	contractStatuses, err := d.bigtable.GetAddressContractInteractionsAt(contractStatusRequests)
	if err != nil {
		return nil, err
	}
	contractStatusPerAddress := make(map[string]int)
	for i, contractStatus := range contractStatusRequests {
		contractStatusPerAddress["0x"+contractStatus.Address] = i
	}
	for i := range notificationDetails.MinimumCollateralReached {
		if address, ok := addressMapping[string(notificationDetails.MinimumCollateralReached[i].Hash)]; ok {
			notificationDetails.MinimumCollateralReached[i] = *address
		}
	}
	for i := range notificationDetails.MaximumCollateralReached {
		if address, ok := addressMapping[string(notificationDetails.MaximumCollateralReached[i].Hash)]; ok {
			notificationDetails.MaximumCollateralReached[i] = *address
		}
	}
	for i := range notificationDetails.Withdrawal {
		if address, ok := addressMapping[string(notificationDetails.Withdrawal[i].Address.Hash)]; ok {
			notificationDetails.Withdrawal[i].Address = *address
		}
		contractStatus := contractStatuses[contractStatusPerAddress[string(notificationDetails.Withdrawal[i].Address.Hash)]]
		notificationDetails.Withdrawal[i].Address.IsContract = contractStatus == types.CONTRACT_CREATION || contractStatus == types.CONTRACT_PRESENT
	}

	return &notificationDetails, nil
}

func (d *DataAccessService) GetAccountDashboardNotificationDetails(ctx context.Context, dashboardId uint64, groupId uint64, epoch uint64, search string) (*t.NotificationAccountDashboardDetail, error) {
	return d.dummy.GetAccountDashboardNotificationDetails(ctx, dashboardId, groupId, epoch, search)
}

func (d *DataAccessService) GetMachineNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationMachinesColumn], search string, limit uint64) ([]t.NotificationMachinesTableRow, *t.Paging, error) {
	result := make([]t.NotificationMachinesTableRow, 0)
	var paging t.Paging

	// Initialize the cursor
	var currentCursor t.NotificationMachinesCursor
	var err error
	if cursor != "" {
		currentCursor, err = utils.StringToCursor[t.NotificationMachinesCursor](cursor)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse passed cursor as NotificationMachinesCursor: %w", err)
		}
	}

	// -------------------------------------
	// Get the machine notification history
	notificationHistory := []struct {
		Epoch          uint64          `db:"epoch"`
		MachineId      uint64          `db:"machine_id"`
		MachineName    string          `db:"machine_name"`
		EventType      types.EventName `db:"event_type"`
		EventThreshold float64         `db:"event_threshold"`
	}{}

	ds := goqu.Dialect("postgres").
		Select(
			goqu.L("epoch"),
			goqu.L("machine_id"),
			goqu.L("machine_name"),
			goqu.L("event_type"),
			goqu.L("event_threshold")).
		From("machine_notifications_history").
		Where(goqu.L("user_id = ?", userId)).
		Limit(uint(limit + 1))

	// Search
	if search != "" {
		ds = ds.Where(goqu.L("machine_name ILIKE ?", search))
	}

	// Sorting and limiting if cursor is present
	defaultColumns := []t.SortColumn{
		{Column: enums.NotificationsMachinesColumns.Timestamp.ToString(), Desc: true, Offset: currentCursor.Epoch},
		{Column: enums.NotificationsMachinesColumns.MachineId.ToString(), Desc: false, Offset: currentCursor.MachineId},
		{Column: enums.NotificationsMachinesColumns.EventType.ToString(), Desc: false, Offset: currentCursor.EventType},
	}
	var offset interface{}
	switch colSort.Column {
	case enums.NotificationsMachinesColumns.MachineName:
		offset = currentCursor.MachineName
	case enums.NotificationsMachinesColumns.Threshold:
		offset = currentCursor.EventThreshold
	}

	order, directions := applySortAndPagination(defaultColumns, t.SortColumn{Column: colSort.Column.ToString(), Desc: colSort.Desc, Offset: offset}, currentCursor.GenericCursor)
	ds = ds.Order(order...)
	if directions != nil {
		ds = ds.Where(directions)
	}

	query, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return nil, nil, fmt.Errorf("error preparing machine notifications query: %w", err)
	}

	err = d.userReader.SelectContext(ctx, &notificationHistory, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf(`error retrieving data for machine notifications: %w`, err)
	}

	// -------------------------------------
	// Calculate the result
	cursorData := notificationHistory
	for _, notification := range notificationHistory {
		resultEntry := t.NotificationMachinesTableRow{
			MachineName: notification.MachineName,
			Threshold:   notification.EventThreshold,
			Timestamp:   utils.EpochToTime(notification.Epoch).Unix(),
		}
		switch notification.EventType {
		case types.MonitoringMachineOfflineEventName:
			resultEntry.EventType = "offline"
		case types.MonitoringMachineDiskAlmostFullEventName:
			resultEntry.EventType = "storage"
		case types.MonitoringMachineCpuLoadEventName:
			resultEntry.EventType = "cpu"
		case types.MonitoringMachineMemoryUsageEventName:
			resultEntry.EventType = "memory"
		default:
			return nil, nil, fmt.Errorf("invalid event name for machine notification: %v", notification.EventType)
		}
		result = append(result, resultEntry)
	}

	// -------------------------------------
	// Paging

	// Flag if above limit
	moreDataFlag := len(result) > int(limit)
	if !moreDataFlag && !currentCursor.IsValid() {
		// No paging required
		return result, &paging, nil
	}

	// Remove the last entries from data
	if moreDataFlag {
		result = result[:limit]
		cursorData = cursorData[:limit]
	}

	if currentCursor.IsReverse() {
		slices.Reverse(result)
		slices.Reverse(cursorData)
	}

	p, err := utils.GetPagingFromData(cursorData, currentCursor, moreDataFlag)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get paging: %w", err)
	}

	return result, p, nil
}
func (d *DataAccessService) GetClientNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationClientsColumn], search string, limit uint64) ([]t.NotificationClientsTableRow, *t.Paging, error) {
	result := make([]t.NotificationClientsTableRow, 0)
	var paging t.Paging

	// Initialize the cursor
	var currentCursor t.NotificationClientsCursor
	var err error
	if cursor != "" {
		currentCursor, err = utils.StringToCursor[t.NotificationClientsCursor](cursor)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse passed cursor as NotificationClientsCursor: %w", err)
		}
	}

	// -------------------------------------
	// Get the client notification history
	notificationHistory := []struct {
		Epoch   uint64 `db:"epoch"`
		Client  string `db:"client"`
		Version string `db:"client_version"`
		Url     string `db:"client_url"`
	}{}

	ds := goqu.Dialect("postgres").
		Select(
			goqu.L("epoch"),
			goqu.L("client"),
			goqu.L("client_version"),
			goqu.L("client_url")).
		From("client_notifications_history").
		Where(goqu.L("user_id = ?", userId)).
		Limit(uint(limit + 1))

	// Search
	if search != "" {
		ds = ds.Where(goqu.L("client ILIKE ?", search))
	}

	// Sorting and limiting if cursor is present
	// Rows can be uniquely identified by (epoch, client)
	defaultColumns := []t.SortColumn{
		{Column: enums.NotificationsClientsColumns.Timestamp.ToString(), Desc: true, Offset: currentCursor.Epoch},
		{Column: enums.NotificationsClientsColumns.ClientName.ToString(), Desc: false, Offset: currentCursor.Client},
	}
	order, directions := applySortAndPagination(defaultColumns, t.SortColumn{Column: colSort.Column.ToString(), Desc: colSort.Desc}, currentCursor.GenericCursor)
	ds = ds.Order(order...)
	if directions != nil {
		ds = ds.Where(directions)
	}

	query, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return nil, nil, fmt.Errorf("error preparing client notifications query: %w", err)
	}

	err = d.userReader.SelectContext(ctx, &notificationHistory, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf(`error retrieving data for client notifications: %w`, err)
	}

	// -------------------------------------
	// Calculate the result
	cursorData := notificationHistory
	for _, notification := range notificationHistory {
		resultEntry := t.NotificationClientsTableRow{
			ClientName: notification.Client,
			Version:    notification.Version,
			Url:        notification.Url,
			Timestamp:  utils.EpochToTime(notification.Epoch).Unix(),
		}
		result = append(result, resultEntry)
	}

	// -------------------------------------
	// Paging

	// Flag if above limit
	moreDataFlag := len(result) > int(limit)
	if !moreDataFlag && !currentCursor.IsValid() {
		// No paging required
		return result, &paging, nil
	}

	// Remove the last entries from data
	if moreDataFlag {
		result = result[:limit]
		cursorData = cursorData[:limit]
	}

	if currentCursor.IsReverse() {
		slices.Reverse(result)
		slices.Reverse(cursorData)
	}

	p, err := utils.GetPagingFromData(cursorData, currentCursor, moreDataFlag)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get paging: %w", err)
	}

	return result, p, nil
}
func (d *DataAccessService) GetRocketPoolNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationRocketPoolColumn], search string, limit uint64) ([]t.NotificationRocketPoolTableRow, *t.Paging, error) {
	return d.dummy.GetRocketPoolNotifications(ctx, userId, cursor, colSort, search, limit)

	// TODO: Adjust after db structure has been clarified
	// result := make([]t.NotificationRocketPoolTableRow, 0)
	// var paging t.Paging

	// // Initialize the cursor
	// var currentCursor t.NotificationRocketPoolsCursor
	// var err error
	// if cursor != "" {
	// 	currentCursor, err = utils.StringToCursor[t.NotificationRocketPoolsCursor](cursor)
	// 	if err != nil {
	// 		return nil, nil, fmt.Errorf("failed to parse passed cursor as NotificationRocketPoolsCursor: %w", err)
	// 	}
	// }

	// isReverseDirection := (colSort.Desc && !currentCursor.IsReverse()) || (!colSort.Desc && currentCursor.IsReverse())
	// sortSearchDirection := ">"
	// if isReverseDirection {
	// 	sortSearchDirection = "<"
	// }

	// // -------------------------------------
	// // Get the machine notification history
	// notificationHistory := []struct {
	// 	Epoch          uint64          `db:"epoch"`
	// 	LastBlock      int64           `db:"last_block"`
	// 	EventType      types.EventName `db:"event_type"`
	// 	EventThreshold float64         `db:"event_threshold"`
	// 	NodeAddress    []byte          `db:"node_address"`
	// }{}

	// ds := goqu.Dialect("postgres").
	// 	Select(
	// 		goqu.L("epoch"),
	// 		goqu.L("last_block"),
	// 		goqu.L("event_type"),
	// 		goqu.L("event_threshold"),
	// 		goqu.L("node_address")).
	// 	From("rocketpool_notifications_history").
	// 	Where(goqu.L("user_id = ?", userId)).
	// 	Limit(uint(limit + 1))

	// // Search
	// if search != "" {
	// 	if !utils.IsEth1Address(search) {
	// 		// If search is not a valid address, return empty result
	// 		return result, &paging, nil
	// 	}
	// 	nodeAddress, err := hexutil.Decode(search)
	// 	if err != nil {
	// 		return nil, nil, fmt.Errorf("failed to decode node address: %w", err)
	// 	}
	// 	ds = ds.Where(goqu.L("node_address = ?", nodeAddress))
	// }

	// // Sorting and limiting if cursor is present
	// // Rows can be uniquely identified by (epoch, event_type, node_address)
	// sortDirFunc := func(column string) exp.OrderedExpression {
	// 	return goqu.I(column).Asc()
	// }
	// if isReverseDirection {
	// 	sortDirFunc = func(column string) exp.OrderedExpression {
	// 		return goqu.I(column).Desc()
	// 	}
	// }
	// switch colSort.Column {
	// case enums.NotificationRocketPoolColumns.Timestamp:
	// 	if currentCursor.IsValid() {
	// 		ds = ds.Where(goqu.Or(
	// 			goqu.L(fmt.Sprintf("(epoch %s ?)", sortSearchDirection), currentCursor.Epoch),
	// 			goqu.L(fmt.Sprintf("(epoch = ? AND event_type %s ?)", sortSearchDirection), currentCursor.Epoch, currentCursor.EventType),
	// 			goqu.L(fmt.Sprintf("(epoch = ? AND event_type = ? AND node_address %s ?)", sortSearchDirection), currentCursor.Epoch, currentCursor.EventType, currentCursor.NodeAddress),
	// 		))
	// 	}
	// 	ds = ds.Order(
	// 		sortDirFunc("epoch"),
	// 		sortDirFunc("event_type"),
	// 		sortDirFunc("node_address"))
	// case enums.NotificationRocketPoolColumns.EventType:
	// 	if currentCursor.IsValid() {
	// 		ds = ds.Where(goqu.Or(
	// 			goqu.L(fmt.Sprintf("(event_type %s ?)", sortSearchDirection), currentCursor.EventType),
	// 			goqu.L(fmt.Sprintf("(event_type = ? AND epoch %s ?)", sortSearchDirection), currentCursor.EventType, currentCursor.Epoch),
	// 			goqu.L(fmt.Sprintf("(event_type = ? AND epoch = ? AND node_address %s ?)", sortSearchDirection), currentCursor.EventType, currentCursor.Epoch, currentCursor.NodeAddress),
	// 		))
	// 	}
	// 	ds = ds.Order(
	// 		sortDirFunc("event_type"),
	// 		sortDirFunc("epoch"),
	// 		sortDirFunc("node_address"))
	// case enums.NotificationRocketPoolColumns.NodeAddress:
	// 	if currentCursor.IsValid() {
	// 		ds = ds.Where(goqu.Or(
	// 			goqu.L(fmt.Sprintf("(node_address %s ?)", sortSearchDirection), currentCursor.NodeAddress),
	// 			goqu.L(fmt.Sprintf("(node_address = ? AND epoch %s ?)", sortSearchDirection), currentCursor.NodeAddress, currentCursor.Epoch),
	// 			goqu.L(fmt.Sprintf("(node_address = ? AND epoch = ? AND event_type %s ?)", sortSearchDirection), currentCursor.NodeAddress, currentCursor.Epoch, currentCursor.EventType),
	// 		))
	// 	}
	// 	ds = ds.Order(
	// 		sortDirFunc("node_address"),
	// 		sortDirFunc("epoch"),
	// 		sortDirFunc("event_type"))
	// default:
	// 	return nil, nil, fmt.Errorf("invalid column for sorting of rocketpool notification history: %v", colSort.Column)
	// }

	// query, args, err := ds.Prepared(true).ToSQL()
	// if err != nil {
	// 	return nil, nil, fmt.Errorf("error preparing rocketpool notifications query: %w", err)
	// }

	// err = d.userReader.SelectContext(ctx, &notificationHistory, query, args...)
	// if err != nil {
	// 	return nil, nil, fmt.Errorf(`error retrieving data for rocketpool notifications: %w`, err)
	// }

	// // -------------------------------------
	// // Get the node address info
	// addressMapping := make(map[string]*t.Address)
	// contractStatusRequests := make([]db.ContractInteractionAtRequest, 0)

	// for _, notification := range notificationHistory {
	// 	addressMapping[hexutil.Encode(notification.NodeAddress)] = nil
	// 	contractStatusRequests = append(contractStatusRequests, db.ContractInteractionAtRequest{
	// 		Address:  fmt.Sprintf("%x", notification.NodeAddress),
	// 		Block:    notification.LastBlock,
	// 		TxIdx:    -1,
	// 		TraceIdx: -1,
	// 	})
	// }

	// err = d.GetNamesAndEnsForAddresses(ctx, addressMapping)
	// if err != nil {
	// 	return nil, nil, err
	// }

	// contractStatuses, err := d.bigtable.GetAddressContractInteractionsAt(contractStatusRequests)
	// if err != nil {
	// 	return nil, nil, err
	// }

	// // -------------------------------------
	// // Calculate the result
	// cursorData := notificationHistory
	// for idx, notification := range notificationHistory {
	// 	resultEntry := t.NotificationRocketPoolTableRow{
	// 		Timestamp: utils.EpochToTime(notification.Epoch).Unix(),
	// 		Threshold: notification.EventThreshold,
	// 		Node:      *addressMapping[hexutil.Encode(notification.NodeAddress)],
	// 	}
	// 	resultEntry.Node.IsContract = contractStatuses[idx] == types.CONTRACT_CREATION || contractStatuses[idx] == types.CONTRACT_PRESENT

	// 	switch notification.EventType {
	// 	case types.RocketpoolNewClaimRoundStartedEventName:
	// 		resultEntry.EventType = "reward_round"
	// 	case types.RocketpoolCollateralMinReachedEventName:
	// 		resultEntry.EventType = "collateral_min"
	// 	case types.RocketpoolCollateralMaxReachedEventName:
	// 		resultEntry.EventType = "collateral_max"
	// 	default:
	// 		return nil, nil, fmt.Errorf("invalid event name for rocketpool notification: %v", notification.EventType)
	// 	}
	// 	result = append(result, resultEntry)
	// }

	// // -------------------------------------
	// // Paging

	// // Flag if above limit
	// moreDataFlag := len(result) > int(limit)
	// if !moreDataFlag && !currentCursor.IsValid() {
	// 	// No paging required
	// 	return result, &paging, nil
	// }

	// // Remove the last entries from data
	// if moreDataFlag {
	// 	result = result[:limit]
	// 	cursorData = cursorData[:limit]
	// }

	// if currentCursor.IsReverse() {
	// 	slices.Reverse(result)
	// 	slices.Reverse(cursorData)
	// }

	// p, err := utils.GetPagingFromData(cursorData, currentCursor, moreDataFlag)
	// if err != nil {
	// 	return nil, nil, fmt.Errorf("failed to get paging: %w", err)
	// }

	// return result, p, nil
}
func (d *DataAccessService) GetNetworkNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationNetworksColumn], limit uint64) ([]t.NotificationNetworksTableRow, *t.Paging, error) {
	result := make([]t.NotificationNetworksTableRow, 0)
	var paging t.Paging

	// Initialize the cursor
	var currentCursor t.NotificationNetworksCursor
	var err error
	if cursor != "" {
		currentCursor, err = utils.StringToCursor[t.NotificationNetworksCursor](cursor)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse passed cursor as NotificationNetworksCursor: %w", err)
		}
	}

	// -------------------------------------
	// Get the network notification history
	notificationHistory := []struct {
		Epoch          uint64          `db:"epoch"`
		Network        uint64          `db:"network"`
		EventType      types.EventName `db:"event_type"`
		EventThreshold float64         `db:"event_threshold"`
	}{}

	ds := goqu.Dialect("postgres").
		Select(
			goqu.L("epoch"),
			goqu.L("network"),
			goqu.L("event_type"),
			goqu.L("event_threshold")).
		From("network_notifications_history").
		Where(goqu.L("user_id = ?", userId)).
		Limit(uint(limit + 1))

	// Sorting and limiting if cursor is present
	// Rows can be uniquely identified by (epoch, network, event_type)
	defaultColumns := []t.SortColumn{
		{Column: enums.NotificationNetworksColumns.Timestamp.ToString(), Desc: true, Offset: currentCursor.Epoch},
		{Column: enums.NotificationNetworksColumns.Network.ToString(), Desc: false, Offset: currentCursor.Network},
		{Column: enums.NotificationNetworksColumns.EventType.ToString(), Desc: false, Offset: currentCursor.EventType},
	}
	order, directions := applySortAndPagination(defaultColumns, t.SortColumn{Column: colSort.Column.ToString(), Desc: colSort.Desc}, currentCursor.GenericCursor)
	ds = ds.Order(order...)
	if directions != nil {
		ds = ds.Where(directions)
	}

	query, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return nil, nil, fmt.Errorf("error preparing network notifications query: %w", err)
	}

	err = d.userReader.SelectContext(ctx, &notificationHistory, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf(`error retrieving data for network notifications: %w`, err)
	}

	// -------------------------------------
	// Calculate the result
	cursorData := notificationHistory
	for _, notification := range notificationHistory {
		resultEntry := t.NotificationNetworksTableRow{
			ChainId:   notification.Network,
			Timestamp: utils.EpochToTime(notification.Epoch).Unix(),
		}
		switch notification.EventType {
		case types.NetworkGasAboveThresholdEventName:
			resultEntry.EventType = "gas_above"
			resultEntry.Threshold = decimal.NewFromFloat(notification.EventThreshold).Mul(decimal.NewFromInt(params.GWei))
		case types.NetworkGasBelowThresholdEventName:
			resultEntry.EventType = "gas_below"
			resultEntry.Threshold = decimal.NewFromFloat(notification.EventThreshold).Mul(decimal.NewFromInt(params.GWei))
		case types.NetworkParticipationRateThresholdEventName:
			resultEntry.EventType = "participation_rate"
			resultEntry.Threshold = decimal.NewFromFloat(notification.EventThreshold)
		case types.RocketpoolNewClaimRoundStartedEventName:
			resultEntry.EventType = "new_reward_round"
		default:
			return nil, nil, fmt.Errorf("invalid event name for network notification: %v", notification.EventType)
		}
		result = append(result, resultEntry)
	}

	// -------------------------------------
	// Paging

	// Flag if above limit
	moreDataFlag := len(result) > int(limit)
	if !moreDataFlag && !currentCursor.IsValid() {
		// No paging required
		return result, &paging, nil
	}

	// Remove the last entries from data
	if moreDataFlag {
		result = result[:limit]
		cursorData = cursorData[:limit]
	}

	if currentCursor.IsReverse() {
		slices.Reverse(result)
		slices.Reverse(cursorData)
	}

	p, err := utils.GetPagingFromData(cursorData, currentCursor, moreDataFlag)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get paging: %w", err)
	}

	return result, p, nil
}
func (d *DataAccessService) GetNotificationSettings(ctx context.Context, userId uint64) (*t.NotificationSettings, error) {
	wg := errgroup.Group{}

	// -------------------------------------
	// Create the default settings
	result := &t.NotificationSettings{
		GeneralSettings: t.NotificationSettingsGeneral{
			MachineStorageUsageThreshold: MachineStorageUsageThresholdDefault,
			MachineCpuUsageThreshold:     MachineCpuUsageThresholdDefault,
			MachineMemoryUsageThreshold:  MachineMemoryUsageThresholdDefault,
		},
	}

	// For networks
	networks, err := d.GetAllNetworks()
	if err != nil {
		return nil, err
	}
	networksSettings := make(map[string]*t.NotificationNetwork, len(networks))
	for _, network := range networks {
		networksSettings[network.NotificationsName] = &t.NotificationNetwork{
			ChainId: network.ChainId,
			Settings: t.NotificationSettingsNetwork{
				GasAboveThreshold:          decimal.NewFromFloat(GasAboveThresholdDefault).Mul(decimal.NewFromInt(params.GWei)),
				GasBelowThreshold:          decimal.NewFromFloat(GasBelowThresholdDefault).Mul(decimal.NewFromInt(params.GWei)),
				ParticipationRateThreshold: ParticipationRateThresholdDefault,
			},
		}
	}

	// For clients
	clients, err := d.GetAllClients()
	if err != nil {
		return nil, err
	}
	clientSettings := make(map[string]*t.NotificationSettingsClient, len(clients))
	for _, client := range clients {
		clientSettings[client.DbName] = &t.NotificationSettingsClient{
			Id:       client.Id,
			Name:     client.Name,
			Category: client.Category,
		}
	}

	// -------------------------------------
	// Get the "do not disturb" setting
	var doNotDisturbTimestamp sql.NullTime
	wg.Go(func() error {
		err := d.userReader.GetContext(ctx, &doNotDisturbTimestamp, `
		SELECT
			notifications_do_not_disturb_ts
		FROM users
		WHERE id = $1`, userId)
		if err != nil {
			return fmt.Errorf(`error retrieving data for notifications "do not disturb" setting: %w`, err)
		}

		return nil
	})

	// -------------------------------------
	// Get the notification channels
	notificationChannels := []struct {
		Channel types.NotificationChannel `db:"channel"`
		Active  bool                      `db:"active"`
	}{}
	wg.Go(func() error {
		err := d.userReader.SelectContext(ctx, &notificationChannels, `
		SELECT
			channel,
			active
		FROM users_notification_channels
		WHERE user_id = $1`, userId)
		if err != nil {
			return fmt.Errorf(`error retrieving data for notifications channels: %w`, err)
		}

		return nil
	})

	// -------------------------------------
	// Get the subscribed events
	subscribedEvents := []struct {
		Name      types.EventName `db:"event_name"`
		Filter    string          `db:"event_filter"`
		Threshold float64         `db:"event_threshold"`
	}{}
	wg.Go(func() error {
		err := d.userReader.SelectContext(ctx, &subscribedEvents, `
		SELECT
			event_name,
			event_filter,
			event_threshold
		FROM users_subscriptions
		WHERE user_id = $1`, userId)
		if err != nil {
			return fmt.Errorf(`error retrieving data for notifications subscribed events: %w`, err)
		}

		return nil
	})

	// -------------------------------------
	// Get the paired devices
	pairedDevices := []struct {
		DeviceId      uint64    `db:"id"`
		CreatedTs     time.Time `db:"created_ts"`
		DeviceName    string    `db:"device_name"`
		NotifyEnabled bool      `db:"notify_enabled"`
	}{}
	wg.Go(func() error {
		err := d.userReader.SelectContext(ctx, &pairedDevices, `
		SELECT
			id,
			created_ts,
			device_name,
			COALESCE(notify_enabled, false) AS notify_enabled
		FROM users_devices
		WHERE user_id = $1`, userId)
		if err != nil {
			return fmt.Errorf(`error retrieving data for notifications paired devices: %w`, err)
		}

		return nil
	})

	// -------------------------------------
	// Get the machines
	hasMachines := false
	wg.Go(func() error {
		machineNames, err := db.BigtableClient.GetMachineMetricsMachineNames(types.UserId(userId))
		if err != nil {
			return fmt.Errorf(`error retrieving data for notifications machine names: %w`, err)
		}
		if len(machineNames) > 0 {
			hasMachines = true
		}
		return nil
	})

	err = wg.Wait()
	if err != nil {
		return nil, err
	}

	// -------------------------------------
	// Fill the result
	result.HasMachines = hasMachines
	if doNotDisturbTimestamp.Valid {
		result.GeneralSettings.DoNotDisturbTimestamp = doNotDisturbTimestamp.Time.Unix()
	}

	for _, channel := range notificationChannels {
		if channel.Channel == types.EmailNotificationChannel {
			result.GeneralSettings.IsEmailNotificationsEnabled = channel.Active
		} else if channel.Channel == types.PushNotificationChannel {
			result.GeneralSettings.IsPushNotificationsEnabled = channel.Active
		}
	}

	for _, event := range subscribedEvents {
		eventSplit := strings.Split(string(event.Name), ":")

		if len(eventSplit) == 2 {
			networkName := eventSplit[0]
			networkEvent := types.EventName(eventSplit[1])

			if _, ok := networksSettings[networkName]; !ok {
				return nil, fmt.Errorf("network is not defined: %s", networkName)
			}

			switch networkEvent {
			case types.RocketpoolNewClaimRoundStartedEventName:
				networksSettings[networkName].Settings.IsNewRewardRoundSubscribed = true
			case types.NetworkGasAboveThresholdEventName:
				networksSettings[networkName].Settings.IsGasAboveSubscribed = true
				networksSettings[networkName].Settings.GasAboveThreshold = decimal.NewFromFloat(event.Threshold).Mul(decimal.NewFromInt(params.GWei))
			case types.NetworkGasBelowThresholdEventName:
				networksSettings[networkName].Settings.IsGasBelowSubscribed = true
				networksSettings[networkName].Settings.GasBelowThreshold = decimal.NewFromFloat(event.Threshold).Mul(decimal.NewFromInt(params.GWei))
			case types.NetworkParticipationRateThresholdEventName:
				networksSettings[networkName].Settings.IsParticipationRateSubscribed = true
				networksSettings[networkName].Settings.ParticipationRateThreshold = event.Threshold
			}
		} else {
			switch event.Name {
			case types.MonitoringMachineOfflineEventName:
				result.GeneralSettings.IsMachineOfflineSubscribed = true
			case types.MonitoringMachineDiskAlmostFullEventName:
				result.GeneralSettings.IsMachineStorageUsageSubscribed = true
				result.GeneralSettings.MachineStorageUsageThreshold = event.Threshold
			case types.MonitoringMachineCpuLoadEventName:
				result.GeneralSettings.IsMachineCpuUsageSubscribed = true
				result.GeneralSettings.MachineCpuUsageThreshold = event.Threshold
			case types.MonitoringMachineMemoryUsageEventName:
				result.GeneralSettings.IsMachineMemoryUsageSubscribed = true
				result.GeneralSettings.MachineMemoryUsageThreshold = event.Threshold
			case types.EthClientUpdateEventName:
				if clientSettings[event.Filter] != nil {
					clientSettings[event.Filter].IsSubscribed = true
				} else {
					log.Warnf("client %s is not found in the client settings", event.Filter)
				}
			}
		}
	}

	for _, settings := range networksSettings {
		result.Networks = append(result.Networks, *settings)
	}

	for _, device := range pairedDevices {
		result.PairedDevices = append(result.PairedDevices, t.NotificationPairedDevice{
			Id:                     device.DeviceId,
			PairedTimestamp:        device.CreatedTs.Unix(),
			Name:                   device.DeviceName,
			IsNotificationsEnabled: device.NotifyEnabled,
		})
	}

	for _, settings := range clientSettings {
		result.Clients = append(result.Clients, *settings)
	}

	// properly sort the responses
	sort.Slice(result.Networks, func(i, j int) bool { // sort by chain id ascending
		return result.Networks[i].ChainId < result.Networks[j].ChainId
	})
	sort.Slice(result.Clients, func(i, j int) bool { // sort by client name ascending
		return result.Clients[i].Name < result.Clients[j].Name
	})
	sort.Slice(result.PairedDevices, func(i, j int) bool { // sort by paired timestamp descending
		return result.PairedDevices[i].PairedTimestamp > result.PairedDevices[j].PairedTimestamp
	})

	return result, nil
}

func (d *DataAccessService) GetNotificationSettingsDefaultValues(ctx context.Context) (*t.NotificationSettingsDefaultValues, error) {
	return &t.NotificationSettingsDefaultValues{
		GroupEfficiencyBelowThreshold:     GroupEfficiencyBelowThresholdDefault,
		MaxCollateralThreshold:            MaxCollateralThresholdDefault,
		MinCollateralThreshold:            MinCollateralThresholdDefault,
		ERC20TokenTransfersValueThreshold: ERC20TokenTransfersValueThresholdDefault,

		MachineStorageUsageThreshold: MachineStorageUsageThresholdDefault,
		MachineCpuUsageThreshold:     MachineCpuUsageThresholdDefault,
		MachineMemoryUsageThreshold:  MachineMemoryUsageThresholdDefault,

		GasAboveThreshold: decimal.NewFromFloat(GasAboveThresholdDefault).Mul(decimal.NewFromInt(params.GWei)),
		GasBelowThreshold: decimal.NewFromFloat(GasAboveThresholdDefault).Mul(decimal.NewFromInt(params.GWei)),

		NetworkParticipationRateThreshold: ParticipationRateThresholdDefault,
	}, nil
}

func (d *DataAccessService) UpdateNotificationSettingsGeneral(ctx context.Context, userId uint64, settings t.NotificationSettingsGeneral) error {
	epoch := utils.TimeToEpoch(time.Now())

	var eventsToInsert []goqu.Record
	var eventsToDelete []goqu.Expression

	tx, err := d.userWriter.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting db transactions to update general notification settings: %w", err)
	}
	defer utils.Rollback(tx)

	// -------------------------------------
	// Set the "do not disturb" setting
	_, err = tx.ExecContext(ctx, `
		UPDATE users 
		SET notifications_do_not_disturb_ts = 
		    CASE 
		        WHEN $1 = 0 THEN NULL
		        ELSE TO_TIMESTAMP($1)
		    END 
		WHERE id = $2`, settings.DoNotDisturbTimestamp, userId)
	if err != nil {
		return err
	}

	// -------------------------------------
	// Set the notification channels
	_, err = tx.ExecContext(ctx, `
		INSERT INTO users_notification_channels (user_id, channel, active)
    		VALUES ($1, $2, $3), ($1, $4, $5)
    	ON CONFLICT (user_id, channel) 
    		DO UPDATE SET active = EXCLUDED.active`,
		userId, types.EmailNotificationChannel, settings.IsEmailNotificationsEnabled, types.PushNotificationChannel, settings.IsPushNotificationsEnabled)
	if err != nil {
		return err
	}

	// -------------------------------------
	// Collect the machine and rocketpool events to set and delete

	//Machine events
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsMachineOfflineSubscribed, userId, types.MonitoringMachineOfflineEventName, "", "", epoch, 0)
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsMachineStorageUsageSubscribed, userId, types.MonitoringMachineDiskAlmostFullEventName, "", "", epoch, settings.MachineStorageUsageThreshold)
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsMachineCpuUsageSubscribed, userId, types.MonitoringMachineCpuLoadEventName, "", "", epoch, settings.MachineCpuUsageThreshold)
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsMachineMemoryUsageSubscribed, userId, types.MonitoringMachineMemoryUsageEventName, "", "", epoch, settings.MachineMemoryUsageThreshold)

	// Insert all the events or update the threshold if they already exist
	if len(eventsToInsert) > 0 {
		insertDs := goqu.Dialect("postgres").
			Insert("users_subscriptions").
			Cols("user_id", "event_name", "event_filter", "created_ts", "created_epoch", "event_threshold").
			Rows(eventsToInsert).
			OnConflict(goqu.DoUpdate(
				"user_id, event_name, event_filter",
				goqu.Record{"event_threshold": goqu.L("EXCLUDED.event_threshold")},
			))

		query, args, err := insertDs.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %v", err)
		}

		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return err
		}
	}

	// Delete all the events
	if len(eventsToDelete) > 0 {
		deleteDs := goqu.Dialect("postgres").
			Delete("users_subscriptions").
			Where(goqu.Or(eventsToDelete...))

		query, args, err := deleteDs.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %v", err)
		}

		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing tx to update general notification settings: %w", err)
	}
	return nil
}
func (d *DataAccessService) UpdateNotificationSettingsNetworks(ctx context.Context, userId uint64, chainId uint64, settings t.NotificationSettingsNetwork) error {
	epoch := utils.TimeToEpoch(time.Now())

	networks, err := d.GetAllNetworks()
	if err != nil {
		return err
	}

	networkName := ""
	for _, network := range networks {
		if network.ChainId == chainId {
			networkName = network.NotificationsName
			break
		}
	}
	if networkName == "" {
		return fmt.Errorf("network with chain id %d to update general notification settings not found", chainId)
	}

	var eventsToInsert []goqu.Record
	var eventsToDelete []goqu.Expression

	tx, err := d.userWriter.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting db transactions to update general notification settings: %w", err)
	}
	defer utils.Rollback(tx)

	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsGasAboveSubscribed, userId, types.NetworkGasAboveThresholdEventName, networkName, "", epoch, settings.GasAboveThreshold.Div(decimal.NewFromInt(params.GWei)).InexactFloat64())
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsGasBelowSubscribed, userId, types.NetworkGasBelowThresholdEventName, networkName, "", epoch, settings.GasBelowThreshold.Div(decimal.NewFromInt(params.GWei)).InexactFloat64())
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsParticipationRateSubscribed, userId, types.NetworkParticipationRateThresholdEventName, networkName, "", epoch, settings.ParticipationRateThreshold)
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsNewRewardRoundSubscribed, userId, types.RocketpoolNewClaimRoundStartedEventName, networkName, "", epoch, 0)

	// Insert all the events or update the threshold if they already exist
	if len(eventsToInsert) > 0 {
		insertDs := goqu.Dialect("postgres").
			Insert("users_subscriptions").
			Cols("user_id", "event_name", "event_filter", "created_ts", "created_epoch", "event_threshold").
			Rows(eventsToInsert).
			OnConflict(goqu.DoUpdate(
				"user_id, event_name, event_filter",
				goqu.Record{"event_threshold": goqu.L("EXCLUDED.event_threshold")},
			))

		query, args, err := insertDs.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %v", err)
		}

		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return err
		}
	}

	// Delete all the events
	if len(eventsToDelete) > 0 {
		deleteDs := goqu.Dialect("postgres").
			Delete("users_subscriptions").
			Where(goqu.Or(eventsToDelete...))

		query, args, err := deleteDs.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %v", err)
		}

		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing tx to update general notification settings: %w", err)
	}
	return nil
}
func (d *DataAccessService) UpdateNotificationSettingsPairedDevice(ctx context.Context, userId uint64, pairedDeviceId uint64, name string, IsNotificationsEnabled bool) error {
	result, err := d.userWriter.ExecContext(ctx, `
		UPDATE users_devices 
		SET 
			device_name = $1,
			notify_enabled = $2
		WHERE user_id = $3 AND id = $4`,
		name, IsNotificationsEnabled, userId, pairedDeviceId)
	if err != nil {
		return err
	}

	// TODO: This can be deleted when the API layer has an improved check for the device id
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("device with id %v to update notification settings not found", pairedDeviceId)
	}
	return nil
}
func (d *DataAccessService) DeleteNotificationSettingsPairedDevice(ctx context.Context, userId uint64, pairedDeviceId uint64) error {
	result, err := d.userWriter.ExecContext(ctx, `
		DELETE FROM users_devices 
		WHERE user_id = $1 AND id = $2`,
		userId, pairedDeviceId)
	if err != nil {
		return err
	}

	// TODO: This can be deleted when the API layer has an improved check for the device id
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("device with id %v to delete not found", pairedDeviceId)
	}
	return nil
}
func (d *DataAccessService) UpdateNotificationSettingsClients(ctx context.Context, userId uint64, clientId uint64, IsSubscribed bool) (*t.NotificationSettingsClient, error) {
	result := &t.NotificationSettingsClient{Id: clientId, IsSubscribed: IsSubscribed}

	var clientInfo *t.ClientInfo

	clients, err := d.GetAllClients()
	if err != nil {
		return nil, err
	}
	for _, client := range clients {
		if client.Id == clientId {
			clientInfo = &client
			break
		}
	}
	if clientInfo == nil {
		return nil, fmt.Errorf("client with id %d to update client notification settings not found", clientId)
	}

	if IsSubscribed {
		_, err = d.userWriter.ExecContext(ctx, `
			INSERT INTO users_subscriptions (user_id, event_name, event_filter, created_ts, created_epoch)
				VALUES ($1, $2, $3, NOW(), $4)
			ON CONFLICT (user_id, event_name, event_filter) 
				DO NOTHING`,
			userId, types.EthClientUpdateEventName, clientInfo.DbName, utils.TimeToEpoch(time.Now()))
	} else {
		_, err = d.userWriter.ExecContext(ctx, `DELETE FROM users_subscriptions WHERE user_id = $1 AND event_name = $2 AND event_filter = $3`,
			userId, types.EthClientUpdateEventName, clientInfo.DbName)
	}
	if err != nil {
		return nil, err
	}

	result.Name = clientInfo.Name
	result.Category = clientInfo.Category

	return result, nil
}
func (d *DataAccessService) GetNotificationSettingsDashboards(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationSettingsDashboardColumn], search string, limit uint64) ([]t.NotificationSettingsDashboardsTableRow, *t.Paging, error) {
	result := make([]t.NotificationSettingsDashboardsTableRow, 0)
	var paging t.Paging

	wg := errgroup.Group{}

	// Initialize the cursor
	var currentCursor t.NotificationSettingsCursor
	var err error
	if cursor != "" {
		currentCursor, err = utils.StringToCursor[t.NotificationSettingsCursor](cursor)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse passed cursor as NotificationSettingsCursor: %w", err)
		}
	}

	isReverseDirection := (colSort.Desc && !currentCursor.IsReverse()) || (!colSort.Desc && currentCursor.IsReverse())

	// -------------------------------------
	// Get the events
	events := []struct {
		Name      types.EventName `db:"event_name"`
		Filter    string          `db:"event_filter"`
		Threshold float64         `db:"event_threshold"`
	}{}
	wg.Go(func() error {
		err := d.userReader.SelectContext(ctx, &events, `
			SELECT
				event_name,
				event_filter,
				event_threshold
			FROM users_subscriptions
			WHERE user_id = $1`, userId)
		if err != nil {
			return fmt.Errorf(`error retrieving data for account dashboard notifications: %w`, err)
		}

		return nil
	})

	// -------------------------------------
	// Get the validator dashboards
	valDashboards := []struct {
		DashboardId   uint64         `db:"dashboard_id"`
		DashboardName string         `db:"dashboard_name"`
		GroupId       uint64         `db:"group_id"`
		GroupName     string         `db:"group_name"`
		Network       uint64         `db:"network"`
		WebhookUrl    sql.NullString `db:"webhook_target"`
		WebhookFormat sql.NullString `db:"webhook_format"`
	}{}
	wg.Go(func() error {
		err := d.alloyReader.SelectContext(ctx, &valDashboards, `
			SELECT
				d.id AS dashboard_id,
				d.name AS dashboard_name,
				g.id AS group_id,
				g.name AS group_name,
				d.network,
				g.webhook_target,
				g.webhook_format
			FROM users_val_dashboards d
			INNER JOIN users_val_dashboards_groups g ON d.id = g.dashboard_id
			WHERE d.user_id = $1`, userId)
		if err != nil {
			return fmt.Errorf(`error retrieving data for validator dashboard notifications: %w`, err)
		}

		return nil
	})

	// -------------------------------------
	// Get the account dashboards
	accDashboards := []struct {
		DashboardId                     uint64         `db:"dashboard_id"`
		DashboardName                   string         `db:"dashboard_name"`
		GroupId                         uint64         `db:"group_id"`
		GroupName                       string         `db:"group_name"`
		WebhookUrl                      sql.NullString `db:"webhook_target"`
		WebhookFormat                   sql.NullString `db:"webhook_format"`
		IsIgnoreSpamTransactionsEnabled bool           `db:"ignore_spam_transactions"`
		SubscribedChainIds              []uint64       `db:"subscribed_chain_ids"`
	}{}
	// TODO: Account dashboard handling will be handled later
	// wg.Go(func() error {
	// 	err := d.alloyReader.SelectContext(ctx, &accDashboards, `
	// 		SELECT
	// 			d.id AS dashboard_id,
	// 			d.name AS dashboard_name,
	// 			g.id AS group_id,
	// 			g.name AS group_name,
	// 			g.webhook_target,
	// 			g.webhook_format,
	// 			g.ignore_spam_transactions,
	// 			g.subscribed_chain_ids
	// 		FROM users_acc_dashboards d
	// 		INNER JOIN users_acc_dashboards_groups g ON d.id = g.dashboard_id
	// 		WHERE d.user_id = $1`, userId)
	// 	if err != nil {
	// 		return fmt.Errorf(`error retrieving data for validator dashboard notifications: %w`, err)
	// 	}

	// 	return nil
	// })

	err = wg.Wait()
	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving dashboard notification data: %w", err)
	}

	// -------------------------------------
	// Evaluate the data
	resultMap := make(map[string]*t.NotificationSettingsDashboardsTableRow)

	for _, event := range events {
		eventFilterSplit := strings.Split(event.Filter, ":")
		if len(eventFilterSplit) != 3 {
			continue
		}
		dashboardType := eventFilterSplit[0]

		eventNameSplit := strings.Split(string(event.Name), ":")
		if len(eventNameSplit) != 2 && dashboardType == ValidatorDashboardEventPrefix {
			return nil, nil, fmt.Errorf("invalid event name formatting for val dashboard notification: expected {network:event_name}, got %v", event.Name)
		}

		eventName := event.Name
		if len(eventNameSplit) == 2 {
			eventName = types.EventName(eventNameSplit[1])
		}

		if _, ok := resultMap[event.Filter]; !ok {
			if dashboardType == ValidatorDashboardEventPrefix {
				resultMap[event.Filter] = &t.NotificationSettingsDashboardsTableRow{
					Settings: t.NotificationSettingsValidatorDashboard{
						GroupEfficiencyBelowThreshold: GroupEfficiencyBelowThresholdDefault,
						MaxCollateralThreshold:        MaxCollateralThresholdDefault,
						MinCollateralThreshold:        MinCollateralThresholdDefault,
					},
				}
			} else if dashboardType == AccountDashboardEventPrefix {
				resultMap[event.Filter] = &t.NotificationSettingsDashboardsTableRow{
					Settings: t.NotificationSettingsAccountDashboard{
						ERC20TokenTransfersValueThreshold: ERC20TokenTransfersValueThresholdDefault,
					},
				}
			}
		}

		switch settings := resultMap[event.Filter].Settings.(type) {
		case t.NotificationSettingsValidatorDashboard:
			switch eventName {
			case types.ValidatorIsOfflineEventName:
				settings.IsValidatorOfflineSubscribed = true
			case types.ValidatorGroupEfficiencyEventName:
				settings.IsGroupEfficiencyBelowSubscribed = true
				settings.GroupEfficiencyBelowThreshold = event.Threshold
			case types.ValidatorMissedAttestationEventName:
				settings.IsAttestationsMissedSubscribed = true
			case types.ValidatorMissedProposalEventName, types.ValidatorExecutedProposalEventName:
				settings.IsBlockProposalSubscribed = true
			case types.ValidatorUpcomingProposalEventName:
				settings.IsUpcomingBlockProposalSubscribed = true
			case types.SyncCommitteeSoon:
				settings.IsSyncSubscribed = true
			case types.ValidatorReceivedWithdrawalEventName:
				settings.IsWithdrawalProcessedSubscribed = true
			case types.ValidatorGotSlashedEventName:
				settings.IsSlashedSubscribed = true
			case types.RocketpoolCollateralMinReachedEventName:
				settings.IsMinCollateralSubscribed = true
				settings.MinCollateralThreshold = event.Threshold
			case types.RocketpoolCollateralMaxReachedEventName:
				settings.IsMaxCollateralSubscribed = true
				settings.MaxCollateralThreshold = event.Threshold
			}
			resultMap[event.Filter].Settings = settings
		case t.NotificationSettingsAccountDashboard:
			switch eventName {
			case types.IncomingTransactionEventName:
				settings.IsIncomingTransactionsSubscribed = true
			case types.OutgoingTransactionEventName:
				settings.IsOutgoingTransactionsSubscribed = true
			case types.ERC20TokenTransferEventName:
				settings.IsERC20TokenTransfersSubscribed = true
				settings.ERC20TokenTransfersValueThreshold = event.Threshold
			case types.ERC721TokenTransferEventName:
				settings.IsERC721TokenTransfersSubscribed = true
			case types.ERC1155TokenTransferEventName:
				settings.IsERC1155TokenTransfersSubscribed = true
			}
			resultMap[event.Filter].Settings = settings
		}
	}

	// Validator dashboards
	for _, valDashboard := range valDashboards {
		key := fmt.Sprintf("%s:%d:%d", ValidatorDashboardEventPrefix, valDashboard.DashboardId, valDashboard.GroupId)

		if _, ok := resultMap[key]; !ok {
			resultMap[key] = &t.NotificationSettingsDashboardsTableRow{
				Settings: t.NotificationSettingsValidatorDashboard{
					GroupEfficiencyBelowThreshold: GroupEfficiencyBelowThresholdDefault,
					MaxCollateralThreshold:        MaxCollateralThresholdDefault,
					MinCollateralThreshold:        MinCollateralThresholdDefault,
				},
			}
		}

		// Set general info
		resultMap[key].IsAccountDashboard = false
		resultMap[key].DashboardId = valDashboard.DashboardId
		resultMap[key].DashboardName = valDashboard.DashboardName
		resultMap[key].GroupId = valDashboard.GroupId
		resultMap[key].GroupName = valDashboard.GroupName
		resultMap[key].ChainIds = []uint64{valDashboard.Network}

		// Set the settings
		if valSettings, ok := resultMap[key].Settings.(t.NotificationSettingsValidatorDashboard); ok {
			valSettings.WebhookUrl = valDashboard.WebhookUrl.String
			valSettings.IsWebhookDiscordEnabled = valDashboard.WebhookFormat.Valid &&
				types.NotificationChannel(valDashboard.WebhookFormat.String) == types.WebhookDiscordNotificationChannel

			resultMap[key].Settings = valSettings
		}
	}

	// Account dashboards
	for _, accDashboard := range accDashboards {
		key := fmt.Sprintf("%s:%d:%d", AccountDashboardEventPrefix, accDashboard.DashboardId, accDashboard.GroupId)

		if _, ok := resultMap[key]; !ok {
			resultMap[key] = &t.NotificationSettingsDashboardsTableRow{
				Settings: t.NotificationSettingsAccountDashboard{
					ERC20TokenTransfersValueThreshold: ERC20TokenTransfersValueThresholdDefault,
				},
			}
		}

		// Set general info
		resultMap[key].IsAccountDashboard = true
		resultMap[key].DashboardId = accDashboard.DashboardId
		resultMap[key].DashboardName = accDashboard.DashboardName
		resultMap[key].GroupId = accDashboard.GroupId
		resultMap[key].GroupName = accDashboard.GroupName
		resultMap[key].ChainIds = accDashboard.SubscribedChainIds

		// Set the settings
		if accSettings, ok := resultMap[key].Settings.(t.NotificationSettingsAccountDashboard); ok {
			accSettings.WebhookUrl = accDashboard.WebhookUrl.String
			accSettings.IsWebhookDiscordEnabled = accDashboard.WebhookFormat.Valid &&
				types.NotificationChannel(accDashboard.WebhookFormat.String) == types.WebhookDiscordNotificationChannel
			accSettings.IsIgnoreSpamTransactionsEnabled = accDashboard.IsIgnoreSpamTransactionsEnabled
			accSettings.SubscribedChainIds = accDashboard.SubscribedChainIds

			resultMap[key].Settings = accSettings
		}
	}

	// Apply filter
	if search != "" {
		lowerSearch := strings.ToLower(search)
		for key, resultEntry := range resultMap {
			if !strings.HasPrefix(strings.ToLower(resultEntry.DashboardName), lowerSearch) &&
				!strings.HasPrefix(strings.ToLower(resultEntry.GroupName), lowerSearch) {
				delete(resultMap, key)
			}
		}
	}

	// Convert to a slice for sorting and paging
	for _, resultEntry := range resultMap {
		result = append(result, *resultEntry)
	}

	// -------------------------------------
	// Sort
	// Each row is uniquely defined by the dashboardId, groupId, and isAccountDashboard so the sort order is DashboardName/GroupName => DashboardId => GroupId => IsAccountDashboard
	var primarySortParam func(resultEntry t.NotificationSettingsDashboardsTableRow) string
	switch colSort.Column {
	case enums.NotificationSettingsDashboardColumns.DashboardName:
		primarySortParam = func(resultEntry t.NotificationSettingsDashboardsTableRow) string { return resultEntry.DashboardName }
	case enums.NotificationSettingsDashboardColumns.GroupName:
		primarySortParam = func(resultEntry t.NotificationSettingsDashboardsTableRow) string { return resultEntry.GroupName }
	default:
		return nil, nil, fmt.Errorf("invalid sort column for notification subscriptions: %v", colSort.Column)
	}
	sort.Slice(result, func(i, j int) bool {
		if isReverseDirection {
			if primarySortParam(result[i]) == primarySortParam(result[j]) {
				if result[i].DashboardId == result[j].DashboardId {
					if result[i].GroupId == result[j].GroupId {
						return result[i].IsAccountDashboard
					}
					return result[i].GroupId > result[j].GroupId
				}
				return result[i].DashboardId > result[j].DashboardId
			}
			return primarySortParam(result[i]) > primarySortParam(result[j])
		} else {
			if primarySortParam(result[i]) == primarySortParam(result[j]) {
				if result[i].DashboardId == result[j].DashboardId {
					if result[i].GroupId == result[j].GroupId {
						return result[j].IsAccountDashboard
					}
					return result[i].GroupId < result[j].GroupId
				}
				return result[i].DashboardId < result[j].DashboardId
			}
			return primarySortParam(result[i]) < primarySortParam(result[j])
		}
	})

	// -------------------------------------
	// Paging

	// Find the index for the cursor and limit the data
	if currentCursor.IsValid() {
		for idx, row := range result {
			if row.DashboardId == currentCursor.DashboardId &&
				row.GroupId == currentCursor.GroupId &&
				row.IsAccountDashboard == currentCursor.IsAccountDashboard {
				result = result[idx+1:]
				break
			}
		}
	}

	// Flag if above limit
	moreDataFlag := len(result) > int(limit)
	if !moreDataFlag && !currentCursor.IsValid() {
		// No paging required
		return result, &paging, nil
	}

	// Remove the last entries from data
	if moreDataFlag {
		result = result[:limit]
	}

	if currentCursor.IsReverse() {
		slices.Reverse(result)
	}

	p, err := utils.GetPagingFromData(result, currentCursor, moreDataFlag)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get paging: %w", err)
	}

	return result, p, nil
}
func (d *DataAccessService) UpdateNotificationSettingsValidatorDashboard(ctx context.Context, userId uint64, dashboardId t.VDBIdPrimary, groupId uint64, settings t.NotificationSettingsValidatorDashboard) error {
	// For the given dashboardId and groupId update users_subscriptions and users_val_dashboards_groups with the given settings
	epoch := utils.TimeToEpoch(time.Now())

	var eventsToInsert []goqu.Record
	var eventsToDelete []goqu.Expression

	// Get the network for the validator dashboard
	var chainId uint64
	err := d.alloyReader.GetContext(ctx, &chainId, `SELECT network FROM users_val_dashboards WHERE id = $1 AND user_id = $2`, dashboardId, userId)
	if err != nil {
		return fmt.Errorf("error getting network for validator dashboard: %w", err)
	}

	networks, err := d.GetAllNetworks()
	if err != nil {
		return err
	}

	networkName := ""
	for _, network := range networks {
		if network.ChainId == chainId {
			networkName = network.NotificationsName
			break
		}
	}
	if networkName == "" {
		return fmt.Errorf("network with chain id %d to update general notification settings not found", chainId)
	}

	// Add and remove the events in users_subscriptions
	tx, err := d.userWriter.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting db transactions to update validator dashboard notification settings: %w", err)
	}
	defer utils.Rollback(tx)

	eventFilter := fmt.Sprintf("%s:%d:%d", ValidatorDashboardEventPrefix, dashboardId, groupId)

	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsValidatorOfflineSubscribed, userId, types.ValidatorIsOfflineEventName, networkName, eventFilter, epoch, 0)
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsGroupEfficiencyBelowSubscribed, userId, types.ValidatorGroupEfficiencyEventName, networkName, eventFilter, epoch, settings.GroupEfficiencyBelowThreshold)
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsAttestationsMissedSubscribed, userId, types.ValidatorMissedAttestationEventName, networkName, eventFilter, epoch, 0)
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsUpcomingBlockProposalSubscribed, userId, types.ValidatorUpcomingProposalEventName, networkName, eventFilter, epoch, 0)
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsSyncSubscribed, userId, types.SyncCommitteeSoon, networkName, eventFilter, epoch, 0)
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsWithdrawalProcessedSubscribed, userId, types.ValidatorReceivedWithdrawalEventName, networkName, eventFilter, epoch, 0)
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsSlashedSubscribed, userId, types.ValidatorGotSlashedEventName, networkName, eventFilter, epoch, 0)
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsMaxCollateralSubscribed, userId, types.RocketpoolCollateralMaxReachedEventName, networkName, eventFilter, epoch, settings.MaxCollateralThreshold)
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsMinCollateralSubscribed, userId, types.RocketpoolCollateralMinReachedEventName, networkName, eventFilter, epoch, settings.MinCollateralThreshold)
	// Set two events for IsBlockProposalSubscribed
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsBlockProposalSubscribed, userId, types.ValidatorMissedProposalEventName, networkName, eventFilter, epoch, 0)
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsBlockProposalSubscribed, userId, types.ValidatorExecutedProposalEventName, networkName, eventFilter, epoch, 0)

	// Insert all the events or update the threshold if they already exist
	if len(eventsToInsert) > 0 {
		insertDs := goqu.Dialect("postgres").
			Insert("users_subscriptions").
			Cols("user_id", "event_name", "event_filter", "created_ts", "created_epoch", "event_threshold").
			Rows(eventsToInsert).
			OnConflict(goqu.DoUpdate(
				"user_id, event_name, event_filter",
				goqu.Record{"event_threshold": goqu.L("EXCLUDED.event_threshold")},
			))

		query, args, err := insertDs.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %v", err)
		}

		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return err
		}
	}

	// Delete all the events
	if len(eventsToDelete) > 0 {
		deleteDs := goqu.Dialect("postgres").
			Delete("users_subscriptions").
			Where(goqu.Or(eventsToDelete...))

		query, args, err := deleteDs.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %v", err)
		}

		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing tx to update validator dashboard notification settings: %w", err)
	}

	// Set non-event settings
	var webhookFormat sql.NullString
	if settings.WebhookUrl != "" {
		webhookFormat.String = string(types.WebhookNotificationChannel)
		webhookFormat.Valid = true
		if settings.IsWebhookDiscordEnabled {
			webhookFormat.String = string(types.WebhookDiscordNotificationChannel)
		}
	}

	_, err = d.alloyWriter.ExecContext(ctx, `
		UPDATE users_val_dashboards_groups 
		SET 
			webhook_target = NULLIF($1, ''),
			webhook_format = $2
		WHERE dashboard_id = $3 AND id = $4`, settings.WebhookUrl, webhookFormat, dashboardId, groupId)
	if err != nil {
		return err
	}

	return nil
}
func (d *DataAccessService) UpdateNotificationSettingsAccountDashboard(ctx context.Context, userId uint64, dashboardId t.VDBIdPrimary, groupId uint64, settings t.NotificationSettingsAccountDashboard) error {
	// TODO: Account dashboard handling will be handled later
	// // For the given dashboardId and groupId update users_subscriptions and users_acc_dashboards_groups with the given settings
	// epoch := utils.TimeToEpoch(time.Now())

	// var eventsToInsert []goqu.Record
	// var eventsToDelete []goqu.Expression

	// tx, err := d.userWriter.BeginTxx(ctx, nil)
	// if err != nil {
	// 	return fmt.Errorf("error starting db transactions to update validator dashboard notification settings: %w", err)
	// }
	// defer utils.Rollback(tx)

	// eventFilter := fmt.Sprintf("%s:%d:%d", AccountDashboardEventPrefix, dashboardId, groupId)

	// d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsIncomingTransactionsSubscribed, userId, types.IncomingTransactionEventName, "", eventFilter, epoch, 0)
	// d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsOutgoingTransactionsSubscribed, userId, types.OutgoingTransactionEventName, "", eventFilter, epoch, 0)
	// d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsERC20TokenTransfersSubscribed, userId, types.ERC20TokenTransferEventName, "", eventFilter, epoch, settings.ERC20TokenTransfersValueThreshold)
	// d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsERC721TokenTransfersSubscribed, userId, types.ERC721TokenTransferEventName, "", eventFilter, epoch, 0)
	// d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsERC1155TokenTransfersSubscribed, userId, types.ERC1155TokenTransferEventName, "", eventFilter, epoch, 0)

	// // Insert all the events or update the threshold if they already exist
	// if len(eventsToInsert) > 0 {
	// 	insertDs := goqu.Dialect("postgres").
	// 		Insert("users_subscriptions").
	// 		Cols("user_id", "event_name", "event_filter", "created_ts", "created_epoch", "event_threshold").
	// 		Rows(eventsToInsert).
	// 		OnConflict(goqu.DoUpdate(
	// 			"user_id, event_name, event_filter",
	// 			goqu.Record{"event_threshold": goqu.L("EXCLUDED.event_threshold")},
	// 		))

	// 	query, args, err := insertDs.Prepared(true).ToSQL()
	// 	if err != nil {
	// 		return fmt.Errorf("error preparing query: %v", err)
	// 	}

	// 	_, err = tx.ExecContext(ctx, query, args...)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	// // Delete all the events
	// if len(eventsToDelete) > 0 {
	// 	deleteDs := goqu.Dialect("postgres").
	// 		Delete("users_subscriptions").
	// 		Where(goqu.Or(eventsToDelete...))

	// 	query, args, err := deleteDs.Prepared(true).ToSQL()
	// 	if err != nil {
	// 		return fmt.Errorf("error preparing query: %v", err)
	// 	}

	// 	_, err = tx.ExecContext(ctx, query, args...)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	// err = tx.Commit()
	// if err != nil {
	// 	return fmt.Errorf("error committing tx to update validator dashboard notification settings: %w", err)
	// }

	// // Set non-event settings
	// var webhookFormat sql.NullString
	// if settings.WebhookUrl != "" {
	// 	webhookFormat.String = string(types.WebhookNotificationChannel)
	// 	webhookFormat.Valid = true
	// 	if settings.IsWebhookDiscordEnabled {
	// 		webhookFormat.String = string(types.WebhookDiscordNotificationChannel)
	// 	}
	// }

	// _, err = d.alloyWriter.ExecContext(ctx, `
	// 	UPDATE users_acc_dashboards_groups
	// 	SET
	// 		webhook_target = NULLIF($1, ''),
	// 		webhook_format = $2,
	// 		ignore_spam_transactions = $3,
	// 		subscribed_chain_ids = $4
	// 	WHERE dashboard_id = $5 AND id = $6`, settings.WebhookUrl, webhookFormat, settings.IsIgnoreSpamTransactionsEnabled, settings.SubscribedChainIds, dashboardId, groupId)
	// if err != nil {
	// 	return err
	// }

	return d.dummy.UpdateNotificationSettingsAccountDashboard(ctx, userId, dashboardId, groupId, settings)
}

func (d *DataAccessService) AddOrRemoveEvent(eventsToInsert *[]goqu.Record, eventsToDelete *[]goqu.Expression, isSubscribed bool, userId uint64, eventName types.EventName, network, eventFilter string, epoch int64, threshold float64) {
	fullEventName := string(eventName)
	if network != "" {
		fullEventName = fmt.Sprintf("%s:%s", network, eventName)
	}

	if isSubscribed {
		event := goqu.Record{"user_id": userId, "event_name": fullEventName, "event_filter": eventFilter, "created_ts": goqu.L("NOW()"), "created_epoch": epoch, "event_threshold": threshold}
		*eventsToInsert = append(*eventsToInsert, event)
	} else {
		*eventsToDelete = append(*eventsToDelete, goqu.Ex{"user_id": userId, "event_name": fullEventName, "event_filter": eventFilter})
	}
}

func (d *DataAccessService) QueueTestEmailNotification(ctx context.Context, userId uint64) error {
	return notification.SendTestEmail(ctx, types.UserId(userId), d.userReader)
}
func (d *DataAccessService) QueueTestPushNotification(ctx context.Context, userId uint64) error {
	return notification.QueueTestPushNotification(ctx, types.UserId(userId), d.userReader, d.readerDb)
}
func (d *DataAccessService) QueueTestWebhookNotification(ctx context.Context, userId uint64, webhookUrl string, isDiscordWebhook bool) error {
	return notification.SendTestWebhookNotification(ctx, types.UserId(userId), webhookUrl, isDiscordWebhook)
}
