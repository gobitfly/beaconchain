package dataaccess

import (
	"context"
	"database/sql"
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/ethereum/go-ethereum/params"
	"github.com/go-redis/redis/v8"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
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
	UpdateNotificationSettingsGeneral(ctx context.Context, userId uint64, settings t.NotificationSettingsGeneral) error
	UpdateNotificationSettingsNetworks(ctx context.Context, userId uint64, chainId uint64, settings t.NotificationSettingsNetwork) error
	UpdateNotificationSettingsPairedDevice(ctx context.Context, userId uint64, pairedDeviceId string, name string, IsNotificationsEnabled bool) error
	DeleteNotificationSettingsPairedDevice(ctx context.Context, userId uint64, pairedDeviceId string) error
	UpdateNotificationSettingsClients(ctx context.Context, userId uint64, clientId uint64, IsSubscribed bool) (*t.NotificationSettingsClient, error)
	GetNotificationSettingsDashboards(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationSettingsDashboardColumn], search string, limit uint64) ([]t.NotificationSettingsDashboardsTableRow, *t.Paging, error)
	UpdateNotificationSettingsValidatorDashboard(ctx context.Context, userId uint64, dashboardId t.VDBIdPrimary, groupId uint64, settings t.NotificationSettingsValidatorDashboard) error
	UpdateNotificationSettingsAccountDashboard(ctx context.Context, userId uint64, dashboardId t.VDBIdPrimary, groupId uint64, settings t.NotificationSettingsAccountDashboard) error
}

const (
	ValidatorDashboardEventPrefix string = "vdb"
	AccountDashboardEventPrefix   string = "adb"

	DiscordWebhookFormat string = "discord"

	GroupOfflineThresholdDefault             float64 = 0.1
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
	latestSlot, err := d.GetLatestSlot()
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
			Select("name").
			From(query.As("history")).
			LeftJoin(goqu.I(groupsTable).As("groups"), goqu.On(
				goqu.Ex{"groups.dashboard_id": goqu.I("history.dashboard_id")},
				goqu.Ex{"groups.id": goqu.I("history.group_id")},
			))

		mostNotifiedGroups := [3]string{}
		querySql, args, err := query.Prepared(true).ToSQL()
		if err != nil {
			return mostNotifiedGroups, err
		}
		res := []string{}
		err = d.alloyReader.SelectContext(ctx, &res, querySql, args...)
		if err != nil {
			return mostNotifiedGroups, err
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
			res := d.persistentRedisDbClient.Get(ctx, fmt.Sprintf("%s:%d:%d", prefix, userId, day))
			if res.Err() == redis.Nil {
				return 0, nil
			} else if res.Err() != nil {
				return 0, res.Err()
			}
			return res.Uint64()
		}
		response.Last24hPushCount, err = getMessageCount("n_mails")
		if err != nil {
			return err
		}
		response.Last24hPushCount, err = getMessageCount("n_push")
		if err != nil {
			return err
		}
		response.Last24hPushCount, err = getMessageCount("n_webhook")
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
			whereNetwork += "event_name like '" + network.Name + ":rocketpool_%' OR event_name like '" + network.Name + ":network_%'"
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

		err = d.alloyReader.GetContext(ctx, &response, querySql, args...)
		return err
	})

	err = eg.Wait()
	return &response, err
}
func (d *DataAccessService) GetDashboardNotifications(ctx context.Context, userId uint64, chainIds []uint64, cursor string, colSort t.Sort[enums.NotificationDashboardsColumn], search string, limit uint64) ([]t.NotificationDashboardsTableRow, *t.Paging, error) {
	return d.dummy.GetDashboardNotifications(ctx, userId, chainIds, cursor, colSort, search, limit)
}

func (d *DataAccessService) GetValidatorDashboardNotificationDetails(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, epoch uint64, search string) (*t.NotificationValidatorDashboardDetail, error) {
	return d.dummy.GetValidatorDashboardNotificationDetails(ctx, dashboardId, groupId, epoch, search)
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

	isReverseDirection := (colSort.Desc && !currentCursor.IsReverse()) || (!colSort.Desc && currentCursor.IsReverse())
	sortSearchDirection := ">"
	if isReverseDirection {
		sortSearchDirection = "<"
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
	// Rows can be uniquely identified by (epoch, machine_id, event_type)
	sortDirFunc := func(column string) exp.OrderedExpression {
		return goqu.I(column).Asc()
	}
	if isReverseDirection {
		sortDirFunc = func(column string) exp.OrderedExpression {
			return goqu.I(column).Desc()
		}
	}
	switch colSort.Column {
	case enums.NotificationsMachinesColumns.MachineName:
		if currentCursor.IsValid() {
			ds = ds.Where(goqu.Or(
				goqu.L(fmt.Sprintf("(machine_name %s ?)", sortSearchDirection), currentCursor.MachineName),
				goqu.L(fmt.Sprintf("(machine_name = ? AND epoch %s ?)", sortSearchDirection), currentCursor.MachineName, currentCursor.Epoch),
				goqu.L(fmt.Sprintf("(machine_name = ? AND epoch = ? AND machine_id %s ?)", sortSearchDirection), currentCursor.MachineName, currentCursor.Epoch, currentCursor.MachineId),
				goqu.L(fmt.Sprintf("(machine_name = ? AND epoch = ? AND machine_id = ? AND event_type %s ?)", sortSearchDirection), currentCursor.MachineName, currentCursor.Epoch, currentCursor.MachineId, currentCursor.EventType),
			))
		}
		ds = ds.Order(
			sortDirFunc("machine_name"),
			sortDirFunc("epoch"),
			sortDirFunc("machine_id"),
			sortDirFunc("event_type"))
	case enums.NotificationsMachinesColumns.Threshold:
		if currentCursor.IsValid() {
			ds = ds.Where(goqu.Or(
				goqu.L(fmt.Sprintf("(event_threshold %s ?)", sortSearchDirection), currentCursor.EventThreshold),
				goqu.L(fmt.Sprintf("(event_threshold = ? AND epoch %s ?)", sortSearchDirection), currentCursor.EventThreshold, currentCursor.Epoch),
				goqu.L(fmt.Sprintf("(event_threshold = ? AND epoch = ? AND machine_id %s ?)", sortSearchDirection), currentCursor.EventThreshold, currentCursor.Epoch, currentCursor.MachineId),
				goqu.L(fmt.Sprintf("(event_threshold = ? AND epoch = ? AND machine_id = ? AND event_type %s ?)", sortSearchDirection), currentCursor.EventThreshold, currentCursor.Epoch, currentCursor.MachineId, currentCursor.EventType),
			))
		}
		ds = ds.Order(
			sortDirFunc("event_threshold"),
			sortDirFunc("epoch"),
			sortDirFunc("machine_id"),
			sortDirFunc("event_type"))
	case enums.NotificationsMachinesColumns.EventType:
		if currentCursor.IsValid() {
			ds = ds.Where(goqu.Or(
				goqu.L(fmt.Sprintf("(event_type %s ?)", sortSearchDirection), currentCursor.EventType),
				goqu.L(fmt.Sprintf("(event_type = ? AND epoch %s ?)", sortSearchDirection), currentCursor.EventType, currentCursor.Epoch),
				goqu.L(fmt.Sprintf("(event_type = ? AND epoch = ? AND machine_id %s ?)", sortSearchDirection), currentCursor.EventType, currentCursor.Epoch, currentCursor.MachineId),
			))
		}
		ds = ds.Order(
			sortDirFunc("event_type"),
			sortDirFunc("epoch"),
			sortDirFunc("machine_id"))
	case enums.NotificationsMachinesColumns.Timestamp:
		if currentCursor.IsValid() {
			ds = ds.Where(goqu.Or(
				goqu.L(fmt.Sprintf("(epoch %s ?)", sortSearchDirection), currentCursor.Epoch),
				goqu.L(fmt.Sprintf("(epoch = ? AND machine_id %s ?)", sortSearchDirection), currentCursor.Epoch, currentCursor.MachineId),
				goqu.L(fmt.Sprintf("(epoch = ? AND machine_id = ? AND event_type %s ?)", sortSearchDirection), currentCursor.Epoch, currentCursor.MachineId, currentCursor.EventType),
			))
		}
		ds = ds.Order(
			sortDirFunc("epoch"),
			sortDirFunc("machine_id"),
			sortDirFunc("event_type"))
	default:
		return nil, nil, fmt.Errorf("invalid column for sorting of machine notification history: %v", colSort.Column)
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

	isReverseDirection := (colSort.Desc && !currentCursor.IsReverse()) || (!colSort.Desc && currentCursor.IsReverse())
	sortSearchDirection := ">"
	if isReverseDirection {
		sortSearchDirection = "<"
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
	sortDirFunc := func(column string) exp.OrderedExpression {
		return goqu.I(column).Asc()
	}
	if isReverseDirection {
		sortDirFunc = func(column string) exp.OrderedExpression {
			return goqu.I(column).Desc()
		}
	}
	switch colSort.Column {
	case enums.NotificationsClientsColumns.ClientName:
		if currentCursor.IsValid() {
			ds = ds.Where(goqu.Or(
				goqu.L(fmt.Sprintf("(client %s ?)", sortSearchDirection), currentCursor.Client),
				goqu.L(fmt.Sprintf("(client = ? AND epoch %s ?)", sortSearchDirection), currentCursor.Client, currentCursor.Epoch),
			))
		}
		ds = ds.Order(
			sortDirFunc("client"),
			sortDirFunc("epoch"))
	case enums.NotificationsClientsColumns.Timestamp:
		if currentCursor.IsValid() {
			ds = ds.Where(goqu.Or(
				goqu.L(fmt.Sprintf("(epoch %s ?)", sortSearchDirection), currentCursor.Epoch),
				goqu.L(fmt.Sprintf("(epoch = ? AND client %s ?)", sortSearchDirection), currentCursor.Epoch, currentCursor.Client),
			))
		}
		ds = ds.Order(
			sortDirFunc("epoch"),
			sortDirFunc("client"))
	default:
		return nil, nil, fmt.Errorf("invalid column for sorting of client notification history: %v", colSort.Column)
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
	// 	case types.RocketpoolCollateralMinReached:
	// 		resultEntry.EventType = "collateral_min"
	// 	case types.RocketpoolCollateralMaxReached:
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

	isReverseDirection := (colSort.Desc && !currentCursor.IsReverse()) || (!colSort.Desc && currentCursor.IsReverse())
	sortSearchDirection := ">"
	if isReverseDirection {
		sortSearchDirection = "<"
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
	sortDirFunc := func(column string) exp.OrderedExpression {
		return goqu.I(column).Asc()
	}
	if isReverseDirection {
		sortDirFunc = func(column string) exp.OrderedExpression {
			return goqu.I(column).Desc()
		}
	}
	switch colSort.Column {
	case enums.NotificationNetworksColumns.EventType:
		if currentCursor.IsValid() {
			ds = ds.Where(goqu.Or(
				goqu.L(fmt.Sprintf("(event_type %s ?)", sortSearchDirection), currentCursor.EventType),
				goqu.L(fmt.Sprintf("(event_type = ? AND epoch %s ?)", sortSearchDirection), currentCursor.EventType, currentCursor.Epoch),
				goqu.L(fmt.Sprintf("(event_type = ? AND epoch = ? AND network %s ?)", sortSearchDirection), currentCursor.EventType, currentCursor.Epoch, currentCursor.Network),
			))
		}
		ds = ds.Order(
			sortDirFunc("event_type"),
			sortDirFunc("epoch"),
			sortDirFunc("network"))
	case enums.NotificationNetworksColumns.Timestamp:
		if currentCursor.IsValid() {
			ds = ds.Where(goqu.Or(
				goqu.L(fmt.Sprintf("(epoch %s ?)", sortSearchDirection), currentCursor.Epoch),
				goqu.L(fmt.Sprintf("(epoch = ? AND network %s ?)", sortSearchDirection), currentCursor.Epoch, currentCursor.Network),
				goqu.L(fmt.Sprintf("(epoch = ? AND network = ? AND event_type %s ?)", sortSearchDirection), currentCursor.Epoch, currentCursor.Network, currentCursor.EventType),
			))
		}
		ds = ds.Order(
			sortDirFunc("epoch"),
			sortDirFunc("network"),
			sortDirFunc("event_type"))
	default:
		return nil, nil, fmt.Errorf("invalid column for sorting of network notification history: %v", colSort.Column)
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
		networksSettings[network.Name] = &t.NotificationNetwork{
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
		clientSettings[client.Name] = &t.NotificationSettingsClient{
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
		DeviceIdentifier sql.NullString `db:"device_identifier"`
		CreatedTs        time.Time      `db:"created_ts"`
		DeviceName       string         `db:"device_name"`
		NotifyEnabled    bool           `db:"notify_enabled"`
	}{}
	wg.Go(func() error {
		err := d.userReader.SelectContext(ctx, &pairedDevices, `
		SELECT
			device_identifier,
			created_ts,
			device_name,
			notify_enabled
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
				clientSettings[event.Filter].IsSubscribed = true
			}
		}
	}

	for _, settings := range networksSettings {
		result.Networks = append(result.Networks, *settings)
	}

	for _, device := range pairedDevices {
		result.PairedDevices = append(result.PairedDevices, t.NotificationPairedDevice{
			Id:                     device.DeviceIdentifier.String,
			PairedTimestamp:        device.CreatedTs.Unix(),
			Name:                   device.DeviceName,
			IsNotificationsEnabled: device.NotifyEnabled,
		})
	}

	for _, settings := range clientSettings {
		result.Clients = append(result.Clients, *settings)
	}

	return result, nil
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
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsMachineOfflineSubscribed, userId, string(types.MonitoringMachineOfflineEventName), "", epoch, 0)
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsMachineStorageUsageSubscribed, userId, string(types.MonitoringMachineDiskAlmostFullEventName), "", epoch, settings.MachineStorageUsageThreshold)
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsMachineCpuUsageSubscribed, userId, string(types.MonitoringMachineCpuLoadEventName), "", epoch, settings.MachineCpuUsageThreshold)
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsMachineMemoryUsageSubscribed, userId, string(types.MonitoringMachineMemoryUsageEventName), "", epoch, settings.MachineMemoryUsageThreshold)

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
			networkName = network.Name
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

	eventName := fmt.Sprintf("%s:%s", networkName, types.NetworkGasAboveThresholdEventName)
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsGasAboveSubscribed, userId, eventName, "", epoch, settings.GasAboveThreshold.Div(decimal.NewFromInt(params.GWei)).InexactFloat64())
	eventName = fmt.Sprintf("%s:%s", networkName, types.NetworkGasBelowThresholdEventName)
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsGasBelowSubscribed, userId, eventName, "", epoch, settings.GasBelowThreshold.Div(decimal.NewFromInt(params.GWei)).InexactFloat64())
	eventName = fmt.Sprintf("%s:%s", networkName, types.NetworkParticipationRateThresholdEventName)
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsParticipationRateSubscribed, userId, eventName, "", epoch, settings.ParticipationRateThreshold)
	eventName = fmt.Sprintf("%s:%s", networkName, types.RocketpoolNewClaimRoundStartedEventName)
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsNewRewardRoundSubscribed, userId, eventName, "", epoch, 0)

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
func (d *DataAccessService) UpdateNotificationSettingsPairedDevice(ctx context.Context, userId uint64, pairedDeviceId string, name string, IsNotificationsEnabled bool) error {
	result, err := d.userWriter.ExecContext(ctx, `
		UPDATE users_devices 
		SET 
			device_name = $1,
			notify_enabled = $2
		WHERE user_id = $3 AND device_identifier = $4`,
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
		return fmt.Errorf("device with id %s to update notification settings not found", pairedDeviceId)
	}
	return nil
}
func (d *DataAccessService) DeleteNotificationSettingsPairedDevice(ctx context.Context, userId uint64, pairedDeviceId string) error {
	result, err := d.userWriter.ExecContext(ctx, `
		DELETE FROM users_devices 
		WHERE user_id = $1 AND device_identifier = $2`,
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
		return fmt.Errorf("device with id %s to delete not found", pairedDeviceId)
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
			userId, types.EthClientUpdateEventName, clientInfo.Name, utils.TimeToEpoch(time.Now()))
	} else {
		_, err = d.userWriter.ExecContext(ctx, `DELETE FROM users_subscriptions WHERE user_id = $1 AND event_name = $2 AND event_filter = $3`,
			userId, types.EthClientUpdateEventName, clientInfo.Name)
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
		DashboardId             uint64         `db:"dashboard_id"`
		DashboardName           string         `db:"dashboard_name"`
		GroupId                 uint64         `db:"group_id"`
		GroupName               string         `db:"group_name"`
		Network                 uint64         `db:"network"`
		WebhookUrl              sql.NullString `db:"webhook_target"`
		IsWebhookDiscordEnabled sql.NullBool   `db:"discord_webhook"`
		IsRealTimeModeEnabled   sql.NullBool   `db:"realtime_notifications"`
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
				(g.webhook_format = $1) AS discord_webhook,
				g.realtime_notifications
			FROM users_val_dashboards d
			INNER JOIN users_val_dashboards_groups g ON d.id = g.dashboard_id
			WHERE d.user_id = $2`, DiscordWebhookFormat, userId)
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
		IsWebhookDiscordEnabled         sql.NullBool   `db:"discord_webhook"`
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
	// 			(g.webhook_format = $1) AS discord_webhook,
	// 			g.ignore_spam_transactions,
	// 			g.subscribed_chain_ids
	// 		FROM users_acc_dashboards d
	// 		INNER JOIN users_acc_dashboards_groups g ON d.id = g.dashboard_id
	// 		WHERE d.user_id = $2`, DiscordWebhookFormat, userId)
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
	type NotificationSettingsDashboardsInfo struct {
		IsAccountDashboard bool // if false it's a validator dashboard
		DashboardId        uint64
		DashboardName      string
		GroupId            uint64
		GroupName          string
		// if it's a validator dashboard, Settings is NotificationSettingsAccountDashboard, otherwise NotificationSettingsValidatorDashboard
		Settings interface{}
		ChainIds []uint64
	}
	settingsMap := make(map[string]*NotificationSettingsDashboardsInfo)

	for _, event := range events {
		eventSplit := strings.Split(event.Filter, ":")
		if len(eventSplit) != 3 {
			continue
		}
		dashboardType := eventSplit[0]

		if _, ok := settingsMap[event.Filter]; !ok {
			if dashboardType == ValidatorDashboardEventPrefix {
				settingsMap[event.Filter] = &NotificationSettingsDashboardsInfo{
					Settings: t.NotificationSettingsValidatorDashboard{
						GroupOfflineThreshold:  GroupOfflineThresholdDefault,
						MaxCollateralThreshold: MaxCollateralThresholdDefault,
						MinCollateralThreshold: MinCollateralThresholdDefault,
					},
				}
			} else if dashboardType == AccountDashboardEventPrefix {
				settingsMap[event.Filter] = &NotificationSettingsDashboardsInfo{
					Settings: t.NotificationSettingsAccountDashboard{
						ERC20TokenTransfersValueThreshold: ERC20TokenTransfersValueThresholdDefault,
					},
				}
			}
		}

		switch settings := settingsMap[event.Filter].Settings.(type) {
		case t.NotificationSettingsValidatorDashboard:
			switch event.Name {
			case types.ValidatorIsOfflineEventName:
				settings.IsValidatorOfflineSubscribed = true
			case types.GroupIsOfflineEventName:
				settings.IsGroupOfflineSubscribed = true
				settings.GroupOfflineThreshold = event.Threshold
			case types.ValidatorMissedAttestationEventName:
				settings.IsAttestationsMissedSubscribed = true
			case types.ValidatorProposalEventName:
				settings.IsBlockProposalSubscribed = true
			case types.ValidatorUpcomingProposalEventName:
				settings.IsUpcomingBlockProposalSubscribed = true
			case types.SyncCommitteeSoon:
				settings.IsSyncSubscribed = true
			case types.ValidatorReceivedWithdrawalEventName:
				settings.IsWithdrawalProcessedSubscribed = true
			case types.ValidatorGotSlashedEventName:
				settings.IsSlashedSubscribed = true
			case types.RocketpoolCollateralMinReached:
				settings.IsMinCollateralSubscribed = true
				settings.MinCollateralThreshold = event.Threshold
			case types.RocketpoolCollateralMaxReached:
				settings.IsMaxCollateralSubscribed = true
				settings.MaxCollateralThreshold = event.Threshold
			}
			settingsMap[event.Filter].Settings = settings
		case t.NotificationSettingsAccountDashboard:
			switch event.Name {
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
			settingsMap[event.Filter].Settings = settings
		}
	}

	// Validator dashboards
	for _, valDashboard := range valDashboards {
		key := fmt.Sprintf("%s:%d:%d", ValidatorDashboardEventPrefix, valDashboard.DashboardId, valDashboard.GroupId)

		if _, ok := settingsMap[key]; !ok {
			settingsMap[key] = &NotificationSettingsDashboardsInfo{
				Settings: t.NotificationSettingsValidatorDashboard{
					GroupOfflineThreshold:  GroupOfflineThresholdDefault,
					MaxCollateralThreshold: MaxCollateralThresholdDefault,
					MinCollateralThreshold: MinCollateralThresholdDefault,
				},
			}
		}

		// Set general info
		settingsMap[key].IsAccountDashboard = false
		settingsMap[key].DashboardId = valDashboard.DashboardId
		settingsMap[key].DashboardName = valDashboard.DashboardName
		settingsMap[key].GroupId = valDashboard.GroupId
		settingsMap[key].GroupName = valDashboard.GroupName
		settingsMap[key].ChainIds = []uint64{valDashboard.Network}

		// Set the settings
		if valSettings, ok := settingsMap[key].Settings.(*t.NotificationSettingsValidatorDashboard); ok {
			valSettings.WebhookUrl = valDashboard.WebhookUrl.String
			valSettings.IsWebhookDiscordEnabled = valDashboard.IsWebhookDiscordEnabled.Bool
			valSettings.IsRealTimeModeEnabled = valDashboard.IsRealTimeModeEnabled.Bool
		}
	}

	// Account dashboards
	for _, accDashboard := range accDashboards {
		key := fmt.Sprintf("%s:%d:%d", AccountDashboardEventPrefix, accDashboard.DashboardId, accDashboard.GroupId)

		if _, ok := settingsMap[key]; !ok {
			settingsMap[key] = &NotificationSettingsDashboardsInfo{
				Settings: t.NotificationSettingsAccountDashboard{
					ERC20TokenTransfersValueThreshold: ERC20TokenTransfersValueThresholdDefault,
				},
			}
		}

		// Set general info
		settingsMap[key].IsAccountDashboard = true
		settingsMap[key].DashboardId = accDashboard.DashboardId
		settingsMap[key].DashboardName = accDashboard.DashboardName
		settingsMap[key].GroupId = accDashboard.GroupId
		settingsMap[key].GroupName = accDashboard.GroupName
		settingsMap[key].ChainIds = accDashboard.SubscribedChainIds

		// Set the settings
		if accSettings, ok := settingsMap[key].Settings.(*t.NotificationSettingsAccountDashboard); ok {
			accSettings.WebhookUrl = accDashboard.WebhookUrl.String
			accSettings.IsWebhookDiscordEnabled = accDashboard.IsWebhookDiscordEnabled.Bool
			accSettings.IsIgnoreSpamTransactionsEnabled = accDashboard.IsIgnoreSpamTransactionsEnabled
			accSettings.SubscribedChainIds = accDashboard.SubscribedChainIds
		}
	}

	// Apply filter
	if search != "" {
		lowerSearch := strings.ToLower(search)
		for key, setting := range settingsMap {
			if !strings.HasPrefix(strings.ToLower(setting.DashboardName), lowerSearch) &&
				!strings.HasPrefix(strings.ToLower(setting.GroupName), lowerSearch) {
				delete(settingsMap, key)
			}
		}
	}

	// Convert to a slice for sorting and paging
	settings := slices.Collect(maps.Values(settingsMap))

	// -------------------------------------
	// Sort
	// Each row is uniquely defined by the dashboardId, groupId, and isAccountDashboard so the sort order is DashboardName/GroupName => DashboardId => GroupId => IsAccountDashboard
	var primarySortParam func(resultEntry *NotificationSettingsDashboardsInfo) string
	switch colSort.Column {
	case enums.NotificationSettingsDashboardColumns.DashboardName:
		primarySortParam = func(resultEntry *NotificationSettingsDashboardsInfo) string { return resultEntry.DashboardName }
	case enums.NotificationSettingsDashboardColumns.GroupName:
		primarySortParam = func(resultEntry *NotificationSettingsDashboardsInfo) string { return resultEntry.GroupName }
	default:
		return nil, nil, fmt.Errorf("invalid sort column for notification subscriptions: %v", colSort.Column)
	}
	sort.Slice(settings, func(i, j int) bool {
		if isReverseDirection {
			if primarySortParam(settings[i]) == primarySortParam(settings[j]) {
				if settings[i].DashboardId == settings[j].DashboardId {
					if settings[i].GroupId == settings[j].GroupId {
						return settings[i].IsAccountDashboard
					}
					return settings[i].GroupId > settings[j].GroupId
				}
				return settings[i].DashboardId > settings[j].DashboardId
			}
			return primarySortParam(settings[i]) > primarySortParam(settings[j])
		} else {
			if primarySortParam(settings[i]) == primarySortParam(settings[j]) {
				if settings[i].DashboardId == settings[j].DashboardId {
					if settings[i].GroupId == settings[j].GroupId {
						return settings[j].IsAccountDashboard
					}
					return settings[i].GroupId < settings[j].GroupId
				}
				return settings[i].DashboardId < settings[j].DashboardId
			}
			return primarySortParam(settings[i]) < primarySortParam(settings[j])
		}
	})

	// -------------------------------------
	// Convert to the final result format
	for _, setting := range settings {
		result = append(result, t.NotificationSettingsDashboardsTableRow{
			IsAccountDashboard: setting.IsAccountDashboard,
			DashboardId:        setting.DashboardId,
			GroupId:            setting.GroupId,
			GroupName:          setting.GroupName,
			Settings:           setting.Settings,
			ChainIds:           setting.ChainIds,
		})
	}

	// -------------------------------------
	// Paging

	// Find the index for the cursor and limit the data
	if currentCursor.IsValid() {
		for idx, row := range settings {
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

	tx, err := d.userWriter.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting db transactions to update validator dashboard notification settings: %w", err)
	}
	defer utils.Rollback(tx)

	eventFilter := fmt.Sprintf("%s:%d:%d", ValidatorDashboardEventPrefix, dashboardId, groupId)

	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsValidatorOfflineSubscribed, userId, string(types.ValidatorIsOfflineEventName), eventFilter, epoch, 0)
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsGroupOfflineSubscribed, userId, string(types.GroupIsOfflineEventName), eventFilter, epoch, settings.GroupOfflineThreshold)
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsAttestationsMissedSubscribed, userId, string(types.ValidatorMissedAttestationEventName), eventFilter, epoch, 0)
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsBlockProposalSubscribed, userId, string(types.ValidatorProposalEventName), eventFilter, epoch, 0)
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsUpcomingBlockProposalSubscribed, userId, string(types.ValidatorUpcomingProposalEventName), eventFilter, epoch, 0)
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsSyncSubscribed, userId, string(types.SyncCommitteeSoon), eventFilter, epoch, 0)
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsWithdrawalProcessedSubscribed, userId, string(types.ValidatorReceivedWithdrawalEventName), eventFilter, epoch, 0)
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsSlashedSubscribed, userId, string(types.ValidatorGotSlashedEventName), eventFilter, epoch, 0)
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsMaxCollateralSubscribed, userId, string(types.RocketpoolCollateralMaxReached), eventFilter, epoch, settings.MaxCollateralThreshold)
	d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsMinCollateralSubscribed, userId, string(types.RocketpoolCollateralMinReached), eventFilter, epoch, settings.MinCollateralThreshold)

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
	_, err = d.alloyWriter.ExecContext(ctx, `
		UPDATE users_val_dashboards_groups 
		SET 
			webhook_target = NULLIF($1, ''),
			webhook_format = CASE WHEN $2 THEN $3 ELSE NULL END,
			realtime_notifications = CASE WHEN $4 THEN TRUE ELSE NULL END
		WHERE dashboard_id = $5 AND id = $6`, settings.WebhookUrl, settings.IsWebhookDiscordEnabled, DiscordWebhookFormat, settings.IsRealTimeModeEnabled, dashboardId, groupId)
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

	// d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsIncomingTransactionsSubscribed, userId, string(types.IncomingTransactionEventName), eventFilter, epoch, 0)
	// d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsOutgoingTransactionsSubscribed, userId, string(types.OutgoingTransactionEventName), eventFilter, epoch, 0)
	// d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsERC20TokenTransfersSubscribed, userId, string(types.ERC20TokenTransferEventName), eventFilter, epoch, settings.ERC20TokenTransfersValueThreshold)
	// d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsERC721TokenTransfersSubscribed, userId, string(types.ERC721TokenTransferEventName), eventFilter, epoch, 0)
	// d.AddOrRemoveEvent(&eventsToInsert, &eventsToDelete, settings.IsERC1155TokenTransfersSubscribed, userId, string(types.ERC1155TokenTransferEventName), eventFilter, epoch, 0)

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
	// _, err = d.alloyWriter.ExecContext(ctx, `
	// 	UPDATE users_acc_dashboards_groups
	// 	SET
	// 		webhook_target = NULLIF($1, ''),
	// 		webhook_format = CASE WHEN $2 THEN $3 ELSE NULL END,
	// 		ignore_spam_transactions = $4,
	// 		subscribed_chain_ids = $5
	// 	WHERE dashboard_id = $6 AND id = $7`, settings.WebhookUrl, settings.IsWebhookDiscordEnabled, DiscordWebhookFormat, settings.IsIgnoreSpamTransactionsEnabled, settings.SubscribedChainIds, dashboardId, groupId)
	// if err != nil {
	// 	return err
	// }

	return d.dummy.UpdateNotificationSettingsAccountDashboard(ctx, userId, dashboardId, groupId, settings)
}

func (d *DataAccessService) AddOrRemoveEvent(eventsToInsert *[]goqu.Record, eventsToDelete *[]goqu.Expression, isSubscribed bool, userId uint64, eventName string, eventFilter string, epoch int64, threshold float64) {
	if isSubscribed {
		event := goqu.Record{"user_id": userId, "event_name": eventName, "event_filter": eventFilter, "created_ts": goqu.L("NOW()"), "created_epoch": epoch, "event_threshold": threshold}
		*eventsToInsert = append(*eventsToInsert, event)
	} else {
		*eventsToDelete = append(*eventsToDelete, goqu.Ex{"user_id": userId, "event_name": eventName, "event_filter": eventFilter})
	}
}
