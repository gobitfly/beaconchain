package dataaccess

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand/v2"
	"reflect"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/go-faker/faker/v4/pkg/options"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	"github.com/gobitfly/beaconchain/pkg/api/types"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/userservice"
	"github.com/shopspring/decimal"
)

type DummyService struct {
}

// ensure DummyService pointer implements DataAccessor
var _ DataAccessor = (*DummyService)(nil)

func NewDummyService() *DummyService {
	// define custom tags for faker
	_ = faker.AddProvider("eth", func(v reflect.Value) (interface{}, error) {
		return randomEthDecimal(), nil
	})
	_ = faker.AddProvider("cl_el_eth", func(v reflect.Value) (interface{}, error) {
		return t.ClElValue[decimal.Decimal]{
			Cl: randomEthDecimal(),
			El: randomEthDecimal(),
		}, nil
	})
	_ = faker.AddProvider("chain_ids", func(v reflect.Value) (interface{}, error) {
		possibleChainIds := []uint64{1, 100, 17000, 10200}
		rand.Shuffle(len(possibleChainIds), func(i, j int) {
			possibleChainIds[i], possibleChainIds[j] = possibleChainIds[j], possibleChainIds[i]
		})
		return possibleChainIds[:rand.IntN(len(possibleChainIds))], nil //nolint:gosec
	})
	return &DummyService{}
}

// generate random decimal.Decimal, should result in somewhere around 0.001 ETH (+/- a few decimal places) in Wei
func randomEthDecimal() decimal.Decimal {
	decimal, _ := decimal.NewFromString(fmt.Sprintf("%d00000000000", rand.Int64N(10000000))) //nolint:gosec
	return decimal
}

// must pass a pointer to the data
func commonFakeData(a interface{}) error {
	// TODO fake decimal.Decimal
	return faker.FakeData(a, options.WithRandomMapAndSliceMaxSize(5))
}

// used for any non-pointer data, e.g. all primitive types or slices
func getDummyData[T any]() (T, error) {
	var r T
	err := commonFakeData(&r)
	return r, err
}

// used for any struct data that should be returned as a pointer
func getDummyStruct[T any]() (*T, error) {
	var r T
	err := commonFakeData(&r)
	return &r, err
}

// used for any table data that should be returned with paging
func getDummyWithPaging[T any]() ([]T, *t.Paging, error) {
	r := []T{}
	p := t.Paging{}
	_ = commonFakeData(&r)
	err := commonFakeData(&p)
	return r, &p, err
}

func (d *DummyService) Close() {
	// nothing to close
}

func (d *DummyService) GetLatestSlot() (uint64, error) {
	return getDummyData[uint64]()
}

func (d *DummyService) GetLatestFinalizedEpoch() (uint64, error) {
	return getDummyData[uint64]()
}

func (d *DummyService) GetLatestBlock() (uint64, error) {
	return getDummyData[uint64]()
}

func (d *DummyService) GetBlockHeightAt(slot uint64) (uint64, error) {
	return getDummyData[uint64]()
}

func (d *DummyService) GetLatestExchangeRates() ([]t.EthConversionRate, error) {
	return getDummyData[[]t.EthConversionRate]()
}

func (d *DummyService) GetUserByEmail(ctx context.Context, email string) (uint64, error) {
	return getDummyData[uint64]()
}

func (d *DummyService) CreateUser(ctx context.Context, email, password string) (uint64, error) {
	return getDummyData[uint64]()
}

func (d *DummyService) RemoveUser(ctx context.Context, userId uint64) error {
	return nil
}

func (d *DummyService) UpdateUserEmail(ctx context.Context, userId uint64) error {
	return nil
}

func (d *DummyService) UpdateUserPassword(ctx context.Context, userId uint64, password string) error {
	return nil
}

func (d *DummyService) GetEmailConfirmationTime(ctx context.Context, userId uint64) (time.Time, error) {
	return getDummyData[time.Time]()
}

func (d *DummyService) GetPasswordResetTime(ctx context.Context, userId uint64) (time.Time, error) {
	return getDummyData[time.Time]()
}

func (d *DummyService) UpdateEmailConfirmationTime(ctx context.Context, userId uint64) error {
	return nil
}

func (d *DummyService) IsPasswordResetAllowed(ctx context.Context, userId uint64) (bool, error) {
	return true, nil
}

func (d *DummyService) UpdatePasswordResetTime(ctx context.Context, userId uint64) error {
	return nil
}

func (d *DummyService) UpdateEmailConfirmationHash(ctx context.Context, userId uint64, email, confirmationHash string) error {
	return nil
}

func (d *DummyService) UpdatePasswordResetHash(ctx context.Context, userId uint64, confirmationHash string) error {
	return nil
}

func (d *DummyService) GetUserInfo(ctx context.Context, userId uint64) (*t.UserInfo, error) {
	return getDummyStruct[t.UserInfo]()
}

func (d *DummyService) GetUserCredentialInfo(ctx context.Context, userId uint64) (*t.UserCredentialInfo, error) {
	return getDummyStruct[t.UserCredentialInfo]()
}

func (d *DummyService) GetUserIdByApiKey(ctx context.Context, apiKey string) (uint64, error) {
	return getDummyData[uint64]()
}

func (d *DummyService) GetUserIdByConfirmationHash(ctx context.Context, hash string) (uint64, error) {
	return getDummyData[uint64]()
}

func (d *DummyService) GetUserIdByResetHash(ctx context.Context, hash string) (uint64, error) {
	return getDummyData[uint64]()
}

func (d *DummyService) GetProductSummary(ctx context.Context) (*t.ProductSummary, error) {
	return getDummyStruct[t.ProductSummary]()
}

func (d *DummyService) GetFreeTierPerks(ctx context.Context) (*t.PremiumPerks, error) {
	return getDummyStruct[t.PremiumPerks]()
}

func (d *DummyService) GetValidatorDashboardUser(ctx context.Context, dashboardId t.VDBIdPrimary) (*t.DashboardUser, error) {
	return getDummyStruct[t.DashboardUser]()
}

func (d *DummyService) GetValidatorDashboardIdByPublicId(ctx context.Context, publicDashboardId t.VDBIdPublic) (*t.VDBIdPrimary, error) {
	return getDummyStruct[t.VDBIdPrimary]()
}

func (d *DummyService) GetValidatorDashboardInfo(ctx context.Context, dashboardId t.VDBIdPrimary) (*t.ValidatorDashboard, error) {
	r, err := getDummyStruct[t.ValidatorDashboard]()
	// return semi-valid data to not break staging
	r.IsArchived = false
	return r, err
}

func (d *DummyService) GetValidatorDashboardName(ctx context.Context, dashboardId t.VDBIdPrimary) (string, error) {
	return getDummyData[string]()
}

func (d *DummyService) GetValidatorsFromSlices(indices []uint64, publicKeys []string) ([]t.VDBValidator, error) {
	return getDummyData[[]t.VDBValidator]()
}

func (d *DummyService) GetUserDashboards(ctx context.Context, userId uint64) (*t.UserDashboardsData, error) {
	return getDummyStruct[t.UserDashboardsData]()
}

func (d *DummyService) CreateValidatorDashboard(ctx context.Context, userId uint64, name string, network uint64) (*t.VDBPostReturnData, error) {
	return getDummyStruct[t.VDBPostReturnData]()
}

func (d *DummyService) GetValidatorDashboardOverview(ctx context.Context, dashboardId t.VDBId, protocolModes t.VDBProtocolModes) (*t.VDBOverviewData, error) {
	return getDummyStruct[t.VDBOverviewData]()
}

func (d *DummyService) RemoveValidatorDashboard(ctx context.Context, dashboardId t.VDBIdPrimary) error {
	return nil
}

func (d *DummyService) UpdateValidatorDashboardArchiving(ctx context.Context, dashboardId t.VDBIdPrimary, archived bool) (*t.VDBPostArchivingReturnData, error) {
	return getDummyStruct[t.VDBPostArchivingReturnData]()
}

func (d *DummyService) UpdateValidatorDashboardName(ctx context.Context, dashboardId t.VDBIdPrimary, name string) (*t.VDBPostReturnData, error) {
	return getDummyStruct[t.VDBPostReturnData]()
}

func (d *DummyService) CreateValidatorDashboardGroup(ctx context.Context, dashboardId t.VDBIdPrimary, name string) (*t.VDBPostCreateGroupData, error) {
	return getDummyStruct[t.VDBPostCreateGroupData]()
}

func (d *DummyService) UpdateValidatorDashboardGroup(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, name string) (*t.VDBPostCreateGroupData, error) {
	return getDummyStruct[t.VDBPostCreateGroupData]()
}

func (d *DummyService) RemoveValidatorDashboardGroup(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64) error {
	return nil
}

func (d *DummyService) GetValidatorDashboardGroupExists(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64) (bool, error) {
	return true, nil
}

func (d *DummyService) GetValidatorDashboardExistingValidatorCount(ctx context.Context, dashboardId t.VDBIdPrimary, validators []t.VDBValidator) (uint64, error) {
	return getDummyData[uint64]()
}

func (d *DummyService) AddValidatorDashboardValidators(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, validators []t.VDBValidator) ([]t.VDBPostValidatorsData, error) {
	return getDummyData[[]t.VDBPostValidatorsData]()
}

func (d *DummyService) AddValidatorDashboardValidatorsByDepositAddress(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, address string, limit uint64) ([]t.VDBPostValidatorsData, error) {
	return getDummyData[[]t.VDBPostValidatorsData]()
}

func (d *DummyService) AddValidatorDashboardValidatorsByWithdrawalAddress(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, address string, limit uint64) ([]t.VDBPostValidatorsData, error) {
	return getDummyData[[]t.VDBPostValidatorsData]()
}

func (d *DummyService) AddValidatorDashboardValidatorsByGraffiti(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, graffiti string, limit uint64) ([]t.VDBPostValidatorsData, error) {
	return getDummyData[[]t.VDBPostValidatorsData]()
}

func (d *DummyService) GetValidatorDashboardValidators(ctx context.Context, dashboardId t.VDBId, groupId int64, cursor string, colSort t.Sort[enums.VDBManageValidatorsColumn], search string, limit uint64) ([]t.VDBManageValidatorsTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.VDBManageValidatorsTableRow]()
}

func (d *DummyService) RemoveValidatorDashboardValidators(ctx context.Context, dashboardId t.VDBIdPrimary, validators []t.VDBValidator) error {
	return nil
}

func (d *DummyService) CreateValidatorDashboardPublicId(ctx context.Context, dashboardId t.VDBIdPrimary, name string, shareGroups bool) (*t.VDBPublicId, error) {
	return getDummyStruct[t.VDBPublicId]()
}

func (d *DummyService) GetValidatorDashboardPublicId(ctx context.Context, publicDashboardId t.VDBIdPublic) (*t.VDBPublicId, error) {
	return getDummyStruct[t.VDBPublicId]()
}

func (d *DummyService) UpdateValidatorDashboardPublicId(ctx context.Context, publicDashboardId t.VDBIdPublic, name string, shareGroups bool) (*t.VDBPublicId, error) {
	return getDummyStruct[t.VDBPublicId]()
}

func (d *DummyService) RemoveValidatorDashboardPublicId(ctx context.Context, publicDashboardId t.VDBIdPublic) error {
	return nil
}

func (d *DummyService) GetValidatorDashboardSlotViz(ctx context.Context, dashboardId t.VDBId, groupIds []uint64) ([]t.SlotVizEpoch, error) {
	r := struct {
		Epochs []t.SlotVizEpoch `faker:"slice_len=4"`
	}{}
	err := commonFakeData(&r)
	return r.Epochs, err
}

func (d *DummyService) GetValidatorDashboardSummary(ctx context.Context, dashboardId t.VDBId, period enums.TimePeriod, cursor string, colSort t.Sort[enums.VDBSummaryColumn], search string, limit uint64, protocolModes t.VDBProtocolModes) ([]t.VDBSummaryTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.VDBSummaryTableRow]()
}
func (d *DummyService) GetValidatorDashboardGroupSummary(ctx context.Context, dashboardId t.VDBId, groupId int64, period enums.TimePeriod, protocolModes t.VDBProtocolModes) (*t.VDBGroupSummaryData, error) {
	return getDummyStruct[t.VDBGroupSummaryData]()
}

func (d *DummyService) GetValidatorDashboardSummaryChart(ctx context.Context, dashboardId t.VDBId, groupIds []int64, efficiency enums.VDBSummaryChartEfficiencyType, aggregation enums.ChartAggregation, afterTs uint64, beforeTs uint64) (*t.ChartData[int, float64], error) {
	return getDummyStruct[t.ChartData[int, float64]]()
}

func (d *DummyService) GetValidatorDashboardSummaryValidators(ctx context.Context, dashboardId t.VDBId, groupId int64) (*t.VDBGeneralSummaryValidators, error) {
	return getDummyStruct[t.VDBGeneralSummaryValidators]()
}
func (d *DummyService) GetValidatorDashboardSyncSummaryValidators(ctx context.Context, dashboardId t.VDBId, groupId int64, period enums.TimePeriod) (*t.VDBSyncSummaryValidators, error) {
	return getDummyStruct[t.VDBSyncSummaryValidators]()
}
func (d *DummyService) GetValidatorDashboardSlashingsSummaryValidators(ctx context.Context, dashboardId t.VDBId, groupId int64, period enums.TimePeriod) (*t.VDBSlashingsSummaryValidators, error) {
	return getDummyStruct[t.VDBSlashingsSummaryValidators]()
}
func (d *DummyService) GetValidatorDashboardProposalSummaryValidators(ctx context.Context, dashboardId t.VDBId, groupId int64, period enums.TimePeriod) (*t.VDBProposalSummaryValidators, error) {
	return getDummyStruct[t.VDBProposalSummaryValidators]()
}

func (d *DummyService) GetValidatorDashboardRewards(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBRewardsColumn], search string, limit uint64, protocolModes t.VDBProtocolModes) ([]t.VDBRewardsTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.VDBRewardsTableRow]()
}

func (d *DummyService) GetValidatorDashboardGroupRewards(ctx context.Context, dashboardId t.VDBId, groupId int64, epoch uint64, protocolModes t.VDBProtocolModes) (*t.VDBGroupRewardsData, error) {
	return getDummyStruct[t.VDBGroupRewardsData]()
}

func (d *DummyService) GetValidatorDashboardRewardsChart(ctx context.Context, dashboardId t.VDBId, protocolModes t.VDBProtocolModes) (*t.ChartData[int, decimal.Decimal], error) {
	return getDummyStruct[t.ChartData[int, decimal.Decimal]]()
}

func (d *DummyService) GetValidatorDashboardDuties(ctx context.Context, dashboardId t.VDBId, epoch uint64, groupId int64, cursor string, colSort t.Sort[enums.VDBDutiesColumn], search string, limit uint64, protocolModes t.VDBProtocolModes) ([]t.VDBEpochDutiesTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.VDBEpochDutiesTableRow]()
}

func (d *DummyService) GetValidatorDashboardBlocks(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBBlocksColumn], search string, limit uint64, protocolModes t.VDBProtocolModes) ([]t.VDBBlocksTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.VDBBlocksTableRow]()
}

func (d *DummyService) GetValidatorDashboardHeatmap(ctx context.Context, dashboardId t.VDBId, protocolModes t.VDBProtocolModes, aggregation enums.ChartAggregation, afterTs uint64, beforeTs uint64) (*t.VDBHeatmap, error) {
	return getDummyStruct[t.VDBHeatmap]()
}

func (d *DummyService) GetValidatorDashboardGroupHeatmap(ctx context.Context, dashboardId t.VDBId, groupId uint64, protocolModes t.VDBProtocolModes, aggregation enums.ChartAggregation, timestamp uint64) (*t.VDBHeatmapTooltipData, error) {
	return getDummyStruct[t.VDBHeatmapTooltipData]()
}

func (d *DummyService) GetValidatorDashboardElDeposits(ctx context.Context, dashboardId t.VDBId, cursor string, search string, limit uint64) ([]t.VDBExecutionDepositsTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.VDBExecutionDepositsTableRow]()
}

func (d *DummyService) GetValidatorDashboardClDeposits(ctx context.Context, dashboardId t.VDBId, cursor string, search string, limit uint64) ([]t.VDBConsensusDepositsTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.VDBConsensusDepositsTableRow]()
}

func (d *DummyService) GetValidatorDashboardTotalElDeposits(ctx context.Context, dashboardId t.VDBId) (*t.VDBTotalExecutionDepositsData, error) {
	return getDummyStruct[t.VDBTotalExecutionDepositsData]()
}

func (d *DummyService) GetValidatorDashboardTotalClDeposits(ctx context.Context, dashboardId t.VDBId) (*t.VDBTotalConsensusDepositsData, error) {
	return getDummyStruct[t.VDBTotalConsensusDepositsData]()
}

func (d *DummyService) GetValidatorDashboardWithdrawals(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBWithdrawalsColumn], search string, limit uint64, protocolModes t.VDBProtocolModes) ([]t.VDBWithdrawalsTableRow, *t.Paging, error) {
	return []t.VDBWithdrawalsTableRow{}, &t.Paging{}, nil
}

func (d *DummyService) GetValidatorDashboardTotalWithdrawals(ctx context.Context, dashboardId t.VDBId, search string, protocolModes t.VDBProtocolModes) (*t.VDBTotalWithdrawalsData, error) {
	return getDummyStruct[t.VDBTotalWithdrawalsData]()
}

func (d *DummyService) GetValidatorDashboardRocketPool(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBRocketPoolColumn], search string, limit uint64) ([]t.VDBRocketPoolTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.VDBRocketPoolTableRow]()
}

func (d *DummyService) GetValidatorDashboardTotalRocketPool(ctx context.Context, dashboardId t.VDBId, search string) (*t.VDBRocketPoolTableRow, error) {
	return getDummyStruct[t.VDBRocketPoolTableRow]()
}

func (d *DummyService) GetValidatorDashboardNodeRocketPool(ctx context.Context, dashboardId t.VDBId, node string) (*t.VDBNodeRocketPoolData, error) {
	return getDummyStruct[t.VDBNodeRocketPoolData]()
}

func (d *DummyService) GetValidatorDashboardRocketPoolMinipools(ctx context.Context, dashboardId t.VDBId, node string, cursor string, colSort t.Sort[enums.VDBRocketPoolMinipoolsColumn], search string, limit uint64) ([]t.VDBRocketPoolMinipoolsTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.VDBRocketPoolMinipoolsTableRow]()
}

func (d *DummyService) GetAllNetworks() ([]t.NetworkInfo, error) {
	return getDummyData[[]t.NetworkInfo]()
}

func (d *DummyService) GetSearchValidatorByIndex(ctx context.Context, chainId, index uint64) (*t.SearchValidator, error) {
	return getDummyStruct[t.SearchValidator]()
}

func (d *DummyService) GetSearchValidatorByPublicKey(ctx context.Context, chainId uint64, publicKey []byte) (*t.SearchValidator, error) {
	return getDummyStruct[t.SearchValidator]()
}

func (d *DummyService) GetSearchValidatorsByDepositAddress(ctx context.Context, chainId uint64, address []byte) (*t.SearchValidatorsByDepositAddress, error) {
	return getDummyStruct[t.SearchValidatorsByDepositAddress]()
}

func (d *DummyService) GetSearchValidatorsByDepositEnsName(ctx context.Context, chainId uint64, ensName string) (*t.SearchValidatorsByDepositEnsName, error) {
	return getDummyStruct[t.SearchValidatorsByDepositEnsName]()
}

func (d *DummyService) GetSearchValidatorsByWithdrawalCredential(ctx context.Context, chainId uint64, credential []byte) (*t.SearchValidatorsByWithdrwalCredential, error) {
	return getDummyStruct[t.SearchValidatorsByWithdrwalCredential]()
}

func (d *DummyService) GetSearchValidatorsByWithdrawalEnsName(ctx context.Context, chainId uint64, ensName string) (*t.SearchValidatorsByWithrawalEnsName, error) {
	return getDummyStruct[t.SearchValidatorsByWithrawalEnsName]()
}

func (d *DummyService) GetSearchValidatorsByGraffiti(ctx context.Context, chainId uint64, graffiti string) (*t.SearchValidatorsByGraffiti, error) {
	return getDummyStruct[t.SearchValidatorsByGraffiti]()
}

func (d *DummyService) GetUserValidatorDashboardCount(ctx context.Context, userId uint64, active bool) (uint64, error) {
	return getDummyData[uint64]()
}

func (d *DummyService) GetValidatorDashboardGroupCount(ctx context.Context, dashboardId t.VDBIdPrimary) (uint64, error) {
	return getDummyData[uint64]()
}

func (d *DummyService) GetValidatorDashboardValidatorsCount(ctx context.Context, dashboardId t.VDBIdPrimary) (uint64, error) {
	return getDummyData[uint64]()
}

func (d *DummyService) GetValidatorDashboardPublicIdCount(ctx context.Context, dashboardId t.VDBIdPrimary) (uint64, error) {
	return getDummyData[uint64]()
}

func (d *DummyService) GetNotificationOverview(ctx context.Context, userId uint64) (*t.NotificationOverviewData, error) {
	return getDummyStruct[t.NotificationOverviewData]()
}
func (d *DummyService) GetDashboardNotifications(ctx context.Context, userId uint64, chainId uint64, cursor string, colSort t.Sort[enums.NotificationDashboardsColumn], search string, limit uint64) ([]t.NotificationDashboardsTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.NotificationDashboardsTableRow]()
}

func (d *DummyService) GetValidatorDashboardNotificationDetails(ctx context.Context, notificationId string) (*t.NotificationValidatorDashboardDetail, error) {
	return getDummyStruct[t.NotificationValidatorDashboardDetail]()
}

func (d *DummyService) GetAccountDashboardNotificationDetails(ctx context.Context, notificationId string) (*t.NotificationAccountDashboardDetail, error) {
	return getDummyStruct[t.NotificationAccountDashboardDetail]()
}

func (d *DummyService) GetMachineNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationMachinesColumn], search string, limit uint64) ([]t.NotificationMachinesTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.NotificationMachinesTableRow]()
}
func (d *DummyService) GetClientNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationClientsColumn], search string, limit uint64) ([]t.NotificationClientsTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.NotificationClientsTableRow]()
}
func (d *DummyService) GetRocketPoolNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationRocketPoolColumn], search string, limit uint64) ([]t.NotificationRocketPoolTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.NotificationRocketPoolTableRow]()
}
func (d *DummyService) GetNetworkNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationNetworksColumn], search string, limit uint64) ([]t.NotificationNetworksTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.NotificationNetworksTableRow]()
}

func (d *DummyService) GetNotificationSettings(ctx context.Context, userId uint64) (*t.NotificationSettings, error) {
	return getDummyStruct[t.NotificationSettings]()
}
func (d *DummyService) UpdateNotificationSettingsGeneral(ctx context.Context, userId uint64, settings t.NotificationSettingsGeneral) error {
	return nil
}
func (d *DummyService) UpdateNotificationSettingsNetworks(ctx context.Context, userId uint64, chainId uint64, settings t.NotificationSettingsNetwork) error {
	return nil
}
func (d *DummyService) UpdateNotificationSettingsPairedDevice(ctx context.Context, pairedDeviceId string, name string, IsNotificationsEnabled bool) error {
	return nil
}
func (d *DummyService) DeleteNotificationSettingsPairedDevice(ctx context.Context, pairedDeviceId string) error {
	return nil
}
func (d *DummyService) GetNotificationSettingsDashboards(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationSettingsDashboardColumn], search string, limit uint64) ([]t.NotificationSettingsDashboardsTableRow, *t.Paging, error) {
	r, p, err := getDummyWithPaging[t.NotificationSettingsDashboardsTableRow]()
	for i, n := range r {
		var settings interface{}
		if n.IsAccountDashboard {
			settings = t.NotificationSettingsAccountDashboard{}
		} else {
			settings = t.NotificationSettingsValidatorDashboard{}
		}
		_ = commonFakeData(&settings)
		r[i].Settings = settings
	}
	return r, p, err
}
func (d *DummyService) UpdateNotificationSettingsValidatorDashboard(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, settings t.NotificationSettingsValidatorDashboard) error {
	return nil
}
func (d *DummyService) UpdateNotificationSettingsAccountDashboard(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, settings t.NotificationSettingsAccountDashboard) error {
	return nil
}
func (d *DummyService) CreateAdConfiguration(ctx context.Context, key, jquerySelector string, insertMode enums.AdInsertMode, refreshInterval uint64, forAllUsers bool, bannerId uint64, htmlContent string, enabled bool) error {
	return nil
}

func (d *DummyService) GetAdConfigurations(ctx context.Context, keys []string) ([]t.AdConfigurationData, error) {
	return getDummyData[[]t.AdConfigurationData]()
}

func (d *DummyService) UpdateAdConfiguration(ctx context.Context, key, jquerySelector string, insertMode enums.AdInsertMode, refreshInterval uint64, forAllUsers bool, bannerId uint64, htmlContent string, enabled bool) error {
	return nil
}

func (d *DummyService) RemoveAdConfiguration(ctx context.Context, key string) error {
	return nil
}

func (d *DummyService) GetLatestExportedChartTs(ctx context.Context, aggregation enums.ChartAggregation) (uint64, error) {
	return getDummyData[uint64]()
}

func (d *DummyService) GetUserIdByRefreshToken(claimUserID, claimAppID, claimDeviceID uint64, hashedRefreshToken string) (uint64, error) {
	return getDummyData[uint64]()
}

func (d *DummyService) MigrateMobileSession(oldHashedRefreshToken, newHashedRefreshToken, deviceID, deviceName string) error {
	return nil
}

func (d *DummyService) GetAppDataFromRedirectUri(callback string) (*t.OAuthAppData, error) {
	return getDummyStruct[t.OAuthAppData]()
}

func (d *DummyService) AddUserDevice(userID uint64, hashedRefreshToken string, deviceID, deviceName string, appID uint64) error {
	return nil
}

func (d *DummyService) AddMobileNotificationToken(userID uint64, deviceID, notifyToken string) error {
	return nil
}

func (d *DummyService) GetAppSubscriptionCount(userID uint64) (uint64, error) {
	return getDummyData[uint64]()
}

func (d *DummyService) AddMobilePurchase(tx *sql.Tx, userID uint64, paymentDetails t.MobileSubscription, verifyResponse *userservice.VerifyResponse, extSubscriptionId string) error {
	return nil
}

func (d *DummyService) GetBlockOverview(ctx context.Context, chainId, block uint64) (*t.BlockOverview, error) {
	return getDummyStruct[t.BlockOverview]()
}

func (d *DummyService) GetBlockTransactions(ctx context.Context, chainId, block uint64) ([]t.BlockTransactionTableRow, error) {
	return getDummyData[[]t.BlockTransactionTableRow]()
}

func (d *DummyService) GetBlock(ctx context.Context, chainId, block uint64) (*t.BlockSummary, error) {
	return getDummyStruct[t.BlockSummary]()
}

func (d *DummyService) GetBlockVotes(ctx context.Context, chainId, block uint64) ([]t.BlockVoteTableRow, error) {
	return getDummyData[[]t.BlockVoteTableRow]()
}

func (d *DummyService) GetBlockAttestations(ctx context.Context, chainId, block uint64) ([]t.BlockAttestationTableRow, error) {
	return getDummyData[[]t.BlockAttestationTableRow]()
}

func (d *DummyService) GetBlockWithdrawals(ctx context.Context, chainId, block uint64) ([]t.BlockWithdrawalTableRow, error) {
	return getDummyData[[]t.BlockWithdrawalTableRow]()
}

func (d *DummyService) GetBlockBlsChanges(ctx context.Context, chainId, block uint64) ([]t.BlockBlsChangeTableRow, error) {
	return getDummyData[[]t.BlockBlsChangeTableRow]()
}

func (d *DummyService) GetBlockVoluntaryExits(ctx context.Context, chainId, block uint64) ([]t.BlockVoluntaryExitTableRow, error) {
	return getDummyData[[]t.BlockVoluntaryExitTableRow]()
}

func (d *DummyService) GetBlockBlobs(ctx context.Context, chainId, block uint64) ([]t.BlockBlobTableRow, error) {
	return getDummyData[[]t.BlockBlobTableRow]()
}

func (d *DummyService) GetSlot(ctx context.Context, chainId, block uint64) (*t.BlockSummary, error) {
	return getDummyStruct[t.BlockSummary]()
}

func (d *DummyService) GetSlotOverview(ctx context.Context, chainId, block uint64) (*t.BlockOverview, error) {
	return getDummyStruct[t.BlockOverview]()
}

func (d *DummyService) GetSlotTransactions(ctx context.Context, chainId, block uint64) ([]t.BlockTransactionTableRow, error) {
	return getDummyData[[]t.BlockTransactionTableRow]()
}

func (d *DummyService) GetSlotVotes(ctx context.Context, chainId, block uint64) ([]t.BlockVoteTableRow, error) {
	return getDummyData[[]t.BlockVoteTableRow]()
}

func (d *DummyService) GetSlotAttestations(ctx context.Context, chainId, block uint64) ([]t.BlockAttestationTableRow, error) {
	return getDummyData[[]t.BlockAttestationTableRow]()
}

func (d *DummyService) GetSlotWithdrawals(ctx context.Context, chainId, block uint64) ([]t.BlockWithdrawalTableRow, error) {
	return getDummyData[[]t.BlockWithdrawalTableRow]()
}

func (d *DummyService) GetSlotBlsChanges(ctx context.Context, chainId, block uint64) ([]t.BlockBlsChangeTableRow, error) {
	return getDummyData[[]t.BlockBlsChangeTableRow]()
}

func (d *DummyService) GetSlotVoluntaryExits(ctx context.Context, chainId, block uint64) ([]t.BlockVoluntaryExitTableRow, error) {
	return getDummyData[[]t.BlockVoluntaryExitTableRow]()
}

func (d *DummyService) GetSlotBlobs(ctx context.Context, chainId, block uint64) ([]t.BlockBlobTableRow, error) {
	return getDummyData[[]t.BlockBlobTableRow]()
}

func (d *DummyService) GetRocketPoolOverview(ctx context.Context) (*t.RocketPoolData, error) {
	return getDummyStruct[t.RocketPoolData]()
}

func (d *DummyService) GetApiWeights(ctx context.Context) ([]t.ApiWeightItem, error) {
	r := []t.ApiWeightItem{}
	err := commonFakeData(&r)
	return r, err
}

func (d *DummyService) GetHealthz(ctx context.Context, showAll bool) types.HealthzData {
	r, _ := getDummyData[types.HealthzData]()
	return r
}
