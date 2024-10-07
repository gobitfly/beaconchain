package dataaccess

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/ethereum/go-ethereum/params"
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

const (
	MachineStorageUsageThresholdDefault float64 = 0.9
	MachineCpuUsageThresholdDefault     float64 = 0.6
	MachineMemoryUsageThresholdDefault  float64 = 0.8

	GasAboveThresholdDefault          float64 = 950
	GasBelowThresholdDefault          float64 = 150
	ParticipationRateThresholdDefault float64 = 0.8
)

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
	return d.dummy.GetMachineNotifications(ctx, userId, cursor, colSort, search, limit)
}
func (d *DataAccessService) GetClientNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationClientsColumn], search string, limit uint64) ([]t.NotificationClientsTableRow, *t.Paging, error) {
	return d.dummy.GetClientNotifications(ctx, userId, cursor, colSort, search, limit)
}
func (d *DataAccessService) GetRocketPoolNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationRocketPoolColumn], search string, limit uint64) ([]t.NotificationRocketPoolTableRow, *t.Paging, error) {
	return d.dummy.GetRocketPoolNotifications(ctx, userId, cursor, colSort, search, limit)
}
func (d *DataAccessService) GetNetworkNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationNetworksColumn], limit uint64) ([]t.NotificationNetworksTableRow, *t.Paging, error) {
	return d.dummy.GetNetworkNotifications(ctx, userId, cursor, colSort, limit)
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
		DeviceIdentifier string    `db:"device_identifier"`
		CreatedTs        time.Time `db:"created_ts"`
		DeviceName       string    `db:"device_name"`
		NotifyEnabled    bool      `db:"notify_enabled"`
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
			Id:                     device.DeviceIdentifier,
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
	return d.dummy.GetNotificationSettingsDashboards(ctx, userId, cursor, colSort, search, limit)
}
func (d *DataAccessService) UpdateNotificationSettingsValidatorDashboard(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, settings t.NotificationSettingsValidatorDashboard) error {
	return d.dummy.UpdateNotificationSettingsValidatorDashboard(ctx, dashboardId, groupId, settings)
}
func (d *DataAccessService) UpdateNotificationSettingsAccountDashboard(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, settings t.NotificationSettingsAccountDashboard) error {
	return d.dummy.UpdateNotificationSettingsAccountDashboard(ctx, dashboardId, groupId, settings)
}

func (d *DataAccessService) AddOrRemoveEvent(eventsToInsert *[]goqu.Record, eventsToDelete *[]goqu.Expression, isSubscribed bool, userId uint64, eventName string, eventFilter string, epoch int64, threshold float64) {
	if isSubscribed {
		event := goqu.Record{"user_id": userId, "event_name": eventName, "event_filter": eventFilter, "created_ts": goqu.L("NOW()"), "created_epoch": epoch, "event_threshold": threshold}
		*eventsToInsert = append(*eventsToInsert, event)
	} else {
		*eventsToDelete = append(*eventsToDelete, goqu.Ex{"user_id": userId, "event_name": eventName, "event_filter": eventFilter})
	}
}
