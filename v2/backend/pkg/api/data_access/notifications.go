package dataaccess

import (
	"context"
	"fmt"
	"slices"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/ethereum/go-ethereum/params"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/shopspring/decimal"
)

type NotificationsRepository interface {
	GetNotificationOverview(ctx context.Context, userId uint64) (*t.NotificationOverviewData, error)

	GetDashboardNotifications(ctx context.Context, userId uint64, chainIds []uint64, cursor string, colSort t.Sort[enums.NotificationDashboardsColumn], search string, limit uint64) ([]t.NotificationDashboardsTableRow, *t.Paging, error)
	// depending on how notifications are implemented, we may need to use something other than `notificationId` for identifying the notification
	GetValidatorDashboardNotificationDetails(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, epoch uint64) (*t.NotificationValidatorDashboardDetail, error)
	GetAccountDashboardNotificationDetails(ctx context.Context, dashboardId uint64, groupId uint64, epoch uint64) (*t.NotificationAccountDashboardDetail, error)

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
	UpdateNotificationSettingsValidatorDashboard(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, settings t.NotificationSettingsValidatorDashboard) error
	UpdateNotificationSettingsAccountDashboard(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, settings t.NotificationSettingsAccountDashboard) error
}

func (d *DataAccessService) GetNotificationOverview(ctx context.Context, userId uint64) (*t.NotificationOverviewData, error) {
	return d.dummy.GetNotificationOverview(ctx, userId)
}
func (d *DataAccessService) GetDashboardNotifications(ctx context.Context, userId uint64, chainIds []uint64, cursor string, colSort t.Sort[enums.NotificationDashboardsColumn], search string, limit uint64) ([]t.NotificationDashboardsTableRow, *t.Paging, error) {
	return d.dummy.GetDashboardNotifications(ctx, userId, chainIds, cursor, colSort, search, limit)
}

func (d *DataAccessService) GetValidatorDashboardNotificationDetails(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, epoch uint64) (*t.NotificationValidatorDashboardDetail, error) {
	return d.dummy.GetValidatorDashboardNotificationDetails(ctx, dashboardId, groupId, epoch)
}

func (d *DataAccessService) GetAccountDashboardNotificationDetails(ctx context.Context, dashboardId uint64, groupId uint64, epoch uint64) (*t.NotificationAccountDashboardDetail, error) {
	return d.dummy.GetAccountDashboardNotificationDetails(ctx, dashboardId, groupId, epoch)
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
			goqu.L("event_name"),
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
	return d.dummy.GetNotificationSettings(ctx, userId)
}
func (d *DataAccessService) UpdateNotificationSettingsGeneral(ctx context.Context, userId uint64, settings t.NotificationSettingsGeneral) error {
	return d.dummy.UpdateNotificationSettingsGeneral(ctx, userId, settings)
}
func (d *DataAccessService) UpdateNotificationSettingsNetworks(ctx context.Context, userId uint64, chainId uint64, settings t.NotificationSettingsNetwork) error {
	return d.dummy.UpdateNotificationSettingsNetworks(ctx, userId, chainId, settings)
}
func (d *DataAccessService) UpdateNotificationSettingsPairedDevice(ctx context.Context, userId uint64, pairedDeviceId string, name string, IsNotificationsEnabled bool) error {
	return d.dummy.UpdateNotificationSettingsPairedDevice(ctx, userId, pairedDeviceId, name, IsNotificationsEnabled)
}
func (d *DataAccessService) DeleteNotificationSettingsPairedDevice(ctx context.Context, userId uint64, pairedDeviceId string) error {
	return d.dummy.DeleteNotificationSettingsPairedDevice(ctx, userId, pairedDeviceId)
}
func (d *DataAccessService) UpdateNotificationSettingsClients(ctx context.Context, userId uint64, clientId uint64, IsSubscribed bool) (*t.NotificationSettingsClient, error) {
	return d.dummy.UpdateNotificationSettingsClients(ctx, userId, clientId, IsSubscribed)
}
func (d *DataAccessService) GetNotificationSettingsDashboards(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationSettingsDashboardColumn], search string, limit uint64) ([]t.NotificationSettingsDashboardsTableRow, *t.Paging, error) {
	return d.dummy.GetNotificationSettingsDashboards(ctx, userId, cursor, colSort, search, limit)
}
func (d *DataAccessService) UpdateNotificationSettingsValidatorDashboard(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, settings t.NotificationSettingsValidatorDashboard) error {
	return d.dummy.UpdateNotificationSettingsValidatorDashboard(ctx, dashboardId, groupId, settings)
}
func (d *DataAccessService) UpdateNotificationSettingsAccountDashboard(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, settings t.NotificationSettingsAccountDashboard) error {
	return d.dummy.UpdateNotificationSettingsAccountDashboard(ctx, dashboardId, groupId, settings)
}
