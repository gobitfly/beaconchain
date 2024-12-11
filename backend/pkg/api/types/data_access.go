package types

import (
	"database/sql"
	"time"

	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/consapi/types"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
	"github.com/shopspring/decimal"
)

// everything that goes in this file is for the data access layer only
// it won't be converted to typescript or used in the frontend

const DefaultGroupId = 0
const AllGroups = -1
const NetworkAverage = -2
const DefaultGroupName = "default"
const DefaultDashboardName = DefaultGroupName

type Sort[T enums.Enum] struct {
	Column T
	Desc   bool
}

type SortColumn struct {
	// defaults
	Column enums.OrderableSortable
	Desc   bool
	Offset any // nil to indicate null value
}

type VDBIdPrimary int
type VDBIdPublic string
type VDBIdValidatorSet []VDBValidator
type VDBId struct {
	Validators      VDBIdValidatorSet // if this is nil, then use the id
	Id              VDBIdPrimary
	AggregateGroups bool
}

// could replace if we want the import in all files
type VDBValidator = types.ValidatorIndex

type DashboardUser struct {
	Id     VDBIdPrimary `db:"id"` // this must be the bigint id
	UserId uint64       `db:"user_id"`
}

type CursorLike interface {
	IsCursor() bool
	IsValid() bool
	IsReverse() bool
}

type GenericCursor struct {
	Reverse bool `json:"r"`
	Valid   bool `json:"-"`
}

func (c GenericCursor) IsCursor() bool {
	return true
}

func (c GenericCursor) IsValid() bool {
	return c.Valid
}

// note: dont have to check for valid when calling this
func (c GenericCursor) IsReverse() bool {
	return c.Reverse && c.Valid
}

type CLDepositsCursor struct {
	GenericCursor
	Slot      int64
	SlotIndex int64
}

type ELDepositsCursor struct {
	GenericCursor
	BlockNumber int64
	LogIndex    int64
}

type ValidatorsCursor struct {
	GenericCursor

	Index uint64 `json:"vi"`
}

type RewardsCursor struct {
	GenericCursor

	Epoch   uint64
	GroupId int64
}

type ValidatorDutiesCursor struct {
	GenericCursor

	Index  uint64
	Reward decimal.Decimal
}

type WithdrawalsCursor struct {
	GenericCursor

	Slot            uint64
	WithdrawalIndex uint64
	Index           uint64
	Recipient       []byte
	Amount          uint64
}

type NotificationSettingsCursor struct {
	GenericCursor

	IsAccountDashboard bool // if false it's a validator dashboard
	DashboardId        uint64
	GroupId            uint64
}

type NotificationMachinesCursor struct {
	GenericCursor

	MachineId      uint64
	MachineName    string
	EventType      string
	EventThreshold float64
	Ts             time.Time
}

type NotificationClientsCursor struct {
	GenericCursor

	Client string
	Ts     time.Time
}

type NotificationNetworksCursor struct {
	GenericCursor

	Network   uint64
	Ts        time.Time
	EventType t.EventName
}

type UserCredentialInfo struct {
	Id             uint64 `db:"id"`
	Email          string `db:"email"`
	EmailConfirmed bool   `db:"email_confirmed"`
	Password       string `db:"password"`
	ProductId      string `db:"product_id"`
	UserGroup      string `db:"user_group"`
}

type BlocksCursor struct {
	GenericCursor

	Proposer uint64
	Slot     uint64 // same as Age
	Block    sql.NullInt64
	Status   uint64
	Reward   decimal.Decimal
}

type NotificationsDashboardsCursor struct {
	GenericCursor

	Epoch         uint64
	ChainId       uint64
	DashboardName string
	DashboardId   uint64
	GroupName     string
	GroupId       uint64
}

type NetworkInfo struct {
	ChainId           uint64
	Name              string
	NotificationsName string
}

type ClientInfo struct {
	Id       uint64
	Name     string
	DbName   string
	Category string
}

// -------------------------
// validator indices structs, only used between data access and api layer

type VDBGeneralSummaryValidators struct {
	// fill slices with indices of validators
	Deposited   []uint64
	Pending     []IndexTimestamp
	Online      []uint64
	Offline     []uint64
	Slashing    []uint64
	Exiting     []IndexTimestamp
	Slashed     []uint64
	Exited      []uint64
	Withdrawing []IndexTimestamp
	Withdrawn   []uint64
}

type IndexTimestamp struct {
	Index     uint64
	Timestamp uint64
}

type VDBValidatorSyncPast struct {
	Index uint64
	Count uint64
}
type VDBSyncSummaryValidators struct {
	// fill slices with indices of validators
	Upcoming []uint64
	Current  []uint64
	Past     []VDBValidatorSyncPast
}

type VDBValidatorGotSlashed struct {
	Index     uint64
	SlashedBy uint64
}
type VDBValidatorHasSlashed struct {
	Index          uint64
	SlashedIndices []uint64
}
type VDBSlashingsSummaryValidators struct {
	// fill with the validator index that got slashed and the index of the validator that slashed it
	GotSlashed []VDBValidatorGotSlashed
	// fill with the validator index that slashed and the index of the validators that got slashed
	HasSlashed []VDBValidatorHasSlashed
}

type VDBProposalSummaryValidators struct {
	Proposed []IndexSlots
	Missed   []IndexSlots
}

type VDBProtocolModes struct {
	RocketPool bool
}

type MobileSubscription struct {
	ProductIDUnverified string                               `json:"id"`
	PriceMicros         uint64                               `json:"priceMicros"`
	Currency            string                               `json:"currency"`
	Transaction         MobileSubscriptionTransactionGeneric `json:"transaction"`
	ValidUnverified     bool                                 `json:"valid"`
}

type MobileSubscriptionTransactionGeneric struct {
	Type    string `json:"type"`
	Receipt string `json:"receipt"`
	ID      string `json:"id"`
}

type VDBValidatorSummaryChartRow struct {
	Timestamp              time.Time `db:"ts"`
	GroupId                int64     `db:"group_id"`
	AttestationReward      float64   `db:"attestations_reward"`
	AttestationIdealReward float64   `db:"attestations_ideal_reward"`
	BlocksProposed         float64   `db:"blocks_proposed"`
	BlocksScheduled        float64   `db:"blocks_scheduled"`
	SyncExecuted           float64   `db:"sync_executed"`
	SyncScheduled          float64   `db:"sync_scheduled"`
}

// healthz structs

type HealthzResult struct {
	EventId string               `db:"event_id" json:"-"`
	Status  constants.StatusType `db:"status" json:"status"`
	Result  []map[string]string  `db:"result" json:"reports"`
}

type HealthzData struct {
	TotalOkPercentage float64                    `json:"total_ok_percentage"`
	ReportingUUID     string                     `json:"reporting_uuid"`
	DeploymentType    string                     `json:"deployment_type"`
	Reports           map[string][]HealthzResult `json:"status_reports"`
}

// -------------------------
// Mobile structs

type MobileAppBundleStats struct {
	LatestBundleVersion uint64 `db:"bundle_version"`
	BundleUrl           string `db:"bundle_url"`
	TargetCount         int64  `db:"target_count"` // coalesce to -1 if column is null
	DeliveryCount       int64  `db:"delivered_count"`
	MaxNativeVersion    uint64 `db:"max_native_version"` // the max native version of the whole table for the given environment
}

// Notification structs

type NotificationSettingsDefaultValues struct {
	GroupEfficiencyBelowThreshold     float64
	MaxCollateralThreshold            float64
	MinCollateralThreshold            float64
	ERC20TokenTransfersValueThreshold float64

	MachineStorageUsageThreshold float64
	MachineCpuUsageThreshold     float64
	MachineMemoryUsageThreshold  float64

	GasAboveThreshold                 decimal.Decimal
	GasBelowThreshold                 decimal.Decimal
	NetworkParticipationRateThreshold float64
}

// ------------------------------

type CtxKey string

const CtxUserIdKey CtxKey = "user_id"
const CtxIsMockedKey CtxKey = "is_mocked"
const CtxMockSeedKey CtxKey = "mock_seed"
const CtxDashboardIdKey CtxKey = "dashboard_id"
