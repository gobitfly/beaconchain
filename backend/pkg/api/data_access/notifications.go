package dataaccess

import (
	"context"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
)

type NotificationsRepository interface {
	GetNotificationOverview(ctx context.Context, userId uint64) (*t.NotificationOverviewData, error)

	GetDashboardNotifications(ctx context.Context, userId uint64, chainId uint64, cursor string, colSort t.Sort[enums.NotificationDashboardsColumn], search string, limit uint64) ([]t.NotificationDashboardsTableRow, *t.Paging, error)
	// depending on how notifications are implemented, we may need to use something other than `notificationId` for identifying the notification
	GetValidatorDashboardNotificationDetails(ctx context.Context, notificationId string) (*t.NotificationValidatorDashboardDetail, error)
	GetAccountDashboardNotificationDetails(ctx context.Context, notificationId string) (*t.NotificationAccountDashboardDetail, error)

	GetMachineNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationMachinesColumn], search string, limit uint64) ([]t.NotificationMachinesTableRow, *t.Paging, error)
	GetClientNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationClientsColumn], search string, limit uint64) ([]t.NotificationClientsTableRow, *t.Paging, error)
	GetRocketPoolNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationRocketPoolColumn], search string, limit uint64) ([]t.NotificationRocketPoolTableRow, *t.Paging, error)
	GetNetworkNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationNetworksColumn], search string, limit uint64) ([]t.NotificationNetworksTableRow, *t.Paging, error)

	GetNotificationSettings(ctx context.Context, userId uint64) (*t.NotificationSettings, error)
	UpdateNotificationSettingsGeneral(ctx context.Context, userId uint64, settings t.NotificationSettingsGeneral) error
	UpdateNotificationSettingsNetworks(ctx context.Context, userId uint64, chainId uint64, settings t.NotificationSettingsNetwork) error
	UpdateNotificationSettingsPairedDevice(ctx context.Context, pairedDeviceId string, name string, IsNotificationsEnabled bool) error
	DeleteNotificationSettingsPairedDevice(ctx context.Context, pairedDeviceId string) error
	GetNotificationSettingsDashboards(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationSettingsDashboardColumn], search string, limit uint64) ([]t.NotificationSettingsDashboardsTableRow, *t.Paging, error)
	UpdateNotificationSettingsValidatorDashboard(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, settings t.NotificationSettingsValidatorDashboard) error
	UpdateNotificationSettingsAccountDashboard(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, settings t.NotificationSettingsAccountDashboard) error
}

func (d *DataAccessService) GetNotificationOverview(ctx context.Context, userId uint64) (*t.NotificationOverviewData, error) {
	return d.dummy.GetNotificationOverview(ctx, userId)
}
func (d *DataAccessService) GetDashboardNotifications(ctx context.Context, userId uint64, chainId uint64, cursor string, colSort t.Sort[enums.NotificationDashboardsColumn], search string, limit uint64) ([]t.NotificationDashboardsTableRow, *t.Paging, error) {
	return d.dummy.GetDashboardNotifications(ctx, userId, chainId, cursor, colSort, search, limit)
}

func (d *DataAccessService) GetValidatorDashboardNotificationDetails(ctx context.Context, notificationId string) (*t.NotificationValidatorDashboardDetail, error) {
	return d.dummy.GetValidatorDashboardNotificationDetails(ctx, notificationId)
}

func (d *DataAccessService) GetAccountDashboardNotificationDetails(ctx context.Context, notificationId string) (*t.NotificationAccountDashboardDetail, error) {
	return d.dummy.GetAccountDashboardNotificationDetails(ctx, notificationId)
}

func (d *DataAccessService) GetMachineNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationMachinesColumn], search string, limit uint64) ([]t.NotificationMachinesTableRow, *t.Paging, error) {
	return d.dummy.GetMachineNotifications(ctx, userId, cursor, colSort, search, limit)
}
func (d *DataAccessService) GetClientNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationClientsColumn], search string, limit uint64) ([]t.NotificationClientsTableRow, *t.Paging, error) {
	return d.dummy.GetClientNotifications(ctx, userId, cursor, colSort, search, limit)
}
func (d *DataAccessService) GetRocketPoolNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationRocketPoolColumn], search string, limit uint64) ([]t.NotificationRocketPoolTableRow, *t.Paging, error) {
	return d.dummy.GetRocketPoolNotifications(ctx, userId, cursor, colSort, search, limit)
}
func (d *DataAccessService) GetNetworkNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationNetworksColumn], search string, limit uint64) ([]t.NotificationNetworksTableRow, *t.Paging, error) {
	return d.dummy.GetNetworkNotifications(ctx, userId, cursor, colSort, search, limit)
}
func (d *DataAccessService) GetNotificationSettings(ctx context.Context, userId uint64) (*t.NotificationSettings, error) {
	return d.dummy.GetNotificationSettings(ctx, userId)
}
func (d *DataAccessService) UpdateNotificationSettingsGeneral(ctx context.Context, userId uint64, settings t.NotificationSettingsGeneral) error {
	// TODO: Are there plans to have the network as a prefix like before for rocketpool?

	latestEpoch := cache.LatestEpoch.Get()

	var eventsToInsert []goqu.Record
	var eventsToDelete []goqu.Expression

	tx, err := d.alloyWriter.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting db transactions to update general notification settings: %w", err)
	}
	defer utils.Rollback(tx)

	// -------------------------------------
	// Set the "do not disturb" setting
	_, err = tx.ExecContext(ctx, `UPDATE users SET notifications_do_not_disturb_ts = TO_TIMESTAMP($1) where legacy_receipt is null WHERE id = $2`,
		settings.DoNotDisturbTimestamp, userId)
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
	// Subscribed clients events
	_, err = tx.ExecContext(ctx, `DELETE FROM users_subscriptions WHERE user_id = $1 AND event_name = $2 AND NOT (event_filter = ANY($3))`,
		userId, types.EthClientUpdateEventName, settings.SubscribedClients)
	if err != nil {
		return err
	}

	for _, client := range settings.SubscribedClients {
		eventsToInsert = append(eventsToInsert, goqu.Record{"user_id": userId, "event_name": types.EthClientUpdateEventName, "event_filter": client, "created_ts": goqu.L("NOW()"), "created_epoch": latestEpoch, "event_threshold": 0})
	}

	// -------------------------------------
	// Collect the machine and rocketpool events to set and delete

	//Machine events
	if settings.IsMachineOfflineSubscribed {
		event := goqu.Record{"user_id": userId, "event_name": types.MonitoringMachineOfflineEventName, "event_filter": "", "created_ts": goqu.L("NOW()"), "created_epoch": latestEpoch, "event_threshold": 0}
		eventsToInsert = append(eventsToInsert, event)
	} else {
		event := goqu.Ex{"user_id": userId, "event_name": types.MonitoringMachineOfflineEventName, "event_filter": ""}
		eventsToDelete = append(eventsToDelete, event)
	}
	if settings.MachineStorageUsageThreshold > 0 {
		event := goqu.Record{"user_id": userId, "event_name": types.MonitoringMachineDiskAlmostFullEventName, "event_filter": "", "created_ts": goqu.L("NOW()"), "created_epoch": latestEpoch, "event_threshold": settings.MachineStorageUsageThreshold}
		eventsToInsert = append(eventsToInsert, event)
	} else {
		event := goqu.Ex{"user_id": userId, "event_name": types.MonitoringMachineDiskAlmostFullEventName, "event_filter": ""}
		eventsToDelete = append(eventsToDelete, event)
	}
	if settings.MachineCpuUsageThreshold > 0 {
		event := goqu.Record{"user_id": userId, "event_name": types.MonitoringMachineCpuLoadEventName, "event_filter": "", "created_ts": goqu.L("NOW()"), "created_epoch": latestEpoch, "event_threshold": settings.MachineCpuUsageThreshold}
		eventsToInsert = append(eventsToInsert, event)
	} else {
		event := goqu.Ex{"user_id": userId, "event_name": types.MonitoringMachineCpuLoadEventName, "event_filter": ""}
		eventsToDelete = append(eventsToDelete, event)
	}
	if settings.MachineMemoryUsageThreshold > 0 {
		event := goqu.Record{"user_id": userId, "event_name": types.MonitoringMachineMemoryUsageEventName, "event_filter": "", "created_ts": goqu.L("NOW()"), "created_epoch": latestEpoch, "event_threshold": settings.MachineMemoryUsageThreshold}
		eventsToInsert = append(eventsToInsert, event)
	} else {
		event := goqu.Ex{"user_id": userId, "event_name": types.MonitoringMachineMemoryUsageEventName, "event_filter": ""}
		eventsToDelete = append(eventsToDelete, event)
	}

	// RocketPool events
	if settings.IsRocketPoolNewRewardRoundSubscribed {
		event := goqu.Record{"user_id": userId, "event_name": types.RocketpoolNewClaimRoundStartedEventName, "event_filter": "", "created_ts": goqu.L("NOW()"), "created_epoch": latestEpoch, "event_threshold": 0}
		eventsToInsert = append(eventsToInsert, event)
	} else {
		event := goqu.Ex{"user_id": userId, "event_name": types.RocketpoolNewClaimRoundStartedEventName, "event_filter": ""}
		eventsToDelete = append(eventsToDelete, event)
	}
	if settings.RocketPoolMaxCollateralThreshold > 0 {
		event := goqu.Record{"user_id": userId, "event_name": types.RocketpoolCollateralMaxReached, "event_filter": "", "created_ts": goqu.L("NOW()"), "created_epoch": latestEpoch, "event_threshold": settings.RocketPoolMaxCollateralThreshold}
		eventsToInsert = append(eventsToInsert, event)
	} else {
		event := goqu.Ex{"user_id": userId, "event_name": types.RocketpoolCollateralMaxReached, "event_filter": ""}
		eventsToDelete = append(eventsToDelete, event)
	}
	if settings.RocketPoolMinCollateralThreshold > 0 {
		event := goqu.Record{"user_id": userId, "event_name": types.RocketpoolCollateralMinReached, "event_filter": "", "created_ts": goqu.L("NOW()"), "created_epoch": latestEpoch, "event_threshold": settings.RocketPoolMinCollateralThreshold}
		eventsToInsert = append(eventsToInsert, event)
	} else {
		event := goqu.Ex{"user_id": userId, "event_name": types.RocketpoolCollateralMinReached, "event_filter": ""}
		eventsToDelete = append(eventsToDelete, event)
	}

	// Insert all the events or update the threshold if they already exist
	insertDs := goqu.Dialect("postgres").
		Insert("users_subscriptions").
		Cols("user_id", "event_name", "event_filter", "created_ts", "created_epoch", "event_threshold").
		Rows(eventsToInsert).
		OnConflict(goqu.DoUpdate("user_id, event_name, event_filter", goqu.L("event_threshold = EXCLUDED.event_threshold")))

	query, args, err := insertDs.Prepared(true).ToSQL()
	if err != nil {
		return fmt.Errorf("error preparing query: %v", err)
	}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	// Delete all the events
	deleteDs := goqu.Dialect("postgres").
		Delete("users_subscriptions").
		Where(goqu.Or(eventsToDelete...))

	query, args, err = deleteDs.Prepared(true).ToSQL()
	if err != nil {
		return fmt.Errorf("error preparing query: %v", err)
	}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing tx to update general notification settings: %w", err)
	}
	return nil
}
func (d *DataAccessService) UpdateNotificationSettingsNetworks(ctx context.Context, userId uint64, chainId uint64, settings t.NotificationSettingsNetwork) error {
	return d.dummy.UpdateNotificationSettingsNetworks(ctx, userId, chainId, settings)
}
func (d *DataAccessService) UpdateNotificationSettingsPairedDevice(ctx context.Context, pairedDeviceId string, name string, IsNotificationsEnabled bool) error {
	return d.dummy.UpdateNotificationSettingsPairedDevice(ctx, pairedDeviceId, name, IsNotificationsEnabled)
}
func (d *DataAccessService) DeleteNotificationSettingsPairedDevice(ctx context.Context, pairedDeviceId string) error {
	return d.dummy.DeleteNotificationSettingsPairedDevice(ctx, pairedDeviceId)
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
