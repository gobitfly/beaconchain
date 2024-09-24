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

const (
	ValidatorDashboardEventPrefix string = "vdb"
	AccountDashboardEventPrefix   string = "adb"
)

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
	result := make([]t.NotificationSettingsDashboardsTableRow, 0)
	var paging t.Paging

	wg := errgroup.Group{}

	// Initialize the cursor
	var currentCursor t.NotificationSettingsCursor
	var err error
	if cursor != "" {
		currentCursor, err = utils.StringToCursor[t.NotificationSettingsCursor](cursor)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse passed cursor as WithdrawalsCursor: %w", err)
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
		IsRealTimeModeEnabled   bool           `db:"real_time_mode"`
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
				g.discord_webhook,
				g.real_time_mode
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
		IsWebhookDiscordEnabled         sql.NullBool   `db:"discord_webhook"`
		IsIgnoreSpamTransactionsEnabled bool           `db:"ignore_spam_transactions"`
		SubscribedChainIds              []uint64       `db:"subscribed_chain_ids"`
	}{}
	wg.Go(func() error {
		err := d.alloyReader.SelectContext(ctx, &accDashboards, `
			SELECT
				d.id AS dashboard_id,
				d.name AS dashboard_name,
				g.id AS group_id,
				g.name AS group_name,
				g.webhook_target,
				g.discord_webhook,
				g.ignore_spam_transactions,
				g.subscribed_chain_ids
			FROM users_val_dashboards d
			INNER JOIN users_val_dashboards_groups g ON d.id = g.dashboard_id
			WHERE d.user_id = $1`, userId)
		if err != nil {
			return fmt.Errorf(`error retrieving data for validator dashboard notifications: %w`, err)
		}

		return nil
	})

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
					Settings: t.NotificationSettingsValidatorDashboard{},
				}
			} else if dashboardType == AccountDashboardEventPrefix {
				settingsMap[event.Filter] = &NotificationSettingsDashboardsInfo{
					Settings: t.NotificationSettingsAccountDashboard{},
				}
			}
		}

		if valSettings, ok := settingsMap[event.Filter].Settings.(*t.NotificationSettingsValidatorDashboard); ok {
			switch event.Name {
			case types.ValidatorIsOfflineEventName:
				valSettings.IsValidatorOfflineSubscribed = true
			case types.GroupIsOfflineEventName:
				valSettings.IsGroupOfflineSubscribed = true
				valSettings.GroupOfflineThreshold = event.Threshold
			case types.ValidatorMissedAttestationEventName:
				valSettings.IsAttestationsMissedSubscribed = true
			case types.ValidatorProposalEventName:
				valSettings.IsBlockProposalSubscribed = true
			case types.ValidatorUpcomingProposalEventName:
				valSettings.IsUpcomingBlockProposalSubscribed = true
			case types.SyncCommitteeSoon:
				valSettings.IsSyncSubscribed = true
			case types.ValidatorReceivedWithdrawalEventName:
				valSettings.IsWithdrawalProcessedSubscribed = true
			case types.ValidatorGotSlashedEventName:
				valSettings.IsSlashedSubscribed = true
			}
		} else if accSettings, ok := settingsMap[event.Filter].Settings.(*t.NotificationSettingsAccountDashboard); ok {
			switch event.Name {
			case types.IncomingTransactionEventName:
				accSettings.IsIncomingTransactionsSubscribed = true
			case types.OutgoingTransactionEventName:
				accSettings.IsOutgoingTransactionsSubscribed = true
			case types.ERC20TokenTransferEventName:
				accSettings.IsERC20TokenTransfersSubscribed = true
				accSettings.ERC20TokenTransfersValueThreshold = event.Threshold
			case types.ERC721TokenTransferEventName:
				accSettings.IsERC721TokenTransfersSubscribed = true
			case types.ERC1155TokenTransferEventName:
				accSettings.IsERC1155TokenTransfersSubscribed = true
			}
		}
	}

	// Validator dashboards
	for _, valDashboard := range valDashboards {
		key := fmt.Sprintf("%s:%d:%d", ValidatorDashboardEventPrefix, valDashboard.DashboardId, valDashboard.GroupId)

		if _, ok := settingsMap[key]; !ok {
			if !valDashboard.WebhookUrl.Valid && !valDashboard.IsRealTimeModeEnabled {
				// No subscriptions for this dashboard
				continue
			}

			settingsMap[key] = &NotificationSettingsDashboardsInfo{
				Settings: t.NotificationSettingsValidatorDashboard{},
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
			valSettings.IsRealTimeModeEnabled = valDashboard.IsRealTimeModeEnabled
		}
	}

	// Account dashboards
	for _, accDashboard := range accDashboards {
		key := fmt.Sprintf("%s:%d:%d", AccountDashboardEventPrefix, accDashboard.DashboardId, accDashboard.GroupId)

		if _, ok := settingsMap[key]; !ok {
			if !accDashboard.WebhookUrl.Valid && !accDashboard.IsIgnoreSpamTransactionsEnabled && len(accDashboard.SubscribedChainIds) == 0 {
				// No subscription for this dashboard
				continue
			}

			settingsMap[key] = &NotificationSettingsDashboardsInfo{
				Settings: t.NotificationSettingsAccountDashboard{},
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
		for key, setting := range settingsMap {
			if search != setting.DashboardName && search != setting.GroupName {
				delete(settingsMap, key)
			}
		}
	}

	// Convert to a slice for sorting and paging
	settings := slices.Collect(maps.Values(settingsMap))

	// -------------------------------------
	// Sort
	var primarySortParam func(resultEntry *NotificationSettingsDashboardsInfo) string
	var secondarySortParam func(resultEntry *NotificationSettingsDashboardsInfo) string
	switch colSort.Column {
	case enums.NotificationSettingsDashboardColumns.DashboardName:
		primarySortParam = func(resultEntry *NotificationSettingsDashboardsInfo) string {
			return resultEntry.DashboardName
		}
		secondarySortParam = func(resultEntry *NotificationSettingsDashboardsInfo) string {
			return resultEntry.GroupName
		}
	case enums.NotificationSettingsDashboardColumns.GroupName:
		primarySortParam = func(resultEntry *NotificationSettingsDashboardsInfo) string {
			return resultEntry.GroupName
		}
		secondarySortParam = func(resultEntry *NotificationSettingsDashboardsInfo) string {
			return resultEntry.DashboardName
		}
	default:
		return nil, nil, fmt.Errorf("invalid sort column for notification subscriptions: %v", colSort.Column)
	}
	sort.Slice(settings, func(i, j int) bool {
		if isReverseDirection {
			if primarySortParam(settings[i]) == primarySortParam(settings[j]) {
				if secondarySortParam(settings[i]) == secondarySortParam(settings[j]) {
					return settings[i].IsAccountDashboard
				}
				return secondarySortParam(settings[i]) > secondarySortParam(settings[j])
			}
			return primarySortParam(settings[i]) > primarySortParam(settings[j])
		} else {
			if primarySortParam(settings[i]) == primarySortParam(settings[j]) {
				if secondarySortParam(settings[i]) == secondarySortParam(settings[j]) {
					return settings[j].IsAccountDashboard
				}
				return secondarySortParam(settings[i]) < secondarySortParam(settings[j])
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

	d.AddOrRemoveEvent(eventsToInsert, eventsToDelete, settings.IsValidatorOfflineSubscribed, userId, types.ValidatorIsOfflineEventName, eventFilter, epoch, 0)
	d.AddOrRemoveEvent(eventsToInsert, eventsToDelete, settings.IsGroupOfflineSubscribed, userId, types.GroupIsOfflineEventName, eventFilter, epoch, settings.GroupOfflineThreshold)
	d.AddOrRemoveEvent(eventsToInsert, eventsToDelete, settings.IsAttestationsMissedSubscribed, userId, types.ValidatorMissedAttestationEventName, eventFilter, epoch, 0)
	d.AddOrRemoveEvent(eventsToInsert, eventsToDelete, settings.IsBlockProposalSubscribed, userId, types.ValidatorProposalEventName, eventFilter, epoch, 0)
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
	_, err = d.alloyWriter.ExecContext(ctx, `
		UPDATE users_val_dashboards_groups 
		SET 
			webhook_target = NULLIF($1, ''),
			discord_webhook = CASE WHEN $1 = '' THEN NULL ELSE $2 END,
			real_time_mode = $3
		WHERE dashboard_id = $4 AND id = $5`, settings.WebhookUrl, settings.IsWebhookDiscordEnabled, settings.IsRealTimeModeEnabled, dashboardId, groupId)
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

	eventFilter := fmt.Sprintf("%s:%d:%d", AccountDashboardEventPrefix, dashboardId, groupId)

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
	_, err = d.alloyWriter.ExecContext(ctx, `
		UPDATE users_acc_dashboards_groups 
		SET 
			webhook_target = NULLIF($1, ''),
			discord_webhook = CASE WHEN $1 = '' THEN NULL ELSE $2 END,
			ignore_spam_transactions = $3,
			subscribed_chain_ids = $4
		WHERE dashboard_id = $5 AND id = $6`, settings.WebhookUrl, settings.IsWebhookDiscordEnabled, settings.IsIgnoreSpamTransactionsEnabled, settings.SubscribedChainIds, dashboardId, groupId)
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
