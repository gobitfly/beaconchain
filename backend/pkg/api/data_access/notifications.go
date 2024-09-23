package dataaccess

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
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
	UpdateNotificationSettingsValidatorDashboard(ctx context.Context, userId uint64, dashboardId t.VDBIdPrimary, groupId uint64, settings t.NotificationSettingsValidatorDashboard) error
	UpdateNotificationSettingsAccountDashboard(ctx context.Context, userId uint64, dashboardId t.VDBIdPrimary, groupId uint64, settings t.NotificationSettingsAccountDashboard) error
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
func (d *DataAccessService) GetNotificationSettingsDashboards(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationSettingsDashboardColumn], search string, limit uint64) ([]t.NotificationSettingsDashboardsTableRow, *t.Paging, error) {
	// Get settings with userId from users_val_dashboards_groups, users_acc_dashboards_groups and users_subscriptions
	result := make([]t.NotificationSettingsDashboardsTableRow, 0)
	var paging t.Paging

	wg := errgroup.Group{}

	// -------------------------------------
	// Get the validator dashboards
	valDashboards := []struct {
		WebhookUrl            sql.NullString `db:"webhook_target"`
		Destination           string         `db:"destination"`
		IsRealTimeModeEnabled bool           `db:"real_time_mode"`
	}{}
	wg.Go(func() error {
		err := d.alloyReader.SelectContext(ctx, &valDashboards, `
			SELECT
				webhook_target,
				destination,
				real_time_mode
			FROM users_val_dashboards_groups
			WHERE user_id = $1 AND (webhook_target IS NOT NULL OR real_time_mode)`, userId)
		if err != nil {
			return fmt.Errorf(`error retrieving data for validator dashboard notifications: %w`, err)
		}

		return nil
	})

	// -------------------------------------
	// Get the account dashboards
	accDashboards := []struct {
		WebhookUrl                      sql.NullString `db:"webhook_target"`
		Destination                     string         `db:"destination"`
		IsIgnoreSpamTransactionsEnabled bool           `db:"ignore_spam_transactions"`
		SubscribedChainIds              []uint64       `db:"subscribed_chain_ids"`
	}{}
	wg.Go(func() error {
		err := d.alloyReader.SelectContext(ctx, &accDashboards, `
			SELECT
				webhook_target,
				destination,
				ignore_spam_transactions,
				subscribed_chain_ids
			FROM users_acc_dashboards_groups
			WHERE user_id = $1 AND (webhook_target IS NOT NULL OR IsIgnoreSpamTransactionsEnabled OR array_length(subscribed_chain_ids, 1) > 0)`, userId)
		if err != nil {
			return fmt.Errorf(`error retrieving data for validator dashboard notifications: %w`, err)
		}

		return nil
	})

	// -------------------------------------
	// Get the events
	events := []struct {
		EventName      string  `db:"event_name"`
		EventFilter    string  `db:"event_filter"`
		EventThreshold float64 `db:"event_threshold"`
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

	err := wg.Wait()
	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving dashboard notification data: %w", err)
	}

	return result, &paging, nil
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

	eventFilter := fmt.Sprintf("vdb:%d:%d", dashboardId, groupId)

	d.AddOrRemoveEvent(eventsToInsert, eventsToDelete, settings.IsValidatorOfflineSubscribed, userId, types.ValidatorIsOfflineEventName, eventFilter, epoch, 0)
	d.AddOrRemoveEvent(eventsToInsert, eventsToDelete, settings.IsGroupOfflineSubscribed, userId, types.GroupIsOfflineEventName, eventFilter, epoch, settings.GroupOfflineThreshold)
	d.AddOrRemoveEvent(eventsToInsert, eventsToDelete, settings.IsAttestationsMissedSubscribed, userId, types.ValidatorMissedAttestationEventName, eventFilter, epoch, 0)
	d.AddOrRemoveEvent(eventsToInsert, eventsToDelete, settings.IsBlockProposalSubscribed, userId, types.ValidatorMissedProposalEventName, eventFilter, epoch, 0)
	d.AddOrRemoveEvent(eventsToInsert, eventsToDelete, settings.IsUpcomingBlockProposalSubscribed, userId, types.ValidatorUpcomingProposalEventName, eventFilter, epoch, 0)
	d.AddOrRemoveEvent(eventsToInsert, eventsToDelete, settings.IsSyncSubscribed, userId, types.SyncCommitteeSoon, eventFilter, epoch, 0)
	d.AddOrRemoveEvent(eventsToInsert, eventsToDelete, settings.IsWithdrawalProcessedSubscribed, userId, types.ValidatorReceivedWithdrawalEventName, eventFilter, epoch, 0)
	d.AddOrRemoveEvent(eventsToInsert, eventsToDelete, settings.IsSlashedSubscribed, userId, types.ValidatorGotSlashedEventName, eventFilter, epoch, 0)

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
	destination := types.WebhookNotificationChannel
	if settings.IsWebhookDiscordEnabled {
		destination = types.WebhookDiscordNotificationChannel
	}
	_, err = d.alloyWriter.ExecContext(ctx, `
		UPDATE users_val_dashboards_groups 
		SET 
			webhook_target = NULLIF($1, ''),
			destination = $2,
			real_time_mode = $3
		WHERE dashboard_id = $4 AND id = $5`, settings.WebhookUrl, destination, settings.IsRealTimeModeEnabled, dashboardId, groupId)
	if err != nil {
		return err
	}

	return nil
}
func (d *DataAccessService) UpdateNotificationSettingsAccountDashboard(ctx context.Context, userId uint64, dashboardId t.VDBIdPrimary, groupId uint64, settings t.NotificationSettingsAccountDashboard) error {
	// For the given dashboardId and groupId update users_subscriptions and users_acc_dashboards_groups with the given settings
	epoch := utils.TimeToEpoch(time.Now())

	var eventsToInsert []goqu.Record
	var eventsToDelete []goqu.Expression

	tx, err := d.userWriter.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting db transactions to update validator dashboard notification settings: %w", err)
	}
	defer utils.Rollback(tx)

	eventFilter := fmt.Sprintf("adb:%d:%d", dashboardId, groupId)

	d.AddOrRemoveEvent(eventsToInsert, eventsToDelete, settings.IsIncomingTransactionsSubscribed, userId, types.IncomingTransactionEventName, eventFilter, epoch, 0)
	d.AddOrRemoveEvent(eventsToInsert, eventsToDelete, settings.IsOutgoingTransactionsSubscribed, userId, types.OutgoingTransactionEventName, eventFilter, epoch, 0)
	d.AddOrRemoveEvent(eventsToInsert, eventsToDelete, settings.IsERC20TokenTransfersSubscribed, userId, types.ERC20TokenTransferEventName, eventFilter, epoch, settings.ERC20TokenTransfersValueThreshold)
	d.AddOrRemoveEvent(eventsToInsert, eventsToDelete, settings.IsERC721TokenTransfersSubscribed, userId, types.ERC721TokenTransferEventName, eventFilter, epoch, 0)
	d.AddOrRemoveEvent(eventsToInsert, eventsToDelete, settings.IsERC1155TokenTransfersSubscribed, userId, types.ERC1155TokenTransferEventName, eventFilter, epoch, 0)

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
	destination := types.WebhookNotificationChannel
	if settings.IsWebhookDiscordEnabled {
		destination = types.WebhookDiscordNotificationChannel
	}
	_, err = d.alloyWriter.ExecContext(ctx, `
		UPDATE users_acc_dashboards_groups 
		SET 
			webhook_target = NULLIF($1, ''),
			destination = $2,
			ignore_spam_transactions = $3,
			subscribed_chain_ids = $4
		WHERE dashboard_id = $5 AND id = $6`, settings.WebhookUrl, destination, settings.IsIgnoreSpamTransactionsEnabled, settings.SubscribedChainIds, dashboardId, groupId)
	if err != nil {
		return err
	}

	return nil
}

func (d *DataAccessService) AddOrRemoveEvent(eventsToInsert []goqu.Record, eventsToDelete []goqu.Expression, isSubscribed bool, userId uint64, eventName types.EventName, eventFilter string, epoch int64, threshold float64) {
	if isSubscribed {
		event := goqu.Record{"user_id": userId, "event_name": eventName, "event_filter": eventFilter, "created_ts": goqu.L("NOW()"), "created_epoch": epoch, "event_threshold": threshold}
		eventsToInsert = append(eventsToInsert, event)
	} else {
		eventsToDelete = append(eventsToDelete, goqu.Ex{"user_id": userId, "event_name": eventName, "event_filter": eventFilter})
	}
}
