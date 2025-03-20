package dataaccess

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand/v2"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"time"

	mathrand "math/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-faker/faker/v4"
	"github.com/go-faker/faker/v4/pkg/interfaces"
	"github.com/go-faker/faker/v4/pkg/options"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	commontypes "github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/userservice"
	"github.com/lucasjones/reggen"
	"github.com/shopspring/decimal"
)

type DummyService struct{}

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
	_ = faker.AddProvider("past_timestamp", func(v reflect.Value) (interface{}, error) {
		past_timestamp, _ := time.Parse("2006-Jan-02", "2023-Jan-01")
		newer_timestamp, _ := time.Parse("2006-Jan-02", "2025-Jan-01")
		return randomTimestamp(past_timestamp, newer_timestamp), nil
	})
	_ = faker.AddProvider("future_timestamp", func(v reflect.Value) (interface{}, error) {
		older_timestamp, _ := time.Parse("2006-Jan-02", "2026-Jan-01")
		future_timestamp, _ := time.Parse("2006-Jan-02", "2028-Jan-01")
		return randomTimestamp(older_timestamp, future_timestamp), nil
	})
	addTagFromRegex("address", t.ReEthereumAddress, func(s string) interface{} {
		// convert to EIP55
		return common.HexToAddress(s).Hex()
	})
	addTagFromRegex("ens", t.ReEnsName, func(s string) interface{} { return s })
	addTagFromRegex("pubkey", t.ReValidatorPublicKeyWithPrefix, func(s string) interface{} { return strings.ToLower(s) })
	addTagFromRegex("tx_hash", t.ReTransactionHash, func(s string) interface{} { return strings.ToLower(s) })
	addTagFromRegex("withdrawal_credentials", t.ReWithdrawalCredential, func(s string) interface{} {
		s = strings.ToLower(s)
		if !strings.HasPrefix(s, "0x") {
			s = "0x" + s
		}
		return s
	})
	return &DummyService{}
}

// generate random string matching the provided regex
// accepts a function to apply formatting to the generated string
func addTagFromRegex(name string, regex *regexp.Regexp, format func(string) interface{}) {
	_ = faker.AddProvider(name, func(v reflect.Value) (interface{}, error) {
		gen, err := reggen.NewGenerator(regex.String())
		if err != nil {
			return nil, err
		}
		gen.SetSeed(source.Int63())
		s := gen.Generate(10)
		return format(s), nil
	})
}

// generate random decimal.Decimal, result is between 0.001 and 1000 GWei (returned in Wei)
func randomEthDecimal() decimal.Decimal {
	decimal, _ := decimal.NewFromString(fmt.Sprintf("%d000000", randomIntFromSeed(1000000)))
	return decimal
}

func randomIntFromSeed(max int64) int64 {
	return source.Int63() % max
}

// generate random timestamp between two dates
func randomTimestamp(t1, t2 time.Time) int64 {
	min, max := t1.Unix(), t2.Unix()
	if max < min {
		min, max = max, min
	}
	return randomIntFromSeed(max-min) + min
}

var source mathrand.Source

// must pass a pointer to the data
func populateWithFakeData(ctx context.Context, a interface{}) error {
	seed, ok := ctx.Value(t.CtxMockSeedKey).(int64)
	if !ok {
		seed = time.Now().UnixNano()
	}
	source = faker.NewSafeSource(mathrand.NewSource(seed))
	faker.SetRandomSource(source)
	return faker.FakeData(a, options.WithRandomMapAndSliceMaxSize(10), options.WithRandomFloatBoundaries(interfaces.RandomFloatBoundary{Start: 0, End: 1}))
}

func (*DummyService) StartDataAccessServices() {
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
	r := struct {
		Data   []T
		Paging t.Paging
	}{}
	err := populateWithFakeData(ctx, &r)
	return r.Data, &r.Paging, err
}

func (*DummyService) Close() {
	// nothing to close
}

func (*DummyService) GetLatestSlot(ctx context.Context) (uint64, error) {
	return getDummyData[uint64](ctx)
}

func (*DummyService) GetLatestFinalizedEpoch(ctx context.Context) (uint64, error) {
	return getDummyData[uint64](ctx)
}

func (*DummyService) GetLatestBlock(ctx context.Context) (uint64, error) {
	return getDummyData[uint64](ctx)
}

func (*DummyService) GetLatestExchangeRates(ctx context.Context) ([]t.EthConversionRate, error) {
	return getDummyData[[]t.EthConversionRate](ctx)
}

func (*DummyService) GetUserByEmail(ctx context.Context, email string) (uint64, error) {
	return getDummyData[uint64](ctx)
}

func (*DummyService) CreateUser(ctx context.Context, email, password string) (uint64, error) {
	return getDummyData[uint64](ctx)
}

func (*DummyService) RemoveUser(ctx context.Context, userId uint64) error {
	return nil
}

func (*DummyService) UpdateUserEmail(ctx context.Context, userId uint64) error {
	return nil
}

func (*DummyService) UpdateUserPassword(ctx context.Context, userId uint64, password string) error {
	return nil
}

func (*DummyService) GetEmailConfirmationTime(ctx context.Context, userId uint64) (time.Time, error) {
	return getDummyData[time.Time](ctx)
}

func (*DummyService) GetPasswordResetTime(ctx context.Context, userId uint64) (time.Time, error) {
	return getDummyData[time.Time](ctx)
}

func (*DummyService) UpdateEmailConfirmationTime(ctx context.Context, userId uint64) error {
	return nil
}

func (*DummyService) IsPasswordResetAllowed(ctx context.Context, userId uint64) (bool, error) {
	return true, nil
}

func (*DummyService) UpdatePasswordResetTime(ctx context.Context, userId uint64) error {
	return nil
}

func (*DummyService) UpdateEmailConfirmationHash(ctx context.Context, userId uint64, email, confirmationHash string) error {
	return nil
}

func (*DummyService) UpdatePasswordResetHash(ctx context.Context, userId uint64, confirmationHash string) error {
	return nil
}

func (*DummyService) GetUserInfo(ctx context.Context, userId uint64) (*t.UserInfo, error) {
	return getDummyStruct[t.UserInfo](ctx)
}

func (*DummyService) GetUserCredentialInfo(ctx context.Context, userId uint64) (*t.UserCredentialInfo, error) {
	return getDummyStruct[t.UserCredentialInfo](ctx)
}

func (*DummyService) GetUserIdByApiKey(ctx context.Context, apiKey string) (uint64, error) {
	return getDummyData[uint64](ctx)
}

func (*DummyService) GetUserIdByConfirmationHash(ctx context.Context, hash string) (uint64, error) {
	return getDummyData[uint64](ctx)
}

func (*DummyService) GetUserIdByResetHash(ctx context.Context, hash string) (uint64, error) {
	return getDummyData[uint64](ctx)
}

func (*DummyService) GetProductSummary(ctx context.Context) (*t.ProductSummary, error) {
	return getDummyStruct[t.ProductSummary](ctx)
}

func (*DummyService) GetFreeTierPerks(ctx context.Context) (*t.PremiumPerks, error) {
	return getDummyStruct[t.PremiumPerks](ctx)
}

func (*DummyService) GetValidatorDashboardUser(ctx context.Context, dashboardId t.VDBIdPrimary) (*t.DashboardUser, error) {
	return getDummyStruct[t.DashboardUser](ctx)
}

func (*DummyService) GetValidatorDashboardIdByPublicId(ctx context.Context, publicDashboardId t.VDBIdPublic) (*t.VDBIdPrimary, error) {
	return getDummyStruct[t.VDBIdPrimary](ctx)
}

func (*DummyService) GetValidatorDashboardInfo(ctx context.Context, dashboardId t.VDBIdPrimary) (*t.ValidatorDashboard, error) {
	r, err := getDummyStruct[t.ValidatorDashboard](ctx)
	// return semi-valid data to not break staging
	r.IsArchived = false
	return r, err
}

func (*DummyService) GetValidatorDashboardName(ctx context.Context, dashboardId t.VDBIdPrimary) (string, error) {
	return getDummyData[string](ctx)
}

func (*DummyService) GetValidatorsFromSlices(ctx context.Context, indices []uint64, publicKeys []string) ([]t.VDBValidator, error) {
	return getDummyData[[]t.VDBValidator](ctx)
}

func (*DummyService) GetValidatorsEffectiveBalanceTotal(ctx context.Context, indices []uint64) (uint64, error) {
	return getDummyData[uint64](ctx)
}

func (*DummyService) GetUserDashboards(ctx context.Context, userId uint64) (*t.UserDashboardsData, error) {
	return getDummyStruct[t.UserDashboardsData](ctx)
}

func (*DummyService) CreateValidatorDashboard(ctx context.Context, userId uint64, name string, network uint64) (*t.VDBPostReturnData, error) {
	return getDummyStruct[t.VDBPostReturnData](ctx)
}

func (*DummyService) GetValidatorDashboardOverview(ctx context.Context, dashboardId t.VDBId, protocolModes t.VDBProtocolModes) (*t.VDBOverviewData, error) {
	return getDummyStruct[t.VDBOverviewData](ctx)
}

func (*DummyService) RemoveValidatorDashboard(ctx context.Context, dashboardId t.VDBIdPrimary) error {
	return nil
}

func (*DummyService) RemoveValidatorDashboards(ctx context.Context, dashboardIds []uint64) error {
	return nil
}

func (*DummyService) UpdateValidatorDashboardArchiving(ctx context.Context, dashboardId t.VDBIdPrimary, archivedReason *enums.VDBArchivedReason) (*t.VDBPostArchivingReturnData, error) {
	return getDummyStruct[t.VDBPostArchivingReturnData](ctx)
}

func (*DummyService) UpdateValidatorDashboardsArchiving(ctx context.Context, dashboards []t.ArchiverDashboardArchiveReason) error {
	return nil
}

func (*DummyService) UpdateValidatorDashboardName(ctx context.Context, dashboardId t.VDBIdPrimary, name string) (*t.VDBPostReturnData, error) {
	return getDummyStruct[t.VDBPostReturnData](ctx)
}

func (*DummyService) CreateValidatorDashboardGroup(ctx context.Context, dashboardId t.VDBIdPrimary, name string) (*t.VDBPostCreateGroupData, error) {
	return getDummyStruct[t.VDBPostCreateGroupData](ctx)
}

func (*DummyService) UpdateValidatorDashboardGroup(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, name string) (*t.VDBPostCreateGroupData, error) {
	return getDummyStruct[t.VDBPostCreateGroupData](ctx)
}

func (*DummyService) RemoveValidatorDashboardGroup(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64) error {
	return nil
}

func (*DummyService) RemoveValidatorDashboardGroupValidators(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64) error {
	return nil
}

func (*DummyService) GetValidatorDashboardGroupExists(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64) (bool, error) {
	return true, nil
}

func (*DummyService) AddValidatorDashboardValidators(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, validators []t.VDBValidator) ([]t.VDBPostValidatorsData, error) {
	return getDummyData[[]t.VDBPostValidatorsData](ctx)
}

func (*DummyService) GetValidatorsByDepositAddress(ctx context.Context, depositAddress string) ([]t.VDBValidator, error) {
	return getDummyData[[]t.VDBValidator](ctx)
}

func (*DummyService) GetValidatorsByWithdrawalCredentials(ctx context.Context, withdrawalCredentials string) ([]t.VDBValidator, error) {
	return getDummyData[[]t.VDBValidator](ctx)
}

func (*DummyService) GetValidatorsByGraffiti(ctx context.Context, graffiti string) ([]t.VDBValidator, error) {
	return getDummyData[[]t.VDBValidator](ctx)
}

func (*DummyService) GetValidatorDashboardValidators(ctx context.Context, dashboardId t.VDBId, groupId int64, cursor string, colSort t.Sort[enums.VDBManageValidatorsColumn], search string, limit uint64) ([]t.VDBManageValidatorsTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.VDBManageValidatorsTableRow](ctx)
}

func (*DummyService) RemoveValidatorDashboardValidators(ctx context.Context, dashboardId t.VDBIdPrimary, validators []t.VDBValidator) error {
	return nil
}

func (*DummyService) CreateValidatorDashboardPublicId(ctx context.Context, dashboardId t.VDBIdPrimary, name string, shareGroups bool) (*t.VDBPublicId, error) {
	return getDummyStruct[t.VDBPublicId](ctx)
}

func (*DummyService) GetValidatorDashboardPublicId(ctx context.Context, publicDashboardId t.VDBIdPublic) (*t.VDBPublicId, error) {
	return getDummyStruct[t.VDBPublicId](ctx)
}

func (*DummyService) UpdateValidatorDashboardPublicId(ctx context.Context, publicDashboardId t.VDBIdPublic, name string, shareGroups bool) (*t.VDBPublicId, error) {
	return getDummyStruct[t.VDBPublicId](ctx)
}

func (*DummyService) RemoveValidatorDashboardPublicId(ctx context.Context, publicDashboardId t.VDBIdPublic) error {
	return nil
}

func (*DummyService) GetValidatorDashboardSlotViz(ctx context.Context, dashboardId t.VDBId, groupIds []uint64) ([]t.SlotVizEpoch, error) {
	r := struct {
		Epochs []t.SlotVizEpoch `faker:"slice_len=4"`
	}{}
	err := populateWithFakeData(ctx, &r)
	return r.Epochs, err
}

func (*DummyService) GetValidatorDashboardSummary(ctx context.Context, dashboardId t.VDBId, period enums.TimePeriod, cursor string, colSort t.Sort[enums.VDBSummaryColumn], search string, limit uint64, protocolModes t.VDBProtocolModes) ([]t.VDBSummaryTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.VDBSummaryTableRow](ctx)
}
func (*DummyService) GetValidatorDashboardGroupSummary(ctx context.Context, dashboardId t.VDBId, groupId int64, period enums.TimePeriod, protocolModes t.VDBProtocolModes) (*t.VDBGroupSummaryData, error) {
	return getDummyStruct[t.VDBGroupSummaryData](ctx)
}

func (*DummyService) GetValidatorDashboardSummaryChart(ctx context.Context, dashboardId t.VDBId, groupIds []int64, efficiency enums.VDBSummaryChartEfficiencyType, aggregation enums.ChartAggregation, afterTs uint64, beforeTs uint64) (*t.ChartData[int, float64], error) {
	return getDummyStruct[t.ChartData[int, float64]](ctx)
}

func (*DummyService) GetValidatorDashboardSummaryValidators(ctx context.Context, dashboardId t.VDBId, groupId int64) (*t.VDBGeneralSummaryValidators, error) {
	return getDummyStruct[t.VDBGeneralSummaryValidators](ctx)
}
func (*DummyService) GetValidatorDashboardSyncSummaryValidators(ctx context.Context, dashboardId t.VDBId, groupId int64, period enums.TimePeriod) (*t.VDBSyncSummaryValidators, error) {
	return getDummyStruct[t.VDBSyncSummaryValidators](ctx)
}
func (*DummyService) GetValidatorDashboardSlashingsSummaryValidators(ctx context.Context, dashboardId t.VDBId, groupId int64, period enums.TimePeriod) (*t.VDBSlashingsSummaryValidators, error) {
	return getDummyStruct[t.VDBSlashingsSummaryValidators](ctx)
}
func (*DummyService) GetValidatorDashboardProposalSummaryValidators(ctx context.Context, dashboardId t.VDBId, groupId int64, period enums.TimePeriod) (*t.VDBProposalSummaryValidators, error) {
	return getDummyStruct[t.VDBProposalSummaryValidators](ctx)
}

func (*DummyService) GetValidatorDashboardRewards(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBRewardsColumn], search string, limit uint64, protocolModes t.VDBProtocolModes) ([]t.VDBRewardsTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.VDBRewardsTableRow](ctx)
}

func (*DummyService) GetValidatorDashboardGroupRewards(ctx context.Context, dashboardId t.VDBId, groupId int64, epoch uint64, protocolModes t.VDBProtocolModes) (*t.VDBGroupRewardsData, error) {
	return getDummyStruct[t.VDBGroupRewardsData](ctx)
}

func (*DummyService) GetValidatorDashboardRewardsChart(ctx context.Context, dashboardId t.VDBId, protocolModes t.VDBProtocolModes) (*t.ChartData[int, decimal.Decimal], error) {
	return getDummyStruct[t.ChartData[int, decimal.Decimal]](ctx)
}

func (*DummyService) GetValidatorDashboardDuties(ctx context.Context, dashboardId t.VDBId, epoch uint64, groupId int64, cursor string, colSort t.Sort[enums.VDBDutiesColumn], search string, limit uint64, protocolModes t.VDBProtocolModes) ([]t.VDBEpochDutiesTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.VDBEpochDutiesTableRow](ctx)
}

func (*DummyService) GetValidatorDashboardBlocks(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBBlocksColumn], search string, limit uint64, protocolModes t.VDBProtocolModes) ([]t.VDBBlocksTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.VDBBlocksTableRow](ctx)
}

func (*DummyService) GetValidatorDashboardHeatmap(ctx context.Context, dashboardId t.VDBId, protocolModes t.VDBProtocolModes, aggregation enums.ChartAggregation, afterTs uint64, beforeTs uint64) (*t.VDBHeatmap, error) {
	return getDummyStruct[t.VDBHeatmap](ctx)
}

func (*DummyService) GetValidatorDashboardGroupHeatmap(ctx context.Context, dashboardId t.VDBId, groupId uint64, protocolModes t.VDBProtocolModes, aggregation enums.ChartAggregation, timestamp uint64) (*t.VDBHeatmapTooltipData, error) {
	return getDummyStruct[t.VDBHeatmapTooltipData](ctx)
}

func (*DummyService) GetValidatorDashboardElDeposits(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBDepositsElColumn], search string, limit uint64) ([]t.VDBExecutionDepositsTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.VDBExecutionDepositsTableRow](ctx)
}

func (*DummyService) GetValidatorDashboardClDeposits(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBDepositsClColumn], search string, limit uint64) ([]t.VDBConsensusDepositsTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.VDBConsensusDepositsTableRow](ctx)
}

func (*DummyService) GetValidatorDashboardTotalElDeposits(ctx context.Context, dashboardId t.VDBId, search string) (*t.VDBTotalExecutionDepositsData, error) {
	return getDummyStruct[t.VDBTotalExecutionDepositsData](ctx)
}

func (*DummyService) GetValidatorDashboardTotalClDeposits(ctx context.Context, dashboardId t.VDBId, search string) (*t.VDBTotalConsensusDepositsData, error) {
	return getDummyStruct[t.VDBTotalConsensusDepositsData](ctx)
}

func (*DummyService) GetValidatorDashboardElWithdrawals(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBWithdrawalsElColumn], search string, limit uint64, protocolModes t.VDBProtocolModes) ([]t.VDBWithdrawalsElTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.VDBWithdrawalsElTableRow](ctx)
}

func (*DummyService) GetValidatorDashboardClWithdrawals(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBWithdrawalsClColumn], search string, limit uint64, protocolModes t.VDBProtocolModes) ([]t.VDBWithdrawalsClTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.VDBWithdrawalsClTableRow](ctx)
}

func (*DummyService) GetValidatorDashboardTotalElWithdrawals(ctx context.Context, dashboardId t.VDBId, search string, protocolModes t.VDBProtocolModes) (*t.VDBTotalExecutionWithdrawalsData, error) {
	return getDummyStruct[t.VDBTotalExecutionWithdrawalsData](ctx)
}

func (*DummyService) GetValidatorDashboardTotalClWithdrawals(ctx context.Context, dashboardId t.VDBId, search string, protocolModes t.VDBProtocolModes) (*t.VDBTotalConsensusWithdrawalsData, error) {
	return getDummyStruct[t.VDBTotalConsensusWithdrawalsData](ctx)
}

func (*DummyService) GetValidatorDashboardExecutionLayerConsolidations(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBConsolidationsElColumn], search string, limit uint64) ([]t.VDBConsolidationsElTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.VDBConsolidationsElTableRow](ctx)
}

func (*DummyService) GetValidatorDashboardConsensusLayerConsolidations(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBConsolidationsClColumn], search string, limit uint64) ([]t.VDBConsolidationsClTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.VDBConsolidationsClTableRow](ctx)
}

func (*DummyService) GetValidatorDashboardRocketPool(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBRocketPoolColumn], search string, limit uint64) ([]t.VDBRocketPoolTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.VDBRocketPoolTableRow](ctx)
}

func (*DummyService) GetValidatorDashboardTotalRocketPool(ctx context.Context, dashboardId t.VDBId, search string) (*t.VDBRocketPoolTableRow, error) {
	return getDummyStruct[t.VDBRocketPoolTableRow](ctx)
}

func (*DummyService) GetValidatorDashboardRocketPoolMinipools(ctx context.Context, dashboardId t.VDBId, node string, cursor string, colSort t.Sort[enums.VDBRocketPoolMinipoolsColumn], search string, limit uint64) ([]t.VDBRocketPoolMinipoolsTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.VDBRocketPoolMinipoolsTableRow](ctx)
}

func (*DummyService) GetAllNetworks() ([]t.NetworkInfo, error) {
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
		{
			ChainId:           560048,
			Name:              "hoodi",
			NotificationsName: "hoodi",
		},
	}, nil
}

func (*DummyService) GetAllClients() ([]t.ClientInfo, error) {
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

func (*DummyService) GetSearchValidatorByIndex(ctx context.Context, chainId, index uint64) (*t.SearchValidator, error) {
	return getDummyStruct[t.SearchValidator](ctx)
}

func (*DummyService) GetSearchValidatorByPublicKey(ctx context.Context, chainId uint64, publicKey []byte) (*t.SearchValidator, error) {
	return getDummyStruct[t.SearchValidator](ctx)
}

func (*DummyService) GetSearchValidatorsByDepositAddress(ctx context.Context, chainId uint64, address []byte) (*t.SearchValidatorsByDepositAddress, error) {
	return getDummyStruct[t.SearchValidatorsByDepositAddress](ctx)
}

func (*DummyService) GetSearchValidatorsByDepositEnsName(ctx context.Context, chainId uint64, ensName string) (*t.SearchValidatorsByDepositAddress, error) {
	return getDummyStruct[t.SearchValidatorsByDepositAddress](ctx)
}

func (*DummyService) GetSearchValidatorsByWithdrawalCredential(ctx context.Context, chainId uint64, credential []byte) (*t.SearchValidatorsByWithdrawalCredential, error) {
	return getDummyStruct[t.SearchValidatorsByWithdrawalCredential](ctx)
}

func (*DummyService) GetSearchValidatorsByWithdrawalEnsName(ctx context.Context, chainId uint64, ensName string) (*t.SearchValidatorsByWithdrawalCredential, error) {
	return getDummyStruct[t.SearchValidatorsByWithdrawalCredential](ctx)
}

func (*DummyService) GetSearchValidatorsByGraffiti(ctx context.Context, chainId uint64, graffiti string) (*t.SearchValidatorsByGraffiti, error) {
	return getDummyStruct[t.SearchValidatorsByGraffiti](ctx)
}

func (*DummyService) GetSearchValidatorsByGraffitiHex(ctx context.Context, chainId uint64, graffiti []byte) (*t.SearchValidatorsByGraffiti, error) {
	return getDummyStruct[t.SearchValidatorsByGraffiti](ctx)
}

func (*DummyService) GetUserValidatorDashboardCount(ctx context.Context, userId uint64, active bool) (uint64, error) {
	return getDummyData[uint64](ctx)
}

func (*DummyService) GetValidatorDashboardGroupCount(ctx context.Context, dashboardId t.VDBIdPrimary) (uint64, error) {
	return getDummyData[uint64](ctx)
}

func (*DummyService) GetValidatorDashboardEffectiveBalanceTotal(ctx context.Context, dashboardId t.VDBId, onlyActive bool) (uint64, error) {
	return getDummyData[uint64](ctx)
}

func (*DummyService) GetValidatorsEffectiveBalances(ctx context.Context, validators []t.VDBValidator, onlyActive bool) (map[t.VDBValidator]uint64, error) {
	return map[t.VDBValidator]uint64{}, nil
}

func (*DummyService) GetValidatorDashboardPublicIdCount(ctx context.Context, dashboardId t.VDBIdPrimary) (uint64, error) {
	return getDummyData[uint64](ctx)
}

func (*DummyService) GetNotificationOverview(ctx context.Context, userId uint64) (*t.NotificationOverviewData, error) {
	return getDummyStruct[t.NotificationOverviewData](ctx)
}
func (*DummyService) GetDashboardNotifications(ctx context.Context, userId uint64, chainIds []uint64, cursor string, colSort t.Sort[enums.NotificationDashboardsColumn], search string, limit uint64) ([]t.NotificationDashboardsTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.NotificationDashboardsTableRow](ctx)
}

func (*DummyService) GetValidatorDashboardNotificationDetails(ctx context.Context, dashboardId t.VDBIdPrimary, groupId uint64, epoch uint64, search string) (*t.NotificationValidatorDashboardDetail, error) {
	return getDummyStruct[t.NotificationValidatorDashboardDetail](ctx)
}

func (*DummyService) GetAccountDashboardNotificationDetails(ctx context.Context, dashboardId uint64, groupId uint64, epoch uint64, search string) (*t.NotificationAccountDashboardDetail, error) {
	return getDummyStruct[t.NotificationAccountDashboardDetail](ctx)
}

func (*DummyService) GetMachineNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationMachinesColumn], search string, limit uint64) ([]t.NotificationMachinesTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.NotificationMachinesTableRow](ctx)
}
func (*DummyService) GetClientNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationClientsColumn], search string, limit uint64) ([]t.NotificationClientsTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.NotificationClientsTableRow](ctx)
}
func (*DummyService) GetNetworkNotifications(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationNetworksColumn], limit uint64) ([]t.NotificationNetworksTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.NotificationNetworksTableRow](ctx)
}

func (*DummyService) GetNotificationSettings(ctx context.Context, userId uint64) (*t.NotificationSettings, error) {
	return getDummyStruct[t.NotificationSettings](ctx)
}
func (*DummyService) GetNotificationSettingsDefaultValues(ctx context.Context) (*t.NotificationSettingsDefaultValues, error) {
	return getDummyStruct[t.NotificationSettingsDefaultValues](ctx)
}
func (*DummyService) UpdateNotificationSettingsGeneral(ctx context.Context, userId uint64, settings t.NotificationSettingsGeneral) error {
	return nil
}
func (*DummyService) UpdateNotificationSettingsNetworks(ctx context.Context, userId uint64, chainId uint64, settings t.NotificationSettingsNetwork) error {
	return nil
}
func (*DummyService) UpdateNotificationSettingsPairedDevice(ctx context.Context, pairedDeviceId uint64, name string, IsNotificationsEnabled bool) error {
	return nil
}
func (*DummyService) DeleteNotificationSettingsPairedDevice(ctx context.Context, pairedDeviceId uint64) error {
	return nil
}

func (*DummyService) UpdateNotificationSettingsClients(ctx context.Context, userId uint64, clientId uint64, IsSubscribed bool) (*t.NotificationSettingsClient, error) {
	return getDummyStruct[t.NotificationSettingsClient](ctx)
}

func (*DummyService) GetNotificationSettingsDashboards(ctx context.Context, userId uint64, cursor string, colSort t.Sort[enums.NotificationSettingsDashboardColumn], search string, limit uint64) ([]t.NotificationSettingsDashboardsTableRow, *t.Paging, error) {
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
func (*DummyService) UpdateNotificationSettingsValidatorDashboard(ctx context.Context, userId uint64, dashboardId t.VDBIdPrimary, groupId uint64, settings t.NotificationSettingsValidatorDashboard) error {
	return nil
}
func (*DummyService) UpdateNotificationSettingsAccountDashboard(ctx context.Context, userId uint64, dashboardId t.VDBIdPrimary, groupId uint64, settings t.NotificationSettingsAccountDashboard) error {
	return nil
}
func (*DummyService) CreateAdConfiguration(ctx context.Context, key, jquerySelector string, insertMode enums.AdInsertMode, refreshInterval uint64, forAllUsers bool, bannerId uint64, htmlContent string, enabled bool) error {
	return nil
}

func (*DummyService) GetAdConfigurations(ctx context.Context, keys []string) ([]t.AdConfigurationData, error) {
	return getDummyData[[]t.AdConfigurationData](ctx)
}

func (*DummyService) UpdateAdConfiguration(ctx context.Context, key, jquerySelector string, insertMode enums.AdInsertMode, refreshInterval uint64, forAllUsers bool, bannerId uint64, htmlContent string, enabled bool) error {
	return nil
}

func (*DummyService) RemoveAdConfiguration(ctx context.Context, key string) error {
	return nil
}

func (*DummyService) GetLatestExportedChartTs(ctx context.Context, aggregation enums.ChartAggregation) (uint64, error) {
	return getDummyData[uint64](ctx)
}

func (*DummyService) GetUserIdByRefreshToken(ctx context.Context, claimUserID, claimAppID, claimDeviceID uint64, hashedRefreshToken string) (uint64, error) {
	return getDummyData[uint64](ctx)
}

func (*DummyService) MigrateMobileSession(ctx context.Context, oldHashedRefreshToken, newHashedRefreshToken, deviceID, deviceName string) error {
	return nil
}

func (*DummyService) GetAppDataFromRedirectUri(ctx context.Context, callback string) (*t.OAuthAppData, error) {
	return getDummyStruct[t.OAuthAppData](ctx)
}

func (*DummyService) AddUserDevice(ctx context.Context, userID uint64, hashedRefreshToken string, deviceID, deviceName string, appID uint64) error {
	return nil
}

func (*DummyService) AddMobileNotificationToken(ctx context.Context, userID uint64, deviceID, notifyToken string) error {
	return nil
}

func (*DummyService) GetAppSubscriptionCount(ctx context.Context, userID uint64) (uint64, error) {
	return getDummyData[uint64](ctx)
}

func (*DummyService) AddMobilePurchase(ctx context.Context, tx *sql.Tx, userID uint64, paymentDetails t.MobileSubscription, verifyResponse *userservice.VerifyResponse, extSubscriptionId string) error {
	return nil
}

func (*DummyService) GetBlockOverview(ctx context.Context, chainId, block uint64) (*t.BlockOverview, error) {
	return getDummyStruct[t.BlockOverview](ctx)
}

func (*DummyService) GetBlockTransactions(ctx context.Context, chainId, block uint64) ([]t.BlockTransactionTableRow, error) {
	return getDummyData[[]t.BlockTransactionTableRow](ctx)
}

func (*DummyService) GetBlock(ctx context.Context, chainId, block uint64) (*t.BlockSummary, error) {
	return getDummyStruct[t.BlockSummary](ctx)
}

func (*DummyService) GetBlockVotes(ctx context.Context, chainId, block uint64) ([]t.BlockVoteTableRow, error) {
	return getDummyData[[]t.BlockVoteTableRow](ctx)
}

func (*DummyService) GetBlockAttestations(ctx context.Context, chainId, block uint64) ([]t.BlockAttestationTableRow, error) {
	return getDummyData[[]t.BlockAttestationTableRow](ctx)
}

func (*DummyService) GetBlockWithdrawals(ctx context.Context, chainId, block uint64) ([]t.BlockWithdrawalTableRow, error) {
	return getDummyData[[]t.BlockWithdrawalTableRow](ctx)
}

func (*DummyService) GetBlockBlsChanges(ctx context.Context, chainId, block uint64) ([]t.BlockBlsChangeTableRow, error) {
	return getDummyData[[]t.BlockBlsChangeTableRow](ctx)
}

func (*DummyService) GetBlockVoluntaryExits(ctx context.Context, chainId, block uint64) ([]t.BlockVoluntaryExitTableRow, error) {
	return getDummyData[[]t.BlockVoluntaryExitTableRow](ctx)
}

func (*DummyService) GetBlockBlobs(ctx context.Context, chainId, block uint64) ([]t.BlockBlobTableRow, error) {
	return getDummyData[[]t.BlockBlobTableRow](ctx)
}

func (*DummyService) GetSlot(ctx context.Context, chainId, block uint64) (*t.BlockSummary, error) {
	return getDummyStruct[t.BlockSummary](ctx)
}

func (*DummyService) GetSlotOverview(ctx context.Context, chainId, block uint64) (*t.BlockOverview, error) {
	return getDummyStruct[t.BlockOverview](ctx)
}

func (*DummyService) GetSlotTransactions(ctx context.Context, chainId, block uint64) ([]t.BlockTransactionTableRow, error) {
	return getDummyData[[]t.BlockTransactionTableRow](ctx)
}

func (*DummyService) GetSlotVotes(ctx context.Context, chainId, block uint64) ([]t.BlockVoteTableRow, error) {
	return getDummyData[[]t.BlockVoteTableRow](ctx)
}

func (*DummyService) GetSlotAttestations(ctx context.Context, chainId, block uint64) ([]t.BlockAttestationTableRow, error) {
	return getDummyData[[]t.BlockAttestationTableRow](ctx)
}

func (*DummyService) GetSlotWithdrawals(ctx context.Context, chainId, block uint64) ([]t.BlockWithdrawalTableRow, error) {
	return getDummyData[[]t.BlockWithdrawalTableRow](ctx)
}

func (*DummyService) GetSlotBlsChanges(ctx context.Context, chainId, block uint64) ([]t.BlockBlsChangeTableRow, error) {
	return getDummyData[[]t.BlockBlsChangeTableRow](ctx)
}

func (*DummyService) GetSlotVoluntaryExits(ctx context.Context, chainId, block uint64) ([]t.BlockVoluntaryExitTableRow, error) {
	return getDummyData[[]t.BlockVoluntaryExitTableRow](ctx)
}

func (*DummyService) GetSlotBlobs(ctx context.Context, chainId, block uint64) ([]t.BlockBlobTableRow, error) {
	return getDummyData[[]t.BlockBlobTableRow](ctx)
}

func (*DummyService) GetValidatorDashboardsCountInfo(ctx context.Context) (map[uint64][]t.ArchiverDashboard, error) {
	return getDummyData[map[uint64][]t.ArchiverDashboard](ctx)
}

func (*DummyService) GetRocketPoolOverview(ctx context.Context) (*t.RocketPoolData, error) {
	return getDummyStruct[t.RocketPoolData](ctx)
}

func (*DummyService) GetApiWeights(ctx context.Context) ([]t.ApiWeightItem, error) {
	return getDummyData[[]t.ApiWeightItem](ctx)
}

func (*DummyService) GetHealthz(ctx context.Context, showAll bool) t.HealthzData {
	r, _ := getDummyData[t.HealthzData](ctx)
	return r
}

func (*DummyService) GetLatestBundleForNativeVersion(ctx context.Context, nativeVersion uint64) (*t.MobileAppBundleStats, error) {
	return getDummyStruct[t.MobileAppBundleStats](ctx)
}

func (*DummyService) IncrementBundleDeliveryCount(ctx context.Context, bundleVerison uint64) error {
	return nil
}

func (*DummyService) GetValidatorDashboardMobileWidget(ctx context.Context, dashboardId t.VDBIdPrimary) (*t.MobileWidgetData, error) {
	return getDummyStruct[t.MobileWidgetData](ctx)
}

func (*DummyService) GetUserMachineMetrics(ctx context.Context, userID uint64, limit int, offset int) (*t.MachineMetricsData, error) {
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

func (*DummyService) PostUserMachineMetrics(ctx context.Context, userID uint64, machine, process string, data []byte) error {
	return nil
}

func (*DummyService) GetValidatorDashboardMobileValidators(ctx context.Context, dashboardId t.VDBId, groupId int64, period enums.TimePeriod, cursor string, colSort t.Sort[enums.VDBManageValidatorsColumn], search string, limit uint64) ([]t.MobileValidatorDashboardValidatorsTableRow, *t.Paging, error) {
	return getDummyWithPaging[t.MobileValidatorDashboardValidatorsTableRow](ctx)
}

func (*DummyService) QueueTestEmailNotification(ctx context.Context, userId uint64) error {
	return nil
}
func (*DummyService) QueueTestPushNotification(ctx context.Context, userId uint64) error {
	return nil
}
func (*DummyService) QueueTestWebhookNotification(ctx context.Context, userId uint64, webhookUrl string, isDiscordWebhook bool) error {
	return nil
}

func (*DummyService) GetPairedDeviceUserId(ctx context.Context, pairedDeviceId uint64) (uint64, error) {
	return getDummyData[uint64](ctx)
}

func (*DummyService) GetHasUserActiveSubscription(ctx context.Context, userId uint64) (bool, error) {
	return getDummyData[bool](ctx)
}

func (*DummyService) GetValidatorDashboardValidatorsOfList(ctx context.Context, dashboardId t.VDBIdPrimary, validators []t.VDBValidator) ([]t.VDBValidator, error) {
	return getDummyData[[]t.VDBValidator](ctx)
}
