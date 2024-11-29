package dataaccess

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand/v2"
	"reflect"
	"slices"
	"sync"
	"time"

	mathrand "math/rand"

	"github.com/go-faker/faker/v4"
	"github.com/go-faker/faker/v4/pkg/interfaces"
	"github.com/go-faker/faker/v4/pkg/options"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	commontypes "github.com/gobitfly/beaconchain/pkg/commons/types"
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

// generate random decimal.Decimal, result is between 0.001 and 1000 GWei (returned in Wei)
func randomEthDecimal() decimal.Decimal {
	decimal, _ := decimal.NewFromString(fmt.Sprintf("%d000000", rand.Int64N(1000000)+1)) //nolint:gosec
	return decimal
}

var mockLock sync.Mutex = sync.Mutex{}

// must pass a pointer to the data
func populateWithFakeData(ctx context.Context, a interface{}) error {
	if seed, ok := ctx.Value(t.CtxMockSeedKey).(int64); ok {
		mockLock.Lock()
		defer mockLock.Unlock()
		faker.SetRandomSource(mathrand.NewSource(seed))
	}

	return faker.FakeData(a, options.WithRandomMapAndSliceMaxSize(10), options.WithRandomFloatBoundaries(interfaces.RandomFloatBoundary{Start: 0, End: 1}))
}

func (d *DummyService) StartDataAccessServices() {
	// nothing to start
}

// used for any non-pointer data, e.g. all primitive types or slices
func getDummyData[T any](ctx context.Context) (T, error) {
	var r T
	err := populateWithFakeData(ctx, &r)
	return r, err
}

// used for any struct data that should be returned as a pointer
func getDummyStruct[T any](ctx context.Context) (*T, error) {
	var r T
	err := populateWithFakeData(ctx, &r)
	return &r, err
}

// used for any table data that should be returned with paging
func getDummyWithPaging[T any](ctx context.Context) ([]T, *t.Paging, error) {
	r := []T{}
	p := t.Paging{}
	_ = populateWithFakeData(ctx, &r)
	err := populateWithFakeData(ctx, &p)
	return r, &p, err
}

func (d *DummyService) Close() {
	// nothing to close
}

func (d *DummyService) GetLatestSlot(ctx context.Context) (uint64, error) {
	return getDummyData[uint64](ctx)
}

func (d *DummyService) GetLatestFinalizedEpoch(ctx context.Context) (uint64, error) {
	return getDummyData[uint64](ctx)
}

func (d *DummyService) GetLatestBlock(ctx context.Context) (uint64, error) {
	return getDummyData[uint64](ctx)
}

func (d *DummyService) GetLatestExchangeRates(ctx context.Context) ([]t.EthConversionRate, error) {
	return getDummyData[[]t.EthConversionRate](ctx)
}

func (d *DummyService) GetUserByEmail(ctx context.Context, email string) (uint64, error) {
	return getDummyData[uint64](ctx)
}

func (d *DummyService) CreateUser(ctx context.Context, email, password string) (uint64, error) {
	return getDummyData[uint64](ctx)
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
	return getDummyData[time.Time](ctx)
}

func (d *DummyService) GetPasswordResetTime(ctx context.Context, userId uint64) (time.Time, error) {
	return getDummyData[time.Time](ctx)
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
	return getDummyStruct[t.UserInfo](ctx)
}

func (d *DummyService) GetUserCredentialInfo(ctx context.Context, userId uint64) (*t.UserCredentialInfo, error) {
	return getDummyStruct[t.UserCredentialInfo](ctx)
}

func (d *DummyService) GetUserIdByApiKey(ctx context.Context, apiKey string) (uint64, error) {
	return getDummyData[uint64](ctx)
}

func (d *DummyService) GetUserIdByConfirmationHash(ctx context.Context, hash string) (uint64, error) {
	return getDummyData[uint64](ctx)
}

func (d *DummyService) GetUserIdByResetHash(ctx context.Context, hash string) (uint64, error) {
	return getDummyData[uint64](ctx)
}

func (d *DummyService) GetProductSummary(ctx context.Context) (*t.ProductSummary, error) {
	return getDummyStruct[t.ProductSummary](ctx)
}

func (d *DummyService) GetFreeTierPerks(ctx context.Context) (*t.PremiumPerks, error) {
	return getDummyStruct[t.PremiumPerks](ctx)
}

func (d *DummyService) GetValidatorDashboardUser(ctx context.Context, dashboardId t.VDBIdPrimary) (*t.DashboardUser, error) {
	return getDummyStruct[t.DashboardUser](ctx)
}

func (d *DummyService) GetValidatorDashboardIdByPublicId(ctx context.Context, publicDashboardId t.VDBIdPublic) (*t.VDBIdPrimary, error) {
	return getDummyStruct[t.VDBIdPrimary](ctx)
}

func (d *DummyService) GetValidatorDashboardInfo(ctx context.Context, dashboardId t.VDBIdPrimary) (*t.ValidatorDashboard, error) {
	r, err := getDummyStruct[t.ValidatorDashboard](ctx)
	// return semi-valid data to not break staging
	r.IsArchived = false
	return r, err
}

func (d *DummyService) GetValidatorDashboardName(ctx context.Context, dashboardId t.VDBIdPrimary) (string, error) {
	return getDummyData[string](ctx)
}

func (d *DummyService) GetValidatorsFromSlices(ctx context.Context, indices []uint64, publicKeys []string) ([]t.VDBValidator, error) {
	return getDummyData[[]t.VDBValidator](ctx)
}

func (d *DummyService) GetUserDashboards(ctx context.Context, userId uint64) (*t.UserDashboardsData, error) {
	return getDummyStruct[t.UserDashboardsData](ctx)
}

func (d *DummyService) CreateValidatorDashboard(ctx context.Context, userId uint64, name string, network uint64) (*t.VDBPostReturnData, error) {
	return getDummyStruct[t.VDBPostReturnData](ctx)
}

func (d *DummyService) GetValidatorDashboardOverview(ctx context.Context, dashboardId t.VDBId, protocolModes t.VDBProtocolModes) (*t.VDBOverviewData, error) {
	return getDummyStruct[t.VDBOverviewData](ctx)
}

func (d *DummyService) RemoveValidatorDashboard(ctx context.Context, dashboardId t.VDBIdPrimary) error {
	return nil
}

func (d *DummyService) RemoveValidatorDashboards(ctx context.Context, dashboardIds []uint64) error {
	return nil
}

func (d *DummyService) UpdateValidatorDashboardArchiving(ctx context.Context, dashboardId t.VDBIdPrimary, archivedReason *enums.VDBArchivedReason) (*t.VDBPostArchivingReturnData, error) {
	return getDummyStruct[t.VDBPostArchivingReturnData](ctx)
}

func (d *DummyService) UpdateValidatorDashboardsArchiving(ctx context.Context, dashboards []t.ArchiverDashboardArchiveReason) error {
	return nil
}

func (d *DummyService) UpdateValidatorDashboardName(ctx context.Context, dashboardId t.VDBIdPrimary, name string) (*t.VDBPostReturnData, error) {
	return getDummyStruct[t.VDBPostReturnData](ctx)
}

func (d *DummyService) CreateValidatorDashboardGroup(ctx context.Context, dashboardId t.VDBIdPrimary, name string) (*t.VDBPostCreateGroupData, error) {
	return getDummyStruct[t.VDBPostCreateGroupData](ctx)
}

func (d *DummyService) UpdateValidatorDashboardGroup(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, name string) (*t.VDBPostCreateGroupData, error) {
	return getDummyStruct[t.VDBPostCreateGroupData](ctx)
}

func (d *DummyService) RemoveValidatorDashboardGroup(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64) error {
	return nil
}

func (d *DummyService) RemoveValidatorDashboardGroupValidators(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64) error {
	return nil
}

func (d *DummyService) GetValidatorDashboardGroupExists(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64) (bool, error) {
	return true, nil
}

func (d *DummyService) AddValidatorDashboardValidators(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, validators []t.VDBValidator) ([]t.VDBPostValidatorsData, error) {
	return getDummyData[[]t.VDBPostValidatorsData](ctx)
}

func (d *DummyService) AddValidatorDashboardValidatorsByDepositAddress(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, address string, limit uint64) ([]t.VDBPostValidatorsData, error) {
	return getDummyData[[]t.VDBPostValidatorsData](ctx)
}

func (d *DummyService) AddValidatorDashboardValidatorsByWithdrawalCredential(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, address string, limit uint64) ([]t.VDBPostValidatorsData, error) {
	return getDummyData[[]t.VDBPostValidatorsData](ctx)
}

func (d *DummyService) AddValidatorDashboardValidatorsByGraffiti(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, graffiti string, limit uint64) ([]t.VDBPostValidatorsData, error) {
	return getDummyData[[]t.VDBPostValidatorsData](ctx)
}

func (d *DummyService) GetValidatorDashboardValidators(ctx context.Context, dashboardId t.VDBId, groupId int64, cursor string, colSort t.Sort[enums.VDBManageValidatorsColumn], search string, limit uint64) ([]t.VDBManageValidatorsTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.VDBManageValidatorsTableRow](ctx)
}

func (d *DummyService) RemoveValidatorDashboardValidators(ctx context.Context, dashboardId t.VDBIdPrimary, validators []t.VDBValidator) error {
	return nil
}

func (d *DummyService) CreateValidatorDashboardPublicId(ctx context.Context, dashboardId t.VDBIdPrimary, name string, shareGroups bool) (*t.VDBPublicId, error) {
	return getDummyStruct[t.VDBPublicId](ctx)
}

func (d *DummyService) GetValidatorDashboardPublicId(ctx context.Context, publicDashboardId t.VDBIdPublic) (*t.VDBPublicId, error) {
	return getDummyStruct[t.VDBPublicId](ctx)
}

func (d *DummyService) UpdateValidatorDashboardPublicId(ctx context.Context, publicDashboardId t.VDBIdPublic, name string, shareGroups bool) (*t.VDBPublicId, error) {
	return getDummyStruct[t.VDBPublicId](ctx)
}

func (d *DummyService) RemoveValidatorDashboardPublicId(ctx context.Context, publicDashboardId t.VDBIdPublic) error {
	return nil
}

func (d *DummyService) GetValidatorDashboardSlotViz(ctx context.Context, dashboardId t.VDBId, groupIds []uint64) ([]t.SlotVizEpoch, error) {
	r := struct {
		Epochs []t.SlotVizEpoch `faker:"slice_len=4"`
	}{}
	err := populateWithFakeData(ctx, &r)
	return r.Epochs, err
}

func (d *DummyService) GetValidatorDashboardSummary(ctx context.Context, dashboardId t.VDBId, period enums.TimePeriod, cursor string, colSort t.Sort[enums.VDBSummaryColumn], search string, limit uint64, protocolModes t.VDBProtocolModes) ([]t.VDBSummaryTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.VDBSummaryTableRow](ctx)
}
func (d *DummyService) GetValidatorDashboardGroupSummary(ctx context.Context, dashboardId t.VDBId, groupId int64, period enums.TimePeriod, protocolModes t.VDBProtocolModes) (*t.VDBGroupSummaryData, error) {
	return getDummyStruct[t.VDBGroupSummaryData](ctx)
}

func (d *DummyService) GetValidatorDashboardSummaryChart(ctx context.Context, dashboardId t.VDBId, groupIds []int64, efficiency enums.VDBSummaryChartEfficiencyType, aggregation enums.ChartAggregation, afterTs uint64, beforeTs uint64) (*t.ChartData[int, float64], error) {
	return getDummyStruct[t.ChartData[int, float64]](ctx)
}

func (d *DummyService) GetValidatorDashboardSummaryValidators(ctx context.Context, dashboardId t.VDBId, groupId int64) (*t.VDBGeneralSummaryValidators, error) {
	return getDummyStruct[t.VDBGeneralSummaryValidators](ctx)
}
func (d *DummyService) GetValidatorDashboardSyncSummaryValidators(ctx context.Context, dashboardId t.VDBId, groupId int64, period enums.TimePeriod) (*t.VDBSyncSummaryValidators, error) {
	return getDummyStruct[t.VDBSyncSummaryValidators](ctx)
}
func (d *DummyService) GetValidatorDashboardSlashingsSummaryValidators(ctx context.Context, dashboardId t.VDBId, groupId int64, period enums.TimePeriod) (*t.VDBSlashingsSummaryValidators, error) {
	return getDummyStruct[t.VDBSlashingsSummaryValidators](ctx)
}
func (d *DummyService) GetValidatorDashboardProposalSummaryValidators(ctx context.Context, dashboardId t.VDBId, groupId int64, period enums.TimePeriod) (*t.VDBProposalSummaryValidators, error) {
	return getDummyStruct[t.VDBProposalSummaryValidators](ctx)
}

func (d *DummyService) GetValidatorDashboardRewards(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBRewardsColumn], search string, limit uint64, protocolModes t.VDBProtocolModes) ([]t.VDBRewardsTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.VDBRewardsTableRow](ctx)
}

func (d *DummyService) GetValidatorDashboardGroupRewards(ctx context.Context, dashboardId t.VDBId, groupId int64, epoch uint64, protocolModes t.VDBProtocolModes) (*t.VDBGroupRewardsData, error) {
	return getDummyStruct[t.VDBGroupRewardsData](ctx)
}

func (d *DummyService) GetValidatorDashboardRewardsChart(ctx context.Context, dashboardId t.VDBId, protocolModes t.VDBProtocolModes) (*t.ChartData[int, decimal.Decimal], error) {
	return getDummyStruct[t.ChartData[int, decimal.Decimal]](ctx)
}

func (d *DummyService) GetValidatorDashboardDuties(ctx context.Context, dashboardId t.VDBId, epoch uint64, groupId int64, cursor string, colSort t.Sort[enums.VDBDutiesColumn], search string, limit uint64, protocolModes t.VDBProtocolModes) ([]t.VDBEpochDutiesTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.VDBEpochDutiesTableRow](ctx)
}

func (d *DummyService) GetValidatorDashboardBlocks(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBBlocksColumn], search enums.VDBBlocksSearches, limit uint64, protocolModes t.VDBProtocolModes) ([]t.VDBBlocksTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.VDBBlocksTableRow](ctx)
}

func (d *DummyService) GetValidatorDashboardHeatmap(ctx context.Context, dashboardId t.VDBId, protocolModes t.VDBProtocolModes, aggregation enums.ChartAggregation, afterTs uint64, beforeTs uint64) (*t.VDBHeatmap, error) {
	return getDummyStruct[t.VDBHeatmap](ctx)
}

func (d *DummyService) GetValidatorDashboardGroupHeatmap(ctx context.Context, dashboardId t.VDBId, groupId uint64, protocolModes t.VDBProtocolModes, aggregation enums.ChartAggregation, timestamp uint64) (*t.VDBHeatmapTooltipData, error) {
	return getDummyStruct[t.VDBHeatmapTooltipData](ctx)
}

func (d *DummyService) GetValidatorDashboardElDeposits(ctx context.Context, dashboardId t.VDBId, cursor string, limit uint64) ([]t.VDBExecutionDepositsTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.VDBExecutionDepositsTableRow](ctx)
}

func (d *DummyService) GetValidatorDashboardClDeposits(ctx context.Context, dashboardId t.VDBId, cursor string, limit uint64) ([]t.VDBConsensusDepositsTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.VDBConsensusDepositsTableRow](ctx)
}

func (d *DummyService) GetValidatorDashboardTotalElDeposits(ctx context.Context, dashboardId t.VDBId) (*t.VDBTotalExecutionDepositsData, error) {
	return getDummyStruct[t.VDBTotalExecutionDepositsData](ctx)
}

func (d *DummyService) GetValidatorDashboardTotalClDeposits(ctx context.Context, dashboardId t.VDBId) (*t.VDBTotalConsensusDepositsData, error) {
	return getDummyStruct[t.VDBTotalConsensusDepositsData](ctx)
}

func (d *DummyService) GetValidatorDashboardWithdrawals(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBWithdrawalsColumn], search string, limit uint64, protocolModes t.VDBProtocolModes) ([]t.VDBWithdrawalsTableRow, *t.Paging, error) {
	return []t.VDBWithdrawalsTableRow{}, &t.Paging{}, nil
}

func (d *DummyService) GetValidatorDashboardTotalWithdrawals(ctx context.Context, dashboardId t.VDBId, search string, protocolModes t.VDBProtocolModes) (*t.VDBTotalWithdrawalsData, error) {
	return getDummyStruct[t.VDBTotalWithdrawalsData](ctx)
}

func (d *DummyService) GetValidatorDashboardRocketPool(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBRocketPoolColumn], search string, limit uint64) ([]t.VDBRocketPoolTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.VDBRocketPoolTableRow](ctx)
}

func (d *DummyService) GetValidatorDashboardTotalRocketPool(ctx context.Context, dashboardId t.VDBId, search string) (*t.VDBRocketPoolTableRow, error) {
	return getDummyStruct[t.VDBRocketPoolTableRow](ctx)
}

func (d *DummyService) GetValidatorDashboardRocketPoolMinipools(ctx context.Context, dashboardId t.VDBId, node string, cursor string, colSort t.Sort[enums.VDBRocketPoolMinipoolsColumn], search string, limit uint64) ([]t.VDBRocketPoolMinipoolsTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.VDBRocketPoolMinipoolsTableRow](ctx)
}

func (d *DummyService) GetAllNetworks() ([]t.NetworkInfo, error) {
	return []t.NetworkInfo{
		{
			ChainId:           1,
			Name:              "ethereum",
			NotificationsName: "mainnet",
		},
		{
			ChainId:           100,
			Name:              "gnosis",
			NotificationsName: "gnosis",
		},
		{
			ChainId:           17000,
			Name:              "holesky",
			NotificationsName: "holesky",
		},
	}, nil
}

func (d *DummyService) GetAllClients() ([]t.ClientInfo, error) {
	return []t.ClientInfo{
		// execution_layer
		{
			Id:       0,
			Name:     "Geth",
			DbName:   "geth",
			Category: "execution_layer",
		},
		{
			Id:       1,
			Name:     "Nethermind",
			DbName:   "nethermind",
			Category: "execution_layer",
		},
		{
			Id:       2,
			Name:     "Besu",
			DbName:   "besu",
			Category: "execution_layer",
		},
		{
			Id:       3,
			Name:     "Erigon",
			DbName:   "erigon",
			Category: "execution_layer",
		},
		{
			Id:       4,
			Name:     "Reth",
			DbName:   "reth",
			Category: "execution_layer",
		},
		// consensus_layer
		{
			Id:       5,
			Name:     "Teku",
			DbName:   "teku",
			Category: "consensus_layer",
		},
		{
			Id:       6,
			Name:     "Prysm",
			DbName:   "prysm",
			Category: "consensus_layer",
		},
		{
			Id:       7,
			Name:     "Nimbus",
			DbName:   "nimbus",
			Category: "consensus_layer",
		},
		{
			Id:       8,
			Name:     "Lighthouse",
			DbName:   "lighthouse",
			Category: "consensus_layer",
		},
		{
			Id:       9,
			Name:     "Lodestar",
			DbName:   "lodestar",
			Category: "consensus_layer",
		},
		// other
		{
			Id:       10,
			Name:     "Rocketpool Smart Node",
			DbName:   "rocketpool",
			Category: "other",
		},
		{
			Id:       11,
			Name:     "MEV-Boost",
			DbName:   "mev-boost",
			Category: "other",
		},
	}, nil
}

func (d *DummyService) GetSearchValidatorByIndex(ctx context.Context, chainId, index uint64) (*t.SearchValidator, error) {
	return getDummyStruct[t.SearchValidator](ctx)
}

func (d *DummyService) GetSearchValidatorByPublicKey(ctx context.Context, chainId uint64, publicKey []byte) (*t.SearchValidator, error) {
	return getDummyStruct[t.SearchValidator](ctx)
}

func (d *DummyService) GetSearchValidatorsByDepositAddress(ctx context.Context, chainId uint64, address []byte) (*t.SearchValidatorsByDepositAddress, error) {
	return getDummyStruct[t.SearchValidatorsByDepositAddress](ctx)
}

func (d *DummyService) GetSearchValidatorsByDepositEnsName(ctx context.Context, chainId uint64, ensName string) (*t.SearchValidatorsByDepositAddress, error) {
	return getDummyStruct[t.SearchValidatorsByDepositAddress](ctx)
}

func (d *DummyService) GetSearchValidatorsByWithdrawalCredential(ctx context.Context, chainId uint64, credential []byte) (*t.SearchValidatorsByWithdrwalCredential, error) {
	return getDummyStruct[t.SearchValidatorsByWithdrwalCredential](ctx)
}

func (d *DummyService) GetSearchValidatorsByWithdrawalEnsName(ctx context.Context, chainId uint64, ensName string) (*t.SearchValidatorsByWithdrwalCredential, error) {
	return getDummyStruct[t.SearchValidatorsByWithdrwalCredential](ctx)
}

func (d *DummyService) GetSearchValidatorsByGraffiti(ctx context.Context, chainId uint64, graffiti string) (*t.SearchValidatorsByGraffiti, error) {
	return getDummyStruct[t.SearchValidatorsByGraffiti](ctx)
}

func (d *DummyService) GetUserValidatorDashboardCount(ctx context.Context, userId uint64, active bool) (uint64, error) {
	return getDummyData[uint64](ctx)
}

func (d *DummyService) GetValidatorDashboardGroupCount(ctx context.Context, dashboardId t.VDBIdPrimary) (uint64, error) {
	return getDummyData[uint64](ctx)
}

func (d *DummyService) GetValidatorDashboardValidatorsCount(ctx context.Context, dashboardId t.VDBIdPrimary) (uint64, error) {
	return getDummyData[uint64](ctx)
}

func (d *DummyService) GetValidatorDashboardPublicIdCount(ctx context.Context, dashboardId t.VDBIdPrimary) (uint64, error) {
	return getDummyData[uint64](ctx)
}

func (d *DummyService) GetNotificationOverview(ctx context.Context, userId uint64) (*t.NotificationOverviewData, error) {
	return getDummyStruct[t.NotificationOverviewData](ctx)
}
func (d *DummyService) GetDashboardNotifications(ctx context.Context, userId uint64, chainIds []uint64, cursor string, colSort t.Sort[enums.NotificationDashboardsColumn], search string, limit uint64) ([]t.NotificationDashboardsTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.NotificationDashboardsTableRow](ctx)
}

func (d *DummyService) GetValidatorDashboardNotificationDetails(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, epoch uint64, search string) (*t.NotificationValidatorDashboardDetail, error) {
	return getDummyStruct[t.NotificationValidatorDashboardDetail](ctx)
}

func (d *DummyService) GetAccountDashboardNotificationDetails(ctx context.Context, dashboardId uint64, groupId uint64, epoch uint64, search string) (*t.NotificationAccountDashboardDetail, error) {
	return getDummyStruct[t.NotificationAccountDashboardDetail](ctx)
}

func (d *DummyService) GetMachineNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationMachinesColumn], search string, limit uint64) ([]t.NotificationMachinesTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.NotificationMachinesTableRow](ctx)
}
func (d *DummyService) GetClientNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationClientsColumn], search string, limit uint64) ([]t.NotificationClientsTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.NotificationClientsTableRow](ctx)
}
func (d *DummyService) GetNetworkNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationNetworksColumn], limit uint64) ([]t.NotificationNetworksTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.NotificationNetworksTableRow](ctx)
}

func (d *DummyService) GetNotificationSettings(ctx context.Context, userId uint64) (*t.NotificationSettings, error) {
	return getDummyStruct[t.NotificationSettings](ctx)
}
func (d *DummyService) GetNotificationSettingsDefaultValues(ctx context.Context) (*t.NotificationSettingsDefaultValues, error) {
	return getDummyStruct[t.NotificationSettingsDefaultValues](ctx)
}
func (d *DummyService) UpdateNotificationSettingsGeneral(ctx context.Context, userId uint64, settings t.NotificationSettingsGeneral) error {
	return nil
}
func (d *DummyService) UpdateNotificationSettingsNetworks(ctx context.Context, userId uint64, chainId uint64, settings t.NotificationSettingsNetwork) error {
	return nil
}
func (d *DummyService) UpdateNotificationSettingsPairedDevice(ctx context.Context, pairedDeviceId uint64, name string, IsNotificationsEnabled bool) error {
	return nil
}
func (d *DummyService) DeleteNotificationSettingsPairedDevice(ctx context.Context, pairedDeviceId uint64) error {
	return nil
}

func (d *DummyService) UpdateNotificationSettingsClients(ctx context.Context, userId uint64, clientId uint64, IsSubscribed bool) (*t.NotificationSettingsClient, error) {
	return getDummyStruct[t.NotificationSettingsClient](ctx)
}

func (d *DummyService) GetNotificationSettingsDashboards(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationSettingsDashboardColumn], search string, limit uint64) ([]t.NotificationSettingsDashboardsTableRow, *t.Paging, error) {
	r, p, err := getDummyWithPaging[t.NotificationSettingsDashboardsTableRow](ctx)
	for i, n := range r {
		var settings interface{}
		if n.IsAccountDashboard {
			settings = t.NotificationSettingsAccountDashboard{}
		} else {
			settings = t.NotificationSettingsValidatorDashboard{}
		}
		_ = populateWithFakeData(ctx, &settings)
		r[i].Settings = settings
	}
	return r, p, err
}
func (d *DummyService) UpdateNotificationSettingsValidatorDashboard(ctx context.Context, userId uint64, dashboardId t.VDBIdPrimary, groupId uint64, settings t.NotificationSettingsValidatorDashboard) error {
	return nil
}
func (d *DummyService) UpdateNotificationSettingsAccountDashboard(ctx context.Context, userId uint64, dashboardId t.VDBIdPrimary, groupId uint64, settings t.NotificationSettingsAccountDashboard) error {
	return nil
}
func (d *DummyService) CreateAdConfiguration(ctx context.Context, key, jquerySelector string, insertMode enums.AdInsertMode, refreshInterval uint64, forAllUsers bool, bannerId uint64, htmlContent string, enabled bool) error {
	return nil
}

func (d *DummyService) GetAdConfigurations(ctx context.Context, keys []string) ([]t.AdConfigurationData, error) {
	return getDummyData[[]t.AdConfigurationData](ctx)
}

func (d *DummyService) UpdateAdConfiguration(ctx context.Context, key, jquerySelector string, insertMode enums.AdInsertMode, refreshInterval uint64, forAllUsers bool, bannerId uint64, htmlContent string, enabled bool) error {
	return nil
}

func (d *DummyService) RemoveAdConfiguration(ctx context.Context, key string) error {
	return nil
}

func (d *DummyService) GetLatestExportedChartTs(ctx context.Context, aggregation enums.ChartAggregation) (uint64, error) {
	return getDummyData[uint64](ctx)
}

func (d *DummyService) GetUserIdByRefreshToken(ctx context.Context, claimUserID, claimAppID, claimDeviceID uint64, hashedRefreshToken string) (uint64, error) {
	return getDummyData[uint64](ctx)
}

func (d *DummyService) MigrateMobileSession(ctx context.Context, oldHashedRefreshToken, newHashedRefreshToken, deviceID, deviceName string) error {
	return nil
}

func (d *DummyService) GetAppDataFromRedirectUri(ctx context.Context, callback string) (*t.OAuthAppData, error) {
	return getDummyStruct[t.OAuthAppData](ctx)
}

func (d *DummyService) AddUserDevice(ctx context.Context, userID uint64, hashedRefreshToken string, deviceID, deviceName string, appID uint64) error {
	return nil
}

func (d *DummyService) AddMobileNotificationToken(ctx context.Context, userID uint64, deviceID, notifyToken string) error {
	return nil
}

func (d *DummyService) GetAppSubscriptionCount(ctx context.Context, userID uint64) (uint64, error) {
	return getDummyData[uint64](ctx)
}

func (d *DummyService) AddMobilePurchase(ctx context.Context, tx *sql.Tx, userID uint64, paymentDetails t.MobileSubscription, verifyResponse *userservice.VerifyResponse, extSubscriptionId string) error {
	return nil
}

func (d *DummyService) GetBlockOverview(ctx context.Context, chainId, block uint64) (*t.BlockOverview, error) {
	return getDummyStruct[t.BlockOverview](ctx)
}

func (d *DummyService) GetBlockTransactions(ctx context.Context, chainId, block uint64) ([]t.BlockTransactionTableRow, error) {
	return getDummyData[[]t.BlockTransactionTableRow](ctx)
}

func (d *DummyService) GetBlock(ctx context.Context, chainId, block uint64) (*t.BlockSummary, error) {
	return getDummyStruct[t.BlockSummary](ctx)
}

func (d *DummyService) GetBlockVotes(ctx context.Context, chainId, block uint64) ([]t.BlockVoteTableRow, error) {
	return getDummyData[[]t.BlockVoteTableRow](ctx)
}

func (d *DummyService) GetBlockAttestations(ctx context.Context, chainId, block uint64) ([]t.BlockAttestationTableRow, error) {
	return getDummyData[[]t.BlockAttestationTableRow](ctx)
}

func (d *DummyService) GetBlockWithdrawals(ctx context.Context, chainId, block uint64) ([]t.BlockWithdrawalTableRow, error) {
	return getDummyData[[]t.BlockWithdrawalTableRow](ctx)
}

func (d *DummyService) GetBlockBlsChanges(ctx context.Context, chainId, block uint64) ([]t.BlockBlsChangeTableRow, error) {
	return getDummyData[[]t.BlockBlsChangeTableRow](ctx)
}

func (d *DummyService) GetBlockVoluntaryExits(ctx context.Context, chainId, block uint64) ([]t.BlockVoluntaryExitTableRow, error) {
	return getDummyData[[]t.BlockVoluntaryExitTableRow](ctx)
}

func (d *DummyService) GetBlockBlobs(ctx context.Context, chainId, block uint64) ([]t.BlockBlobTableRow, error) {
	return getDummyData[[]t.BlockBlobTableRow](ctx)
}

func (d *DummyService) GetSlot(ctx context.Context, chainId, block uint64) (*t.BlockSummary, error) {
	return getDummyStruct[t.BlockSummary](ctx)
}

func (d *DummyService) GetSlotOverview(ctx context.Context, chainId, block uint64) (*t.BlockOverview, error) {
	return getDummyStruct[t.BlockOverview](ctx)
}

func (d *DummyService) GetSlotTransactions(ctx context.Context, chainId, block uint64) ([]t.BlockTransactionTableRow, error) {
	return getDummyData[[]t.BlockTransactionTableRow](ctx)
}

func (d *DummyService) GetSlotVotes(ctx context.Context, chainId, block uint64) ([]t.BlockVoteTableRow, error) {
	return getDummyData[[]t.BlockVoteTableRow](ctx)
}

func (d *DummyService) GetSlotAttestations(ctx context.Context, chainId, block uint64) ([]t.BlockAttestationTableRow, error) {
	return getDummyData[[]t.BlockAttestationTableRow](ctx)
}

func (d *DummyService) GetSlotWithdrawals(ctx context.Context, chainId, block uint64) ([]t.BlockWithdrawalTableRow, error) {
	return getDummyData[[]t.BlockWithdrawalTableRow](ctx)
}

func (d *DummyService) GetSlotBlsChanges(ctx context.Context, chainId, block uint64) ([]t.BlockBlsChangeTableRow, error) {
	return getDummyData[[]t.BlockBlsChangeTableRow](ctx)
}

func (d *DummyService) GetSlotVoluntaryExits(ctx context.Context, chainId, block uint64) ([]t.BlockVoluntaryExitTableRow, error) {
	return getDummyData[[]t.BlockVoluntaryExitTableRow](ctx)
}

func (d *DummyService) GetSlotBlobs(ctx context.Context, chainId, block uint64) ([]t.BlockBlobTableRow, error) {
	return getDummyData[[]t.BlockBlobTableRow](ctx)
}

func (d *DummyService) GetValidatorDashboardsCountInfo(ctx context.Context) (map[uint64][]t.ArchiverDashboard, error) {
	return getDummyData[map[uint64][]t.ArchiverDashboard](ctx)
}

func (d *DummyService) GetRocketPoolOverview(ctx context.Context) (*t.RocketPoolData, error) {
	return getDummyStruct[t.RocketPoolData](ctx)
}

func (d *DummyService) GetApiWeights(ctx context.Context) ([]t.ApiWeightItem, error) {
	return getDummyData[[]t.ApiWeightItem](ctx)
}

func (d *DummyService) GetHealthz(ctx context.Context, showAll bool) t.HealthzData {
	r, _ := getDummyData[t.HealthzData](ctx)
	return r
}

func (d *DummyService) GetLatestBundleForNativeVersion(ctx context.Context, nativeVersion uint64) (*t.MobileAppBundleStats, error) {
	return getDummyStruct[t.MobileAppBundleStats](ctx)
}

func (d *DummyService) IncrementBundleDeliveryCount(ctx context.Context, bundleVerison uint64) error {
	return nil
}

func (d *DummyService) GetValidatorDashboardMobileWidget(ctx context.Context, dashboardId t.VDBIdPrimary) (*t.MobileWidgetData, error) {
	return getDummyStruct[t.MobileWidgetData](ctx)
}

func (d *DummyService) GetUserMachineMetrics(ctx context.Context, userID uint64, limit int, offset int) (*t.MachineMetricsData, error) {
	data, err := getDummyStruct[t.MachineMetricsData](ctx)
	if err != nil {
		return nil, err
	}
	data.SystemMetrics = slices.SortedFunc(slices.Values(data.SystemMetrics), func(i, j *commontypes.MachineMetricSystem) int {
		return int(i.Timestamp) - int(j.Timestamp)
	})
	data.ValidatorMetrics = slices.SortedFunc(slices.Values(data.ValidatorMetrics), func(i, j *commontypes.MachineMetricValidator) int {
		return int(i.Timestamp) - int(j.Timestamp)
	})
	data.NodeMetrics = slices.SortedFunc(slices.Values(data.NodeMetrics), func(i, j *commontypes.MachineMetricNode) int {
		return int(i.Timestamp) - int(j.Timestamp)
	})
	return data, nil
}

func (d *DummyService) PostUserMachineMetrics(ctx context.Context, userID uint64, machine, process string, data []byte) error {
	return nil
}

func (d *DummyService) GetValidatorDashboardMobileValidators(ctx context.Context, dashboardId t.VDBId, groupId int64, period enums.TimePeriod, cursor string, colSort t.Sort[enums.VDBManageValidatorsColumn], search string, limit uint64) ([]t.MobileValidatorDashboardValidatorsTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.MobileValidatorDashboardValidatorsTableRow](ctx)
}

func (d *DummyService) QueueTestEmailNotification(ctx context.Context, userId uint64) error {
	return nil
}
func (d *DummyService) QueueTestPushNotification(ctx context.Context, userId uint64) error {
	return nil
}
func (d *DummyService) QueueTestWebhookNotification(ctx context.Context, userId uint64, webhookUrl string, isDiscordWebhook bool) error {
	return nil
}

func (d *DummyService) GetPairedDeviceUserId(ctx context.Context, pairedDeviceId uint64) (uint64, error) {
	return getDummyData[uint64](ctx)
}
