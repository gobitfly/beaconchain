package dataaccess

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/params"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
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
	MachineStorageUsageThresholdDefault     float64 = 0.1
	MachineCpuUsageThresholdDefault         float64 = 0.2
	MachineMemoryUsageThresholdDefault      float64 = 0.3
	RocketPoolMaxCollateralThresholdDefault float64 = 0.4
	RocketPoolMinCollateralThresholdDefault float64 = 0.5

	GasAboveThresholdDefault          float64 = 1000.0001
	GasBelowThresholdDefault          float64 = 1000.0002
	ParticipationRateThresholdDefault float64 = 0.6
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
