package dataaccess

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
)

type NotificationsRepository interface {
	GetNotificationOverview(ctx context.Context, userId uint64) (*t.NotificationOverviewData, error)

	GetDashboardNotifications(ctx context.Context, userId uint64, chainId uint64, cursor string, colSort t.Sort[enums.NotificationDashboardsColumn], search string, limit uint64) ([]t.NotificationDashboardsTableRow, *t.Paging, error)
	// depending on how notifications are implemented, we may need to use something other than `notificationId` for identifying the notification
	GetValidatorDashboardNotificationDetails(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, epoch uint64) (*t.NotificationValidatorDashboardDetail, error)
	GetAccountDashboardNotificationDetails(ctx context.Context, dashboardId uint64, groupId uint64, epoch uint64) (*t.NotificationAccountDashboardDetail, error)

	GetMachineNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationMachinesColumn], search string, limit uint64) ([]t.NotificationMachinesTableRow, *t.Paging, error)
	GetClientNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationClientsColumn], search string, limit uint64) ([]t.NotificationClientsTableRow, *t.Paging, error)
	GetRocketPoolNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationRocketPoolColumn], search string, limit uint64) ([]t.NotificationRocketPoolTableRow, *t.Paging, error)
	GetNetworkNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationNetworksColumn], search string, limit uint64) ([]t.NotificationNetworksTableRow, *t.Paging, error)

	GetNotificationSettings(ctx context.Context, userId uint64) (*t.NotificationSettings, error)
	UpdateNotificationSettingsGeneral(ctx context.Context, userId uint64, settings t.NotificationSettingsGeneral) error
	UpdateNotificationSettingsNetworks(ctx context.Context, userId uint64, chainId uint64, settings t.NotificationSettingsNetwork) error
	UpdateNotificationSettingsPairedDevice(ctx context.Context, userId uint64, pairedDeviceId string, name string, IsNotificationsEnabled bool) error
	DeleteNotificationSettingsPairedDevice(ctx context.Context, userId uint64, pairedDeviceId string) error
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
func (d *DataAccessService) GetNetworkNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationNetworksColumn], search string, limit uint64) ([]t.NotificationNetworksTableRow, *t.Paging, error) {
	return d.dummy.GetNetworkNotifications(ctx, userId, cursor, colSort, search, limit)
}
func (d *DataAccessService) GetNotificationSettings(ctx context.Context, userId uint64) (*t.NotificationSettings, error) {
	result := &t.NotificationSettings{}

	wg := errgroup.Group{}

	networks, err := d.GetAllNetworks()
	if err != nil {
		return nil, err
	}

	chainIds := make(map[string]uint64, len(networks))
	for _, network := range networks {
		chainIds[network.Name] = network.ChainId
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

	err = wg.Wait()
	if err != nil {
		return nil, err
	}

	// -------------------------------------
	// Fill the result
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
	networkEvents := make(map[string]*t.NotificationSettingsNetwork)
	for _, event := range subscribedEvents {
		eventSplit := strings.Split(string(event.Name), ":")

		if len(eventSplit) == 2 {
			networkName := eventSplit[0]
			networkEvent := types.EventName(eventSplit[1])

			if _, ok := networkEvents[networkName]; !ok {
				networkEvents[networkName] = &t.NotificationSettingsNetwork{}
			}
			switch networkEvent {
			case types.RocketpoolNewClaimRoundStartedEventName:
				result.GeneralSettings.IsRocketPoolNewRewardRoundSubscribed = true
			case types.RocketpoolCollateralMaxReached:
				result.GeneralSettings.IsRocketPoolMaxCollateralSubscribed = true
				result.GeneralSettings.RocketPoolMaxCollateralThreshold = event.Threshold
			case types.RocketpoolCollateralMinReached:
				result.GeneralSettings.IsRocketPoolMinCollateralSubscribed = true
				result.GeneralSettings.RocketPoolMinCollateralThreshold = event.Threshold
			case types.NetworkGasAboveThresholdEventName:
				networkEvents[networkName].IsGasAboveSubscribed = true
				networkEvents[networkName].GasAboveThreshold = decimal.NewFromFloat(event.Threshold).Mul(decimal.NewFromInt(1e9))
			case types.NetworkGasBelowThresholdEventName:
				networkEvents[networkName].IsGasBelowSubscribed = true
				networkEvents[networkName].GasBelowThreshold = decimal.NewFromFloat(event.Threshold).Mul(decimal.NewFromInt(1e9))
			case types.NetworkParticipationRateThresholdEventName:
				networkEvents[networkName].IsParticipationRateSubscribed = true
				networkEvents[networkName].ParticipationRateThreshold = event.Threshold
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
				result.GeneralSettings.SubscribedClients = append(result.GeneralSettings.SubscribedClients, event.Filter)
			}
		}
	}

	for network, settings := range networkEvents {
		result.Networks = append(result.Networks, t.NotificationNetwork{
			ChainId:  chainIds[network],
			Settings: *settings,
		})

	}

	for _, device := range pairedDevices {
		result.PairedDevices = append(result.PairedDevices, t.NotificationPairedDevice{
			Id:                     device.DeviceIdentifier,
			PairedTimestamp:        device.CreatedTs.Unix(),
			Name:                   device.DeviceName,
			IsNotificationsEnabled: device.NotifyEnabled,
		})
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
func (d *DataAccessService) GetNotificationSettingsDashboards(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationSettingsDashboardColumn], search string, limit uint64) ([]t.NotificationSettingsDashboardsTableRow, *t.Paging, error) {
	return d.dummy.GetNotificationSettingsDashboards(ctx, userId, cursor, colSort, search, limit)
}
func (d *DataAccessService) UpdateNotificationSettingsValidatorDashboard(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, settings t.NotificationSettingsValidatorDashboard) error {
	return d.dummy.UpdateNotificationSettingsValidatorDashboard(ctx, dashboardId, groupId, settings)
}
func (d *DataAccessService) UpdateNotificationSettingsAccountDashboard(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, settings t.NotificationSettingsAccountDashboard) error {
	return d.dummy.UpdateNotificationSettingsAccountDashboard(ctx, dashboardId, groupId, settings)
}
