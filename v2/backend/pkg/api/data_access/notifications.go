package dataaccess

import (
	"context"

	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"golang.org/x/sync/errgroup"
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
	eg := errgroup.Group{}

	response := []t.NotificationDashboardsTableRow{}

	// validator dashboards
	eg.Go(func() error {
		type subscritionResult struct {
			network     uint64 `db:"network"`
			dashboardId uint64 `db:"dashboard_id"`
			groupId     uint64 `db:"group_id"`
			groupName   string `db:"group_name"`
			entityCount uint64 `db:"entity_count"`
			eventName   string `db:"event_name"`
		}
		subscriptions := []struct {
			subscriptionId uint64 `db:"subscription_id"`
			subscritionResult
		}{}
		// NOTE: subscriptions and notifications are in different databases so we can't join/group efficiently; this makes things ugly...
		// 1. get subscribed notifications for user
		// TODO could change this: minimize data transfer by splitting the query into dashboards + subscriptions retrieval; need to test performance
		query := `SELECT
			uvd.network,
			uvd.id as dashboard_id,
			uvdg.id as group_id,
			uvdg.name as group_name,
			COUNT(uvdv.validator_index) as entity_count,
			event_name
		FROM
			(
			SELECT
				split_part(event_filter, ':', 2)::int as dashboard_id,
				split_part(event_filter, ':', 3)::int as group_id,
				event_name
			FROM
				users_subscriptions
			WHERE
				user_id = $1 AND event_filter like 'vdb:%'
			) as us
		INNER JOIN
			users_val_dashboards uvd ON us.dashboard_id = uvd.id AND uvd.network = $2
		INNER JOIN
			users_val_dashboards_groups uvdg ON us.dashboard_id = uvdg.dashboard_id AND us.group_id = uvdg.id
		LEFT JOIN
			users_val_dashboards_validators uvdv ON us.dashboard_id = uvdv.dashboard_id AND us.group_id = uvdv.group_id
		GROUP BY
			us.id, uvd.network, uvd.id, uvdg.id, uvdg.name, event_name
		`
		err := d.alloyReader.GetContext(ctx, &subscriptions, query, userId, chainId)
		if err != nil {
			return err
		}
		subscriptionsMap := make(map[uint64]subscritionResult)
		// dashboardId -> groupId -> epoch -> notificationIds
		r := map[uint64]map[uint64]map[uint64][]uint64{}
		for _, subscription := range subscriptions {
			subscriptionsMap[subscription.subscriptionId] = subscription.subscritionResult
			r[subscription.dashboardId][subscription.groupId] = make(map[uint64][]uint64)
		}

		// 2. get requested notifications (based on cursor, search etc.)
		// providing grouped data so
		subscriptionIds := make([]uint64, len(subscriptions))
		for _, subscription := range subscriptions {
			subscriptionIds = append(subscriptionIds, subscription.subscriptionId)
		}
		result := []struct {
			Epoch           uint64   `db:"epoch"`
			SubscriptionIds []uint64 `db:"subscription_ids"`
		}{}
		query = `SELECT
			ARRAY_AGG(DISTINCT subscription_id) AS ids,
			epoch
		FROM
			notification_queue
		WHERE
			subscription_id = ANY($1)
		GROUP BY
			epoch
		`
		err = d.alloyReader.GetContext(ctx, &result, query, subscriptionIds)
		if err != nil {
			return err
		}

		// 3. compose response rows

		for _, row := range result {
			tableRow := t.NotificationDashboardsTableRow{
				IsAccountDashboard: false,
				Epoch:              row.Epoch,
			}
			for _, subscriptionId := range row.SubscriptionIds {
				subscription := subscriptionsMap[subscriptionId]
				if _, ok := r[subscription.dashboardId][subscription.groupId][row.Epoch]; !ok {
					r[subscription.dashboardId][subscription.groupId][row.Epoch] = []uint64{}
				}
				r[subscription.dashboardId][subscription.groupId][row.Epoch] = append(r[subscription.dashboardId][subscription.groupId][row.Epoch], subscriptionId)
				tableRow.ChainId = subscription.network
				tableRow.DashboardId = subscription.dashboardId
				tableRow.GroupId = subscription.groupId
				tableRow.GroupName = subscription.groupName
				tableRow.EntityCount = subscription.entityCount
				tableRow.EventTypes = append(tableRow.EventTypes, subscription.eventName)
			}
			response = append(response, tableRow)
		}
		return nil
	})

	// account dashboards
	eg.Go(func() error {
		query := `SELECT 
			last_sent_ts,
			uvd.network,					// chainID
			us.id,							// notification id
			us.created_ts,					// timestamp
			uvd.id,		 					// dashboardId
			uvd.name,						// group name
			COUNT(uvdv.validator_index),	// entity count
			event_name 						// event types
		FROM
			(
			SELECT
				id,
				created_ts,
				split_part(event_filter, ':', 2) as dashboard_id
			FROM
				users_subscriptions
			WHERE
				event_filter like 'adb:%'
			) as us
		LEFT JOIN
			users_val_dashboards uvd ON
		LEFT JOIN
			users_val_dashboards_validators uvdv ON
		GROUP BY
			id
		`
		return d.alloyReader.GetContext(ctx, nil, query)
	})

	return response, nil, eg.Wait()
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
	return d.dummy.UpdateNotificationSettingsGeneral(ctx, userId, settings)
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
