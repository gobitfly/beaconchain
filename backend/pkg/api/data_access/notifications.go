package dataaccess

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/gob"
	"fmt"
	"io"
	"sync"

	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/notification"
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

func (*DataAccessService) initNotifications() {
	var once sync.Once
	once.Do(func() {
		gob.Register(&notification.ValidatorProposalNotification{})
		gob.Register(&notification.ValidatorAttestationNotification{})
		gob.Register(&notification.ValidatorIsOfflineNotification{})
		gob.Register(&notification.ValidatorGotSlashedNotification{})
		gob.Register(&notification.ValidatorWithdrawalNotification{})
		gob.Register(&notification.NetworkNotification{})
		gob.Register(&notification.RocketpoolNotification{})
		gob.Register(&notification.MonitorMachineNotification{})
		gob.Register(&notification.TaxReportNotification{})
		gob.Register(&notification.EthClientNotification{})
		gob.Register(&notification.SyncCommitteeSoonNotification{})
	})
}

func (d *DataAccessService) GetNotificationOverview(ctx context.Context, userId uint64) (*t.NotificationOverviewData, error) {
	return d.dummy.GetNotificationOverview(ctx, userId)
}
func (d *DataAccessService) GetDashboardNotifications(ctx context.Context, userId uint64, chainIds []uint64, cursor string, colSort t.Sort[enums.NotificationDashboardsColumn], search string, limit uint64) ([]t.NotificationDashboardsTableRow, *t.Paging, error) {
	return d.dummy.GetDashboardNotifications(ctx, userId, chainIds, cursor, colSort, search, limit)
}

func (d *DataAccessService) GetValidatorDashboardNotificationDetails(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, epoch uint64) (*t.NotificationValidatorDashboardDetail, error) {
	var notificationDetails t.NotificationValidatorDashboardDetail

	var result []byte
	query := `SELECT details FROM users_val_dashboards_notifications_history WHERE dashboard_id = $1 AND group_id = $2 AND epoch = $3`
	err := d.alloyReader.SelectContext(ctx, &result, query)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(result)
	gz, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	decompressedData, err := io.ReadAll(gz)
	if err != nil {
		return nil, err
	}

	decoder := gob.NewDecoder(bytes.NewReader(decompressedData))

	var notifications []any
	err = decoder.Decode(&notifications)
	if err != nil {
		return nil, err
	}

	type ProposalInfo struct {
		Proposed  []uint64
		Scheduled []uint64
		Missed    []uint64
	}

	proposalsInfo := make(map[t.VDBValidator]*ProposalInfo)
	addressMapping := make(map[string]*t.Address)
	for _, not := range notifications {
		n := not.(types.Notification)
		switch n.GetEventName() {
		case types.ValidatorMissedProposalEventName, types.ValidatorExecutedProposalEventName /*, types.ValidatorScheduledProposalEventName*/ :
			// aggregate proposals
			curNotification, ok := not.(notification.ValidatorProposalNotification)
			if !ok {
				return nil, fmt.Errorf("failed to cast notification to ValidatorProposalNotification")
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
			curNotification, ok := not.(notification.ValidatorAttestationNotification)
			if !ok {
				return nil, fmt.Errorf("failed to cast notification to ValidatorAttestationNotification")
			}
			if curNotification.Status == 0 {
				continue
			}
			notificationDetails.AttestationMissed = append(notificationDetails.AttestationMissed, t.IndexEpoch{curNotification.ValidatorIndex, curNotification.Epoch})
		case types.ValidatorGotSlashedEventName:
			curNotification, ok := not.(notification.ValidatorGotSlashedNotification)
			if !ok {
				return nil, fmt.Errorf("failed to cast notification to ValidatorGotSlashedNotification")
			}
			notificationDetails.Slashed = append(notificationDetails.Slashed, curNotification.ValidatorIndex)
		case types.ValidatorIsOfflineEventName:
			curNotification, ok := not.(notification.ValidatorIsOfflineNotification)
			if !ok {
				return nil, fmt.Errorf("failed to cast notification to ValidatorIsOfflineNotification")
			}
			if curNotification.IsOffline {
				notificationDetails.ValidatorOffline = append(notificationDetails.ValidatorOffline, curNotification.ValidatorIndex)
			} else {
				// TODO EpochCount is not correct, missing / cumbersome to retrieve from backend - using "back online since" instead atm
				notificationDetails.ValidatorBackOnline = append(notificationDetails.ValidatorBackOnline, t.NotificationEventValidatorBackOnline{Index: curNotification.ValidatorIndex, EpochCount: curNotification.Epoch})
			}
			// TODO not present in backend yet
			//notificationDetails.ValidatorOfflineReminder = ...
		case types.ValidatorGroupIsOfflineEventName:
			// TODO type / collection not present yet, skipping
			/*curNotification, ok := not.(notification.validatorGroupIsOfflineNotification)
			if !ok {
				return nil, fmt.Errorf("failed to cast notification to validatorGroupIsOfflineNotification")
			}
			if curNotification.Status == 0 {
				notificationDetails.GroupOffline = ...
				notificationDetails.GroupOfflineReminder = ...
			} else {
				notificationDetails.GroupBackOnline = ...
			}
			*/
			continue
		case types.ValidatorReceivedWithdrawalEventName:
			curNotification, ok := not.(notification.ValidatorWithdrawalNotification)
			if !ok {
				return nil, fmt.Errorf("failed to cast notification to ValidatorWithdrawalNotification")
			}
			// TODO might need to take care of automatic + exit withdrawal happening in the same epoch ?
			notificationDetails.Withdrawal = append(notificationDetails.Withdrawal, t.IndexBlocks{curNotification.ValidatorIndex, []uint64{curNotification.Slot}})
		case types.NetworkLivenessIncreasedEventName,
			types.EthClientUpdateEventName,
			types.MonitoringMachineOfflineEventName,
			types.MonitoringMachineDiskAlmostFullEventName,
			types.MonitoringMachineCpuLoadEventName,
			types.MonitoringMachineMemoryUsageEventName,
			types.TaxReportEventName:
			// not vdb notifications, skip
			continue
		case types.ValidatorDidSlashEventName:
		case types.RocketpoolCommissionThresholdEventName,
			types.RocketpoolNewClaimRoundStartedEventName:
			// these could maybe returned later (?)
			continue
		case types.RocketpoolCollateralMinReachedEventName, types.RocketpoolCollateralMaxReachedEventName:
			_, ok := not.(notification.RocketpoolNotification)
			if !ok {
				return nil, fmt.Errorf("failed to cast notification to RocketpoolNotification")
			}
			addr := t.Address{Hash: t.Hash(n.GetEventFilter()), IsContract: true}
			addressMapping[n.GetEventFilter()] = &addr
			if n.GetEventName() == types.RocketpoolCollateralMinReachedEventName {
				notificationDetails.MinimumCollateralReached = append(notificationDetails.MinimumCollateralReached, addr)
			} else {
				notificationDetails.MaximumCollateralReached = append(notificationDetails.MaximumCollateralReached, addr)
			}
		case types.SyncCommitteeSoonEventName:
			curNotification, ok := not.(notification.SyncCommitteeSoonNotification)
			if !ok {
				return nil, fmt.Errorf("failed to cast notification to SyncCommitteeSoonNotification")
			}
			notificationDetails.SyncCommittee = append(notificationDetails.SyncCommittee, curNotification.Validator)
		default:
			log.Debugf("Unhandled notification type: %s", n.GetEventName())
		}
	}

	// fill proposals
	for validatorIndex, proposalInfo := range proposalsInfo {
		if len(proposalInfo.Proposed) > 0 {
			notificationDetails.ProposalDone = append(notificationDetails.ProposalDone, t.IndexBlocks{validatorIndex, proposalInfo.Proposed})
		}
		if len(proposalInfo.Scheduled) > 0 {
			notificationDetails.UpcomingProposals = append(notificationDetails.UpcomingProposals, t.IndexBlocks{validatorIndex, proposalInfo.Scheduled})
		}
		if len(proposalInfo.Missed) > 0 {
			notificationDetails.ProposalMissed = append(notificationDetails.ProposalMissed, t.IndexBlocks{validatorIndex, proposalInfo.Missed})
		}
	}

	// fill addresses
	if err := d.GetNamesAndEnsForAddresses(ctx, addressMapping); err != nil {
		return nil, err
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

	return &notificationDetails, nil
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
